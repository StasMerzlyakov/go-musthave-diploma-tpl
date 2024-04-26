package order_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/StasMerzlyakov/gophermart/internal/gophermart/adapter/in/http/handler/mocks"
	"github.com/StasMerzlyakov/gophermart/internal/gophermart/adapter/in/http/handler/order"
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

	app := mocks.NewMockGetOrderApp(ctrl)

	tm1, err := time.Parse(time.RFC3339, "2020-12-10T15:15:45+03:00")
	require.NoError(t, err)

	tm2, err := time.Parse(time.RFC3339, "2020-12-10T15:12:01+03:00")
	require.NoError(t, err)

	tm3, err := time.Parse(time.RFC3339, "2020-12-09T16:09:53+03:00")
	require.NoError(t, err)

	app.EXPECT().All(gomock.Any()).
		DoAndReturn(func(ctx context.Context) ([]domain.OrderData, error) {

			return []domain.OrderData{
				{
					Number:     "9278923470",
					Status:     "PROCESSED",
					Accrual:    domain.Float64Ptr(500),
					UploadedAt: *domain.TimePtr(tm1),
				},
				{
					Number:     "12345678903",
					Status:     "PROCESSING",
					UploadedAt: *domain.TimePtr(tm2),
				},
				{
					Number:     "346436439",
					Status:     "INVALID",
					UploadedAt: *domain.TimePtr(tm3),
				},
			}, nil
		})

	mux := http.NewServeMux()
	handler := order.GetHandler(app)
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
			"number": "9278923470",
			"status": "PROCESSED",
			"accrual": 500,
			"uploaded_at": "2020-12-10T15:15:45+03:00"
		},
		{
			"number": "12345678903",
			"status": "PROCESSING",
			"uploaded_at": "2020-12-10T15:12:01+03:00"
		},
		{
			"number": "346436439",
			"status": "INVALID",
			"uploaded_at": "2020-12-09T16:09:53+03:00"
		}
	]`, "unexpected content")

}
