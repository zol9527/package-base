-- login.lua: wrk 脚本占位 (后续可扩展为实际基准测试)
-- 目的：为性能基线任务 (T041, T056, T068) 提供初始文件，避免脚本缺失导致 CI 失败。
-- 后续扩展建议：
-- 1. 预热阶段请求
-- 2. 随机化测试用户凭证
-- 3. 输出 JSON 指标 (p50, p95, throughput)

wrk.method = "POST"
wrk.path   = "/api/auth/login"
wrk.headers["Content-Type"] = "application/json"

local username = os.getenv("PERF_USERNAME") or "perfuser"
local password = os.getenv("PERF_PASSWORD") or "Passw0rd!"

request = function()
  local body = string.format('{"identifier":"%s","password":"%s"}', username, password)
  return wrk.format(nil, nil, nil, body)
end
