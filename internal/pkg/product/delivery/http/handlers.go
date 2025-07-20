package http

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"path"
	"strconv"
	"strings"

	"github.com/satori/uuid"
	"github.com/totorialman/vk-marketplace-api/internal/models"
	"github.com/totorialman/vk-marketplace-api/internal/pkg/product"
	"github.com/totorialman/vk-marketplace-api/internal/pkg/utils/log"
	utils "github.com/totorialman/vk-marketplace-api/internal/pkg/utils/send_error"
)

type ProductHandler struct {
	uc product.ProductUsecase
}

func CreateProductHandler(uc product.ProductUsecase) *ProductHandler {
	return &ProductHandler{uc: uc}
}

const maxImageSize = 5 * 1024 * 1024 

func isValidImageURL(url string) bool {
	ext := strings.ToLower(path.Ext(url))
	switch ext {
	case ".jpg", ".jpeg", ".png", ".webp":
		return true
	default:
		return false
	}
}

func validateImage(url string) error {
	if !isValidImageURL(url) {
		return fmt.Errorf("недопустимый формат изображения (разрешены: .jpg, .jpeg, .png, .webp)")
	}

	resp, err := http.Head(url)
	if err != nil {
		return fmt.Errorf("не удалось получить заголовки изображения: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("изображение недоступно, статус: %d", resp.StatusCode)
	}

	if sizeStr := resp.Header.Get("Content-Length"); sizeStr != "" {
		size, err := strconv.Atoi(sizeStr)
		if err == nil && size > maxImageSize {
			return fmt.Errorf("размер изображения превышает %d МБ", maxImageSize/1024/1024)
		}
	}

	contentType := resp.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "image/") {
		return fmt.Errorf("контент не является изображением, а %s", contentType)
	}

	return nil
}

func (h *ProductHandler) Create(w http.ResponseWriter, r *http.Request) {
	logger := log.GetLoggerFromContext(r.Context()).With(slog.String("func", log.GetFuncName()))

	userID, ok := r.Context().Value("user_id").(uuid.UUID)
	if !ok {
		log.LogHandlerError(logger, fmt.Errorf("только авторизованные пользователи могут размещать объявления"), http.StatusUnauthorized)
		utils.SendError(w, "только авторизованные пользователи могут размещать объявления", http.StatusUnauthorized)
		return
	}

	var req models.Product
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.LogHandlerError(logger, fmt.Errorf("ошибка парсинга JSON: %w", err), http.StatusBadRequest)
		utils.SendError(w, "Ошибка парсинга JSON", http.StatusBadRequest)
		return
	}
	req.Sanitize()

	if len(req.Title) < 5 || len(req.Title) > 255 {
		utils.SendError(w, "заголовок должен быть от 5 до 255 символов", http.StatusBadRequest)
		return
	}

	if len(req.Description) < 10 || len(req.Description) > 1000 {
		utils.SendError(w, "описание должно быть от 10 до 1000 символов", http.StatusBadRequest)
		return
	}

	if req.Price <= 0 {
		utils.SendError(w, "цена должна быть больше 0", http.StatusBadRequest)
		return
	}

	if err := validateImage(req.ImageURL); err != nil {
		utils.SendError(w, "неверное изображение: "+err.Error(), http.StatusBadRequest)
		return
	}

	req.UserID = userID

	product, err := h.uc.Create(r.Context(), &req)
	if err != nil {
		log.LogHandlerError(logger, err, http.StatusInternalServerError)
		utils.SendError(w, "не удалось создать объявление", http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(product)
	if err != nil {
		log.LogHandlerError(logger, fmt.Errorf("ошибка маршалинга: %w", err), http.StatusInternalServerError)
		utils.SendError(w, "Ошибка сервера", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
	log.LogHandlerInfo(logger, "Success", http.StatusOK)
}

func (h *ProductHandler) List(w http.ResponseWriter, r *http.Request) {
	logger := log.GetLoggerFromContext(r.Context()).With(slog.String("func", log.GetFuncName()))

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	sortBy := r.URL.Query().Get("sort_by")
	sortDir := r.URL.Query().Get("sort_dir")
	minPriceStr := r.URL.Query().Get("min_price")
	maxPriceStr := r.URL.Query().Get("max_price")

	minPrice := 0.0
	maxPrice := 1e20

	if minPriceStr != "" {
		minPrice, _ = strconv.ParseFloat(minPriceStr, 64)
	}
	if maxPriceStr != "" {
		maxPrice, _ = strconv.ParseFloat(maxPriceStr, 64)
	}
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	if sortBy == "" {
		sortBy = "created_at"
	}
	if sortDir == "" {
		sortDir = "desc"
	}

	products, err := h.uc.List(r.Context(), page, limit, sortBy, sortDir, minPrice, maxPrice)
	if err != nil {
		log.LogHandlerError(logger, err, http.StatusInternalServerError)
		utils.SendError(w, "не удалось получить список объявлений", http.StatusInternalServerError)
		return
	}

	userID, authorized := r.Context().Value("user_id").(uuid.UUID)

	if authorized {
		for i := range products {
			products[i].IsMine = (products[i].UserID == userID)
		}
	}

	data, err := json.Marshal(products)
	if err != nil {
		log.LogHandlerError(logger, fmt.Errorf("ошибка маршалинга: %w", err), http.StatusInternalServerError)
		utils.SendError(w, "Ошибка сервера", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
	log.LogHandlerInfo(logger, "Success", http.StatusOK)
}
