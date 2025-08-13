package users

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"redditclone/internal/models"
	"redditclone/pkg/logger"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

var (
	applJSON = "application/json"
	register = "/register"
	login    = "/login"
	token    = "TOKEN"
	uname    = "username123"
	passw    = "password123"
	errTest  = errors.New("test error")
)

type rwRec interface {
	http.ResponseWriter
	Result() *http.Response
}

type failingWriter struct {
	*httptest.ResponseRecorder
	failOnce bool
}

func newFailingWriter() rwRec {
	return &failingWriter{ResponseRecorder: httptest.NewRecorder(), failOnce: true}
}

func (fw *failingWriter) Write(b []byte) (int, error) {
	if fw.failOnce {
		fw.failOnce = false
		return 0, errTest
	}
	return fw.ResponseRecorder.Write(b)
}

func mustJSON(t *testing.T, v any) *bytes.Reader {
	t.Helper()
	b, err := json.Marshal(v)
	require.NoError(t, err)
	return bytes.NewReader(b)
}

func TestHandlerRegister(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSrv := NewMockUsersService(ctrl)
	h := NewUserHandler(mockSrv, &logger.Logger{
		Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
	})

	testCases := []struct {
		name             string
		req              *http.Request
		recorder         func() rwRec
		mockExpect       func()
		wantedStatusCode int
		wantedInBody     string
	}{
		{
			"wrong content-type",
			httptest.NewRequest(http.MethodPost, register, nil),
			func() rwRec { return httptest.NewRecorder() },
			func() {},
			http.StatusBadRequest,
			"unknown payload",
		},
		{
			"bad json",
			func() *http.Request {
				r := httptest.NewRequest(http.MethodPost, register, bytes.NewBufferString("{"))
				r.Header.Set("Content-Type", applJSON)
				return r
			}(),
			func() rwRec { return httptest.NewRecorder() },
			func() {},
			http.StatusOK,
			"",
		},
		{
			"user already exists",
			func() *http.Request {
				r := httptest.NewRequest(http.MethodPost, register,
					mustJSON(t, models.LoginForm{Username: uname, Password: passw}))
				r.Header.Set("Content-Type", applJSON)
				return r
			}(),
			newFailingWriter,
			func() {
				mockSrv.EXPECT().
					RegisterUser(gomock.Any(), uname, passw).
					Return("", models.ErrUserExists)
			},
			http.StatusUnprocessableEntity,
			"",
		},
		{
			"internal error",
			func() *http.Request {
				r := httptest.NewRequest(http.MethodPost, register,
					mustJSON(t, models.LoginForm{Username: uname, Password: passw}))
				r.Header.Set("Content-Type", applJSON)
				return r
			}(),
			newFailingWriter,
			func() {
				mockSrv.EXPECT().
					RegisterUser(gomock.Any(), uname, passw).
					Return("", errTest)
			},
			http.StatusInternalServerError,
			"",
		},
		{
			"encoder error",
			func() *http.Request {
				r := httptest.NewRequest(http.MethodPost, register,
					mustJSON(t, models.LoginForm{Username: uname, Password: passw}))
				r.Header.Set("Content-Type", applJSON)
				return r
			}(),
			newFailingWriter,
			func() {
				mockSrv.EXPECT().
					RegisterUser(gomock.Any(), uname, passw).
					Return(token, nil)
			},
			http.StatusCreated,
			"JSON encoder error",
		},
		{
			"success",
			func() *http.Request {
				r := httptest.NewRequest(http.MethodPost, register,
					mustJSON(t, models.LoginForm{Username: uname, Password: passw}))
				r.Header.Set("Content-Type", applJSON)
				return r
			}(),
			func() rwRec { return httptest.NewRecorder() },
			func() {
				mockSrv.EXPECT().
					RegisterUser(gomock.Any(), uname, passw).
					Return(token, nil)
			},
			http.StatusCreated,
			`{"token":"` + token + `"}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tc.mockExpect()
			rec := tc.recorder()

			h.Register(rec, tc.req)

			res := rec.Result()
			defer res.Body.Close()
			body, err := io.ReadAll(res.Body)
			require.NoError(t, err)

			require.Equal(t, tc.wantedStatusCode, res.StatusCode, string(body))
			if tc.wantedInBody != "" {
				require.Contains(t, string(body), tc.wantedInBody)
			}
		})
	}
}

func TestHandlerLogin(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSrv := NewMockUsersService(ctrl)
	h := NewUserHandler(mockSrv, &logger.Logger{
		Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
	})

	testCases := []struct {
		name             string
		req              *http.Request
		recorder         func() rwRec
		mockExpect       func()
		wantedStatusCode int
		wantedInBody     string
	}{
		{
			"wrong content-type",
			httptest.NewRequest(http.MethodPost, login, nil),
			func() rwRec { return httptest.NewRecorder() },
			func() {},
			http.StatusBadRequest,
			"unknown payload",
		},
		{
			"bad json",
			func() *http.Request {
				r := httptest.NewRequest(http.MethodPost, login, bytes.NewBufferString("{"))
				r.Header.Set("Content-Type", applJSON)
				return r
			}(),
			func() rwRec { return httptest.NewRecorder() },
			func() {},
			http.StatusInternalServerError,
			"EOF",
		},
		{
			"user not found",
			func() *http.Request {
				r := httptest.NewRequest(http.MethodPost, login,
					mustJSON(t, models.LoginForm{Username: uname, Password: passw}))
				r.Header.Set("Content-Type", applJSON)
				return r
			}(),
			newFailingWriter,
			func() {
				mockSrv.EXPECT().
					LoginUser(gomock.Any(), uname, passw).
					Return("", models.ErrUserNotFound)
			},
			http.StatusUnauthorized,
			"",
		},
		{
			"invalid password",
			func() *http.Request {
				r := httptest.NewRequest(http.MethodPost, login,
					mustJSON(t, models.LoginForm{Username: uname, Password: uname}))
				r.Header.Set("Content-Type", applJSON)
				return r
			}(),
			newFailingWriter,
			func() {
				mockSrv.EXPECT().
					LoginUser(gomock.Any(), uname, uname).
					Return("", models.ErrInvalidLogin)
			},
			http.StatusUnauthorized,
			"",
		},
		{
			"internal error",
			func() *http.Request {
				r := httptest.NewRequest(http.MethodPost, login,
					mustJSON(t, models.LoginForm{Username: uname, Password: passw}))
				r.Header.Set("Content-Type", applJSON)
				return r
			}(),
			newFailingWriter,
			func() {
				mockSrv.EXPECT().
					LoginUser(gomock.Any(), uname, passw).
					Return("", errTest)
			},
			http.StatusInternalServerError,
			"",
		},
		{
			"encoder error",
			func() *http.Request {
				r := httptest.NewRequest(http.MethodPost, login,
					mustJSON(t, models.LoginForm{Username: uname, Password: passw}))
				r.Header.Set("Content-Type", applJSON)
				return r
			}(),
			newFailingWriter,
			func() {
				mockSrv.EXPECT().
					LoginUser(gomock.Any(), uname, passw).
					Return(token, nil)
			},
			http.StatusInternalServerError,
			"JSON encoder error",
		},
		{
			"success",
			func() *http.Request {
				r := httptest.NewRequest(http.MethodPost, login,
					mustJSON(t, models.LoginForm{Username: uname, Password: passw}))
				r.Header.Set("Content-Type", applJSON)
				return r
			}(),
			func() rwRec { return httptest.NewRecorder() },
			func() {
				mockSrv.EXPECT().
					LoginUser(gomock.Any(), uname, passw).
					Return(token, nil)
			},
			http.StatusOK,
			`{"token":"` + token + `"}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tc.mockExpect()
			rec := tc.recorder()

			h.Login(rec, tc.req)

			res := rec.Result()
			defer res.Body.Close()
			body, err := io.ReadAll(res.Body)
			require.NoError(t, err)

			require.Equal(t, tc.wantedStatusCode, res.StatusCode, string(body))
			if tc.wantedInBody != "" {
				require.Contains(t, string(body), tc.wantedInBody)
			}
		})
	}
}
