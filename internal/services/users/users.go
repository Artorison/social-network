package users

import (
	"context"
	"fmt"
	"time"

	"github.com/Artorison/social-network/internal/models"
	"github.com/Artorison/social-network/internal/sessions"
	"github.com/Artorison/social-network/pkg/helpers"
)

func (s *UserService) RegisterUser(ctx context.Context, username, password string) (string, error) {
	hash, err := sessions.GenerateHashPassword(password)
	if err != nil {
		return "", err
	}

	user := &models.User{
		ID:           helpers.GenerateID(),
		PasswordHash: hash,
		Username:     username,
	}

	if err = s.Repo.CreateUser(ctx, user); err != nil {
		return "", err
	}

	token, err := sessions.NewJWT(user)
	if err != nil {
		return "", fmt.Errorf("JWT generate: %w", err)
	}

	expiresAt := time.Now().Add(time.Hour * 24)

	if err := s.Sm.Create(ctx, token, sessions.NewSession(user), expiresAt); err != nil {
		return "", fmt.Errorf("create session: %w", err)
	}

	return token, nil
}

func (s *UserService) LoginUser(ctx context.Context, username, password string) (string, error) {
	user, err := s.Repo.GetUser(ctx, username)
	if err != nil {
		return "", err
	}

	if !sessions.CheckHashPassword(user.PasswordHash, password) {
		return "", models.ErrInvalidLogin
	}

	token, err := sessions.NewJWT(user)
	if err != nil {
		return "", fmt.Errorf("JWT generate: %w", err)
	}

	expiresAt := time.Now().Add(time.Hour * 24)

	if err := s.Sm.Create(ctx, token, sessions.NewSession(user), expiresAt); err != nil {
		return "", fmt.Errorf("create session: %w", err)
	}

	return token, nil
}
