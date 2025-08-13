package models

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"redditclone/pkg/logger"
)

var (
	ErrUserExists      = errors.New("already exists")
	ErrUserNotFound    = errors.New("user not found")
	ErrInvalidLogin    = errors.New("invalid login or password")
	ErrPostNotFound    = errors.New("post not found")
	ErrCommentNotFound = errors.New("comment not found")
)

var ErrInvalidToken = errors.New("invalid or expired token")

type ErrorResponse struct {
	Message string `json:"message"`
}

func JSONError(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(ErrorResponse{Message: msg}); err != nil {
		slog.Error("JSONError encode error", logger.Err(err))
	}
}

func FieldErr(w http.ResponseWriter, code int, err ValidationErrors) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if er := json.NewEncoder(w).Encode(err); er != nil {
		slog.Error("FieldErr encode error", logger.Err(er))
	}
}

type FieldError struct {
	Location string `json:"location"`
	Param    string `json:"param"`
	Value    string `json:"value"`
	Msg      string `json:"msg"`
}

type ValidationErrors struct {
	Errors []FieldError `json:"errors"`
}

func SendJSONResponse(
	w http.ResponseWriter, status int, payload any, log *logger.Logger,
) {
	w.Header().Set("Content-Type", "application/json")
	if status != http.StatusOK {
		w.WriteHeader(status)
	}
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		JSONError(w, http.StatusInternalServerError, "internal")
		log.Error("encode JSON", logger.Err(err))
		return
	}
}
