package posts

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/Artorison/social-network/internal/models"
	"github.com/Artorison/social-network/internal/sessions"
	"github.com/Artorison/social-network/pkg/logger"
	"github.com/gorilla/mux"
)

func (h *PostsHandler) AddComment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	postID, ok := vars[ParamPostID]
	if !ok {
		models.JSONError(w, http.StatusBadRequest, "invalid post id")
		return
	}

	var dto models.AddCommentDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		models.JSONError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	session, ok := sessions.GetSessionFromCtx(r.Context())
	if !ok || session == nil {
		models.JSONError(w, http.StatusUnauthorized, "no session in context")
		return
	}

	h.Logger.Info("add comment", slog.String("comment", dto.CommentMsg))
	post, err := h.Service.AddComment(r.Context(), postID, dto.CommentMsg, session)
	if errors.Is(err, models.ErrPostNotFound) {
		models.JSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err != nil {
		models.JSONError(w, http.StatusInternalServerError, "internal")
		h.Logger.Error("upvote get post", logger.Err(err))
		return
	}
	models.SendJSONResponse(w, http.StatusCreated, post, h.Logger)
}

func (h *PostsHandler) DeleteComment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	postID, ok := vars[ParamPostID]
	if !ok {
		models.JSONError(w, http.StatusBadRequest, "invalid post id")
		return
	}
	commentID, ok := vars[ParamCommentID]
	if !ok {
		models.JSONError(w, http.StatusBadRequest, "invalid comment id")
		return
	}

	post, err := h.Service.DeleteComment(r.Context(), postID, commentID)
	if errors.Is(err, models.ErrPostNotFound) {
		models.JSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err != nil {
		models.JSONError(w, http.StatusInternalServerError, "internal")
		h.Logger.Error("delete comment post", logger.Err(err))
		return
	}
	models.SendJSONResponse(w, http.StatusOK, post, h.Logger)
}
