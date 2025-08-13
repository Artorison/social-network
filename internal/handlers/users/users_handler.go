package users

import (
	"context"

	"github.com/Artorison/social-network/pkg/logger"
)

type UsersService interface {
	RegisterUser(ctx context.Context, username, password string) (string, error)
	LoginUser(ctx context.Context, username, password string) (string, error)
}
type UsersHandler struct {
	Service UsersService
	Logger  *logger.Logger
}

func NewUserHandler(service UsersService, logger *logger.Logger) *UsersHandler {
	return &UsersHandler{Service: service, Logger: logger}
}
