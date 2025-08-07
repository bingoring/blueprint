package services

import (
	"blueprint/internal/models"
	"blueprint/internal/queue"
	"blueprint/internal/redis"
	"container/heap"
	"fmt"
	"log"
	"sync"
	"time"

	"gorm.io/gorm"
)

// 🚀 High-Performance Matching Engine (Polymarket Style)

// MatchingEngine 고성능 매칭 엔진
type MatchingEngine struct {
	db             *gorm.DB
	queuePublisher *queue.Publisher

	// 매칭 엔진 상태
	isRunning      bool
	stopChan       chan struct{}
	orderChan      chan *OrderMatchRequest
	mutex          sync.RWMutex

	// 시장별 주문장 (인메모리 고속 처리)
	orderBooks     map[string]*OrderBookEngine // milestoneID:optionID -> OrderBook

	// 성능 통계
	stats          MatchingStats
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
	orderIndex map[uint]*models.Order // orderID -> order
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
	TotalMatches       int64     `json:"total_matches"`
	TotalVolume        int64     `json:"total_volume"`
	AvgMatchTime       float64   `json:"avg_match_time_ms"`
	OrdersProcessed    int64     `json:"orders_processed"`
	ActiveOrderBooks   int       `json:"active_order_books"`
	CacheHitRate       float64   `json:"cache_hit_rate"`
	LastMatchTime      time.Time `json:"last_match_time"`
	StartTime          time.Time `json:"start_time"`
}

// NewMatchingEngine 매칭 엔진 생성자
func NewMatchingEngine(db *gorm.DB) *MatchingEngine {
	return &MatchingEngine{
		db:             db,
		queuePublisher: queue.NewPublisher(),
		stopChan:       make(chan struct{}),
		orderChan:      make(chan *OrderMatchRequest, 10000), // 고성능 버퍼
		orderBooks:     make(map[string]*OrderBookEngine),
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
		// 데이터베이스에 저장 (비동기)
		go me.persistTrades(trades)

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

			trade := models.Trade{
				ProjectID:    order.ProjectID,
				MilestoneID:  order.MilestoneID,
				OptionID:     order.OptionID,
				BuyOrderID:   order.ID,
				SellOrderID:  bestSell.ID,
				BuyerID:      order.UserID,
				SellerID:     bestSell.UserID,
				Quantity:     matchQuantity,
				Price:        bestSell.Price,
				TotalAmount:  int64(float64(matchQuantity) * bestSell.Price),
				CreatedAt:    time.Now(),
			}

			trades = append(trades, trade)

			remaining -= matchQuantity
			bestSell.Remaining -= matchQuantity
			bestSell.Filled += matchQuantity

			if bestSell.Remaining <= 0 {
				heap.Pop(orderBook.SellOrders)
				bestSell.Status = models.OrderStatusFilled
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

			trade := models.Trade{
				ProjectID:    order.ProjectID,
				MilestoneID:  order.MilestoneID,
				OptionID:     order.OptionID,
				BuyOrderID:   bestBuy.ID,
				SellOrderID:  order.ID,
				BuyerID:      bestBuy.UserID,
				SellerID:     order.UserID,
				Quantity:     matchQuantity,
				Price:        bestBuy.Price,
				TotalAmount:  int64(float64(matchQuantity) * bestBuy.Price),
				CreatedAt:    time.Now(),
			}

			trades = append(trades, trade)

			remaining -= matchQuantity
			bestBuy.Remaining -= matchQuantity
			bestBuy.Filled += matchQuantity

			if bestBuy.Remaining <= 0 {
				heap.Pop(orderBook.BuyOrders)
				bestBuy.Status = models.OrderStatusFilled
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
	} else if order.Filled > 0 {
		order.Status = models.OrderStatusPartial
	}

	return trades
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

func (me *MatchingEngine) persistTrades(trades []models.Trade) {
	for _, trade := range trades {
		if err := me.db.Create(&trade).Error; err != nil {
			log.Printf("❌ Failed to persist trade: %v", err)
		}
	}
}

func (me *MatchingEngine) broadcastTrades(trades []models.Trade) {
	for _, trade := range trades {
		// 실시간 브로드캐스트
		redis.BroadcastTradeUpdate(trade.MilestoneID, trade.OptionID, trade)
		redis.BroadcastPriceChange(trade.MilestoneID, trade.OptionID, trade.Price)

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

func (me *MatchingEngine) updateStats(processingTime time.Duration) {
	me.mutex.Lock()
	defer me.mutex.Unlock()

	me.stats.OrdersProcessed++
	me.stats.TotalMatches++
	me.stats.LastMatchTime = time.Now()

	// 이동 평균으로 평균 매칭 시간 계산
	me.stats.AvgMatchTime = (me.stats.AvgMatchTime*0.95) + (processingTime.Seconds()*1000*0.05)
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
