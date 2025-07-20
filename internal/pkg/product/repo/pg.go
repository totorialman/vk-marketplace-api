package repo

import (
	"context"
	"fmt"

	"github.com/jackc/pgtype/pgxtype"
	"github.com/satori/uuid"
	"github.com/totorialman/vk-marketplace-api/internal/models"
)

const (
	createProduct = `INSERT INTO products (id, title, description, image_url, price, user_id)
	 VALUES ($1, $2, $3, $4, $5, $6)
	 RETURNING created_at`

	getProductList = `SELECT id, title, description, image_url, price, user_id, created_at
        FROM products
        WHERE price BETWEEN $1 AND $2
        ORDER BY %s %s
        LIMIT $3 OFFSET $4` // в %s %s подставляю только валидированные через мапу значения
)

type ProductRepo struct {
	db pgxtype.Querier
}

func CreateProductRepo(db pgxtype.Querier) *ProductRepo {
	return &ProductRepo{db: db}
}

func (r *ProductRepo) Create(ctx context.Context, product *models.Product) (*models.Product, error) {
	product.Id = uuid.NewV4()
	product.Sanitize()
	err := r.db.QueryRow(ctx, createProduct,
		product.Id, product.Title, product.Description, product.ImageURL, product.Price, product.UserID,
	).Scan(&product.CreatedAt)

	if err != nil {
		return nil, err
	}

	return product, nil
}

func (r *ProductRepo) List(ctx context.Context, page int, limit int, sortBy string, sortDir string, minPrice float64, maxPrice float64) ([]models.Product, error) {
	offset := (page - 1) * limit

	validSort := map[string]string{
		"created_at": "created_at",
		"price":      "price",
		"":           "created_at",
	}
	validSortDir := map[string]string{
		"asc":  "ASC",
		"desc": "DESC",
		"":     "DESC",
	}

	sortBy = validSort[sortBy]
	sortDir = validSortDir[sortDir]

	query := fmt.Sprintf(getProductList, sortBy, sortDir)

	rows, err := r.db.Query(ctx, query, minPrice, maxPrice, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("ошибка выполнения запроса: %w", err)
	}
	defer rows.Close()

	var products []models.Product
	for rows.Next() {
		var p models.Product
		if err := rows.Scan(&p.Id, &p.Title, &p.Description, &p.ImageURL, &p.Price, &p.UserID, &p.CreatedAt); err != nil {
			return nil, fmt.Errorf("ошибка сканирования строки: %w", err)
		}
		products = append(products, p)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка итерации по строкам: %w", err)
	}

	return products, nil
}
