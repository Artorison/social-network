package app

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Artorison/social-network/config"
	"github.com/Artorison/social-network/database/mongodb"
	"github.com/Artorison/social-network/database/mysqldb"
	postshandler "github.com/Artorison/social-network/internal/handlers/posts"
	usershandlers "github.com/Artorison/social-network/internal/handlers/users"
	"github.com/Artorison/social-network/internal/middleware"
	"github.com/Artorison/social-network/internal/repo/posts/postmongo"
	usersmysql "github.com/Artorison/social-network/internal/repo/users/my_sql"
	postservice "github.com/Artorison/social-network/internal/services/posts"
	userservice "github.com/Artorison/social-network/internal/services/users"
	"github.com/Artorison/social-network/internal/sessions"
	"github.com/Artorison/social-network/pkg/logger"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/mongo"
)

func Run(cfg *config.Config, l *logger.Logger) {
	router := mux.NewRouter()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

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
	s.InitDependencies()

	srv := &http.Server{
		Addr:              ":" + s.Cfg.AppPort,
		Handler:           s.Router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	s.Logger.Info("Server listening on http://localhost:" + s.Cfg.AppPort)

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.Logger.Error("server crashed", logger.Err(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		s.Logger.Error("graceful shutdown failed", logger.Err(err))
	} else {
		s.Logger.Info("server stopped gracefully")
	}
}

func (s *App) LoadStatic() {
	staticFS := http.FileServer(http.Dir("static/"))
	s.Router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", staticFS))
	s.Logger.Info("register static")
}

func (s *App) InitDependencies() {
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
