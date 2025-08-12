package services

import (
	"blueprint-module/pkg/redis"
	"blueprint/internal/models"
	"context"
	"fmt"
	"math"
	"time"

	"gorm.io/gorm"
)

// 🎯 Dynamic Fee System (Polymarket Style)

// FeeService 동적 수수료 서비스
type FeeService struct {
	db *gorm.DB
}

// FeeConfig 수수료 설정
type FeeConfig struct {
	// 기본 수수료율
	BaseMakerFee float64 `json:"base_maker_fee"` // 0.1% (매이커)
	BaseTakerFee float64 `json:"base_taker_fee"` // 0.2% (테이커)

	// VIP 단계별 할인
	VIPTiers []VIPTier `json:"vip_tiers"`

	// 거래량 기반 할인
	VolumeDiscounts []VolumeDiscount `json:"volume_discounts"`

	// 시장 조건 기반 조정
	LiquidityMultiplier float64 `json:"liquidity_multiplier"` // 유동성 부족시 할증
	VolatilityMultiplier float64 `json:"volatility_multiplier"` // 변동성 높을 때 할증

	// 최소/최대 수수료
	MinFee float64 `json:"min_fee"` // 0.05%
	MaxFee float64 `json:"max_fee"` // 1.0%
}

// VIPTier VIP 등급별 혜택
type VIPTier struct {
	Level           int     `json:"level"`           // 1-10
	MinVolume30D    int64   `json:"min_volume_30d"`  // 30일 거래량 조건
	MinTrades30D    int     `json:"min_trades_30d"`  // 30일 거래 횟수 조건
	MakerDiscount   float64 `json:"maker_discount"`  // 매이커 할인율
	TakerDiscount   float64 `json:"taker_discount"`  // 테이커 할인율
	SpecialBenefits string  `json:"special_benefits"` // 특별 혜택
}

// VolumeDiscount 거래량 할인
type VolumeDiscount struct {
	MinVolume24H int64   `json:"min_volume_24h"` // 24시간 거래량 조건
	Discount     float64 `json:"discount"`       // 할인율
}

// FeeCalculation 수수료 계산 결과
type FeeCalculation struct {
	// 거래 정보
	TradeAmount   int64   `json:"trade_amount"`   // 거래 금액
	IsMaker       bool    `json:"is_maker"`       // 매이커 여부

	// 수수료 구성
	BaseFeeRate   float64 `json:"base_fee_rate"`   // 기본 수수료율
	VIPDiscount   float64 `json:"vip_discount"`    // VIP 할인
	VolumeDiscount float64 `json:"volume_discount"` // 거래량 할인
	LiquidityFee  float64 `json:"liquidity_fee"`   // 유동성 수수료
	VolatilityFee float64 `json:"volatility_fee"`  // 변동성 수수료

	// 최종 결과
	FinalFeeRate  float64 `json:"final_fee_rate"`  // 최종 수수료율
	FeeAmount     int64   `json:"fee_amount"`      // 수수료 금액 (points)

	// 메타데이터
	UserVIPLevel  int     `json:"user_vip_level"`  // 사용자 VIP 등급
	MarketLiquidity float64 `json:"market_liquidity"` // 시장 유동성
	Explanation   string  `json:"explanation"`     // 수수료 설명
}

// UserTradingStats 사용자 거래 통계
type UserTradingStats struct {
	UserID          uint    `json:"user_id"`
	Volume24H       int64   `json:"volume_24h"`       // 24시간 거래량
	Volume30D       int64   `json:"volume_30d"`       // 30일 거래량
	Trades24H       int     `json:"trades_24h"`       // 24시간 거래 횟수
	Trades30D       int     `json:"trades_30d"`       // 30일 거래 횟수
	MakerRatio      float64 `json:"maker_ratio"`      // 매이커 비율
	AvgTradeSize    float64 `json:"avg_trade_size"`   // 평균 거래 크기
	VIPLevel        int     `json:"vip_level"`        // VIP 등급
	LastCalculated  time.Time `json:"last_calculated"` // 마지막 계산 시간
}

// NewFeeService 수수료 서비스 생성자
func NewFeeService(db *gorm.DB) *FeeService {
	return &FeeService{
		db: db,
	}
}

// GetDefaultConfig 기본 수수료 설정
func (fs *FeeService) GetDefaultConfig() FeeConfig {
	return FeeConfig{
		BaseMakerFee: 0.001, // 0.1%
		BaseTakerFee: 0.002, // 0.2%

		VIPTiers: []VIPTier{
			{Level: 1, MinVolume30D: 10000, MinTrades30D: 50, MakerDiscount: 0.05, TakerDiscount: 0.02, SpecialBenefits: "기본 혜택"},
			{Level: 2, MinVolume30D: 50000, MinTrades30D: 200, MakerDiscount: 0.10, TakerDiscount: 0.05, SpecialBenefits: "수수료 할인 + 분석 도구"},
			{Level: 3, MinVolume30D: 100000, MinTrades30D: 500, MakerDiscount: 0.15, TakerDiscount: 0.08, SpecialBenefits: "전용 고객 지원"},
			{Level: 4, MinVolume30D: 500000, MinTrades30D: 1000, MakerDiscount: 0.20, TakerDiscount: 0.12, SpecialBenefits: "API 우선 처리"},
			{Level: 5, MinVolume30D: 1000000, MinTrades30D: 2000, MakerDiscount: 0.25, TakerDiscount: 0.15, SpecialBenefits: "전용 매니저"},
		},

		VolumeDiscounts: []VolumeDiscount{
			{MinVolume24H: 1000, Discount: 0.01},   // 1% 할인
			{MinVolume24H: 5000, Discount: 0.02},   // 2% 할인
			{MinVolume24H: 10000, Discount: 0.03},  // 3% 할인
			{MinVolume24H: 50000, Discount: 0.05},  // 5% 할인
		},

		LiquidityMultiplier: 1.5,  // 유동성 부족시 1.5배
		VolatilityMultiplier: 1.3, // 변동성 높을 때 1.3배

		MinFee: 0.0005, // 0.05%
		MaxFee: 0.01,   // 1.0%
	}
}

// CalculateFee 동적 수수료 계산
func (fs *FeeService) CalculateFee(userID uint, milestoneID uint, optionID string, tradeAmount int64, isMaker bool) (*FeeCalculation, error) {
	config := fs.GetDefaultConfig()

	// 1. 사용자 거래 통계 조회
	userStats, err := fs.GetUserTradingStats(userID)
	if err != nil {
		return nil, err
	}

	// 2. 시장 조건 조회
	marketLiquidity := fs.GetMarketLiquidity(milestoneID, optionID)
	marketVolatility := fs.GetMarketVolatility(milestoneID, optionID)

	// 3. 기본 수수료율 결정
	var baseFeeRate float64
	if isMaker {
		baseFeeRate = config.BaseMakerFee
	} else {
		baseFeeRate = config.BaseTakerFee
	}

	// 4. VIP 할인 계산
	vipTier := fs.GetVIPTier(userStats, config.VIPTiers)
	var vipDiscount float64
	if vipTier != nil {
		if isMaker {
			vipDiscount = vipTier.MakerDiscount
		} else {
			vipDiscount = vipTier.TakerDiscount
		}
	}

	// 5. 거래량 할인 계산
	volumeDiscount := fs.GetVolumeDiscount(userStats.Volume24H, config.VolumeDiscounts)

	// 6. 시장 조건 기반 조정
	liquidityFee := 0.0
	if marketLiquidity < 0.3 { // 유동성이 30% 미만
		liquidityFee = baseFeeRate * (config.LiquidityMultiplier - 1.0)
	}

	volatilityFee := 0.0
	if marketVolatility > 0.1 { // 변동성이 10% 초과
		volatilityFee = baseFeeRate * (config.VolatilityMultiplier - 1.0)
	}

	// 7. 최종 수수료율 계산
	finalFeeRate := baseFeeRate
	finalFeeRate *= (1.0 - vipDiscount)      // VIP 할인 적용
	finalFeeRate *= (1.0 - volumeDiscount)   // 거래량 할인 적용
	finalFeeRate += liquidityFee             // 유동성 수수료 추가
	finalFeeRate += volatilityFee            // 변동성 수수료 추가

	// 최소/최대 제한 적용
	if finalFeeRate < config.MinFee {
		finalFeeRate = config.MinFee
	}
	if finalFeeRate > config.MaxFee {
		finalFeeRate = config.MaxFee
	}

	// 8. 수수료 금액 계산
	feeAmount := int64(float64(tradeAmount) * finalFeeRate)

	// 9. 설명 생성
	explanation := fs.generateExplanation(baseFeeRate, vipDiscount, volumeDiscount, liquidityFee, volatilityFee, isMaker)

	return &FeeCalculation{
		TradeAmount:     tradeAmount,
		IsMaker:         isMaker,
		BaseFeeRate:     baseFeeRate,
		VIPDiscount:     vipDiscount,
		VolumeDiscount:  volumeDiscount,
		LiquidityFee:    liquidityFee,
		VolatilityFee:   volatilityFee,
		FinalFeeRate:    finalFeeRate,
		FeeAmount:       feeAmount,
		UserVIPLevel:    userStats.VIPLevel,
		MarketLiquidity: marketLiquidity,
		Explanation:     explanation,
	}, nil
}

// GetUserTradingStats 사용자 거래 통계 조회
func (fs *FeeService) GetUserTradingStats(userID uint) (*UserTradingStats, error) {
	// Redis 캐시에서 먼저 조회
	cacheKey := fmt.Sprintf("user_stats:%d", userID)
	var stats UserTradingStats
	ctx := context.Background()

	if err := redis.Client.Get(ctx, cacheKey).Scan(&stats); err == nil {
		// 캐시에서 1시간 이내 데이터면 사용
		if time.Since(stats.LastCalculated) < time.Hour {
			return &stats, nil
		}
	}

	// 데이터베이스에서 실시간 계산
	now := time.Now()
	twentyFourHoursAgo := now.Add(-24 * time.Hour)
	thirtyDaysAgo := now.Add(-30 * 24 * time.Hour)

	// 24시간 통계
	var result24h struct {
		Volume int64
		Trades int64
	}

	fs.db.Model(&models.Trade{}).
		Select("COALESCE(SUM(total_amount), 0) as volume, COUNT(*) as trades").
		Where("(buyer_id = ? OR seller_id = ?) AND created_at > ?", userID, userID, twentyFourHoursAgo).
		Scan(&result24h)

	// 30일 통계
	var result30d struct {
		Volume int64
		Trades int64
	}

	fs.db.Model(&models.Trade{}).
		Select("COALESCE(SUM(total_amount), 0) as volume, COUNT(*) as trades").
		Where("(buyer_id = ? OR seller_id = ?) AND created_at > ?", userID, userID, thirtyDaysAgo).
		Scan(&result30d)

	// 매이커 비율 계산 (30일 기준)
	var makerTrades int64
	fs.db.Model(&models.Order{}).
		Where("user_id = ? AND created_at > ? AND status = ?", userID, thirtyDaysAgo, models.OrderStatusFilled).
		Count(&makerTrades)

	makerRatio := 0.0
	if result30d.Trades > 0 {
		makerRatio = float64(makerTrades) / float64(result30d.Trades)
	}

	// 평균 거래 크기
	avgTradeSize := 0.0
	if result30d.Trades > 0 {
		avgTradeSize = float64(result30d.Volume) / float64(result30d.Trades)
	}

	// VIP 레벨 계산
	vipLevel := fs.calculateVIPLevel(result30d.Volume, int(result30d.Trades))

	stats = UserTradingStats{
		UserID:         userID,
		Volume24H:      result24h.Volume,
		Volume30D:      result30d.Volume,
		Trades24H:      int(result24h.Trades),
		Trades30D:      int(result30d.Trades),
		MakerRatio:     makerRatio,
		AvgTradeSize:   avgTradeSize,
		VIPLevel:       vipLevel,
		LastCalculated: now,
	}

	// Redis에 캐시 (1시간 TTL)
	redis.Client.Set(ctx, cacheKey, stats, time.Hour)

	return &stats, nil
}

// GetMarketLiquidity 시장 유동성 조회
func (fs *FeeService) GetMarketLiquidity(milestoneID uint, optionID string) float64 {
	// 호가창 깊이로 유동성 측정
	var orderBookDepth struct {
		BidVolume int64
		AskVolume int64
	}

	fs.db.Model(&models.Order{}).
		Select("COALESCE(SUM(CASE WHEN side = 'buy' THEN remaining ELSE 0 END), 0) as bid_volume, COALESCE(SUM(CASE WHEN side = 'sell' THEN remaining ELSE 0 END), 0) as ask_volume").
		Where("milestone_id = ? AND option_id = ? AND status = ?", milestoneID, optionID, models.OrderStatusPending).
		Scan(&orderBookDepth)

	totalVolume := orderBookDepth.BidVolume + orderBookDepth.AskVolume

	// 유동성 점수 (0.0 - 1.0)
	// 10,000 이상이면 높은 유동성 (1.0)
	liquidity := math.Min(1.0, float64(totalVolume)/10000.0)

	return liquidity
}

// GetMarketVolatility 시장 변동성 조회
func (fs *FeeService) GetMarketVolatility(milestoneID uint, optionID string) float64 {
	// 최근 24시간 가격 변동성 계산
	var priceData []struct {
		Price float64
		Time  time.Time
	}

	twentyFourHoursAgo := time.Now().Add(-24 * time.Hour)

	fs.db.Model(&models.Trade{}).
		Select("price, created_at as time").
		Where("milestone_id = ? AND option_id = ? AND created_at > ?", milestoneID, optionID, twentyFourHoursAgo).
		Order("created_at ASC").
		Scan(&priceData)

	if len(priceData) < 2 {
		return 0.0 // 데이터 부족
	}

	// 가격 변화율의 표준편차 계산
	var priceChanges []float64
	for i := 1; i < len(priceData); i++ {
		change := (priceData[i].Price - priceData[i-1].Price) / priceData[i-1].Price
		priceChanges = append(priceChanges, change)
	}

	// 표준편차 계산
	if len(priceChanges) == 0 {
		return 0.0
	}

	var sum float64
	for _, change := range priceChanges {
		sum += change
	}
	mean := sum / float64(len(priceChanges))

	var variance float64
	for _, change := range priceChanges {
		variance += math.Pow(change-mean, 2)
	}
	variance /= float64(len(priceChanges))

	volatility := math.Sqrt(variance)

	return volatility
}

// Helper functions

func (fs *FeeService) GetVIPTier(userStats *UserTradingStats, tiers []VIPTier) *VIPTier {
	for i := len(tiers) - 1; i >= 0; i-- {
		tier := tiers[i]
		if userStats.Volume30D >= tier.MinVolume30D && userStats.Trades30D >= tier.MinTrades30D {
			return &tier
		}
	}
	return nil
}

func (fs *FeeService) GetVolumeDiscount(volume24h int64, discounts []VolumeDiscount) float64 {
	for i := len(discounts) - 1; i >= 0; i-- {
		discount := discounts[i]
		if volume24h >= discount.MinVolume24H {
			return discount.Discount
		}
	}
	return 0.0
}

func (fs *FeeService) calculateVIPLevel(volume30d int64, trades30d int) int {
	config := fs.GetDefaultConfig()
	for i := len(config.VIPTiers) - 1; i >= 0; i-- {
		tier := config.VIPTiers[i]
		if volume30d >= tier.MinVolume30D && trades30d >= tier.MinTrades30D {
			return tier.Level
		}
	}
	return 0
}

func (fs *FeeService) generateExplanation(baseFee, vipDiscount, volumeDiscount, liquidityFee, volatilityFee float64, isMaker bool) string {
	roleStr := "테이커"
	if isMaker {
		roleStr = "매이커"
	}

	explanation := fmt.Sprintf("기본 %s 수수료 %.2f%%", roleStr, baseFee*100)

	if vipDiscount > 0 {
		explanation += fmt.Sprintf(" - VIP 할인 %.1f%%", vipDiscount*100)
	}

	if volumeDiscount > 0 {
		explanation += fmt.Sprintf(" - 거래량 할인 %.1f%%", volumeDiscount*100)
	}

	if liquidityFee > 0 {
		explanation += fmt.Sprintf(" + 유동성 수수료 %.2f%%", liquidityFee*100)
	}

	if volatilityFee > 0 {
		explanation += fmt.Sprintf(" + 변동성 수수료 %.2f%%", volatilityFee*100)
	}

	return explanation
}

// GetUserVIPLevel 사용자 VIP 레벨 조회
func (fs *FeeService) GetUserVIPLevel(userID uint) (int, error) {
	stats, err := fs.GetUserTradingStats(userID)
	if err != nil {
		return 0, err
	}
	return stats.VIPLevel, nil
}

// EstimateFee 수수료 예상 계산 (주문 전 미리보기)
func (fs *FeeService) EstimateFee(userID uint, milestoneID uint, optionID string, tradeAmount int64, orderType models.OrderType) (*FeeCalculation, error) {
	// 주문 타입에 따라 매이커/테이커 결정
	// 시장가는 테이커, 지정가는 매이커로 가정 (실제로는 더 복잡)
	isMaker := orderType == models.OrderTypeLimit

	return fs.CalculateFee(userID, milestoneID, optionID, tradeAmount, isMaker)
}
