package order_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

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

func TestCreateHandler1(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mLog := dmocks.NewMockLogger(ctrl)
	mLog.EXPECT().Infow(gomock.Any(), gomock.Any()).AnyTimes()

	domain.SetMainLogger(mLog)
	erMW := logging.EncrichWithRequestIDMW

	app := mocks.NewMockCreateOrderApp(ctrl)

	app.EXPECT().New(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, number domain.OrderNumber) error {
			require.Equal(t, domain.OrderNumber("1232"), number)
			return nil
		})

	mux := http.NewServeMux()
	handler := order.CreateHandler(app)
	path := "/orders"
	mux.Handle(path, middleware.Conveyor(handler, erMW))

	srv := httptest.NewServer(mux)
	defer srv.Close()

	req := resty.New().R().
		SetHeader("Content-Type", domain.TextPlain)
	req.Method = http.MethodPost
	req.URL = srv.URL + path

	req.SetBody("1232")

	resp, err := req.Send()
	require.Nil(t, err)
	require.Equal(t, http.StatusAccepted, resp.StatusCode())
}
