package posts

import (
	"context"

	"github.com/Artorison/social-network/internal/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PostsRepo interface {
	CreatePost(ctx context.Context, post *models.Post) error
	GetAllPosts(ctx context.Context) ([]*models.Post, error)
	GetPostByID(ctx context.Context, postID primitive.ObjectID) (
		*models.Post, error)

	GetPostsByCategory(ctx context.Context, category string) (
		[]*models.Post, error)

	DeletePost(ctx context.Context, postID primitive.ObjectID) error
	GetUsersPosts(ctx context.Context, username string) (
		[]*models.Post, error)

	AddCom(ctx context.Context, newCom *models.Comment) error
	DeleteCom(ctx context.Context, postID, commentID primitive.ObjectID) error

	ChangeVote(ctx context.Context, postID primitive.ObjectID,
		vote *models.Vote) (*models.Post, error)

	UnVote(ctx context.Context, postID primitive.ObjectID,
		userID string) (*models.Post, error)
}

type PostService struct {
	Repo PostsRepo
}

func NewPostService(repo PostsRepo) *PostService {
	return &PostService{Repo: repo}
}
