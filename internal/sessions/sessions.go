package sessions

import (
	"context"
	"redditclone/internal/models"
	"redditclone/pkg/helpers"
)

type Session struct {
	ID   string
	User *models.User
}

func NewSession(user *models.User) *Session {
	return &Session{User: user, ID: helpers.GenerateID()}
}

type ctxKey struct{}

var key ctxKey

func SessionToCtx(ctx context.Context, s *Session) context.Context {
	return context.WithValue(ctx, key, s)
}

func GetSessionFromCtx(ctx context.Context) (*Session, bool) {
	s, ok := ctx.Value(key).(*Session)
	return s, ok
}
