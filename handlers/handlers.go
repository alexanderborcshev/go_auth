package handlers

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"

	"tg-bot-test/middleware"
	"tg-bot-test/models"
	"tg-bot-test/pkg/security"
)

type Handlers struct {
	db        *gorm.DB
	jwtSecret string
}

func New(db *gorm.DB, jwtSecret string) *Handlers {
	return &Handlers{db: db, jwtSecret: jwtSecret}
}

// DefaultRole is applied when no role provided during registration.
const DefaultRole = "user"

// RegisterRequest represents the payload for user registration.
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3"`
	Password string `json:"password" binding:"required,min=6"`
	Role     string `json:"role"` // optional; defaults to "user"
}

// LoginRequest represents payload for user login.
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// UpdateProfileRequest allows partial updates of the current user's profile.
type UpdateProfileRequest struct {
	Username *string `json:"username"`
	Password *string `json:"password"`
}

// Register creates a new user with a hashed password.
func (h *Handlers) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeJSONError(c, http.StatusBadRequest, err.Error())
		return
	}
	if req.Role == "" {
		req.Role = DefaultRole
	}

	hashed, err := security.HashPassword(req.Password)
	if err != nil {
		writeJSONError(c, http.StatusInternalServerError, "failed to hash password")
		return
	}
	user := models.User{Username: req.Username, Password: hashed, Role: req.Role}
	if err := h.db.Create(&user).Error; err != nil {
		writeJSONError(c, http.StatusBadRequest, "could not create user: "+err.Error())
		return
	}
	writeJSON(c, http.StatusCreated, gin.H{"id": user.ID, "username": user.Username, "role": user.Role})
}

// Login verifies credentials and returns JWT token.
func (h *Handlers) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeJSONError(c, http.StatusBadRequest, err.Error())
		return
	}
	var user models.User
	if err := h.db.Where("username = ?", req.Username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeJSONError(c, http.StatusUnauthorized, "invalid credentials")
			return
		}
		writeJSONError(c, http.StatusInternalServerError, "db error")
		return
	}
	if err := security.CheckPassword(user.Password, req.Password); err != nil {
		writeJSONError(c, http.StatusUnauthorized, "invalid credentials")
		return
	}

	tokenString, err := h.generateJWT(&user)
	if err != nil {
		writeJSONError(c, http.StatusInternalServerError, "failed to sign token")
		return
	}
	writeJSON(c, http.StatusOK, gin.H{"token": tokenString})
}

// GetProfile returns current user's profile; requires authentication.
func (h *Handlers) GetProfile(c *gin.Context) {
	uid, ok := getUserIDFromCtx(c)
	if !ok {
		writeJSONError(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	var user models.User
	if err := h.db.First(&user, uid).Error; err != nil {
		writeJSONError(c, http.StatusNotFound, "user not found")
		return
	}
	writeJSON(c, http.StatusOK, gin.H{"id": user.ID, "username": user.Username, "role": user.Role})
}

// UpdateProfile allows updating the username and password of the current user.
func (h *Handlers) UpdateProfile(c *gin.Context) {
	uid, ok := getUserIDFromCtx(c)
	if !ok {
		writeJSONError(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeJSONError(c, http.StatusBadRequest, err.Error())
		return
	}
	updates := map[string]any{}
	if req.Username != nil {
		updates["username"] = *req.Username
	}
	if req.Password != nil {
		hash, err := security.HashPassword(*req.Password)
		if err != nil {
			writeJSONError(c, http.StatusInternalServerError, "failed to hash password")
			return
		}
		updates["password"] = hash
	}
	if len(updates) == 0 {
		writeJSONError(c, http.StatusBadRequest, "nothing to update")
		return
	}
	if err := h.db.Model(&models.User{}).Where("id = ?", uid).Updates(updates).Error; err != nil {
		writeJSONError(c, http.StatusBadRequest, "update failed: "+err.Error())
		return
	}
	c.Status(http.StatusNoContent)
}

// DeleteUser deletes a user by id; admin only (enforced by middleware).
func (h *Handlers) DeleteUser(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		writeJSONError(c, http.StatusBadRequest, "missing id")
		return
	}
	if err := h.db.Delete(&models.User{}, id).Error; err != nil {
		writeJSONError(c, http.StatusBadRequest, "delete failed: "+err.Error())
		return
	}
	c.Status(http.StatusNoContent)
}

func getUserIDFromCtx(c *gin.Context) (uint, bool) {
	v, ok := c.Get(middleware.CtxUserIDKey)
	if !ok {
		return 0, false
	}
	switch t := v.(type) {
	case float64:
		return uint(t), true
	case int:
		return uint(t), true
	case uint:
		return t, true
	case int64:
		return uint(t), true
	case string:
		// not expected but try to parse
		return 0, false
	default:
		return 0, false
	}
}

// writeJSON sends a JSON response with the given status code.
func writeJSON(c *gin.Context, status int, body any) {
	c.JSON(status, body)
}

// writeJSONError sends a standardized error JSON response.
func writeJSONError(c *gin.Context, status int, message string) {
	writeJSON(c, status, gin.H{"error": message})
}

// generateJWT builds a signed JWT for the given user.
func (h *Handlers) generateJWT(user *models.User) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"role":    user.Role,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	})
	return token.SignedString([]byte(h.jwtSecret))
}
