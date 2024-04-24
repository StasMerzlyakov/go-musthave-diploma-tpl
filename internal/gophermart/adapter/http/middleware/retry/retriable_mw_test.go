package retry_test

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/StasMerzlyakov/gophermart/internal/gophermart/domain"
	"github.com/StasMerzlyakov/gophermart/internal/gophermart/domain/mocks"
	"github.com/go-resty/resty/v2"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/StasMerzlyakov/gophermart/internal/gophermart/adapter/http/middleware"
	"github.com/StasMerzlyakov/gophermart/internal/gophermart/adapter/http/middleware/logging"
	mmocks "github.com/StasMerzlyakov/gophermart/internal/gophermart/adapter/http/middleware/mocks"
	"github.com/StasMerzlyakov/gophermart/internal/gophermart/adapter/http/middleware/retry"
)

func TestRetriableMW1(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mLog := mocks.NewMockLogger(ctrl)
	mLog.EXPECT().Infow(gomock.Any(), gomock.Any()).AnyTimes()

	rMW := retry.NewRetriableRequestMWConf(time.Duration(time.Second), time.Duration(2*time.Second), 4)

	domain.SetMainLogger(mLog)
	erMW := logging.EncrichWithRequestIDMW

	mux := http.NewServeMux()
	mux.Handle("/json", middleware.Conveyor(createOkMockHandler(ctrl), rMW, erMW))

	srv := httptest.NewServer(mux)
	defer srv.Close()

	req := resty.New().R()
	req.Method = http.MethodPost

	req.URL = srv.URL + "/json"

	resp, err := req.Send()
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode())
}

func TestRetriableMW2(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mLog := mocks.NewMockLogger(ctrl)
	mLog.EXPECT().Infow(gomock.Any(), gomock.Any()).AnyTimes()

	rMW := retry.NewRetriableRequestMWConf(time.Duration(time.Second), time.Duration(2*time.Second), 4)

	domain.SetMainLogger(mLog)
	erMW := logging.EncrichWithRequestIDMW

	mux := http.NewServeMux()
	mux.Handle("/json", middleware.Conveyor(createAnyErrHandler(ctrl), rMW, erMW))

	srv := httptest.NewServer(mux)
	defer srv.Close()

	req := resty.New().R()
	req.Method = http.MethodPost

	req.URL = srv.URL + "/json"

	resp, err := req.Send()
	require.Nil(t, err)
	require.Equal(t, http.StatusAccepted, resp.StatusCode())
}

func TestRetriableMW3(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mLog := mocks.NewMockLogger(ctrl)
	mLog.EXPECT().Infow(gomock.Any(), gomock.Any()).AnyTimes()

	rMW := retry.NewRetriableRequestMWConf(time.Duration(time.Second), time.Duration(2*time.Second), 4)

	domain.SetMainLogger(mLog)
	erMW := logging.EncrichWithRequestIDMW

	mux := http.NewServeMux()
	mux.Handle("/json", middleware.Conveyor(createAnyErrHandler(ctrl), rMW, erMW))

	srv := httptest.NewServer(mux)
	defer srv.Close()

	req := resty.New().R()
	req.Method = http.MethodPost

	req.URL = srv.URL + "/json"

	resp, err := req.Send()
	require.Nil(t, err)
	require.Equal(t, http.StatusAccepted, resp.StatusCode())
}

func TestRetriableMW4(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mLog := mocks.NewMockLogger(ctrl)
	mLog.EXPECT().Infow(gomock.Any(), gomock.Any()).AnyTimes()

	rMW := retry.NewRetriableRequestMWConf(time.Duration(time.Second), time.Duration(2*time.Second), 4)

	domain.SetMainLogger(mLog)
	erMW := logging.EncrichWithRequestIDMW

	mux := http.NewServeMux()
	mux.Handle("/json", middleware.Conveyor(createInernalErrExceptLastHandler(ctrl), rMW, erMW))

	srv := httptest.NewServer(mux)
	defer srv.Close()

	req := resty.New().R()
	req.Method = http.MethodPost

	req.URL = srv.URL + "/json"

	resp, err := req.Send()
	require.Nil(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode())

	// Проверим что в ответе только данные от посленего вызова
	errHeader := resp.Header().Get("ERR")
	assert.Equal(t, "ERROR4", errHeader)

	respContent := string(resp.Body())
	assert.Equal(t, "test error 4", respContent)
}

func createOkMockHandler(ctrl *gomock.Controller) http.Handler {
	mockHandler := mmocks.NewMockHandler(ctrl)

	mockHandler.EXPECT().ServeHTTP(gomock.Any(), gomock.Any()).DoAndReturn(
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			io.WriteString(w, "{ "+strings.Repeat(`"msg":"Hello, world",`, 19)+`"msg":"Hello, world"`+"}")
		}).Times(1)
	return mockHandler
}

func createAnyErrHandler(ctrl *gomock.Controller) http.Handler {
	mockHandler := mmocks.NewMockHandler(ctrl)

	mockHandler.EXPECT().ServeHTTP(gomock.Any(), gomock.Any()).DoAndReturn(
		func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("test err"))
			http.Error(w, "test err", http.StatusAccepted)
		}).Times(1)
	return mockHandler
}

func createInernalErrExceptLastHandler(ctrl *gomock.Controller) http.Handler {
	mockHandler := mmocks.NewMockHandler(ctrl)

	counter := atomic.Int32{}
	mockHandler.EXPECT().ServeHTTP(gomock.Any(), gomock.Any()).DoAndReturn(
		func(w http.ResponseWriter, r *http.Request) {
			new := counter.Add(1)
			if new == 4 {
				w.WriteHeader(http.StatusCreated)
				w.Header().Add("ERR", fmt.Sprintf("ERROR%d", new))
				w.Write([]byte(fmt.Sprintf("test error %d", new)))
			} else {
				w.WriteHeader(http.StatusInternalServerError)
				w.Header().Add("ERR", fmt.Sprintf("ERROR%d", new))
				w.Write([]byte(fmt.Sprintf("test error %d", new)))
			}
		}).Times(4) // 4 попытки
	return mockHandler
}
