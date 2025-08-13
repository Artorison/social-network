package posts

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"redditclone/internal/models"
	"redditclone/internal/sessions"
	"redditclone/pkg/logger"

	"github.com/gorilla/mux"
)

func (h *PostsHandler) GetAllPosts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "application/json")
	posts, err := h.Service.GetAllPosts(r.Context())
	if err != nil {
		models.JSONError(w, http.StatusInternalServerError, "internal")
		h.Logger.Error("GetAllPosts", logger.Err(err))
		return
	}
	models.SendJSONResponse(w, http.StatusOK, posts, h.Logger)
}

func (h *PostsHandler) CreatePost(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "application/json")
	var dto models.CreatePostDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		models.JSONError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	session, ok := sessions.GetSessionFromCtx(r.Context())
	if !ok {
		models.JSONError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	h.Logger.Info("create post", slog.String("username", dto.Title))
	post, err := h.Service.CreatePost(r.Context(), dto, session)
	if err != nil {
		http.Error(w, "create post error", http.StatusInternalServerError)
		h.Logger.Error("create post internal", logger.Err(err))
		return
	}

	models.SendJSONResponse(w, http.StatusCreated, post, h.Logger)

}

func (h *PostsHandler) GetPostByCategory(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)

	category, ok := vars[ParamCategory]
	if !ok {
		models.JSONError(w, http.StatusBadRequest, "category not specified")
		return
	}

	h.Logger.Info("Get Post By Category", slog.String(ParamCategory, category))
	posts, err := h.Service.GetPostsByCategory(r.Context(), category)
	if err != nil {
		http.Error(w, "GetPostsByCategory error", http.StatusInternalServerError)
		h.Logger.Error("GetPostsByCategory", logger.Err(err))
		return
	}

	models.SendJSONResponse(w, http.StatusOK, posts, h.Logger)
}

func (h *PostsHandler) GetPostByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	postID, ok := vars[ParamPostID]
	if !ok {
		models.JSONError(w, http.StatusBadRequest, "invalid post id")
		return
	}

	h.Logger.Info("Get Post By ID", slog.String(ParamPostID, postID))
	post, err := h.Service.GetPostByID(r.Context(), postID)
	if errors.Is(err, models.ErrPostNotFound) {
		models.JSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err != nil {
		models.JSONError(w, http.StatusInternalServerError, "internal")
		h.Logger.Error("get post by id", logger.Err(err))
		return
	}
	models.SendJSONResponse(w, http.StatusOK, post, h.Logger)
}

func (h *PostsHandler) GetPostsByUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	username, ok := vars[ParamUsername]
	if !ok {
		models.JSONError(w, http.StatusBadRequest, "invalid username")
		return
	}

	h.Logger.Info("Get Post By User", slog.String(ParamUsername, username))
	posts, err := h.Service.GetUserPosts(r.Context(), username)
	if err != nil {
		models.JSONError(w, http.StatusInternalServerError, "internal")
		h.Logger.Error("GetPostsByUser", logger.Err(err))
		return
	}

	models.SendJSONResponse(w, http.StatusOK, posts, h.Logger)
}

func (h *PostsHandler) DeletePost(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	postID, ok := vars[ParamPostID]
	if !ok {
		models.JSONError(w, http.StatusBadRequest, "invalid post id")
		return
	}

	h.Logger.Info("Delete Post", slog.String(ParamPostID, postID))
	err := h.Service.DeletePost(r.Context(), postID)
	if err != nil {
		models.JSONError(w, http.StatusInternalServerError, "internal")
		h.Logger.Error("DeletePost", logger.Err(err))
		return
	}

	models.SendJSONResponse(w, http.StatusOK, map[string]string{"message": "success"}, h.Logger)
}
