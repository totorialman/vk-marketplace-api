package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v4/pgxpool"
	authHandler "github.com/totorialman/vk-marketplace-api/internal/pkg/auth/delivery/http"
	authRepo "github.com/totorialman/vk-marketplace-api/internal/pkg/auth/repo"
	authUsecase "github.com/totorialman/vk-marketplace-api/internal/pkg/auth/usecase"
	productHandler "github.com/totorialman/vk-marketplace-api/internal/pkg/product/delivery/http"
	productRepo "github.com/totorialman/vk-marketplace-api/internal/pkg/product/repo"
	productUsecase "github.com/totorialman/vk-marketplace-api/internal/pkg/product/usecase"
	middleware "github.com/totorialman/vk-marketplace-api/internal/pkg/middleware/auth"
	"github.com/totorialman/vk-marketplace-api/internal/pkg/middleware/log"
	_ "github.com/totorialman/vk-marketplace-api/docs" 
    httpSwagger "github.com/swaggo/http-swagger"
)

func initDB(logger *slog.Logger) (*pgxpool.Pool, error) {
	connStr := os.Getenv("POSTGRES_CONN")

	pool, err := pgxpool.Connect(context.Background(), connStr)
	if err != nil {
		return nil, err
	}

	err = pool.Ping(context.Background())
	if err != nil {
		return nil, err
	}

	logger.Info("Успешное подключение к PostgreSQL")
	return pool, nil
}

// @title VK Marketplace API
// @version 1.0
// @description API для маркетплейса ВК
// @host localhost:8080
// @BasePath /

// @securityDefinitions.apikey MarketplaceJWT
// @in header
// @name MarketplaceJWT
func main() {
	logFile, err := os.OpenFile(os.Getenv("MAIN_LOG_FILE"), os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		fmt.Println("ошибка при открытии лог файла: " + err.Error())
		return
	}
	defer logFile.Close()

	logger := slog.New(slog.NewJSONHandler(io.MultiWriter(logFile, os.Stdout), &slog.HandlerOptions{Level: slog.LevelInfo}))

	pool, err := initDB(logger)
	if err != nil {
		logger.Error("Ошибка при подключении к PostgreSQL: " + err.Error())
	}
	defer pool.Close()

	logMW := log.CreateLoggerMiddleware(logger)

	authRepo := authRepo.CreateAuthRepo(pool)
	authUsecase := authUsecase.CreateAuthUsecase(authRepo)
	authHandler := authHandler.CreateAuthHandler(authUsecase)

	productRepo := productRepo.CreateProductRepo(pool)
	productUsecase := productUsecase.CreateProductUsecase(productRepo)
	productHandler := productHandler.CreateProductHandler(productUsecase)

	authMiddleware := middleware.WithAuth(os.Getenv("JWT_SECRET"))

	r := mux.NewRouter().PathPrefix("/api").Subrouter()
	r.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Не найдено", http.StatusTeapot)
	})

	r.Use(logMW)
	r.Use(authMiddleware)

	auth := r.PathPrefix("/auth").Subrouter()
	{
		auth.HandleFunc("/signup", authHandler.SignUp).Methods("POST")
		auth.HandleFunc("/signin", authHandler.SignIn).Methods("POST")
	}

	products := r.PathPrefix("/products").Subrouter()
	{
		products.HandleFunc("", productHandler.Create).Methods("POST")
		products.HandleFunc("", productHandler.List).Methods("GET")
	}

	r.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	http.Handle("/", r)
	srv := http.Server{
		Handler:           r,
		Addr:              ":8080",
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		err := srv.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			logger.Error("Ошибка при запуске сервера: " + err.Error())
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	<-stop
	logger.Info("Получен сигнал остановки")

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	err = srv.Shutdown(ctx)
	if err != nil {
		logger.Error("Ошибка при остановке сервера: " + err.Error())
	} else {
		logger.Info("Сервер успешно остановлен")
	}
}
