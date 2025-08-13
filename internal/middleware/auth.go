package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/Artorison/social-network/internal/models"
	"github.com/Artorison/social-network/internal/sessions"
	"github.com/gorilla/mux"
)

func Auth(sm sessions.SessionManager) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, err := getBearerFromHeader(r.Header.Get("Authorization"))
			if err != nil {
				models.JSONError(w, http.StatusUnauthorized, err.Error())
			}

			if _, err = sessions.ParseJWT(token); err != nil {
				models.JSONError(w, http.StatusUnauthorized, models.ErrInvalidToken.Error())
				return
			}

			user, err := sm.GetUserByToken(r.Context(), token)
			if err != nil {
				models.JSONError(w, http.StatusUnauthorized, models.ErrInvalidToken.Error())
				return
			}

			ctx := sessions.SessionToCtx(r.Context(), sessions.NewSession(user))

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func getBearerFromHeader(header string) (string, error) {
	if header == "" {
		return "", fmt.Errorf("token required")
	}
	headerParts := strings.Split(header, " ")
	if len(headerParts) != 2 || headerParts[0] != "Bearer" {
		return "", fmt.Errorf("invalid auth header")
	}

	if len(headerParts[1]) == 0 {
		return "", fmt.Errorf("token is empty")
	}
	return headerParts[1], nil
}
