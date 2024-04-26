package balance_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

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

func TestGetWithdrawals1(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mLog := dmocks.NewMockLogger(ctrl)
	mLog.EXPECT().Infow(gomock.Any(), gomock.Any()).AnyTimes()

	domain.SetMainLogger(mLog)
	erMW := logging.EncrichWithRequestIDMW()

	app := mocks.NewMockGetWithdrawalsApp(ctrl)

	tm1, err := time.Parse(time.RFC3339, "2020-12-09T16:09:57+03:00")
	require.NoError(t, err)

	app.EXPECT().Withdrawals(gomock.Any()).
		DoAndReturn(func(ctx context.Context) ([]domain.WithdrawalData, error) {

			return []domain.WithdrawalData{
				{
					Order:       domain.OrderNumber("2377225624"),
					Sum:         500,
					ProcessedAt: domain.RFC3339Time(tm1),
				},
			}, nil
		})

	mux := http.NewServeMux()
	handler := balance.GetWithdrawals(app)
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

	require.JSONEqf(t, string(resp.Body()), `[
		{
			"order": "2377225624",
			"sum": 500,
			"processed_at": "2020-12-09T16:09:57+03:00"
		}
	]`, "unexpected content")
}
