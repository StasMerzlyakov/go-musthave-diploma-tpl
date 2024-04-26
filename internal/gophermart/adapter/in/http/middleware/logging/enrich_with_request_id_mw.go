package logging

import (
	"net/http"

	"github.com/StasMerzlyakov/gophermart/internal/gophermart/adapter/in/http/middleware"
	"github.com/StasMerzlyakov/gophermart/internal/gophermart/domain"
	"github.com/google/uuid"
)

// Добавляет к запросу RequestID и устанавливает в контекст логгер
func EncrichWithRequestIDMW() middleware.Middleware {
	return func(next http.Handler) http.Handler {
		logReqFn := func(w http.ResponseWriter, req *http.Request) {
			logger := domain.GetMainLogger()
			requestUUID := uuid.New()
			ctx := domain.EnrichWithRequestIDLogger(req.Context(), requestUUID, logger)
			req = req.WithContext(ctx)
			next.ServeHTTP(w, req)
		}
		return http.HandlerFunc(logReqFn)
	}
}
