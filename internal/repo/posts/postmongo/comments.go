package postmongo

import (
	"context"
	"fmt"
	"redditclone/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (r *PostMongoDB) AddCom(ctx context.Context, newCom *models.Comment) error {

	if _, err := r.Comments.InsertOne(ctx, newCom); err != nil {
		return fmt.Errorf("insertOne com: %w", err)
	}
	return nil
}

func (r *PostMongoDB) DeleteCom(ctx context.Context,
	postID, commentID primitive.ObjectID) error {

	res, err := r.Comments.DeleteOne(ctx, bson.M{"_id": commentID, "post_id": postID})
	if err != nil {
		return fmt.Errorf("deleteOne: %w", err)
	}
	if res.DeletedCount == 0 {
		return models.ErrCommentNotFound
	}

	return nil
}

func (r PostMongoDB) GetPostCom(ctx context.Context, postID primitive.ObjectID) (
	[]*models.Comment, error) {

	cursor, err := r.Comments.Find(ctx, bson.M{"post_id": postID})
	if err != nil {
		return nil, fmt.Errorf("find coms: %w", err)
	}

	defer cursor.Close(ctx)

	comments := make([]*models.Comment, 0, 10)

	for cursor.Next(ctx) {
		var c models.Comment

		if err := cursor.Decode(&c); err != nil {
			return nil, fmt.Errorf("decode com: %w", err)
		}
		comments = append(comments, &c)
	}
	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor err: %w", err)
	}
	return comments, nil
}

func (r *PostMongoDB) CountPostComments(ctx context.Context,
	postID primitive.ObjectID) (int, error) {
	count, err := r.Comments.CountDocuments(ctx, bson.M{"post_id": postID})
	if err != nil {
		return 0, fmt.Errorf("count docs: %w", err)
	}

	return int(count), nil
}
