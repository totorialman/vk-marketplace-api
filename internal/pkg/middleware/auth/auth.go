package auth

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/golang-jwt/jwt"
	"github.com/gorilla/mux"
	"github.com/satori/uuid"
	"github.com/totorialman/vk-marketplace-api/internal/pkg/utils/log"
)

const (
	UserIDKey    = "user_id"
	UserLoginKey = "user_login"
)

func WithAuth(secret string) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger := log.GetLoggerFromContext(r.Context()).With(slog.String("middleware", "WithAuth"))

			tokenStr := r.Header.Get("MarketplaceJWT")
			if tokenStr == "" {
				logger.Info("Отсутствует MarketplaceJWT в заголовках")
				next.ServeHTTP(w, r)
				return
			}

			claims := jwt.MapClaims{}
			token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method")
				}
				if secret == "" {
					return nil, fmt.Errorf("JWT_SECRET не установлен")
				}
				return []byte(secret), nil
			})

			if err != nil {
				logger.Warn("Ошибка парсинга JWT токена: " + err.Error())
				next.ServeHTTP(w, r)
				return
			}

			if !token.Valid {
				logger.Warn("JWT токен не валиден")
				next.ServeHTTP(w, r)
				return
			}

			userIDStr, ok := claims["id"].(string)
			if !ok {
				logger.Warn("JWT токен не содержит id пользователя")
				next.ServeHTTP(w, r)
				return
			}

			userID, err := uuid.FromString(userIDStr)
			if err != nil {
				logger.Warn("Невалидный UUID пользователя в JWT: " + err.Error())
				next.ServeHTTP(w, r)
				return
			}

			userLogin, ok := claims["login"].(string)
			if !ok {
				logger.Warn("JWT токен не содержит login пользователя")
				next.ServeHTTP(w, r)
				return
			}

			ctx := context.WithValue(r.Context(), UserIDKey, userID)
			ctx = context.WithValue(ctx, UserLoginKey, userLogin)
			r = r.WithContext(ctx)

			logger.Info("Пользователь аутентифицирован", slog.String("userID", userID.String()), slog.String("login", userLogin))

			next.ServeHTTP(w, r)
		})
	}
}
