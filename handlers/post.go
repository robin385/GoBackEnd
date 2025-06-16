package handlers

import (
	"strconv"

	"gobackend/models"

	"github.com/valyala/fasthttp"
)

type PostHandler struct {
	postRepo *models.PostRepository
	userRepo *models.UserRepository
}

func NewPostHandler(postRepo *models.PostRepository, userRepo *models.UserRepository) *PostHandler {
	return &PostHandler{
		postRepo: postRepo,
		userRepo: userRepo,
	}
}

// CreatePost creates a new post
func (h *PostHandler) CreatePost(ctx *fasthttp.RequestCtx) {
	var post models.Post
	if err := readJSON(ctx, &post); err != nil {
		writeJSON(ctx, fasthttp.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	user, err := h.userRepo.GetByID(post.UserID)
	if err != nil {
		writeJSON(ctx, fasthttp.StatusInternalServerError, map[string]string{"error": "Failed to verify user"})
		return
	}
	if user == nil {
		writeJSON(ctx, fasthttp.StatusBadRequest, map[string]string{"error": "User not found"})
		return
	}

	if err := h.postRepo.Create(&post); err != nil {
		writeJSON(ctx, fasthttp.StatusInternalServerError, map[string]string{"error": "Failed to create post"})
		return
	}

	createdPost, err := h.postRepo.GetByID(post.ID)
	if err != nil {
		writeJSON(ctx, fasthttp.StatusInternalServerError, map[string]string{"error": "Failed to get created post"})
		return
	}

	writeJSON(ctx, fasthttp.StatusCreated, createdPost)
}

// GetPost retrieves a post by ID
func (h *PostHandler) GetPost(ctx *fasthttp.RequestCtx) {
	idStr := ctx.UserValue("id").(string)
	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeJSON(ctx, fasthttp.StatusBadRequest, map[string]string{"error": "Invalid post ID"})
		return
	}

	post, err := h.postRepo.GetByID(id)
	if err != nil {
		writeJSON(ctx, fasthttp.StatusInternalServerError, map[string]string{"error": "Failed to get post"})
		return
	}

	if post == nil {
		writeJSON(ctx, fasthttp.StatusNotFound, map[string]string{"error": "Post not found"})
		return
	}

	writeJSON(ctx, fasthttp.StatusOK, post)
}

// GetPosts retrieves all posts
func (h *PostHandler) GetPosts(ctx *fasthttp.RequestCtx) {
	posts, err := h.postRepo.GetAll()
	if err != nil {
		writeJSON(ctx, fasthttp.StatusInternalServerError, map[string]string{"error": "Failed to get posts"})
		return
	}

	writeJSON(ctx, fasthttp.StatusOK, posts)
}

// GetPostsByUser retrieves all posts by a specific user
func (h *PostHandler) GetPostsByUser(ctx *fasthttp.RequestCtx) {
	userIDStr := ctx.UserValue("userId").(string)
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		writeJSON(ctx, fasthttp.StatusBadRequest, map[string]string{"error": "Invalid user ID"})
		return
	}

	posts, err := h.postRepo.GetByUserID(userID)
	if err != nil {
		writeJSON(ctx, fasthttp.StatusInternalServerError, map[string]string{"error": "Failed to get posts"})
		return
	}

	writeJSON(ctx, fasthttp.StatusOK, posts)
}

// UpdatePost updates a post
func (h *PostHandler) UpdatePost(ctx *fasthttp.RequestCtx) {
	idStr := ctx.UserValue("id").(string)
	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeJSON(ctx, fasthttp.StatusBadRequest, map[string]string{"error": "Invalid post ID"})
		return
	}

	var post models.Post
	if err := readJSON(ctx, &post); err != nil {
		writeJSON(ctx, fasthttp.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	post.ID = id
	if err := h.postRepo.Update(&post); err != nil {
		writeJSON(ctx, fasthttp.StatusInternalServerError, map[string]string{"error": "Failed to update post"})
		return
	}

	updatedPost, err := h.postRepo.GetByID(id)
	if err != nil {
		writeJSON(ctx, fasthttp.StatusInternalServerError, map[string]string{"error": "Failed to get updated post"})
		return
	}

	writeJSON(ctx, fasthttp.StatusOK, updatedPost)
}

// DeletePost deletes a post
func (h *PostHandler) DeletePost(ctx *fasthttp.RequestCtx) {
	idStr := ctx.UserValue("id").(string)
	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeJSON(ctx, fasthttp.StatusBadRequest, map[string]string{"error": "Invalid post ID"})
		return
	}

	if err := h.postRepo.Delete(id); err != nil {
		writeJSON(ctx, fasthttp.StatusInternalServerError, map[string]string{"error": "Failed to delete post"})
		return
	}

	writeJSON(ctx, fasthttp.StatusOK, map[string]string{"message": "Post deleted successfully"})
}
