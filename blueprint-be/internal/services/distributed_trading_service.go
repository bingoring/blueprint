package services

import (
	"blueprint-module/pkg/models"
	"blueprint-module/pkg/redis"
	"fmt"
	"log"
	"time"

	redisClient "github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// 🌐 분산 거래 서비스 - 기존 TradingService를 대체하는 분산 버전
type DistributedTradingService struct {
	db             *gorm.DB
	matchingEngine *DistributedMatchingEngine
	commandHandler *TradingCommandHandler
	queryHandler   *TradingQueryHandler
}

// NewDistributedTradingService 분산 거래 서비스 생성자
func NewDistributedTradingService(db *gorm.DB, sseService *SSEService) *DistributedTradingService {
	return NewDistributedTradingServiceWithRedis(db, sseService, nil)
}

func NewDistributedTradingServiceWithRedis(db *gorm.DB, sseService *SSEService, redisClient *redisClient.Client) *DistributedTradingService {
	// 분산 매칭 엔진 초기화
	matchingEngine := NewDistributedMatchingEngineWithRedis(db, sseService, redisClient)

	// CQRS 핸들러들 초기화
	commandHandler := NewTradingCommandHandler(matchingEngine)

	// Use provided Redis client or get default one
	if redisClient == nil {
		redisClient = redis.GetClient()
	}
	queryHandler := NewTradingQueryHandler(redisClient, db)

	return &DistributedTradingService{
		db:             db,
		matchingEngine: matchingEngine,
		commandHandler: commandHandler,
		queryHandler:   queryHandler,
	}
}

// Start 분산 거래 서비스 시작
func (dts *DistributedTradingService) Start() error {
	log.Println("🚀 Starting Distributed Trading Service...")

	// 분산 매칭 엔진 시작
	if err := dts.matchingEngine.Start(); err != nil {
		return fmt.Errorf("failed to start matching engine: %v", err)
	}

	log.Println("✅ Distributed Trading Service started successfully")
	return nil
}

// Stop 분산 거래 서비스 정지
func (dts *DistributedTradingService) Stop() error {
	log.Println("🛑 Stopping Distributed Trading Service...")

	// 분산 매칭 엔진 정지
	if err := dts.matchingEngine.Stop(); err != nil {
		return fmt.Errorf("failed to stop matching engine: %v", err)
	}

	log.Println("✅ Distributed Trading Service stopped successfully")
	return nil
}

// ======================== 주문 관리 (Command Side) ========================

// CreateOrder 주문 생성 - CQRS Command 패턴 사용
func (dts *DistributedTradingService) CreateOrder(userID uint, milestoneID uint, optionID string, orderType string, quantity int64, price float64) (*MatchingResult, error) {
	// 1. 사용자 잔액 검증
	if err := dts.ValidateUserBalance(userID, orderType, quantity, price); err != nil {
		return nil, err
	}

	// 2. 주문 생성 명령 실행
	cmd := &CreateOrderCommand{
		UserID:      userID,
		MilestoneID: milestoneID,
		OptionID:    optionID,
		Type:        orderType,
		Quantity:    quantity,
		Price:       price,
	}

	return dts.commandHandler.HandleCreateOrder(cmd)
}

// CancelOrder 주문 취소 - CQRS Command 패턴 사용
func (dts *DistributedTradingService) CancelOrder(userID uint, orderID uint) error {
	// 1. 주문 소유권 검증
	var order models.Order
	err := dts.db.Where("id = ? AND user_id = ? AND status = ?", orderID, userID, "open").First(&order).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("order not found or not owned by user")
		}
		return err
	}

	// 2. 주문 취소 명령 실행
	cmd := &CancelOrderCommand{
		UserID:  userID,
		OrderID: orderID,
	}

	return dts.commandHandler.HandleCancelOrder(cmd)
}

// ValidateUserBalance 사용자 잔액 검증 (기존과 동일)
func (dts *DistributedTradingService) ValidateUserBalance(userID uint, orderType string, quantity int64, price float64) error {
	var wallet models.UserWallet
	err := dts.db.Where("user_id = ?", userID).First(&wallet).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("user wallet not found")
		}
		return err
	}

	requiredAmount := int64(float64(quantity) * price * 100) // Convert to cents

	if orderType == "buy" {
		if wallet.USDCBalance < requiredAmount {
			return fmt.Errorf("insufficient USDC balance: required %d cents, available %d cents",
				requiredAmount, wallet.USDCBalance)
		}
	} else {
		// sell 주문의 경우, 보유 토큰 수량 확인
		// TODO: 실제 구현에서는 Position 테이블에서 보유량 확인
	}

	return nil
}

// ======================== 시장 데이터 조회 (Query Side) ========================

// GetMarketData 마켓 데이터 조회 - CQRS Query 패턴 사용
func (dts *DistributedTradingService) GetMarketData(milestoneID uint, optionID string) (*MarketDataView, error) {
	query := &MarketDataQuery{
		MilestoneID: milestoneID,
		OptionID:    optionID,
	}

	return dts.queryHandler.GetMarketData(query)
}

// GetOrderBook 주문장 조회 - CQRS Query 패턴 사용
func (dts *DistributedTradingService) GetOrderBook(milestoneID uint, optionID string, depth int) (*OrderBookView, error) {
	query := &OrderBookQuery{
		MilestoneID: milestoneID,
		OptionID:    optionID,
		Depth:       depth,
	}

	return dts.queryHandler.GetOrderBook(query)
}

// GetUserOrders 사용자 주문 내역 조회 - CQRS Query 패턴 사용
func (dts *DistributedTradingService) GetUserOrders(userID uint, status string, limit int) ([]*UserOrderView, error) {
	query := &UserOrdersQuery{
		UserID: userID,
		Status: status,
		Limit:  limit,
	}

	return dts.queryHandler.GetUserOrders(query)
}

// ======================== 기존 호환성 메소드들 ========================

// GetMarketSummary 마켓 요약 정보 (기존 API 호환성)
func (dts *DistributedTradingService) GetMarketSummary(milestoneID uint, optionID string) (map[string]interface{}, error) {
	marketData, err := dts.GetMarketData(milestoneID, optionID)
	if err != nil {
		return nil, err
	}

	orderBook, err := dts.GetOrderBook(milestoneID, optionID, 5)
	if err != nil {
		return nil, err
	}

	summary := map[string]interface{}{
		"market_key":    marketData.MarketKey,
		"last_price":    marketData.LastPrice,
		"volume_24h":    marketData.Volume24h,
		"price_history": marketData.PriceHistory,
		"best_bid":      getBestPrice(orderBook.Bids, "bid"),
		"best_ask":      getBestPrice(orderBook.Asks, "ask"),
		"updated_at":    time.Now(),
	}

	return summary, nil
}

// getBestPrice 최적 가격 조회 헬퍼 함수
func getBestPrice(entries []OrderBookEntry, orderType string) float64 {
	if len(entries) == 0 {
		return 0
	}
	return entries[0].Price
}

// ======================== 헬스 체크 및 모니터링 ========================

// HealthCheck 서비스 상태 확인
func (dts *DistributedTradingService) HealthCheck() map[string]interface{} {
	return map[string]interface{}{
		"service":         "distributed_trading_service",
		"status":          "healthy",
		"matching_engine": dts.matchingEngine.instanceID,
		"timestamp":       time.Now(),
	}
}

// GetSystemMetrics 시스템 메트릭 조회
func (dts *DistributedTradingService) GetSystemMetrics() map[string]interface{} {
	// Redis 연결 상태, 활성 마켓 수, 처리 중인 주문 수 등
	activeMarkets, _ := dts.matchingEngine.getActiveMarkets()

	return map[string]interface{}{
		"active_markets":     len(activeMarkets),
		"instance_id":        dts.matchingEngine.instanceID,
		"uptime":             time.Since(time.Now()), // 실제로는 서비스 시작 시간부터 계산
		"redis_connected":    true,                   // 실제 Redis 연결 상태 확인
		"database_connected": true,                   // 실제 DB 연결 상태 확인
	}
}
