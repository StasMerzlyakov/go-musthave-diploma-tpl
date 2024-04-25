package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/StasMerzlyakov/gophermart/internal/gophermart/domain"
)

//go:generate mockgen -destination "../mocks/$GOFILE" -package mocks . LogingApp
type LogingApp interface {
	Login(ctx context.Context, userData *domain.AuthentificationData) (domain.TokenString, error)
}

// POST /api/user/login
func LoginHandler(app LogingApp) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {

		handlerName := "LoginHandler"

		logger, err := domain.GetCtxLogger(req.Context())
		if err != nil {
			fmt.Printf("%v:  can't extract logger\n", handlerName)
			http.Error(w, "can't extract logger", http.StatusInternalServerError)
			return
		}

		contentType := req.Header.Get("Content-Type")
		if contentType != domain.ApplicationJSON {
			logger.Infow(handlerName, "err", "unexpected content type")
			http.Error(w, "unexpected content type", http.StatusBadRequest)
			return
		}

		var login *domain.AuthentificationData

		if err := json.NewDecoder(req.Body).Decode(&login); err != nil {
			logger.Infow(handlerName, "err", fmt.Sprintf("json decode error - %v", err.Error()))
			http.Error(w, "json decode error", http.StatusBadRequest)
			return
		}

		tokenString, err := app.Login(req.Context(), login)
		if err != nil {
			http.Error(w, "registration error", domain.MapDomainErrorToHttpStatusErr(err))
			return
		}

		w.Header().Set(domain.AuthorizationHeader, fmt.Sprintf("Bearer %v", tokenString))
	}
}
