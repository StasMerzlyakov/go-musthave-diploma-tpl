package auth_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/StasMerzlyakov/gophermart/internal/gophermart/adapter/http/handler/auth"
	amock "github.com/StasMerzlyakov/gophermart/internal/gophermart/adapter/http/handler/mocks"
	"github.com/StasMerzlyakov/gophermart/internal/gophermart/adapter/http/middleware"
	"github.com/StasMerzlyakov/gophermart/internal/gophermart/adapter/http/middleware/logging"
	"github.com/StasMerzlyakov/gophermart/internal/gophermart/domain"
	"github.com/StasMerzlyakov/gophermart/internal/gophermart/domain/mocks"
	"github.com/go-resty/resty/v2"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestRegisterHandler1(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mLog := mocks.NewMockLogger(ctrl)
	mLog.EXPECT().Infow(gomock.Any(), gomock.Any()).AnyTimes()

	domain.SetMainLogger(mLog)
	erMW := logging.EncrichWithRequestIDMW

	authApp := amock.NewMockRegisterApp(ctrl)

	authApp.EXPECT().Register(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, regData *domain.RegistrationData) (domain.TokenString, error) {
			require.Equal(t, "login", regData.Login)
			require.Equal(t, "pass", regData.Password)
			return "JWT", nil
		})

	mux := http.NewServeMux()
	registerHandler := auth.RegisterHandler(authApp)
	mux.Handle("/register", middleware.Conveyor(registerHandler, erMW))

	srv := httptest.NewServer(mux)
	defer srv.Close()

	req := resty.New().R().
		SetHeader("Content-Type", domain.ApplicationJSON)
	req.Method = http.MethodPost
	req.URL = srv.URL + "/register"

	req.SetBody(domain.RegistrationData{
		Login:    "login",
		Password: "pass",
	})

	resp, err := req.Send()
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode())

	authHeader := resp.Header().Get("Authorization")
	require.Equal(t, "Bearer JWT", authHeader)
}
