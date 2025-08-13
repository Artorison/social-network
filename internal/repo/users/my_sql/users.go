package usersmysql

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"redditclone/internal/models"
)

type UsersMysqlRepo struct {
	DB *sql.DB
}

func NewUsersMysqlRepo(db *sql.DB) *UsersMysqlRepo {
	return &UsersMysqlRepo{DB: db}
}

func (r *UsersMysqlRepo) CreateUser(ctx context.Context, u *models.User) error {

	const check = `SELECT 1 FROM users WHERE username = ? LIMIT 1;`
	var exists int
	err := r.DB.QueryRowContext(ctx, check, u.Username).Scan(&exists)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	if exists == 1 {
		return models.ErrUserExists
	}

	const query = `INSERT INTO users (id, username, password_hash)
	VALUES (?, ?, ?);`

	_, err = r.DB.ExecContext(ctx, query, u.ID, u.Username, u.PasswordHash)
	if err != nil {
		return err
	}

	return nil
}

func (r *UsersMysqlRepo) GetUser(ctx context.Context, username string) (*models.User, error) {
	const query = `SELECT id, password_hash FROM users WHERE username = ?;`

	var user = models.User{Username: username}
	err := r.DB.QueryRowContext(ctx, query, username).Scan(&user.ID, &user.PasswordHash)
	if err == sql.ErrNoRows {
		slog.Error("user not found", slog.String("username", username))
		return nil, models.ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}

	return &user, nil
}
