package services

import (
	"blueprint-module/pkg/models"
	"blueprint-module/pkg/redis"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	redisClient "github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// 🌐 분산 매칭 엔진 (다중 서버 지원)
// Redis Streams + Distributed Locks + Event Sourcing

type DistributedMatchingEngine struct {
	db             *gorm.DB
	redisClient    *redisClient.Client
	sseService     *SSEService
	instanceID     string // 서버 인스턴스 고유 ID
	
	// 분산 락 및 상태 관리
	lockManager    *DistributedLockManager
	eventSourcing  *OrderEventSourcing
	
	// 실시간 처리
	orderStreams   *RedisStreamManager
	priceOracle    *DistributedPriceOracle
	
	// 로컬 캐시 (성능 최적화용)
	localCache     *LocalOrderBookCache
	
	// 컨트롤 채널
	ctx            context.Context
	cancel         context.CancelFunc
	wg             sync.WaitGroup
}

// OrderEvent 주문 이벤트 (Event Sourcing)
type OrderEvent struct {
	EventID     string                 `json:"event_id"`
	EventType   OrderEventType         `json:"event_type"`
	OrderID     uint                   `json:"order_id"`
	MilestoneID uint                   `json:"milestone_id"`
	OptionID    string                 `json:"option_id"`
	Payload     map[string]interface{} `json:"payload"`
	Timestamp   int64                  `json:"timestamp"`
	ServerID    string                 `json:"server_id"`
	Version     int                    `json:"version"`
}

type OrderEventType string

const (
	EventOrderCreated   OrderEventType = "ORDER_CREATED"
	EventOrderMatched   OrderEventType = "ORDER_MATCHED"
	EventOrderCancelled OrderEventType = "ORDER_CANCELLED"
	EventOrderExpired   OrderEventType = "ORDER_EXPIRED"
	EventTradeExecuted  OrderEventType = "TRADE_EXECUTED"
	EventPriceUpdated   OrderEventType = "PRICE_UPDATED"
)

// 🔐 분산 락 매니저
type DistributedLockManager struct {
	redisClient *redisClient.Client
}

func NewDistributedLockManager(redisClient *redisClient.Client) *DistributedLockManager {
	return &DistributedLockManager{
		redisClient: redisClient,
	}
}

// AcquireLock 분산 락 획득 (Redlock 알고리즘)
func (dlm *DistributedLockManager) AcquireLock(ctx context.Context, key string, ttl time.Duration, instanceID string) (bool, error) {
	script := `
		if redis.call("GET", KEYS[1]) == false then
			redis.call("SETEX", KEYS[1], ARGV[2], ARGV[1])
			return 1
		end
		return 0
	`
	
	result, err := dlm.redisClient.Eval(ctx, script, []string{fmt.Sprintf("lock:%s", key)}, instanceID, int(ttl.Seconds())).Result()
	if err != nil {
		return false, err
	}
	
	return result.(int64) == 1, nil
}

// ReleaseLock 분산 락 해제
func (dlm *DistributedLockManager) ReleaseLock(ctx context.Context, key string, instanceID string) error {
	script := `
		if redis.call("GET", KEYS[1]) == ARGV[1] then
			return redis.call("DEL", KEYS[1])
		else
			return 0
		end
	`
	
	_, err := dlm.redisClient.Eval(ctx, script, []string{fmt.Sprintf("lock:%s", key)}, instanceID).Result()
	return err
}

// 📊 이벤트 소싱 기반 주문 관리
type OrderEventSourcing struct {
	redisClient *redisClient.Client
}

func NewOrderEventSourcing(redisClient *redisClient.Client) *OrderEventSourcing {
	return &OrderEventSourcing{
		redisClient: redisClient,
	}
}

// AppendEvent 이벤트 추가 (원자성 보장)
func (oes *OrderEventSourcing) AppendEvent(ctx context.Context, marketKey string, event *OrderEvent) error {
	eventJSON, err := json.Marshal(event)
	if err != nil {
		return err
	}
	
	streamKey := fmt.Sprintf("events:%s", marketKey)
	
	// Redis Streams에 이벤트 추가
	_, err = oes.redisClient.XAdd(ctx, &redisClient.XAddArgs{
		Stream: streamKey,
		Values: map[string]interface{}{
			"event_id":   event.EventID,
			"event_type": event.EventType,
			"order_id":   event.OrderID,
			"payload":    string(eventJSON),
			"timestamp":  event.Timestamp,
			"server_id":  event.ServerID,
		},
	}).Result()
	
	return err
}

// ReadEvents 이벤트 읽기 (특정 시점부터)
func (oes *OrderEventSourcing) ReadEvents(ctx context.Context, marketKey string, fromID string) ([]*OrderEvent, error) {
	streamKey := fmt.Sprintf("events:%s", marketKey)
	
	result, err := oes.redisClient.XRead(ctx, &redisClient.XReadArgs{
		Streams: []string{streamKey, fromID},
		Count:   100,
		Block:   0,
	}).Result()
	
	if err != nil {
		return nil, err
	}
	
	var events []*OrderEvent
	for _, stream := range result {
		for _, message := range stream.Messages {
			var event OrderEvent
			if payloadStr, ok := message.Values["payload"].(string); ok {
				if err := json.Unmarshal([]byte(payloadStr), &event); err == nil {
					events = append(events, &event)
				}
			}
		}
	}
	
	return events, nil
}

// 🌊 Redis Streams 기반 실시간 주문 처리
type RedisStreamManager struct {
	redisClient *redisClient.Client
	instanceID  string
}

func NewRedisStreamManager(redisClient *redisClient.Client, instanceID string) *RedisStreamManager {
	return &RedisStreamManager{
		redisClient: redisClient,
		instanceID:  instanceID,
	}
}

// ProcessOrderStream 주문 스트림 처리 (컨슈머 그룹)
func (rsm *RedisStreamManager) ProcessOrderStream(ctx context.Context, marketKey string, processor func(*OrderEvent) error) error {
	streamKey := fmt.Sprintf("orders:%s", marketKey)
	consumerGroup := "matching-engines"
	consumerName := fmt.Sprintf("engine-%s", rsm.instanceID)
	
	// 컨슈머 그룹 생성 (이미 존재하면 무시)
	rsm.redisClient.XGroupCreate(ctx, streamKey, consumerGroup, "0").Err()
	
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			// 스트림에서 메시지 읽기
			result, err := rsm.redisClient.XReadGroup(ctx, &redisClient.XReadGroupArgs{
				Group:    consumerGroup,
				Consumer: consumerName,
				Streams:  []string{streamKey, ">"},
				Count:    10,
				Block:    time.Second,
			}).Result()
			
			if err != nil {
				if err == redisClient.Nil {
					continue
				}
				log.Printf("❌ Stream read error: %v", err)
				continue
			}
			
			// 메시지 처리
			for _, stream := range result {
				for _, message := range stream.Messages {
					if err := rsm.processMessage(ctx, streamKey, consumerGroup, message, processor); err != nil {
						log.Printf("❌ Message processing error: %v", err)
					}
				}
			}
		}
	}
}

// processMessage 개별 메시지 처리
func (rsm *RedisStreamManager) processMessage(ctx context.Context, streamKey, consumerGroup string, message redisClient.XMessage, processor func(*OrderEvent) error) error {
	// 메시지를 OrderEvent로 변환
	var event OrderEvent
	if payloadStr, ok := message.Values["payload"].(string); ok {
		if err := json.Unmarshal([]byte(payloadStr), &event); err != nil {
			return err
		}
	} else {
		return fmt.Errorf("invalid message format")
	}
	
	// 이벤트 처리
	if err := processor(&event); err != nil {
		return err
	}
	
	// 메시지 확인 (ACK)
	return rsm.redisClient.XAck(ctx, streamKey, consumerGroup, message.ID).Err()
}

// 💰 분산 가격 오라클
type DistributedPriceOracle struct {
	redisClient *redisClient.Client
}

func NewDistributedPriceOracle(redisClient *redisClient.Client) *DistributedPriceOracle {
	return &DistributedPriceOracle{
		redisClient: redisClient,
	}
}

// UpdatePrice 가격 업데이트 (원자적)
func (dpo *DistributedPriceOracle) UpdatePrice(ctx context.Context, marketKey string, price float64, volume int64) error {
	script := `
		local priceKey = "price:" .. KEYS[1]
		local volumeKey = "volume:" .. KEYS[1]
		local historyKey = "history:" .. KEYS[1]
		
		-- 현재 가격과 볼륨 업데이트
		redis.call("SET", priceKey, ARGV[1])
		redis.call("INCRBY", volumeKey, ARGV[2])
		
		-- 가격 히스토리 추가 (최근 1000개 유지)
		local timestamp = redis.call("TIME")[1]
		redis.call("ZADD", historyKey, timestamp, ARGV[1])
		redis.call("ZREMRANGEBYRANK", historyKey, 0, -1001)
		
		return 1
	`
	
	_, err := dpo.redisClient.Eval(ctx, script, []string{marketKey}, price, volume).Result()
	return err
}

// GetPrice 현재 가격 조회
func (dpo *DistributedPriceOracle) GetPrice(ctx context.Context, marketKey string) (float64, error) {
	priceStr, err := dpo.redisClient.Get(ctx, fmt.Sprintf("price:%s", marketKey)).Result()
	if err != nil {
		if err == redisClient.Nil {
			return 0, nil
		}
		return 0, err
	}
	
	return strconv.ParseFloat(priceStr, 64)
}

// 🚀 분산 매칭 엔진 생성자
func NewDistributedMatchingEngine(db *gorm.DB, sseService *SSEService) *DistributedMatchingEngine {
	ctx, cancel := context.WithCancel(context.Background())
	instanceID := fmt.Sprintf("engine-%d", time.Now().UnixNano())
	
	redisClient := redis.GetClient()
	
	return &DistributedMatchingEngine{
		db:            db,
		redisClient:   redisClient,
		sseService:    sseService,
		instanceID:    instanceID,
		ctx:           ctx,
		cancel:        cancel,
		lockManager:   NewDistributedLockManager(redisClient),
		eventSourcing: NewOrderEventSourcing(redisClient),
		orderStreams:  NewRedisStreamManager(redisClient, instanceID),
		priceOracle:   NewDistributedPriceOracle(redisClient),
		localCache:    NewLocalOrderBookCache(),
	}
}

// Start 분산 매칭 엔진 시작
func (dme *DistributedMatchingEngine) Start() error {
	log.Printf("🌐 Starting Distributed Matching Engine: %s", dme.instanceID)
	
	// 주요 마켓 스트림 처리 시작
	markets, err := dme.getActiveMarkets()
	if err != nil {
		return err
	}
	
	for _, marketKey := range markets {
		dme.wg.Add(1)
		go func(market string) {
			defer dme.wg.Done()
			dme.orderStreams.ProcessOrderStream(dme.ctx, market, dme.processOrderEvent)
		}(marketKey)
	}
	
	// 가격 오라클 업데이터 시작
	dme.wg.Add(1)
	go func() {
		defer dme.wg.Done()
		dme.runPriceOracleUpdater()
	}()
	
	log.Printf("✅ Distributed Matching Engine started with %d market streams", len(markets))
	return nil
}

// Stop 분산 매칭 엔진 정지
func (dme *DistributedMatchingEngine) Stop() error {
	log.Printf("🛑 Stopping Distributed Matching Engine: %s", dme.instanceID)
	dme.cancel()
	dme.wg.Wait()
	log.Printf("✅ Distributed Matching Engine stopped")
	return nil
}

// SubmitOrder 주문 제출 (분산 환경)
func (dme *DistributedMatchingEngine) SubmitOrder(order *models.Order) (*MatchingResult, error) {
	marketKey := dme.getMarketKey(order.MilestoneID, order.OptionID)
	
	// 1. 분산 락 획득 (매칭 원자성 보장)
	lockKey := fmt.Sprintf("match:%s", marketKey)
	locked, err := dme.lockManager.AcquireLock(dme.ctx, lockKey, 5*time.Second, dme.instanceID)
	if err != nil {
		return nil, fmt.Errorf("failed to acquire lock: %v", err)
	}
	if !locked {
		return nil, fmt.Errorf("market is locked by another instance")
	}
	defer dme.lockManager.ReleaseLock(dme.ctx, lockKey, dme.instanceID)
	
	// 2. 주문 이벤트 생성
	event := &OrderEvent{
		EventID:     fmt.Sprintf("%s-%d", dme.instanceID, time.Now().UnixNano()),
		EventType:   EventOrderCreated,
		OrderID:     order.ID,
		MilestoneID: order.MilestoneID,
		OptionID:    order.OptionID,
		Payload: map[string]interface{}{
			"order": order,
		},
		Timestamp: time.Now().UnixMilli(),
		ServerID:  dme.instanceID,
		Version:   1,
	}
	
	// 3. 이벤트 소싱에 기록
	if err := dme.eventSourcing.AppendEvent(dme.ctx, marketKey, event); err != nil {
		return nil, fmt.Errorf("failed to append event: %v", err)
	}
	
	// 4. 매칭 실행
	return dme.executeMatching(marketKey, order)
}

// 핵심 매칭 로직들...
func (dme *DistributedMatchingEngine) executeMatching(marketKey string, order *models.Order) (*MatchingResult, error) {
	// 1. Redis에서 현재 주문장 상태 로드
	orderBook, err := dme.loadOrderBook(marketKey)
	if err != nil {
		return nil, fmt.Errorf("failed to load order book: %v", err)
	}

	// 2. 매칭 엔진 실행
	trades := []models.Trade{}
	remainingQuantity := order.Quantity

	if order.Side == "buy" {
		// Buy 주문: 가장 낮은 Ask부터 매칭
		for len(orderBook.Asks) > 0 && remainingQuantity > 0 {
			askOrder := orderBook.Asks[0]
			if askOrder.Price > order.Price {
				break // 가격이 맞지 않음
			}

			// 거래 체결
			tradeQuantity := minInt64(remainingQuantity, askOrder.Quantity)
			trade := models.Trade{
				BuyOrderID:   order.ID,
				SellOrderID:  askOrder.ID,
				MilestoneID:  order.MilestoneID,
				OptionID:     order.OptionID,
				BuyerID:      order.UserID,
				SellerID:     askOrder.UserID,
				Quantity:     tradeQuantity,
				Price:        askOrder.Price,
				CreatedAt:    time.Now(),
			}

			trades = append(trades, trade)
			remainingQuantity -= tradeQuantity

			// Ask 주문 업데이트
			askOrder.Quantity -= tradeQuantity
			if askOrder.Quantity == 0 {
				orderBook.Asks = orderBook.Asks[1:]
			}

			// 거래 이벤트 발생
			dme.emitTradeEvent(marketKey, &trade)
		}
	} else {
		// Sell 주문: 가장 높은 Bid부터 매칭
		for len(orderBook.Bids) > 0 && remainingQuantity > 0 {
			bidOrder := orderBook.Bids[0]
			if bidOrder.Price < order.Price {
				break // 가격이 맞지 않음
			}

			// 거래 체결
			tradeQuantity := minInt64(remainingQuantity, bidOrder.Quantity)
			trade := models.Trade{
				BuyOrderID:   bidOrder.ID,
				SellOrderID:  order.ID,
				MilestoneID:  order.MilestoneID,
				OptionID:     order.OptionID,
				BuyerID:      bidOrder.UserID,
				SellerID:     order.UserID,
				Quantity:     tradeQuantity,
				Price:        bidOrder.Price,
				CreatedAt:    time.Now(),
			}

			trades = append(trades, trade)
			remainingQuantity -= tradeQuantity

			// Bid 주문 업데이트
			bidOrder.Quantity -= tradeQuantity
			if bidOrder.Quantity == 0 {
				orderBook.Bids = orderBook.Bids[1:]
			}

			// 거래 이벤트 발생
			dme.emitTradeEvent(marketKey, &trade)
		}
	}

	// 3. 남은 수량이 있으면 주문장에 추가
	if remainingQuantity > 0 {
		order.Quantity = remainingQuantity
		dme.addOrderToBook(orderBook, order)
	}

	// 4. 업데이트된 주문장을 Redis에 저장
	if err := dme.saveOrderBook(marketKey, orderBook); err != nil {
		return nil, fmt.Errorf("failed to save order book: %v", err)
	}

	// 5. 가격 오라클 업데이트
	if len(trades) > 0 {
		lastTrade := trades[len(trades)-1]
		totalVolume := int64(0)
		for _, trade := range trades {
			totalVolume += int64(trade.Quantity)
		}
		dme.priceOracle.UpdatePrice(dme.ctx, marketKey, lastTrade.Price, totalVolume)
	}

	// 6. SSE로 실시간 업데이트 전송
	dme.broadcastMarketUpdate(marketKey, orderBook, trades)

	return &MatchingResult{
		Trades:   trades,
		Error:    nil,
		Executed: len(trades) > 0,
	}, nil
}

func (dme *DistributedMatchingEngine) processOrderEvent(event *OrderEvent) error {
	marketKey := dme.getMarketKey(event.MilestoneID, event.OptionID)

	switch event.EventType {
	case EventOrderCreated:
		log.Printf("🔄 Processing order created event: %s", event.EventID)
		// 주문 생성 이벤트는 이미 executeMatching에서 처리됨
		
	case EventOrderCancelled:
		log.Printf("❌ Processing order cancelled event: %s", event.EventID)
		return dme.handleOrderCancellation(marketKey, event.OrderID)
		
	case EventOrderExpired:
		log.Printf("⏰ Processing order expired event: %s", event.EventID)
		return dme.handleOrderExpiry(marketKey, event.OrderID)
		
	case EventTradeExecuted:
		log.Printf("💰 Processing trade executed event: %s", event.EventID)
		// 거래 실행 이벤트는 SSE 브로드캐스팅용
		
	case EventPriceUpdated:
		log.Printf("📊 Processing price updated event: %s", event.EventID)
		// 가격 업데이트는 이미 priceOracle에서 처리됨
		
	default:
		log.Printf("⚠️ Unknown event type: %s", event.EventType)
	}
	
	return nil
}

func (dme *DistributedMatchingEngine) getActiveMarkets() ([]string, error) {
	// 활성 상태의 마일스톤들을 조회
	var milestones []models.Milestone
	err := dme.db.Where("status IN ?", []string{"funding", "active"}).Find(&milestones).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get active milestones: %v", err)
	}
	
	var markets []string
	for _, milestone := range milestones {
		// 각 마일스톤에 대해 success/fail 마켓 생성
		markets = append(markets, fmt.Sprintf("%d:success", milestone.ID))
		markets = append(markets, fmt.Sprintf("%d:fail", milestone.ID))
	}
	
	log.Printf("🎯 Found %d active markets from %d milestones", len(markets), len(milestones))
	return markets, nil
}

func (dme *DistributedMatchingEngine) getMarketKey(milestoneID uint, optionID string) string {
	return fmt.Sprintf("%d:%s", milestoneID, optionID)
}

func (dme *DistributedMatchingEngine) parseMarketKey(marketKey string) (uint, string) {
	parts := strings.Split(marketKey, ":")
	if len(parts) != 2 {
		return 0, ""
	}
	
	milestoneID, err := strconv.ParseUint(parts[0], 10, 32)
	if err != nil {
		return 0, ""
	}
	
	return uint(milestoneID), parts[1]
}

func (dme *DistributedMatchingEngine) runPriceOracleUpdater() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-dme.ctx.Done():
			return
		case <-ticker.C:
			// 가격 오라클 업데이트 로직
		}
	}
}

// 🏎️ 로컬 캐시 (성능 최적화)
type LocalOrderBookCache struct {
	cache map[string]*CachedOrderBook
	mutex sync.RWMutex
}

type CachedOrderBook struct {
	BestBid   float64   `json:"best_bid"`
	BestAsk   float64   `json:"best_ask"`
	LastPrice float64   `json:"last_price"`
	Volume24h int64     `json:"volume_24h"`
	UpdatedAt time.Time `json:"updated_at"`
}

func NewLocalOrderBookCache() *LocalOrderBookCache {
	return &LocalOrderBookCache{
		cache: make(map[string]*CachedOrderBook),
	}
}

func (lobc *LocalOrderBookCache) Get(marketKey string) (*CachedOrderBook, bool) {
	lobc.mutex.RLock()
	defer lobc.mutex.RUnlock()
	
	book, exists := lobc.cache[marketKey]
	return book, exists
}

func (lobc *LocalOrderBookCache) Set(marketKey string, book *CachedOrderBook) {
	lobc.mutex.Lock()
	defer lobc.mutex.Unlock()
	
	lobc.cache[marketKey] = book
}

// 📚 Redis 기반 주문장 관리
type DistributedOrderBook struct {
	MarketKey string          `json:"market_key"`
	Bids      []*models.Order `json:"bids"`  // 매수 주문 (가격 높은 순)
	Asks      []*models.Order `json:"asks"`  // 매도 주문 (가격 낮은 순)
	LastPrice float64         `json:"last_price"`
	UpdatedAt time.Time       `json:"updated_at"`
}

// loadOrderBook Redis에서 주문장 로드
func (dme *DistributedMatchingEngine) loadOrderBook(marketKey string) (*DistributedOrderBook, error) {
	orderBookKey := fmt.Sprintf("orderbook:%s", marketKey)
	
	// Redis에서 주문장 데이터 가져오기
	data, err := dme.redisClient.Get(dme.ctx, orderBookKey).Result()
	if err != nil {
		if err == redisClient.Nil {
			// 새로운 주문장 생성
			return &DistributedOrderBook{
				MarketKey: marketKey,
				Bids:      []*models.Order{},
				Asks:      []*models.Order{},
				UpdatedAt: time.Now(),
			}, nil
		}
		return nil, err
	}
	
	var orderBook DistributedOrderBook
	if err := json.Unmarshal([]byte(data), &orderBook); err != nil {
		return nil, err
	}
	
	return &orderBook, nil
}

// saveOrderBook Redis에 주문장 저장
func (dme *DistributedMatchingEngine) saveOrderBook(marketKey string, orderBook *DistributedOrderBook) error {
	orderBook.UpdatedAt = time.Now()
	orderBookKey := fmt.Sprintf("orderbook:%s", marketKey)
	
	data, err := json.Marshal(orderBook)
	if err != nil {
		return err
	}
	
	// TTL 설정 (24시간)
	return dme.redisClient.Set(dme.ctx, orderBookKey, data, 24*time.Hour).Err()
}

// addOrderToBook 주문장에 주문 추가
func (dme *DistributedMatchingEngine) addOrderToBook(orderBook *DistributedOrderBook, order *models.Order) {
	if order.Side == "buy" {
		// Bid 추가 (가격 높은 순으로 정렬)
		inserted := false
		for i, existingOrder := range orderBook.Bids {
			if order.Price > existingOrder.Price {
				orderBook.Bids = append(orderBook.Bids[:i], append([]*models.Order{order}, orderBook.Bids[i:]...)...)
				inserted = true
				break
			}
		}
		if !inserted {
			orderBook.Bids = append(orderBook.Bids, order)
		}
	} else {
		// Ask 추가 (가격 낮은 순으로 정렬)
		inserted := false
		for i, existingOrder := range orderBook.Asks {
			if order.Price < existingOrder.Price {
				orderBook.Asks = append(orderBook.Asks[:i], append([]*models.Order{order}, orderBook.Asks[i:]...)...)
				inserted = true
				break
			}
		}
		if !inserted {
			orderBook.Asks = append(orderBook.Asks, order)
		}
	}
}

// emitTradeEvent 거래 이벤트 발행
func (dme *DistributedMatchingEngine) emitTradeEvent(marketKey string, trade *models.Trade) {
	event := &OrderEvent{
		EventID:     fmt.Sprintf("trade-%s-%d", dme.instanceID, time.Now().UnixNano()),
		EventType:   EventTradeExecuted,
		MilestoneID: trade.MilestoneID,
		OptionID:    trade.OptionID,
		Payload: map[string]interface{}{
			"trade": trade,
		},
		Timestamp: time.Now().UnixMilli(),
		ServerID:  dme.instanceID,
		Version:   1,
	}
	
	dme.eventSourcing.AppendEvent(dme.ctx, marketKey, event)
}

// broadcastMarketUpdate SSE로 마켓 업데이트 브로드캐스트
func (dme *DistributedMatchingEngine) broadcastMarketUpdate(marketKey string, orderBook *DistributedOrderBook, trades []models.Trade) {
	if dme.sseService == nil {
		return
	}
	
	// 주문장 상태 업데이트
	orderBookData := map[string]interface{}{
		"bids":       orderBook.Bids,
		"asks":       orderBook.Asks,
		"timestamp":  time.Now().UnixMilli(),
	}
	
	// Extract milestone and option IDs from marketKey
	milestoneID, optionID := dme.parseMarketKey(marketKey)
	dme.sseService.BroadcastOrderBookUpdate(milestoneID, optionID, orderBookData)
	
	// 거래 내역 브로드캐스트
	if len(trades) > 0 {
		for _, trade := range trades {
			tradeData := map[string]interface{}{
				"trade":       trade,
				"timestamp":   time.Now().UnixMilli(),
			}
			
			dme.sseService.BroadcastTradeUpdate(trade.MilestoneID, trade.OptionID, tradeData)
		}
	}
}

// handleOrderCancellation 주문 취소 처리
func (dme *DistributedMatchingEngine) handleOrderCancellation(marketKey string, orderID uint) error {
	lockKey := fmt.Sprintf("cancel:%s:%d", marketKey, orderID)
	locked, err := dme.lockManager.AcquireLock(dme.ctx, lockKey, 5*time.Second, dme.instanceID)
	if err != nil || !locked {
		return fmt.Errorf("failed to acquire cancellation lock")
	}
	defer dme.lockManager.ReleaseLock(dme.ctx, lockKey, dme.instanceID)
	
	orderBook, err := dme.loadOrderBook(marketKey)
	if err != nil {
		return err
	}
	
	// Bids에서 주문 제거
	for i, order := range orderBook.Bids {
		if order.ID == orderID {
			orderBook.Bids = append(orderBook.Bids[:i], orderBook.Bids[i+1:]...)
			break
		}
	}
	
	// Asks에서 주문 제거
	for i, order := range orderBook.Asks {
		if order.ID == orderID {
			orderBook.Asks = append(orderBook.Asks[:i], orderBook.Asks[i+1:]...)
			break
		}
	}
	
	return dme.saveOrderBook(marketKey, orderBook)
}

// handleOrderExpiry 주문 만료 처리
func (dme *DistributedMatchingEngine) handleOrderExpiry(marketKey string, orderID uint) error {
	return dme.handleOrderCancellation(marketKey, orderID)
}

// minInt64 utility function
func minInt64(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

// 🔄 CQRS Pattern: Command Query Responsibility Segregation
// 쓰기 작업과 읽기 작업을 분리하여 성능과 확장성 향상

// TradingCommandHandler 거래 명령 처리 (쓰기 작업)
type TradingCommandHandler struct {
	matchingEngine *DistributedMatchingEngine
}

func NewTradingCommandHandler(engine *DistributedMatchingEngine) *TradingCommandHandler {
	return &TradingCommandHandler{
		matchingEngine: engine,
	}
}

// CreateOrderCommand 주문 생성 명령
type CreateOrderCommand struct {
	UserID      uint    `json:"user_id"`
	MilestoneID uint    `json:"milestone_id"`
	OptionID    string  `json:"option_id"`
	Type        string  `json:"type"`        // buy/sell
	Quantity    int64   `json:"quantity"`
	Price       float64 `json:"price"`
}

// CancelOrderCommand 주문 취소 명령
type CancelOrderCommand struct {
	UserID  uint `json:"user_id"`
	OrderID uint `json:"order_id"`
}

// HandleCreateOrder 주문 생성 명령 처리
func (tch *TradingCommandHandler) HandleCreateOrder(cmd *CreateOrderCommand) (*MatchingResult, error) {
	// 1. 명령 검증
	if err := tch.validateCreateOrderCommand(cmd); err != nil {
		return nil, fmt.Errorf("invalid command: %v", err)
	}
	
	// 2. Order 모델 생성
	order := &models.Order{
		UserID:      cmd.UserID,
		MilestoneID: cmd.MilestoneID,
		OptionID:    cmd.OptionID,
		Side:        models.OrderSide(cmd.Type),
		Quantity:    cmd.Quantity,
		Price:       cmd.Price,
		Status:      models.OrderStatusPending,
		CreatedAt:   time.Now(),
	}
	
	// 3. 분산 매칭 엔진으로 처리
	return tch.matchingEngine.SubmitOrder(order)
}

// HandleCancelOrder 주문 취소 명령 처리
func (tch *TradingCommandHandler) HandleCancelOrder(cmd *CancelOrderCommand) error {
	// 1. 명령 검증
	if cmd.OrderID == 0 || cmd.UserID == 0 {
		return fmt.Errorf("invalid cancel order command")
	}
	
	// 2. 주문 취소 이벤트 생성
	marketKey := "unknown" // 실제로는 주문 조회 후 결정
	event := &OrderEvent{
		EventID:     fmt.Sprintf("cancel-%s-%d", tch.matchingEngine.instanceID, time.Now().UnixNano()),
		EventType:   EventOrderCancelled,
		OrderID:     cmd.OrderID,
		Payload: map[string]interface{}{
			"user_id": cmd.UserID,
		},
		Timestamp: time.Now().UnixMilli(),
		ServerID:  tch.matchingEngine.instanceID,
		Version:   1,
	}
	
	// 3. 이벤트 소싱에 기록
	return tch.matchingEngine.eventSourcing.AppendEvent(tch.matchingEngine.ctx, marketKey, event)
}

func (tch *TradingCommandHandler) validateCreateOrderCommand(cmd *CreateOrderCommand) error {
	if cmd.UserID == 0 {
		return fmt.Errorf("user_id is required")
	}
	if cmd.MilestoneID == 0 {
		return fmt.Errorf("milestone_id is required")
	}
	if cmd.OptionID == "" {
		return fmt.Errorf("option_id is required")
	}
	if cmd.Type != "buy" && cmd.Type != "sell" {
		return fmt.Errorf("type must be 'buy' or 'sell'")
	}
	if cmd.Quantity <= 0 {
		return fmt.Errorf("quantity must be positive")
	}
	if cmd.Price <= 0 {
		return fmt.Errorf("price must be positive")
	}
	return nil
}

// TradingQueryHandler 거래 조회 처리 (읽기 작업)
type TradingQueryHandler struct {
	redisClient *redisClient.Client
	db          *gorm.DB
}

func NewTradingQueryHandler(redisClient *redisClient.Client, db *gorm.DB) *TradingQueryHandler {
	return &TradingQueryHandler{
		redisClient: redisClient,
		db:          db,
	}
}

// MarketDataQuery 마켓 데이터 조회
type MarketDataQuery struct {
	MilestoneID uint   `json:"milestone_id"`
	OptionID    string `json:"option_id"`
}

// OrderBookQuery 주문장 조회
type OrderBookQuery struct {
	MilestoneID uint   `json:"milestone_id"`
	OptionID    string `json:"option_id"`
	Depth       int    `json:"depth"` // 주문장 깊이
}

// UserOrdersQuery 사용자 주문 내역 조회
type UserOrdersQuery struct {
	UserID uint   `json:"user_id"`
	Status string `json:"status,omitempty"`
	Limit  int    `json:"limit,omitempty"`
}

// GetMarketData 마켓 데이터 조회
func (tqh *TradingQueryHandler) GetMarketData(query *MarketDataQuery) (*MarketDataView, error) {
	marketKey := fmt.Sprintf("%d:%s", query.MilestoneID, query.OptionID)
	
	// Redis에서 실시간 데이터 조회
	pipe := tqh.redisClient.Pipeline()
	
	priceCmd := pipe.Get(context.Background(), fmt.Sprintf("price:%s", marketKey))
	volumeCmd := pipe.Get(context.Background(), fmt.Sprintf("volume:%s", marketKey))
	historyCmd := pipe.ZRevRange(context.Background(), fmt.Sprintf("history:%s", marketKey), 0, 23) // 최근 24개
	
	_, err := pipe.Exec(context.Background())
	if err != nil && err != redisClient.Nil {
		return nil, err
	}
	
	// 결과 파싱
	price, _ := priceCmd.Float64()
	volume, _ := volumeCmd.Int64()
	history := historyCmd.Val()
	
	return &MarketDataView{
		MarketKey:   marketKey,
		MilestoneID: query.MilestoneID,
		OptionID:    query.OptionID,
		LastPrice:   price,
		Volume24h:   volume,
		PriceHistory: history,
		UpdatedAt:   time.Now(),
	}, nil
}

// GetOrderBook 주문장 조회
func (tqh *TradingQueryHandler) GetOrderBook(query *OrderBookQuery) (*OrderBookView, error) {
	marketKey := fmt.Sprintf("%d:%s", query.MilestoneID, query.OptionID)
	orderBookKey := fmt.Sprintf("orderbook:%s", marketKey)
	
	data, err := tqh.redisClient.Get(context.Background(), orderBookKey).Result()
	if err != nil {
		if err == redisClient.Nil {
			return &OrderBookView{
				MarketKey: marketKey,
				Bids:      []OrderBookEntry{},
				Asks:      []OrderBookEntry{},
			}, nil
		}
		return nil, err
	}
	
	var orderBook DistributedOrderBook
	if err := json.Unmarshal([]byte(data), &orderBook); err != nil {
		return nil, err
	}
	
	// 뷰 모델로 변환
	view := &OrderBookView{
		MarketKey: marketKey,
		Bids:      tqh.convertToOrderBookEntries(orderBook.Bids, query.Depth),
		Asks:      tqh.convertToOrderBookEntries(orderBook.Asks, query.Depth),
		UpdatedAt: orderBook.UpdatedAt,
	}
	
	return view, nil
}

// GetUserOrders 사용자 주문 내역 조회
func (tqh *TradingQueryHandler) GetUserOrders(query *UserOrdersQuery) ([]*UserOrderView, error) {
	dbQuery := tqh.db.Where("user_id = ?", query.UserID)
	
	if query.Status != "" {
		dbQuery = dbQuery.Where("status = ?", query.Status)
	}
	
	if query.Limit > 0 {
		dbQuery = dbQuery.Limit(query.Limit)
	} else {
		dbQuery = dbQuery.Limit(50) // 기본 50개
	}
	
	var orders []models.Order
	err := dbQuery.Order("created_at DESC").Find(&orders).Error
	if err != nil {
		return nil, err
	}
	
	// 뷰 모델로 변환
	var views []*UserOrderView
	for _, order := range orders {
		views = append(views, &UserOrderView{
			ID:          order.ID,
			MilestoneID: order.MilestoneID,
			OptionID:    order.OptionID,
			Type:        string(order.Side),
			Quantity:    int(order.Quantity),
			Price:       order.Price,
			Status:      string(order.Status),
			CreatedAt:   order.CreatedAt,
		})
	}
	
	return views, nil
}

func (tqh *TradingQueryHandler) convertToOrderBookEntries(orders []*models.Order, depth int) []OrderBookEntry {
	entries := []OrderBookEntry{}
	
	limit := len(orders)
	if depth > 0 && depth < limit {
		limit = depth
	}
	
	for i := 0; i < limit; i++ {
		entry := OrderBookEntry{
			Price:    orders[i].Price,
			Quantity: int(orders[i].Quantity),
		}
		entries = append(entries, entry)
	}
	
	return entries
}

// 🔍 View Models (읽기 전용 모델들)
type MarketDataView struct {
	MarketKey    string    `json:"market_key"`
	MilestoneID  uint      `json:"milestone_id"`
	OptionID     string    `json:"option_id"`
	LastPrice    float64   `json:"last_price"`
	Volume24h    int64     `json:"volume_24h"`
	PriceHistory []string  `json:"price_history"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type OrderBookView struct {
	MarketKey string             `json:"market_key"`
	Bids      []OrderBookEntry   `json:"bids"`
	Asks      []OrderBookEntry   `json:"asks"`
	UpdatedAt time.Time          `json:"updated_at"`
}

type OrderBookEntry struct {
	Price    float64 `json:"price"`
	Quantity int     `json:"quantity"`
}

type UserOrderView struct {
	ID          uint      `json:"id"`
	MilestoneID uint      `json:"milestone_id"`
	OptionID    string    `json:"option_id"`
	Type        string    `json:"type"`
	Quantity    int       `json:"quantity"`
	Price       float64   `json:"price"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}