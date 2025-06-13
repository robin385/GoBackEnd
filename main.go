package main

import (
	"log"
	"net/http"
	"os"

	"gobackend/database"
	"gobackend/handlers"
	"gobackend/models"

	"github.com/gin-gonic/gin"
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

	userHandler := handlers.NewUserHandler(userRepo)
	trashHandler := handlers.NewTrashPostHandler(trashRepo, userRepo)

	r := gin.Default()
	r.GET("/health", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })

	r.POST("/users", userHandler.CreateUser)

	r.POST("/trashposts", trashHandler.CreateTrashPost)
	r.GET("/trashposts", trashHandler.GetTrashPosts)
	r.DELETE("/trashposts/:id", trashHandler.DeleteTrashPost)

	if err := r.Run(":" + port); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
