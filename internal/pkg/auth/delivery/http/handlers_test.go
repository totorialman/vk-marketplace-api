package http

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/totorialman/vk-marketplace-api/internal/models"
	"github.com/totorialman/vk-marketplace-api/internal/pkg/auth/mocks"
	"github.com/golang/mock/gomock"
	"github.com/satori/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/totorialman/vk-marketplace-api/internal/pkg/auth"
)

func TestSignUp(t *testing.T) {
	type args struct {
		login    string
		password string
	}

	var tests = []struct {
		name           string
		requestBody    string
		args           args
		ucErr          error
		userIdInCtx    bool
		expectedStatus int
	}{
		{
			name:        "Success",
			requestBody: `{"login":"test123","password":"Pass@123"}`,
			args: args{
				login:    "test123",
				password: "Pass@123",
			},
			ucErr:          nil,
			userIdInCtx:    false,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid JSON",
			requestBody:    `{"login":"testuser","password":"abc123"`,
			userIdInCtx:    false,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:        "Invalid Login or Password",
			requestBody: `{"login":"testuser","password":"wrongpass"}`,
			args: args{
				login:    "testuser",
				password: "wrongpass",
			},
			ucErr:          auth.ErrInvalidLogin,
			userIdInCtx:    false,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:        "User Creation Failed",
			requestBody: `{"login":"newuser","password":"validPass123"}`,
			args: args{
				login:    "newuser",
				password: "validPass123",
			},
			ucErr:          auth.ErrCreatingUser,
			userIdInCtx:    false,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:        "Unknown Error",
			requestBody: `{"login":"testuser","password":"errorpass"}`,
			args: args{
				login:    "testuser",
				password: "errorpass",
			},
			ucErr:          errors.New("unknown error"),
			userIdInCtx:    false,
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:        "Already Authorized",
			requestBody: `{"login":"testuser","password":"Pass@123"}`,
			args: args{
				login:    "testuser",
				password: "Pass@123",
			},
			userIdInCtx:    true,
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUsecase := mocks.NewMockAuthUsecase(ctrl)

			if tt.name != "Invalid JSON" && !tt.userIdInCtx {
				mockUsecase.EXPECT().SignUp(gomock.Any(), models.UserReq{
					Login:    tt.args.login,
					Password: tt.args.password,
				}).Return(models.User{
					Login:          tt.args.login,
					PasswordHash:   []byte("hashed_password"),
					Id:             uuid.NewV4(),
					MarketplaceJWT: "jwt_token",
				}, tt.ucErr)
			}

			r := httptest.NewRequest("POST", "/api/auth/signup", bytes.NewBufferString(tt.requestBody))
			if tt.userIdInCtx {
				r = r.WithContext(context.WithValue(r.Context(), "user_id", uuid.NewV4()))
			}
			w := httptest.NewRecorder()

			handler := CreateAuthHandler(mockUsecase)
			handler.SignUp(w, r)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestSignIn(t *testing.T) {
	type args struct {
		login    string
		password string
	}

	var tests = []struct {
		name           string
		requestBody    string
		args           args
		ucErr          error
		expectedStatus int
	}{
		{
			name:           "Success",
			requestBody:    `{"login":"test123","password":"Pass@123"}`,
			args: args{
				login:    "test123",
				password: "Pass@123",
			},
			ucErr:          nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid JSON",
			requestBody:    `{"login":"testuser","password":"abc123"`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "User Not Found",
			requestBody:    `{"login":"unknown","password":"somepass"}`,
			args: args{
				login:    "unknown",
				password: "somepass",
			},
			ucErr:          auth.ErrUserNotFound,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Invalid Credentials",
			requestBody:    `{"login":"testuser","password":"wrongpass"}`,
			args: args{
				login:    "testuser",
				password: "wrongpass",
			},
			ucErr:          auth.ErrInvalidCredentials,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Unknown Error",
			requestBody:    `{"login":"testuser","password":"errorpass"}`,
			args: args{
				login:    "testuser",
				password: "errorpass",
			},
			ucErr:          errors.New("unknown error"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUsecase := mocks.NewMockAuthUsecase(ctrl)

			if tt.name != "Invalid JSON" {
				mockUsecase.EXPECT().SignIn(gomock.Any(), models.UserReq{
					Login:    tt.args.login,
					Password: tt.args.password,
				}).Return(models.User{
					Login:          tt.args.login,
					PasswordHash:   []byte("hashed_password"),
					Id:             uuid.NewV4(),
					MarketplaceJWT: "jwt_token",
				}, tt.ucErr)
			}

			r := httptest.NewRequest("POST", "/api/auth/signin", bytes.NewBufferString(tt.requestBody))
			w := httptest.NewRecorder()

			handler := CreateAuthHandler(mockUsecase)
			handler.SignIn(w, r)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}
