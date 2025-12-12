package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"go-auth/handlers"
	"go-auth/internal/database"
	"go-auth/middleware"
	"go-auth/models"
	cfg "go-auth/pkg/config"
)

func setupRouter(db *gorm.DB, jwtSecret string) *gin.Engine {
	r := gin.Default()

	h := handlers.New(db, jwtSecret)
	authMW := middleware.AuthMiddleware(jwtSecret)
	adminMW := middleware.RoleMiddleware("admin")

	// Public endpoints
	r.POST("/register", h.Register)
	r.POST("/login", h.Login)

	// Protected endpoints
	auth := r.Group("/")
	auth.Use(authMW)
	{
		auth.GET("/profile", h.GetProfile)
		auth.PUT("/profile", h.UpdateProfile)
	}

	// Admin endpoints
	admin := r.Group("/")
	admin.Use(authMW, adminMW)
	{
		admin.DELETE("/user/:id", h.DeleteUser)
	}

	return r
}

func main() {
	config := cfg.Load()

	db, err := database.OpenSQLite(config.DBPath)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	if err := db.AutoMigrate(&models.User{}); err != nil {
		log.Fatalf("failed to migrate database: %v", err)
	}

	r := setupRouter(db, config.JWTSecret)
	addr := ":" + config.Port
	if port := os.Getenv("PORT"); port != "" {
		addr = ":" + port
	}
	if err := r.Run(addr); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
