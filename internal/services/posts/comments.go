package posts

import (
	"context"
	"fmt"
	"redditclone/internal/models"
	"redditclone/internal/sessions"
	"redditclone/pkg/helpers"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (s *PostService) AddComment(ctx context.Context, postID string, commentMsg string, ss *sessions.Session) (*models.Post, error) {

	postObjID, err := primitive.ObjectIDFromHex(postID)
	if err != nil {
		return nil, fmt.Errorf("invalid postID: %w", err)
	}
	var comment = models.Comment{
		Created: helpers.GetTime(),
		Author:  ss.User,
		Body:    commentMsg,
		ID:      primitive.NewObjectID(),
		PostID:  postObjID,
	}
	err = s.Repo.AddCom(ctx, &comment)
	if err != nil {
		return nil, err
	}

	return s.GetPostByID(ctx, postID)
}

func (s *PostService) DeleteComment(ctx context.Context, postID, commentID string) (*models.Post, error) {

	pID, err := primitive.ObjectIDFromHex(postID)
	if err != nil {
		return nil, fmt.Errorf("invalid postID: %w", err)
	}

	commID, err := primitive.ObjectIDFromHex(commentID)
	if err != nil {
		return nil, fmt.Errorf("invalid postID: %w", err)
	}

	err = s.Repo.DeleteCom(ctx, pID, commID)
	if err != nil {
		return nil, err
	}

	return s.GetPostByID(ctx, postID)
}
