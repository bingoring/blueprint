package services

import (
	"blueprint-module/pkg/queue"
	"blueprint-module/pkg/redis"
	"blueprint/internal/models"
	"context"
	"fmt"
	"log"
	"math"
	"sync"
	"time"

	"gorm.io/gorm"
)

// 💎 Liquidity Mining Program (Polymarket Style)

// LiquidityMiningService 유동성 마이닝 서비스
type LiquidityMiningService struct {
	db             *gorm.DB
	queuePublisher *queue.Publisher

	// 마이닝 상태
	isRunning      bool
	stopChan       chan struct{}
	mutex          sync.RWMutex

	// 설정
	config         LiquidityMiningConfig

	// 통계
	stats          LiquidityMiningStats
}

// LiquidityMiningConfig 유동성 마이닝 설정
type LiquidityMiningConfig struct {
	// 리워드 설정
	DailyRewardPool     int64   `json:"daily_reward_pool"`     // 일일 리워드 풀 (tokens)
	MinLiquidityAmount  int64   `json:"min_liquidity_amount"`  // 최소 유동성 제공량
	RewardCalculationInterval time.Duration `json:"reward_calculation_interval"` // 리워드 계산 주기

	// 부스터 설정
	EarlyProviderBonus  float64 `json:"early_provider_bonus"`  // 초기 유동성 제공자 보너스
	LongTermBonus       float64 `json:"long_term_bonus"`       // 장기 제공자 보너스 (30일+)
	VIPBonus            float64 `json:"vip_bonus"`             // VIP 사용자 보너스

	// 마켓별 승수
	MarketMultipliers   map[string]float64 `json:"market_multipliers"` // 특정 마켓 승수

	// 이벤트 기간 설정
	EventMultiplier     float64 `json:"event_multiplier"`      // 이벤트 기간 승수
	EventStartTime      time.Time `json:"event_start_time"`    // 이벤트 시작 시간
	EventEndTime        time.Time `json:"event_end_time"`      // 이벤트 종료 시간
}

// LiquidityProvider 유동성 제공자 정보
type LiquidityProvider struct {
	ID             uint      `json:"id" gorm:"primaryKey"`
	UserID         uint      `json:"user_id" gorm:"index"`
	MilestoneID    uint      `json:"milestone_id" gorm:"index"`
	OptionID       string    `json:"option_id" gorm:"index"`

	// 유동성 정보
	BidQuantity    int64     `json:"bid_quantity"`     // 매수 유동성
	AskQuantity    int64     `json:"ask_quantity"`     // 매도 유동성
	TotalLiquidity int64     `json:"total_liquidity"`  // 총 유동성
	AvgSpread      float64   `json:"avg_spread"`       // 평균 스프레드

	// 시간 정보
	StartTime      time.Time `json:"start_time"`       // 제공 시작 시간
	LastActive     time.Time `json:"last_active"`      // 마지막 활동 시간
	Duration       int64     `json:"duration"`         // 제공 지속 시간 (분)

	// 리워드 정보
	EarnedRewards  int64     `json:"earned_rewards"`   // 획득한 리워드
	PendingRewards int64     `json:"pending_rewards"`  // 대기 중인 리워드
	LastClaimTime  time.Time `json:"last_claim_time"`  // 마지막 청구 시간

	// 부스터 정보
	EarlyBonus     float64   `json:"early_bonus"`      // 초기 제공자 보너스
	LongTermBonus  float64   `json:"long_term_bonus"`  // 장기 제공자 보너스
	VIPLevel       int       `json:"vip_level"`        // VIP 레벨

	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`

	// 관계
	User           models.User      `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Milestone      models.Milestone `json:"milestone,omitempty" gorm:"foreignKey:MilestoneID"`
}

// LiquidityReward 유동성 리워드 기록
type LiquidityReward struct {
	ID               uint      `json:"id" gorm:"primaryKey"`
	UserID           uint      `json:"user_id" gorm:"index"`
	MilestoneID      uint      `json:"milestone_id"`
	OptionID         string    `json:"option_id"`

	// 리워드 정보
	RewardAmount     int64     `json:"reward_amount"`     // 리워드 금액
	LiquidityScore   float64   `json:"liquidity_score"`   // 유동성 점수
	TimeWeight       float64   `json:"time_weight"`       // 시간 가중치
	MarketShare      float64   `json:"market_share"`      // 시장 점유율

	// 부스터 적용
	BaseReward       int64     `json:"base_reward"`       // 기본 리워드
	BonusReward      int64     `json:"bonus_reward"`      // 보너스 리워드
	TotalMultiplier  float64   `json:"total_multiplier"`  // 총 승수

	// 기간 정보
	PeriodStart      time.Time `json:"period_start"`      // 리워드 기간 시작
	PeriodEnd        time.Time `json:"period_end"`        // 리워드 기간 종료

	// 상태
	Status           string    `json:"status"`            // pending, claimed, expired
	ClaimedAt        *time.Time `json:"claimed_at"`       // 청구 시간

	CreatedAt        time.Time `json:"created_at"`
}

// LiquidityMiningStats 유동성 마이닝 통계
type LiquidityMiningStats struct {
	TotalProviders       int     `json:"total_providers"`        // 총 제공자 수
	TotalLiquidity       int64   `json:"total_liquidity"`        // 총 유동성
	TotalRewardsDistributed int64 `json:"total_rewards_distributed"` // 총 배분된 리워드
	AverageAPY           float64 `json:"average_apy"`             // 평균 연수익률
	TopMarkets           []MarketLiquidityInfo `json:"top_markets"` // 상위 마켓들
	ActivePools          int     `json:"active_pools"`           // 활성 풀 수
	DailyVolume          int64   `json:"daily_volume"`           // 일일 거래량
}

// MarketLiquidityInfo 마켓별 유동성 정보
type MarketLiquidityInfo struct {
	MilestoneID    uint    `json:"milestone_id"`
	OptionID       string  `json:"option_id"`
	TotalLiquidity int64   `json:"total_liquidity"`
	Providers      int     `json:"providers"`
	APY            float64 `json:"apy"`
	Volume24h      int64   `json:"volume_24h"`
}

// NewLiquidityMiningService 유동성 마이닝 서비스 생성자
func NewLiquidityMiningService(db *gorm.DB) *LiquidityMiningService {
	return &LiquidityMiningService{
		db:             db,
		queuePublisher: queue.NewPublisher(),
		stopChan:       make(chan struct{}),
		config: LiquidityMiningConfig{
			DailyRewardPool:           100000, // 100,000 tokens per day
			MinLiquidityAmount:        1000,   // 최소 1,000 points
			RewardCalculationInterval: 1 * time.Hour, // 1시간마다 계산
			EarlyProviderBonus:        0.5,    // 50% 보너스
			LongTermBonus:             0.3,    // 30% 보너스
			VIPBonus:                  0.2,    // 20% 보너스
			MarketMultipliers:         make(map[string]float64),
			EventMultiplier:           2.0,    // 이벤트 기간 2배
		},
		stats: LiquidityMiningStats{},
	}
}

// Start 유동성 마이닝 시작
func (lms *LiquidityMiningService) Start() error {
	lms.mutex.Lock()
	defer lms.mutex.Unlock()

	if lms.isRunning {
		return nil
	}

	lms.isRunning = true
	log.Println("💎 Liquidity Mining Service started!")

	// 리워드 계산 워커 시작
	go lms.rewardCalculationWorker()

	// 통계 업데이트 워커
	go lms.statsUpdateWorker()

	// 만료된 리워드 정리 워커
	go lms.cleanupWorker()

	return nil
}

// Stop 유동성 마이닝 중지
func (lms *LiquidityMiningService) Stop() error {
	lms.mutex.Lock()
	defer lms.mutex.Unlock()

	if !lms.isRunning {
		return nil
	}

	lms.isRunning = false
	close(lms.stopChan)

	log.Println("🛑 Liquidity Mining Service stopped!")
	return nil
}

// TrackLiquidityProvider 유동성 제공자 추적
func (lms *LiquidityMiningService) TrackLiquidityProvider(userID uint, milestoneID uint, optionID string, bidQuantity, askQuantity int64) error {
	provider := &LiquidityProvider{
		UserID:         userID,
		MilestoneID:    milestoneID,
		OptionID:       optionID,
		BidQuantity:    bidQuantity,
		AskQuantity:    askQuantity,
		TotalLiquidity: bidQuantity + askQuantity,
		StartTime:      time.Now(),
		LastActive:     time.Now(),
	}

	// 기존 제공자 정보가 있으면 업데이트, 없으면 생성
	var existingProvider LiquidityProvider
	err := lms.db.Where("user_id = ? AND milestone_id = ? AND option_id = ?",
		userID, milestoneID, optionID).First(&existingProvider).Error

	if err == gorm.ErrRecordNotFound {
		// 새로운 제공자
		provider.EarlyBonus = lms.calculateEarlyProviderBonus(milestoneID, optionID)
		return lms.db.Create(provider).Error
	} else if err != nil {
		return err
	} else {
		// 기존 제공자 업데이트
		updates := map[string]interface{}{
			"bid_quantity":    bidQuantity,
			"ask_quantity":    askQuantity,
			"total_liquidity": bidQuantity + askQuantity,
			"last_active":     time.Now(),
			"duration":        int64(time.Since(existingProvider.StartTime).Minutes()),
		}
		return lms.db.Model(&existingProvider).Updates(updates).Error
	}
}

// CalculateRewards 리워드 계산 및 배분
func (lms *LiquidityMiningService) CalculateRewards() error {
	log.Println("💰 Calculating liquidity mining rewards...")

	periodStart := time.Now().Add(-lms.config.RewardCalculationInterval)
	periodEnd := time.Now()

	// 활성 유동성 제공자들 조회
	var providers []LiquidityProvider
	err := lms.db.Where("last_active > ? AND total_liquidity >= ?",
		periodStart, lms.config.MinLiquidityAmount).Find(&providers).Error
	if err != nil {
		return err
	}

	if len(providers) == 0 {
		log.Println("📊 No active liquidity providers found")
		return nil
	}

	// 총 유동성 점수 계산
	totalLiquidityScore := 0.0
	providerScores := make(map[uint]float64)

	for _, provider := range providers {
		score := lms.calculateLiquidityScore(&provider)
		providerScores[provider.ID] = score
		totalLiquidityScore += score
	}

	// 일일 리워드 풀을 시간 비례로 계산
	periodRewardPool := float64(lms.config.DailyRewardPool) *
		lms.config.RewardCalculationInterval.Hours() / 24.0

	// 각 제공자에게 리워드 배분
	for _, provider := range providers {
		score := providerScores[provider.ID]
		if score <= 0 {
			continue
		}

		// 기본 리워드 계산
		baseReward := int64(periodRewardPool * score / totalLiquidityScore)

		// 부스터 적용
		multiplier := lms.calculateTotalMultiplier(&provider)
		bonusReward := int64(float64(baseReward) * (multiplier - 1.0))
		totalReward := baseReward + bonusReward

		// 리워드 기록 생성
		reward := &LiquidityReward{
			UserID:          provider.UserID,
			MilestoneID:     provider.MilestoneID,
			OptionID:        provider.OptionID,
			RewardAmount:    totalReward,
			LiquidityScore:  score,
			BaseReward:      baseReward,
			BonusReward:     bonusReward,
			TotalMultiplier: multiplier,
			PeriodStart:     periodStart,
			PeriodEnd:       periodEnd,
			Status:          "pending",
			CreatedAt:       time.Now(),
		}

		if err := lms.db.Create(reward).Error; err != nil {
			log.Printf("❌ Failed to create reward record: %v", err)
			continue
		}

		// 제공자의 대기 중인 리워드 업데이트
		lms.db.Model(&provider).Update("pending_rewards",
			provider.PendingRewards + totalReward)

		log.Printf("💎 Reward calculated for user %d: %d tokens (%.2fx multiplier)",
			provider.UserID, totalReward, multiplier)
	}

	log.Printf("✅ Reward calculation completed for %d providers", len(providers))
	return nil
}

// ClaimRewards 리워드 청구
func (lms *LiquidityMiningService) ClaimRewards(userID uint) (*ClaimResult, error) {
	// 대기 중인 리워드 조회
	var pendingRewards []LiquidityReward
	err := lms.db.Where("user_id = ? AND status = ?", userID, "pending").
		Find(&pendingRewards).Error
	if err != nil {
		return nil, err
	}

	if len(pendingRewards) == 0 {
		return &ClaimResult{
			Success: false,
			Message: "청구할 리워드가 없습니다",
		}, nil
	}

	// 총 리워드 계산
	totalReward := int64(0)
	for _, reward := range pendingRewards {
		totalReward += reward.RewardAmount
	}

	// 사용자 지갑에 BLUEPRINT 리워드 추가
	tx := lms.db.Begin()

	// 지갑 업데이트
	var wallet models.UserWallet
	err = tx.Where("user_id = ?", userID).First(&wallet).Error
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	// BLUEPRINT 토큰으로 리워드 지급
	wallet.BlueprintBalance += totalReward
	wallet.TotalBlueprintEarned += totalReward
	if err := tx.Save(&wallet).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// 리워드 상태 업데이트
	now := time.Now()
	err = tx.Model(&LiquidityReward{}).
		Where("user_id = ? AND status = ?", userID, "pending").
		Updates(map[string]interface{}{
			"status":     "claimed",
			"claimed_at": &now,
		}).Error
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	// 유동성 제공자 정보 업데이트
	tx.Model(&LiquidityProvider{}).
		Where("user_id = ?", userID).
		Updates(map[string]interface{}{
			"earned_rewards":  gorm.Expr("earned_rewards + ?", totalReward),
			"pending_rewards": 0,
			"last_claim_time": now,
		})

	tx.Commit()

	log.Printf("🎉 User %d claimed %d tokens in liquidity rewards", userID, totalReward)

	return &ClaimResult{
		Success:      true,
		Message:      fmt.Sprintf("%d 토큰을 성공적으로 청구했습니다", totalReward),
		RewardAmount: totalReward,
		ClaimedAt:    now,
	}, nil
}

// Helper functions

func (lms *LiquidityMiningService) calculateLiquidityScore(provider *LiquidityProvider) float64 {
	// 기본 점수: 유동성 * 시간 가중치
	baseScore := float64(provider.TotalLiquidity)

	// 시간 가중치 (더 오래 제공할수록 높은 점수)
	timeWeight := math.Min(1.0 + float64(provider.Duration)/1440.0, 2.0) // 최대 2배 (24시간 기준)

	// 스프레드 패널티 (스프레드가 클수록 점수 감소)
	spreadPenalty := math.Max(0.5, 1.0 - provider.AvgSpread*10) // 최소 50%

	finalScore := baseScore * timeWeight * spreadPenalty

	return finalScore
}

func (lms *LiquidityMiningService) calculateTotalMultiplier(provider *LiquidityProvider) float64 {
	multiplier := 1.0

	// 초기 제공자 보너스
	if provider.EarlyBonus > 0 {
		multiplier += provider.EarlyBonus
	}

	// 장기 제공자 보너스 (30일 이상)
	if provider.Duration >= 30*24*60 { // 30일
		multiplier += lms.config.LongTermBonus
	}

	// VIP 보너스
	if provider.VIPLevel > 0 {
		vipBonus := float64(provider.VIPLevel) * lms.config.VIPBonus / 5.0 // 최대 5레벨
		multiplier += vipBonus
	}

	// 이벤트 기간 보너스
	now := time.Now()
	if now.After(lms.config.EventStartTime) && now.Before(lms.config.EventEndTime) {
		multiplier *= lms.config.EventMultiplier
	}

	return multiplier
}

func (lms *LiquidityMiningService) calculateEarlyProviderBonus(milestoneID uint, optionID string) float64 {
	// 해당 마켓의 총 제공자 수 확인
	var providerCount int64
	lms.db.Model(&LiquidityProvider{}).
		Where("milestone_id = ? AND option_id = ?", milestoneID, optionID).
		Count(&providerCount)

	// 초기 10명까지는 보너스 제공
	if providerCount < 10 {
		return lms.config.EarlyProviderBonus * (1.0 - float64(providerCount)/10.0)
	}

	return 0.0
}

// Worker functions

func (lms *LiquidityMiningService) rewardCalculationWorker() {
	ticker := time.NewTicker(lms.config.RewardCalculationInterval)
	defer ticker.Stop()

	for {
		select {
		case <-lms.stopChan:
			return
		case <-ticker.C:
			if err := lms.CalculateRewards(); err != nil {
				log.Printf("❌ Error calculating rewards: %v", err)
			}
		}
	}
}

func (lms *LiquidityMiningService) statsUpdateWorker() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-lms.stopChan:
			return
		case <-ticker.C:
			lms.updateStats()
		}
	}
}

func (lms *LiquidityMiningService) cleanupWorker() {
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-lms.stopChan:
			return
		case <-ticker.C:
			// 30일 이상 된 만료된 리워드 삭제
			expiredTime := time.Now().Add(-30 * 24 * time.Hour)
			lms.db.Where("status = 'expired' AND created_at < ?", expiredTime).
				Delete(&LiquidityReward{})
		}
	}
}

func (lms *LiquidityMiningService) updateStats() {
	var stats LiquidityMiningStats

	var totalProviders int64
	// 총 제공자 수
	lms.db.Model(&LiquidityProvider{}).Count(&totalProviders)
	stats.TotalProviders = int(totalProviders)

	// 총 유동성
	lms.db.Model(&LiquidityProvider{}).
		Select("COALESCE(SUM(total_liquidity), 0)").
		Row().Scan(&stats.TotalLiquidity)

	// 총 배분된 리워드
	lms.db.Model(&LiquidityReward{}).
		Where("status = 'claimed'").
		Select("COALESCE(SUM(reward_amount), 0)").
		Row().Scan(&stats.TotalRewardsDistributed)

	// 통계 업데이트
	lms.mutex.Lock()
	lms.stats = stats
	lms.mutex.Unlock()

	// Redis에 캐시
	ctx := context.Background()
	redis.Client.Set(ctx, "liquidity_mining_stats", stats, 5*time.Minute)
}

// ClaimResult 청구 결과
type ClaimResult struct {
	Success      bool      `json:"success"`
	Message      string    `json:"message"`
	RewardAmount int64     `json:"reward_amount"`
	ClaimedAt    time.Time `json:"claimed_at"`
}

// GetUserLiquidityInfo 사용자 유동성 정보 조회
func (lms *LiquidityMiningService) GetUserLiquidityInfo(userID uint) (*UserLiquidityInfo, error) {
	var providers []LiquidityProvider
	err := lms.db.Where("user_id = ?", userID).Find(&providers).Error
	if err != nil {
		return nil, err
	}

	var totalLiquidity int64
	var totalEarned int64
	var totalPending int64

	for _, provider := range providers {
		totalLiquidity += provider.TotalLiquidity
		totalEarned += provider.EarnedRewards
		totalPending += provider.PendingRewards
	}

	// 예상 일일 수익 계산
	dailyEstimate := lms.estimateDailyRewards(userID, totalLiquidity)

	return &UserLiquidityInfo{
		TotalLiquidity:   totalLiquidity,
		ActiveProvisions: len(providers),
		TotalEarned:      totalEarned,
		PendingRewards:   totalPending,
		EstimatedDaily:   dailyEstimate,
		Providers:        providers,
	}, nil
}

func (lms *LiquidityMiningService) estimateDailyRewards(userID uint, liquidity int64) int64 {
	if liquidity == 0 {
		return 0
	}

	// 간단한 추정: 전체 유동성 대비 비율로 계산
	totalMarketLiquidity := lms.stats.TotalLiquidity
	if totalMarketLiquidity == 0 {
		return 0
	}

	userShare := float64(liquidity) / float64(totalMarketLiquidity)
	dailyEstimate := int64(float64(lms.config.DailyRewardPool) * userShare)

	return dailyEstimate
}

// UserLiquidityInfo 사용자 유동성 정보
type UserLiquidityInfo struct {
	TotalLiquidity   int64               `json:"total_liquidity"`
	ActiveProvisions int                 `json:"active_provisions"`
	TotalEarned      int64               `json:"total_earned"`
	PendingRewards   int64               `json:"pending_rewards"`
	EstimatedDaily   int64               `json:"estimated_daily"`
	Providers        []LiquidityProvider `json:"providers"`
}

// GetStats 통계 조회
func (lms *LiquidityMiningService) GetStats() LiquidityMiningStats {
	lms.mutex.RLock()
	defer lms.mutex.RUnlock()
	return lms.stats
}
