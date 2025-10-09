# Quickstart: 基础认证系统启动指南

## 1. 环境准备
```bash
cp .env.example .env  # 若存在
export DB_DRIVER=sqlite            # 可选: mysql|postgres|sqlite
export OBJECT_STORAGE=local        # 可选: oss|local
export TOKEN_TTL=60                # 分钟
export JWT_SECRET=dev_local_secret # HS256 密钥（开发环境）
```

## 2. 启动开发
```bash
go run ./cmd/server/main.go
```
期望输出：服务成功监听端口且 `/health` 返回 200。

## 3. 创建测试用户（临时脚本示例）
```bash
go test ./internal/... -run TestCreateDevUser -count=1
```
（或直接在 SQLite 中插入一条用户记录）

## 4. 调用登录接口
```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"identifier":"devuser","password":"Passw0rd!"}'
```
成功返回：`{"code":0,"message":"ok","data":{"token":"...","expires_in":3600}}`

## 5. Token 验证
将返回的 token 作为 `Authorization: Bearer <token>` 调用未来受保护接口。

## 6. 切换数据库驱动
MySQL:
```bash
export DB_DRIVER=mysql
export DB_DSN="user:pass@tcp(localhost:3306)/dbname?parseTime=true"
```
PostgreSQL:
```bash
export DB_DRIVER=postgres
export DB_DSN="host=127.0.0.1 user=xxx password=xxx dbname=xxx sslmode=disable"
```
重新启动服务即自动使用对应驱动。

## 7. 切换对象存储
OSS:
```bash
export OBJECT_STORAGE=oss
export OSS_ENDPOINT=... 
export OSS_BUCKET=... 
export OSS_ACCESS_KEY=... 
export OSS_ACCESS_SECRET=...
```
本地文件无需额外配置。

## 8. 运行测试
```bash
go test ./... -cover
```
目标：新增代码覆盖率 ≥70%，安全路径 ≥90%。

## 9. 性能基线(可选)
```bash
wrk -t2 -c20 -d30s --latency http://localhost:8080/api/auth/login -s scripts/login.lua
```
记录 p95 与平均值；若 p95 >120ms 需分析（加密/数据库）。

## 10. 故障排查
| 问题 | 可能原因 | 处理 |
|------|----------|------|
| 启动报错缺少 JWT_SECRET | 未设置环境变量 | 设置并重启 |
| 登录总是失败 | 用户不存在或密码不匹配 | 确认散列生成流程 |
| Token exp 不正确 | TOKEN_TTL 未设置或时间漂移 | 校验服务器时钟 |

## 11. 下一步可选扩展
- 引入刷新 token 机制
- 增加密码复杂度策略
- MFA (TOTP) 集成
- 权限/角色模型
