package utils

import (
	"fmt"

	"github.com/golang-jwt/jwt"
)

func GetLoginFromJWT(JWTStr string, claims jwt.MapClaims, secret string) (string, bool) {
	token, err := jwt.ParseWithClaims(JWTStr, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		if secret == "" {
			return nil, fmt.Errorf("JWT_SECRET не установлен")
		}
		return []byte(secret), nil
	})
	if err != nil || !token.Valid {
		return "", false
	}

	login, ok := claims["login"].(string)
	return login, ok
}

func GetIdFromJWT(JWTStr string, claims jwt.MapClaims, secret string) (string, bool) {
	token, err := jwt.ParseWithClaims(JWTStr, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		if secret == "" {
			return nil, fmt.Errorf("JWT_SECRET не установлен")
		}
		return []byte(secret), nil
	})
	if err != nil || !token.Valid {
		return "", false
	}

	id, ok := claims["id"].(string)
	return id, ok
}
