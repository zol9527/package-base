# Quickstart: RBAC 权限管理

## 步骤
1. 启用 RBAC：部署包含授权中间件的版本（详见实现文档）。
2. 创建角色：POST /rbac/roles { name:"Viewer" }
3. 添加权限：POST /rbac/permissions { resource:"billing.*", action:"read", effect:"allow" }
4. 赋予角色权限：PUT /rbac/roles/{roleId}/permissions [...]
5. 绑定用户角色：PUT /rbac/users/{userId}/roles [...]
6. 验证授权：POST /rbac/authorize { userId, resource, action }
7. 查询审计：GET /rbac/audit?userId=...

## 提示
- 默认拒绝：未命中允许规则即 403。
- 显式拒绝优先：任何层级 deny 均直接拒绝。
- 变更立即生效：角色/权限变更后新请求立即按新策略评估。
