package utils

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
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

// GenerateAccessToken 生成访问Token
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
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(j.secret)
	if err != nil {
		return "", time.Time{}, errors.Wrap(err, "failed to sign access token")
	}

	return tokenString, expiresAt, nil
}

// GenerateRefreshToken 生成刷新Token
func (j *JWTManager) GenerateRefreshToken(userID string) (string, time.Time, error) {
	expiresAt := time.Now().Add(j.refreshTokenDuration)

	claims := &jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(expiresAt),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		NotBefore: jwt.NewNumericDate(time.Now()),
		Subject:   userID,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(j.secret)
	if err != nil {
		return "", time.Time{}, errors.Wrap(err, "failed to sign refresh token")
	}

	return tokenString, expiresAt, nil
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
