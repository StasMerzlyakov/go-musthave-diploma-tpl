package logging

import (
	"fmt"
	"net/http"
	"time"

	"github.com/StasMerzlyakov/gophermart/internal/gophermart/adapter/http/middleware"
	"github.com/StasMerzlyakov/gophermart/internal/gophermart/domain"
)

func NewLoggingRequestMW() middleware.Middleware {
	return func(next http.Handler) http.Handler {
		logReqFn := func(w http.ResponseWriter, req *http.Request) {

			log, err := domain.GetCtxLogger(req.Context())
			if err != nil {
				fmt.Printf("LoggingRequestMW - can't extract logger\n")
				http.Error(w, "logging configuration error", http.StatusInternalServerError)
			}

			start := time.Now()
			uri := req.RequestURI
			method := req.Method

			next.ServeHTTP(w, req)

			duration := time.Since(start)

			log.Infow("requestStatus",
				"uri", uri,
				"method", method,
				"duration", duration,
			)
		}
		return http.HandlerFunc(logReqFn)
	}
}
