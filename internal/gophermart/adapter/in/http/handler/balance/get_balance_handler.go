package balance

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/StasMerzlyakov/gophermart/internal/gophermart/domain"
)

//go:generate mockgen -destination "../mocks/$GOFILE" -package mocks . GetBalanceApp
type GetBalanceApp interface {
	Get(ctx context.Context) (*domain.Balance, error)
}

// GET /api/user/balance
func GetHandler(app GetBalanceApp) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {

		handlerName := "balance.GetHandler"

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

		data, err := app.Get(req.Context())
		if err != nil {
			http.Error(w, err.Error(), domain.MapDomainErrorToHttpStatusErr(err))
			return
		}

		w.Header().Set("Content-Type", domain.ApplicationJSON)

		if err := json.NewEncoder(w).Encode(data); err != nil {
			logger.Infow(handlerName, "err", fmt.Sprintf("json encode error: %v", err.Error()))
			http.Error(w, err.Error(), domain.MapDomainErrorToHttpStatusErr(err))
			return
		}
	}
}
