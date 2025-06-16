package handlers

import (
	"os"
	"strconv"
	"time"

	"gobackend/models"

	"github.com/golang-jwt/jwt/v5"
	"github.com/valyala/fasthttp"
)

type UserHandler struct {
	userRepo *models.UserRepository
}

type createUserRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
	IsAdmin  bool   `json:"is_admin"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func NewUserHandler(userRepo *models.UserRepository) *UserHandler {
	return &UserHandler{userRepo: userRepo}
}

// Login authenticates a user and returns a JWT
func (h *UserHandler) Login(ctx *fasthttp.RequestCtx) {
	var req loginRequest
	if err := readJSON(ctx, &req); err != nil {
		writeJSON(ctx, fasthttp.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	user, err := h.userRepo.GetByEmail(req.Email)
	if err != nil || user == nil || !user.CheckPassword(req.Password) {
		writeJSON(ctx, fasthttp.StatusUnauthorized, map[string]string{"error": "invalid credentials"})
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"exp":     time.Now().Add(72 * time.Hour).Unix(),
	})
	secret := os.Getenv("JWT_SECRET")
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		writeJSON(ctx, fasthttp.StatusInternalServerError, map[string]string{"error": "failed to sign token"})
		return
	}

	writeJSON(ctx, fasthttp.StatusOK, map[string]string{"token": signed})
}

// CreateUser creates a new user
func (h *UserHandler) CreateUser(ctx *fasthttp.RequestCtx) {
	var req createUserRequest
	if err := readJSON(ctx, &req); err != nil {
		writeJSON(ctx, fasthttp.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	user := models.User{Name: req.Name, Email: req.Email, IsAdmin: req.IsAdmin}
	if err := user.SetPassword(req.Password); err != nil {
		writeJSON(ctx, fasthttp.StatusInternalServerError, map[string]string{"error": "failed to set password"})
		return
	}

	if err := h.userRepo.Create(&user); err != nil {
		writeJSON(ctx, fasthttp.StatusInternalServerError, map[string]string{"error": "Failed to create user"})
		return
	}

	writeJSON(ctx, fasthttp.StatusCreated, user)
}

// GetUser retrieves a user by ID
func (h *UserHandler) GetUser(ctx *fasthttp.RequestCtx) {
	idStr := ctx.UserValue("id").(string)
	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeJSON(ctx, fasthttp.StatusBadRequest, map[string]string{"error": "Invalid user ID"})
		return
	}

	user, err := h.userRepo.GetByID(id)
	if err != nil {
		writeJSON(ctx, fasthttp.StatusInternalServerError, map[string]string{"error": "Failed to get user"})
		return
	}

	if user == nil {
		writeJSON(ctx, fasthttp.StatusNotFound, map[string]string{"error": "User not found"})
		return
	}

	writeJSON(ctx, fasthttp.StatusOK, user)
}

// GetUsers retrieves all users
func (h *UserHandler) GetUsers(ctx *fasthttp.RequestCtx) {
	users, err := h.userRepo.GetAll()
	if err != nil {
		writeJSON(ctx, fasthttp.StatusInternalServerError, map[string]string{"error": "Failed to get users"})
		return
	}

	writeJSON(ctx, fasthttp.StatusOK, users)
}

// UpdateUser updates a user
func (h *UserHandler) UpdateUser(ctx *fasthttp.RequestCtx) {
	idStr := ctx.UserValue("id").(string)
	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeJSON(ctx, fasthttp.StatusBadRequest, map[string]string{"error": "Invalid user ID"})
		return
	}

	var req createUserRequest
	if err := readJSON(ctx, &req); err != nil {
		writeJSON(ctx, fasthttp.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	user := models.User{ID: id, Name: req.Name, Email: req.Email, IsAdmin: req.IsAdmin}
	if req.Password != "" {
		if err := user.SetPassword(req.Password); err != nil {
			writeJSON(ctx, fasthttp.StatusInternalServerError, map[string]string{"error": "failed to set password"})
			return
		}
	}

	if err := h.userRepo.Update(&user); err != nil {
		writeJSON(ctx, fasthttp.StatusInternalServerError, map[string]string{"error": "Failed to update user"})
		return
	}

	updatedUser, err := h.userRepo.GetByID(id)
	if err != nil {
		writeJSON(ctx, fasthttp.StatusInternalServerError, map[string]string{"error": "Failed to get updated user"})
		return
	}

	writeJSON(ctx, fasthttp.StatusOK, updatedUser)
}

// DeleteUser deletes a user
func (h *UserHandler) DeleteUser(ctx *fasthttp.RequestCtx) {
	idStr := ctx.UserValue("id").(string)
	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeJSON(ctx, fasthttp.StatusBadRequest, map[string]string{"error": "Invalid user ID"})
		return
	}

	if err := h.userRepo.Delete(id); err != nil {
		writeJSON(ctx, fasthttp.StatusInternalServerError, map[string]string{"error": "Failed to delete user"})
		return
	}

	writeJSON(ctx, fasthttp.StatusOK, map[string]string{"message": "User deleted successfully"})
}
