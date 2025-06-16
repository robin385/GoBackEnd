package handlers

import (
	"strconv"

	"gobackend/models"

	"github.com/valyala/fasthttp"
)

// CommentHandler handles comment endpoints
type CommentHandler struct {
	repo     *models.CommentRepository
	userRepo *models.UserRepository
	postRepo *models.TrashPostRepository
}

func NewCommentHandler(repo *models.CommentRepository, userRepo *models.UserRepository, postRepo *models.TrashPostRepository) *CommentHandler {
	return &CommentHandler{repo: repo, userRepo: userRepo, postRepo: postRepo}
}

// createCommentRequest represents the payload for creating a comment
type createCommentRequest struct {
	Content string `json:"content"`
}

// CreateComment adds a new comment to a post
func (h *CommentHandler) CreateComment(ctx *fasthttp.RequestCtx) {
	postIDStr := ctx.UserValue("id").(string)
	postID, err := strconv.Atoi(postIDStr)
	if err != nil {
		writeJSON(ctx, fasthttp.StatusBadRequest, map[string]string{"error": "invalid post id"})
		return
	}

	userID, err := getUserIDFromToken(ctx)
	if err != nil {
		writeJSON(ctx, fasthttp.StatusUnauthorized, map[string]string{"error": err.Error()})
		return
	}

	var req createCommentRequest
	if err := readJSON(ctx, &req); err != nil {
		writeJSON(ctx, fasthttp.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	user, err := h.userRepo.GetByID(userID)
	if err != nil || user == nil {
		writeJSON(ctx, fasthttp.StatusBadRequest, map[string]string{"error": "invalid user"})
		return
	}

	post, err := h.postRepo.GetByID(postID)
	if err != nil || post == nil {
		writeJSON(ctx, fasthttp.StatusBadRequest, map[string]string{"error": "invalid post"})
		return
	}

	c := models.Comment{PostID: postID, UserID: userID, Content: req.Content, User: user}
	if err := h.repo.Create(&c); err != nil {
		writeJSON(ctx, fasthttp.StatusInternalServerError, map[string]string{"error": "failed to create comment"})
		return
	}

	_ = h.userRepo.AddExp(userID, 10)
	_ = h.userRepo.AddExp(post.UserID, 10)

	writeJSON(ctx, fasthttp.StatusCreated, c)
}

// GetComments returns comments for a post
func (h *CommentHandler) GetComments(ctx *fasthttp.RequestCtx) {
	postIDStr := ctx.UserValue("id").(string)
	postID, err := strconv.Atoi(postIDStr)
	if err != nil {
		writeJSON(ctx, fasthttp.StatusBadRequest, map[string]string{"error": "invalid post id"})
		return
	}

	comments, err := h.repo.GetByPostID(postID)
	if err != nil {
		writeJSON(ctx, fasthttp.StatusInternalServerError, map[string]string{"error": "failed to get comments"})
		return
	}

	writeJSON(ctx, fasthttp.StatusOK, comments)
}
