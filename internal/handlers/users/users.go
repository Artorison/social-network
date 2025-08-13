package users

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"redditclone/internal/models"
	"redditclone/pkg/logger"
)

func (h *UsersHandler) Register(w http.ResponseWriter, r *http.Request) {

	if r.Header.Get("Content-Type") != "application/json" {
		models.JSONError(w, http.StatusBadRequest, "unknown payload")
		return
	}

	var req models.LoginForm
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.Logger.Error("decode JSON", logger.Err(err))
		return
	}

	h.Logger.Info("register user", slog.String("username", req.Username))
	token, err := h.Service.RegisterUser(r.Context(), req.Username, req.Password)
	if errors.Is(err, models.ErrUserExists) {
		resp := models.ValidationErrors{
			Errors: []models.FieldError{{
				Location: "body",
				Param:    "username",
				Value:    req.Username,
				Msg:      err.Error(),
			}},
		}
		models.FieldErr(w, http.StatusUnprocessableEntity, resp)
		return
	}
	if err != nil {
		models.JSONError(w, http.StatusInternalServerError, "internal")
		h.Logger.Error("register get token", logger.Err(err))
		return
	}

	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(models.RequestToken{Token: token})
	if err != nil {
		http.Error(w, "JSON encoder error", http.StatusInternalServerError)
		h.Logger.Error("register encode JSON", logger.Err(err))
		return
	}
}

func (h *UsersHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != "application/json" {
		models.JSONError(w, http.StatusBadRequest, "unknown payload")
		return
	}

	var req models.LoginForm
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		models.JSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.Logger.Info("login user", slog.String("username", req.Username))
	token, err := h.Service.LoginUser(r.Context(), req.Username, req.Password)
	if errors.Is(err, models.ErrUserNotFound) {
		models.JSONError(w, http.StatusUnauthorized, err.Error())
		return
	}
	if errors.Is(err, models.ErrInvalidLogin) {
		models.JSONError(w, http.StatusUnauthorized, err.Error())
		return
	}
	if err != nil {
		models.JSONError(w, http.StatusInternalServerError, "internal")
		h.Logger.Error("login user get token", logger.Err(err))
		return
	}

	w.Header().Set("Content-type", "application/json")
	err = json.NewEncoder(w).Encode(models.RequestToken{Token: token})
	if err != nil {
		http.Error(w, "JSON encoder error", http.StatusInternalServerError)
		h.Logger.Error("login encode JSON", logger.Err(err))
		return
	}
}
