package queue

import (
	"blueprint-module/pkg/redis"
	"encoding/json"
	"fmt"
	"time"

	redislib "github.com/redis/go-redis/v9"
	"golang.org/x/net/context"
)

var ctx = context.Background()

// EventType 이벤트 타입
type EventType string

const (
	EventTypeTrade       EventType = "trade"
	EventTypePriceUpdate EventType = "price_update"
	EventTypeMarketMake  EventType = "market_make"
	EventTypeUserJoin    EventType = "user_join"
	EventTypeUserLeave   EventType = "user_leave"

	// 🆕 비동기 초기화 이벤트들
	EventTypeUserCreated EventType = "user_created"  // 회원가입 후 처리
	EventTypeWalletCreate EventType = "wallet_create" // 지갑 생성
	EventTypeMarketInit  EventType = "market_init"   // 마켓 초기화
	EventTypeWelcomeUser EventType = "welcome_user"  // 웰컴 처리 (이메일, 온보딩 등)
)

// QueueEvent 큐 이벤트 구조체
type QueueEvent struct {
	ID          string                 `json:"id"`
	Type        EventType              `json:"type"`
	MilestoneID uint                   `json:"milestone_id"`
	OptionID    string                 `json:"option_id,omitempty"`
	UserID      uint                   `json:"user_id,omitempty"`
	Data        map[string]interface{} `json:"data"`
	Timestamp   int64                  `json:"timestamp"`
	Retry       int                    `json:"retry"`
}

// TradeEventData 거래 이벤트 데이터
type TradeEventData struct {
	TradeID     uint    `json:"trade_id"`
	BuyerID     uint    `json:"buyer_id"`
	SellerID    uint    `json:"seller_id"`
	Quantity    int64   `json:"quantity"`
	Price       float64 `json:"price"`
	TotalAmount int64   `json:"total_amount"`
}

// PriceUpdateEventData 가격 업데이트 이벤트 데이터
type PriceUpdateEventData struct {
	OldPrice float64 `json:"old_price"`
	NewPrice float64 `json:"new_price"`
	Volume   int64   `json:"volume"`
}

// MarketMakeEventData 마켓 메이킹 이벤트 데이터
type MarketMakeEventData struct {
	Action       string  `json:"action"` // "create_orders", "cancel_orders", "adjust_spread"
	CurrentPrice float64 `json:"current_price"`
	Spread       float64 `json:"spread"`
	Volume       int64   `json:"volume"`
}

// UserCreatedEventData 회원가입 완료 이벤트 데이터
type UserCreatedEventData struct {
	UserID   uint   `json:"user_id"`
	Email    string `json:"email"`
	Username string `json:"username"`
	Provider string `json:"provider"` // "local", "google"
}

// WalletCreateEventData 지갑 생성 이벤트 데이터
type WalletCreateEventData struct {
	UserID        uint  `json:"user_id"`
	InitialAmount int64 `json:"initial_amount"` // 초기 지급 포인트
}

// MarketInitEventData 마켓 초기화 이벤트 데이터
type MarketInitEventData struct {
	ProjectID   uint     `json:"project_id"`
	MilestoneID uint     `json:"milestone_id"`
	Options     []string `json:"options"`
}

// WelcomeUserEventData 웰컴 처리 이벤트 데이터
type WelcomeUserEventData struct {
	UserID    uint   `json:"user_id"`
	Email     string `json:"email"`
	Username  string `json:"username"`
	FirstName string `json:"first_name,omitempty"`
}

// QueueNames 큐 이름들
const (
	QueueTrades      = "queue:trades"
	QueuePrices      = "queue:prices"
	QueueMarketMake  = "queue:market_make"
	QueueNotify      = "queue:notify"
	QueueAnalytics   = "queue:analytics"

	// 🆕 비동기 초기화 큐들
	QueueUserTasks   = "queue:user_tasks"   // 사용자 관련 후속 작업
	QueueWallet      = "queue:wallet"       // 지갑 생성/업데이트
	QueueMarket      = "queue:market"       // 마켓 초기화
	QueueWelcome     = "queue:welcome"      // 웰컴 처리
)

// Publisher 이벤트 발행자
type Publisher struct {
	client *redislib.Client
}

// NewPublisher 발행자 생성
func NewPublisher() *Publisher {
	return &Publisher{
		client: redis.Client,
	}
}

// EnqueueTradeWork 거래 작업을 큐에 추가 (기존 PublishTradeEvent)
func (p *Publisher) EnqueueTradeWork(milestoneID uint, optionID string, data TradeEventData) error {
	event := QueueEvent{
		ID:          fmt.Sprintf("trade_%d_%s_%d", milestoneID, optionID, time.Now().UnixNano()),
		Type:        EventTypeTrade,
		MilestoneID: milestoneID,
		OptionID:    optionID,
		Data: map[string]interface{}{
			"trade_id":     data.TradeID,
			"buyer_id":     data.BuyerID,
			"seller_id":    data.SellerID,
			"quantity":     data.Quantity,
			"price":        data.Price,
			"total_amount": data.TotalAmount,
		},
		Timestamp: time.Now().Unix(),
	}

	return p.publishEvent(QueueTrades, event)
}

// EnqueuePriceUpdateWork 가격 업데이트 작업을 큐에 추가 (기존 PublishPriceUpdateEvent)
func (p *Publisher) EnqueuePriceUpdateWork(milestoneID uint, optionID string, data PriceUpdateEventData) error {
	event := QueueEvent{
		ID:          fmt.Sprintf("price_%d_%s_%d", milestoneID, optionID, time.Now().UnixNano()),
		Type:        EventTypePriceUpdate,
		MilestoneID: milestoneID,
		OptionID:    optionID,
		Data: map[string]interface{}{
			"old_price": data.OldPrice,
			"new_price": data.NewPrice,
			"volume":    data.Volume,
		},
		Timestamp: time.Now().Unix(),
	}

	return p.publishEvent(QueuePrices, event)
}

// EnqueueMarketMakeWork 마켓 메이킹 작업을 큐에 추가 (기존 PublishMarketMakeEvent)
func (p *Publisher) EnqueueMarketMakeWork(milestoneID uint, optionID string, data MarketMakeEventData) error {
	event := QueueEvent{
		ID:          fmt.Sprintf("mm_%d_%s_%d", milestoneID, optionID, time.Now().UnixNano()),
		Type:        EventTypeMarketMake,
		MilestoneID: milestoneID,
		OptionID:    optionID,
		Data: map[string]interface{}{
			"action":        data.Action,
			"current_price": data.CurrentPrice,
			"spread":        data.Spread,
			"volume":        data.Volume,
		},
		Timestamp: time.Now().Unix(),
	}

	return p.publishEvent(QueueMarketMake, event)
}

// EnqueueUserCreated 사용자 생성 후 처리 작업을 큐에 추가
func (p *Publisher) EnqueueUserCreated(data UserCreatedEventData) error {
	event := QueueEvent{
		ID:       fmt.Sprintf("user_created_%d_%d", data.UserID, time.Now().UnixNano()),
		Type:     EventTypeUserCreated,
		UserID:   data.UserID,
		Data: map[string]interface{}{
			"user_id":  data.UserID,
			"email":    data.Email,
			"username": data.Username,
			"provider": data.Provider,
		},
		Timestamp: time.Now().Unix(),
	}

	return p.publishEvent(QueueUserTasks, event)
}

// EnqueueWalletCreate 지갑 생성 작업을 큐에 추가
func (p *Publisher) EnqueueWalletCreate(data WalletCreateEventData) error {
	event := QueueEvent{
		ID:     fmt.Sprintf("wallet_create_%d_%d", data.UserID, time.Now().UnixNano()),
		Type:   EventTypeWalletCreate,
		UserID: data.UserID,
		Data: map[string]interface{}{
			"user_id":        data.UserID,
			"initial_amount": data.InitialAmount,
		},
		Timestamp: time.Now().Unix(),
	}

	return p.publishEvent(QueueWallet, event)
}

// EnqueueMarketInit 마켓 초기화 작업을 큐에 추가
func (p *Publisher) EnqueueMarketInit(data MarketInitEventData) error {
	event := QueueEvent{
		ID:          fmt.Sprintf("market_init_%d_%d", data.MilestoneID, time.Now().UnixNano()),
		Type:        EventTypeMarketInit,
		MilestoneID: data.MilestoneID,
		Data: map[string]interface{}{
			"project_id":   data.ProjectID,
			"milestone_id": data.MilestoneID,
			"options":      data.Options,
		},
		Timestamp: time.Now().Unix(),
	}

	return p.publishEvent(QueueMarket, event)
}

// EnqueueWelcomeUser 웰컴 처리 작업을 큐에 추가
func (p *Publisher) EnqueueWelcomeUser(data WelcomeUserEventData) error {
	event := QueueEvent{
		ID:     fmt.Sprintf("welcome_user_%d_%d", data.UserID, time.Now().UnixNano()),
		Type:   EventTypeWelcomeUser,
		UserID: data.UserID,
		Data: map[string]interface{}{
			"user_id":    data.UserID,
			"email":      data.Email,
			"username":   data.Username,
			"first_name": data.FirstName,
		},
		Timestamp: time.Now().Unix(),
	}

	return p.publishEvent(QueueWelcome, event)
}

// publishEvent 내부 이벤트 발행 메서드
func (p *Publisher) publishEvent(queueName string, event QueueEvent) error {
	jsonData, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %v", err)
	}

	// Redis Streams에 이벤트 추가
	args := &redislib.XAddArgs{
		Stream: queueName,
		MaxLen: 10000, // 최대 10,000개 이벤트 유지
		Approx: true,
		Values: map[string]interface{}{
			"event": string(jsonData),
		},
	}

	_, err = p.client.XAdd(ctx, args).Result()
	if err != nil {
		return fmt.Errorf("failed to add event to stream: %v", err)
	}

	// log.Printf("📤 Published event: %s to %s", event.Type, queueName) // Original code had this line commented out
	return nil
}

// Consumer 이벤트 소비자
type Consumer struct {
	client      *redislib.Client
	consumerID  string
	groupName   string
	isRunning   bool
	stopChan    chan struct{}
}

// NewConsumer 소비자 생성
func NewConsumer(consumerID, groupName string) *Consumer {
	return &Consumer{
		client:     redis.Client,
		consumerID: consumerID,
		groupName:  groupName,
		stopChan:   make(chan struct{}),
	}
}

// EventHandler 이벤트 핸들러 타입
type EventHandler func(event QueueEvent) error

// StartConsuming 이벤트 소비 시작
func (c *Consumer) StartConsuming(queueName string, handler EventHandler) error {
	c.isRunning = true

	// Consumer Group 생성 (이미 존재하면 무시)
	c.client.XGroupCreateMkStream(ctx, queueName, c.groupName, "0").Err()

	// log.Printf("🎧 Started consuming queue: %s with consumer: %s", queueName, c.consumerID) // Original code had this line commented out

	go func() {
		for c.isRunning {
			select {
			case <-c.stopChan:
				return
			default:
				c.processMessages(queueName, handler)
			}
		}
	}()

	return nil
}

// StopConsuming 이벤트 소비 중지
func (c *Consumer) StopConsuming() {
	c.isRunning = false
	close(c.stopChan)
}

// processMessages 메시지 처리
func (c *Consumer) processMessages(queueName string, handler EventHandler) {
	streams, err := c.client.XReadGroup(ctx, &redislib.XReadGroupArgs{
		Group:    c.groupName,
		Consumer: c.consumerID,
		Streams:  []string{queueName, ">"},
		Count:    10,
		Block:    1 * time.Second,
	}).Result()

	if err != nil {
		if err != redislib.Nil {
			// log.Printf("❌ Error reading from stream: %v", err) // Original code had this line commented out
		}
		return
	}

	for _, stream := range streams {
		for _, message := range stream.Messages {
			if err := c.handleMessage(queueName, message, handler); err != nil {
				// log.Printf("❌ Error handling message: %v", err) // Original code had this line commented out
			}
		}
	}
}

// handleMessage 개별 메시지 처리
func (c *Consumer) handleMessage(queueName string, message redislib.XMessage, handler EventHandler) error {
	eventData, exists := message.Values["event"]
	if !exists {
		return fmt.Errorf("no event data in message")
	}

	var event QueueEvent
	if err := json.Unmarshal([]byte(eventData.(string)), &event); err != nil {
		return fmt.Errorf("failed to unmarshal event: %v", err)
	}

	// 이벤트 처리
	if err := handler(event); err != nil {
		// log.Printf("❌ Handler error for event %s: %v", event.ID, err) // Original code had this line commented out

		// 재시도 로직 (3회까지)
		if event.Retry < 3 {
			event.Retry++
			return c.retryEvent(queueName, event)
		}

		// 실패한 이벤트는 별도 큐로 이동
		return c.moveToDeadLetterQueue(queueName, event)
	}

	// 성공적으로 처리된 메시지 확인
	return c.client.XAck(ctx, queueName, c.groupName, message.ID).Err()
}

// retryEvent 이벤트 재시도
func (c *Consumer) retryEvent(queueName string, event QueueEvent) error {
	retryQueue := fmt.Sprintf("%s:retry", queueName)
	return c.client.XAdd(ctx, &redislib.XAddArgs{
		Stream: retryQueue,
		Values: map[string]interface{}{
			"event": event,
		},
	}).Err()
}

// moveToDeadLetterQueue 실패한 이벤트를 데드레터 큐로 이동
func (c *Consumer) moveToDeadLetterQueue(queueName string, event QueueEvent) error {
	dlqName := fmt.Sprintf("%s:dlq", queueName)
	return c.client.XAdd(ctx, &redislib.XAddArgs{
		Stream: dlqName,
		Values: map[string]interface{}{
			"event":      event,
			"failed_at":  time.Now().Unix(),
			"queue_name": queueName,
		},
	}).Err()
}

// GetQueueStats 큐 통계 조회
func GetQueueStats(queueName string) (map[string]interface{}, error) {
	client := redis.Client

	info, err := client.XInfoStream(ctx, queueName).Result()
	if err != nil {
		return nil, err
	}

	stats := map[string]interface{}{
		"length":         info.Length,
		"consumers":      info.Groups,
		"last_entry_id":  info.LastGeneratedID,
	}

	return stats, nil
}

// PurgeQueue 큐 정리 (오래된 메시지 삭제)
func PurgeQueue(queueName string, maxAge time.Duration) error {
	client := redis.Client

	// maxAge보다 오래된 메시지들의 ID 계산
	cutoffTime := time.Now().Add(-maxAge).UnixMilli()
	cutoffID := fmt.Sprintf("%d-0", cutoffTime)

	// 오래된 메시지들 삭제
	return client.XTrimMinID(ctx, queueName, cutoffID).Err()
}

// HealthCheck 큐 시스템 상태 확인
func HealthCheck() map[string]interface{} {
	queues := []string{QueueTrades, QueuePrices, QueueMarketMake, QueueNotify, QueueAnalytics}
	health := make(map[string]interface{})

	for _, queue := range queues {
		stats, err := GetQueueStats(queue)
		if err != nil {
			health[queue] = map[string]interface{}{
				"status": "error",
				"error":  err.Error(),
			}
		} else {
			health[queue] = map[string]interface{}{
				"status": "healthy",
				"stats":  stats,
			}
		}
	}

	return health
}

// SetWithExpiry Redis에 키-값을 만료시간과 함께 저장
func SetWithExpiry(key, value string, expiry time.Duration) error {
	client := redis.GetClient()
	if client == nil {
		return fmt.Errorf("redis client is not available")
	}

	return client.Set(ctx, key, value, expiry).Err()
}

// Get Redis에서 값 조회
func Get(key string) (string, error) {
	client := redis.GetClient()
	if client == nil {
		return "", fmt.Errorf("redis client is not available")
	}

	result := client.Get(ctx, key)
	if result.Err() == redislib.Nil {
		return "", fmt.Errorf("key not found")
	}

	return result.Val(), result.Err()
}

// Delete Redis에서 키 삭제
func Delete(key string) error {
	client := redis.GetClient()
	if client == nil {
		return fmt.Errorf("redis client is not available")
	}

	return client.Del(ctx, key).Err()
}

// PublishJob Redis Stream에 작업을 발행
func PublishJob(queueName string, job map[string]interface{}) error {
	client := redis.GetClient()
	if client == nil {
		return fmt.Errorf("redis client is not available")
	}

	// job을 JSON으로 직렬화
	jobData, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("failed to marshal job: %w", err)
	}

	// Redis Stream에 메시지 추가
	args := &redislib.XAddArgs{
		Stream: queueName,
		Values: map[string]interface{}{
			"job_data": string(jobData),
			"created_at": time.Now().Unix(),
		},
	}

	_, err = client.XAdd(ctx, args).Result()
	if err != nil {
		return fmt.Errorf("failed to publish job to %s: %w", queueName, err)
	}

	return nil
}

// ConsumeJobs Redis Stream에서 작업을 소비 (워커용)
func ConsumeJobs(queueName, consumerGroup, consumerName string, handler func(map[string]interface{}) error) error {
	client := redis.GetClient()
	if client == nil {
		return fmt.Errorf("redis client is not available")
	}

	// 컨슈머 그룹 생성 (이미 존재하면 무시)
	_, err := client.XGroupCreateMkStream(ctx, queueName, consumerGroup, "0").Result()
	if err != nil && err.Error() != "BUSYGROUP Consumer Group name already exists" {
		return fmt.Errorf("failed to create consumer group: %w", err)
	}

	for {
		// 새로운 메시지 읽기
		msgs, err := client.XReadGroup(ctx, &redislib.XReadGroupArgs{
			Group:    consumerGroup,
			Consumer: consumerName,
			Streams:  []string{queueName, ">"},
			Count:    1,
			Block:    time.Second * 5, // 5초 블록
		}).Result()

		if err != nil {
			if err == redislib.Nil {
				continue // 타임아웃, 다시 시도
			}
			return fmt.Errorf("failed to read from stream: %w", err)
		}

		// 메시지 처리
		for _, stream := range msgs {
			for _, msg := range stream.Messages {
				jobDataStr, ok := msg.Values["job_data"].(string)
				if !ok {
					continue
				}

				var jobData map[string]interface{}
				if err := json.Unmarshal([]byte(jobDataStr), &jobData); err != nil {
					continue
				}

				// 핸들러 실행
				if err := handler(jobData); err != nil {
					// 처리 실패 시 로그만 출력하고 계속
					fmt.Printf("Failed to process job %s: %v\n", msg.ID, err)
				}

				// 메시지 ACK
				client.XAck(ctx, queueName, consumerGroup, msg.ID)
			}
		}
	}
}

// GetQueueLength 큐의 길이 조회 (모니터링용)
func GetQueueLength(queueName string) (int64, error) {
	client := redis.GetClient()
	if client == nil {
		return 0, fmt.Errorf("redis client is not available")
	}

	length, err := client.XLen(ctx, queueName).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get queue length: %w", err)
	}

	return length, nil
}

// ConsumeJobsWithContext Redis Stream에서 작업을 소비 (context 지원)
func ConsumeJobsWithContext(ctx context.Context, queueName, consumerGroup, consumerName string, handler func(map[string]interface{}) error) error {
	client := redis.GetClient()
	if client == nil {
		return fmt.Errorf("redis client is not available")
	}

	// 컨슈머 그룹 생성 (이미 존재하면 무시)
	_, err := client.XGroupCreateMkStream(ctx, queueName, consumerGroup, "0").Result()
	if err != nil && err.Error() != "BUSYGROUP Consumer Group name already exists" {
		return fmt.Errorf("failed to create consumer group: %w", err)
	}

	for {
		// Context 취소 확인
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// 새로운 메시지 읽기
		msgs, err := client.XReadGroup(ctx, &redislib.XReadGroupArgs{
			Group:    consumerGroup,
			Consumer: consumerName,
			Streams:  []string{queueName, ">"},
			Count:    1,
			Block:    time.Second * 5, // 5초 블록
		}).Result()

		if err != nil {
			// Context가 취소된 경우
			if err == context.Canceled {
				return nil
			}
			if err == redislib.Nil {
				continue // 타임아웃, 다시 시도
			}
			return fmt.Errorf("failed to read from stream: %w", err)
		}

		// 메시지 처리
		for _, stream := range msgs {
			for _, msg := range stream.Messages {
				jobDataStr, ok := msg.Values["job_data"].(string)
				if !ok {
					continue
				}

				var jobData map[string]interface{}
				if err := json.Unmarshal([]byte(jobDataStr), &jobData); err != nil {
					continue
				}

				// 핸들러 실행
				if err := handler(jobData); err != nil {
					// 처리 실패 시 로그만 출력하고 계속
					fmt.Printf("Failed to process job %s: %v\n", msg.ID, err)
				}

				// 메시지 ACK
				client.XAck(ctx, queueName, consumerGroup, msg.ID)
			}
		}
	}
}
