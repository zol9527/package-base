package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
)

// Function: main
//
// Description: 基于性能结果 JSON 执行阈值守护，超过阈值返回非0退出码。支持 summary 或 single/stair 输出。
//
// Parameters:
//
// Returns:
func main() {
	file := flag.String("file", "perf/login_single_baseline.json", "要检查的性能结果文件")
	maxP95 := flag.Float64("max-p95", 350.0, "允许的最大 p95 毫秒（超过则失败）")
	mode := flag.String("mode", "auto", "auto|single|summary|stair")
	flag.Parse()

	b, err := os.ReadFile(*file)
	if err != nil {
		fmt.Printf("ERROR: 读取文件失败: %v\n", err)
		os.Exit(2)
	}

	var m map[string]interface{}
	if err := json.Unmarshal(b, &m); err != nil {
		fmt.Printf("ERROR: 解析 JSON 失败: %v\n", err)
		os.Exit(2)
	}

	chosen := *mode
	if chosen == "auto" {
		if _, ok := m["median_p95_ms"]; ok && m["runs"] != nil {
			chosen = "single"
		} else if _, ok := m["single_concurrency_8"]; ok {
			chosen = "summary"
		} else if m["mode"] == "stair" {
			chosen = "stair"
		} else if m["mode"] == "single" {
			chosen = "single"
		}
	}

	switch chosen {
	case "single":
		if v, ok := m["median_p95_ms"].(float64); ok {
			check(*maxP95, v)
		} else {
			fmt.Println("ERROR: 未找到 median_p95_ms")
			os.Exit(2)
		}
	case "summary":
		sc8, ok := m["single_concurrency_8"].(map[string]interface{})
		if !ok {
			fmt.Println("ERROR: summary 缺少 single_concurrency_8")
			os.Exit(2)
		}
		v, ok := sc8["median_p95_ms"].(float64)
		if !ok {
			fmt.Println("ERROR: summary 缺少 median_p95_ms")
			os.Exit(2)
		}
		check(*maxP95, v)
	case "stair":
		// 对 stair 取所有 concurrency 的 median_p95 最大值进行校验
		pcs, ok := m["per_concurrency"].([]interface{})
		if !ok {
			fmt.Println("ERROR: stair 缺少 per_concurrency")
			os.Exit(2)
		}
		maxFound := 0.0
		for _, it := range pcs {
			mm, _ := it.(map[string]interface{})
			if v, ok := mm["median_p95_ms"].(float64); ok {
				if v > maxFound {
					maxFound = v
				}
			}
		}
		check(*maxP95, maxFound)
	default:
		fmt.Println("ERROR: 未知模式")
		os.Exit(2)
	}
}

func check(threshold, value float64) {
	fmt.Printf("Gate: p95=%.3fms threshold=%.3fms\n", value, threshold)
	if value > threshold {
		fmt.Println("FAIL: 性能不达标")
		os.Exit(1)
	}
	fmt.Println("PASS: 性能达标")
}
