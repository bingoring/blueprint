package main

import (
	"blueprint/internal/config"
	"blueprint/internal/database"
	"blueprint/internal/handlers"
	"blueprint/internal/middleware"
	"blueprint/internal/services"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	// 설정 로드
	cfg := config.LoadConfig()

	// Gin 모드 설정
	gin.SetMode(cfg.Server.Mode)

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
	router.Use(middleware.ResponseWrapper()) // 응답 래핑 미들웨어 추가

	// 핸들러 초기화
	authHandler := handlers.NewAuthHandler(cfg)

	// Initialize services
	aiService := services.NewBridgeAIService(cfg, database.GetDB())
	goalHandler := handlers.NewGoalHandler(aiService)

	// API 라우트 그룹
	api := router.Group("/api/v1")

	// 인증 관련 API
	auth := api.Group("/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
		auth.GET("/google/login", authHandler.GoogleLogin)
		auth.GET("/google/callback", authHandler.GoogleCallback)
	}

	// 보호된 라우트 (인증 필요)
	protected := api.Group("/")
	protected.Use(middleware.AuthMiddleware(cfg))
	{
		// 사용자 관련
		protected.GET("/me", authHandler.Me)
		protected.POST("/auth/logout", authHandler.Logout)     // 로그아웃
		protected.POST("/auth/refresh", authHandler.RefreshToken) // 토큰 갱신
		protected.GET("/auth/token-expiry", authHandler.CheckTokenExpiry) // 토큰 만료 확인

		// 목표 관리
		goals := protected.Group("/goals")
		{
			goals.POST("", goalHandler.CreateGoal)                    // 목표 생성
			goals.GET("", goalHandler.GetGoals)                      // 목표 목록 조회 (필터링, 페이지네이션)
			goals.GET("/:id", goalHandler.GetGoal)                   // 특정 목표 조회
			goals.PUT("/:id", goalHandler.UpdateGoal)                // 목표 수정
			goals.DELETE("/:id", goalHandler.DeleteGoal)             // 목표 삭제
			goals.PATCH("/:id/status", goalHandler.UpdateGoalStatus) // 목표 상태 변경
		}

		// 꿈 등록 (마일스톤 포함) ✨
		protected.POST("/dreams", goalHandler.CreateGoalWithMilestones)

		// AI 마일스톤 제안 🤖
		protected.POST("/ai/milestones", goalHandler.GenerateAIMilestones)

		// AI 사용 정보 조회 📊
		protected.GET("/ai/usage", goalHandler.GetAIUsageInfo)

		// 목표 메타데이터
		protected.GET("/goal-categories", goalHandler.GetGoalCategories) // 카테고리 목록
		protected.GET("/goal-statuses", goalHandler.GetGoalStatuses)     // 상태 목록
	}

	// 헬스 체크
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"message": "Blueprint API Server is running",
		})
	})

	// 서버 시작
	log.Printf("Server starting on port %s", cfg.Server.Port)
	if err := router.Run(":" + cfg.Server.Port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
