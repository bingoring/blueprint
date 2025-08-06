package main

import (
	"blueprint/internal/config"
	"blueprint/internal/database"
	"blueprint/internal/handlers"
	"blueprint/internal/middleware"
	"blueprint/internal/redis"
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

	// Redis 연결
	if err := redis.InitRedis(cfg); err != nil {
		log.Fatal("Failed to connect to Redis:", err)
	}
	defer redis.CloseRedis()

	// Gin 라우터 초기화
	router := gin.Default()

	// 미들웨어 설정
	router.Use(middleware.CORSMiddleware(cfg))
	router.Use(middleware.ResponseWrapper()) // 응답 래핑 미들웨어 추가

	// Initialize services
	// AI Service 초기화
	aiService := services.NewBridgeAIService(cfg, database.GetDB())

	// SSE Service 초기화
	sseService := services.NewSSEService()

	// Trading Service 초기화
	tradingService := services.NewTradingService(database.GetDB(), sseService)

	// 고성능 매칭 엔진 초기화 및 시작
	matchingEngine := services.NewMatchingEngine(database.GetDB())
	go func() {
		if err := matchingEngine.Start(); err != nil {
			log.Printf("Failed to start matching engine: %v", err)
		}
	}()

	// Market Maker 봇 초기화 및 시작
	marketMakerBot := services.NewMarketMakerBot(database.GetDB(), tradingService)

	// 🆕 워커 서비스 초기화 및 시작 (비동기 작업 처리)
	workerService := services.NewWorkerService()
	go func() {
		if err := workerService.Start(); err != nil {
			log.Printf("Failed to start worker service: %v", err)
		}
	}()

	// Market Maker 봇 백그라운드 시작
	go func() {
		if err := marketMakerBot.Start(); err != nil {
			log.Printf("Failed to start market maker bot: %v", err)
		}
	}()

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(cfg)
	projectHandler := handlers.NewProjectHandler(aiService)
	tradingHandler := handlers.NewTradingHandler(tradingService) // P2P 거래 핸들러

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

	// 🔐 인증이 필요한 라우터
	protected := api.Group("/")
	protected.Use(middleware.AuthMiddleware(cfg))
	{
		// 🔐 사용자 정보
		protected.GET("/users/me", authHandler.Me)                            // 기존 Me 메서드 사용
		protected.POST("/auth/logout", authHandler.Logout)                    // 로그아웃
		protected.POST("/auth/refresh", authHandler.RefreshToken)             // 토큰 갱신

		// 🏗️ 프로젝트 관리
		protected.POST("/projects", projectHandler.CreateProjectWithMilestones) // 기존 메서드 사용
		protected.GET("/projects", projectHandler.GetProjects)                  // 프로젝트 목록
		protected.GET("/projects/:id", projectHandler.GetProject)               // 특정 프로젝트
		protected.PUT("/projects/:id", projectHandler.UpdateProject)            // 프로젝트 수정
		protected.DELETE("/projects/:id", projectHandler.DeleteProject)         // 프로젝트 삭제
		protected.POST("/ai/milestones", projectHandler.GenerateAIMilestones)   // AI 마일스톤 제안

		// 💰 지갑 관리
		protected.GET("/wallet", tradingHandler.GetUserWallet)              // 사용자 지갑 조회

		// 📈 P2P 거래 시스템
		protected.POST("/orders", tradingHandler.CreateOrder)              // 주문 생성
		protected.GET("/orders/my", tradingHandler.GetMyOrders)            // 내 주문 내역
		protected.DELETE("/orders/:id", tradingHandler.CancelOrder)        // 주문 취소
		protected.GET("/trades/my", tradingHandler.GetMyTrades)            // 내 거래 내역
		protected.GET("/positions/my", tradingHandler.GetMyPositions)      // 내 포지션
		protected.GET("/milestones/:id/position/:option", tradingHandler.GetMilestonePosition) // 특정 포지션
	}

		// 📊 공개 마켓 데이터 API
		api.GET("/milestones/:id/market", tradingHandler.GetMilestoneMarket)           // 마켓 정보 조회
		api.POST("/milestones/:id/market/init", tradingHandler.InitializeMarket)       // 마켓 초기화
		api.GET("/milestones/:id/orderbook/:option", tradingHandler.GetOrderBook)      // 호가창 조회 (option별)
		api.GET("/milestones/:id/trades/:option", tradingHandler.GetRecentTrades)      // 최근 거래 조회 (option별)

		// 📡 실시간 연결
		api.GET("/milestones/:id/stream", tradingHandler.HandleSSEConnection)          // SSE 연결

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
