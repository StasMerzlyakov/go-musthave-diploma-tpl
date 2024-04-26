package auth

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/StasMerzlyakov/gophermart/internal/gophermart/adapter/in/http/middleware"
	"github.com/StasMerzlyakov/gophermart/internal/gophermart/domain"
)

type AuthApp interface {
	Authorize(ctx context.Context, tokenString domain.TokenString) (*domain.AuthData, error)
}

// Отвечает за авторизацию по JWT-токену.
func NewJwtRequestMW(app AuthApp) middleware.Middleware {
	return func(next http.Handler) http.Handler {
		authFn := func(w http.ResponseWriter, req *http.Request) {

			log, err := domain.GetCtxLogger(req.Context())
			if err != nil {
				fmt.Printf("JwtRequestMW - can't extract logger")
				http.Error(w, "logging configuration error", http.StatusInternalServerError)
				return
			}

			reqToken := req.Header.Get(domain.AuthorizationHeader)
			if reqToken == "" {
				errMsg := "Authorization header is not set"
				log.Infow("JwtRequestMW", "err", errMsg)
				http.Error(w, errMsg, http.StatusUnauthorized)
				return
			}
			splitToken := strings.Split(reqToken, "Bearer ")
			if len(splitToken) != 2 {
				errMsg := "unexpected Authorization header value"
				log.Infow("JwtRequestMW", "err", errMsg)
				http.Error(w, errMsg, http.StatusUnauthorized)
				return
			}
			tokenString := domain.TokenString(splitToken[1])

			authData, err := app.Authorize(req.Context(), tokenString)
			if err != nil {
				log.Infow("JwtRequestMW", "err", err.Error())
				http.Error(w, err.Error(), domain.MapDomainErrorToHTTPStatusErr(err))
				return
			}

			ctx, err := domain.EnrichWithAuthData(req.Context(), authData)
			if err != nil {
				log.Infow("JwtRequestMW", "err", err.Error())
				http.Error(w, err.Error(), domain.MapDomainErrorToHTTPStatusErr(err))
				return
			}

			req = req.WithContext(ctx)
			next.ServeHTTP(w, req)

		}
		return http.HandlerFunc(authFn)
	}
}
