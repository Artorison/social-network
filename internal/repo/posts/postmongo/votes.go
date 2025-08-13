package postmongo

import (
	"context"
	"errors"
	"fmt"
	"math"
	"slices"

	"github.com/Artorison/social-network/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func (r *PostMongoDB) ChangeVote(ctx context.Context, postID primitive.ObjectID,
	vote *models.Vote,
) (*models.Post, error) {
	var post models.Post
	if err := r.Posts.FindOne(ctx, bson.M{"_id": postID}).Decode(&post); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, models.ErrPostNotFound
		}
		return nil, fmt.Errorf("find post: %w", err)
	}

	found := false
	for i, v := range post.Votes {
		if v.UserID == vote.UserID {
			post.Score -= v.Vote
			post.Votes[i].Vote = vote.Vote
			found = true
			break
		}
	}
	if !found {
		post.Votes = append(post.Votes, &models.Vote{
			UserID: vote.UserID,
			Vote:   vote.Vote,
		})
	}
	post.Score += vote.Vote
	post.UpvotePercentage = CalculateUpvotePercentage(post.Votes)

	update := bson.M{"$set": bson.M{
		"votes":            post.Votes,
		"score":            post.Score,
		"upvotePercentage": post.UpvotePercentage,
	}}

	if _, err := r.Posts.UpdateOne(ctx, bson.M{"_id": postID}, update); err != nil {
		return nil, fmt.Errorf("update vote: %w", err)
	}

	comments, err := r.GetPostCom(ctx, postID)
	if err != nil {
		return nil, err
	}
	post.Comments = comments

	return &post, nil
}

func (r *PostMongoDB) UnVote(ctx context.Context, postID primitive.ObjectID,
	userID string,
) (*models.Post, error) {
	var post models.Post
	if err := r.Posts.FindOne(ctx, bson.M{"_id": postID}).Decode(&post); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, models.ErrPostNotFound
		}
		return nil, fmt.Errorf("find post: %w", err)
	}

	for i, v := range post.Votes {
		if v.UserID == userID {
			post.Score -= v.Vote
			post.Votes = slices.Delete(post.Votes, i, i+1)
			break
		}
	}
	post.UpvotePercentage = CalculateUpvotePercentage(post.Votes)

	update := bson.M{"$set": bson.M{
		"votes":            post.Votes,
		"score":            post.Score,
		"upvotePercentage": post.UpvotePercentage,
	}}
	if _, err := r.Posts.UpdateOne(ctx, bson.M{"_id": postID}, update); err != nil {
		return nil, fmt.Errorf("update unvote: %w", err)
	}

	comments, err := r.GetPostCom(ctx, postID)
	if err != nil {
		return nil, err
	}
	post.Comments = comments

	return &post, nil
}

func CalculateUpvotePercentage(votes []*models.Vote) int {
	var up int
	total := len(votes)
	if total == 0 {
		return 0
	}
	for _, v := range votes {
		if v.Vote > 0 {
			up++
		}
	}
	return int(math.Round(float64(up * 100 / total)))
}
