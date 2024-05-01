package balance

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/StasMerzlyakov/gophermart/internal/gophermart/domain"
)

//go:generate mockgen -destination "../mocks/$GOFILE" -package mocks . WithdrawApp
type WithdrawApp interface {
	Withdraw(ctx context.Context, withdraw *domain.WithdrawData) error
}

// POST /api/user/balance/withdraw
func WithdrawHandler(app WithdrawApp) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {

		handlerName := "balance.WithdrawHandler"

		logger, err := domain.GetCtxLogger(req.Context())
		if err != nil {
			fmt.Printf("%v:  can't extract logger\n", handlerName)
			http.Error(w, "can't extract logger", http.StatusInternalServerError)
			return
		}

		contentType := req.Header.Get("Content-Type")
		if contentType != domain.ApplicationJSON {
			logger.Errorw(handlerName, "err", "unexpected content type")
			http.Error(w, "unexpected content type", http.StatusBadRequest)
			return
		}

		var data *domain.WithdrawData

		if err := json.NewDecoder(req.Body).Decode(&data); err != nil {
			logger.Errorw(handlerName, "err", fmt.Sprintf("json decode error - %v", err.Error()))
			http.Error(w, "json decode error", http.StatusBadRequest)
			return
		}

		if err = app.Withdraw(req.Context(), data); err != nil {
			http.Error(w, "withraw error", domain.MapDomainErrorToHTTPStatusErr(err))
		}
	}
}
