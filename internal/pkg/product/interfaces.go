package product

import (
	"context"

	"github.com/totorialman/vk-marketplace-api/internal/models"
)

type ProductUsecase interface {
	Create(ctx context.Context, product models.ProductReq) (models.Product, error)
	List(ctx context.Context, page, limit int, sortBy, sortDir string, minPrice, maxPrice float64) ([]models.Product, error)
}

type ProductRepo interface {
	Create(ctx context.Context, product models.ProductReq) (models.Product, error)
	List(ctx context.Context, page, limit int, sortBy, sortDir string, minPrice, maxPrice float64) ([]models.Product, error)
}

type ProductFilter struct {
	Page     int
	Limit    int
	SortBy   string
	SortDir  string
	MinPrice float64
	MaxPrice float64
}
