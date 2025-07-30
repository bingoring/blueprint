package main

import (
	"blueprint/internal/config"
	"blueprint/internal/database"
	"blueprint/internal/handlers"
	"blueprint/internal/middleware"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	// 설정 로드
	cfg := config.LoadConfig()

	// Gin 모드 설정
	gin.SetMode(cfg.Server.GinMode)

	// 데이터베이스 연결
	if err := database.Connect(cfg); err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// 데이터베이스 마이그레이션
	if err := database.AutoMigrate(); err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	// Gin 라우터 초기화
	router := gin.Default()

	// 미들웨어 설정
	router.Use(middleware.CORSMiddleware(cfg))

	// 핸들러 초기화
	authHandler := handlers.NewAuthHandler(cfg)

	// API 라우트 그룹
	api := router.Group("/api/v1")

	// 인증 관련 라우트 (인증 불필요)
	auth := api.Group("/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
		auth.GET("/google", authHandler.GoogleLogin)
		auth.GET("/google/callback", authHandler.GoogleCallback)
	}

	// 보호된 라우트 (인증 필요)
	protected := api.Group("/")
	protected.Use(middleware.AuthMiddleware(cfg))
	{
		protected.GET("/me", authHandler.Me)
		// TODO: 추후 사용자 프로필, 목표 관련 라우트 추가
	}

	// 헬스 체크
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"message": "LifePathDAO API Server is running",
		})
	})

	// 서버 시작
	log.Printf("Server starting on port %s", cfg.Server.Port)
	if err := router.Run(":" + cfg.Server.Port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
