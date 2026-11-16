# package-base Development Guidelines

Auto-generated from all feature plans. Last updated: 2025-10-08

## Active Technologies
- Go 1.24.x + Fiber v2 (HTTP 框架), golang-jwt/jwt v5, GORM, zap, validator, redis(v8)（短期保留仅作缓存/未来用途), aliyun-oss-sdk (可选), go-redis (可选), sqlite / mysql / postgres 驱动（当前只启用 sqlite -> gorm sqlite 驱动 待引入） (001-logto-session)
- Go 1.24.x（项目既有） + Fiber v2（HTTP）、GORM、zap、validator、JWT、sqlite/mysql/postgres 驱动（项目已有，仅启用 sqlite 驱动待引入说明） (002-rbac)
- 复用现有数据库（按项目当前驱动）；新增 RBAC 相关表/迁移（角色、权限、映射、绑定、决策日志） (002-rbac)

## Project Structure
```
src/
tests/
```

## Commands
# Add commands for Go 1.24.x

## Code Style
Go 1.24.x: Follow standard conventions

## Recent Changes
- 002-rbac: Added Go 1.24.x（项目既有） + Fiber v2（HTTP）、GORM、zap、validator、JWT、sqlite/mysql/postgres 驱动（项目已有，仅启用 sqlite 驱动待引入说明）
- 001-logto-session: Added Go 1.24.x + Fiber v2 (HTTP 框架), golang-jwt/jwt v5, GORM, zap, validator, redis(v8)（短期保留仅作缓存/未来用途), aliyun-oss-sdk (可选), go-redis (可选), sqlite / mysql / postgres 驱动（当前只启用 sqlite -> gorm sqlite 驱动 待引入）

<!-- MANUAL ADDITIONS START -->
<!-- MANUAL ADDITIONS END -->
