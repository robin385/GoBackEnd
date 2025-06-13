package handlers

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"os"
	"time"

	"gobackend/models"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
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

func (h *OAuthHandler) Login(c *gin.Context) {
	url := h.config.AuthCodeURL("state")
	c.Redirect(http.StatusFound, url)
}

func (h *OAuthHandler) Callback(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "code required"})
		return
	}

	token, err := h.config.Exchange(c, code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "token exchange failed"})
		return
	}

	client := h.config.Client(c, token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user info"})
		return
	}
	defer resp.Body.Close()

	var info struct {
		Email string `json:"email"`
		Name  string `json:"name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "decode user info"})
		return
	}

	user, err := h.userRepo.GetByEmail(info.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
		return
	}
	if user == nil {
		user = &models.User{Name: info.Name, Email: info.Email}
		_ = user.SetPassword(randomString())
		if err := h.userRepo.Create(user); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "create user"})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "token sign"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": signed})
}

func randomString() string {
	b := make([]byte, 16)
	for i := range b {
		b[i] = byte('a' + rand.Intn(26))
	}
	return string(b)
}
