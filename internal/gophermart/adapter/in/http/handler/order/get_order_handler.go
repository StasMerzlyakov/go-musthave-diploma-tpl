package order

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/StasMerzlyakov/gophermart/internal/gophermart/domain"
)

//go:generate mockgen -destination "../mocks/$GOFILE" -package mocks . GetOrderApp
type GetOrderApp interface {
	All(ctx context.Context) ([]domain.OrderData, error)
}

// GET /api/user/orders
func GetHandler(app GetOrderApp) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {

		handlerName := "order.GetHandler"

		logger, err := domain.GetCtxLogger(req.Context())
		if err != nil {
			fmt.Printf("%v:  can't extract logger\n", handlerName)
			http.Error(w, "can't extract logger", http.StatusInternalServerError)
			return
		}

		_, err = io.ReadAll(req.Body)
		defer req.Body.Close()

		if err != nil {
			logger.Errorw(handlerName, "err", "can't read body")
			http.Error(w, "can't read body", http.StatusBadRequest)
			return
		}

		orderData, err := app.All(req.Context())
		if err != nil {
			logger.Errorw(handlerName, "err", err.Error())
			http.Error(w, err.Error(), domain.MapDomainErrorToHTTPStatusErr(err))
			return
		}

		w.Header().Set("Content-Type", domain.ApplicationJSON)

		if err := json.NewEncoder(w).Encode(orderData); err != nil {
			logger.Errorw(handlerName, "err", fmt.Sprintf("json encode error: %v", err.Error()))
			http.Error(w, err.Error(), domain.MapDomainErrorToHTTPStatusErr(err))
			return
		}
	}
}
