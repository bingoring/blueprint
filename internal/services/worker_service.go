package services

import (
	"blueprint/internal/database"
	"blueprint/internal/models"
	"blueprint/internal/queue"
	"fmt"
	"log"
	"sync"
	"time"

	"gorm.io/gorm"
)

// WorkerService 백그라운드 작업 처리 서비스
type WorkerService struct {
	db         *gorm.DB
	consumers  map[string]*queue.Consumer
	isRunning  bool
	stopChan   chan struct{}
	wg         sync.WaitGroup
	mutex      sync.RWMutex
}

// NewWorkerService 워커 서비스 생성
func NewWorkerService() *WorkerService {
	return &WorkerService{
		db:        database.GetDB(),
		consumers: make(map[string]*queue.Consumer),
		stopChan:  make(chan struct{}),
	}
}

// Start 워커 서비스 시작
func (w *WorkerService) Start() error {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	if w.isRunning {
		return fmt.Errorf("worker service is already running")
	}

	w.isRunning = true
	log.Printf("🔧 Starting Worker Service...")

	// 각 큐별로 워커 시작
	w.startQueueWorker(queue.QueueUserTasks, "user-worker", w.handleUserTasks)
	w.startQueueWorker(queue.QueueWallet, "wallet-worker", w.handleWalletTasks)
	w.startQueueWorker(queue.QueueMarket, "market-worker", w.handleMarketTasks)
	w.startQueueWorker(queue.QueueWelcome, "welcome-worker", w.handleWelcomeTasks)

	log.Printf("✅ Worker Service started with %d workers", len(w.consumers))
	return nil
}

// Stop 워커 서비스 중지
func (w *WorkerService) Stop() {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	if !w.isRunning {
		return
	}

	log.Printf("🛑 Stopping Worker Service...")
	close(w.stopChan)

	// 모든 컨슈머 중지
	for name, consumer := range w.consumers {
		log.Printf("🛑 Stopping consumer: %s", name)
		consumer.StopConsuming()
	}

	// 모든 워커 완료 대기
	w.wg.Wait()
	w.isRunning = false
	log.Printf("✅ Worker Service stopped")
}

// startQueueWorker 큐 워커 시작
func (w *WorkerService) startQueueWorker(queueName, workerName string, handler queue.EventHandler) {
	consumer := queue.NewConsumer(workerName, "blueprint-workers")
	w.consumers[workerName] = consumer

	w.wg.Add(1)
	go func() {
		defer w.wg.Done()
		if err := consumer.StartConsuming(queueName, handler); err != nil {
			log.Printf("❌ Error starting consumer %s: %v", workerName, err)
		}
	}()

	log.Printf("🔧 Started worker: %s for queue: %s", workerName, queueName)
}

// handleUserTasks 사용자 작업 처리
func (w *WorkerService) handleUserTasks(event queue.QueueEvent) error {
	switch event.Type {
	case queue.EventTypeUserCreated:
		return w.processUserCreated(event)
	default:
		return fmt.Errorf("unknown user task type: %s", event.Type)
	}
}

// handleWalletTasks 지갑 작업 처리
func (w *WorkerService) handleWalletTasks(event queue.QueueEvent) error {
	switch event.Type {
	case queue.EventTypeWalletCreate:
		return w.processWalletCreate(event)
	default:
		return fmt.Errorf("unknown wallet task type: %s", event.Type)
	}
}

// handleMarketTasks 마켓 작업 처리
func (w *WorkerService) handleMarketTasks(event queue.QueueEvent) error {
	switch event.Type {
	case queue.EventTypeMarketInit:
		return w.processMarketInit(event)
	default:
		return fmt.Errorf("unknown market task type: %s", event.Type)
	}
}

// handleWelcomeTasks 웰컴 작업 처리
func (w *WorkerService) handleWelcomeTasks(event queue.QueueEvent) error {
	switch event.Type {
	case queue.EventTypeWelcomeUser:
		return w.processWelcomeUser(event)
	default:
		return fmt.Errorf("unknown welcome task type: %s", event.Type)
	}
}

// processUserCreated 사용자 생성 후속 처리
func (w *WorkerService) processUserCreated(event queue.QueueEvent) error {
	userID := uint(event.Data["user_id"].(float64))
	email := event.Data["email"].(string)
	username := event.Data["username"].(string)
	_ = event.Data["provider"].(string) // 현재 사용하지 않지만 나중에 확장 가능

	log.Printf("🔧 Processing user created: UserID=%d, Email=%s", userID, email)

	// 1. 지갑 생성 큐에 추가
	publisher := queue.NewPublisher()
	err := publisher.EnqueueWalletCreate(queue.WalletCreateEventData{
		UserID:        userID,
		InitialAmount: 10000, // 초기 10,000 포인트
	})
	if err != nil {
		log.Printf("❌ Failed to enqueue wallet creation: %v", err)
	}

	// 2. 웰컴 처리 큐에 추가
	err = publisher.EnqueueWelcomeUser(queue.WelcomeUserEventData{
		UserID:   userID,
		Email:    email,
		Username: username,
	})
	if err != nil {
		log.Printf("❌ Failed to enqueue welcome user: %v", err)
	}

	log.Printf("✅ User created tasks queued: UserID=%d", userID)
	return nil
}

// processWalletCreate 지갑 생성 처리
func (w *WorkerService) processWalletCreate(event queue.QueueEvent) error {
	userID := uint(event.Data["user_id"].(float64))
	initialAmount := int64(event.Data["initial_amount"].(float64))

	log.Printf("🔧 Processing wallet creation: UserID=%d, Amount=%d", userID, initialAmount)

	// 기존 지갑 확인
	var existingWallet models.UserWallet
	err := w.db.Where("user_id = ?", userID).First(&existingWallet).Error
	if err == nil {
		log.Printf("⚠️ Wallet already exists for UserID=%d", userID)
		return nil // 이미 지갑이 있으면 생성하지 않음
	}

	// 새 지갑 생성 (하이브리드 시스템)
	wallet := models.UserWallet{
		UserID:                 userID,
		USDCBalance:           initialAmount,  // 초기 USDC 지급
		USDCLockedBalance:     0,
		BlueprintBalance:      1000,          // 초기 BLUEPRINT 토큰 지급
		BlueprintLockedBalance: 0,
		TotalUSDCDeposit:      initialAmount,
		TotalBlueprintEarned:  1000,          // 회원가입 보상
		WinRate:               0,
		TotalTrades:           0,
		CreatedAt:             time.Now(),
		UpdatedAt:             time.Now(),
	}

	if err := w.db.Create(&wallet).Error; err != nil {
		log.Printf("❌ Failed to create wallet for UserID=%d: %v", userID, err)
		return err
	}

	log.Printf("✅ Wallet created: UserID=%d, USDC=%d, BLUEPRINT=%d", userID, initialAmount, 1000)
	return nil
}

// processMarketInit 마켓 초기화 처리
func (w *WorkerService) processMarketInit(event queue.QueueEvent) error {
	projectID := uint(event.Data["project_id"].(float64))
	milestoneID := uint(event.Data["milestone_id"].(float64))
	options := event.Data["options"].([]interface{})

	log.Printf("🔧 Processing market init: ProjectID=%d, MilestoneID=%d", projectID, milestoneID)

	// options를 string slice로 변환
	optionStrings := make([]string, len(options))
	for i, opt := range options {
		optionStrings[i] = opt.(string)
	}

	// 🎯 폴리마켓 스타일: 확률 합계 검증
	optionCount := len(optionStrings)
	if optionCount < 2 {
		return fmt.Errorf("market must have at least 2 options")
	}

	// 각 옵션의 초기 확률은 1/N (균등 분배)
	initialPrice := 1.0 / float64(optionCount)

	// 범위 검증 (0.01-0.99)
	if initialPrice < 0.01 {
		initialPrice = 0.01
	} else if initialPrice > 0.99 {
		initialPrice = 0.99
	}

	log.Printf("🎯 Initializing market with %d options at %.4f probability each", optionCount, initialPrice)

	// 각 옵션별로 MarketData 생성
	for _, option := range optionStrings {
		var existingMarket models.MarketData
		err := w.db.Where("milestone_id = ? AND option_id = ?", milestoneID, option).First(&existingMarket).Error
		if err == nil {
			continue // 이미 존재하면 스킵
		}

		marketData := models.MarketData{
			MilestoneID:   milestoneID,
			OptionID:      option,
			CurrentPrice:  initialPrice, // 균등 확률로 초기화
			PreviousPrice: initialPrice,
			Volume24h:     0,
			Trades24h:     0,
		}

		if err := w.db.Create(&marketData).Error; err != nil {
			log.Printf("❌ Failed to create market data: MilestoneID=%d, Option=%s, Error=%v", milestoneID, option, err)
			return err
		}
	}

	// 확률 합계 검증 (로그)
	totalProbability := float64(optionCount) * initialPrice
	log.Printf("✅ Market initialized: MilestoneID=%d, Total probability=%.4f", milestoneID, totalProbability)

	log.Printf("✅ Market initialized: MilestoneID=%d, Options=%v", milestoneID, optionStrings)
	return nil
}

// processWelcomeUser 웰컴 사용자 처리
func (w *WorkerService) processWelcomeUser(event queue.QueueEvent) error {
	userID := uint(event.Data["user_id"].(float64))
	email := event.Data["email"].(string)
	username := event.Data["username"].(string)

	log.Printf("🔧 Processing welcome user: UserID=%d, Email=%s", userID, email)

	// TODO: 실제 웰컴 이메일 발송, 온보딩 데이터 생성 등
	// 여기서는 로그만 출력
	log.Printf("📧 Welcome email would be sent to: %s", email)
	log.Printf("🎉 Onboarding data would be created for: %s", username)

	log.Printf("✅ Welcome user processed: UserID=%d", userID)
	return nil
}

// GetStats 워커 통계 조회
func (w *WorkerService) GetStats() map[string]interface{} {
	w.mutex.RLock()
	defer w.mutex.RUnlock()

	stats := map[string]interface{}{
		"is_running":     w.isRunning,
		"active_workers": len(w.consumers),
		"uptime":         time.Since(time.Now()), // TODO: 실제 시작 시간 추적
	}

	// 각 큐별 통계
	queueStats := make(map[string]interface{})
	for queueName := range w.consumers {
		if qStats, err := queue.GetQueueStats(queueName); err == nil {
			queueStats[queueName] = qStats
		}
	}
	stats["queue_stats"] = queueStats

	return stats
}
