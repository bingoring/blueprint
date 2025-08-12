package utils

import (
	"blueprint-module/pkg/models"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID   uint   `json:"user_id"`
	Email    string `json:"email"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// GenerateToken JWT 토큰 생성 (설정 가능한 만료 시간)
func GenerateToken(user *models.User, jwtSecret string) (string, error) {
	return GenerateTokenWithExpiry(user, jwtSecret, 24*time.Hour) // 기본 24시간
}

// GenerateTokenWithExpiry 만료 시간을 지정하여 JWT 토큰 생성
func GenerateTokenWithExpiry(user *models.User, jwtSecret string, expiry time.Duration) (string, error) {
	expirationTime := time.Now().Add(expiry)

	claims := &Claims{
		UserID:   user.ID,
		Email:    user.Email,
		Username: user.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "blueprint",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ValidateToken 토큰 유효성 검사
func ValidateToken(tokenString, jwtSecret string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(jwtSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

// IsTokenExpired 토큰이 만료되었는지 확인
func IsTokenExpired(tokenString, jwtSecret string) bool {
	claims, err := ValidateToken(tokenString, jwtSecret)
	if err != nil {
		return true // 토큰이 유효하지 않으면 만료된 것으로 간주
	}

	// 현재 시간과 만료 시간 비교
	return time.Now().After(claims.ExpiresAt.Time)
}

// GetTokenExpirationTime 토큰의 만료 시간 반환
func GetTokenExpirationTime(tokenString, jwtSecret string) (*time.Time, error) {
	claims, err := ValidateToken(tokenString, jwtSecret)
	if err != nil {
		return nil, err
	}

	expirationTime := claims.ExpiresAt.Time
	return &expirationTime, nil
}

// GetTokenRemainingTime 토큰의 남은 유효 시간 반환
func GetTokenRemainingTime(tokenString, jwtSecret string) (time.Duration, error) {
	expirationTime, err := GetTokenExpirationTime(tokenString, jwtSecret)
	if err != nil {
		return 0, err
	}

	remaining := time.Until(*expirationTime)
	if remaining < 0 {
		return 0, errors.New("token has expired")
	}

	return remaining, nil
}
