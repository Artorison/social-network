package usersmysql

import (
	"context"
	"database/sql"
	"errors"
	"redditclone/internal/models"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func newMock(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	require.NoError(t, err, "cant create mock")
	t.Cleanup(func() {
		db.Close()
		require.NoError(t, mock.ExpectationsWereMet())
	})
	return db, mock
}

func newUser(id, username, passw string) *models.User {
	return &models.User{
		Username:     username,
		ID:           id,
		PasswordHash: passw,
	}
}

func TestGetUser(t *testing.T) {
	db, mock := newMock(t)
	uRepo := NewUsersMysqlRepo(db)
	ctx := context.Background()
	const query = `SELECT id, password_hash FROM users WHERE username = ?;`

	user := newUser("afnoieiolkd3oinfosienf", "Artor", "ngsoiesnfgo")

	rows := sqlmock.NewRows([]string{"id", "password_hash"}).
		AddRow(user.ID, user.PasswordHash)

	mock.ExpectQuery(query).WithArgs(user.Username).WillReturnRows(rows)

	userRes, err := uRepo.GetUser(ctx, user.Username)
	require.NoError(t, err, "get user from DB")
	require.Equal(t, user, userRes)

	mock.ExpectQuery(query).WithArgs("no_user").WillReturnError(sql.ErrNoRows)
	userRes, err = uRepo.GetUser(ctx, "no_user")
	require.ErrorIs(t, err, models.ErrUserNotFound)
	require.Nil(t, userRes)

	dbErr := errors.New("bad connect")
	mock.ExpectQuery(query).WithArgs(user.Username).WillReturnError(dbErr)

	userRes, err = uRepo.GetUser(ctx, user.Username)
	require.ErrorIs(t, err, dbErr)
	require.Nil(t, userRes)
}

func TestCreateUser(t *testing.T) {
	const check = `SELECT 1 FROM users WHERE username = ? LIMIT 1;`
	const insert = `INSERT INTO users (id, username, password_hash)
	VALUES (?, ?, ?);`
	dbErr := errors.New("bad connect")

	testCases := []struct {
		name      string
		user      *models.User
		mockFunc  func(mock sqlmock.Sqlmock)
		wantedErr error
	}{
		{
			name: "OK",
			user: newUser("id1", "name1", "password1"),
			mockFunc: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(check).
					WithArgs("name1").
					WillReturnError(sql.ErrNoRows)

				mock.ExpectExec(insert).
					WithArgs("id1", "name1", "password1").
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			wantedErr: nil,
		},
		{
			name: "user_exists",
			user: newUser("id1", "name1", "password1"),
			mockFunc: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"1"}).AddRow("1")
				mock.ExpectQuery(check).
					WithArgs("name1").
					WillReturnRows(rows)
			},
			wantedErr: models.ErrUserExists,
		},
		{
			name: "check err",
			user: newUser("id1", "name1", "password1"),
			mockFunc: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(check).
					WithArgs("name1").
					WillReturnError(dbErr)
			},
			wantedErr: dbErr,
		},
		{
			name: "insert err",
			user: newUser("id1", "name1", "password1"),
			mockFunc: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(check).
					WithArgs("name1").
					WillReturnError(sql.ErrNoRows)

				mock.ExpectExec(insert).
					WithArgs("id1", "name1", "password1").
					WillReturnError(dbErr)
			},
			wantedErr: dbErr,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			t.Parallel()
			db, mock := newMock(t)
			uRepo := NewUsersMysqlRepo(db)

			tC.mockFunc(mock)

			err := uRepo.CreateUser(context.Background(), tC.user)

			if tC.wantedErr != nil {
				require.ErrorIs(t, err, tC.wantedErr)
			} else {
				require.NoError(t, err)
			}
		})
	}

}
