package auth

import (
	"context"
	"errors"

	"github.com/totorialman/vk-marketplace-api/internal/models"
)

var (
	ErrCreatingUser       = errors.New("ошибка в создании пользователя")
	ErrUserNotFound       = errors.New("пользователь не найден")
	ErrInvalidLogin       = errors.New("неверный формат логина")
	ErrInvalidPassword    = errors.New("неверный формат пароля")
	ErrInvalidCredentials = errors.New("неверный логин или пароль")
	ErrGeneratingToken    = errors.New("ошибка генерации токена")
)

type AuthUsecase interface {
	SignUp(ctx context.Context, data models.UserReq) (models.User, error)
	SignIn(ctx context.Context, data models.UserReq) (models.User, error)
}

type AuthRepo interface {
	CreateUser(ctx context.Context, user models.User) error
	SelectUserByLogin(ctx context.Context, login string) (models.User, error)
}
