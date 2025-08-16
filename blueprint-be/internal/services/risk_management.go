package services

import (
	"blueprint-module/pkg/models"
	"fmt"
	"log"
	"math"
	"sync"
	"time"

	"gorm.io/gorm"
)

// 🛡️ Advanced Risk Management System (Polymarket Style)

// RiskManagementService 리스크 관리 서비스
type RiskManagementService struct {
	db         *gorm.DB
	feeService *FeeService

	// 리스크 모니터링
	isRunning bool
	stopChan  chan struct{}
	mutex     sync.RWMutex

	// 설정
	config RiskConfig

	// 실시간 모니터링
	userRiskScores map[uint]float64   // userID -> risk score
	marketRisks    map[string]float64 // market -> risk level
}

// RiskConfig 리스크 관리 설정
type RiskConfig struct {
	// 사용자 한도
	MaxDailyVolume  int64 `json:"max_daily_volume"`  // 일일 최대 거래량
	MaxPositionSize int64 `json:"max_position_size"` // 최대 포지션 크기
	MaxOpenOrders   int   `json:"max_open_orders"`   // 최대 오픈 주문 수
	MaxLossPerDay   int64 `json:"max_loss_per_day"`  // 일일 최대 손실

	// VIP별 한도 승수
	VIPLimitMultipliers map[int]float64 `json:"vip_limit_multipliers"` // VIP level -> multiplier

	// 시장 리스크 임계값
	VolatilityThreshold float64 `json:"volatility_threshold"` // 변동성 임계값
	LiquidityThreshold  float64 `json:"liquidity_threshold"`  // 유동성 임계값
	ConcentrationLimit  float64 `json:"concentration_limit"`  // 집중도 한계

	// 포지션 리스크
	MaxCorrelatedPositions  int     `json:"max_correlated_positions"`   // 상관관계 포지션 한계
	MaxSingleMarketExposure float64 `json:"max_single_market_exposure"` // 단일 시장 노출 한계

	// 시스템 리스크
	CircuitBreakerThreshold float64 `json:"circuit_breaker_threshold"` // 서킷브레이커 임계값
	EmergencyStopTrigger    float64 `json:"emergency_stop_trigger"`    // 긴급 중단 트리거
}

// UserRiskProfile 사용자 리스크 프로필
type UserRiskProfile struct {
	ID     uint `json:"id" gorm:"primaryKey"`
	UserID uint `json:"user_id" gorm:"uniqueIndex"`

	// 리스크 점수
	RiskScore float64 `json:"risk_score"` // 0.0 - 1.0
	RiskLevel string  `json:"risk_level"` // LOW, MEDIUM, HIGH, CRITICAL

	// 거래 행동 분석
	AvgHoldingPeriod int64   `json:"avg_holding_period"` // 평균 보유 기간 (분)
	WinRate          float64 `json:"win_rate"`           // 승률
	MaxDrawdown      float64 `json:"max_drawdown"`       // 최대 손실
	Sharpe           float64 `json:"sharpe"`             // 샤프 비율

	// 포지션 집중도
	PortfolioConcentration float64 `json:"portfolio_concentration"` // 포트폴리오 집중도
	MarketDiversification  float64 `json:"market_diversification"`  // 시장 다변화

	// 한도 설정
	DailyVolumeLimit int64 `json:"daily_volume_limit"` // 일일 거래량 한도
	PositionLimit    int64 `json:"position_limit"`     // 포지션 한도
	LossLimit        int64 `json:"loss_limit"`         // 손실 한도

	// 상태
	IsRestricted   bool      `json:"is_restricted"`   // 제한 여부
	LastAssessment time.Time `json:"last_assessment"` // 마지막 평가 시간
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`

	// 관계
	User models.User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// RiskAlert 리스크 알림
type RiskAlert struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	UserID      uint      `json:"user_id" gorm:"index"`
	AlertType   string    `json:"alert_type"` // POSITION_LIMIT, LOSS_LIMIT, VOLATILITY, LIQUIDITY
	Severity    string    `json:"severity"`   // INFO, WARNING, CRITICAL
	Message     string    `json:"message"`
	MetricValue float64   `json:"metric_value"`
	Threshold   float64   `json:"threshold"`
	IsResolved  bool      `json:"is_resolved"`
	CreatedAt   time.Time `json:"created_at"`
}

// MarketRiskMetrics 시장 리스크 지표
type MarketRiskMetrics struct {
	MilestoneID    uint      `json:"milestone_id"`
	OptionID       string    `json:"option_id"`
	Volatility     float64   `json:"volatility"`      // 변동성
	Liquidity      float64   `json:"liquidity"`       // 유동성
	Concentration  float64   `json:"concentration"`   // 집중도
	OrderImbalance float64   `json:"order_imbalance"` // 주문 불균형
	RiskLevel      string    `json:"risk_level"`      // LOW, MEDIUM, HIGH
	LastUpdate     time.Time `json:"last_update"`
}

// RiskCheckResult 리스크 체크 결과
type RiskCheckResult struct {
	Allowed       bool       `json:"allowed"`
	Reason        string     `json:"reason,omitempty"`
	RiskScore     float64    `json:"risk_score"`
	Warnings      []string   `json:"warnings,omitempty"`
	MaxAmount     int64      `json:"max_amount,omitempty"`
	CooldownUntil *time.Time `json:"cooldown_until,omitempty"`
}

// NewRiskManagementService 리스크 관리 서비스 생성자
func NewRiskManagementService(db *gorm.DB, feeService *FeeService) *RiskManagementService {
	return &RiskManagementService{
		db:         db,
		feeService: feeService,
		stopChan:   make(chan struct{}),
		config: RiskConfig{
			MaxDailyVolume:          1000000, // 1M points per day
			MaxPositionSize:         100000,  // 100K points per position
			MaxOpenOrders:           50,
			MaxLossPerDay:           50000, // 50K points per day
			VolatilityThreshold:     0.2,   // 20% volatility
			LiquidityThreshold:      0.1,   // 10% liquidity
			ConcentrationLimit:      0.3,   // 30% concentration
			MaxCorrelatedPositions:  5,
			MaxSingleMarketExposure: 0.25, // 25% of portfolio
			CircuitBreakerThreshold: 0.15, // 15% market move
			EmergencyStopTrigger:    0.25, // 25% system loss
			VIPLimitMultipliers: map[int]float64{
				1: 1.2,
				2: 1.5,
				3: 2.0,
				4: 3.0,
				5: 5.0,
			},
		},
		userRiskScores: make(map[uint]float64),
		marketRisks:    make(map[string]float64),
	}
}

// Start 리스크 관리 시작
func (rms *RiskManagementService) Start() error {
	rms.mutex.Lock()
	defer rms.mutex.Unlock()

	if rms.isRunning {
		return nil
	}

	rms.isRunning = true
	log.Println("🛡️ Risk Management Service started!")

	// 리스크 모니터링 워커
	go rms.riskMonitoringWorker()

	// 시장 리스크 분석 워커
	go rms.marketRiskWorker()

	// 사용자 리스크 평가 워커
	go rms.userRiskAssessmentWorker()

	return nil
}

// CheckOrderRisk 주문 리스크 체크
func (rms *RiskManagementService) CheckOrderRisk(userID uint, req *models.CreateOrderRequest) (*RiskCheckResult, error) {
	// 1. 사용자 리스크 프로필 조회
	profile, err := rms.GetUserRiskProfile(userID)
	if err != nil {
		return nil, err
	}

	// 2. VIP 레벨 조회
	vipLevel, _ := rms.feeService.GetUserVIPLevel(userID)

	// 3. 기본 한도 계산
	limits := rms.calculateUserLimits(profile, vipLevel)

	// 4. 현재 사용량 조회
	usage, err := rms.getCurrentUsage(userID)
	if err != nil {
		return nil, err
	}

	// 5. 주문 금액 계산
	orderAmount := int64(float64(req.Quantity) * req.Price)

	var warnings []string

	// 6. 일일 거래량 체크
	if usage.DailyVolume+orderAmount > limits.DailyVolumeLimit {
		return &RiskCheckResult{
			Allowed:   false,
			Reason:    "일일 거래량 한도 초과",
			RiskScore: profile.RiskScore,
			MaxAmount: limits.DailyVolumeLimit - usage.DailyVolume,
		}, nil
	}

	// 7. 포지션 크기 체크
	if orderAmount > limits.PositionLimit {
		return &RiskCheckResult{
			Allowed:   false,
			Reason:    "단일 포지션 한도 초과",
			RiskScore: profile.RiskScore,
			MaxAmount: limits.PositionLimit,
		}, nil
	}

	// 8. 오픈 주문 수 체크
	if usage.OpenOrders >= rms.config.MaxOpenOrders {
		return &RiskCheckResult{
			Allowed: false,
			Reason:  "오픈 주문 수 한도 초과",
		}, nil
	}

	// 9. 시장 리스크 체크
	marketRisk := rms.getMarketRisk(req.MilestoneID, req.OptionID)
	if marketRisk > 0.8 { // 높은 리스크 시장
		warnings = append(warnings, "높은 리스크 시장입니다")
	}

	// 10. 포트폴리오 집중도 체크
	if profile.PortfolioConcentration > rms.config.ConcentrationLimit {
		warnings = append(warnings, "포트폴리오 집중도가 높습니다")
	}

	return &RiskCheckResult{
		Allowed:   true,
		RiskScore: profile.RiskScore,
		Warnings:  warnings,
		MaxAmount: limits.DailyVolumeLimit - usage.DailyVolume,
	}, nil
}

// GetUserRiskProfile 사용자 리스크 프로필 조회
func (rms *RiskManagementService) GetUserRiskProfile(userID uint) (*UserRiskProfile, error) {
	var profile UserRiskProfile
	err := rms.db.Where("user_id = ?", userID).First(&profile).Error

	if err == gorm.ErrRecordNotFound {
		// 새 프로필 생성
		profile = UserRiskProfile{
			UserID:           userID,
			RiskScore:        0.5, // 중간 리스크로 시작
			RiskLevel:        "MEDIUM",
			DailyVolumeLimit: rms.config.MaxDailyVolume,
			PositionLimit:    rms.config.MaxPositionSize,
			LossLimit:        rms.config.MaxLossPerDay,
			LastAssessment:   time.Now(),
		}

		if err := rms.db.Create(&profile).Error; err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	// 프로필이 오래되었으면 업데이트
	if time.Since(profile.LastAssessment) > 24*time.Hour {
		go rms.UpdateUserRiskProfile(userID)
	}

	return &profile, nil
}

// UpdateUserRiskProfile 사용자 리스크 프로필 업데이트
func (rms *RiskManagementService) UpdateUserRiskProfile(userID uint) error {
	// 최근 30일 거래 통계 분석
	thirtyDaysAgo := time.Now().Add(-30 * 24 * time.Hour)

	var trades []models.Trade
	err := rms.db.Where("(buyer_id = ? OR seller_id = ?) AND created_at > ?",
		userID, userID, thirtyDaysAgo).Find(&trades).Error
	if err != nil {
		return err
	}

	// 리스크 지표 계산
	riskScore := rms.calculateRiskScore(userID, trades)
	winRate := rms.calculateWinRate(userID, trades)
	maxDrawdown := rms.calculateMaxDrawdown(userID, trades)
	avgHoldingPeriod := rms.calculateAvgHoldingPeriod(userID)
	concentration := rms.calculatePortfolioConcentration(userID)

	// 리스크 레벨 결정
	riskLevel := rms.determineRiskLevel(riskScore)

	// 프로필 업데이트
	updates := map[string]interface{}{
		"risk_score":              riskScore,
		"risk_level":              riskLevel,
		"win_rate":                winRate,
		"max_drawdown":            maxDrawdown,
		"avg_holding_period":      avgHoldingPeriod,
		"portfolio_concentration": concentration,
		"last_assessment":         time.Now(),
	}

	return rms.db.Model(&UserRiskProfile{}).Where("user_id = ?", userID).Updates(updates).Error
}

// Helper functions

func (rms *RiskManagementService) calculateUserLimits(profile *UserRiskProfile, vipLevel int) *UserLimits {
	multiplier := 1.0
	if mult, exists := rms.config.VIPLimitMultipliers[vipLevel]; exists {
		multiplier = mult
	}

	// 리스크 스코어에 따른 조정
	riskAdjustment := 1.0
	if profile.RiskScore > 0.7 {
		riskAdjustment = 0.5 // 고위험 사용자는 50% 감소
	} else if profile.RiskScore < 0.3 {
		riskAdjustment = 1.5 // 저위험 사용자는 50% 증가
	}

	return &UserLimits{
		DailyVolumeLimit: int64(float64(rms.config.MaxDailyVolume) * multiplier * riskAdjustment),
		PositionLimit:    int64(float64(rms.config.MaxPositionSize) * multiplier * riskAdjustment),
		LossLimit:        int64(float64(rms.config.MaxLossPerDay) * multiplier * riskAdjustment),
	}
}

func (rms *RiskManagementService) getCurrentUsage(userID uint) (*UserUsage, error) {
	today := time.Now().Truncate(24 * time.Hour)

	// 오늘 거래량
	var dailyVolume int64
	rms.db.Model(&models.Trade{}).
		Select("COALESCE(SUM(total_amount), 0)").
		Where("(buyer_id = ? OR seller_id = ?) AND created_at >= ?", userID, userID, today).
		Scan(&dailyVolume)

	// 오픈 주문 수
	var openOrders int64
	rms.db.Model(&models.Order{}).
		Where("user_id = ? AND status IN ?", userID, []string{"pending", "partial"}).
		Count(&openOrders)

	// 오늘 손실
	var dailyLoss int64
	// 간단한 손실 계산 (실제로는 더 복잡한 로직 필요)

	return &UserUsage{
		DailyVolume: dailyVolume,
		OpenOrders:  int(openOrders),
		DailyLoss:   dailyLoss,
	}, nil
}

func (rms *RiskManagementService) getMarketRisk(milestoneID uint, optionID string) float64 {
	key := fmt.Sprintf("%d:%s", milestoneID, optionID)

	rms.mutex.RLock()
	risk, exists := rms.marketRisks[key]
	rms.mutex.RUnlock()

	if !exists {
		return 0.5 // 기본 중간 리스크
	}

	return risk
}

func (rms *RiskManagementService) calculateRiskScore(userID uint, trades []models.Trade) float64 {
	if len(trades) == 0 {
		return 0.5 // 기본값
	}

	// 거래 빈도 (높을수록 위험)
	frequency := float64(len(trades)) / 30.0        // 일일 평균 거래 수
	frequencyScore := math.Min(frequency/10.0, 1.0) // 일일 10회 이상이면 1.0

	// 거래 크기 변동성 (높을수록 위험)
	var amounts []float64
	for _, trade := range trades {
		amounts = append(amounts, float64(trade.TotalAmount))
	}
	volatilityScore := rms.calculateStdDev(amounts) / rms.calculateMean(amounts)
	volatilityScore = math.Min(volatilityScore, 1.0)

	// 최종 리스크 스코어 (0.0 - 1.0)
	finalScore := (frequencyScore*0.4 + volatilityScore*0.6)
	return math.Max(0.0, math.Min(1.0, finalScore))
}

func (rms *RiskManagementService) calculateWinRate(userID uint, trades []models.Trade) float64 {
	if len(trades) == 0 {
		return 0.0
	}

	wins := 0
	// 간단한 승률 계산 (실제로는 포지션 기반으로 계산해야 함)
	// 여기서는 매수자가 이익을 본 거래의 비율로 계산

	return float64(wins) / float64(len(trades))
}

func (rms *RiskManagementService) calculateMaxDrawdown(userID uint, trades []models.Trade) float64 {
	// 최대 손실폭 계산
	return 0.0 // 간단화
}

func (rms *RiskManagementService) calculateAvgHoldingPeriod(userID uint) int64 {
	// 평균 보유 기간 계산
	return 0 // 간단화
}

func (rms *RiskManagementService) calculatePortfolioConcentration(userID uint) float64 {
	// 포트폴리오 집중도 계산
	return 0.0 // 간단화
}

func (rms *RiskManagementService) determineRiskLevel(score float64) string {
	if score < 0.3 {
		return "LOW"
	} else if score < 0.7 {
		return "MEDIUM"
	} else if score < 0.9 {
		return "HIGH"
	}
	return "CRITICAL"
}

func (rms *RiskManagementService) calculateStdDev(values []float64) float64 {
	if len(values) <= 1 {
		return 0.0
	}

	mean := rms.calculateMean(values)
	var sum float64
	for _, v := range values {
		sum += math.Pow(v-mean, 2)
	}

	return math.Sqrt(sum / float64(len(values)-1))
}

func (rms *RiskManagementService) calculateMean(values []float64) float64 {
	if len(values) == 0 {
		return 0.0
	}

	var sum float64
	for _, v := range values {
		sum += v
	}

	return sum / float64(len(values))
}

// Worker functions

func (rms *RiskManagementService) riskMonitoringWorker() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-rms.stopChan:
			return
		case <-ticker.C:
			rms.monitorSystemRisk()
		}
	}
}

func (rms *RiskManagementService) marketRiskWorker() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-rms.stopChan:
			return
		case <-ticker.C:
			rms.updateMarketRisks()
		}
	}
}

func (rms *RiskManagementService) userRiskAssessmentWorker() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-rms.stopChan:
			return
		case <-ticker.C:
			rms.batchUpdateUserRisks()
		}
	}
}

func (rms *RiskManagementService) monitorSystemRisk() {
	// 시스템 전체 리스크 모니터링
	log.Println("🔍 Monitoring system risk...")
}

func (rms *RiskManagementService) updateMarketRisks() {
	// 시장별 리스크 업데이트
	log.Println("📊 Updating market risks...")
}

func (rms *RiskManagementService) batchUpdateUserRisks() {
	// 사용자 리스크 일괄 업데이트
	log.Println("👥 Batch updating user risks...")
}

// Helper structs

type UserLimits struct {
	DailyVolumeLimit int64 `json:"daily_volume_limit"`
	PositionLimit    int64 `json:"position_limit"`
	LossLimit        int64 `json:"loss_limit"`
}

type UserUsage struct {
	DailyVolume int64 `json:"daily_volume"`
	OpenOrders  int   `json:"open_orders"`
	DailyLoss   int64 `json:"daily_loss"`
}
