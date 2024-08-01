package token

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTMaker struct {
	SecretKey string
}

func NewJWTMaker(secretKey string) *JWTMaker {
	return &JWTMaker{
		SecretKey: secretKey,
	}
}

func (maker *JWTMaker) CreateToken(id int, email string, isAdmin bool, duration time.Duration) (string, *UserClaims, error) {
	claims, err := NewUserClaims(id, email, isAdmin, duration)
	if err != nil {
		return "", nil, err
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(maker.SecretKey))
	if err != nil {
		return "", nil, fmt.Errorf("Error signing method: %w", err)
	}
	return tokenString, claims, nil
}

func (maker *JWTMaker) VerifyToken(tokenStr string) (*UserClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		_, ok := token.Method.(*jwt.SigningMethodHMAC)
		if !ok {
			return nil, fmt.Errorf("invalid signing method token")
		}
		return []byte(maker.SecretKey), nil
	})
	if err != nil {
		return nil, fmt.Errorf("Error parsing token: %w", err)
	}
	claims, ok := token.Claims.(*UserClaims)
	if !ok {
		return nil, fmt.Errorf("Invalid token claims")
	}
	return claims, nil
}
