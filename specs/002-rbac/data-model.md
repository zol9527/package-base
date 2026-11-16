# Data Model: RBAC 权限管理与资源访问控制

## Entities

### Role
- Fields: id, name, description, built_in(bool), enabled(bool), parents[] (for inheritance), created_at, updated_at
- Rules: name 全局唯一；禁止循环继承；内置角色不可删除（可禁用分配）。

### Permission
- Fields: id, resource(string), action(enum: read|write|delete|manage|custom), effect(enum: allow|deny), note, created_at, updated_at
- Rules: resource 使用 “域:子域:标识” 命名；支持通配符 `*`；(resource, action, effect) 在角色聚合内需可去重。

### RolePermission
- Fields: id, role_id, permission_id
- Rules: (role_id, permission_id) 唯一。

### UserRoleBinding
- Fields: id, user_id, role_id, scope(optional), expires_at(optional), created_at, updated_at
- Rules: (user_id, role_id, scope) 唯一；过期后无效。

### AccessDecision
- Fields: id, user_id, resource, action, allowed(bool), reason(string), request_id, client_info(optional), created_at
- Rules: 仅记录必要信息，避免泄露策略细节。

## Relationships
- Role 1..* Permission（经 RolePermission）
- User 1..* Role（经 UserRoleBinding），可选 scope
- Role *..* Role（parents[] 继承）

## Validation & State
- 继承更新时执行环检测；禁用角色后，其权限不再参与聚合。
- 绑定过期或用户禁用时，授权直接拒绝。
