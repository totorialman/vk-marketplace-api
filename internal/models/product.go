package models

import (
	"html"
	"time"

	"github.com/satori/uuid"
)

// easyjson:json
type Product struct {
	Id          uuid.UUID `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	ImageURL    string    `json:"image_url"`
	Price       float64   `json:"price"`
	UserID      uuid.UUID `json:"user_id"`
	CreatedAt   time.Time `json:"created_at"`
	IsMine      bool      `json:"is_mine,omitempty"`
}

func (a *Product) Sanitize() {
	a.Title = html.EscapeString(a.Title)
	a.Description = html.EscapeString(a.Description)
}
