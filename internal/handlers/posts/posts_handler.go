package posts

import (
	"context"
	"redditclone/internal/models"
	"redditclone/internal/sessions"
	"redditclone/pkg/logger"
)

type PostsService interface {
	GetAllPosts(ctx context.Context) ([]*models.Post, error)
	CreatePost(ctx context.Context, dto models.CreatePostDTO, ss *sessions.Session) (*models.Post, error)
	GetPostsByCategory(ctx context.Context, category string) ([]*models.Post, error)
	GetPostByID(ctx context.Context, postID string) (*models.Post, error)
	GetUserPosts(ctx context.Context, username string) ([]*models.Post, error)
	DeletePost(ctx context.Context, postID string) error

	AddComment(ctx context.Context, postID string, commentMsg string, ss *sessions.Session) (*models.Post, error)
	DeleteComment(ctx context.Context, postID, commentID string) (*models.Post, error)

	UpVote(ctx context.Context, postID string, ss *sessions.Session) (*models.Post, error)
	DownVote(ctx context.Context, postID string, ss *sessions.Session) (*models.Post, error)
	UnVote(ctx context.Context, postID string, ss *sessions.Session) (*models.Post, error)
}

type PostsHandler struct {
	Service PostsService
	Logger  *logger.Logger
}

func NewPostHandler(service PostsService, logger *logger.Logger) *PostsHandler {
	return &PostsHandler{
		Service: service,
		Logger:  logger,
	}
}
