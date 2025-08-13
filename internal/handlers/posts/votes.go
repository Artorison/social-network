package posts

import (
	"errors"
	"log/slog"
	"net/http"
	"redditclone/internal/models"
	"redditclone/internal/sessions"
	"redditclone/pkg/logger"

	"github.com/gorilla/mux"
)

func (h *PostsHandler) UppVote(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	postID, ok := vars[ParamPostID]
	if !ok {
		models.JSONError(w, http.StatusBadRequest, "invalid post id")
		return
	}

	session, ok := sessions.GetSessionFromCtx(r.Context())
	if !ok {
		models.JSONError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	h.Logger.Info("UppVote", slog.String(ParamPostID, postID))
	post, err := h.Service.UpVote(r.Context(), postID, session)
	if errors.Is(err, models.ErrPostNotFound) {
		models.JSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err != nil {
		models.JSONError(w, http.StatusInternalServerError, "internal")
		h.Logger.Error("upvote get post", logger.Err(err))
		return
	}
	models.SendJSONResponse(w, http.StatusOK, post, h.Logger)
}

func (h *PostsHandler) DownVote(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	postID, ok := vars[ParamPostID]
	if !ok {
		models.JSONError(w, http.StatusBadRequest, "invalid post id")
		return
	}

	session, ok := sessions.GetSessionFromCtx(r.Context())
	if !ok {
		models.JSONError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	h.Logger.Info("DownVote", slog.String(ParamPostID, postID))
	post, err := h.Service.DownVote(r.Context(), postID, session)
	if errors.Is(err, models.ErrPostNotFound) {
		models.JSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err != nil {
		models.JSONError(w, http.StatusInternalServerError, "internal")
		h.Logger.Error("downvote get post", logger.Err(err))
		return
	}
	models.SendJSONResponse(w, http.StatusOK, post, h.Logger)
}

func (h *PostsHandler) UnVote(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	postID, ok := vars[ParamPostID]
	if !ok {
		models.JSONError(w, http.StatusBadRequest, "invalid post id")
		return
	}

	session, ok := sessions.GetSessionFromCtx(r.Context())
	if !ok || session == nil {
		models.JSONError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	h.Logger.Info("UnVote", slog.String(ParamPostID, postID))
	post, err := h.Service.UnVote(r.Context(), postID, session)
	if errors.Is(err, models.ErrPostNotFound) {
		models.JSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err != nil {
		models.JSONError(w, http.StatusInternalServerError, "internal")
		h.Logger.Error("unvote get post", logger.Err(err))
		return
	}
	models.SendJSONResponse(w, http.StatusOK, post, h.Logger)
}
