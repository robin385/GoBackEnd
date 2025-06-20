package main

import (
	"log"
	"os"

	"gobackend/database"
	"gobackend/handlers"
	"gobackend/models"

	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/prefork"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./data/app.db"
	}

	db, err := database.Initialize(dbPath)
	if err != nil {
		log.Fatalf("failed to init db: %v", err)
	}

	userRepo := models.NewUserRepository(db.DB)
	trashRepo := models.NewTrashPostRepository(db.DB)
	commentRepo := models.NewCommentRepository(db.DB)

	userHandler := handlers.NewUserHandler(userRepo)
	trashHandler := handlers.NewTrashPostHandler(trashRepo, userRepo)
	commentHandler := handlers.NewCommentHandler(commentRepo, userRepo, trashRepo)
	oauthHandler := handlers.NewOAuthHandler(userRepo)

	r := router.New()
	r.GET("/health", func(ctx *fasthttp.RequestCtx) {
		ctx.Response.Header.Set("Content-Type", "application/json")
		ctx.SetStatusCode(fasthttp.StatusOK)
		ctx.Write([]byte(`{"status":"ok"}`))
	})

	r.POST("/users", userHandler.CreateUser)
	r.POST("/login", userHandler.Login)
	r.GET("/leaderboard", userHandler.Leaderboard)
	r.GET("/auth/google/login", oauthHandler.Login)
	r.GET("/auth/google/callback", oauthHandler.Callback)
	r.ServeFiles("/uploads/{filepath:*}", "./uploads")
	r.POST("/trashposts", trashHandler.CreateTrashPost)
	r.GET("/trashposts", trashHandler.GetTrashPosts)
	r.DELETE("/trashposts/{id}", trashHandler.DeleteTrashPost)
	r.POST("/trashposts/{id}/comments", commentHandler.CreateComment)
	r.GET("/trashposts/{id}/comments", commentHandler.GetComments)

	server := &fasthttp.Server{Handler: r.Handler}

	if err := prefork.New(server).ListenAndServe(":" + port); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
