package posts

import (
	"context"
	"fmt"

	"github.com/Artorison/social-network/internal/models"
	"github.com/Artorison/social-network/internal/sessions"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (s *PostService) UpVote(ctx context.Context, postID string, ss *sessions.Session) (*models.Post, error) {
	vote := models.Vote{
		Vote:   1,
		UserID: ss.User.ID,
	}
	pID, err := primitive.ObjectIDFromHex(postID)
	if err != nil {
		return nil, fmt.Errorf("invalid postID: %w", err)
	}

	return s.Repo.ChangeVote(ctx, pID, &vote)
}

func (s *PostService) DownVote(ctx context.Context, postID string, ss *sessions.Session) (*models.Post, error) {
	vote := models.Vote{
		Vote:   -1,
		UserID: ss.User.ID,
	}
	pID, err := primitive.ObjectIDFromHex(postID)
	if err != nil {
		return nil, fmt.Errorf("invalid postID: %w", err)
	}

	return s.Repo.ChangeVote(ctx, pID, &vote)
}

func (s *PostService) UnVote(ctx context.Context, postID string, ss *sessions.Session) (*models.Post, error) {
	pID, err := primitive.ObjectIDFromHex(postID)
	if err != nil {
		return nil, fmt.Errorf("invalid postID: %w", err)
	}
	return s.Repo.UnVote(ctx, pID, ss.User.ID)
}
