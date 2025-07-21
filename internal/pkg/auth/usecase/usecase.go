package usecase

import (
	"bytes"
	"context"
	"crypto/rand"
	"log/slog"
	"os"
	"strings"
	"time"
	"unicode"

	"github.com/golang-jwt/jwt"
	"github.com/satori/uuid"
	"github.com/totorialman/vk-marketplace-api/internal/models"
	"github.com/totorialman/vk-marketplace-api/internal/pkg/auth"
	"github.com/totorialman/vk-marketplace-api/internal/pkg/utils/log"
	"golang.org/x/crypto/argon2"
)

func HashPassword(salt []byte, plainPassword string) []byte {
	hashedPass := argon2.IDKey([]byte(plainPassword), salt, 1, 64*1024, 4, 32)
	return append(salt, hashedPass...)
}

func checkPassword(passHash []byte, plainPassword string) bool {
	salt := make([]byte, 8)
	copy(salt, passHash[:8])
	userPassHash := HashPassword(salt, plainPassword)
	return bytes.Equal(userPassHash, passHash)
}

const (
	maxLoginLength = 20
	minLoginLength = 3
	minPassLength  = 8
	maxPassLength  = 100
)

const allowedChars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_-"

func validLogin(login string) bool {
	if len(login) < minLoginLength || len(login) > maxLoginLength {
		return false
	}
	for _, char := range login {
		if !strings.Contains(allowedChars, string(char)) {
			return false
		}
	}
	return true
}

func validPassword(password string) bool {
	var up, low, digit bool

	if len(password) < minPassLength || len(password) > maxPassLength {
		return false
	}

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			up = true
		case unicode.IsLower(char):
			low = true
		case unicode.IsDigit(char):
			digit = true
		}
	}

	return up && low && digit
}

func generateToken(login string, id uuid.UUID) (string, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return "", auth.ErrGeneratingToken
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"login": login,
		"id":    id,
		"exp":   time.Now().Add(24 * time.Hour).Unix(),
	})

	return token.SignedString([]byte(secret))
}

type AuthUsecase struct {
	repo auth.AuthRepo
}

func CreateAuthUsecase(repo auth.AuthRepo) *AuthUsecase {
	return &AuthUsecase{repo: repo}
}

func (uc *AuthUsecase) SignIn(ctx context.Context, data models.UserReq) (models.User, error) {
	logger := log.GetLoggerFromContext(ctx).With(slog.String("func", log.GetFuncName()))

	if !validLogin(data.Login) {
		logger.Error(auth.ErrInvalidLogin.Error())
		return models.User{}, auth.ErrInvalidLogin
	}

	user, err := uc.repo.SelectUserByLogin(ctx, data.Login)
	if err != nil {
		logger.Error(auth.ErrUserNotFound.Error())
		return models.User{}, auth.ErrUserNotFound
	}

	if !checkPassword(user.PasswordHash, data.Password) {
		logger.Error(auth.ErrInvalidCredentials.Error())
		return models.User{}, auth.ErrInvalidCredentials
	}

	token, err := generateToken(user.Login, user.Id)
	if err != nil {
		logger.Error(auth.ErrGeneratingToken.Error())
		return models.User{}, auth.ErrGeneratingToken
	}

	dataResp := models.User{
		Id:             user.Id,
		Login:          user.Login,
		MarketplaceJWT: token,
	}
	dataResp.Sanitize()

	logger.Info("Successful")
	return dataResp, nil
}

func (uc *AuthUsecase) SignUp(ctx context.Context, data models.UserReq) (models.User, error) {
	logger := log.GetLoggerFromContext(ctx).With(slog.String("func", log.GetFuncName()))

	if !validLogin(data.Login) {
		logger.Error(auth.ErrInvalidLogin.Error())
		return models.User{}, auth.ErrInvalidLogin
	}

	if !validPassword(data.Password) {
		logger.Error(auth.ErrInvalidPassword.Error())
		return models.User{}, auth.ErrInvalidPassword
	}

	salt := make([]byte, 8)
	rand.Read(salt)
	hashedPassword := HashPassword(salt, data.Password)

	newUser := models.User{
		Id:           uuid.NewV4(),
		Login:        data.Login,
		PasswordHash: hashedPassword,
	}
	newUser.Sanitize()

	if err := uc.repo.CreateUser(ctx, newUser); err != nil {
		logger.Error(err.Error())
		return models.User{}, auth.ErrCreatingUser
	}

	token, err := generateToken(newUser.Login, newUser.Id)
	if err != nil {
		logger.Error(auth.ErrGeneratingToken.Error())
		return models.User{}, auth.ErrGeneratingToken
	}

	dataResp := models.User{
		Id:             newUser.Id,
		Login:          data.Login,
		MarketplaceJWT: token,
	}
	dataResp.Sanitize()

	logger.Info("Successful")
	return dataResp, nil
}
