package services

import (
	"blueprint-module/pkg/models"
	"blueprint-module/pkg/queue"
	"fmt"
	"log"
	"math"
	"sync"
	"time"

	"gorm.io/gorm"
)

// MarketMakerBot 폴리마켓 스타일 마켓메이커 봇
type MarketMakerBot struct {
	db             *gorm.DB
	tradingService *TradingService
	queuePublisher *queue.Publisher

	// 봇 설정
	isRunning bool
	stopChan  chan struct{}
	mutex     sync.RWMutex

	// 마켓 메이킹 설정
	config        MarketMakerConfig
	activeMarkets map[string]*MarketInfo // milestone_id:option_id -> MarketInfo

	// 성과 추적
	stats MarketMakerStats
}

// MarketMakerConfig 마켓메이커 설정
type MarketMakerConfig struct {
	UserID           uint    `json:"user_id"`           // 마켓메이커 봇 사용자 ID
	MinSpread        float64 `json:"min_spread"`        // 최소 스프레드 (0.01 = 1%)
	MaxSpread        float64 `json:"max_spread"`        // 최대 스프레드 (0.05 = 5%)
	BaseOrderSize    int64   `json:"base_order_size"`   // 기본 주문 수량
	MaxOrderSize     int64   `json:"max_order_size"`    // 최대 주문 수량
	MinPrice         float64 `json:"min_price"`         // 최소 가격 (0.01)
	MaxPrice         float64 `json:"max_price"`         // 최대 가격 (0.99)
	RefreshInterval  int     `json:"refresh_interval"`  // 주문 갱신 주기 (초)
	VolatilityFactor float64 `json:"volatility_factor"` // 변동성 기반 스프레드 조정
	InventoryLimit   int64   `json:"inventory_limit"`   // 포지션 한도
	RiskTolerance    float64 `json:"risk_tolerance"`    // 리스크 허용도
	EnabledMarkets   []uint  `json:"enabled_markets"`   // 활성화된 마일스톤 ID들
}

// MarketInfo 개별 마켓 정보
type MarketInfo struct {
	MilestoneID   uint                   `json:"milestone_id"`
	OptionID      string                 `json:"option_id"`
	CurrentPrice  float64                `json:"current_price"`
	LastUpdate    time.Time              `json:"last_update"`
	Volatility    float64                `json:"volatility"`
	Volume24h     int64                  `json:"volume_24h"`
	Spread        float64                `json:"spread"`
	BidPrice      float64                `json:"bid_price"`
	AskPrice      float64                `json:"ask_price"`
	Position      int64                  `json:"position"`      // 현재 포지션 (+매수, -매도)
	ActiveOrders  []uint                 `json:"active_orders"` // 활성 주문 ID들
	LastTradeTime time.Time              `json:"last_trade_time"`
	PriceHistory  []float64              `json:"price_history"` // 최근 가격 히스토리 (변동성 계산용)
	Metadata      map[string]interface{} `json:"metadata"`
}

// MarketMakerStats 마켓메이커 성과 통계
type MarketMakerStats struct {
	StartTime             time.Time `json:"start_time"`
	TotalProfit           int64     `json:"total_profit"`
	TotalVolume           int64     `json:"total_volume"`
	TotalTrades           int64     `json:"total_trades"`
	SuccessfulTrades      int64     `json:"successful_trades"`
	FailedTrades          int64     `json:"failed_trades"`
	AverageProfitPerTrade int64     `json:"avg_profit_per_trade"`
	MaxDrawdown           int64     `json:"max_drawdown"`
	SharpeRatio           float64   `json:"sharpe_ratio"`
	ActiveMarkets         int       `json:"active_markets"`
	TotalOrdersPlaced     int64     `json:"total_orders_placed"`
	OrderCancelRate       float64   `json:"order_cancel_rate"`
}

// NewMarketMakerBot 마켓메이커 봇 생성자
func NewMarketMakerBot(db *gorm.DB, tradingService *TradingService) *MarketMakerBot {
	return &MarketMakerBot{
		db:             db,
		tradingService: tradingService,
		queuePublisher: queue.NewPublisher(),
		stopChan:       make(chan struct{}),
		activeMarkets:  make(map[string]*MarketInfo),
		config: MarketMakerConfig{
			UserID:           1,    // 시스템 봇 계정
			MinSpread:        0.02, // 2%
			MaxSpread:        0.08, // 8%
			BaseOrderSize:    10,
			MaxOrderSize:     100,
			MinPrice:         0.05,
			MaxPrice:         0.95,
			RefreshInterval:  5, // 30초마다 갱신
			VolatilityFactor: 2.0,
			InventoryLimit:   1000,
			RiskTolerance:    0.1,
		},
		stats: MarketMakerStats{
			StartTime: time.Now(),
		},
	}
}

// Start 마켓메이커 봇 시작
func (mm *MarketMakerBot) Start() error {
	mm.mutex.Lock()
	defer mm.mutex.Unlock()

	if mm.isRunning {
		return fmt.Errorf("market maker bot is already running")
	}

	mm.isRunning = true
	log.Println("🤖 Market Maker Bot started!")

	// 초기 마켓 스캔 (지연 후 실행)
	go func() {
		log.Printf("🤖 Market maker will start scanning in 15 seconds...")
		time.Sleep(15 * time.Second) // 15초 대기하여 모든 서비스가 완전히 준비될 시간 제공
		log.Printf("🤖 Starting market scan...")
		if err := mm.scanActiveMarkets(); err != nil {
			log.Printf("❌ Error scanning markets: %v", err)
		}
	}()

	// 메인 루프 시작
	go mm.mainLoop()

	// 통계 출력 루프
	go mm.statsLoop()

	return nil
}

// Stop 마켓메이커 봇 중지
func (mm *MarketMakerBot) Stop() error {
	mm.mutex.Lock()
	defer mm.mutex.Unlock()

	if !mm.isRunning {
		return fmt.Errorf("market maker bot is not running")
	}

	mm.isRunning = false
	close(mm.stopChan)

	// 모든 활성 주문 취소
	mm.cancelAllOrders()

	log.Println("🛑 Market Maker Bot stopped!")
	return nil
}

// mainLoop 메인 실행 루프
func (mm *MarketMakerBot) mainLoop() {
	ticker := time.NewTicker(time.Duration(mm.config.RefreshInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-mm.stopChan:
			return
		case <-ticker.C:
			mm.runMarketMakingCycle()
		}
	}
}

// runMarketMakingCycle 마켓메이킹 사이클 실행
func (mm *MarketMakerBot) runMarketMakingCycle() {
	mm.mutex.Lock()
	defer mm.mutex.Unlock()

	// 1. 마켓 상태 업데이트
	mm.updateMarketStates()

	// 2. 기존 주문 관리
	mm.manageExistingOrders()

	// 3. 새로운 주문 생성
	mm.placeNewOrders()

	// 4. 리스크 관리
	mm.performRiskManagement()

	// 5. 통계 업데이트
	mm.updateStats()
}

// scanActiveMarkets 활성 마켓 스캔
func (mm *MarketMakerBot) scanActiveMarkets() error {
	var milestones []models.Milestone

	// 활성화된 마일스톤들 조회
	err := mm.db.Where("status = ? AND target_date > ?",
		models.MilestoneStatusPending, time.Now()).Find(&milestones).Error
	if err != nil {
		return err
	}

	for _, milestone := range milestones {
		// 설정에서 활성화된 마켓만 처리
		if len(mm.config.EnabledMarkets) > 0 {
			enabled := false
			for _, id := range mm.config.EnabledMarkets {
				if id == milestone.ID {
					enabled = true
					break
				}
			}
			if !enabled {
				continue
			}
		}

		// 성공/실패 두 옵션에 대해 마켓 정보 생성
		for _, option := range []string{"success", "fail"} {
			key := fmt.Sprintf("%d:%s", milestone.ID, option)

			if _, exists := mm.activeMarkets[key]; !exists {
				// 현재 시장 가격 조회
				currentPrice := mm.getCurrentPrice(milestone.ID, option)

				mm.activeMarkets[key] = &MarketInfo{
					MilestoneID:  milestone.ID,
					OptionID:     option,
					CurrentPrice: currentPrice,
					LastUpdate:   time.Now(),
					Volatility:   0.05, // 기본 변동성 5%
					Spread:       mm.config.MinSpread,
					ActiveOrders: make([]uint, 0),
					PriceHistory: make([]float64, 0),
					Metadata:     make(map[string]interface{}),
				}

				// 🎯 새 마켓에 초기 유동성 제공
				go mm.provideInitialLiquidity(milestone.ID, option, currentPrice)

				log.Printf("🎯 Added market: %s (price: %.4f)", key, currentPrice)
			}
		}
	}

	log.Printf("📊 Market scan completed. Found %d active markets", len(mm.activeMarkets))
	return nil
}

// updateMarketStates 마켓 상태 업데이트
func (mm *MarketMakerBot) updateMarketStates() {
	for _, market := range mm.activeMarkets {
		// 현재 가격 업데이트
		newPrice := mm.getCurrentPrice(market.MilestoneID, market.OptionID)
		if newPrice > 0 {
			// 가격 히스토리 업데이트 (최대 100개 유지)
			market.PriceHistory = append(market.PriceHistory, newPrice)
			if len(market.PriceHistory) > 100 {
				market.PriceHistory = market.PriceHistory[1:]
			}

			// 변동성 계산
			market.Volatility = mm.calculateVolatility(market.PriceHistory)

			// 가격 변동시 스프레드 조정
			if math.Abs(newPrice-market.CurrentPrice) > 0.01 {
				market.Spread = mm.calculateOptimalSpread(market)
			}

			market.CurrentPrice = newPrice
			market.LastUpdate = time.Now()
		}

		// 포지션 업데이트
		market.Position = mm.getCurrentPosition(market.MilestoneID, market.OptionID)

		// 24시간 거래량 업데이트
		market.Volume24h = mm.getVolume24h(market.MilestoneID, market.OptionID)
	}
}

// manageExistingOrders 기존 주문 관리
func (mm *MarketMakerBot) manageExistingOrders() {
	for _, market := range mm.activeMarkets {
		var ordersToCancel []uint

		for _, orderID := range market.ActiveOrders {
			order := mm.getOrder(orderID)
			if order == nil {
				// 주문이 체결되었거나 취소됨
				continue
			}

			// 가격이 크게 변동했거나 오래된 주문 취소
			shouldCancel := false

			// 1. 가격 변동 체크
			if order.Side == models.OrderSideBuy {
				if order.Price < market.CurrentPrice*(1-market.Spread*2) {
					shouldCancel = true
				}
			} else {
				if order.Price > market.CurrentPrice*(1+market.Spread*2) {
					shouldCancel = true
				}
			}

			// 2. 시간 체크 (30분 이상 된 주문)
			if time.Since(order.CreatedAt) > 30*time.Minute {
				shouldCancel = true
			}

			// 3. 리스크 체크 (포지션이 한도 초과)
			if math.Abs(float64(market.Position)) > float64(mm.config.InventoryLimit) {
				if (market.Position > 0 && order.Side == models.OrderSideBuy) ||
					(market.Position < 0 && order.Side == models.OrderSideSell) {
					shouldCancel = true
				}
			}

			if shouldCancel {
				ordersToCancel = append(ordersToCancel, orderID)
			}
		}

		// 주문 취소 실행
		for _, orderID := range ordersToCancel {
			mm.cancelOrder(orderID)
			mm.removeOrderFromMarket(market, orderID)
		}
	}
}

// placeNewOrders 새로운 주문 생성
func (mm *MarketMakerBot) placeNewOrders() {
	for _, market := range mm.activeMarkets {
		// 활성 주문이 너무 많으면 스킵
		if len(market.ActiveOrders) >= 4 { // 최대 4개 주문 (매수2, 매도2)
			continue
		}

		// 매수/매도 주문 생성 조건 (균형 잡힌 접근)
		shouldPlaceBuyOrder := len(market.ActiveOrders) < 2  // 최대 2개 주문만
		shouldPlaceSellOrder := len(market.ActiveOrders) < 2 // 최대 2개 주문만

		// 현재 가격 기준으로 Bid/Ask 가격 계산
		bidPrice := market.CurrentPrice * (1 - market.Spread)
		askPrice := market.CurrentPrice * (1 + market.Spread)

		// 가격 범위 제한
		bidPrice = math.Max(bidPrice, mm.config.MinPrice)
		askPrice = math.Min(askPrice, mm.config.MaxPrice)

		// 주문 수량 계산 (변동성과 포지션에 따라 조정)
		orderSize := mm.calculateOrderSize(market)

		// 매수 주문 생성
		if shouldPlaceBuyOrder && bidPrice > mm.config.MinPrice {
			buyOrderID := mm.placeOrder(market.MilestoneID, market.OptionID,
				models.OrderSideBuy, orderSize, bidPrice)
			if buyOrderID > 0 {
				market.ActiveOrders = append(market.ActiveOrders, buyOrderID)
				market.BidPrice = bidPrice
			}
		}

		// 매도 주문 생성
		if shouldPlaceSellOrder && askPrice < mm.config.MaxPrice {
			sellOrderID := mm.placeOrder(market.MilestoneID, market.OptionID,
				models.OrderSideSell, orderSize, askPrice)
			if sellOrderID > 0 {
				market.ActiveOrders = append(market.ActiveOrders, sellOrderID)
				market.AskPrice = askPrice
			}
		}

		// 마켓메이킹 이벤트 발행
		mm.queuePublisher.EnqueueMarketMakeWork(market.MilestoneID, market.OptionID,
			queue.MarketMakeEventData{
				Action:       "create_orders",
				CurrentPrice: market.CurrentPrice,
				Spread:       market.Spread,
				Volume:       market.Volume24h,
			})
	}
}

// calculateOptimalSpread 최적 스프레드 계산
func (mm *MarketMakerBot) calculateOptimalSpread(market *MarketInfo) float64 {
	// 기본 스프레드
	baseSpread := mm.config.MinSpread

	// 변동성 기반 조정
	volatilityAdjustment := market.Volatility * mm.config.VolatilityFactor

	// 포지션 기반 조정 (포지션이 클수록 스프레드 증가)
	positionRatio := math.Abs(float64(market.Position)) / float64(mm.config.InventoryLimit)
	positionAdjustment := positionRatio * 0.02 // 최대 2% 추가

	// 거래량 기반 조정 (거래량이 적을수록 스프레드 증가)
	volumeAdjustment := 0.0
	if market.Volume24h < 100 {
		volumeAdjustment = 0.01 // 1% 추가
	}

	// 최종 스프레드 계산
	finalSpread := baseSpread + volatilityAdjustment + positionAdjustment + volumeAdjustment

	// 범위 제한
	finalSpread = math.Max(finalSpread, mm.config.MinSpread)
	finalSpread = math.Min(finalSpread, mm.config.MaxSpread)

	return finalSpread
}

// calculateOrderSize 주문 수량 계산
func (mm *MarketMakerBot) calculateOrderSize(market *MarketInfo) int64 {
	baseSize := mm.config.BaseOrderSize

	// 변동성에 따른 조정 (변동성이 높을수록 수량 감소)
	volatilityFactor := 1.0 - market.Volatility
	if volatilityFactor < 0.3 {
		volatilityFactor = 0.3
	}

	// 거래량에 따른 조정 (거래량이 많을수록 수량 증가)
	volumeFactor := 1.0
	if market.Volume24h > 1000 {
		volumeFactor = 1.5
	} else if market.Volume24h > 500 {
		volumeFactor = 1.2
	}

	finalSize := int64(float64(baseSize) * volatilityFactor * volumeFactor)

	// 범위 제한
	if finalSize < 1 {
		finalSize = 1
	}
	if finalSize > mm.config.MaxOrderSize {
		finalSize = mm.config.MaxOrderSize
	}

	return finalSize
}

// Helper functions (simplified implementations)

func (mm *MarketMakerBot) getCurrentPrice(milestoneID uint, optionID string) float64 {
	var marketData models.MarketData
	err := mm.db.Where("milestone_id = ? AND option_id = ?", milestoneID, optionID).
		First(&marketData).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// 새로운 마켓이므로 기본 가격 사용 (로그 없음)
			return 0.5
		}
		// 다른 에러인 경우에만 로그 출력
		log.Printf("⚠️ Error getting market price for %d:%s: %v", milestoneID, optionID, err)
		return 0.5
	}

	return marketData.CurrentPrice
}

func (mm *MarketMakerBot) calculateVolatility(prices []float64) float64 {
	if len(prices) < 2 {
		return 0.05 // 기본 변동성
	}

	// 단순 변동성 계산 (표준편차)
	var sum, mean, variance float64
	n := float64(len(prices))

	for _, price := range prices {
		sum += price
	}
	mean = sum / n

	for _, price := range prices {
		variance += math.Pow(price-mean, 2)
	}
	variance /= n

	return math.Sqrt(variance) / mean // 상대 변동성
}

func (mm *MarketMakerBot) getCurrentPosition(milestoneID uint, optionID string) int64 {
	var position models.Position
	err := mm.db.Where("user_id = ? AND milestone_id = ? AND option_id = ?",
		mm.config.UserID, milestoneID, optionID).First(&position).Error
	if err != nil {
		return 0
	}
	return position.Quantity
}

func (mm *MarketMakerBot) getVolume24h(milestoneID uint, optionID string) int64 {
	var result struct {
		TotalVolume int64
	}

	mm.db.Model(&models.Trade{}).
		Select("COALESCE(SUM(quantity), 0) as total_volume").
		Where("milestone_id = ? AND option_id = ? AND created_at > ?",
			milestoneID, optionID, time.Now().Add(-24*time.Hour)).
		Scan(&result)

	return result.TotalVolume
}

func (mm *MarketMakerBot) placeOrder(milestoneID uint, optionID string, side models.OrderSide, quantity int64, price float64) uint {
	// milestone에서 project_id 조회
	var milestone struct {
		ProjectID uint `json:"project_id"`
	}

	if err := mm.db.Table("milestones").
		Select("project_id").
		Where("id = ?", milestoneID).
		First(&milestone).Error; err != nil {
		log.Printf("❌ Failed to get project_id for milestone %d: %v", milestoneID, err)
		return 0
	}

	request := models.CreateOrderRequest{
		ProjectID:   milestone.ProjectID, // 올바른 project_id 설정
		MilestoneID: milestoneID,
		OptionID:    optionID,
		Type:        models.OrderTypeLimit,
		Side:        side,
		Quantity:    quantity,
		Price:       price,
	}

	response, err := mm.tradingService.CreateOrder(mm.config.UserID, request, "system", "market-maker-bot")
	if err != nil {
		log.Printf("❌ Failed to place order: %v", err)
		return 0
	}

	mm.stats.TotalOrdersPlaced++
	log.Printf("📝 Order placed: %s %d@%.4f for %s", side, quantity, price, optionID)

	return response.Order.ID
}

func (mm *MarketMakerBot) cancelOrder(orderID uint) error {
	// 주문 취소 로직 구현
	err := mm.db.Model(&models.Order{}).Where("id = ?", orderID).
		Update("status", models.OrderStatusCancelled).Error
	if err != nil {
		return err
	}

	log.Printf("❌ Order cancelled: %d", orderID)
	return nil
}

func (mm *MarketMakerBot) getOrder(orderID uint) *models.Order {
	var order models.Order
	err := mm.db.Where("id = ?", orderID).First(&order).Error
	if err != nil {
		return nil
	}
	return &order
}

func (mm *MarketMakerBot) removeOrderFromMarket(market *MarketInfo, orderID uint) {
	for i, id := range market.ActiveOrders {
		if id == orderID {
			market.ActiveOrders = append(market.ActiveOrders[:i], market.ActiveOrders[i+1:]...)
			break
		}
	}
}

func (mm *MarketMakerBot) cancelAllOrders() {
	for _, market := range mm.activeMarkets {
		for _, orderID := range market.ActiveOrders {
			mm.cancelOrder(orderID)
		}
		market.ActiveOrders = make([]uint, 0)
	}
}

func (mm *MarketMakerBot) performRiskManagement() {
	// 리스크 관리 로직 (포지션 한도, 손실 제한 등)
	for _, market := range mm.activeMarkets {
		// 포지션이 한도를 초과하면 반대 주문만 생성하도록 설정
		if math.Abs(float64(market.Position)) > float64(mm.config.InventoryLimit)*0.9 {
			log.Printf("⚠️ Position limit approaching for %s: %d", market.OptionID, market.Position)
		}
	}
}

func (mm *MarketMakerBot) updateStats() {
	mm.stats.ActiveMarkets = len(mm.activeMarkets)

	// 수익률 계산 등 추가 통계 업데이트
	// (실제 구현에서는 더 정교한 수익률 계산 필요)
}

func (mm *MarketMakerBot) statsLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-mm.stopChan:
			return
		case <-ticker.C:
			mm.printStats()
		}
	}
}

func (mm *MarketMakerBot) printStats() {
	mm.mutex.RLock()
	defer mm.mutex.RUnlock()

	log.Printf("📊 Market Maker Stats:")
	log.Printf("   Active Markets: %d", mm.stats.ActiveMarkets)
	log.Printf("   Total Orders: %d", mm.stats.TotalOrdersPlaced)
	log.Printf("   Total Trades: %d", mm.stats.TotalTrades)
	log.Printf("   Runtime: %v", time.Since(mm.stats.StartTime))
}

// GetConfig 설정 조회
func (mm *MarketMakerBot) GetConfig() MarketMakerConfig {
	mm.mutex.RLock()
	defer mm.mutex.RUnlock()
	return mm.config
}

// UpdateConfig 설정 업데이트
func (mm *MarketMakerBot) UpdateConfig(config MarketMakerConfig) {
	mm.mutex.Lock()
	defer mm.mutex.Unlock()
	mm.config = config
	log.Println("🔧 Market Maker config updated")
}

// GetStats 통계 조회
func (mm *MarketMakerBot) GetStats() MarketMakerStats {
	mm.mutex.RLock()
	defer mm.mutex.RUnlock()
	return mm.stats
}

// GetActiveMarkets 활성 마켓 조회
func (mm *MarketMakerBot) GetActiveMarkets() map[string]*MarketInfo {
	mm.mutex.RLock()
	defer mm.mutex.RUnlock()

	result := make(map[string]*MarketInfo)
	for k, v := range mm.activeMarkets {
		result[k] = v
	}
	return result
}

// IsRunning 실행 상태 확인
func (mm *MarketMakerBot) IsRunning() bool {
	mm.mutex.RLock()
	defer mm.mutex.RUnlock()
	return mm.isRunning
}

// provideInitialLiquidity 새 마켓에 초기 유동성 제공
func (mm *MarketMakerBot) provideInitialLiquidity(milestoneID uint, optionID string, currentPrice float64) {
	// 🔍 마일스톤에서 프로젝트 ID 조회
	var milestone models.Milestone
	if err := mm.db.Where("id = ?", milestoneID).First(&milestone).Error; err != nil {
		log.Printf("❌ Failed to get milestone %d: %v", milestoneID, err)
		return
	}

	// 🔍 MarketData가 존재하는지 확인
	var marketData models.MarketData
	if err := mm.db.Where("milestone_id = ? AND option_id = ?", milestoneID, optionID).First(&marketData).Error; err != nil {
		log.Printf("⚠️ MarketData not found for %d:%s, skipping liquidity provision", milestoneID, optionID)
		return
	}

	// 현재 가격 주변에 매수/매도 주문 생성
	spread := mm.config.MinSpread
	bidPrice := currentPrice - spread/2
	askPrice := currentPrice + spread/2

	// 가격 범위 검증
	if bidPrice < mm.config.MinPrice {
		bidPrice = mm.config.MinPrice
	}
	if askPrice > mm.config.MaxPrice {
		askPrice = mm.config.MaxPrice
	}

	// 매수 주문 생성
	buyOrder := models.CreateOrderRequest{
		ProjectID:   milestone.ProjectID, // 마일스톤에서 프로젝트 ID 가져오기
		MilestoneID: milestoneID,
		OptionID:    optionID,
		Type:        models.OrderTypeLimit,
		Side:        models.OrderSideBuy,
		Quantity:    mm.config.BaseOrderSize,
		Price:       bidPrice,
		Currency:    models.CurrencyUSDC,
	}

	// 매도 주문 생성
	sellOrder := models.CreateOrderRequest{
		ProjectID:   milestone.ProjectID, // 마일스톤에서 프로젝트 ID 가져오기
		MilestoneID: milestoneID,
		OptionID:    optionID,
		Type:        models.OrderTypeLimit,
		Side:        models.OrderSideSell,
		Quantity:    mm.config.BaseOrderSize,
		Price:       askPrice,
		Currency:    models.CurrencyUSDC,
	}

	log.Printf("🤖 Providing initial liquidity for %s: bid=%.2f¢, ask=%.2f¢",
		optionID, bidPrice*100, askPrice*100)

	// 🔍 마켓메이커 봇 지갑 확인/생성
	mm.ensureMarketMakerWallet()

	// 주문 생성 (에러 발생 시 로그만 출력)
	if _, err := mm.tradingService.CreateOrder(mm.config.UserID, buyOrder, "market-maker", "market-maker-bot"); err != nil {
		log.Printf("❌ Failed to create initial buy order: %v", err)
	}

	if _, err := mm.tradingService.CreateOrder(mm.config.UserID, sellOrder, "market-maker", "market-maker-bot"); err != nil {
		log.Printf("❌ Failed to create initial sell order: %v", err)
	}
}

// ensureMarketMakerWallet 마켓메이커 봇 지갑 확인/생성
func (mm *MarketMakerBot) ensureMarketMakerWallet() {
	var wallet models.UserWallet
	err := mm.db.Where("user_id = ?", mm.config.UserID).First(&wallet).Error

	if err == gorm.ErrRecordNotFound {
		// 마켓메이커 봇 지갑 생성
		wallet = models.UserWallet{
			UserID:                 mm.config.UserID,
			USDCBalance:            10000000, // 100,000 USDC (센트 단위)
			USDCLockedBalance:      0,
			BlueprintBalance:       0, // 봇은 BLUEPRINT 필요 없음
			BlueprintLockedBalance: 0,
			TotalUSDCDeposit:       10000000,
			TotalUSDCWithdraw:      0,
			TotalUSDCProfit:        0,
			TotalUSDCLoss:          0,
			TotalUSDCFees:          0,
			TotalBlueprintEarned:   0,
			TotalBlueprintSpent:    0,
			WinRate:                0,
			TotalTrades:            0,
		}

		if err := mm.db.Create(&wallet).Error; err != nil {
			log.Printf("❌ Failed to create market maker wallet: %v", err)
		} else {
			log.Printf("🤖 Created market maker wallet with $%.2f USDC",
				float64(wallet.USDCBalance)/100)
		}
	} else if err != nil {
		log.Printf("❌ Failed to check market maker wallet: %v", err)
	}
}
