package usecase

import (
	"context"
	"log/slog"

	"github.com/totorialman/vk-marketplace-api/internal/models"
	"github.com/totorialman/vk-marketplace-api/internal/pkg/product"
	"github.com/totorialman/vk-marketplace-api/internal/pkg/utils/log"
)

type ProductUsecase struct {
	repo product.ProductRepo
}

func CreateProductUsecase(repo product.ProductRepo) *ProductUsecase {
	return &ProductUsecase{repo: repo}
}

func (u *ProductUsecase) Create(ctx context.Context, product models.ProductReq) (models.Product, error) {
	logger := log.GetLoggerFromContext(ctx).With(slog.String("func", log.GetFuncName()))

	logger.Info("Successful")
	return u.repo.Create(ctx, product)
}

func (u *ProductUsecase) List(ctx context.Context, page int, limit int, sortBy string, sortDir string, minPrice float64, maxPrice float64) ([]models.Product, error) {
	logger := log.GetLoggerFromContext(ctx).With(slog.String("func", log.GetFuncName()))

	logger.Info("Successful")
	return u.repo.List(ctx, page, limit, sortBy, sortDir, minPrice, maxPrice)
}
