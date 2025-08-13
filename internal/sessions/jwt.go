package sessions

import (
	"fmt"
	"log/slog"
	"redditclone/internal/models"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var secretKey string

func MustSecretKey(key string) {
	if key == "" {
		secretKey = "default-secret-key"
		slog.Warn("empty secret key in sessions.MustSecretKey")
		return
	}
	secretKey = key
}

type Claims struct {
	User *models.User `json:"user"`
	jwt.RegisteredClaims
}

func NewJWT(user *models.User) (string, error) {
	now := time.Now()
	claims := &Claims{
		User: user,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secretKey))
}

func ParseJWT(tokenStr string) (*models.User, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected jwt method")
		}
		return ([]byte(secretKey)), nil
	})

	if err != nil || !token.Valid || claims.User == nil {
		return nil, models.ErrInvalidToken
	}

	return claims.User, nil
}
