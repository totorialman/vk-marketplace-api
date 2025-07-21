package usecase

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/totorialman/vk-marketplace-api/internal/models"
	"github.com/totorialman/vk-marketplace-api/internal/pkg/auth"
	"github.com/totorialman/vk-marketplace-api/internal/pkg/auth/mocks"
	"github.com/golang/mock/gomock"
	"github.com/satori/uuid"
)

func TestSignIn(t *testing.T) {
	salt := make([]byte, 8)
	os.Setenv("JWT_SECRET", "testsecret")

	type args struct {
		data models.UserReq
	}
	tests := []struct {
		name       string
		repoMocker func(*mocks.MockAuthRepo, string, string)
		args       args
		wantErr    error
	}{
		{
			name: "Success",
			repoMocker: func(repo *mocks.MockAuthRepo, login, password string) {
				repo.EXPECT().SelectUserByLogin(gomock.Any(), login).Return(models.User{
					Id:           uuid.NewV4(),
					Login:        login,
					PasswordHash: HashPassword(salt, password),
				}, nil).Times(1)
			},
			args: args{
				data: models.UserReq{
					Login:    "testuser",
					Password: "Pass1234",
				},
			},
			wantErr: nil,
		},
		{
			name:       "Invalid login format",
			repoMocker: func(repo *mocks.MockAuthRepo, _, _ string) {},
			args: args{
				data: models.UserReq{
					Login:    "inv@lid!",
					Password: "Pass1234",
				},
			},
			wantErr: auth.ErrInvalidLogin,
		},
		{
			name: "User not found",
			repoMocker: func(repo *mocks.MockAuthRepo, login, _ string) {
				repo.EXPECT().SelectUserByLogin(gomock.Any(), login).Return(models.User{}, auth.ErrUserNotFound).Times(1)
			},
			args: args{
				data: models.UserReq{
					Login:    "nouser",
					Password: "Pass1234",
				},
			},
			wantErr: auth.ErrUserNotFound,
		},
		{
			name: "Wrong password",
			repoMocker: func(repo *mocks.MockAuthRepo, login, _ string) {
				repo.EXPECT().SelectUserByLogin(gomock.Any(), login).Return(models.User{
					Id:           uuid.NewV4(),
					Login:        login,
					PasswordHash: HashPassword(salt, "Correct1234"),
				}, nil).Times(1)
			},
			args: args{
				data: models.UserReq{
					Login:    "testuser",
					Password: "Wrong1234",
				},
			},
			wantErr: auth.ErrInvalidCredentials, 
		},
		{
			name: "Token generation error (no secret)",
			repoMocker: func(repo *mocks.MockAuthRepo, login, password string) {
				repo.EXPECT().SelectUserByLogin(gomock.Any(), login).Return(models.User{
					Id:           uuid.NewV4(),
					Login:        login,
					PasswordHash: HashPassword(salt, password),
				}, nil).Times(1)
				os.Setenv("JWT_SECRET", "")
			},
			args: args{
				data: models.UserReq{
					Login:    "testuser",
					Password: "Pass1234",
				},
			},
			wantErr: auth.ErrGeneratingToken,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := mocks.NewMockAuthRepo(ctrl)
			uc := CreateAuthUsecase(repo)

			tt.repoMocker(repo, tt.args.data.Login, tt.args.data.Password)

			user, err := uc.SignIn(context.Background(), tt.args.data)

			if !errors.Is(err, tt.wantErr) {
				t.Errorf("SignIn() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err == nil {
				if user.MarketplaceJWT == "" {
					t.Errorf("Expected non-empty JWT token, got: '%s'", user.MarketplaceJWT)
				}
			}
		})
	}
}

func TestSignUp(t *testing.T) {
	os.Setenv("JWT_SECRET", "testsecret")

	type args struct {
		data models.UserReq
	}
	tests := []struct {
		name       string
		args       args
		repoMocker func(*mocks.MockAuthRepo, models.User)
		wantErr    error
	}{
		{
			name: "Success",
			args: args{
				data: models.UserReq{
					Login:    "newuser",
					Password: "Password123",
				},
			},
			repoMocker: func(repo *mocks.MockAuthRepo, user models.User) {
				repo.EXPECT().CreateUser(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			},
			wantErr: nil,
		},
		{
			name: "Invalid login",
			args: args{
				data: models.UserReq{
					Login:    "inv@lid",
					Password: "Password123",
				},
			},
			repoMocker: func(repo *mocks.MockAuthRepo, user models.User) {},
			wantErr:    auth.ErrInvalidLogin,
		},
		{
			name: "Invalid password",
			args: args{
				data: models.UserReq{
					Login:    "validlogin",
					Password: "nopunct",
				},
			},
			repoMocker: func(repo *mocks.MockAuthRepo, user models.User) {},
			wantErr:    auth.ErrInvalidPassword,
		},
		{
			name: "Create user error",
			args: args{
				data: models.UserReq{
					Login:    "newuser",
					Password: "Password123",
				},
			},
			repoMocker: func(repo *mocks.MockAuthRepo, user models.User) {
				repo.EXPECT().CreateUser(gomock.Any(), gomock.Any()).Return(auth.ErrCreatingUser).Times(1)
			},
			wantErr: auth.ErrCreatingUser,
		},
		{
			name: "Token generation failure (no secret)",
			args: args{
				data: models.UserReq{
					Login:    "newuser",
					Password: "Password123",
				},
			},
			repoMocker: func(repo *mocks.MockAuthRepo, user models.User) {
				os.Setenv("JWT_SECRET", "")
				repo.EXPECT().CreateUser(gomock.Any(), gomock.Any()).Return(nil).Times(1)
			},
			wantErr: auth.ErrGeneratingToken,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := mocks.NewMockAuthRepo(ctrl)
			uc := CreateAuthUsecase(repo)

			testUser := models.User{
				Login: tt.args.data.Login,
			}

			tt.repoMocker(repo, testUser)

			user, err := uc.SignUp(context.Background(), tt.args.data)

			if !errors.Is(err, tt.wantErr) {
				t.Errorf("SignUp() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err == nil {
				if user.MarketplaceJWT == "" {
					t.Errorf("Expected non-empty JWT token, got: '%s'", user.MarketplaceJWT)
				}
			}
		})
	}
}