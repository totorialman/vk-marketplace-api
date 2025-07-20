package log

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/satori/uuid"
	"github.com/gorilla/mux"
)

func CreateLoggerMiddleware(logger *slog.Logger) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), "logger", logger.With(slog.String("ID", uuid.NewV4().String())))
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		})
	}
}