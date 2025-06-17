package handlers

import (
	"context"
	"encoding/json"
	"math/rand"
	"net/http"
	"os"
	"time"

	"gobackend/models"

	"github.com/golang-jwt/jwt/v5"
	"github.com/valyala/fasthttp"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type OAuthHandler struct {
	config   *oauth2.Config
	userRepo *models.UserRepository
}

func NewOAuthHandler(userRepo *models.UserRepository) *OAuthHandler {
	conf := &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		Endpoint:     google.Endpoint,
		RedirectURL:  os.Getenv("GOOGLE_REDIRECT_URL"),
		Scopes:       []string{"email", "profile"},
	}
	return &OAuthHandler{config: conf, userRepo: userRepo}
}

func (h *OAuthHandler) Login(ctx *fasthttp.RequestCtx) {
	url := h.config.AuthCodeURL("state")
	ctx.Redirect(url, http.StatusFound)
}

func (h *OAuthHandler) Callback(ctx *fasthttp.RequestCtx) {
	code := string(ctx.QueryArgs().Peek("code"))
	if code == "" {
		writeJSON(ctx, fasthttp.StatusBadRequest, map[string]string{"error": "code required"})
		return
	}

	token, err := h.config.Exchange(context.Background(), code)
	if err != nil {
		writeJSON(ctx, fasthttp.StatusInternalServerError, map[string]string{"error": "token exchange failed"})
		return
	}

	client := h.config.Client(context.Background(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		writeJSON(ctx, fasthttp.StatusInternalServerError, map[string]string{"error": "failed to get user info"})
		return
	}
	defer resp.Body.Close()

	var info struct {
		Email string `json:"email"`
		Name  string `json:"name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		writeJSON(ctx, fasthttp.StatusInternalServerError, map[string]string{"error": "decode user info"})
		return
	}

	user, err := h.userRepo.GetByEmail(info.Email)
	if err != nil {
		writeJSON(ctx, fasthttp.StatusInternalServerError, map[string]string{"error": "db error"})
		return
	}
	if user == nil {
		user = &models.User{Name: info.Name, Email: info.Email}
		_ = user.SetPassword(randomString())
		if err := h.userRepo.Create(user); err != nil {
			writeJSON(ctx, fasthttp.StatusInternalServerError, map[string]string{"error": "create user"})
			return
		}
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"exp":     time.Now().Add(72 * time.Hour).Unix(),
	})
	secret := os.Getenv("JWT_SECRET")
	signed, err := jwtToken.SignedString([]byte(secret))
	if err != nil {
		writeJSON(ctx, fasthttp.StatusInternalServerError, map[string]string{"error": "token sign"})
		return
	}

	writeJSON(ctx, fasthttp.StatusOK, map[string]interface{}{"token": signed, "user": user})
}

func randomString() string {
	b := make([]byte, 16)
	for i := range b {
		b[i] = byte('a' + rand.Intn(26))
	}
	return string(b)
}
