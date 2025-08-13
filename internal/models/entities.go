package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	Username     string `json:"username" bson:"username"`
	ID           string `json:"id" bson:"_id"`
	PasswordHash string `json:"-" bson:"-"`
}

type Author struct {
	ID       string `json:"id" bson:"id"`
	Username string `json:"username" bson:"username"`
}

type Post struct {
	Score int    `json:"score" bson:"score"`
	Views int    `json:"views" bson:"views"`
	Type  string `json:"type" bson:"type"`
	Title string `json:"title" bson:"title"`
	URL   string `json:"url,omitempty" bson:"url,omitempty"`
	Text  string `json:"text,omitempty" bson:"text,omitempty"`

	Author *Author `json:"author" bson:"author"`

	Category string `json:"category" bson:"category"`

	Votes    []*Vote    `json:"votes" bson:"votes"`
	Comments []*Comment `json:"comments" bson:"comments"`

	CreatedAt        time.Time          `json:"created" bson:"created"`
	UpvotePercentage int                `json:"upvotePercentage" bson:"upvotePercentage"`
	ID               primitive.ObjectID `json:"id" bson:"_id"`
}

type Comment struct {
	Created time.Time          `json:"created" bson:"created"`
	Author  *User              `json:"author" bson:"author"`
	Body    string             `json:"body" bson:"body"`
	ID      primitive.ObjectID `json:"id" bson:"_id"`
	PostID  primitive.ObjectID `json:"-" bson:"post_id"`
}

type Vote struct {
	UserID string `json:"user" bson:"user_id"`
	Vote   int    `json:"vote" bson:"vote"`
}
