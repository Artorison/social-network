package postmongo

import (
	"context"
	"errors"
	"testing"

	"github.com/Artorison/social-network/internal/models"
	"github.com/Artorison/social-network/pkg/helpers"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
)

var (
	ErrDB    = errors.New("db error")
	u1       = &models.Author{Username: "u1", ID: "u1"}
	u2       = &models.Author{Username: "u2", ID: "u2"}
	post1    = newPost(u1, "music")
	post2    = newPost(u2, "funny")
	postList = []*models.Post{post1, post2}
)

func newPost(author *models.Author, category string) *models.Post {
	return &models.Post{
		Category: category,
		Title:    "Title",
		Text:     "text123",
		Type:     "text",

		ID:               primitive.NewObjectID(),
		Score:            1,
		UpvotePercentage: 100,
		CreatedAt:        helpers.GetTime(),
		Views:            0,

		Comments: make([]*models.Comment, 0, 10),
		Author:   author,
		Votes:    []*models.Vote{{UserID: author.ID, Vote: 1}},
	}
}

func newComment(author *models.User, msg string, postID primitive.ObjectID) *models.Comment {
	return &models.Comment{
		Created: helpers.GetTime(),
		Author:  author,
		Body:    msg,
		ID:      primitive.NewObjectID(),
		PostID:  postID,
	}
}

func toDoc(v any, t *testing.T) bson.D {
	raw, err := bson.Marshal(v)
	require.NoError(t, err)
	var d bson.D
	require.NoError(t, bson.Unmarshal(raw, &d))
	return d
}

func toDocs(vs []*models.Post, t *testing.T) (out []bson.D) {
	for _, v := range vs {
		out = append(out, toDoc(v, t))
	}
	return
}

func oID() primitive.ObjectID {
	return primitive.NewObjectID()
}

func initMtest(t *testing.T) *mtest.T {
	return mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
}

func newDummyComment() *models.Comment {
	return &models.Comment{
		Created: helpers.GetTime(),
		Author: &models.User{
			Username:     "Artor",
			ID:           "124523454",
			PasswordHash: "aiuwfbwdiuanpdaw",
		},
		Body:   "msg",
		ID:     primitive.NewObjectID(),
		PostID: primitive.NewObjectID(),
	}
}

func TestAddComment(t *testing.T) {
	mt := initMtest(t)
	mt.Run("OK", func(mt *mtest.T) {
		postMongoDB := NewModgoDB(mt.Client, mt.DB.Name())
		mt.AddMockResponses(mtest.CreateSuccessResponse())

		err := postMongoDB.AddCom(context.Background(), newDummyComment())

		require.NoError(mt, err)
	})
	mt.Run("insert one error", func(mt *mtest.T) {
		mt.AddMockResponses(
			mtest.CreateCommandErrorResponse(mtest.CommandError{
				Message: ErrDB.Error(),
			}),
		)

		postMongoDB := NewModgoDB(mt.Client, mt.DB.Name())
		err := postMongoDB.AddCom(context.Background(), newDummyComment())
		require.Error(t, err)
		require.Contains(t, err.Error(), "insertOne com")
	})
}

func TestDeleteComment(t *testing.T) {
	mt := initMtest(t)

	postID := oID()
	commentID := oID()

	mt.Run("OK", func(mt *mtest.T) {
		mt.AddMockResponses(bson.D{
			{Key: "ok", Value: 1},
			{Key: "n", Value: 1},
		})

		postDB := NewModgoDB(mt.Client, mt.DB.Name())
		err := postDB.DeleteCom(context.Background(), postID, commentID)
		require.NoError(mt, err)
	})

	mt.Run("Err comment not found", func(mt *mtest.T) {
		mt.AddMockResponses(bson.D{
			{Key: "ok", Value: 1},
			{Key: "n", Value: 0},
		})

		postMongoDB := NewModgoDB(mt.Client, mt.DB.Name())
		err := postMongoDB.DeleteCom(context.Background(), postID, commentID)
		require.ErrorIs(mt, err, models.ErrCommentNotFound)
	})

	mt.Run("deleteOne error", func(mt *mtest.T) {
		mt.AddMockResponses(mtest.CreateCommandErrorResponse(
			mtest.CommandError{
				Message: ErrDB.Error(),
			},
		))
		postDB := NewModgoDB(mt.Client, mt.DB.Name())
		err := postDB.DeleteCom(context.Background(), postID, commentID)
		require.Error(mt, err)
		require.Contains(mt, err.Error(), "deleteOne")
	})
}

func TestCountPostComments(t *testing.T) {
	mt := initMtest(t)

	mt.Run("OK", func(mt *mtest.T) {
		ns := mt.DB.Name() + ".comments"
		mt.AddMockResponses(
			mtest.CreateCursorResponse(
				0,
				ns,
				mtest.FirstBatch,
				bson.D{{Key: "n", Value: int64(5)}},
			),
		)

		postDB := NewModgoDB(mt.Client, mt.DB.Name())
		count, err := postDB.CountPostComments(mt.Context(), oID())
		require.NoError(mt, err)
		require.Equal(mt, 5, count)
	})
	mt.Run("ERR", func(mt *mtest.T) {
		mt.AddMockResponses(
			mtest.CreateCommandErrorResponse(mtest.CommandError{Message: "fail"}),
		)
		postDB := NewModgoDB(mt.Client, mt.DB.Name())
		count, err := postDB.CountPostComments(mt.Context(), oID())
		require.Error(mt, err)
		require.Zero(mt, count)
	})
}

func TestGetPostCom(t *testing.T) {
	mt := initMtest(t)
	postID := oID()

	mt.Run("OK", func(mt *mtest.T) {
		ns := mt.DB.Name() + ".comments"
		doc1 := toDoc(newComment(&models.User{Username: "u1", ID: "u1"}, "msg1", postID), t)
		doc2 := toDoc(newComment(&models.User{Username: "u2", ID: "u2"}, "msg2", postID), t)
		mt.AddMockResponses(mtest.CreateCursorResponse(0, ns, mtest.FirstBatch, doc1, doc2))
		postDB := NewModgoDB(mt.Client, mt.DB.Name())
		cs, err := postDB.GetPostCom(mt.Context(), postID)
		require.NoError(mt, err)
		require.Len(mt, cs, 2)
		require.Equal(mt, "msg1", cs[0].Body)
		require.Equal(mt, "msg2", cs[1].Body)
	})

	mt.Run("find error", func(mt *mtest.T) {
		mt.AddMockResponses(mtest.CreateCommandErrorResponse(mtest.CommandError{Message: "fail"}))
		postDB := NewModgoDB(mt.Client, mt.DB.Name())
		cs, err := postDB.GetPostCom(mt.Context(), postID)
		require.Error(mt, err)
		require.Nil(mt, cs)
	})

	mt.Run("decode error", func(mt *mtest.T) {
		ns := mt.DB.Name() + ".comments"
		bad := bson.D{
			{Key: "_id", Value: oID()},
			{Key: "post_id", Value: "wrong"},
		}
		mt.AddMockResponses(mtest.CreateCursorResponse(0, ns, mtest.FirstBatch, bad))
		postDB := NewModgoDB(mt.Client, mt.DB.Name())
		cs, err := postDB.GetPostCom(mt.Context(), postID)
		require.Error(mt, err)
		require.Nil(mt, cs)
	})
	mt.Run("cursor error", func(mt *mtest.T) {
		ns := mt.DB.Name() + ".comments"

		doc := toDoc(newComment(&models.User{Username: "u1", ID: "u1"}, "msg", postID), t)
		first := mtest.CreateCursorResponse(1, ns, mtest.FirstBatch, doc)
		errGetMore := mtest.CreateCommandErrorResponse(mtest.CommandError{Message: "fail"})

		mt.AddMockResponses(first, errGetMore)

		postDB := NewModgoDB(mt.Client, mt.DB.Name())
		cs, err := postDB.GetPostCom(mt.Context(), postID)

		require.Error(mt, err)
		require.Nil(mt, cs)
	})
}

func TestCreatePost(t *testing.T) {
	mt := initMtest(t)

	mt.Run("cases", func(mt *mtest.T) {
		repo := NewModgoDB(mt.Client, mt.DB.Name())

		cases := []struct {
			name string
			mock func()
			exp  error
		}{
			{
				"ok",
				func() { mt.AddMockResponses(mtest.CreateSuccessResponse()) },
				nil,
			},
			{
				"err",
				func() {
					mt.AddMockResponses(mtest.CreateCommandErrorResponse(mtest.CommandError{Message: ErrDB.Error()}))
				},
				ErrDB,
			},
		}
		for _, c := range cases {
			t.Run(c.name, func(t *testing.T) {
				c.mock()
				err := repo.CreatePost(context.Background(), post1)

				if c.exp == nil {
					require.NoError(t, err)
				} else {
					var cmd mongo.CommandError
					if errors.As(err, &cmd) {
						require.Equal(t, c.exp.Error(), cmd.Message)
					} else {
						require.EqualError(t, err, c.exp.Error())
					}
				}
			})
		}
	})
}

func TestDeletePost(t *testing.T) {
	mt := initMtest(t)

	mt.Run("cases", func(mt *mtest.T) {
		repo := NewModgoDB(mt.Client, mt.DB.Name())

		cases := []struct {
			name string
			mock func()
			exp  error
		}{
			{
				"ok",
				func() { mt.AddMockResponses(mtest.CreateSuccessResponse()) },
				nil,
			},
			{
				"err",
				func() {
					mt.AddMockResponses(mtest.CreateCommandErrorResponse(mtest.CommandError{Message: ErrDB.Error()}))
				},
				ErrDB,
			},
		}

		for _, c := range cases {
			t.Run(c.name, func(t *testing.T) {
				c.mock()
				err := repo.DeletePost(context.Background(), post1.ID)

				if c.exp == nil {
					require.NoError(t, err)
				} else {
					var cmd mongo.CommandError
					if errors.As(err, &cmd) {
						require.Equal(t, c.exp.Error(), cmd.Message)
					} else {
						require.EqualError(t, err, c.exp.Error())
					}
				}
			})
		}
	})
}

func TestGetAllPosts(t *testing.T) {
	mt := initMtest(t)

	mt.Run("cases", func(mt *mtest.T) {
		repo := NewModgoDB(mt.Client, mt.DB.Name())
		ns := mt.DB.Name() + ".posts"

		cases := []struct {
			name string
			mock func()
			expE error
			expL int
		}{
			{
				"ok list",
				func() {
					first := mtest.CreateCursorResponse(1, ns, mtest.FirstBatch, toDoc(post1, t))
					next := mtest.CreateCursorResponse(1, ns, mtest.NextBatch, toDocs(postList[1:], t)...)
					kill := mtest.CreateCursorResponse(0, ns, mtest.NextBatch)
					cnt := mtest.CreateCursorResponse(0, ns, mtest.FirstBatch, bson.D{{Key: "n", Value: int64(0)}})
					mt.AddMockResponses(first, next, kill, cnt, cnt)
				},
				nil,
				len(postList),
			},
			{
				"err find",
				func() {
					mt.AddMockResponses(mtest.CreateCommandErrorResponse(mtest.CommandError{Message: ErrDB.Error()}))
				},
				ErrDB,
				0,
			},
			{
				"err cursor all",
				func() {
					bad := bson.D{{Key: "_id", Value: "bad"}}
					mt.AddMockResponses(
						mtest.CreateCursorResponse(0, ns, mtest.FirstBatch, bad),
					)
				},
				ErrDB,
				0,
			},
			{
				"err count comments",
				func() {
					mt.AddMockResponses(
						mtest.CreateCursorResponse(0, ns, mtest.FirstBatch, toDoc(post1, t)),
						mtest.CreateCommandErrorResponse(
							mtest.CommandError{Message: ErrDB.Error()}),
					)
				},
				ErrDB,
				0,
			},
		}

		for _, c := range cases {
			t.Run(c.name, func(t *testing.T) {
				c.mock()
				res, err := repo.GetAllPosts(context.Background())

				require.Len(t, res, c.expL)
				if c.expE == nil {
					require.NoError(t, err)
				} else {
					require.Error(t, err)
				}
			})
		}
	})
}

func TestGetUsersPosts(t *testing.T) {
	mt := initMtest(t)

	mt.Run("cases", func(mt *mtest.T) {
		repo := NewModgoDB(mt.Client, mt.DB.Name())
		nsPosts := mt.DB.Name() + ".posts"
		nsComments := mt.DB.Name() + ".comments"

		cases := []struct {
			name string
			mock func()
			expE error
			expL int
		}{
			{
				"ok list (2 поста, 0 комментариев)",
				func() {
					mt.AddMockResponses(
						mtest.CreateCursorResponse(
							0,
							nsPosts,
							mtest.FirstBatch,
							toDoc(post1, t),
							toDoc(post2, t),
						),
					)
					empty := mtest.CreateCursorResponse(0, nsComments, mtest.FirstBatch)
					mt.AddMockResponses(empty, empty)
				},
				nil,
				2,
			},
			{
				"find error",
				func() {
					mt.AddMockResponses(
						mtest.CreateCommandErrorResponse(
							mtest.CommandError{Message: ErrDB.Error()},
						),
					)
				},
				ErrDB,
				0,
			},
			{
				"err cursor all",
				func() {
					bad := bson.D{{Key: "_id", Value: "bad"}}
					mt.AddMockResponses(
						mtest.CreateCursorResponse(0, nsPosts, mtest.FirstBatch, bad),
					)
				},
				ErrDB,
				0,
			},
		}

		for _, c := range cases {
			t.Run(c.name, func(t *testing.T) {
				c.mock()
				res, err := repo.GetUsersPosts(context.Background(), u1.Username)

				require.Len(t, res, c.expL)
				if c.expE == nil {
					require.NoError(t, err)
				} else {
					require.Error(t, err)
				}
			})
		}
	})
}

func TestGetPostByID(t *testing.T) {
	mt := initMtest(t)

	mt.Run("cases", func(mt *mtest.T) {
		repo := NewModgoDB(mt.Client, mt.DB.Name())
		nsComments := mt.DB.Name() + ".comments"

		cases := []struct {
			name string
			mock func()
			expE error
		}{
			{
				"ok",
				func() {
					mt.AddMockResponses(
						mtest.CreateSuccessResponse(
							bson.E{Key: "value", Value: toDoc(post1, t)},
						),
					)
					mt.AddMockResponses(
						mtest.CreateCursorResponse(0, nsComments, mtest.FirstBatch),
					)
				},
				nil,
			},
			{
				"find error",
				func() {
					mt.AddMockResponses(
						mtest.CreateCommandErrorResponse(
							mtest.CommandError{Message: ErrDB.Error()},
						),
					)
				},
				ErrDB,
			},
		}

		for _, c := range cases {
			t.Run(c.name, func(t *testing.T) {
				c.mock()
				res, err := repo.GetPostByID(context.Background(), post1.ID)

				if c.expE == nil {
					require.NoError(t, err)
					require.NotNil(t, res)
				} else {
					require.Error(t, err)
					require.Nil(t, res)
				}
			})
		}
	})
}

func TestGetPostsByCategory(t *testing.T) {
	mt := initMtest(t)

	mt.Run("cases", func(mt *mtest.T) {
		repo := NewModgoDB(mt.Client, mt.DB.Name())
		nsPosts := mt.DB.Name() + ".posts"
		nsComments := mt.DB.Name() + ".comments"

		cases := []struct {
			name string
			mock func()
			expE error
			expL int
		}{
			{
				"ok list",
				func() {
					mt.AddMockResponses(
						mtest.CreateCursorResponse(
							0,
							nsPosts,
							mtest.FirstBatch,
							toDoc(post1, t), toDoc(post2, t),
						),
					)
					empty := mtest.CreateCursorResponse(0, nsComments, mtest.FirstBatch)
					mt.AddMockResponses(empty, empty)
				},
				nil,
				2,
			},
			{
				"find error",
				func() {
					mt.AddMockResponses(
						mtest.CreateCommandErrorResponse(
							mtest.CommandError{Message: ErrDB.Error()},
						),
					)
				},
				ErrDB,
				0,
			},
			{
				"err cursor all",
				func() {
					bad := bson.D{{Key: "_id", Value: "bad"}}
					mt.AddMockResponses(
						mtest.CreateCursorResponse(0, nsPosts, mtest.FirstBatch, bad),
					)
				},
				ErrDB,
				0,
			},
		}

		for _, c := range cases {
			t.Run(c.name, func(t *testing.T) {
				c.mock()
				res, err := repo.GetPostsByCategory(context.Background(), "music")

				require.Len(t, res, c.expL)
				if c.expE == nil {
					require.NoError(t, err)
				} else {
					require.Error(t, err)
				}
			})
		}
	})
}

func TestChangeVote(t *testing.T) {
	mt := initMtest(t)

	mt.Run("cases", func(mt *mtest.T) {
		repo := NewModgoDB(mt.Client, mt.DB.Name())
		nsPosts := mt.DB.Name() + ".posts"
		nsComments := mt.DB.Name() + ".comments"

		base := toDoc(post1, t)

		var existingDown models.Post
		raw, err := bson.Marshal(base)
		require.NoError(t, err)
		err = bson.Unmarshal(raw, &existingDown)
		require.NoError(t, err)
		existingDown.Votes = []*models.Vote{{UserID: "u1", Vote: -1}}
		existingDown.Score = -1
		existingDown.UpvotePercentage = 0
		docExistingDown := toDoc(existingDown, t)

		cases := []struct {
			name  string
			vote  *models.Vote
			mock  func()
			expE  error
			check func(*testing.T, *models.Post)
		}{
			{
				"add new vote",
				&models.Vote{UserID: "u2", Vote: 1},
				func() {
					mt.AddMockResponses(
						mtest.CreateCursorResponse(0, nsPosts, mtest.FirstBatch, base),
						mtest.CreateSuccessResponse(),
						mtest.CreateCursorResponse(0, nsComments, mtest.FirstBatch),
					)
				},
				nil,
				func(t *testing.T, p *models.Post) {
					require.Equal(t, 2, p.Score)
					require.Len(t, p.Votes, 2)
				},
			},
			{
				"change existing vote",
				&models.Vote{UserID: "u1", Vote: 1},
				func() {
					mt.AddMockResponses(
						mtest.CreateCursorResponse(0, nsPosts, mtest.FirstBatch, docExistingDown),
						mtest.CreateSuccessResponse(),
						mtest.CreateCursorResponse(0, nsComments, mtest.FirstBatch),
					)
				},
				nil,
				func(t *testing.T, p *models.Post) {
					require.Equal(t, 1, p.Score)
					require.Len(t, p.Votes, 1)
				},
			},
			{
				"post not found",
				&models.Vote{UserID: "x", Vote: 1},
				func() {
					mt.AddMockResponses(
						mtest.CreateCursorResponse(0, nsPosts, mtest.FirstBatch),
					)
				},
				models.ErrPostNotFound,
				nil,
			},
			{
				"update error",
				&models.Vote{UserID: "x", Vote: 1},
				func() {
					mt.AddMockResponses(
						mtest.CreateCursorResponse(0, nsPosts, mtest.FirstBatch, base),
						mtest.CreateCommandErrorResponse(mtest.CommandError{Message: ErrDB.Error()}),
					)
				},
				ErrDB,
				nil,
			},
			{
				"find error",
				&models.Vote{UserID: "any", Vote: 1},
				func() {
					mt.AddMockResponses(
						mtest.CreateCommandErrorResponse(
							mtest.CommandError{Message: ErrDB.Error()},
						),
					)
				},
				ErrDB,
				nil,
			},
		}

		for _, c := range cases {
			t.Run(c.name, func(t *testing.T) {
				c.mock()
				res, err := repo.ChangeVote(context.Background(), post1.ID, c.vote)

				if c.expE == nil {
					require.NoError(t, err)
					require.NotNil(t, res)
					if c.check != nil {
						c.check(t, res)
					}
				} else {
					require.Error(t, err)
					require.Nil(t, res)
				}
			})
		}
	})
}

func TestUnVote(t *testing.T) {
	mt := initMtest(t)

	mt.Run("cases", func(mt *mtest.T) {
		repo := NewModgoDB(mt.Client, mt.DB.Name())
		nsPosts := mt.DB.Name() + ".posts"
		nsComments := mt.DB.Name() + ".comments"

		start := *post1
		start.Votes = []*models.Vote{{UserID: "u1", Vote: 1}}
		start.Score = 1
		start.UpvotePercentage = 100
		docStart := toDoc(&start, t)

		emptyCom := mtest.CreateCursorResponse(0, nsComments, mtest.FirstBatch)

		cases := []struct {
			name  string
			uid   string
			mock  func()
			expE  error
			check func(*testing.T, *models.Post)
		}{
			{
				"remove existing vote",
				"u1",
				func() {
					mt.AddMockResponses(
						mtest.CreateCursorResponse(0, nsPosts, mtest.FirstBatch, docStart),
					)
					mt.AddMockResponses(mtest.CreateSuccessResponse())
					mt.AddMockResponses(emptyCom)
				},
				nil,
				func(t *testing.T, p *models.Post) {
					require.Equal(t, 0, p.Score)
					require.Len(t, p.Votes, 0)
				},
			},
			{
				"post not found",
				"no",
				func() {
					mt.AddMockResponses(
						mtest.CreateCursorResponse(0, nsPosts, mtest.FirstBatch),
					)
				},
				models.ErrPostNotFound,
				nil,
			},
			{
				"update error",
				"u1",
				func() {
					mt.AddMockResponses(
						mtest.CreateCursorResponse(0, nsPosts, mtest.FirstBatch, docStart),
					)
					mt.AddMockResponses(
						mtest.CreateCommandErrorResponse(mtest.CommandError{Message: ErrDB.Error()}),
					)
				},
				ErrDB,
				nil,
			},
			{
				"find error",
				"u1",
				func() {
					mt.AddMockResponses(
						mtest.CreateCommandErrorResponse(
							mtest.CommandError{Message: ErrDB.Error()},
						),
					)
				},
				ErrDB,
				nil,
			},
		}

		for _, c := range cases {
			t.Run(c.name, func(t *testing.T) {
				c.mock()
				res, err := repo.UnVote(context.Background(), post1.ID, c.uid)

				if c.expE == nil {
					require.NoError(t, err)
					require.NotNil(t, res)
					if c.check != nil {
						c.check(t, res)
					}
				} else {
					require.Error(t, err)
					require.Nil(t, res)
				}
			})
		}
	})
}
