package main

import (
	"blueprint/internal/config"
	"blueprint/internal/database"
	"blueprint/internal/handlers"
	"blueprint/internal/middleware"
	"blueprint/internal/services"
	"log"
	"net/http"

	moduleConfig "blueprint-module/pkg/config"
	moduleRedis "blueprint-module/pkg/redis"

	"github.com/gin-gonic/gin"
)

// config 타입 변환 함수
func convertToModuleConfig(cfg *config.Config) *moduleConfig.Config {
	return &moduleConfig.Config{
		Database: moduleConfig.DatabaseConfig{
			Host:     cfg.Database.Host,
			Port:     cfg.Database.Port,
			User:     cfg.Database.User,
			Password: cfg.Database.Password,
			Name:     cfg.Database.Name,
			SSLMode:  cfg.Database.SSLMode,
		},
		JWT: moduleConfig.JWTConfig{
			Secret: cfg.JWT.Secret,
		},
		OAuth: moduleConfig.OAuthConfig{
			Google: moduleConfig.GoogleOAuthConfig{
				ClientID:     cfg.Google.ClientID,
				ClientSecret: cfg.Google.ClientSecret,
				RedirectURL:  cfg.Google.RedirectURL,
				Scopes:       "profile email",
			},
			LinkedIn: moduleConfig.LinkedInOAuthConfig{
				ClientID:     cfg.LinkedIn.ClientID,
				ClientSecret: cfg.LinkedIn.ClientSecret,
				RedirectURL:  cfg.LinkedIn.RedirectURL,
				Scopes:       "r_liteprofile r_emailaddress",
			},
			Twitter: moduleConfig.TwitterOAuthConfig{
				ClientID:     cfg.Twitter.ClientID,
				ClientSecret: cfg.Twitter.ClientSecret,
				RedirectURL:  cfg.Twitter.RedirectURL,
				Scopes:       "tweet.read users.read",
			},
			GitHub: moduleConfig.GitHubOAuthConfig{
				ClientID:     cfg.GitHub.ClientID,
				ClientSecret: cfg.GitHub.ClientSecret,
				RedirectURL:  cfg.GitHub.RedirectURL,
				Scopes:       "user:email",
			},
		},
		Server: moduleConfig.ServerConfig{
			Port:        cfg.Server.Port,
			Mode:        cfg.Server.Mode,
			FrontendURL: cfg.Server.FrontendURL,
		},
		AI: moduleConfig.AIConfig{
			Provider: cfg.AI.Provider,
			OpenAI: moduleConfig.OpenAIConfig{
				APIKey: cfg.AI.OpenAI.APIKey,
				Model:  cfg.AI.OpenAI.Model,
			},
		},
		Redis: moduleConfig.RedisConfig{
			Host:     cfg.Redis.Host,
			Port:     cfg.Redis.Port,
			Password: cfg.Redis.Password,
			DB:       cfg.Redis.DB,
		},
	}
}

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

	// Redis 연결 (blueprint-module 사용)
	moduleCfg := &moduleConfig.Config{
		Redis: moduleConfig.RedisConfig{
			Host:     cfg.Redis.Host,
			Port:     cfg.Redis.Port,
			Password: cfg.Redis.Password,
			DB:       cfg.Redis.DB,
		},
	}

	if err := moduleRedis.InitRedis(moduleCfg); err != nil {
		log.Fatal("Failed to connect to Redis:", err)
	}
	defer moduleRedis.CloseRedis()

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

	// 고성능 매칭 엔진 초기화 및 시작 (SSE 서비스 주입)
	matchingEngine := services.NewMatchingEngine(database.GetDB(), sseService)
	go func() {
		if err := matchingEngine.Start(); err != nil {
			log.Printf("❌ CRITICAL: Failed to start matching engine: %v", err)
			log.Printf("🚨 Trading functionality will not work!")
		} else {
			log.Printf("✅ Matching engine started successfully")
		}
	}()

	// Trading Service 초기화 (매칭 엔진 주입)
	tradingService := services.NewTradingService(database.GetDB(), sseService, matchingEngine)

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
	// 핸들러 초기화
	moduleConfig := convertToModuleConfig(cfg)
	authHandler := handlers.NewAuthHandler(moduleConfig)
	magicLinkHandler := handlers.NewMagicLinkHandler(moduleConfig)
	projectHandler := handlers.NewProjectHandler(moduleConfig, aiService)
	tradingHandler := handlers.NewTradingHandler(tradingService)
	userSettingsHandler := handlers.NewUserSettingsHandler(moduleConfig)
	oauthHandler := handlers.NewOAuthHandler(moduleConfig)
	activityHandler := handlers.NewActivityHandler() // 활동 로그 핸들러 추가
	profileHandler := handlers.NewProfileHandler()   // 프로필 핸들러 추가

	// API 라우트 그룹
	api := router.Group("/api/v1")

	// 🔐 인증 관련 (비보호)
	auth := api.Group("/auth")
	{
		// Google OAuth (기존 로그인용)
		auth.GET("/google/login", authHandler.GoogleLogin)
		auth.GET("/google/callback", authHandler.GoogleCallback)

		// Magic Link 인증
		auth.POST("/magic-link", magicLinkHandler.CreateMagicLink)
		auth.POST("/verify-magic-link", magicLinkHandler.VerifyMagicLink)

		// 소셜 미디어 연결 (신원 증명용)
		auth.GET("/:provider/connect", middleware.AuthMiddleware(cfg), oauthHandler.StartOAuthConnect)
		auth.GET("/:provider/callback", oauthHandler.OAuthCallback)

		// OAuth 제공업체 목록 조회
		auth.GET("/providers", oauthHandler.GetSupportedProviders)
	}

	// 🔐 인증이 필요한 라우터
	protected := api.Group("/")
	protected.Use(middleware.AuthMiddleware(cfg))
	{
		// 🔐 사용자 정보
		protected.GET("/users/me", authHandler.Me)                            // 사용자 정보 조회
		protected.POST("/auth/logout", authHandler.Logout)                    // 로그아웃
		protected.POST("/auth/refresh", authHandler.RefreshToken)             // 토큰 갱신
		protected.GET("/auth/token-expiry", authHandler.CheckTokenExpiry)     // 토큰 만료 확인

		// 🧑‍💼 계정 설정 & 신원 증명
		protected.GET("/users/me/settings", userSettingsHandler.GetMySettings)
		protected.PUT("/users/me/profile", userSettingsHandler.UpdateProfile)
		protected.PUT("/users/me/preferences", userSettingsHandler.UpdatePreferences)
		// 신원 증명 액션
		protected.POST("/users/me/verify/email", userSettingsHandler.RequestVerifyEmail)
		protected.POST("/users/me/verify/email/confirm", userSettingsHandler.VerifyEmailCode)
		protected.POST("/users/me/verify/phone", userSettingsHandler.RequestVerifyPhone)
		protected.POST("/users/me/connect/:provider", userSettingsHandler.ConnectProvider) // linkedin|github|twitter
		protected.POST("/users/me/verify/work-email", userSettingsHandler.VerifyWorkEmail)
		protected.POST("/users/me/verify/professional", userSettingsHandler.SubmitProfessionalDoc)
		protected.POST("/users/me/verify/education", userSettingsHandler.SubmitEducationDoc)

		// 📝 활동 로그
		protected.GET("/users/me/activities", activityHandler.GetUserActivities)       // 사용자 활동 로그 조회
		protected.GET("/users/me/activities/summary", activityHandler.GetActivitySummary) // 활동 요약 (대시보드용)

		// 👤 프로필 조회 (public/private)
		protected.GET("/users/:username/profile", profileHandler.GetUserProfile) // 사용자 프로필 조회

		// 🏗️ 프로젝트 관리
		protected.POST("/projects", projectHandler.CreateProjectWithMilestones) // 기존 메서드 사용
		protected.GET("/projects", projectHandler.GetProjects)                  // 프로젝트 목록
		protected.GET("/projects/:id", projectHandler.GetProject)               // 특정 프로젝트
		protected.PUT("/projects/:id", projectHandler.UpdateProject)            // 프로젝트 수정
		protected.DELETE("/projects/:id", projectHandler.DeleteProject)         // 프로젝트 삭제
		protected.GET("/ai/usage", projectHandler.GetAIUsageInfo)   // AI 마일스톤 제안
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
		api.GET("/milestones/:id/price-history/:option", tradingHandler.GetPriceHistory) // 가격 히스토리 조회 (option별)

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
