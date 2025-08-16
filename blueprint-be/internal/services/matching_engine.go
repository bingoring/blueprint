package services

import (
	"blueprint-module/pkg/models"
	"blueprint-module/pkg/queue"
	"blueprint-module/pkg/redis"
	"container/heap"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"gorm.io/gorm"
)

// 🚀 High-Performance Matching Engine (Polymarket Style)

// MatchingEngine 고성능 매칭 엔진
type MatchingEngine struct {
	db                     *gorm.DB
	queuePublisher         *queue.Publisher
	sseService             *SSEService                 // SSE 실시간 브로드캐스트용
	fundingService         *FundingVerificationService // 🆕 펀딩 검증 서비스
	mentorQualificationSvc *MentorQualificationService // 🆕 멘토 자격 증명 서비스

	// 매칭 엔진 상태
	isRunning bool
	stopChan  chan struct{}
	orderChan chan *OrderMatchRequest
	mutex     sync.RWMutex

	// 시장별 주문장 (인메모리 고속 처리)
	orderBooks map[string]*OrderBookEngine // milestoneID:optionID -> OrderBook

	// 성능 통계
	stats MatchingStats
}

// OrderMatchRequest 매칭 요청
type OrderMatchRequest struct {
	Order    *models.Order
	Response chan<- *MatchingResult
}

// MatchingResult 매칭 결과
type MatchingResult struct {
	Trades   []models.Trade
	Error    error
	Executed bool
}

// OrderBookEngine 개별 시장의 주문장 엔진
type OrderBookEngine struct {
	MilestoneID uint
	OptionID    string

	// Price-Time Priority 힙
	BuyOrders  *BuyOrderHeap  // 높은 가격부터 (매수)
	SellOrders *SellOrderHeap // 낮은 가격부터 (매도)

	// 성능 최적화를 위한 인덱스
	orderIndex map[uint]*models.Order      // orderID -> order
	priceIndex map[float64][]*models.Order // price -> orders

	// 통계
	lastPrice   float64
	volume24h   int64
	tradesCount int64

	mutex sync.RWMutex
}

// BuyOrderHeap 매수 주문 힙 (가격 높은 순, 시간 빠른 순)
type BuyOrderHeap []*models.Order

func (h BuyOrderHeap) Len() int { return len(h) }

func (h BuyOrderHeap) Less(i, j int) bool {
	// 1. 가격이 높은 것이 우선
	if h[i].Price != h[j].Price {
		return h[i].Price > h[j].Price
	}
	// 2. 가격이 같으면 시간이 빠른 것이 우선 (FIFO)
	return h[i].CreatedAt.Before(h[j].CreatedAt)
}

func (h BuyOrderHeap) Swap(i, j int) { h[i], h[j] = h[j], h[i] }

func (h *BuyOrderHeap) Push(x interface{}) {
	*h = append(*h, x.(*models.Order))
}

func (h *BuyOrderHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

// SellOrderHeap 매도 주문 힙 (가격 낮은 순, 시간 빠른 순)
type SellOrderHeap []*models.Order

func (h SellOrderHeap) Len() int { return len(h) }

func (h SellOrderHeap) Less(i, j int) bool {
	// 1. 가격이 낮은 것이 우선
	if h[i].Price != h[j].Price {
		return h[i].Price < h[j].Price
	}
	// 2. 가격이 같으면 시간이 빠른 것이 우선 (FIFO)
	return h[i].CreatedAt.Before(h[j].CreatedAt)
}

func (h SellOrderHeap) Swap(i, j int) { h[i], h[j] = h[j], h[i] }

func (h *SellOrderHeap) Push(x interface{}) {
	*h = append(*h, x.(*models.Order))
}

func (h *SellOrderHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

// MatchingStats 매칭 엔진 통계
type MatchingStats struct {
	TotalMatches     int64     `json:"total_matches"`
	TotalVolume      int64     `json:"total_volume"`
	AvgMatchTime     float64   `json:"avg_match_time_ms"`
	OrdersProcessed  int64     `json:"orders_processed"`
	ActiveOrderBooks int       `json:"active_order_books"`
	CacheHitRate     float64   `json:"cache_hit_rate"`
	LastMatchTime    time.Time `json:"last_match_time"`
	StartTime        time.Time `json:"start_time"`
}

// NewMatchingEngine 매칭 엔진 생성자
func NewMatchingEngine(db *gorm.DB, sseService *SSEService, fundingService *FundingVerificationService, mentorQualificationSvc *MentorQualificationService) *MatchingEngine {
	return &MatchingEngine{
		db:                     db,
		queuePublisher:         queue.NewPublisher(),
		sseService:             sseService,
		fundingService:         fundingService,
		mentorQualificationSvc: mentorQualificationSvc,
		stopChan:               make(chan struct{}),
		orderChan:              make(chan *OrderMatchRequest, 10000), // 고성능 버퍼
		orderBooks:             make(map[string]*OrderBookEngine),
		stats: MatchingStats{
			StartTime: time.Now(),
		},
	}
}

// Start 매칭 엔진 시작
func (me *MatchingEngine) Start() error {
	me.mutex.Lock()
	defer me.mutex.Unlock()

	if me.isRunning {
		log.Println("⚠️ Matching engine is already running")
		return nil
	}

	log.Println("🚀 Starting Matching Engine...")

	// 기존 주문들을 메모리로 로드
	log.Println("📊 Loading existing orders...")
	if err := me.loadExistingOrders(); err != nil {
		log.Printf("❌ CRITICAL ERROR: Failed to load existing orders: %v", err)
		return err // 중요한 오류는 리턴
	}

	me.isRunning = true
	log.Println("🔥 High-Performance Matching Engine started!")

	// 매칭 워커 시작 (동시 처리)
	log.Println("🔧 Starting matching workers...")
	for i := 0; i < 4; i++ { // 4개 워커로 병렬 처리
		go me.matchingWorker(i)
	}

	// 통계 업데이트 워커
	go me.statsWorker()

	log.Println("✅ All matching engine workers started successfully")
	return nil
}

// Stop 매칭 엔진 중지
func (me *MatchingEngine) Stop() error {
	me.mutex.Lock()
	defer me.mutex.Unlock()

	if !me.isRunning {
		return nil
	}

	me.isRunning = false
	close(me.stopChan)
	close(me.orderChan)

	log.Println("🛑 Matching Engine stopped!")
	return nil
}

// SubmitOrder 주문 제출 (비동기 고속 처리)
func (me *MatchingEngine) SubmitOrder(order *models.Order) (*MatchingResult, error) {
	if !me.isRunning {
		return nil, fmt.Errorf("matching engine is not running")
	}

	responseChan := make(chan *MatchingResult, 1)

	request := &OrderMatchRequest{
		Order:    order,
		Response: responseChan,
	}

	// 논블로킹 전송
	select {
	case me.orderChan <- request:
		// 응답 대기 (타임아웃 30초로 증가)
		select {
		case result := <-responseChan:
			return result, nil
		case <-time.After(30 * time.Second):
			log.Printf("❌ Matching timeout for order: %+v", order)
			return nil, fmt.Errorf("matching timeout")
		}
	default:
		return nil, fmt.Errorf("matching queue is full")
	}
}

// matchingWorker 매칭 워커 (병렬 처리)
func (me *MatchingEngine) matchingWorker(workerID int) {
	log.Printf("🔧 Matching worker %d started", workerID)

	for {
		select {
		case <-me.stopChan:
			return
		case request := <-me.orderChan:
			if request == nil {
				return
			}

			startTime := time.Now()
			result := me.processOrder(request.Order)

			// 성능 통계 업데이트
			processingTime := time.Since(startTime)
			me.updateStats(processingTime)

			// 느린 주문만 로그 출력 (100ms 이상)
			if processingTime > 100*time.Millisecond {
				log.Printf("⚠️ Slow order processing: Worker %d, Order %d, Time %v", workerID, request.Order.ID, processingTime)
			}

			// 응답 전송 (논블로킹)
			select {
			case request.Response <- result:
				// 성공적으로 응답 전송
			default:
				// 응답 채널이 이미 닫혔거나 수신자가 없음 (타임아웃 발생)
				log.Printf("⚠️ Response channel unavailable for order %d (likely timeout)", request.Order.ID)
			}
		}
	}
}

// processOrder 주문 처리 (핵심 매칭 로직)
func (me *MatchingEngine) processOrder(order *models.Order) *MatchingResult {
	// 주문장 가져오기 또는 생성
	orderBook := me.getOrCreateOrderBook(order.MilestoneID, order.OptionID)

	orderBook.mutex.Lock()
	defer orderBook.mutex.Unlock()

	var trades []models.Trade

	// 폴리마켓 스타일: Limit Order만 처리
	trades = me.executeLimitOrder(orderBook, order)

	// 체결된 거래가 있으면 처리
	if len(trades) > 0 {
		// 🆕 펀딩 TVL 업데이트 (동기 처리 - 중요)
		go me.updateFundingTVL(order.MilestoneID, order.OptionID, trades)

		// 🆕 멘토 자격 업데이트 (비동기 처리 - "가장 똑똑한 돈" 식별)
		go me.updateMentorQualification(order.MilestoneID, trades)

		// 🆕 멘토 풀 수수료 적립 (비동기 처리 - "The Reward Engine")
		go me.accumulateMentorPoolFees(order.MilestoneID, trades)

		// 데이터베이스에 저장 (비동기)
		go me.persistTrades(trades)

		// 사용자 지갑 잔액 업데이트 (비동기)
		go me.updateUserWallets(trades)

		// 사용자 Position 업데이트 (비동기)
		go me.updateUserPositions(trades)

		// MarketData 업데이트 (비동기)
		go me.updateMarketData(order.MilestoneID, order.OptionID, trades)

		// 실시간 브로드캐스트
		go me.broadcastTrades(trades)

		// 캐시 업데이트
		go me.updateMarketCache(order.MilestoneID, order.OptionID, trades)
	}

	return &MatchingResult{
		Trades:   trades,
		Executed: len(trades) > 0,
		Error:    nil,
	}
}

// executeLimitOrder 지정가 주문 체결
func (me *MatchingEngine) executeLimitOrder(orderBook *OrderBookEngine, order *models.Order) []models.Trade {
	var trades []models.Trade
	remaining := order.Quantity

	if order.Side == models.OrderSideBuy {
		// 매수 지정가: 지정가 이하의 매도 주문과 체결
		for remaining > 0 && orderBook.SellOrders.Len() > 0 {
			bestSell := (*orderBook.SellOrders)[0]

			if bestSell.Price > order.Price {
				break // 가격 조건 불만족
			}

			if bestSell.Remaining <= 0 {
				heap.Pop(orderBook.SellOrders)
				continue
			}

			matchQuantity := min(remaining, bestSell.Remaining)

			totalAmount := int64(float64(matchQuantity) * bestSell.Price * 100) // 센트 단위로 변환
			buyerFee := totalAmount * 25 / 10000                                // 0.25% 수수료
			sellerFee := totalAmount * 25 / 10000                               // 0.25% 수수료

			trade := models.Trade{
				ProjectID:   order.ProjectID,
				MilestoneID: order.MilestoneID,
				OptionID:    order.OptionID,
				BuyOrderID:  order.ID,
				SellOrderID: bestSell.ID,
				BuyerID:     order.UserID,
				SellerID:    bestSell.UserID,
				Quantity:    matchQuantity,
				Price:       bestSell.Price,
				TotalAmount: totalAmount,
				BuyerFee:    buyerFee,
				SellerFee:   sellerFee,
				CreatedAt:   time.Now(),
			}

			trades = append(trades, trade)

			remaining -= matchQuantity
			bestSell.Remaining -= matchQuantity
			bestSell.Filled += matchQuantity

			if bestSell.Remaining <= 0 {
				heap.Pop(orderBook.SellOrders)
				bestSell.Status = models.OrderStatusFilled
				// 🔧 메모리 리크 방지: 완료된 주문은 인덱스에서 제거
				delete(orderBook.orderIndex, bestSell.ID)
			}

			orderBook.lastPrice = bestSell.Price
		}

		// 미체결 물량이 있으면 주문장에 추가
		if remaining > 0 {
			order.Remaining = remaining
			order.Status = models.OrderStatusPending
			heap.Push(orderBook.BuyOrders, order)
			orderBook.orderIndex[order.ID] = order
		}
	} else {
		// 매도 지정가: 지정가 이상의 매수 주문과 체결
		for remaining > 0 && orderBook.BuyOrders.Len() > 0 {
			bestBuy := (*orderBook.BuyOrders)[0]

			if bestBuy.Price < order.Price {
				break // 가격 조건 불만족
			}

			if bestBuy.Remaining <= 0 {
				heap.Pop(orderBook.BuyOrders)
				continue
			}

			matchQuantity := min(remaining, bestBuy.Remaining)

			totalAmount := int64(float64(matchQuantity) * bestBuy.Price * 100) // 센트 단위로 변환
			buyerFee := totalAmount * 25 / 10000                               // 0.25% 수수료
			sellerFee := totalAmount * 25 / 10000                              // 0.25% 수수료

			trade := models.Trade{
				ProjectID:   order.ProjectID,
				MilestoneID: order.MilestoneID,
				OptionID:    order.OptionID,
				BuyOrderID:  bestBuy.ID,
				SellOrderID: order.ID,
				BuyerID:     bestBuy.UserID,
				SellerID:    order.UserID,
				Quantity:    matchQuantity,
				Price:       bestBuy.Price,
				TotalAmount: totalAmount,
				BuyerFee:    buyerFee,
				SellerFee:   sellerFee,
				CreatedAt:   time.Now(),
			}

			trades = append(trades, trade)

			remaining -= matchQuantity
			bestBuy.Remaining -= matchQuantity
			bestBuy.Filled += matchQuantity

			if bestBuy.Remaining <= 0 {
				heap.Pop(orderBook.BuyOrders)
				bestBuy.Status = models.OrderStatusFilled
				// 🔧 메모리 리크 방지: 완료된 주문은 인덱스에서 제거
				delete(orderBook.orderIndex, bestBuy.ID)
			}

			orderBook.lastPrice = bestBuy.Price
		}

		// 미체결 물량이 있으면 주문장에 추가
		if remaining > 0 {
			order.Remaining = remaining
			order.Status = models.OrderStatusPending
			heap.Push(orderBook.SellOrders, order)
			orderBook.orderIndex[order.ID] = order
		}
	}

	// 주문 상태 업데이트
	order.Filled = order.Quantity - remaining

	if remaining <= 0 {
		order.Status = models.OrderStatusFilled
		// 🔧 메모리 리크 방지: 완전 체결된 주문도 인덱스에서 제거
		orderBook.mutex.Lock()
		delete(orderBook.orderIndex, order.ID)
		orderBook.mutex.Unlock()
	} else if order.Filled > 0 {
		order.Status = models.OrderStatusPartial
	}

	return trades
}

// CancelOrder 주문 취소 (매칭 엔진에서 제거)
func (me *MatchingEngine) CancelOrder(order *models.Order) {
	key := me.getMarketKey(order.MilestoneID, order.OptionID)

	me.mutex.RLock()
	orderBook, exists := me.orderBooks[key]
	me.mutex.RUnlock()

	if !exists {
		return // 주문장이 없으면 무시
	}

	orderBook.mutex.Lock()
	defer orderBook.mutex.Unlock()

	// 인덱스에서 주문 제거
	delete(orderBook.orderIndex, order.ID)

	// 힙에서도 제거 (비효율적이지만 정확성 보장)
	me.removeFromHeap(orderBook, order)
}

// removeFromHeap 힙에서 특정 주문 제거
func (me *MatchingEngine) removeFromHeap(orderBook *OrderBookEngine, order *models.Order) {
	if order.Side == models.OrderSideBuy {
		for i, o := range *orderBook.BuyOrders {
			if o.ID == order.ID {
				(*orderBook.BuyOrders)[i] = (*orderBook.BuyOrders)[len(*orderBook.BuyOrders)-1]
				*orderBook.BuyOrders = (*orderBook.BuyOrders)[:len(*orderBook.BuyOrders)-1]
				heap.Init(orderBook.BuyOrders)
				break
			}
		}
	} else {
		for i, o := range *orderBook.SellOrders {
			if o.ID == order.ID {
				(*orderBook.SellOrders)[i] = (*orderBook.SellOrders)[len(*orderBook.SellOrders)-1]
				*orderBook.SellOrders = (*orderBook.SellOrders)[:len(*orderBook.SellOrders)-1]
				heap.Init(orderBook.SellOrders)
				break
			}
		}
	}
}

// 🆕 updateFundingTVL 펀딩 TVL 업데이트
func (me *MatchingEngine) updateFundingTVL(milestoneID uint, optionID string, trades []models.Trade) {
	if me.fundingService == nil {
		return
	}

	// 거래의 총 금액 계산
	var totalAmount int64
	for _, trade := range trades {
		totalAmount += trade.TotalAmount
	}

	// 펀딩 서비스를 통해 TVL 업데이트
	if err := me.fundingService.UpdateTVL(milestoneID, optionID, totalAmount); err != nil {
		log.Printf("❌ Failed to update TVL for milestone %d: %v", milestoneID, err)
	}
}

// 🆕 updateMentorQualification 멘토 자격 업데이트
func (me *MatchingEngine) updateMentorQualification(milestoneID uint, trades []models.Trade) {
	if me.mentorQualificationSvc == nil {
		return
	}

	// 성공 베팅과 관련된 거래만 처리 (optionID가 "success"인 경우)
	hasSuccessBetting := false
	for _, trade := range trades {
		if trade.OptionID == "success" {
			hasSuccessBetting = true
			break
		}
	}

	if !hasSuccessBetting {
		return // 실패 베팅은 멘토 자격과 관련 없음
	}

	// 멘토 자격 재처리 (베팅 순위 변동 반영)
	if _, err := me.mentorQualificationSvc.ProcessMilestoneBetting(milestoneID); err != nil {
		log.Printf("❌ Failed to update mentor qualification for milestone %d: %v", milestoneID, err)
	} else {
		log.Printf("✨ Mentor qualification updated for milestone %d after new trades", milestoneID)
	}
}

// 🆕 accumulateMentorPoolFees 멘토 풀에 수수료 적립
func (me *MatchingEngine) accumulateMentorPoolFees(milestoneID uint, trades []models.Trade) {
	// 총 거래 수수료 계산
	var totalFees int64
	for _, trade := range trades {
		totalFees += trade.BuyerFee + trade.SellerFee
	}

	if totalFees <= 0 {
		return
	}

	// 멘토 풀 조회 및 수수료 적립
	var mentorPool models.MentorPool
	if err := me.db.Where("milestone_id = ?", milestoneID).First(&mentorPool).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Printf("📋 No mentor pool found for milestone %d, skipping fee accumulation", milestoneID)
			return
		}
		log.Printf("❌ Failed to query mentor pool for milestone %d: %v", milestoneID, err)
		return
	}

	// 설정된 비율만큼 멘토 풀에 적립 (기본 50%)
	mentorPoolFees := int64(float64(totalFees) * mentorPool.FeePercentage / 100)

	// 멘토 풀 업데이트
	mentorPool.AccumulatedFees += mentorPoolFees
	mentorPool.TotalPoolAmount += mentorPoolFees

	if err := me.db.Save(&mentorPool).Error; err != nil {
		log.Printf("❌ Failed to update mentor pool fees for milestone %d: %v", milestoneID, err)
		return
	}

	log.Printf("💰 Accumulated $%.2f mentor pool fees for milestone %d (%.1f%% of total fees $%.2f)",
		float64(mentorPoolFees)/100, milestoneID, mentorPool.FeePercentage, float64(totalFees)/100)

	// 실시간 멘토 풀 업데이트 알림
	go me.broadcastMentorPoolUpdate(milestoneID, &mentorPool, mentorPoolFees)
}

// broadcastMentorPoolUpdate 멘토 풀 업데이트 브로드캐스트
func (me *MatchingEngine) broadcastMentorPoolUpdate(milestoneID uint, pool *models.MentorPool, addedAmount int64) {
	if me.sseService == nil {
		return
	}

	event := MarketUpdateEvent{
		MilestoneID: milestoneID,
		MarketData: map[string]interface{}{
			"event_type": "mentor_pool_update",
			"data": map[string]interface{}{
				"milestone_id":      milestoneID,
				"total_pool_amount": pool.TotalPoolAmount,
				"accumulated_fees":  pool.AccumulatedFees,
				"added_amount":      addedAmount,
				"fee_percentage":    pool.FeePercentage,
				"updated_at":        time.Now().Unix(),
			},
		},
		Timestamp: time.Now().Unix(),
	}

	me.sseService.BroadcastMarketUpdate(event)
}

// Helper functions

func (me *MatchingEngine) getMarketKey(milestoneID uint, optionID string) string {
	return fmt.Sprintf("%d:%s", milestoneID, optionID)
}

func (me *MatchingEngine) getOrCreateOrderBook(milestoneID uint, optionID string) *OrderBookEngine {
	me.mutex.Lock()
	defer me.mutex.Unlock()
	return me.getOrCreateOrderBookUnsafe(milestoneID, optionID)
}

// getOrCreateOrderBookUnsafe - mutex 없이 오더북 생성 (내부 호출용)
func (me *MatchingEngine) getOrCreateOrderBookUnsafe(milestoneID uint, optionID string) *OrderBookEngine {
	key := me.getMarketKey(milestoneID, optionID)

	if orderBook, exists := me.orderBooks[key]; exists {
		return orderBook
	}

	orderBook := &OrderBookEngine{
		MilestoneID: milestoneID,
		OptionID:    optionID,
		BuyOrders:   &BuyOrderHeap{},
		SellOrders:  &SellOrderHeap{},
		orderIndex:  make(map[uint]*models.Order),
		priceIndex:  make(map[float64][]*models.Order),
	}

	heap.Init(orderBook.BuyOrders)
	heap.Init(orderBook.SellOrders)

	me.orderBooks[key] = orderBook
	return orderBook
}

func (me *MatchingEngine) loadExistingOrders() error {
	var orders []models.Order
	err := me.db.Where("status IN ?", []models.OrderStatus{
		models.OrderStatusPending,
		models.OrderStatusPartial,
	}).Find(&orders).Error

	if err != nil {
		// 테이블이 존재하지 않는 경우 (깨끗한 데이터베이스) - 정상적인 상황
		if me.isTableNotExistsError(err) {
			log.Printf("📋 No orders table found - starting with clean state")
			return nil
		}
		// 다른 오류는 여전히 critical error로 처리
		return err
	}

	for _, order := range orders {
		// mutex가 이미 Start()에서 잠겨있으므로 Unsafe 버전 사용
		orderBook := me.getOrCreateOrderBookUnsafe(order.MilestoneID, order.OptionID)
		orderBook.mutex.Lock()

		if order.Side == models.OrderSideBuy {
			heap.Push(orderBook.BuyOrders, &order)
		} else {
			heap.Push(orderBook.SellOrders, &order)
		}

		orderBook.orderIndex[order.ID] = &order
		orderBook.mutex.Unlock()
	}

	log.Printf("📊 Loaded %d existing orders into matching engine", len(orders))
	return nil
}

// isTableNotExistsError 테이블이 존재하지 않는 오류인지 확인
func (me *MatchingEngine) isTableNotExistsError(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()
	// PostgreSQL: relation "orders" does not exist
	// MySQL: Table 'db.orders' doesn't exist
	// SQLite: no such table: orders
	return (errStr != "" &&
		(errStr == `ERROR: relation "orders" does not exist (SQLSTATE 42P01)` ||
			strings.Contains(errStr, `relation "orders" does not exist`) ||
			(strings.Contains(errStr, `Table`) && strings.Contains(errStr, `orders`) && strings.Contains(errStr, `doesn't exist`)) ||
			strings.Contains(errStr, `no such table: orders`)))
}

func (me *MatchingEngine) persistTrades(trades []models.Trade) {
	for _, trade := range trades {
		if err := me.db.Create(&trade).Error; err != nil {
			log.Printf("❌ Failed to persist trade: %v", err)
		}
	}
}

func (me *MatchingEngine) broadcastTrades(trades []models.Trade) {
	for _, trade := range trades {
		// Redis 브로드캐스트 (기존)
		redis.BroadcastTradeUpdate(trade.MilestoneID, trade.OptionID, trade)
		redis.BroadcastPriceChange(trade.MilestoneID, trade.OptionID, trade.Price)

		// SSE 실시간 브로드캐스트 (신규 추가)
		if me.sseService != nil {
			// 거래 이벤트 브로드캐스트
			me.sseService.BroadcastTradeUpdate(trade.MilestoneID, trade.OptionID, map[string]interface{}{
				"trade_id":     trade.ID,
				"option_id":    trade.OptionID,
				"buyer_id":     trade.BuyerID,
				"seller_id":    trade.SellerID,
				"quantity":     trade.Quantity,
				"price":        trade.Price,
				"total_amount": trade.TotalAmount,
				"timestamp":    trade.CreatedAt.Unix(),
			})

			// 가격 변동 브로드캐스트
			me.sseService.BroadcastPriceChange(trade.MilestoneID, trade.OptionID, 0, trade.Price)

			// Order Book 업데이트 브로드캐스트
			orderBook := me.getOrCreateOrderBook(trade.MilestoneID, trade.OptionID)
			me.broadcastOrderBookUpdate(orderBook, trade.MilestoneID, trade.OptionID)
		}

		// 큐에 작업 추가
		me.queuePublisher.EnqueueTradeWork(trade.MilestoneID, trade.OptionID, queue.TradeEventData{
			TradeID:     trade.ID,
			BuyerID:     trade.BuyerID,
			SellerID:    trade.SellerID,
			Quantity:    trade.Quantity,
			Price:       trade.Price,
			TotalAmount: trade.TotalAmount,
		})
	}
}

func (me *MatchingEngine) updateMarketCache(milestoneID uint, optionID string, trades []models.Trade) {
	// Redis 캐시 업데이트
	if len(trades) > 0 {
		lastTrade := trades[len(trades)-1]
		redis.SetMarketPrice(milestoneID, optionID, lastTrade.Price)
		redis.SetRecentTrades(milestoneID, optionID, trades)
	}
}

// broadcastOrderBookUpdate Order Book 변경사항을 SSE로 브로드캐스트
func (me *MatchingEngine) broadcastOrderBookUpdate(orderBook *OrderBookEngine, milestoneID uint, optionID string) {
	if me.sseService == nil {
		return
	}

	orderBook.mutex.RLock()
	defer orderBook.mutex.RUnlock()

	// 상위 5개 매수/매도 주문 추출
	buyOrders := make([]map[string]interface{}, 0, 5)
	sellOrders := make([]map[string]interface{}, 0, 5)

	// 매수 주문 (높은 가격순)
	buyCount := 0
	for i := 0; i < orderBook.BuyOrders.Len() && buyCount < 5; i++ {
		order := (*orderBook.BuyOrders)[i]
		if order.Remaining > 0 {
			buyOrders = append(buyOrders, map[string]interface{}{
				"price":    order.Price,
				"quantity": order.Remaining,
			})
			buyCount++
		}
	}

	// 매도 주문 (낮은 가격순)
	sellCount := 0
	for i := 0; i < orderBook.SellOrders.Len() && sellCount < 5; i++ {
		order := (*orderBook.SellOrders)[i]
		if order.Remaining > 0 {
			sellOrders = append(sellOrders, map[string]interface{}{
				"price":    order.Price,
				"quantity": order.Remaining,
			})
			sellCount++
		}
	}

	orderBookData := map[string]interface{}{
		"milestone_id": milestoneID,
		"option_id":    optionID,
		"buy_orders":   buyOrders,
		"sell_orders":  sellOrders,
	}

	me.sseService.BroadcastOrderBookUpdate(milestoneID, optionID, orderBookData)
}

// updateMarketData MarketData 테이블 업데이트
func (me *MatchingEngine) updateMarketData(milestoneID uint, optionID string, trades []models.Trade) {
	if len(trades) == 0 {
		return
	}

	// 최신 거래 정보
	lastTrade := trades[len(trades)-1]
	newPrice := lastTrade.Price
	tradeTime := lastTrade.CreatedAt

	// 기존 MarketData 조회
	var marketData models.MarketData
	err := me.db.Where("milestone_id = ? AND option_id = ?", milestoneID, optionID).First(&marketData).Error

	if err != nil {
		// MarketData가 없으면 새로 생성
		marketData = models.MarketData{
			MilestoneID:   milestoneID,
			OptionID:      optionID,
			CurrentPrice:  newPrice,
			PreviousPrice: newPrice,
			HighPrice24h:  newPrice,
			LowPrice24h:   newPrice,
			LastTradeTime: tradeTime,
		}
	} else {
		// 기존 데이터 업데이트
		marketData.PreviousPrice = marketData.CurrentPrice
		marketData.CurrentPrice = newPrice
		marketData.LastTradeTime = tradeTime

		// 24시간 고가/저가 업데이트
		if newPrice > marketData.HighPrice24h {
			marketData.HighPrice24h = newPrice
		}
		if newPrice < marketData.LowPrice24h || marketData.LowPrice24h == 0 {
			marketData.LowPrice24h = newPrice
		}

		// 24시간 변동폭 계산 (24시간 전 가격과 비교)
		var price24hAgo float64
		me.db.Model(&models.Trade{}).
			Where("milestone_id = ? AND option_id = ? AND created_at <= ?",
				milestoneID, optionID, tradeTime.Add(-24*time.Hour)).
			Order("created_at DESC").
			Limit(1).
			Pluck("price", &price24hAgo)

		if price24hAgo > 0 {
			marketData.Change24h = newPrice - price24hAgo
			marketData.ChangePercent = (marketData.Change24h / price24hAgo) * 100
		} else {
			// 24시간 전 데이터가 없으면 현재 가격 기준
			marketData.Change24h = 0
			marketData.ChangePercent = 0
		}
	}

	// 24시간 거래량 및 거래 수 계산
	var volume24h int64
	var trades24h int

	me.db.Model(&models.Trade{}).
		Where("milestone_id = ? AND option_id = ? AND created_at > ?",
			milestoneID, optionID, tradeTime.Add(-24*time.Hour)).
		Select("COALESCE(SUM(quantity), 0) as volume, COUNT(*) as trades").
		Row().Scan(&volume24h, &trades24h)

	marketData.Volume24h = volume24h
	marketData.Trades24h = trades24h

	// 현재 호가창에서 BidPrice, AskPrice, Spread 계산
	orderBook := me.getOrCreateOrderBook(milestoneID, optionID)
	orderBook.mutex.RLock()

	if orderBook.BuyOrders.Len() > 0 {
		marketData.BidPrice = (*orderBook.BuyOrders)[0].Price
	}
	if orderBook.SellOrders.Len() > 0 {
		marketData.AskPrice = (*orderBook.SellOrders)[0].Price
	}
	if marketData.BidPrice > 0 && marketData.AskPrice > 0 {
		marketData.Spread = marketData.AskPrice - marketData.BidPrice
	}

	orderBook.mutex.RUnlock()
	marketData.UpdatedAt = time.Now()

	// 데이터베이스에 저장
	if marketData.ID == 0 {
		err = me.db.Create(&marketData).Error
	} else {
		err = me.db.Save(&marketData).Error
	}

	if err != nil {
		log.Printf("❌ Failed to update market data for %d:%s: %v", milestoneID, optionID, err)
	} else {
		log.Printf("📊 Updated market data for %d:%s: price %.4f, volume %d",
			milestoneID, optionID, newPrice, volume24h)
	}
}

// updateUserPositions 사용자 포지션 업데이트
func (me *MatchingEngine) updateUserPositions(trades []models.Trade) {
	for _, trade := range trades {
		// 매수자 포지션 업데이트 (+수량)
		me.updateSinglePosition(trade.BuyerID, trade.ProjectID, trade.MilestoneID,
			trade.OptionID, trade.Quantity, trade.Price, trade.TotalAmount, true)

		// 매도자 포지션 업데이트 (-수량)
		me.updateSinglePosition(trade.SellerID, trade.ProjectID, trade.MilestoneID,
			trade.OptionID, -trade.Quantity, trade.Price, trade.TotalAmount, false)
	}
}

// updateSinglePosition 개별 사용자 포지션 업데이트
func (me *MatchingEngine) updateSinglePosition(userID, projectID, milestoneID uint,
	optionID string, quantity int64, price float64, totalAmount int64, isBuy bool) {

	// 기존 포지션 조회
	var position models.Position
	err := me.db.Where("user_id = ? AND project_id = ? AND milestone_id = ? AND option_id = ?",
		userID, projectID, milestoneID, optionID).First(&position).Error

	if err != nil {
		// 새로운 포지션 생성
		if isBuy {
			position = models.Position{
				UserID:      userID,
				ProjectID:   projectID,
				MilestoneID: milestoneID,
				OptionID:    optionID,
				Quantity:    quantity,
				AvgPrice:    price,
				TotalCost:   totalAmount,
				Realized:    0,
				Unrealized:  0,
				UpdatedAt:   time.Now(),
			}
		} else {
			// 매도인데 기존 포지션이 없으면 숏포지션 생성
			position = models.Position{
				UserID:      userID,
				ProjectID:   projectID,
				MilestoneID: milestoneID,
				OptionID:    optionID,
				Quantity:    quantity, // 음수
				AvgPrice:    price,
				TotalCost:   -totalAmount, // 매도로 인한 수익
				Realized:    0,
				Unrealized:  0,
				UpdatedAt:   time.Now(),
			}
		}

		err = me.db.Create(&position).Error
		if err != nil {
			log.Printf("❌ Failed to create position for user %d: %v", userID, err)
		} else {
			log.Printf("🆕 Created new position for user %d: %s %d@%.4f",
				userID, optionID, quantity, price)
		}
	} else {
		// 기존 포지션 업데이트
		oldQuantity := position.Quantity
		newQuantity := oldQuantity + quantity

		if isBuy {
			// 매수: 평균단가 재계산
			if newQuantity > 0 {
				// 순매수 포지션
				totalValue := float64(position.TotalCost) + float64(totalAmount)
				position.AvgPrice = totalValue / float64(newQuantity)
				position.TotalCost += totalAmount
			} else if newQuantity == 0 {
				// 포지션 완전 청산
				position.Realized += totalAmount - int64(float64(quantity)*position.AvgPrice)
				position.AvgPrice = 0
				position.TotalCost = 0
			} else {
				// 일부 청산 (숏포지션으로 전환)
				realizedPnL := int64(float64(oldQuantity) * (price - position.AvgPrice))
				position.Realized += realizedPnL
				position.AvgPrice = price
				position.TotalCost = int64(float64(newQuantity) * price)
			}
		} else {
			// 매도: 실현손익 계산
			if oldQuantity > 0 {
				// 기존 매수 포지션에서 매도
				sellQuantity := -quantity
				realizedPnL := int64(float64(sellQuantity) * (price - position.AvgPrice))
				position.Realized += realizedPnL

				if newQuantity > 0 {
					// 일부 매도
					position.TotalCost = int64(float64(newQuantity) * position.AvgPrice)
				} else if newQuantity == 0 {
					// 전량 매도
					position.AvgPrice = 0
					position.TotalCost = 0
				} else {
					// 과매도 (숏포지션)
					position.AvgPrice = price
					position.TotalCost = int64(float64(newQuantity) * price)
				}
			} else {
				// 기존 숏포지션에서 추가 매도 또는 신규 숏매도
				if oldQuantity == 0 {
					// 신규 숏매도
					position.AvgPrice = price
					position.TotalCost = int64(float64(newQuantity) * price)
				} else {
					// 기존 숏포지션에 추가
					totalValue := float64(position.TotalCost) + float64(totalAmount)
					position.AvgPrice = totalValue / float64(newQuantity)
					position.TotalCost += totalAmount
				}
			}
		}

		position.Quantity = newQuantity
		position.UpdatedAt = time.Now()

		// 미실현 손익 계산 (현재 시장가 기준)
		if newQuantity != 0 {
			currentPrice := me.getCurrentMarketPrice(milestoneID, optionID)
			if currentPrice > 0 {
				position.Unrealized = int64(float64(newQuantity) * (currentPrice - position.AvgPrice))
			}
		} else {
			position.Unrealized = 0
		}

		err = me.db.Save(&position).Error
		if err != nil {
			log.Printf("❌ Failed to update position for user %d: %v", userID, err)
		} else {
			log.Printf("🔄 Updated position for user %d: %s %d@%.4f (realized: %d)",
				userID, optionID, newQuantity, position.AvgPrice, position.Realized)
		}
	}
}

// getCurrentMarketPrice 현재 시장가 조회
func (me *MatchingEngine) getCurrentMarketPrice(milestoneID uint, optionID string) float64 {
	orderBook := me.getOrCreateOrderBook(milestoneID, optionID)
	orderBook.mutex.RLock()
	defer orderBook.mutex.RUnlock()

	// 마지막 체결가가 있으면 사용
	if orderBook.lastPrice > 0 {
		return orderBook.lastPrice
	}

	// 호가창 중간값 사용
	if orderBook.BuyOrders.Len() > 0 && orderBook.SellOrders.Len() > 0 {
		bidPrice := (*orderBook.BuyOrders)[0].Price
		askPrice := (*orderBook.SellOrders)[0].Price
		return (bidPrice + askPrice) / 2
	}

	// 기본값 (초기 확률)
	return 0.33 // 33¢
}

// updateUserWallets 사용자 지갑 잔액 업데이트
func (me *MatchingEngine) updateUserWallets(trades []models.Trade) {
	for _, trade := range trades {
		// 매수자 지갑 업데이트: USDC 차감, LockedBalance 감소
		me.updateBuyerWallet(trade.BuyerID, trade.TotalAmount, trade.BuyerFee)

		// 매도자 지갑 업데이트: USDC 증가, LockedBalance 감소
		me.updateSellerWallet(trade.SellerID, trade.TotalAmount, trade.SellerFee)
	}
}

// updateBuyerWallet 매수자 지갑 업데이트
func (me *MatchingEngine) updateBuyerWallet(buyerID uint, totalAmount, fee int64) {
	var wallet models.UserWallet
	err := me.db.Where("user_id = ?", buyerID).First(&wallet).Error

	if err != nil {
		log.Printf("❌ Failed to find buyer wallet for user %d: %v", buyerID, err)
		return
	}

	// 잠긴 잔액에서 거래금액 차감, 수수료는 일반 잔액에서 차감
	if wallet.USDCLockedBalance >= totalAmount {
		wallet.USDCLockedBalance -= totalAmount
		wallet.USDCBalance -= fee // 수수료는 일반 잔액에서 차감
	} else {
		log.Printf("⚠️ Insufficient locked balance for buyer %d: locked=%d, needed=%d",
			buyerID, wallet.USDCLockedBalance, totalAmount)
		// 부족하면 일반 잔액에서 모두 차감
		remaining := totalAmount - wallet.USDCLockedBalance
		wallet.USDCLockedBalance = 0
		wallet.USDCBalance -= (remaining + fee)
	}

	// 통계 업데이트
	wallet.TotalUSDCFees += fee
	wallet.TotalTrades++
	wallet.UpdatedAt = time.Now()

	err = me.db.Save(&wallet).Error
	if err != nil {
		log.Printf("❌ Failed to update buyer wallet for user %d: %v", buyerID, err)
	} else {
		log.Printf("💰 Updated buyer wallet for user %d: paid %d USDC (fee: %d)",
			buyerID, totalAmount, fee)
	}
}

// updateSellerWallet 매도자 지갑 업데이트
func (me *MatchingEngine) updateSellerWallet(sellerID uint, totalAmount, fee int64) {
	var wallet models.UserWallet
	err := me.db.Where("user_id = ?", sellerID).First(&wallet).Error

	if err != nil {
		log.Printf("❌ Failed to find seller wallet for user %d: %v", sellerID, err)
		return
	}

	// 매도 수익 추가 (수수료 제외)
	netProceeds := totalAmount - fee
	wallet.USDCBalance += netProceeds

	// 통계 업데이트
	wallet.TotalUSDCProfit += netProceeds
	wallet.TotalUSDCFees += fee
	wallet.TotalTrades++
	wallet.UpdatedAt = time.Now()

	err = me.db.Save(&wallet).Error
	if err != nil {
		log.Printf("❌ Failed to update seller wallet for user %d: %v", sellerID, err)
	} else {
		log.Printf("💰 Updated seller wallet for user %d: received %d USDC (fee: %d)",
			sellerID, netProceeds, fee)
	}
}

func (me *MatchingEngine) updateStats(processingTime time.Duration) {
	me.mutex.Lock()
	defer me.mutex.Unlock()

	me.stats.OrdersProcessed++
	me.stats.TotalMatches++
	me.stats.LastMatchTime = time.Now()

	// 이동 평균으로 평균 매칭 시간 계산
	me.stats.AvgMatchTime = (me.stats.AvgMatchTime * 0.95) + (processingTime.Seconds() * 1000 * 0.05)
}

func (me *MatchingEngine) statsWorker() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-me.stopChan:
			return
		case <-ticker.C:
			me.printStats()
		}
	}
}

func (me *MatchingEngine) printStats() {
	me.mutex.RLock()
	defer me.mutex.RUnlock()

	log.Printf("🔥 Matching Engine Stats:")
	log.Printf("   Orders Processed: %d", me.stats.OrdersProcessed)
	log.Printf("   Total Matches: %d", me.stats.TotalMatches)
	log.Printf("   Avg Match Time: %.2fms", me.stats.AvgMatchTime)
	log.Printf("   Active Order Books: %d", len(me.orderBooks))
	log.Printf("   Uptime: %v", time.Since(me.stats.StartTime))
}

// GetStats 통계 조회
func (me *MatchingEngine) GetStats() MatchingStats {
	me.mutex.RLock()
	defer me.mutex.RUnlock()
	return me.stats
}

// GetOrderBook 주문장 조회
func (me *MatchingEngine) GetOrderBook(milestoneID uint, optionID string) *models.OrderBook {
	key := me.getMarketKey(milestoneID, optionID)

	me.mutex.RLock()
	orderBookEngine, exists := me.orderBooks[key]
	me.mutex.RUnlock()

	if !exists {
		return &models.OrderBook{
			MilestoneID: milestoneID,
			OptionID:    optionID,
			Bids:        []models.OrderBookLevel{},
			Asks:        []models.OrderBookLevel{},
			LastUpdate:  time.Now(),
		}
	}

	orderBookEngine.mutex.RLock()
	defer orderBookEngine.mutex.RUnlock()

	// 매수 호가 생성
	bids := make([]models.OrderBookLevel, 0)
	bidPrices := make(map[float64]int64)

	for _, order := range *orderBookEngine.BuyOrders {
		if order.Remaining > 0 {
			bidPrices[order.Price] += order.Remaining
		}
	}

	for price, quantity := range bidPrices {
		bids = append(bids, models.OrderBookLevel{
			Price:    price,
			Quantity: quantity,
			Count:    1,
		})
	}

	// 매도 호가 생성
	asks := make([]models.OrderBookLevel, 0)
	askPrices := make(map[float64]int64)

	for _, order := range *orderBookEngine.SellOrders {
		if order.Remaining > 0 {
			askPrices[order.Price] += order.Remaining
		}
	}

	for price, quantity := range askPrices {
		asks = append(asks, models.OrderBookLevel{
			Price:    price,
			Quantity: quantity,
			Count:    1,
		})
	}

	return &models.OrderBook{
		MilestoneID: milestoneID,
		OptionID:    optionID,
		Bids:        bids,
		Asks:        asks,
		LastUpdate:  time.Now(),
	}
}

func min(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}
