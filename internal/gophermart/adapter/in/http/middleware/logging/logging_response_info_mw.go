package logging

import (
	"fmt"
	"net/http"

	"github.com/StasMerzlyakov/gophermart/internal/gophermart/adapter/in/http/middleware"
	"github.com/StasMerzlyakov/gophermart/internal/gophermart/domain"
)

type responseData struct {
	status int
	size   int
}

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCodeFixed bool
	responseData    *responseData
}

var _ http.ResponseWriter = (*loggingResponseWriter)(nil)

func (lw *loggingResponseWriter) Header() http.Header {
	return lw.ResponseWriter.Header()
}

func (lw *loggingResponseWriter) Write(data []byte) (int, error) {
	if !lw.statusCodeFixed {
		lw.statusCodeFixed = true
	}
	size, err := lw.ResponseWriter.Write(data)
	lw.responseData.size += size
	return size, err
}

func (lw *loggingResponseWriter) WriteHeader(statusCode int) {
	if !lw.statusCodeFixed {
		lw.responseData.status = statusCode
		lw.statusCodeFixed = true
		lw.ResponseWriter.WriteHeader(statusCode)
	}
}

func NewLoggingResponseMW() middleware.Middleware {
	return func(next http.Handler) http.Handler {
		lrw := func(w http.ResponseWriter, req *http.Request) {

			log, err := domain.GetCtxLogger(req.Context())
			if err != nil {
				fmt.Printf("LoggingResponseMW - can't extract logger\n")
				http.Error(w, "logging configuration error", http.StatusInternalServerError)
			}

			lw := &loggingResponseWriter{
				responseData: &responseData{
					status: http.StatusOK,
					size:   0,
				},
				statusCodeFixed: false,
				ResponseWriter:  w,
			}

			next.ServeHTTP(lw, req)
			log.Infow("requestResult", "statusCode", lw.responseData.status, "size", lw.responseData.size)
		}
		return http.HandlerFunc(lrw)
	}
}
