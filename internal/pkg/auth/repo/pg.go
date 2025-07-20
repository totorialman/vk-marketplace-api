package repo

import (
	"context"
	"log/slog"

	"github.com/jackc/pgtype/pgxtype"
	"github.com/totorialman/vk-marketplace-api/internal/models"
	"github.com/totorialman/vk-marketplace-api/internal/pkg/utils/log"
)

const (
	insertUser        = "INSERT INTO users (id, login, password_hash) VALUES ($1, $2, $3)"
	selectUserByLogin = "SELECT id, password_hash FROM users WHERE login = $1"
)

type AuthRepo struct {
	db pgxtype.Querier
}

func CreateAuthRepo(db pgxtype.Querier) *AuthRepo {
	return &AuthRepo{
		db: db,
	}
}

func (repo *AuthRepo) CreateUser(ctx context.Context, user models.User) error {
	logger := log.GetLoggerFromContext(ctx).With(slog.String("func", log.GetFuncName()))

	_, err := repo.db.Exec(ctx, insertUser, user.Id, user.Login, user.PasswordHash)
	if err != nil {
		logger.Error(err.Error())
		return err
	}
	logger.Info("Successful")
	return nil
}

func (repo *AuthRepo) SelectUserByLogin(ctx context.Context, login string) (models.User, error) {
	logger := log.GetLoggerFromContext(ctx).With(slog.String("func", log.GetFuncName()))

	resultUser := models.User{Login: login}
	err := repo.db.QueryRow(ctx, selectUserByLogin, login).Scan(
		&resultUser.Id,
		&resultUser.PasswordHash,
	)

	if err != nil {
		logger.Error(err.Error())
		return models.User{}, err
	}

	logger.Info("Successful")
	return resultUser, nil
}
