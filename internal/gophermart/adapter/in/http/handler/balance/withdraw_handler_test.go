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

func TestWithdraw1(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mLog := dmocks.NewMockLogger(ctrl)
	mLog.EXPECT().Infow(gomock.Any(), gomock.Any()).AnyTimes()

	domain.SetMainLogger(mLog)
	erMW := logging.EncrichWithRequestIDMW

	app := mocks.NewMockWithdrawApp(ctrl)

	app.EXPECT().Withdraw(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, withdraw *domain.WithdrawData) error {
			require.NotNil(t, withdraw)
			require.Equal(t, domain.OrderNumber("2377225624"), withdraw.Order)
			require.Equal(t, float64(751), withdraw.Sum)
			return nil
		})

	mux := http.NewServeMux()
	handler := balance.WithdrawHandler(app)
	path := "/withdraw"
	mux.Handle(path, middleware.Conveyor(handler, erMW))

	srv := httptest.NewServer(mux)
	defer srv.Close()

	req := resty.New().R().
		SetHeader("Content-Type", domain.ApplicationJSON)
	req.Method = http.MethodPost
	req.URL = srv.URL + path

	req.SetBody(`{
		"order": "2377225624",
		"sum": 751
	} `)

	resp, err := req.Send()
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode())
}
