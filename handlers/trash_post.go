package handlers

import (
	"fmt"
	"image/jpeg"
	"mime/multipart"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/disintegration/imaging"
	"github.com/golang-jwt/jwt/v5"
	"github.com/valyala/fasthttp"

	"gobackend/models"
)

// TrashPostHandler handles trash post endpoints
type TrashPostHandler struct {
	repo     *models.TrashPostRepository
	userRepo *models.UserRepository
}

func NewTrashPostHandler(repo *models.TrashPostRepository, userRepo *models.UserRepository) *TrashPostHandler {
	return &TrashPostHandler{repo: repo, userRepo: userRepo}
}

func getUserIDFromToken(ctx *fasthttp.RequestCtx) (int, error) {
	header := string(ctx.Request.Header.Peek("Authorization"))
	if header == "" {
		return 0, fmt.Errorf("authorization header missing")
	}

	parts := strings.SplitN(header, " ", 2)
	tokenString := header
	if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
		tokenString = parts[1]
	}

	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(os.Getenv("JWT_SECRET")), nil
	})
	if err != nil || !token.Valid {
		return 0, fmt.Errorf("invalid token")
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, fmt.Errorf("invalid claims")
	}
	idVal, ok := claims["user_id"].(float64)
	if !ok {
		return 0, fmt.Errorf("user_id missing in token")
	}
	return int(idVal), nil
}

// CreateTrashPost adds a new trash post
func (h *TrashPostHandler) CreateTrashPost(ctx *fasthttp.RequestCtx) {
	userID, err := getUserIDFromToken(ctx)
	if err != nil {
		writeJSON(ctx, fasthttp.StatusUnauthorized, map[string]string{"error": err.Error()})
		return
	}

	lat, err := strconv.ParseFloat(string(ctx.FormValue("latitude")), 64)
	if err != nil {
		writeJSON(ctx, fasthttp.StatusBadRequest, map[string]string{"error": "invalid latitude"})
		return
	}

	lon, err := strconv.ParseFloat(string(ctx.FormValue("longitude")), 64)
	if err != nil {
		writeJSON(ctx, fasthttp.StatusBadRequest, map[string]string{"error": "invalid longitude"})
		return
	}

	post := models.TrashPost{
		UserID:      userID,
		Latitude:    lat,
		Longitude:   lon,
		Description: string(ctx.FormValue("description")),
		Trail:       string(ctx.FormValue("trail")),
	}

	file, err := ctx.FormFile("image")
	if err == nil {
		if path, err := saveCompressedImage(file); err == nil {
			post.ImagePath = path
		} else {
			writeJSON(ctx, fasthttp.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
	}

	user, err := h.userRepo.GetByID(post.UserID)
	if err != nil || user == nil {
		writeJSON(ctx, fasthttp.StatusBadRequest, map[string]string{"error": "invalid user"})
		return
	}

	if err := h.repo.Create(&post); err != nil {
		writeJSON(ctx, fasthttp.StatusInternalServerError, map[string]string{"error": "failed to create post"})
		return
	}

	writeJSON(ctx, fasthttp.StatusCreated, post)
}

// GetTrashPosts returns posts between start and end datetime
func (h *TrashPostHandler) GetTrashPosts(ctx *fasthttp.RequestCtx) {
	startStr := string(ctx.QueryArgs().Peek("start"))
	endStr := string(ctx.QueryArgs().Peek("end"))
	if startStr == "" || endStr == "" {
		writeJSON(ctx, fasthttp.StatusBadRequest, map[string]string{"error": "start and end required"})
		return
	}

	start, err := time.Parse(time.RFC3339, startStr)
	if err != nil {
		writeJSON(ctx, fasthttp.StatusBadRequest, map[string]string{"error": "invalid start"})
		return
	}
	end, err := time.Parse(time.RFC3339, endStr)
	if err != nil {
		writeJSON(ctx, fasthttp.StatusBadRequest, map[string]string{"error": "invalid end"})
		return
	}

	posts, err := h.repo.GetByDateRange(start, end)
	if err != nil {
		writeJSON(ctx, fasthttp.StatusInternalServerError, map[string]string{"error": "failed to get posts"})
		return
	}
	writeJSON(ctx, fasthttp.StatusOK, posts)
}

// DeleteTrashPost deletes a post if user is admin
func (h *TrashPostHandler) DeleteTrashPost(ctx *fasthttp.RequestCtx) {
	idStr := ctx.UserValue("id").(string)
	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeJSON(ctx, fasthttp.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}
	userID, err := strconv.Atoi(string(ctx.QueryArgs().Peek("userId")))
	if err != nil {
		writeJSON(ctx, fasthttp.StatusBadRequest, map[string]string{"error": "invalid user"})
		return
	}
	user, err := h.userRepo.GetByID(userID)
	if err != nil || user == nil {
		writeJSON(ctx, fasthttp.StatusBadRequest, map[string]string{"error": "invalid user"})
		return
	}
	if !user.IsAdmin {
		writeJSON(ctx, fasthttp.StatusForbidden, map[string]string{"error": "admin required"})
		return
	}

	if err := h.repo.Delete(id); err != nil {
		writeJSON(ctx, fasthttp.StatusInternalServerError, map[string]string{"error": "failed to delete post"})
		return
	}
	writeJSON(ctx, fasthttp.StatusOK, map[string]string{"message": "deleted"})
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
