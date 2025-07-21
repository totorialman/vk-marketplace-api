package repo

import (
	"context"
	"errors"
	"testing"

	"github.com/driftprogramming/pgxpoolmock"
	"github.com/golang/mock/gomock"
	"github.com/satori/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/totorialman/vk-marketplace-api/internal/models"
)

func TestCreateUser(t *testing.T) {
	userId := uuid.NewV4()
	testUser := models.User{
		Login:        "test_user",
		PasswordHash: []byte("hashed_password"),
		Id:           userId,
	}

	tests := []struct {
		name           string
		repoMocker     func(*pgxpoolmock.MockPgxPool)
		expectedLogger string
		err            error
	}{
		{
			name: "Success",
			repoMocker: func(mockPool *pgxpoolmock.MockPgxPool) {
				mockPool.EXPECT().Exec(gomock.Any(), insertUser,
					testUser.Id,
					testUser.Login,
					testUser.PasswordHash,
				).Return(nil, nil)
			},
			expectedLogger: "Successful",
			err:            nil,
		},
		{
			name: "DatabaseError",
			repoMocker: func(mockPool *pgxpoolmock.MockPgxPool) {
				mockPool.EXPECT().Exec(gomock.Any(), insertUser,
					testUser.Id,
					testUser.Login,
					testUser.PasswordHash,
				).Return(nil, errors.New("database error"))
			},
			expectedLogger: "database error",
			err:            errors.New("database error"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockPool := pgxpoolmock.NewMockPgxPool(ctrl)
			defer ctrl.Finish()

			test.repoMocker(mockPool)

			repo := CreateAuthRepo(mockPool)
			err := repo.CreateUser(context.Background(), testUser)

			assert.Equal(t, test.err, err)
		})
	}
}

