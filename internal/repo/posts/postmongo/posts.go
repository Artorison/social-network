package postmongo

import (
	"context"
	"fmt"

	"github.com/Artorison/social-network/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (r *PostMongoDB) CreatePost(ctx context.Context, post *models.Post) error {
	_, err := r.Posts.InsertOne(ctx, post)
	if err != nil {
		return fmt.Errorf("insert post: %w", err)
	}
	return nil
}

func (r *PostMongoDB) GetAllPosts(ctx context.Context) ([]*models.Post, error) {
	findOpts := options.Find().SetSort(bson.D{{Key: "score", Value: -1}})
	cursor, err := r.Posts.Find(ctx, bson.D{}, findOpts)
	if err != nil {
		return nil, fmt.Errorf("find post: %w", err)
	}
	defer cursor.Close(ctx)

	posts := []*models.Post{}

	if err := cursor.All(ctx, &posts); err != nil {
		return nil, fmt.Errorf("cursor all posts: %w", err)
	}

	for _, post := range posts {
		counter, err := r.CountPostComments(ctx, post.ID)
		if err != nil {
			return nil, fmt.Errorf("postID: %s err: %w", post.ID.Hex(), err)
		}
		post.Comments = make([]*models.Comment, counter)
	}

	return posts, nil
}

func (r *PostMongoDB) GetPostByID(ctx context.Context, postID primitive.ObjectID) (
	*models.Post, error,
) {
	var post models.Post

	findFilter := bson.M{"_id": postID}
	updateOpts := bson.D{{Key: "$inc", Value: bson.D{{Key: "views", Value: 1}}}}
	findOpts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	err := r.Posts.FindOneAndUpdate(ctx, findFilter, updateOpts, findOpts).Decode(&post)
	if err != nil {
		return nil, fmt.Errorf("find: %w", err)
	}

	comments, err := r.GetPostCom(ctx, postID)
	if err != nil {
		return nil, err
	}

	post.Comments = comments
	return &post, nil
}

func (r *PostMongoDB) GetPostsByCategory(ctx context.Context, category string) ([]*models.Post, error) {
	findFilter := bson.M{"category": category}
	findOpts := options.Find().SetSort(bson.D{{Key: "score", Value: -1}})
	cursor, err := r.Posts.Find(ctx, findFilter, findOpts)
	if err != nil {
		return nil, fmt.Errorf("find post: %w", err)
	}
	defer cursor.Close(ctx)

	posts := []*models.Post{}

	if err := cursor.All(ctx, &posts); err != nil {
		return nil, fmt.Errorf("cursor all posts: %w", err)
	}

	if err := r.AddCommentsToPost(ctx, posts); err != nil {
		return nil, err
	}
	return posts, nil
}

func (r *PostMongoDB) DeletePost(ctx context.Context, postID primitive.ObjectID) error {
	if _, err := r.Posts.DeleteOne(ctx, bson.M{"_id": postID}); err != nil {
		return fmt.Errorf("delete one: %w", err)
	}

	return nil
}

func (r *PostMongoDB) GetUsersPosts(ctx context.Context, username string) ([]*models.Post, error) {
	findFilter := bson.D{{Key: "author.username", Value: username}}
	findOpts := options.Find().SetSort(bson.D{{Key: "score", Value: -1}})
	cursor, err := r.Posts.Find(ctx, findFilter, findOpts)
	if err != nil {
		return nil, fmt.Errorf("find post: %w", err)
	}
	defer cursor.Close(ctx)

	posts := []*models.Post{}

	if err := cursor.All(ctx, &posts); err != nil {
		return nil, fmt.Errorf("cursor all posts: %w", err)
	}

	if err := r.AddCommentsToPost(ctx, posts); err != nil {
		return nil, err
	}
	return posts, nil
}

func (r *PostMongoDB) AddCommentsToPost(ctx context.Context, posts []*models.Post) error {
	for _, post := range posts {
		comments, err := r.GetPostCom(ctx, post.ID)
		if err != nil {
			return fmt.Errorf("add commets: %w", err)
		}
		post.Comments = comments
	}
	return nil
}
