package utils

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

// JWTClaims JWT声明结构
type JWTClaims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     int    `json:"role"`
	jwt.RegisteredClaims
}

// JWTManager JWT管理器
type JWTManager struct {
	secret               []byte
	accessTokenDuration  time.Duration
	refreshTokenDuration time.Duration
}

// NewJWTManager 创建JWT管理器
func NewJWTManager(secret string, accessDuration, refreshDuration time.Duration) *JWTManager {
	return &JWTManager{
		secret:               []byte(secret),
		accessTokenDuration:  accessDuration,
		refreshTokenDuration: refreshDuration,
	}
}

// Function: GenerateAccessToken
//
// Description: 生成访问 Token，携带用户基础身份与角色信息；为保证在同一秒内多次生成也能完成轮换，增加随机 JTI。
//
// Parameters:
//
//	userID(string): 用户ID
//	username(string): 用户名
//	email(string): 邮箱
//	role(int): 角色标识
//
// Returns:
//
//	token(string): 签名后的访问 token 字符串
//	expires_at(time.Time): 过期时间
//	err(error): 可能的错误
func (j *JWTManager) GenerateAccessToken(userID, username, email string, role int) (string, time.Time, error) {
	expiresAt := time.Now().Add(j.accessTokenDuration)

	claims := &JWTClaims{
		UserID:   userID,
		Username: username,
		Email:    email,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Subject:   userID,
			ID:        uuid.NewString(), // 保证同一秒生成的 token 也不同，实现 access_token 轮换
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(j.secret)
	if err != nil {
		return "", time.Time{}, errors.Wrap(err, "failed to sign access token")
	}

	return tokenString, expiresAt, nil
}

// Function: SetAccessTokenDuration
//
// Description: 动态更新访问 token 有效期（用于 TOKEN_TTL 配置热更新场景）。
//
// Parameters:
//
//	d(dur time.Duration): 新的访问 token 时长
//
// Returns:
//
//	无
func (j *JWTManager) SetAccessTokenDuration(d time.Duration) {
	if d > 0 {
		j.accessTokenDuration = d
	}
}

// Function: GetAccessTokenDuration
//
// Description: 返回当前访问 token 时长。
//
// Parameters:
//
//	无
//
// Returns:
//
//	duration(time.Duration): 当前时长
func (j *JWTManager) GetAccessTokenDuration() time.Duration { return j.accessTokenDuration }

// GenerateRefreshToken 生成刷新Token
func (j *JWTManager) GenerateRefreshToken(userID string) (string, time.Time, error) {
	expiresAt := time.Now().Add(j.refreshTokenDuration)

	// 增加随机 JTI 以避免同一秒内重复生成产生相同 token
	claims := &jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(expiresAt),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		NotBefore: jwt.NewNumericDate(time.Now()),
		Subject:   userID,
		ID:        uuid.NewString(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(j.secret)
	if err != nil {
		return "", time.Time{}, errors.Wrap(err, "failed to sign refresh token")
	}

	return tokenString, expiresAt, nil
}

// Function: SetRefreshTokenDuration
//
// Description: 动态更新刷新 token 有效期，用于测试或运行时策略调整（如缩短存活进行安全策略收紧）。
//
// Parameters:
//
//	d(time.Duration): 新的刷新 token 时长，必须 >0 才会生效
//
// Returns:
//
//	无
func (j *JWTManager) SetRefreshTokenDuration(d time.Duration) {
	if d > 0 {
		j.refreshTokenDuration = d
	}
}

// ValidateAccessToken 验证访问Token
func (j *JWTManager) ValidateAccessToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return j.secret, nil
	})

	if err != nil {
		return nil, errors.Wrap(err, "failed to parse access token")
	}

	if !token.Valid {
		return nil, errors.New("invalid access token")
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	return claims, nil
}

// ValidateRefreshToken 验证刷新Token
func (j *JWTManager) ValidateRefreshToken(tokenString string) (string, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return j.secret, nil
	})

	if err != nil {
		return "", errors.Wrap(err, "failed to parse refresh token")
	}

	if !token.Valid {
		return "", errors.New("invalid refresh token")
	}

	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok {
		return "", errors.New("invalid token claims")
	}

	return claims.Subject, nil
}

// ExtractUserIDFromToken 从Token中提取用户ID
func (j *JWTManager) ExtractUserIDFromToken(tokenString string) (string, error) {
	claims, err := j.ValidateAccessToken(tokenString)
	if err != nil {
		return "", err
	}
	return claims.UserID, nil
}
