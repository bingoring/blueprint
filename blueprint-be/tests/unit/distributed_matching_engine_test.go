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

// DistributedMatchingEngineTestSuite 분산 매칭 엔진 테스트 슈트
type DistributedMatchingEngineTestSuite struct {
	suite.Suite
	engine      *services.DistributedMatchingEngine
	db          *gorm.DB
	redisServer *miniredis.Miniredis
	redisClient *redis.Client
}

// SetupSuite 테스트 슈트 초기화
func (suite *DistributedMatchingEngineTestSuite) SetupSuite() {
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

	// 분산 매칭 엔진 초기화 (테스트용 Redis 클라이언트 사용)
	suite.engine = services.NewDistributedMatchingEngineWithRedis(suite.db, nil, suite.redisClient)
}

// TearDownSuite 테스트 슈트 정리
func (suite *DistributedMatchingEngineTestSuite) TearDownSuite() {
	suite.redisServer.Close()
	suite.redisClient.Close()
}

// SetupTest 각 테스트 전 실행
func (suite *DistributedMatchingEngineTestSuite) SetupTest() {
	// Redis 데이터 초기화
	suite.redisServer.FlushAll()
}

// TestOrderBookCreation 주문장 생성 테스트
func (suite *DistributedMatchingEngineTestSuite) TestOrderBookCreation() {
	// 테스트 데이터 준비
	suite.createTestData()

	// 매칭 엔진 시작
	err := suite.engine.Start()
	suite.Assert().NoError(err)

	// 주문 생성
	order := &models.Order{
		UserID:      1,
		MilestoneID: 1,
		OptionID:    "success",
		Side:        models.OrderSideBuy,
		Quantity:    100,
		Price:       0.75,
		Status:      models.OrderStatusPending,
		CreatedAt:   time.Now(),
	}

	result, err := suite.engine.SubmitOrder(order)

	suite.Assert().NoError(err)
	suite.Assert().NotNil(result)
	suite.Assert().False(result.Executed) // 매칭되지 않았으므로 false

	// 매칭 엔진 정지
	err = suite.engine.Stop()
	suite.Assert().NoError(err)
}

// TestOrderMatching 주문 매칭 테스트
func (suite *DistributedMatchingEngineTestSuite) TestOrderMatching() {
	suite.createTestData()

	err := suite.engine.Start()
	suite.Require().NoError(err)
	defer suite.engine.Stop()

	// 매도 주문 먼저 생성
	sellOrder := &models.Order{
		UserID:      2,
		MilestoneID: 1,
		OptionID:    "success",
		Side:        models.OrderSideSell,
		Quantity:    50,
		Price:       0.70,
		Status:      models.OrderStatusPending,
		CreatedAt:   time.Now(),
	}

	_, err = suite.engine.SubmitOrder(sellOrder)
	suite.Require().NoError(err)

	// 매수 주문 생성 (더 높은 가격)
	buyOrder := &models.Order{
		UserID:      1,
		MilestoneID: 1,
		OptionID:    "success",
		Side:        models.OrderSideBuy,
		Quantity:    30,
		Price:       0.75,
		Status:      models.OrderStatusPending,
		CreatedAt:   time.Now(),
	}

	result, err := suite.engine.SubmitOrder(buyOrder)

	suite.Assert().NoError(err)
	suite.Assert().NotNil(result)
	suite.Assert().True(result.Executed) // 매칭되었으므로 true
	suite.Assert().Len(result.Trades, 1) // 1개의 거래 발생

	trade := result.Trades[0]
	suite.Assert().Equal(int64(30), trade.Quantity) // 30주 거래
	suite.Assert().Equal(0.70, trade.Price)         // 매도 주문 가격으로 거래
}

// TestConcurrentOrders 동시 주문 처리 테스트
func (suite *DistributedMatchingEngineTestSuite) TestConcurrentOrders() {
	suite.createTestData()

	err := suite.engine.Start()
	suite.Require().NoError(err)
	defer suite.engine.Stop()

	// 동시에 여러 주문 제출
	results := make(chan error, 10)

	for i := 0; i < 10; i++ {
		go func(orderIndex int) {
			order := &models.Order{
				UserID:      uint(orderIndex + 1),
				MilestoneID: 1,
				OptionID:    "success",
				Side:        models.OrderSideBuy,
				Quantity:    int64(10 * (orderIndex + 1)),
				Price:       0.75 + float64(orderIndex)*0.01,
				Status:      models.OrderStatusPending,
				CreatedAt:   time.Now(),
			}

			_, err := suite.engine.SubmitOrder(order)
			results <- err
		}(i)
	}

	// 모든 고루틴 완료 대기
	for i := 0; i < 10; i++ {
		err := <-results
		suite.Assert().NoError(err)
	}
}

// TestDistributedLocking 분산 락 테스트
func (suite *DistributedMatchingEngineTestSuite) TestDistributedLocking() {
	lockManager := services.NewDistributedLockManager(suite.redisClient)
	ctx := context.Background()

	// 첫 번째 인스턴스가 락 획득
	acquired, err := lockManager.AcquireLock(ctx, "test-market", 5*time.Second, "instance-1")
	suite.Assert().NoError(err)
	suite.Assert().True(acquired)

	// 두 번째 인스턴스가 같은 락 획득 시도 (실패해야 함)
	acquired2, err := lockManager.AcquireLock(ctx, "test-market", 5*time.Second, "instance-2")
	suite.Assert().NoError(err)
	suite.Assert().False(acquired2)

	// 첫 번째 인스턴스가 락 해제
	err = lockManager.ReleaseLock(ctx, "test-market", "instance-1")
	suite.Assert().NoError(err)

	// 이제 두 번째 인스턴스가 락 획득 가능
	acquired3, err := lockManager.AcquireLock(ctx, "test-market", 5*time.Second, "instance-2")
	suite.Assert().NoError(err)
	suite.Assert().True(acquired3)
}

// TestEventSourcing 이벤트 소싱 테스트
func (suite *DistributedMatchingEngineTestSuite) TestEventSourcing() {
	eventSourcing := services.NewOrderEventSourcing(suite.redisClient)
	ctx := context.Background()

	// 이벤트 생성 및 저장
	event := &services.OrderEvent{
		EventID:     "test-event-1",
		EventType:   services.EventOrderCreated,
		OrderID:     1,
		MilestoneID: 1,
		OptionID:    "success",
		Payload: map[string]interface{}{
			"user_id": 1,
			"price":   0.75,
		},
		Timestamp: time.Now().UnixMilli(),
		ServerID:  "test-instance",
		Version:   1,
	}

	err := eventSourcing.AppendEvent(ctx, "1:success", event)
	suite.Assert().NoError(err)

	// 이벤트 읽기
	events, err := eventSourcing.ReadEvents(ctx, "1:success", "0")
	suite.Assert().NoError(err)
	suite.Assert().Len(events, 1)
	suite.Assert().Equal("test-event-1", events[0].EventID)
}

// TestPriceOracle 가격 오라클 테스트
func (suite *DistributedMatchingEngineTestSuite) TestPriceOracle() {
	priceOracle := services.NewDistributedPriceOracle(suite.redisClient)
	ctx := context.Background()

	// 가격 업데이트
	err := priceOracle.UpdatePrice(ctx, "1:success", 0.75, 100)
	suite.Assert().NoError(err)

	// 가격 조회
	price, err := priceOracle.GetPrice(ctx, "1:success")
	suite.Assert().NoError(err)
	suite.Assert().Equal(0.75, price)
}

// createTestData 테스트 데이터 생성
func (suite *DistributedMatchingEngineTestSuite) createTestData() {
	// 사용자 생성
	users := []models.User{
		{ID: 1, Username: "trader1", Email: "trader1@test.com"},
		{ID: 2, Username: "trader2", Email: "trader2@test.com"},
		{ID: 3, Username: "trader3", Email: "trader3@test.com"},
	}

	for _, user := range users {
		suite.db.Create(&user)
		// 지갑 생성
		wallet := models.UserWallet{
			UserID:      user.ID,
			USDCBalance: 10000000, // $100,000 in cents
		}
		suite.db.Create(&wallet)
	}

	// 프로젝트 생성
	project := models.Project{
		ID:          1,
		Title:       "Test Project",
		Description: "Test project for unit tests",
		UserID:      1,
		Status:      "active",
		CreatedAt:   time.Now(),
	}
	suite.db.Create(&project)

	// 마일스톤 생성
	milestone := models.Milestone{
		ID:        1,
		ProjectID: 1,
		Title:     "Test Milestone",
		Status:    "funding",
		Order:     1,
		CreatedAt: time.Now(),
	}
	suite.db.Create(&milestone)
}

// TestDistributedMatchingEngineTestSuite 테스트 슈트 실행
func TestDistributedMatchingEngineTestSuite(t *testing.T) {
	suite.Run(t, new(DistributedMatchingEngineTestSuite))
}
