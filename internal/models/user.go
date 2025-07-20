package models

import (
	"html"

	"github.com/satori/uuid"
)

// easyjson:json
type User struct {
	Id             uuid.UUID `json:"id"`
	Login          string    `json:"login"`
	PasswordHash   []byte    `json:"-"`
	MarketplaceJWT string    `json:"MarketplaceJWT"`
}

// easyjson:json
type UserReq struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

func (u *UserReq) Sanitize() {
	u.Login = html.EscapeString(u.Login)
	u.Password = html.EscapeString(u.Password)
}

func (u *User) Sanitize() {
	u.Login = html.EscapeString(u.Login)
	u.MarketplaceJWT = html.EscapeString(u.MarketplaceJWT)
}

type ErrorResponse struct {
    Error string `json:"error"`
}
