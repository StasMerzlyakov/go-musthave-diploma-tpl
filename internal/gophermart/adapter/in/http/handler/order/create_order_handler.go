package order

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/StasMerzlyakov/gophermart/internal/gophermart/domain"
)

//go:generate mockgen -destination "../mocks/$GOFILE" -package mocks . CreateOrderApp
type CreateOrderApp interface {
	New(ctx context.Context, number domain.OrderNumber) error
}

// POST /api/user/orders
func CreateHandler(app CreateOrderApp) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {

		handlerName := "order.CreateHandler"

		logger, err := domain.GetCtxLogger(req.Context())
		if err != nil {
			fmt.Printf("%v:  can't extract logger\n", handlerName)
			http.Error(w, "can't extract logger", http.StatusInternalServerError)
			return
		}

		number, err := io.ReadAll(req.Body)
		defer req.Body.Close()

		if err != nil {
			logger.Errorw(handlerName, "err", "can't read body")
			http.Error(w, "can't read body", http.StatusBadRequest)
			return
		}

		numberStr := string(number)
		numberStr = strings.Trim(numberStr, " ")

		err = app.New(req.Context(), domain.OrderNumber(numberStr))

		if err != nil {
			http.Error(w, err.Error(), domain.MapDomainErrorToHTTPStatusErr(err))
			return
		}

		w.WriteHeader(http.StatusAccepted)
	}
}
