package http

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/mailru/easyjson"
	"github.com/satori/uuid"
	"github.com/totorialman/vk-marketplace-api/internal/models"
	"github.com/totorialman/vk-marketplace-api/internal/pkg/auth"
	"github.com/totorialman/vk-marketplace-api/internal/pkg/utils/log"
	utils "github.com/totorialman/vk-marketplace-api/internal/pkg/utils/send_error"
)

type AuthHandler struct {
	uc     auth.AuthUsecase
	secret string
}

func CreateAuthHandler(uc auth.AuthUsecase) *AuthHandler {
	return &AuthHandler{uc: uc, secret: os.Getenv("JWT_SECRET")}
}

func (h *AuthHandler) SignUp(w http.ResponseWriter, r *http.Request) {
	logger := log.GetLoggerFromContext(r.Context()).With(slog.String("func", log.GetFuncName()))

	_, ok := r.Context().Value("user_id").(uuid.UUID)
	if ok {
		log.LogHandlerError(logger, fmt.Errorf("вы уже авторизованы"), http.StatusUnauthorized)
		utils.SendError(w, "вы уже авторизованы", http.StatusUnauthorized)
		return
	}

	var req models.UserReq
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		log.LogHandlerError(logger, fmt.Errorf("ошибка парсинга JSON: %w", err), http.StatusBadRequest)
		utils.SendError(w, "Ошибка парсинга JSON", http.StatusBadRequest)
		return
	}
	req.Sanitize()

	user, err := h.uc.SignUp(r.Context(), req)
	if err != nil {
		switch err {
		case auth.ErrInvalidLogin, auth.ErrInvalidPassword:
			log.LogHandlerError(logger, fmt.Errorf("неправильный логин или пароль: %w", err), http.StatusBadRequest)
			utils.SendError(w, "Неправильный логин или пароль", http.StatusBadRequest)
		case auth.ErrCreatingUser:
			log.LogHandlerError(logger, err, http.StatusBadRequest)
			utils.SendError(w, err.Error(), http.StatusBadRequest)
		default:
			log.LogHandlerError(logger, fmt.Errorf("неизвестная ошибка: %w", err), http.StatusInternalServerError)
			utils.SendError(w, "Неизвестная ошибка", http.StatusInternalServerError)
		}
		return
	}

	data, err := easyjson.Marshal(user)
	if err != nil {
		log.LogHandlerError(logger, fmt.Errorf("ошибка маршалинга: %w", err), http.StatusInternalServerError)
		utils.SendError(w, "Ошибка сервера", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
	log.LogHandlerInfo(logger, "Success", http.StatusOK)
}

func (h *AuthHandler) SignIn(w http.ResponseWriter, r *http.Request) {
	logger := log.GetLoggerFromContext(r.Context()).With(slog.String("func", log.GetFuncName()))

	var req models.UserReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.LogHandlerError(logger, fmt.Errorf("ошибка парсинга JSON: %w", err), http.StatusBadRequest)
		utils.SendError(w, "Ошибка парсинга JSON", http.StatusBadRequest)
		return
	}
	req.Sanitize()

	user, err := h.uc.SignIn(r.Context(), req)

	if err != nil {
		switch err {
		case auth.ErrInvalidLogin, auth.ErrUserNotFound:
			log.LogHandlerError(logger, err, http.StatusBadRequest)
			utils.SendError(w, err.Error(), http.StatusBadRequest)
		case auth.ErrInvalidCredentials:
			log.LogHandlerError(logger, err, http.StatusUnauthorized)
			utils.SendError(w, err.Error(), http.StatusUnauthorized)
		default:
			log.LogHandlerError(logger, fmt.Errorf("неизвестная ошибка: %w", err), http.StatusInternalServerError)
			utils.SendError(w, "Неизвестная ошибка", http.StatusInternalServerError)
		}
		return
	}

	data, err := easyjson.Marshal(user)
	if err != nil {
		log.LogHandlerError(logger, fmt.Errorf("ошибка маршалинга: %w", err), http.StatusInternalServerError)
		utils.SendError(w, "Ошибка сервера", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
	log.LogHandlerInfo(logger, "Success", http.StatusOK)
}
