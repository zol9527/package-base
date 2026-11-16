package interfaces

import "context"

// Function: AuthorizationService
//
// Description: 鉴权服务接口，提供对用户在资源上的动作进行授权判定的能力。
//
// Parameters:
//
//    无
//
// Returns:
//
//    无
type AuthorizationService interface {
    // Function: Authorize
    //
    // Description: 判断给定用户对指定资源执行某动作是否被允许，并返回拒绝原因（如有）。
    //
    // Parameters:
    //
    //    ctx(context.Context): 上下文
    //    userID(string): 用户标识
    //    resource(string): 资源名称（如 domain:sub:id 或 路径）
    //    action(string): 动作（如 read/write/delete/manage 或 HTTP 方法）
    //
    // Returns:
    //
    //    allowed(bool): 是否允许
    //    reason(string): 拒绝原因（允许时可为空）
    Authorize(ctx context.Context, userID, resource, action string) (allowed bool, reason string)
}
