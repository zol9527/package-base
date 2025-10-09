//go:build ignore
// +build ignore

// Deprecated: 已迁移至 scripts/perf/summarize_login/main.go。本文件仅为保留历史，构建与 vet 将忽略。
package main

// 汇总 perf 基线：读取 login_single_baseline.json 与 login_stair_baseline.json，生成 perf/login_baseline_summary.json。

import (
	"encoding/json"
	"fmt"
	"os"
)

func main() {
	single := readJSON("perf/login_single_baseline.json")
	stair := readJSON("perf/login_stair_baseline.json")

	out := map[string]interface{}{}
	if single != nil {
		out["single_concurrency_8"] = single
	}
	if stair != nil {
		out["stair"] = stair
	}

	b, _ := json.MarshalIndent(out, "", "  ")
	if err := os.WriteFile("perf/login_baseline_summary.json", b, 0644); err != nil {
		fmt.Println("write error:", err)
		os.Exit(1)
	}
	fmt.Println("summary written: perf/login_baseline_summary.json")
}

func readJSON(path string) map[string]interface{} {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var m map[string]interface{}
	if err := json.Unmarshal(b, &m); err != nil {
		return nil
	}
	return m
}
