package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/StasMerzlyakov/gophermart/internal/gophermart/domain"
)

//go:generate mockgen -destination "../mocks/$GOFILE" -package mocks . RegisterApp
type RegisterApp interface {
	Login(ctx context.Context, userData *domain.AuthentificationData) (domain.TokenString, error)
	Register(ctx context.Context, regData *domain.RegistrationData) (domain.TokenString, error)
}

// POST /api/user/register
func RegisterHandler(app RegisterApp) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {

		handlerName := "auth.RegisterHandler"

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

		var registration *domain.RegistrationData

		if err := json.NewDecoder(req.Body).Decode(&registration); err != nil {
			logger.Errorw(handlerName, "err", fmt.Sprintf("json decode error - %v", err.Error()))
			http.Error(w, "json decode error", http.StatusBadRequest)
			return
		}

		tokenString, err := app.Register(req.Context(), registration)
		if err != nil {
			http.Error(w, "registration error", domain.MapDomainErrorToHttpStatusErr(err))
			return
		}

		w.Header().Set(domain.AuthorizationHeader, fmt.Sprintf("Bearer %v", tokenString))
	}
}
