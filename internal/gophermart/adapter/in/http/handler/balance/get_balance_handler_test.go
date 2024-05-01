package balance_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/StasMerzlyakov/gophermart/internal/gophermart/adapter/in/http/handler/balance"
	"github.com/StasMerzlyakov/gophermart/internal/gophermart/adapter/in/http/handler/mocks"
	"github.com/StasMerzlyakov/gophermart/internal/gophermart/adapter/in/http/middleware"
	"github.com/StasMerzlyakov/gophermart/internal/gophermart/adapter/in/http/middleware/logging"
	"github.com/StasMerzlyakov/gophermart/internal/gophermart/domain"
	dmocks "github.com/StasMerzlyakov/gophermart/internal/gophermart/domain/mocks"
	"github.com/go-resty/resty/v2"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestGetHandler1(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mLog := dmocks.NewMockLogger(ctrl)
	mLog.EXPECT().Infow(gomock.Any(), gomock.Any()).AnyTimes()

	domain.SetMainLogger(mLog)
	erMW := logging.EncrichWithRequestIDMW()

	app := mocks.NewMockGetBalanceApp(ctrl)

	app.EXPECT().Get(gomock.Any()).
		DoAndReturn(func(ctx context.Context) (*domain.Balance, error) {

			return &domain.Balance{
				Current:   500.5,
				Withdrawn: 42,
			}, nil
		})

	mux := http.NewServeMux()
	handler := balance.GetHandler(app)
	path := "/get"
	mux.Handle(path, middleware.Conveyor(handler, erMW))

	srv := httptest.NewServer(mux)
	defer srv.Close()

	req := resty.New().R().
		SetHeader("Content-Type", domain.TextPlain)
	req.Method = http.MethodGet
	req.URL = srv.URL + path

	resp, err := req.Send()
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode())
	contentType := resp.Header().Get("Content-Type")
	require.Equal(t, domain.ApplicationJSON, contentType)

	require.JSONEqf(t, string(resp.Body()), `{
		"current": 500.5,
		"withdrawn": 42
	}`, "unexpected content")
}
