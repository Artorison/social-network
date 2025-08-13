package sessions

import (
	"context"
	"database/sql"
	"fmt"
	"redditclone/internal/models"
	"time"
)

type SessionManager interface {
	GetUserByToken(ctx context.Context, token string) (*models.User, error)
	Create(ctx context.Context, token string, s *Session, expiresAt time.Time) error
	DestroyCurrent(ctx context.Context, s *Session) error
}

type SessionManagerRepo struct {
	DB *sql.DB
}

func NewSessionManager(db *sql.DB) *SessionManagerRepo {
	return &SessionManagerRepo{DB: db}
}

func (sm *SessionManagerRepo) GetUserByToken(ctx context.Context, token string) (*models.User, error) {
	const query = `SELECT u.id, u.username
	FROM sessions s 
	JOIN users u ON s.user_id = u.id
	WHERE s.token = ?
	AND expires_at > NOW();`

	var user models.User
	err := sm.DB.QueryRowContext(ctx, query, token).Scan(&user.ID, &user.Username)
	if err == sql.ErrNoRows {
		return nil, models.ErrInvalidToken
	}
	if err != nil {
		return nil, fmt.Errorf("get user by token: %w", err)
	}

	return &user, nil
}

func (sm *SessionManagerRepo) Create(ctx context.Context, token string, s *Session, expiresAt time.Time) error {
	const query = `INSERT INTO sessions (id, token, user_id, expires_at)
	VALUES (?, ?, ?, ?);`

	_, err := sm.DB.ExecContext(ctx, query, s.ID, token, s.User.ID, expiresAt)
	if err != nil {
		return fmt.Errorf("exec session: %w", err)
	}
	return nil
}

func (sm *SessionManagerRepo) DestroyCurrent(ctx context.Context, s *Session) error {
	const query = `DELETE FROM sessions WHERE id = ?;`
	if _, err := sm.DB.ExecContext(ctx, query, s.ID); err != nil {
		return fmt.Errorf("destroy session: %w", err)
	}
	return nil
}
