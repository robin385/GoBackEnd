package handlers

import (
	"fmt"
	"image/jpeg"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/disintegration/imaging"
	"gobackend/models"

	"github.com/gin-gonic/gin"
)

// TrashPostHandler handles trash post endpoints
type TrashPostHandler struct {
	repo     *models.TrashPostRepository
	userRepo *models.UserRepository
}

func NewTrashPostHandler(repo *models.TrashPostRepository, userRepo *models.UserRepository) *TrashPostHandler {
	return &TrashPostHandler{repo: repo, userRepo: userRepo}
}

// CreateTrashPost adds a new trash post
func (h *TrashPostHandler) CreateTrashPost(c *gin.Context) {
	userID, err := strconv.Atoi(c.PostForm("user_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
		return
	}

	lat, err := strconv.ParseFloat(c.PostForm("latitude"), 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid latitude"})
		return
	}

	lon, err := strconv.ParseFloat(c.PostForm("longitude"), 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid longitude"})
		return
	}

	post := models.TrashPost{
		UserID:      userID,
		Latitude:    lat,
		Longitude:   lon,
		Description: c.PostForm("description"),
		Trail:       c.PostForm("trail"),
	}

	// Handle optional image
	file, err := c.FormFile("image")
	if err == nil {
		if path, err := saveCompressedImage(file); err == nil {
			post.ImagePath = path
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}

	user, err := h.userRepo.GetByID(post.UserID)
	if err != nil || user == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user"})
		return
	}

	if err := h.repo.Create(&post); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create post"})
		return
	}

	c.JSON(http.StatusCreated, post)
}

// GetTrashPosts returns posts between start and end datetime
func (h *TrashPostHandler) GetTrashPosts(c *gin.Context) {
	startStr := c.Query("start")
	endStr := c.Query("end")
	if startStr == "" || endStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "start and end required"})
		return
	}

	start, err := time.Parse(time.RFC3339, startStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start"})
		return
	}
	end, err := time.Parse(time.RFC3339, endStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end"})
		return
	}

	posts, err := h.repo.GetByDateRange(start, end)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get posts"})
		return
	}
	c.JSON(http.StatusOK, posts)
}

// DeleteTrashPost deletes a post if user is admin
func (h *TrashPostHandler) DeleteTrashPost(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	userID, err := strconv.Atoi(c.Query("userId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user"})
		return
	}
	user, err := h.userRepo.GetByID(userID)
	if err != nil || user == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user"})
		return
	}
	if !user.IsAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "admin required"})
		return
	}

	if err := h.repo.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete post"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

func saveCompressedImage(file *multipart.FileHeader) (string, error) {
	f, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("open file: %w", err)
	}
	defer f.Close()

	img, err := imaging.Decode(f)
	if err != nil {
		return "", fmt.Errorf("decode image: %w", err)
	}

	// Resize to 1080x1920 keeping aspect ratio by filling
	resized := imaging.Fill(img, 1080, 1920, imaging.Center, imaging.Lanczos)

	if err := os.MkdirAll("uploads", 0755); err != nil {
		return "", fmt.Errorf("create dir: %w", err)
	}

	filename := fmt.Sprintf("%d.jpg", time.Now().UnixNano())
	path := filepath.Join("uploads", filename)
	out, err := os.Create(path)
	if err != nil {
		return "", fmt.Errorf("create file: %w", err)
	}
	defer out.Close()

	if err := jpeg.Encode(out, resized, &jpeg.Options{Quality: 75}); err != nil {
		return "", fmt.Errorf("encode jpeg: %w", err)
	}

	return path, nil
}
