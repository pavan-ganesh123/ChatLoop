package auth

import (
	"errors"

	"github.com/golang-jwt/jwt/v5"
)

var secretKey = []byte("mysecretkeymysecretkeymysecretkey")

type Claims struct {
	UserID float64  `json:"userId"`
	Email  string `json:"sub"`
	jwt.RegisteredClaims
}

func ValidateToken(tokenStr string) (*Claims, error) {
	
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}