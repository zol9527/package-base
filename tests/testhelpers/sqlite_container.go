package testhelpers

// 此文件原计划提供基于 sqlite 的测试容器，但当前项目尚未实现 sqlite storage。
// 之前版本包含重复的 package 声明与未使用的导入，导致 `go test ./...` 构建失败。
// 为保持测试通过并保留后续扩展入口，暂时仅保留文档注释，不导出任何符号。
// 如需实现：
// 1. 引入 gorm sqlite 驱动
// 2. 在 internal/core/storages 新增 sqlite 实现
// 3. 提供 NewSQLiteTestContainer(t *testing.T) 返回替换后的 container。
