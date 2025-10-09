package main

// 本目录为登录压测工具独立可执行入口，源自原 scripts/perf/login_benchmark.go。
// 拆分目录以避免与同目录的其它 main 冲突，便于 go vet 与构建。

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"sort"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

// Function: main
//
// Description: 入口函数，解析参数，检测/启动服务，执行并发登录压测并输出统计 JSON 到项目根目录 perf_login_result.json。
//
// Parameters:
//
// Returns:
func main() {
	// 参数定义
	duration := flag.Duration("duration", 5*time.Second, "压测持续时间，例如 10s / 30s")
	concurrency := flag.Int("concurrency", 10, "并发 goroutine 数 (与 -stair 互斥)")
	target := flag.String("target", "http://localhost:8080", "服务基础地址 (不含 path)")
	username := flag.String("username", "tester", "登录用户名")
	password := flag.String("password", "Passw0rd!", "登录密码")
	serverCmd := flag.String("server-cmd", "go", "若需要自动启动服务时使用的命令 (go)")
	serverArgsStr := flag.String("server-args", "run ./cmd/server/main.go", "自动启动服务附加参数")
	warmup := flag.Duration("warmup", 1*time.Second, "预热时间 (不计入统计)")
	bcryptCost := flag.Int("bcrypt-cost", -1, "覆盖运行时 BCRYPT_COST (仅脚本自启动服务时生效)")
	output := flag.String("output", "perf_login_result.json", "结果输出文件路径")
	noRegister := flag.Bool("no-register", false, "不尝试注册测试用户")
	runs := flag.Int("runs", 1, "重复执行次数用于统计 p95 中位数")
	sleepBetween := flag.Duration("sleep-between", 500*time.Millisecond, "多轮之间的休眠间隔")
	stair := flag.String("stair", "", "阶梯并发列表, 逗号分隔 例如: 4,8,16 (覆盖 -concurrency)")
	flag.Parse()

	loginURL := fmt.Sprintf("%s/api/auth/login", *target)

	// 捕获中断信号，便于清理子进程
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// 如果服务未启动则尝试拉起
	if !waitHTTP(loginURL, 500*time.Millisecond) {
		fmt.Println("[INFO] 尝试启动服务进程 ...")
		args := splitArgs(*serverArgsStr)
		cmd := exec.CommandContext(ctx, *serverCmd, args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		// 注入可选的 BCRYPT_COST
		if *bcryptCost > 0 {
			env := os.Environ()
			env = append(env, fmt.Sprintf("BCRYPT_COST=%d", *bcryptCost))
			cmd.Env = env
			fmt.Printf("[INFO] 设置 BCRYPT_COST=%d\n", *bcryptCost)
		}
		if err := cmd.Start(); err != nil {
			fmt.Printf("[ERROR] 启动服务失败: %v\n", err)
			writeResultError("start_failed", err)
			return
		}
		// 等待就绪
		ready := waitReady(loginURL, 30*time.Second)
		if !ready {
			fmt.Println("[ERROR] 服务在超时时间内未就绪")
			writeResultError("server_unreachable", errors.New("timeout waiting for server"))
			return
		}
		defer func() {
			_ = cmd.Process.Signal(syscall.SIGINT)
			done := make(chan struct{})
			go func() {
				cmd.Wait()
				close(done)
			}()
			select {
			case <-done:
			case <-time.After(3 * time.Second):
				_ = cmd.Process.Kill()
			}
		}()
	} else {
		fmt.Println("[INFO] 检测到服务已运行，直接进行压测")
	}

	// 确保测试用户存在（尝试注册，忽略已存在错误）除非指定 no-register
	if !*noRegister {
		if err := ensureUser(*target, *username, *password); err != nil {
			fmt.Printf("[WARN] 注册测试用户失败（可能已存在）: %v\n", err)
		}
	} else {
		fmt.Println("[INFO] 跳过测试用户注册 (-no-register)")
	}

	// 预热阶段
	if *warmup > 0 {
		warmCtx, cancel := context.WithTimeout(ctx, *warmup)
		runLoad(warmCtx, loginURL, *username, *password, 5, nil)
		cancel()
		fmt.Println("[INFO] 预热完成")
	}

	// 根据 stair / runs 模式执行
	if *stair != "" {
		concList := parseStair(*stair)
		out := &stairOutput{Mode: "stair", Target: loginURL, Timestamp: time.Now().Format(time.RFC3339)}
		for _, c := range concList {
			fmt.Printf("[INFO] 阶梯并发 %d 开始 (runs=%d) ...\n", c, *runs)
			var runAggs []*agg
			for i := 0; i < *runs; i++ {
				benchCtx, cancelBench := context.WithTimeout(ctx, *duration)
				a := &agg{}
				runLoad(benchCtx, loginURL, *username, *password, c, a)
				cancelBench()
				a.Finalize(*duration)
				runAggs = append(runAggs, a)
				if i < *runs-1 {
					time.Sleep(*sleepBetween)
				}
			}
			out.PerConcurrency = append(out.PerConcurrency, summarizeRuns(c, runAggs))
		}
		writeGenericResult(out, *output)
		return
	}

	// 多轮单并发模式
	var runAggs []*agg
	for i := 0; i < *runs; i++ {
		benchCtx, cancelBench := context.WithTimeout(ctx, *duration)
		a := &agg{}
		runLoad(benchCtx, loginURL, *username, *password, *concurrency, a)
		cancelBench()
		a.Finalize(*duration)
		runAggs = append(runAggs, a)
		if i < *runs-1 {
			time.Sleep(*sleepBetween)
		}
	}
	result := buildMultiRunOutput(loginURL, *concurrency, *duration, runAggs)
	writeGenericResult(result, *output)
}

// Function: splitArgs
//
// Description: 简单拆分命令参数（按空格）；不处理引号嵌套场景。
//
// Parameters:
//
//	s(string): 参数字符串
//
// Returns:
//
//	parts([]string): 拆分后的片段
func splitArgs(s string) []string {
	var parts []string
	for _, p := range bytes.Fields([]byte(s)) {
		parts = append(parts, string(p))
	}
	return parts
}

// Function: waitHTTP
//
// Description: 快速探测单次 HTTP 可用性。
//
// Parameters:
//
//	url(string): 目标 URL
//	timeout(time.Duration): 超时
//
// Returns:
//
//	ok(bool): 是否可达
func waitHTTP(url string, timeout time.Duration) bool {
	c := http.Client{Timeout: timeout}
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	resp, err := c.Do(req)
	if err != nil {
		return false
	}
	resp.Body.Close()
	return true
}

// Function: waitReady
//
// Description: 轮询等待服务就绪（返回 200/400/401 视为端口开放）。
//
// Parameters:
//
//	url(string): 探测 URL
//	timeout(time.Duration): 最大等待
//
// Returns:
//
//	ready(bool): 是否就绪
func waitReady(url string, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if waitHTTP(url, 800*time.Millisecond) {
			return true
		}
		time.Sleep(300 * time.Millisecond)
	}
	return false
}

// Function: ensureUser
//
// Description: 调用注册接口以确保用户存在；若用户已存在则忽略错误。
//
// Parameters:
//
//	base(string): 基础地址 (不含路径)
//	username(string): 用户名
//	password(string): 密码
//
// Returns:
//
//	err(error): 可能的错误（忽略用户名/邮箱已存在）
func ensureUser(base, username, password string) error {
	regURL := fmt.Sprintf("%s/api/auth/register", base)
	payload := map[string]string{
		"username": username,
		"email":    fmt.Sprintf("%s@example.com", username),
		"password": password,
	}
	b, _ := json.Marshal(payload)
	resp, err := http.Post(regURL, "application/json", bytes.NewReader(b))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == 200 {
		return nil
	}
	body, _ := io.ReadAll(resp.Body)
	// 简单判断已存在
	if bytes.Contains(body, []byte("AUTH_USERNAME_EXISTS")) || bytes.Contains(body, []byte("AUTH_EMAIL_EXISTS")) {
		return nil
	}
	return fmt.Errorf("register unexpected status=%d body=%s", resp.StatusCode, string(body))
}

// 登录请求 JSON 模板（复用以减少分配）
func buildPayload(username, password string) []byte {
	m := map[string]string{"username": username, "password": password}
	b, _ := json.Marshal(m)
	return b
}

// 单次登录尝试，返回延迟与是否成功
func doLogin(client *http.Client, url string, payload []byte) (time.Duration, bool, error) {
	start := time.Now()
	resp, err := client.Post(url, "application/json", bytes.NewReader(payload))
	if err != nil {
		return 0, false, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	// 解析 code 字段 (成功 code=0)
	var parsed struct {
		Code interface{} `json:"code"`
	}
	_ = json.Unmarshal(body, &parsed)
	success := false
	switch v := parsed.Code.(type) {
	case float64:
		success = v == 0
	case int:
		success = v == 0
	case string:
		success = v == "0" || v == "SUCCESS" || v == ""
	}
	return time.Since(start), success && resp.StatusCode == 200, nil
}

// 结果聚合结构
type agg struct {
	Latencies   []time.Duration `json:"-"`
	Total       int64           `json:"total"`
	Success     int64           `json:"success"`
	Errors      int64           `json:"errors"`
	RPS         float64         `json:"rps"`
	P50         float64         `json:"p50_ms"`
	P95         float64         `json:"p95_ms"`
	ErrorRate   float64         `json:"error_rate"`
	Duration    string          `json:"duration"`
	Concurrency int             `json:"concurrency"`
	Timestamp   string          `json:"timestamp"`
	Target      string          `json:"target"`
}

// 多轮输出结构
type multiRunOutput struct {
	Mode        string    `json:"mode"`
	Target      string    `json:"target"`
	Concurrency int       `json:"concurrency"`
	Duration    string    `json:"duration"`
	Runs        []runStat `json:"runs"`
	MedianP95   float64   `json:"median_p95_ms"`
	MedianP50   float64   `json:"median_p50_ms"`
	BestP95     float64   `json:"best_p95_ms"`
	WorstP95    float64   `json:"worst_p95_ms"`
	Timestamp   string    `json:"timestamp"`
}

// 阶梯并发输出结构
type stairOutput struct {
	Mode           string      `json:"mode"`
	Target         string      `json:"target"`
	Timestamp      string      `json:"timestamp"`
	PerConcurrency []stairStat `json:"per_concurrency"`
}

type runStat struct {
	Index   int     `json:"index"`
	Total   int64   `json:"total"`
	Success int64   `json:"success"`
	Errors  int64   `json:"errors"`
	RPS     float64 `json:"rps"`
	P50     float64 `json:"p50_ms"`
	P95     float64 `json:"p95_ms"`
}

type stairStat struct {
	Concurrency int       `json:"concurrency"`
	Runs        []runStat `json:"runs"`
	MedianP95   float64   `json:"median_p95_ms"`
	MedianP50   float64   `json:"median_p50_ms"`
	BestP95     float64   `json:"best_p95_ms"`
	WorstP95    float64   `json:"worst_p95_ms"`
}

// Function: Finalize
//
// Description: 根据收集的延迟切片计算统计指标（p50/p95/RPS/错误率）。
//
// Parameters:
//
//	d(time.Duration): 实际压测持续时间
//
// Returns:
//
//	无
func (a *agg) Finalize(d time.Duration) {
	if len(a.Latencies) == 0 {
		return
	}
	sort.Slice(a.Latencies, func(i, j int) bool { return a.Latencies[i] < a.Latencies[j] })
	p := func(pct float64) float64 {
		if len(a.Latencies) == 0 {
			return 0
		}
		k := int(math.Ceil((pct/100.0)*float64(len(a.Latencies)))) - 1
		if k < 0 {
			k = 0
		}
		if k >= len(a.Latencies) {
			k = len(a.Latencies) - 1
		}
		return float64(a.Latencies[k].Microseconds()) / 1000.0
	}
	a.P50 = p(50)
	a.P95 = p(95)
	a.RPS = float64(a.Total) / d.Seconds()
	a.ErrorRate = 0
	if a.Total > 0 {
		a.ErrorRate = float64(a.Errors) / float64(a.Total)
	}
	a.Duration = d.String()
	a.Timestamp = time.Now().Format(time.RFC3339)
}

// 并发执行负载
func runLoad(ctx context.Context, url, username, password string, conc int, out *agg) {
	payload := buildPayload(username, password)
	var wg sync.WaitGroup
	if out != nil {
		out.Concurrency = conc
		out.Target = url
	}
	latMu := sync.Mutex{}
	client := &http.Client{Timeout: 5 * time.Second}
	for i := 0; i < conc; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				default:
				}
				dur, ok, err := doLogin(client, url, payload)
				if out != nil {
					if err != nil {
						atomic.AddInt64(&out.Errors, 1)
					} else {
						if !ok {
							atomic.AddInt64(&out.Errors, 1)
						} else {
							atomic.AddInt64(&out.Success, 1)
						}
						latMu.Lock()
						out.Latencies = append(out.Latencies, dur)
						latMu.Unlock()
					}
					atomic.AddInt64(&out.Total, 1)
				}
			}
		}()
	}
	wg.Wait()
}

// 输出通用结果（单轮 / 多轮 / 阶梯）
func writeGenericResult(v interface{}, file string) {
	f, err := os.Create(file)
	if err != nil {
		fmt.Printf("[ERROR] 写入结果文件失败: %v\n", err)
		return
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		fmt.Printf("[ERROR] 编码结果失败: %v\n", err)
		return
	}
	fmt.Printf("[INFO] 结果已写入 %s\n", file)
}

// 输出错误占位
func writeResultError(code string, err error) {
	file := "perf_login_result.json"
	payload := map[string]interface{}{
		"timestamp": time.Now().Format(time.RFC3339),
		"error":     code,
		"message":   err.Error(),
	}
	b, _ := json.MarshalIndent(payload, "", "  ")
	_ = os.WriteFile(file, b, 0644)
	fmt.Printf("[WARN] 写入占位结果: %s (%v)\n", code, err)
}

// 构建多轮输出
func buildMultiRunOutput(target string, conc int, dur time.Duration, runs []*agg) *multiRunOutput {
	stats := make([]runStat, 0, len(runs))
	var p95s, p50s []float64
	best := math.MaxFloat64
	worst := 0.0
	for i, a := range runs {
		rs := runStat{Index: i, Total: a.Total, Success: a.Success, Errors: a.Errors, RPS: a.RPS, P50: a.P50, P95: a.P95}
		stats = append(stats, rs)
		p95s = append(p95s, a.P95)
		p50s = append(p50s, a.P50)
		if a.P95 < best {
			best = a.P95
		}
		if a.P95 > worst {
			worst = a.P95
		}
	}
	return &multiRunOutput{
		Mode:        "single",
		Target:      target,
		Concurrency: conc,
		Duration:    dur.String(),
		Runs:        stats,
		MedianP95:   median(p95s),
		MedianP50:   median(p50s),
		BestP95:     best,
		WorstP95:    worst,
		Timestamp:   time.Now().Format(time.RFC3339),
	}
}

// 汇总阶梯每个并发的多轮结果
func summarizeRuns(conc int, runs []*agg) stairStat {
	var p95s, p50s []float64
	best := math.MaxFloat64
	worst := 0.0
	var rs []runStat
	for i, a := range runs {
		rs = append(rs, runStat{Index: i, Total: a.Total, Success: a.Success, Errors: a.Errors, RPS: a.RPS, P50: a.P50, P95: a.P95})
		p95s = append(p95s, a.P95)
		p50s = append(p50s, a.P50)
		if a.P95 < best {
			best = a.P95
		}
		if a.P95 > worst {
			worst = a.P95
		}
	}
	return stairStat{Concurrency: conc, Runs: rs, MedianP95: median(p95s), MedianP50: median(p50s), BestP95: best, WorstP95: worst}
}

// 解析阶梯并发列表
func parseStair(s string) []int {
	var res []int
	for _, part := range bytes.Split([]byte(s), []byte{','}) {
		p := bytes.TrimSpace(part)
		if len(p) == 0 {
			continue
		}
		var v int
		fmt.Sscanf(string(p), "%d", &v)
		if v > 0 {
			res = append(res, v)
		}
	}
	return res
}

// 计算中位数
func median(vals []float64) float64 {
	if len(vals) == 0 {
		return 0
	}
	cp := make([]float64, len(vals))
	copy(cp, vals)
	sort.Slice(cp, func(i, j int) bool { return cp[i] < cp[j] })
	n := len(cp)
	if n%2 == 1 {
		return cp[n/2]
	}
	return (cp[n/2-1] + cp[n/2]) / 2
}
