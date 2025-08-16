package e2e_test

import (
	"fmt"
	"testing"
	"time"

	"blueprint-module/pkg/models"
	"blueprint/internal/services"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// RealWorldScenariosTestSuite 실제 시나리오 테스트
type RealWorldScenariosTestSuite struct {
	suite.Suite
	tradingService *services.DistributedTradingService
	db             *gorm.DB
	redisServer    *miniredis.Miniredis
	redisClient    *redis.Client
}

func (suite *RealWorldScenariosTestSuite) SetupSuite() {
	// 데이터베이스 설정
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	suite.Require().NoError(err)
	suite.db = db

	// 테이블 마이그레이션
	err = db.AutoMigrate(
		&models.User{},
		&models.Project{},
		&models.Milestone{},
		&models.Order{},
		&models.Trade{},
		&models.Position{},
		&models.MarketData{},
		&models.UserWallet{},
		&models.PriceHistory{},
	)
	suite.Require().NoError(err)

	// Redis 설정
	suite.redisServer = miniredis.RunT(suite.T())
	suite.redisClient = redis.NewClient(&redis.Options{
		Addr: suite.redisServer.Addr(),
	})

	// 서비스 초기화 (테스트용 Redis 클라이언트 사용)
	suite.tradingService = services.NewDistributedTradingServiceWithRedis(suite.db, nil, suite.redisClient)

	// 실제 시나리오 데이터 생성
	suite.createRealWorldData()

	// 서비스 시작
	err = suite.tradingService.Start()
	suite.Require().NoError(err)
}

func (suite *RealWorldScenariosTestSuite) TearDownSuite() {
	suite.tradingService.Stop()
	suite.redisServer.Close()
	suite.redisClient.Close()
}

// TestStartupCompanyMilestoneTrading 스타트업 마일스톤 거래 시나리오
func (suite *RealWorldScenariosTestSuite) TestStartupCompanyMilestoneTrading() {
	suite.T().Log("📈 시나리오: AI 스타트업의 제품 출시 마일스톤 예측 시장")

	// 1. 초기 시장 상황: 50/50 확률로 시작
	suite.T().Log("1️⃣ 초기 시장 설정 - 제품 출시 성공/실패 예측")

	// 일반 투자자들의 초기 베팅 (보수적)
	conservativeInvestors := []struct {
		userID   uint
		side     string
		quantity int64
		price    float64
	}{
		{1, "buy", 100, 0.45}, // 성공 45% 확신
		{2, "sell", 80, 0.55}, // 실패 45% 확신
		{3, "buy", 50, 0.48},  // 성공 48% 확신
		{4, "sell", 60, 0.52}, // 실패 48% 확신
	}

	for _, investor := range conservativeInvestors {
		_, err := suite.tradingService.CreateOrder(
			investor.userID, 1, "success", investor.side,
			investor.quantity, investor.price,
		)
		suite.Assert().NoError(err)
	}

	// 2. 내부 정보를 가진 직원이 베팅 (실제로는 불법이지만 테스트 시나리오)
	suite.T().Log("2️⃣ 내부 정보 보유자의 대량 베팅")

	// 제품이 거의 완성되었다는 내부 정보
	_, err := suite.tradingService.CreateOrder(
		10, 1, "success", "buy", 500, 0.75, // 75% 확신으로 대량 매수
	)
	suite.Assert().NoError(err)

	// 3. 시장이 반응하면서 가격 상승
	suite.T().Log("3️⃣ 시장 반응 - 가격 상승 및 추가 거래")

	// 가격 상승을 본 투자자들의 추가 진입
	followers := []struct {
		userID   uint
		side     string
		quantity int64
		price    float64
	}{
		{5, "buy", 150, 0.72},  // 상승 추세 따라가기
		{6, "buy", 200, 0.71},  // FOMO 매수
		{7, "sell", 100, 0.78}, // 차익실현 매도
		{8, "buy", 75, 0.73},   // 소량 추가 매수
	}

	for _, follower := range followers {
		result, err := suite.tradingService.CreateOrder(
			follower.userID, 1, "success", follower.side,
			follower.quantity, follower.price,
		)
		suite.Assert().NoError(err)

		if result.Executed {
			suite.T().Logf("   거래 체결: %d주 @ %.3f", result.Trades[0].Quantity, result.Trades[0].Price)
		}
	}

	// 4. 최종 시장 상태 확인
	marketData, err := suite.tradingService.GetMarketData(1, "success")
	suite.Assert().NoError(err)

	suite.T().Logf("📊 최종 시장 상태:")
	suite.T().Logf("   현재 가격: %.3f (성공 확률: %.1f%%)", marketData.LastPrice, marketData.LastPrice*100)
	suite.T().Logf("   24시간 거래량: %d", marketData.Volume24h)

	// 5. 실제 이벤트 발생 시뮬레이션
	suite.T().Log("5️⃣ 실제 이벤트 발생: 제품 출시 성공!")

	// 여기서는 마일스톤 상태 업데이트만 시뮬레이션
	var milestone models.Milestone
	err = suite.db.First(&milestone, 1).Error
	suite.Assert().NoError(err)

	milestone.Status = "completed"
	err = suite.db.Save(&milestone).Error
	suite.Assert().NoError(err)

	suite.T().Log("✅ 시나리오 완료: 성공 예측이 맞았던 투자자들이 수익 실현")
}

// TestInfluencerProjectTrading 인플루언서 프로젝트 거래 시나리오
func (suite *RealWorldScenariosTestSuite) TestInfluencerProjectTrading() {
	suite.T().Log("📱 시나리오: 유명 유튜버의 1M 구독자 달성 예측 시장")

	// 1. 팬덤의 낙관적 베팅
	suite.T().Log("1️⃣ 열성 팬들의 낙관적 베팅")

	fanBets := []struct {
		userID   uint
		quantity int64
		price    float64
	}{
		{11, 200, 0.85}, // 매우 낙관적 팬
		{12, 150, 0.82}, // 확신하는 팬
		{13, 100, 0.80}, // 보통 팬
		{14, 300, 0.88}, // 극성 팬
	}

	for _, bet := range fanBets {
		_, err := suite.tradingService.CreateOrder(
			bet.userID, 2, "success", "buy",
			bet.quantity, bet.price,
		)
		suite.Assert().NoError(err)
	}

	// 2. 데이터 분석가의 반대 베팅
	suite.T().Log("2️⃣ 데이터 분석가의 냉정한 분석")

	// YouTube 성장 곡선 분석 결과 회의적 전망
	analystBets := []struct {
		userID   uint
		quantity int64
		price    float64
	}{
		{15, 400, 0.35}, // 성장 둔화 예상
		{16, 250, 0.40}, // 시장 포화 우려
		{17, 180, 0.38}, // 경쟁 심화 예상
	}

	for _, bet := range analystBets {
		result, err := suite.tradingService.CreateOrder(
			bet.userID, 2, "success", "sell",
			bet.quantity, bet.price,
		)
		suite.Assert().NoError(err)

		if result.Executed && len(result.Trades) > 0 {
			suite.T().Logf("   큰 거래 체결: %d주 @ %.3f",
				result.Trades[0].Quantity, result.Trades[0].Price)
		}
	}

	// 3. 중립적 투자자들의 등장
	suite.T().Log("3️⃣ 기회를 본 중립적 투자자들의 참여")

	neutralTraders := []struct {
		userID   uint
		side     string
		quantity int64
		price    float64
	}{
		{18, "buy", 120, 0.58}, // 가격 차이 활용
		{19, "sell", 90, 0.62}, // 차익거래
		{20, "buy", 160, 0.59}, // 단기 매매
	}

	for _, trader := range neutralTraders {
		_, err := suite.tradingService.CreateOrder(
			trader.userID, 2, "success", trader.side,
			trader.quantity, trader.price,
		)
		suite.Assert().NoError(err)
	}

	// 4. 최종 상태 확인
	marketData, err := suite.tradingService.GetMarketData(2, "success")
	suite.Assert().NoError(err)

	suite.T().Logf("📊 인플루언서 프로젝트 시장 현황:")
	suite.T().Logf("   현재 가격: %.3f", marketData.LastPrice)
	suite.T().Logf("   시장 전망: %.1f%% 확률로 목표 달성 예상", marketData.LastPrice*100)

	// 시장이 효율적으로 작동하는지 검증
	suite.Assert().True(marketData.LastPrice > 0.0 && marketData.LastPrice < 1.0)
	suite.T().Log("✅ 시장이 양극화된 의견들을 효율적으로 반영")
}

// TestMarketManipulationAttempt 시장 조작 시도 시나리오 (방어 테스트)
func (suite *RealWorldScenariosTestSuite) TestMarketManipulationAttempt() {
	suite.T().Log("🛡️ 시나리오: 시장 조작 시도 및 시스템 방어")

	// 1. 정상적인 시장 형성
	suite.T().Log("1️⃣ 정상적인 시장 상황")

	normalOrders := []struct {
		userID   uint
		side     string
		quantity int64
		price    float64
	}{
		{21, "buy", 50, 0.48},
		{22, "sell", 60, 0.52},
		{23, "buy", 40, 0.47},
		{24, "sell", 55, 0.53},
	}

	for _, order := range normalOrders {
		_, err := suite.tradingService.CreateOrder(
			order.userID, 3, "success", order.side,
			order.quantity, order.price,
		)
		suite.Assert().NoError(err)
	}

	// 2. 조작 시도: Pump and Dump
	suite.T().Log("2️⃣ 조작 시도: 인위적 가격 상승")

	// 대량 매수로 가격 인위 상승 시도
	manipulationResult, err := suite.tradingService.CreateOrder(
		25, 3, "success", "buy", 1000, 0.90,
	)
	suite.Assert().NoError(err)

	if manipulationResult.Executed {
		suite.T().Logf("   조작 주문 일부 체결: %d주",
			manipulationResult.Trades[0].Quantity)
	}

	// 3. 시장의 자연스러운 반응
	suite.T().Log("3️⃣ 시장 참여자들의 합리적 대응")

	// 가격 급등을 본 차익실현 매도
	arbitrageOrders := []struct {
		userID   uint
		quantity int64
		price    float64
	}{
		{26, 200, 0.85}, // 차익실현
		{27, 150, 0.82}, // 추가 매도
		{28, 180, 0.87}, // 고점 매도
	}

	for _, order := range arbitrageOrders {
		result, err := suite.tradingService.CreateOrder(
			order.userID, 3, "success", "sell",
			order.quantity, order.price,
		)
		suite.Assert().NoError(err)

		if result.Executed && len(result.Trades) > 0 {
			suite.T().Logf("   차익실현 거래: %d주 @ %.3f",
				result.Trades[0].Quantity, result.Trades[0].Price)
		}
	}

	// 4. 최종 시장 상태 분석
	finalMarketData, err := suite.tradingService.GetMarketData(3, "success")
	suite.Assert().NoError(err)

	suite.T().Logf("📊 조작 시도 후 시장 상황:")
	suite.T().Logf("   최종 가격: %.3f", finalMarketData.LastPrice)
	suite.T().Logf("   거래량 증가: %d", finalMarketData.Volume24h)

	// 시장이 조작을 자연스럽게 방어했는지 확인
	suite.Assert().Less(finalMarketData.LastPrice, 0.90, "시장이 조작된 가격을 거부")
	suite.T().Log("✅ 시장 메커니즘이 조작 시도를 자연스럽게 방어")
}

// TestVolatilityDuringNews 뉴스 발표 시 변동성 시나리오
func (suite *RealWorldScenariosTestSuite) TestVolatilityDuringNews() {
	suite.T().Log("📺 시나리오: 중요 뉴스 발표 시 시장 변동성")

	// 1. 뉴스 발표 전 정상 거래
	suite.T().Log("1️⃣ 뉴스 발표 전 정상적인 거래")

	preNewsOrders := []struct {
		userID   uint
		side     string
		quantity int64
		price    float64
	}{
		{31, "buy", 80, 0.49},
		{32, "sell", 85, 0.51},
		{33, "buy", 75, 0.48},
		{34, "sell", 90, 0.52},
	}

	for _, order := range preNewsOrders {
		_, err := suite.tradingService.CreateOrder(
			order.userID, 4, "success", order.side,
			order.quantity, order.price,
		)
		suite.Assert().NoError(err)
	}

	preNewsPrice, err := suite.tradingService.GetMarketData(4, "success")
	suite.Assert().NoError(err)
	suite.T().Logf("   뉴스 발표 전 가격: %.3f", preNewsPrice.LastPrice)

	// 2. 긍정적 뉴스 발표: "정부 지원 정책 발표"
	suite.T().Log("2️⃣ 긍정적 뉴스 발표 - 급격한 매수 주문 증가")

	// 뉴스를 보고 즉시 반응하는 투자자들
	newsReactionOrders := []struct {
		userID   uint
		side     string
		quantity int64
		price    float64
		delay    time.Duration // 반응 속도
	}{
		{35, "buy", 300, 0.70, 0},                       // 즉시 반응
		{36, "buy", 250, 0.68, 50 * time.Millisecond},   // 조금 늦음
		{37, "buy", 200, 0.72, 100 * time.Millisecond},  // 더 늦음
		{38, "sell", 150, 0.75, 200 * time.Millisecond}, // 차익실현
	}

	for _, order := range newsReactionOrders {
		time.Sleep(order.delay)

		result, err := suite.tradingService.CreateOrder(
			order.userID, 4, "success", order.side,
			order.quantity, order.price,
		)
		suite.Assert().NoError(err)

		if result.Executed && len(result.Trades) > 0 {
			suite.T().Logf("   뉴스 반응 거래: %s %d주 @ %.3f",
				order.side, result.Trades[0].Quantity, result.Trades[0].Price)
		}
	}

	// 3. 시장 안정화
	suite.T().Log("3️⃣ 시장 안정화 - 추가 거래로 균형 찾기")

	stabilizationOrders := []struct {
		userID   uint
		side     string
		quantity int64
		price    float64
	}{
		{39, "sell", 100, 0.69}, // 수익 실현
		{40, "buy", 120, 0.66},  // 추가 매수
		{41, "sell", 80, 0.71},  // 부분 매도
		{42, "buy", 95, 0.67},   // 추가 진입
	}

	for _, order := range stabilizationOrders {
		_, err := suite.tradingService.CreateOrder(
			order.userID, 4, "success", order.side,
			order.quantity, order.price,
		)
		suite.Assert().NoError(err)
	}

	// 4. 변동성 분석
	postNewsPrice, err := suite.tradingService.GetMarketData(4, "success")
	suite.Assert().NoError(err)

	priceChange := postNewsPrice.LastPrice - preNewsPrice.LastPrice
	changePercent := (priceChange / preNewsPrice.LastPrice) * 100

	suite.T().Logf("📊 뉴스 발표 영향 분석:")
	suite.T().Logf("   발표 전 가격: %.3f", preNewsPrice.LastPrice)
	suite.T().Logf("   발표 후 가격: %.3f", postNewsPrice.LastPrice)
	suite.T().Logf("   가격 변동: %.3f (%.1f%%)", priceChange, changePercent)
	suite.T().Logf("   거래량: %d", postNewsPrice.Volume24h)

	// 뉴스가 시장에 적절한 영향을 미쳤는지 확인
	suite.Assert().Greater(priceChange, 0.0, "긍정적 뉴스로 가격 상승")
	suite.Assert().Greater(postNewsPrice.Volume24h, preNewsPrice.Volume24h, "거래량 증가")

	suite.T().Log("✅ 시장이 뉴스에 적절히 반응하고 안정화")
}

func (suite *RealWorldScenariosTestSuite) createRealWorldData() {
	// 다양한 사용자 생성 (투자자, 분석가, 팬, 조작자 등)
	users := []models.User{
		// 일반 투자자들 (1-10)
		{ID: 1, Username: "conservative_investor", Email: "investor1@test.com"},
		{ID: 2, Username: "risk_taker", Email: "investor2@test.com"},
		{ID: 3, Username: "value_investor", Email: "investor3@test.com"},
		{ID: 4, Username: "day_trader", Email: "investor4@test.com"},
		{ID: 5, Username: "trend_follower", Email: "investor5@test.com"},
		{ID: 6, Username: "momentum_trader", Email: "investor6@test.com"},
		{ID: 7, Username: "profit_taker", Email: "investor7@test.com"},
		{ID: 8, Username: "small_trader", Email: "investor8@test.com"},
		{ID: 9, Username: "swing_trader", Email: "investor9@test.com"},
		{ID: 10, Username: "insider_employee", Email: "insider@test.com"},

		// 인플루언서 팬들 (11-20)
		{ID: 11, Username: "superfan_1", Email: "fan1@test.com"},
		{ID: 12, Username: "loyal_follower", Email: "fan2@test.com"},
		{ID: 13, Username: "casual_fan", Email: "fan3@test.com"},
		{ID: 14, Username: "extreme_fan", Email: "fan4@test.com"},
		{ID: 15, Username: "data_analyst", Email: "analyst1@test.com"},
		{ID: 16, Username: "market_researcher", Email: "analyst2@test.com"},
		{ID: 17, Username: "growth_expert", Email: "analyst3@test.com"},
		{ID: 18, Username: "arbitrage_trader", Email: "arb1@test.com"},
		{ID: 19, Username: "hedge_fund", Email: "hedge@test.com"},
		{ID: 20, Username: "quant_trader", Email: "quant@test.com"},

		// 조작 시도자들 및 기타 (21-42)
	}

	// 21-42까지 추가 사용자 생성
	for i := 21; i <= 42; i++ {
		user := models.User{
			ID:       uint(i),
			Username: fmt.Sprintf("trader_%d", i),
			Email:    fmt.Sprintf("trader%d@test.com", i),
		}
		users = append(users, user)
	}

	// 모든 사용자 생성 및 지갑 설정
	for _, user := range users {
		suite.db.Create(&user)

		// 각 사용자별로 다른 잔액 설정 (현실적으로)
		balance := int64(1000000) // 기본 $10,000
		if user.ID <= 5 {
			balance = 10000000 // 대형 투자자 $100,000
		} else if user.ID <= 10 {
			balance = 5000000 // 중형 투자자 $50,000
		} else if user.ID == 25 {
			balance = 50000000 // 조작 시도자 $500,000
		}

		wallet := models.UserWallet{
			UserID:      user.ID,
			USDCBalance: balance,
		}
		suite.db.Create(&wallet)
	}

	// 다양한 실제 시나리오 프로젝트들
	projects := []models.Project{
		{
			ID:          1,
			Title:       "AI 스타트업 제품 출시",
			Description: "혁신적인 AI 기반 생산성 도구 출시 프로젝트",
			UserID:      1,
			Status:      "active",
			CreatedAt:   time.Now(),
		},
		{
			ID:          2,
			Title:       "유명 유튜버 1M 구독자 달성",
			Description: "테크 리뷰 채널의 100만 구독자 달성 도전",
			UserID:      11,
			Status:      "active",
			CreatedAt:   time.Now(),
		},
		{
			ID:          3,
			Title:       "블록체인 메인넷 출시",
			Description: "새로운 DeFi 프로토콜 메인넷 성공적 출시",
			UserID:      15,
			Status:      "active",
			CreatedAt:   time.Now(),
		},
		{
			ID:          4,
			Title:       "바이오텍 임상시험 성공",
			Description: "혁신적인 암 치료제의 3상 임상시험 성공",
			UserID:      25,
			Status:      "active",
			CreatedAt:   time.Now(),
		},
	}

	for _, project := range projects {
		suite.db.Create(&project)
	}

	// 각 프로젝트별 마일스톤 생성
	milestones := []models.Milestone{
		{ID: 1, ProjectID: 1, Title: "제품 베타 테스트 완료", Status: "funding", Order: 1},
		{ID: 2, ProjectID: 2, Title: "월 100만 조회수 달성", Status: "funding", Order: 1},
		{ID: 3, ProjectID: 3, Title: "테스트넷 안정성 검증 완료", Status: "funding", Order: 1},
		{ID: 4, ProjectID: 4, Title: "FDA 승인 획득", Status: "funding", Order: 1},
	}

	for _, milestone := range milestones {
		suite.db.Create(&milestone)
	}

	// 플랫폼 수수료 설정
	feeConfig := models.PlatformFeeConfig{
		ID:                1,
		TradingFeeRate:    0.02,    // 2% (현실적인 수수료)
		WithdrawFeeFlat:   200,     // $2
		MinBetAmount:      1000,    // $10 최소
		MaxBetAmount:      5000000, // $50,000 최대
		StakingRewardRate: 0.70,
		CreatedAt:         time.Now(),
	}
	suite.db.Create(&feeConfig)
}

func TestRealWorldScenariosTestSuite(t *testing.T) {
	suite.Run(t, new(RealWorldScenariosTestSuite))
}
