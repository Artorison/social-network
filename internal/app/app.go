package app

import (
	"context"
	"net/http"
	"redditclone/config"
	"redditclone/database/mongodb"
	"redditclone/database/mysqldb"
	postshandler "redditclone/internal/handlers/posts"
	usershandlers "redditclone/internal/handlers/users"
	"redditclone/internal/middleware"
	"redditclone/internal/repo/posts/postmongo"
	usersmysql "redditclone/internal/repo/users/my_sql"
	postservice "redditclone/internal/services/posts"
	userservice "redditclone/internal/services/users"
	"redditclone/internal/sessions"
	"redditclone/pkg/logger"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/mongo"
)

func Run(cfg *config.Config, l *logger.Logger) {

	router := mux.NewRouter()

	ctx, cansel := context.WithTimeout(context.Background(), time.Second*10)
	defer cansel()

	mongoClient := mongodb.InitMongoDB(ctx, cfg.MongoAddr)

	app := NewApp(cfg, router, l, mongoClient)

	app.Start()
}

type App struct {
	Cfg    *config.Config
	Logger *logger.Logger
	Router *mux.Router
	Mongo  *mongo.Client
}

func NewApp(cfg *config.Config, router *mux.Router, logger *logger.Logger, mongo *mongo.Client) *App {
	return &App{
		Cfg:    cfg,
		Router: router,
		Logger: logger,
		Mongo:  mongo,
	}
}

func (s *App) Start() {
	s.LoadStatic()
	s.NewRealisation()
	s.Logger.Info("Server listening on http://localhost:" + s.Cfg.AppPort)
	if err := http.ListenAndServe(":"+s.Cfg.AppPort, s.Router); err != nil {
		s.Logger.Error("server is crash", logger.Err(err))
	}
}

func (s *App) LoadStatic() {
	staticFS := http.FileServer(http.Dir("static/"))
	s.Router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", staticFS))
	s.Logger.Info("register static")
}

func (s *App) NewRealisation() {
	s.Router.Use(middleware.AccessLog(s.Logger))
	s.Router.Use(middleware.PanicRecover())

	myslqDB := mysqldb.InitMysql(s.Cfg.MysqlDSN)
	sm := sessions.NewSessionManager(myslqDB)
	usersRepo := usersmysql.NewUsersMysqlRepo(myslqDB)
	userService := userservice.NewUserService(usersRepo, sm)
	userHandler := usershandlers.NewUserHandler(userService, s.Logger)

	api := s.Router.PathPrefix("/api").Subrouter()

	api.HandleFunc("/register", userHandler.Register).Methods(http.MethodPost)
	api.HandleFunc("/login", userHandler.Login).Methods(http.MethodPost)

	postRepo := postmongo.NewModgoDB(s.Mongo, s.Cfg.MongoDBName)
	postService := postservice.NewPostService(postRepo)
	postHandler := postshandler.NewPostHandler(postService, s.Logger)

	api.HandleFunc("/posts/", postHandler.GetAllPosts).Methods(http.MethodGet)
	api.HandleFunc("/user/{username}", postHandler.GetPostsByUser).Methods(http.MethodGet)
	api.HandleFunc("/posts/{category}", postHandler.GetPostByCategory).Methods(http.MethodGet)
	api.HandleFunc("/post/{post_id}", postHandler.GetPostByID).Methods(http.MethodGet)

	withAuth := api.NewRoute().Subrouter()
	withAuth.Use(middleware.Auth(sm))
	withAuth.HandleFunc("/posts", postHandler.CreatePost).Methods(http.MethodPost)
	withAuth.HandleFunc("/post/{post_id}", postHandler.AddComment).Methods(http.MethodPost)
	withAuth.HandleFunc("/post/{post_id}/{comment_id}", postHandler.DeleteComment).Methods(http.MethodDelete)
	withAuth.HandleFunc("/post/{post_id}/upvote", postHandler.UppVote).Methods(http.MethodGet)
	withAuth.HandleFunc("/post/{post_id}/downvote", postHandler.DownVote).Methods(http.MethodGet)
	withAuth.HandleFunc("/post/{post_id}/unvote", postHandler.UnVote).Methods(http.MethodGet)
	withAuth.HandleFunc("/post/{post_id}", postHandler.DeletePost).Methods(http.MethodDelete)

	s.Router.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, s.Cfg.HTMLTemplatePath)
	})
}
