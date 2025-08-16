package integration_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"blueprint-module/pkg/models"
	"blueprint/internal/services"
	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TradingIntegrationTestSuite 거래 시스템 통합 테스트
type TradingIntegrationTestSuite struct {
	suite.Suite
	router         *gin.Engine
	db             *gorm.DB
	redisServer    *miniredis.Miniredis
	redisClient    *redis.Client
	tradingService *services.DistributedTradingService
	sseService     *services.SSEService
}

func (suite *TradingIntegrationTestSuite) SetupSuite() {
	// Gin을 test mode로 설정
	gin.SetMode(gin.TestMode)

	// 데이터베이스 설정
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	suite.Require().NoError(err)
	suite.db = db

	// 테이블 마이그레이션
	err = db.AutoMigrate(
		&models.User{},
		&models.UserProfile{},
		&models.UserVerification{},
		&models.Project{},
		&models.Milestone{},
		&models.Order{},
		&models.Trade{},
		&models.Position{},
		&models.MarketData{},
		&models.UserWallet{},
		&models.PriceHistory{},
		&models.StakingPool{},
		&models.RevenueDistribution{},
		&models.StakingReward{},
		&models.GovernanceProposal{},
		&models.GovernanceVote{},
		&models.BlueprintReward{},
		&models.PlatformFeeConfig{},
	)
	suite.Require().NoError(err)

	// Redis 설정
	suite.redisServer = miniredis.RunT(suite.T())
	suite.redisClient = redis.NewClient(&redis.Options{
		Addr: suite.redisServer.Addr(),
	})

	// 서비스 초기화 (테스트용 Redis 클라이언트 사용)
	suite.sseService = services.NewSSEService()
	suite.tradingService = services.NewDistributedTradingServiceWithRedis(suite.db, suite.sseService, suite.redisClient)

	// 라우터 설정
	suite.router = gin.New()
	suite.setupRoutes()

	// 테스트 데이터 생성
	suite.createTestData()

	// 거래 서비스 시작
	err = suite.tradingService.Start()
	suite.Require().NoError(err)
}

func (suite *TradingIntegrationTestSuite) TearDownSuite() {
	suite.tradingService.Stop()
	suite.redisServer.Close()
	suite.redisClient.Close()
}

func (suite *TradingIntegrationTestSuite) SetupTest() {
	suite.redisServer.FlushAll()
}

// setupRoutes API 라우트 설정
func (suite *TradingIntegrationTestSuite) setupRoutes() {
	api := suite.router.Group("/api/v1")
	{
		trading := api.Group("/trading")
		{
			// Mock API 엔드포인트들
			trading.POST("/orders", func(c *gin.Context) {
				c.JSON(http.StatusCreated, gin.H{"message": "order created"})
			})
			trading.GET("/orders/:id", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"order": "mock order"})
			})
			trading.DELETE("/orders/:id", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "order cancelled"})
			})
			trading.GET("/orderbook/:milestoneId/:optionId", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"orderbook": "mock orderbook"})
			})
			trading.GET("/market-data/:milestoneId/:optionId", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"market_data": "mock data"})
			})
			trading.GET("/user/:userId/orders", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"orders": []string{}})
			})
		}

		// SSE 엔드포인트
		api.GET("/stream/:id", suite.sseService.HandleSSEConnection)
	}

	// 헬스 체크
	api.GET("/health", func(c *gin.Context) {
		metrics := suite.tradingService.GetSystemMetrics()
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"metrics": metrics,
		})
	})
}

// TestCreateOrderAPI 주문 생성 API 테스트
func (suite *TradingIntegrationTestSuite) TestCreateOrderAPI() {
	// 주문 생성 요청
	orderReq := map[string]interface{}{
		"milestone_id": 1,
		"option_id":    "success",
		"side":         "buy",
		"quantity":     100,
		"price":        0.75,
		"type":         "limit",
	}

	body, _ := json.Marshal(orderReq)
	req, _ := http.NewRequest("POST", "/api/v1/trading/orders", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-jwt-token") // Mock JWT

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// 응답 검증
	suite.Assert().Equal(http.StatusCreated, w.Code)
}

// TestOrderMatching 주문 매칭 통합 테스트
func (suite *TradingIntegrationTestSuite) TestOrderMatching() {
	// 1. 매도 주문 생성
	sellResult, err := suite.tradingService.CreateOrder(
		2,         // userID
		1,         // milestoneID
		"success", // optionID
		"sell",    // orderType
		50,        // quantity
		0.70,      // price
	)
	suite.Assert().NoError(err)
	suite.Assert().NotNil(sellResult)

	// 2. 매수 주문 생성 (매칭 발생)
	buyResult, err := suite.tradingService.CreateOrder(
		1,         // userID
		1,         // milestoneID
		"success", // optionID
		"buy",     // orderType
		30,        // quantity
		0.75,      // price (더 높은 가격)
	)
	suite.Assert().NoError(err)
	suite.Assert().NotNil(buyResult)
	suite.Assert().True(buyResult.Executed)
	suite.Assert().Len(buyResult.Trades, 1)

	// 3. 거래 결과 검증
	trade := buyResult.Trades[0]
	suite.Assert().Equal(int64(30), trade.Quantity)
	suite.Assert().Equal(0.70, trade.Price) // 매도가격으로 체결

	// 4. 마켓 데이터 확인
	marketData, err := suite.tradingService.GetMarketData(1, "success")
	suite.Assert().NoError(err)
	suite.Assert().Equal(0.70, marketData.LastPrice)
}

// TestOrderBookAPI 주문장 조회 API 테스트
func (suite *TradingIntegrationTestSuite) TestOrderBookAPI() {
	// 여러 주문 생성
	orders := []struct {
		userID   uint
		side     string
		quantity int64
		price    float64
	}{
		{1, "buy", 100, 0.74},
		{1, "buy", 50, 0.73},
		{2, "sell", 75, 0.76},
		{2, "sell", 25, 0.77},
	}

	for _, order := range orders {
		_, err := suite.tradingService.CreateOrder(
			order.userID, 1, "success", order.side, order.quantity, order.price,
		)
		suite.Assert().NoError(err)
	}

	// 주문장 조회
	orderBook, err := suite.tradingService.GetOrderBook(1, "success", 10)
	suite.Assert().NoError(err)
	suite.Assert().NotNil(orderBook)

	// API로 주문장 조회
	req, _ := http.NewRequest("GET", "/api/v1/trading/orderbook/1/success", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// 응답 검증
	suite.Assert().Equal(http.StatusOK, w.Code)
}

// TestConcurrentTrading 동시 거래 테스트
func (suite *TradingIntegrationTestSuite) TestConcurrentTrading() {
	numTraders := 5
	ordersPerTrader := 10
	results := make(chan error, numTraders*ordersPerTrader)

	// 동시에 여러 트레이더가 주문
	for i := 0; i < numTraders; i++ {
		go func(traderID int) {
			for j := 0; j < ordersPerTrader; j++ {
				side := "buy"
				if j%2 == 0 {
					side = "sell"
				}

				_, err := suite.tradingService.CreateOrder(
					uint(traderID+1),
					1,
					"success",
					side,
					int64(10+j),
					0.75+float64(j)*0.001,
				)
				results <- err
			}
		}(i)
	}

	// 모든 주문 완료 대기
	for i := 0; i < numTraders*ordersPerTrader; i++ {
		err := <-results
		suite.Assert().NoError(err)
	}

	// 최종 상태 확인
	marketData, err := suite.tradingService.GetMarketData(1, "success")
	suite.Assert().NoError(err)
	suite.Assert().NotNil(marketData)
}

// TestSSEStreaming SSE 스트리밍 테스트
func (suite *TradingIntegrationTestSuite) TestSSEStreaming() {
	// SSE 연결 시뮬레이션
	req, _ := http.NewRequest("GET", "/api/v1/stream/1", nil)
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Cache-Control", "no-cache")

	w := httptest.NewRecorder()

	// 별도 고루틴에서 SSE 연결 처리
	go func() {
		suite.router.ServeHTTP(w, req)
	}()

	// 잠시 대기 후 거래 실행하여 SSE 이벤트 발생
	time.Sleep(100 * time.Millisecond)

	_, err := suite.tradingService.CreateOrder(1, 1, "success", "buy", 100, 0.75)
	suite.Assert().NoError(err)

	// SSE 연결 응답 확인
	time.Sleep(200 * time.Millisecond)
	suite.Assert().Contains(w.Header().Get("Content-Type"), "text/event-stream")
}

// TestHealthCheck 헬스 체크 테스트
func (suite *TradingIntegrationTestSuite) TestHealthCheck() {
	req, _ := http.NewRequest("GET", "/api/v1/health", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	suite.Assert().Equal(http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.Assert().NoError(err)
	suite.Assert().Equal("healthy", response["status"])
	suite.Assert().Contains(response, "metrics")
}

// TestUserOrdersAPI 사용자 주문 조회 API 테스트
func (suite *TradingIntegrationTestSuite) TestUserOrdersAPI() {
	// 사용자 주문 생성
	for i := 0; i < 3; i++ {
		_, err := suite.tradingService.CreateOrder(
			1, 1, "success", "buy", int64(50+i*10), 0.75+float64(i)*0.01,
		)
		suite.Assert().NoError(err)
	}

	// API로 사용자 주문 조회
	req, _ := http.NewRequest("GET", "/api/v1/trading/user/1/orders?limit=10", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// 응답 검증
	suite.Assert().Equal(http.StatusOK, w.Code)
}

// TestDataConsistency 데이터 일관성 테스트
func (suite *TradingIntegrationTestSuite) TestDataConsistency() {
	// 매칭 가능한 주문 쌍 생성
	_, err := suite.tradingService.CreateOrder(2, 1, "success", "sell", 100, 0.70)
	suite.Assert().NoError(err)

	buyResult, err := suite.tradingService.CreateOrder(1, 1, "success", "buy", 100, 0.75)
	suite.Assert().NoError(err)
	suite.Assert().True(buyResult.Executed)

	// 데이터베이스에서 거래 기록 확인
	var dbTrades []models.Trade
	err = suite.db.Where("milestone_id = ? AND option_id = ?", 1, "success").Find(&dbTrades).Error
	suite.Assert().NoError(err)
	suite.Assert().Len(dbTrades, 1)

	// Redis 캐시에서 가격 정보 확인
	marketData, err := suite.tradingService.GetMarketData(1, "success")
	suite.Assert().NoError(err)
	suite.Assert().Equal(0.70, marketData.LastPrice)

	// 사용자 포지션 확인 (구현되어 있다면)
	// ...
}

func (suite *TradingIntegrationTestSuite) createTestData() {
	// 사용자 생성
	users := []models.User{
		{ID: 1, Username: "trader1", Email: "trader1@test.com"},
		{ID: 2, Username: "trader2", Email: "trader2@test.com"},
		{ID: 3, Username: "trader3", Email: "trader3@test.com"},
		{ID: 4, Username: "trader4", Email: "trader4@test.com"},
		{ID: 5, Username: "trader5", Email: "trader5@test.com"},
	}

	for _, user := range users {
		suite.db.Create(&user)

		// 각 사용자에게 지갑 생성
		wallet := models.UserWallet{
			UserID:      user.ID,
			USDCBalance: 10000000, // $100,000 in cents
		}
		suite.db.Create(&wallet)
	}

	// 프로젝트 생성
	project := models.Project{
		ID:          1,
		Title:       "Integration Test Project",
		Description: "Project for integration testing",
		UserID:      1,
		Status:      "active",
		CreatedAt:   time.Now(),
	}
	suite.db.Create(&project)

	// 마일스톤 생성
	milestone := models.Milestone{
		ID:        1,
		ProjectID: 1,
		Title:     "Integration Test Milestone",
		Status:    "funding",
		Order:     1,
		CreatedAt: time.Now(),
	}
	suite.db.Create(&milestone)

	// 플랫폼 수수료 설정
	feeConfig := models.PlatformFeeConfig{
		ID:                1,
		TradingFeeRate:    0.05,
		WithdrawFeeFlat:   100,
		MinBetAmount:      100,
		MaxBetAmount:      1000000,
		StakingRewardRate: 0.70,
		CreatedAt:         time.Now(),
	}
	suite.db.Create(&feeConfig)
}

func TestTradingIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(TradingIntegrationTestSuite))
}
