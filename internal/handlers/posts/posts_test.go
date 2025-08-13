package posts

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"redditclone/internal/models"
	"redditclone/internal/sessions"
	"redditclone/pkg/logger"

	"log/slog"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

var (
	title        = "some title"
	category     = "programming"
	username     = "user123"
	strPostID    = "awdno2dinonawnd"
	strCommentID = "2adwoiudnawd2"
)

var errSvc = errors.New("service error")

type badRW struct {
	*httptest.ResponseRecorder
	fail bool
}

func (b *badRW) Write(p []byte) (int, error) {
	if b.fail {
		return 0, errors.New("write err")
	}
	return b.ResponseRecorder.Write(p)
}

func newLog() *logger.Logger {
	return &logger.Logger{Logger: slog.New(slog.NewTextHandler(io.Discard, nil))}
}

func TestGetAllPosts(t *testing.T) {
	type ret struct {
		posts []*models.Post
		err   error
	}
	testCases := []struct {
		name             string
		svcRet           ret
		failWrite        bool
		wantedStatusCode int
	}{
		{"ok", ret{[]*models.Post{{Title: title}}, nil}, false, http.StatusOK},
		{"svcErr", ret{nil, errSvc}, false, http.StatusInternalServerError},
		{"encodeErr", ret{[]*models.Post{{Title: title}}, nil}, true, http.StatusInternalServerError},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			svc := NewMockPostsService(ctrl)
			svc.EXPECT().
				GetAllPosts(gomock.Any()).
				Return(tc.svcRet.posts, tc.svcRet.err)

			h := NewPostHandler(svc, newLog())

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rw := &badRW{httptest.NewRecorder(), tc.failWrite}

			h.GetAllPosts(rw, req)

			if rw.Code != tc.wantedStatusCode {
				t.Fatalf("code %d want %d", rw.Code, tc.wantedStatusCode)
			}
		})
	}
}

func TestCreatePost(t *testing.T) {
	dto := models.CreatePostDTO{Title: title}
	okJSON, err := json.Marshal(dto)
	require.NoError(t, err)
	badJSON := []byte("{")

	sess := sessions.NewSession(&models.User{Username: username})
	post := &models.Post{Title: title}

	testCases := []struct {
		name             string
		body             []byte
		withSess         bool
		svcPost          *models.Post
		svcErr           error
		failWrite        bool
		wantedStatusCode int
		callSvc          bool
	}{
		{"badJSON", badJSON, true, nil, nil, false, http.StatusBadRequest, false},
		{"unauth", okJSON, false, nil, nil, false, http.StatusUnauthorized, false},
		{"svcErr", okJSON, true, nil, errSvc, false, http.StatusInternalServerError, true},
		{"encodeErr", okJSON, true, post, nil, true, http.StatusCreated, true},
		{"ok", okJSON, true, post, nil, false, http.StatusCreated, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			svc := NewMockPostsService(ctrl)
			if tc.callSvc {
				svc.EXPECT().
					CreatePost(gomock.Any(), dto, sess).
					Return(tc.svcPost, tc.svcErr)
			}

			h := NewPostHandler(svc, newLog())

			req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(tc.body))
			if tc.withSess {
				req = req.WithContext(sessions.SessionToCtx(req.Context(), sess))
			}
			rw := &badRW{httptest.NewRecorder(), tc.failWrite}

			h.CreatePost(rw, req)

			if rw.Code != tc.wantedStatusCode {
				t.Fatalf("code %d want %d", rw.Code, tc.wantedStatusCode)
			}
		})
	}
}

func TestGetPostByCategory(t *testing.T) {
	type ret struct {
		posts []*models.Post
		err   error
	}
	testCases := []struct {
		name             string
		vars             map[string]string
		svcRet           ret
		failWrite        bool
		wantedStatusCode int
		callSvc          bool
	}{
		{"noVar", nil, ret{}, false, http.StatusBadRequest, false},
		{"svcErr", map[string]string{"category": category}, ret{nil, errSvc}, false, http.StatusInternalServerError, true},
		{"encodeErr", map[string]string{"category": category}, ret{[]*models.Post{{Title: title}}, nil}, true, http.StatusInternalServerError, true},
		{"ok", map[string]string{"category": category}, ret{[]*models.Post{{Title: title}}, nil}, false, http.StatusOK, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			svc := NewMockPostsService(ctrl)
			if tc.callSvc {
				svc.EXPECT().
					GetPostsByCategory(gomock.Any(), category).
					Return(tc.svcRet.posts, tc.svcRet.err)
			}

			h := NewPostHandler(svc, newLog())

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tc.vars != nil {
				req = mux.SetURLVars(req, tc.vars)
			}
			rw := &badRW{httptest.NewRecorder(), tc.failWrite}

			h.GetPostByCategory(rw, req)

			if rw.Code != tc.wantedStatusCode {
				t.Fatalf("code %d want %d", rw.Code, tc.wantedStatusCode)
			}
		})
	}
}

func TestGetPostByID(t *testing.T) {
	found := &models.Post{Title: title}
	type ret struct {
		post *models.Post
		err  error
	}
	testCases := []struct {
		name             string
		vars             map[string]string
		svcRet           ret
		failWrite        bool
		wantedStatusCode int
		callSvc          bool
	}{
		{"noVar", nil, ret{}, false, http.StatusBadRequest, false},
		{"notFound", map[string]string{"post_id": strPostID}, ret{nil, models.ErrPostNotFound}, false, http.StatusBadRequest, true},
		{"svcErr", map[string]string{"post_id": strPostID}, ret{nil, errSvc}, false, http.StatusInternalServerError, true},
		{"encodeErr", map[string]string{"post_id": strPostID}, ret{found, nil}, true, http.StatusInternalServerError, true},
		{"ok", map[string]string{"post_id": strPostID}, ret{found, nil}, false, http.StatusOK, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			svc := NewMockPostsService(ctrl)
			if tc.callSvc {
				svc.EXPECT().
					GetPostByID(gomock.Any(), strPostID).
					Return(tc.svcRet.post, tc.svcRet.err)
			}

			h := NewPostHandler(svc, newLog())

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tc.vars != nil {
				req = mux.SetURLVars(req, tc.vars)
			}
			rw := &badRW{httptest.NewRecorder(), tc.failWrite}

			h.GetPostByID(rw, req)

			if rw.Code != tc.wantedStatusCode {
				t.Fatalf("code %d want %d", rw.Code, tc.wantedStatusCode)
			}
		})
	}
}

func TestGetPostsByUser(t *testing.T) {
	type ret struct {
		posts []*models.Post
		err   error
	}
	testCases := []struct {
		name             string
		vars             map[string]string
		svcRet           ret
		failWrite        bool
		wantedStatusCode int
		callSvc          bool
	}{
		{"noVar", nil, ret{}, false, http.StatusBadRequest, false},
		{"svcErr", map[string]string{"username": username}, ret{nil, errSvc}, false, http.StatusInternalServerError, true},
		{"encodeErr", map[string]string{"username": username}, ret{[]*models.Post{{Title: title}}, nil}, true, http.StatusInternalServerError, true},
		{"ok", map[string]string{"username": username}, ret{[]*models.Post{{Title: title}}, nil}, false, http.StatusOK, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			svc := NewMockPostsService(ctrl)
			if tc.callSvc {
				svc.EXPECT().
					GetUserPosts(gomock.Any(), username).
					Return(tc.svcRet.posts, tc.svcRet.err)
			}

			h := NewPostHandler(svc, newLog())

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tc.vars != nil {
				req = mux.SetURLVars(req, tc.vars)
			}
			rw := &badRW{httptest.NewRecorder(), tc.failWrite}

			h.GetPostsByUser(rw, req)

			if rw.Code != tc.wantedStatusCode {
				t.Fatalf("code %d want %d", rw.Code, tc.wantedStatusCode)
			}
		})
	}
}

func TestDeletePost(t *testing.T) {
	testCases := []struct {
		name             string
		vars             map[string]string
		svcErr           error
		failWrite        bool
		wantedStatusCode int
		callSvc          bool
	}{
		{"noVar", nil, nil, false, http.StatusBadRequest, false},
		{"svcErr", map[string]string{"post_id": strPostID}, errSvc, false, http.StatusInternalServerError, true},
		{"encodeErr", map[string]string{"post_id": strPostID}, nil, true, http.StatusInternalServerError, true},
		{"ok", map[string]string{"post_id": strPostID}, nil, false, http.StatusOK, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			svc := NewMockPostsService(ctrl)
			if tc.callSvc {
				svc.EXPECT().
					DeletePost(gomock.Any(), strPostID).
					Return(tc.svcErr)
			}

			h := NewPostHandler(svc, newLog())

			req := httptest.NewRequest(http.MethodDelete, "/", nil)
			if tc.vars != nil {
				req = mux.SetURLVars(req, tc.vars)
			}
			rw := &badRW{httptest.NewRecorder(), tc.failWrite}

			h.DeletePost(rw, req)

			if rw.Code != tc.wantedStatusCode {
				t.Fatalf("code %d want %d", rw.Code, tc.wantedStatusCode)
			}
		})
	}
}

func TestAddComment(t *testing.T) {
	dto := models.AddCommentDTO{CommentMsg: "hi"}
	okJSON, err := json.Marshal(dto)
	require.NoError(t, err)
	badJSON := []byte("{")

	sess := sessions.NewSession(&models.User{Username: "u"})
	post := &models.Post{Title: "p"}

	testCases := []struct {
		name             string
		vars             map[string]string
		body             []byte
		withSess         bool
		svcPost          *models.Post
		svcErr           error
		failWrite        bool
		wantedStatusCode int
		callSvc          bool
	}{
		{"noVar", nil, nil, false, nil, nil, false, http.StatusBadRequest, false},
		{"badJSON", map[string]string{"post_id": strPostID}, badJSON, true, nil, nil, false, http.StatusBadRequest, false},
		{"noSess", map[string]string{"post_id": strPostID}, okJSON, false, nil, nil, false, http.StatusUnauthorized, false},
		{"notFound", map[string]string{"post_id": strPostID}, okJSON, true, nil, models.ErrPostNotFound, false, http.StatusBadRequest, true},
		{"svcErr", map[string]string{"post_id": strPostID}, okJSON, true, nil, errSvc, false, http.StatusInternalServerError, true},
		{"encodeErr", map[string]string{"post_id": strPostID}, okJSON, true, post, nil, true, http.StatusCreated, true},
		{"ok", map[string]string{"post_id": strPostID}, okJSON, true, post, nil, false, http.StatusCreated, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			svc := NewMockPostsService(ctrl)
			if tc.callSvc {
				svc.EXPECT().
					AddComment(gomock.Any(), strPostID, "hi", sess).
					Return(tc.svcPost, tc.svcErr)
			}

			h := NewPostHandler(svc, newLog())

			req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(tc.body))
			if tc.vars != nil {
				req = mux.SetURLVars(req, tc.vars)
			}
			if tc.withSess {
				req = req.WithContext(sessions.SessionToCtx(req.Context(), sess))
			}
			rw := &badRW{httptest.NewRecorder(), tc.failWrite}

			h.AddComment(rw, req)

			if rw.Code != tc.wantedStatusCode {
				t.Fatalf("code %d want %d", rw.Code, tc.wantedStatusCode)
			}
		})
	}
}
func TestDeleteComent(t *testing.T) {
	post := &models.Post{Title: title}

	testCases := []struct {
		name             string
		vars             map[string]string
		svcPost          *models.Post
		svcErr           error
		failWrite        bool
		wantedStatusCode int
		callSvc          bool
	}{
		{"noPostVar", nil, nil, nil, false, http.StatusBadRequest, false},
		{"noCommentVar", map[string]string{"post_id": strPostID}, nil, nil, false, http.StatusBadRequest, false},
		{"notFound", map[string]string{"post_id": strPostID, "comment_id": strCommentID}, nil, models.ErrPostNotFound, false, http.StatusBadRequest, true},
		{"svcErr", map[string]string{"post_id": strPostID, "comment_id": strCommentID}, nil, errSvc, false, http.StatusInternalServerError, true},
		{"encodeErr", map[string]string{"post_id": strPostID, "comment_id": strCommentID}, post, nil, true, http.StatusInternalServerError, true},
		{"ok", map[string]string{"post_id": strPostID, "comment_id": strCommentID}, post, nil, false, http.StatusOK, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			svc := NewMockPostsService(ctrl)
			if tc.callSvc {
				svc.EXPECT().
					DeleteComment(gomock.Any(), strPostID, strCommentID).
					Return(tc.svcPost, tc.svcErr)
			}

			h := NewPostHandler(svc, newLog())

			req := httptest.NewRequest(http.MethodDelete, "/", nil)
			if tc.vars != nil {
				req = mux.SetURLVars(req, tc.vars)
			}
			rw := &badRW{httptest.NewRecorder(), tc.failWrite}

			h.DeleteComment(rw, req)

			if rw.Code != tc.wantedStatusCode {
				t.Fatalf("code %d want %d", rw.Code, tc.wantedStatusCode)
			}
		})
	}
}

func TestUppVote(t *testing.T) {
	sess := sessions.NewSession(&models.User{Username: username})
	post := &models.Post{Title: title}

	testCases := []struct {
		name             string
		vars             map[string]string
		withSess         bool
		svcPost          *models.Post
		svcErr           error
		failWrite        bool
		wantedStatusCode int
		callSvc          bool
	}{
		{"noVar", nil, false, nil, nil, false, http.StatusBadRequest, false},
		{"noSess", map[string]string{"post_id": strPostID}, false, nil, nil, false, http.StatusUnauthorized, false},
		{"notFound", map[string]string{"post_id": strPostID}, true, nil, models.ErrPostNotFound, false, http.StatusBadRequest, true},
		{"svcErr", map[string]string{"post_id": strPostID}, true, nil, errSvc, false, http.StatusInternalServerError, true},
		{"encodeErr", map[string]string{"post_id": strPostID}, true, post, nil, true, http.StatusInternalServerError, true},
		{"ok", map[string]string{"post_id": strPostID}, true, post, nil, false, http.StatusOK, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			svc := NewMockPostsService(ctrl)
			if tc.callSvc {
				svc.EXPECT().
					UpVote(gomock.Any(), strPostID, sess).
					Return(tc.svcPost, tc.svcErr)
			}

			h := NewPostHandler(svc, newLog())

			req := httptest.NewRequest(http.MethodPost, "/", nil)
			if tc.vars != nil {
				req = mux.SetURLVars(req, tc.vars)
			}
			if tc.withSess {
				req = req.WithContext(sessions.SessionToCtx(req.Context(), sess))
			}
			rw := &badRW{httptest.NewRecorder(), tc.failWrite}

			h.UppVote(rw, req)

			if rw.Code != tc.wantedStatusCode {
				t.Fatalf("code %d want %d", rw.Code, tc.wantedStatusCode)
			}
		})
	}
}

func TestDownVote(t *testing.T) {
	sess := sessions.NewSession(&models.User{Username: username})
	post := &models.Post{Title: title}

	testCases := []struct {
		name             string
		vars             map[string]string
		withSess         bool
		svcPost          *models.Post
		svcErr           error
		failWrite        bool
		wantedStatusCode int
		callSvc          bool
	}{
		{"noVar", nil, false, nil, nil, false, http.StatusBadRequest, false},
		{"noSess", map[string]string{"post_id": strPostID}, false, nil, nil, false, http.StatusUnauthorized, false},
		{"notFound", map[string]string{"post_id": strPostID}, true, nil, models.ErrPostNotFound, false, http.StatusBadRequest, true},
		{"svcErr", map[string]string{"post_id": strPostID}, true, nil, errSvc, false, http.StatusInternalServerError, true},
		{"encodeErr", map[string]string{"post_id": strPostID}, true, post, nil, true, http.StatusInternalServerError, true},
		{"ok", map[string]string{"post_id": strPostID}, true, post, nil, false, http.StatusOK, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			svc := NewMockPostsService(ctrl)
			if tc.callSvc {
				svc.EXPECT().
					DownVote(gomock.Any(), strPostID, sess).
					Return(tc.svcPost, tc.svcErr)
			}

			h := NewPostHandler(svc, newLog())

			req := httptest.NewRequest(http.MethodPost, "/", nil)
			if tc.vars != nil {
				req = mux.SetURLVars(req, tc.vars)
			}
			if tc.withSess {
				req = req.WithContext(sessions.SessionToCtx(req.Context(), sess))
			}
			rw := &badRW{httptest.NewRecorder(), tc.failWrite}

			h.DownVote(rw, req)

			if rw.Code != tc.wantedStatusCode {
				t.Fatalf("code %d want %d", rw.Code, tc.wantedStatusCode)
			}
		})
	}
}

func TestUnVote(t *testing.T) {
	sess := sessions.NewSession(&models.User{Username: username})
	post := &models.Post{Title: title}

	testCases := []struct {
		name             string
		vars             map[string]string
		withSess         bool
		svcPost          *models.Post
		svcErr           error
		failWrite        bool
		wantedStatusCode int
		callSvc          bool
	}{
		{"noVar", nil, false, nil, nil, false, http.StatusBadRequest, false},
		{"noSess", map[string]string{"post_id": strPostID}, false, nil, nil, false, http.StatusUnauthorized, false},
		{"notFound", map[string]string{"post_id": strPostID}, true, nil, models.ErrPostNotFound, false, http.StatusBadRequest, true},
		{"svcErr", map[string]string{"post_id": strPostID}, true, nil, errSvc, false, http.StatusInternalServerError, true},
		{"encodeErr", map[string]string{"post_id": strPostID}, true, post, nil, true, http.StatusInternalServerError, true},
		{"ok", map[string]string{"post_id": strPostID}, true, post, nil, false, http.StatusOK, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			svc := NewMockPostsService(ctrl)
			if tc.callSvc {
				svc.EXPECT().
					UnVote(gomock.Any(), strPostID, sess).
					Return(tc.svcPost, tc.svcErr)
			}

			h := NewPostHandler(svc, newLog())

			req := httptest.NewRequest(http.MethodPost, "/", nil)
			if tc.vars != nil {
				req = mux.SetURLVars(req, tc.vars)
			}
			if tc.withSess {
				req = req.WithContext(sessions.SessionToCtx(req.Context(), sess))
			}
			rw := &badRW{httptest.NewRecorder(), tc.failWrite}

			h.UnVote(rw, req)

			if rw.Code != tc.wantedStatusCode {
				t.Fatalf("code %d want %d", rw.Code, tc.wantedStatusCode)
			}
		})
	}
}
