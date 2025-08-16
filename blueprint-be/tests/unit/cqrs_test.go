package unit_test

import (
	"context"
	"testing"
	"time"

	"blueprint-module/pkg/models"
	"blueprint/internal/services"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// CQRSTestSuite CQRS 패턴 테스트 슈트
type CQRSTestSuite struct {
	suite.Suite
	db             *gorm.DB
	redisServer    *miniredis.Miniredis
	redisClient    *redis.Client
	commandHandler *services.TradingCommandHandler
	queryHandler   *services.TradingQueryHandler
	matchingEngine *services.DistributedMatchingEngine
}

func (suite *CQRSTestSuite) SetupSuite() {
	// In-memory SQLite DB 설정
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	suite.Require().NoError(err)
	suite.db = db

	// 테이블 마이그레이션
	err = db.AutoMigrate(
		&models.User{},
		&models.Project{},
		&models.Milestone{},
		&models.Order{},
		&models.Trade{},
		&models.Position{},
		&models.MarketData{},
		&models.UserWallet{},
	)
	suite.Require().NoError(err)

	// Mock Redis 서버 설정
	suite.redisServer = miniredis.RunT(suite.T())
	suite.redisClient = redis.NewClient(&redis.Options{
		Addr: suite.redisServer.Addr(),
	})

	// CQRS 컴포넌트 초기화
	suite.matchingEngine = services.NewDistributedMatchingEngine(suite.db, nil)
	suite.commandHandler = services.NewTradingCommandHandler(suite.matchingEngine)
	suite.queryHandler = services.NewTradingQueryHandler(suite.redisClient, suite.db)

	// 테스트 데이터 생성
	suite.createTestData()
}

func (suite *CQRSTestSuite) TearDownSuite() {
	suite.redisServer.Close()
	suite.redisClient.Close()
}

func (suite *CQRSTestSuite) SetupTest() {
	suite.redisServer.FlushAll()
}

// TestCreateOrderCommand 주문 생성 명령 테스트
func (suite *CQRSTestSuite) TestCreateOrderCommand() {
	// 매칭 엔진 시작
	err := suite.matchingEngine.Start()
	suite.Require().NoError(err)
	defer suite.matchingEngine.Stop()

	// 주문 생성 명령
	cmd := &services.CreateOrderCommand{
		UserID:      1,
		MilestoneID: 1,
		OptionID:    "success",
		Type:        "buy",
		Quantity:    100,
		Price:       0.75,
	}

	// 명령 실행
	result, err := suite.commandHandler.HandleCreateOrder(cmd)

	suite.Assert().NoError(err)
	suite.Assert().NotNil(result)
	suite.Assert().False(result.Executed) // 매칭 상대 없으므로 false
}

// TestCancelOrderCommand 주문 취소 명령 테스트
func (suite *CQRSTestSuite) TestCancelOrderCommand() {
	// 주문 생성
	order := models.Order{
		ID:          1,
		UserID:      1,
		MilestoneID: 1,
		OptionID:    "success",
		Side:        models.OrderSideBuy,
		Quantity:    100,
		Price:       0.75,
		Status:      models.OrderStatusPending,
		CreatedAt:   time.Now(),
	}
	suite.db.Create(&order)

	// 주문 취소 명령
	cmd := &services.CancelOrderCommand{
		UserID:  1,
		OrderID: 1,
	}

	// 명령 실행
	err := suite.commandHandler.HandleCancelOrder(cmd)
	suite.Assert().NoError(err)
}

// TestMarketDataQuery 마켓 데이터 조회 테스트
func (suite *CQRSTestSuite) TestMarketDataQuery() {
	// Redis에 테스트 데이터 설정
	ctx := context.Background()
	suite.redisClient.Set(ctx, "price:1:success", "0.75", 0)
	suite.redisClient.Set(ctx, "volume:1:success", "1000", 0)
	suite.redisClient.ZAdd(ctx, "history:1:success",
		redis.Z{Score: float64(time.Now().Unix()), Member: "0.75"})

	// 마켓 데이터 조회 쿼리
	query := &services.MarketDataQuery{
		MilestoneID: 1,
		OptionID:    "success",
	}

	// 쿼리 실행
	marketData, err := suite.queryHandler.GetMarketData(query)

	suite.Assert().NoError(err)
	suite.Assert().NotNil(marketData)
	suite.Assert().Equal(uint(1), marketData.MilestoneID)
	suite.Assert().Equal("success", marketData.OptionID)
	suite.Assert().Equal(0.75, marketData.LastPrice)
	suite.Assert().Equal(int64(1000), marketData.Volume24h)
}

// TestOrderBookQuery 주문장 조회 테스트
func (suite *CQRSTestSuite) TestOrderBookQuery() {
	// Redis에 주문장 데이터 설정 (빈 주문장으로 테스트)
	ctx := context.Background()
	orderBookJSON, _ := suite.redisClient.Get(ctx, "orderbook:1:success").Result()
	suite.Assert().Equal("", orderBookJSON) // 비어있음

	// 주문장 조회 쿼리
	query := &services.OrderBookQuery{
		MilestoneID: 1,
		OptionID:    "success",
		Depth:       10,
	}

	// 쿼리 실행
	orderBookView, err := suite.queryHandler.GetOrderBook(query)

	suite.Assert().NoError(err)
	suite.Assert().NotNil(orderBookView)
	suite.Assert().Equal("1:success", orderBookView.MarketKey)
	suite.Assert().Len(orderBookView.Bids, 0) // 빈 주문장
	suite.Assert().Len(orderBookView.Asks, 0)
}

// TestUserOrdersQuery 사용자 주문 조회 테스트
func (suite *CQRSTestSuite) TestUserOrdersQuery() {
	// 테스트 주문 생성
	orders := []models.Order{
		{
			ID:          1,
			UserID:      1,
			MilestoneID: 1,
			OptionID:    "success",
			Side:        models.OrderSideBuy,
			Quantity:    100,
			Price:       0.75,
			Status:      models.OrderStatusPending,
			CreatedAt:   time.Now(),
		},
		{
			ID:          2,
			UserID:      1,
			MilestoneID: 1,
			OptionID:    "fail",
			Side:        models.OrderSideSell,
			Quantity:    50,
			Price:       0.25,
			Status:      models.OrderStatusFilled,
			CreatedAt:   time.Now().Add(-time.Hour),
		},
	}

	for _, order := range orders {
		suite.db.Create(&order)
	}

	// 사용자 주문 조회 쿼리
	query := &services.UserOrdersQuery{
		UserID: 1,
		Status: "", // 모든 상태
		Limit:  10,
	}

	// 쿼리 실행
	userOrders, err := suite.queryHandler.GetUserOrders(query)

	suite.Assert().NoError(err)
	suite.Assert().Len(userOrders, 2)

	// 최신 주문이 먼저 오는지 확인 (ORDER BY created_at DESC)
	suite.Assert().Equal(uint(1), userOrders[0].ID)
	suite.Assert().Equal(uint(2), userOrders[1].ID)
}

// TestUserOrdersQueryWithFilter 필터가 적용된 사용자 주문 조회 테스트
func (suite *CQRSTestSuite) TestUserOrdersQueryWithFilter() {
	// 테스트 주문 생성 (이미 위에서 생성됨)

	// 특정 상태만 조회
	query := &services.UserOrdersQuery{
		UserID: 1,
		Status: string(models.OrderStatusPending),
		Limit:  10,
	}

	_, err := suite.queryHandler.GetUserOrders(query)

	suite.Assert().NoError(err)
	// 주문이 있을 경우 검증 로직 (실제 데이터에 따라 조정)
}

// TestCommandQuerySeparation 명령-조회 분리 테스트
func (suite *CQRSTestSuite) TestCommandQuerySeparation() {
	// 매칭 엔진 시작
	err := suite.matchingEngine.Start()
	suite.Require().NoError(err)
	defer suite.matchingEngine.Stop()

	// 1. Command: 주문 생성
	createCmd := &services.CreateOrderCommand{
		UserID:      1,
		MilestoneID: 1,
		OptionID:    "success",
		Type:        "buy",
		Quantity:    100,
		Price:       0.75,
	}

	_, err = suite.commandHandler.HandleCreateOrder(createCmd)
	suite.Assert().NoError(err)

	// 잠시 대기 (이벤트 처리 시간)
	time.Sleep(100 * time.Millisecond)

	// 2. Query: 사용자 주문 조회
	userQuery := &services.UserOrdersQuery{
		UserID: 1,
		Limit:  10,
	}

	_, err = suite.queryHandler.GetUserOrders(userQuery)
	suite.Assert().NoError(err)

	// 명령으로 생성한 주문이 조회에서 나타나는지 확인
	// (실제 환경에서는 eventual consistency 고려 필요)
}

// TestValidation 유효성 검사 테스트
func (suite *CQRSTestSuite) TestValidation() {
	// 잘못된 명령 테스트
	invalidCmd := &services.CreateOrderCommand{
		UserID:      0, // 잘못된 사용자 ID
		MilestoneID: 1,
		OptionID:    "success",
		Type:        "invalid", // 잘못된 타입
		Quantity:    -100,      // 음수 수량
		Price:       -0.75,     // 음수 가격
	}

	_, err := suite.commandHandler.HandleCreateOrder(invalidCmd)
	suite.Assert().Error(err)
	suite.Assert().Contains(err.Error(), "invalid command")
}

func (suite *CQRSTestSuite) createTestData() {
	// 사용자 생성
	user := models.User{
		ID:       1,
		Username: "testuser",
		Email:    "test@example.com",
	}
	suite.db.Create(&user)

	// 지갑 생성
	wallet := models.UserWallet{
		UserID:      1,
		USDCBalance: 10000000, // $100,000 in cents
	}
	suite.db.Create(&wallet)

	// 프로젝트 생성
	project := models.Project{
		ID:     1,
		Title:  "Test Project",
		UserID: 1,
		Status: "active",
	}
	suite.db.Create(&project)

	// 마일스톤 생성
	milestone := models.Milestone{
		ID:        1,
		ProjectID: 1,
		Title:     "Test Milestone",
		Status:    "funding",
		Order:     1,
	}
	suite.db.Create(&milestone)
}

func TestCQRSTestSuite(t *testing.T) {
	suite.Run(t, new(CQRSTestSuite))
}
