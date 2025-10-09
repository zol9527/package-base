package testhelpers

import (
	"encoding/json"
	"fmt"
	"time"
)

// Function: LoginForTokens
//
// Description: 使用用户名密码调用 /api/auth/login 并返回 access/refresh token 原始值；向上抛错用于测试断言。
//
// Parameters:
//
//	httpGetJSON(func(method,url string, body []byte, timeout time.Duration)(int,int,string,[]byte,error)): GET/POST JSON 执行函数（便于注入 mock）
//	addr(string): 服务监听地址 host:port
//	username(string): 用户名
//	password(string): 密码
//	timeout(time.Duration): 请求超时
//
// Returns:
//
//	accessToken(string): 访问 token
//	refreshToken(string): 刷新 token
//	err(error): 错误
func LoginForTokens(httpGetJSON func(string, string, []byte, time.Duration) (int, int, string, []byte, error), addr, username, password string, timeout time.Duration) (string, string, error) {
	body, _ := json.Marshal(map[string]string{"username": username, "password": password})
	status, code, _, data, err := httpGetJSON("POST", fmt.Sprintf("http://%s/api/auth/login", addr), body, timeout)
	if err != nil {
		return "", "", err
	}
	if status != 200 || code != 0 {
		return "", "", fmt.Errorf("login failed status=%d code=%d", status, code)
	}
	var parsed struct {
		RefreshToken string `json:"refresh_token"`
		AccessToken  string `json:"access_token"`
	}
	if err := json.Unmarshal(data, &parsed); err != nil {
		return "", "", err
	}
	if parsed.AccessToken == "" || parsed.RefreshToken == "" {
		return "", "", fmt.Errorf("tokens empty")
	}
	return parsed.AccessToken, parsed.RefreshToken, nil
}

// Function: RefreshTokens
//
// Description: 调用 /api/auth/refresh 获取新的 access/refresh；不解析业务 user 信息。
//
// Parameters:
//
//	httpGetJSON(func(method,url string, body []byte, timeout time.Duration)(int,int,string,[]byte,error)): JSON 请求函数
//	addr(string): 服务地址
//	refreshToken(string): 旧刷新 token
//	timeout(time.Duration): 超时
//
// Returns:
//
//	newAccess(string): 新 access token
//	newRefresh(string): 新 refresh token
//	err(error): 错误
func RefreshTokens(httpGetJSON func(string, string, []byte, time.Duration) (int, int, string, []byte, error), addr, refreshToken string, timeout time.Duration) (string, string, error) {
	body, _ := json.Marshal(map[string]string{"refresh_token": refreshToken})
	status, code, _, data, err := httpGetJSON("POST", fmt.Sprintf("http://%s/api/auth/refresh", addr), body, timeout)
	if err != nil {
		return "", "", err
	}
	if status != 200 || code != 0 {
		return "", "", fmt.Errorf("refresh failed status=%d code=%d", status, code)
	}
	var parsed struct {
		RefreshToken string `json:"refresh_token"`
		AccessToken  string `json:"access_token"`
	}
	if err := json.Unmarshal(data, &parsed); err != nil {
		return "", "", err
	}
	if parsed.AccessToken == "" || parsed.RefreshToken == "" {
		return "", "", fmt.Errorf("tokens empty")
	}
	return parsed.AccessToken, parsed.RefreshToken, nil
}
