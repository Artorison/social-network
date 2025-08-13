package users

import (
	"context"
	"redditclone/internal/models"
	"redditclone/internal/sessions"
)

type UsersRepo interface {
	CreateUser(ctx context.Context, u *models.User) error
	GetUser(ctx context.Context, username string) (*models.User, error)
}
type UserService struct {
	Repo UsersRepo
	Sm   sessions.SessionManager
}

func NewUserService(repo UsersRepo, sm sessions.SessionManager) *UserService {
	return &UserService{Repo: repo, Sm: sm}
}
