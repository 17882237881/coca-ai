package jwtx

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTHandler struct {
	SigningKey []byte
}

func NewJWTHandler() *JWTHandler {
	return &JWTHandler{
		// TODO: Move to config
		SigningKey: []byte("coca-ai-secret-key-999"),
	}
}

type UserClaims struct {
	jwt.RegisteredClaims
	Uid   int64  `json:"uid"`
	Email string `json:"email"`
	Ssid  string `json:"ssid"`
}

type RefreshClaims struct {
	jwt.RegisteredClaims
	Uid   int64  `json:"uid"`
	Email string `json:"email"`
	Ssid  string `json:"ssid"`
}

// GenerateTokens 生成双 Token (Access + Refresh)
func (h *JWTHandler) GenerateTokens(uid int64, email string, ssid string) (string, string, error) {
	// 1. Access Token (1小时)
	ac := UserClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			Issuer:    "coca-ai",
		},
		Uid:   uid,
		Email: email,
		Ssid:  ssid,
	}
	at := jwt.NewWithClaims(jwt.SigningMethodHS256, ac)
	accessToken, err := at.SignedString(h.SigningKey)
	if err != nil {
		return "", "", err
	}

	// 2. Refresh Token (7天)
	rc := RefreshClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24 * 7)),
			Issuer:    "coca-ai",
		},
		Uid:   uid,
		Email: email,
		Ssid:  ssid,
	}
	rt := jwt.NewWithClaims(jwt.SigningMethodHS256, rc)
	refreshToken, err := rt.SignedString(h.SigningKey)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

// ParseToken 校验 Access Token
func (h *JWTHandler) ParseToken(tokenStr string) (*UserClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		return h.SigningKey, nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*UserClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, errors.New("invalid access token")
}

// ParseRefreshToken 校验 Refresh Token
func (h *JWTHandler) ParseRefreshToken(tokenStr string) (*RefreshClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &RefreshClaims{}, func(token *jwt.Token) (interface{}, error) {
		return h.SigningKey, nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*RefreshClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, errors.New("invalid refresh token")
}
