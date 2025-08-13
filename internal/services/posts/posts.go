package posts

import (
	"context"
	"errors"
	"fmt"

	"github.com/Artorison/social-network/internal/models"
	"github.com/Artorison/social-network/internal/sessions"
	"github.com/Artorison/social-network/pkg/helpers"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var ErrPostNotFound = errors.New("post not exists")

func (s *PostService) GetAllPosts(ctx context.Context) ([]*models.Post, error) {
	return s.Repo.GetAllPosts(ctx)
}

func (s *PostService) CreatePost(
	ctx context.Context, dto models.CreatePostDTO, ss *sessions.Session,
) (*models.Post, error) {
	post := models.Post{
		Category: dto.Category,
		Title:    dto.Title,
		Text:     dto.Text,
		URL:      dto.URL,
		Type:     dto.Type,

		ID:               primitive.NewObjectID(),
		Score:            1,
		UpvotePercentage: 100,
		CreatedAt:        helpers.GetTime(),
		Views:            0,

		Comments: make([]*models.Comment, 0, 10),
		Author:   &models.Author{ID: ss.User.ID, Username: ss.User.Username},
		Votes:    []*models.Vote{{UserID: ss.User.ID, Vote: 1}},
	}
	err := s.Repo.CreatePost(ctx, &post)
	if err != nil {
		return nil, err
	}
	return &post, nil
}

func (s *PostService) GetPostsByCategory(
	ctx context.Context, category string,
) ([]*models.Post, error) {
	return s.Repo.GetPostsByCategory(ctx, category)
}

func (s *PostService) GetPostByID(
	ctx context.Context, postID string,
) (*models.Post, error) {
	pID, err := primitive.ObjectIDFromHex(postID)
	if err != nil {
		return nil, fmt.Errorf("invalid postID: %w", err)
	}
	return s.Repo.GetPostByID(ctx, pID)
}

func (s *PostService) GetUserPosts(
	ctx context.Context, username string,
) ([]*models.Post, error) {
	return s.Repo.GetUsersPosts(ctx, username)
}

func (s *PostService) DeletePost(
	ctx context.Context, postID string,
) error {
	pID, err := primitive.ObjectIDFromHex(postID)
	if err != nil {
		return fmt.Errorf("invalid postID: %w", err)
	}
	return s.Repo.DeletePost(ctx, pID)
}
