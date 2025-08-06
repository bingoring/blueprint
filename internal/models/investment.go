package models

import (
	"time"
)

// 🚀 Modern Trading Models (Polymarket Style)

// OrderType 주문 타입
type OrderType string

const (
	OrderTypeMarket OrderType = "market" // 시장가 주문
	OrderTypeLimit  OrderType = "limit"  // 지정가 주문
)

// OrderSide 주문 방향
type OrderSide string

const (
	OrderSideBuy  OrderSide = "buy"  // 매수
	OrderSideSell OrderSide = "sell" // 매도
)

// OrderStatus 주문 상태
type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "pending"   // 대기 중
	OrderStatusPartial   OrderStatus = "partial"   // 부분 체결
	OrderStatusFilled    OrderStatus = "filled"    // 완전 체결
	OrderStatusCancelled OrderStatus = "cancelled" // 취소됨
	OrderStatusExpired   OrderStatus = "expired"   // 만료됨
)

// Order P2P 주문 (폴리마켓 스타일)
type Order struct {
	ID          uint        `json:"id" gorm:"primaryKey"`
	ProjectID   uint        `json:"project_id"`
	MilestoneID uint        `json:"milestone_id"`
	OptionID    string      `json:"option_id"`
	UserID      uint        `json:"user_id"`
	Type        OrderType   `json:"type"`
	Side        OrderSide   `json:"side"`
	Quantity    int64       `json:"quantity"`     // 주문 수량
	Price       float64     `json:"price"`        // 주문 가격 (0-1 사이)
	Filled      int64       `json:"filled"`       // 체결된 수량
	Remaining   int64       `json:"remaining"`    // 남은 수량
	Status      OrderStatus `json:"status"`
	ExpiresAt   *time.Time  `json:"expires_at,omitempty"`
	IPAddress   string      `json:"ip_address,omitempty"`
	UserAgent   string      `json:"user_agent,omitempty"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`

	// 관계
	User      User      `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Project   Project   `json:"project,omitempty" gorm:"foreignKey:ProjectID"`
	Milestone Milestone `json:"milestone,omitempty" gorm:"foreignKey:MilestoneID"`
}

// Trade 거래 내역
type Trade struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	ProjectID    uint      `json:"project_id"`
	MilestoneID  uint      `json:"milestone_id"`
	OptionID     string    `json:"option_id"`
	BuyOrderID   uint      `json:"buy_order_id"`
	SellOrderID  uint      `json:"sell_order_id"`
	BuyerID      uint      `json:"buyer_id"`
	SellerID     uint      `json:"seller_id"`
	Quantity     int64     `json:"quantity"`     // 거래 수량
	Price        float64   `json:"price"`        // 거래 가격
	TotalAmount  int64     `json:"total_amount"` // 총 거래 금액 (points)
	BuyerFee     int64     `json:"buyer_fee"`    // 매수자 수수료
	SellerFee    int64     `json:"seller_fee"`   // 매도자 수수료
	CreatedAt    time.Time `json:"created_at"`

	// 관계
	BuyOrder  Order     `json:"buy_order,omitempty" gorm:"foreignKey:BuyOrderID"`
	SellOrder Order     `json:"sell_order,omitempty" gorm:"foreignKey:SellOrderID"`
	Buyer     User      `json:"buyer,omitempty" gorm:"foreignKey:BuyerID"`
	Seller    User      `json:"seller,omitempty" gorm:"foreignKey:SellerID"`
	Project   Project   `json:"project,omitempty" gorm:"foreignKey:ProjectID"`
	Milestone Milestone `json:"milestone,omitempty" gorm:"foreignKey:MilestoneID"`
}

// Position 사용자 포지션
type Position struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	UserID      uint      `json:"user_id"`
	ProjectID   uint      `json:"project_id"`
	MilestoneID uint      `json:"milestone_id"`
	OptionID    string    `json:"option_id"`
	Quantity    int64     `json:"quantity"`      // 보유 수량 (+매수, -매도)
	AvgPrice    float64   `json:"avg_price"`     // 평균 취득 가격
	TotalCost   int64     `json:"total_cost"`    // 총 투입 비용
	Realized    int64     `json:"realized"`      // 실현 손익
	Unrealized  int64     `json:"unrealized"`    // 미실현 손익
	UpdatedAt   time.Time `json:"updated_at"`

	// 관계
	User      User      `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Project   Project   `json:"project,omitempty" gorm:"foreignKey:ProjectID"`
	Milestone Milestone `json:"milestone,omitempty" gorm:"foreignKey:MilestoneID"`
}

// MarketData 시장 데이터
type MarketData struct {
	ID              uint      `json:"id" gorm:"primaryKey"`
	MilestoneID     uint      `json:"milestone_id"`
	OptionID        string    `json:"option_id"`
	CurrentPrice    float64   `json:"current_price"`     // 현재 가격
	PreviousPrice   float64   `json:"previous_price"`    // 이전 가격
	Change24h       float64   `json:"change_24h"`        // 24시간 변동폭
	ChangePercent   float64   `json:"change_percent"`    // 변동율 (%)
	Volume24h       int64     `json:"volume_24h"`        // 24시간 거래량
	Trades24h       int       `json:"trades_24h"`        // 24시간 거래 수
	HighPrice24h    float64   `json:"high_price_24h"`    // 24시간 최고가
	LowPrice24h     float64   `json:"low_price_24h"`     // 24시간 최저가
	BidPrice        float64   `json:"bid_price"`         // 현재 매수 호가
	AskPrice        float64   `json:"ask_price"`         // 현재 매도 호가
	Spread          float64   `json:"spread"`            // 호가 스프레드
	MarketCap       int64     `json:"market_cap"`        // 시가총액
	Liquidity       int64     `json:"liquidity"`         // 유동성
	LastTradeTime   time.Time `json:"last_trade_time"`   // 마지막 거래 시간
	UpdatedAt       time.Time `json:"updated_at"`

	// 관계
	Milestone Milestone `json:"milestone,omitempty" gorm:"foreignKey:MilestoneID"`
}

func (MarketData) TableName() string {
	return "market_data"
}

// 🪙 화폐 타입
type CurrencyType string

const (
	CurrencyUSDC      CurrencyType = "USDC"      // 스테이블코인 (베팅/보상)
	CurrencyBLUEPRINT CurrencyType = "BLUEPRINT" // 자체 토큰 (거버넌스/스테이킹)
)

// UserWallet 사용자 지갑 (하이브리드)
type UserWallet struct {
	ID     uint `json:"id" gorm:"primaryKey"`
	UserID uint `json:"user_id" gorm:"uniqueIndex;not null"`

	// 🔵 USDC 잔액 (베팅/보상용)
	USDCBalance       int64 `json:"usdc_balance" gorm:"default:0"`         // 사용 가능한 USDC (센트 단위)
	USDCLockedBalance int64 `json:"usdc_locked_balance" gorm:"default:0"`  // 베팅으로 잠긴 USDC

	// 🟦 BLUEPRINT 토큰 잔액 (거버넌스/스테이킹용)
	BlueprintBalance       int64 `json:"blueprint_balance" gorm:"default:0"`        // 사용 가능한 BLUEPRINT (Wei 단위)
	BlueprintLockedBalance int64 `json:"blueprint_locked_balance" gorm:"default:0"` // 스테이킹/분쟁으로 잠긴 BLUEPRINT

	// 📊 통계 (USDC 기준)
	TotalUSDCDeposit    int64 `json:"total_usdc_deposit" gorm:"default:0"`    // 총 USDC 입금
	TotalUSDCWithdraw   int64 `json:"total_usdc_withdraw" gorm:"default:0"`   // 총 USDC 출금
	TotalUSDCProfit     int64 `json:"total_usdc_profit" gorm:"default:0"`     // 총 USDC 수익
	TotalUSDCLoss       int64 `json:"total_usdc_loss" gorm:"default:0"`       // 총 USDC 손실
	TotalUSDCFees       int64 `json:"total_usdc_fees" gorm:"default:0"`       // 총 USDC 수수료

	// 📈 통계 (BLUEPRINT 기준)
	TotalBlueprintEarned int64 `json:"total_blueprint_earned" gorm:"default:0"` // 총 BLUEPRINT 획득
	TotalBlueprintSpent  int64 `json:"total_blueprint_spent" gorm:"default:0"`  // 총 BLUEPRINT 사용

	// 🎯 성과
	WinRate        float64   `json:"win_rate" gorm:"default:0"`        // 승률
	TotalTrades    int64     `json:"total_trades" gorm:"default:0"`    // 총 거래 수
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`

	// 관계
	User User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

func (UserWallet) TableName() string {
	return "user_wallets"
}

// PriceHistory 가격 히스토리
type PriceHistory struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	MilestoneID uint      `json:"milestone_id"`
	OptionID    string    `json:"option_id"`
	Price       float64   `json:"price"`
	Volume      int64     `json:"volume"`
	CreatedAt   time.Time `json:"created_at"`
}

func (PriceHistory) TableName() string {
	return "price_history"
}

// 🆕 ===== 하이브리드 화폐 시스템 모델들 =====

// 📈 스테이킹 시스템
type StakingPool struct {
	ID       uint      `json:"id" gorm:"primaryKey"`
	UserID   uint      `json:"user_id" gorm:"not null;index"`
	Amount   int64     `json:"amount"`                        // 스테이킹한 BLUEPRINT 양
	StartDate time.Time `json:"start_date"`                   // 스테이킹 시작일
	EndDate   *time.Time `json:"end_date"`                    // 스테이킹 종료일 (활성 시 nil)
	Status    string    `json:"status" gorm:"default:'active'"` // active, withdrawn

	// 누적 보상
	TotalUSDCRewards int64 `json:"total_usdc_rewards" gorm:"default:0"` // 받은 USDC 보상 총액

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// 관계
	User User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

func (StakingPool) TableName() string {
	return "staking_pools"
}

// 💵 수수료 분배 내역
type RevenueDistribution struct {
	ID            uint      `json:"id" gorm:"primaryKey"`
	TotalRevenue  int64     `json:"total_revenue"`   // 해당 기간 총 USDC 수수료
	DistributionDate time.Time `json:"distribution_date"` // 분배 날짜
	TotalStakers  int       `json:"total_stakers"`   // 분배 대상 스테이커 수

	CreatedAt time.Time `json:"created_at"`
}

func (RevenueDistribution) TableName() string {
	return "revenue_distributions"
}

// 💎 개별 스테이커 보상 내역
type StakingReward struct {
	ID                     uint      `json:"id" gorm:"primaryKey"`
	RevenueDistributionID  uint      `json:"revenue_distribution_id" gorm:"not null;index"`
	UserID                 uint      `json:"user_id" gorm:"not null;index"`
	StakedAmount           int64     `json:"staked_amount"`    // 분배 시점의 스테이킹 양
	RewardAmount           int64     `json:"reward_amount"`    // 받은 USDC 보상

	CreatedAt time.Time `json:"created_at"`

	// 관계
	RevenueDistribution RevenueDistribution `json:"revenue_distribution,omitempty" gorm:"foreignKey:RevenueDistributionID"`
	User                User                `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

func (StakingReward) TableName() string {
	return "staking_rewards"
}

// ⚖️ 거버넌스 투표
type GovernanceProposal struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Title       string    `json:"title" gorm:"not null"`
	Description string    `json:"description" gorm:"type:text"`
	ProposerID  uint      `json:"proposer_id" gorm:"not null"`

	// 투표 설정
	VotingStartDate time.Time  `json:"voting_start_date"`
	VotingEndDate   time.Time  `json:"voting_end_date"`
	MinQuorum       int64      `json:"min_quorum"`        // 최소 투표권 수 (BLUEPRINT)

	// 결과
	VotesFor     int64  `json:"votes_for" gorm:"default:0"`
	VotesAgainst int64  `json:"votes_against" gorm:"default:0"`
	Status       string `json:"status" gorm:"default:'pending'"` // pending, active, passed, rejected, executed

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// 관계
	Proposer User `json:"proposer,omitempty" gorm:"foreignKey:ProposerID"`
}

func (GovernanceProposal) TableName() string {
	return "governance_proposals"
}

// 🗳️ 개별 투표
type GovernanceVote struct {
	ID         uint   `json:"id" gorm:"primaryKey"`
	ProposalID uint   `json:"proposal_id" gorm:"not null;index"`
	UserID     uint   `json:"user_id" gorm:"not null;index"`
	VotePower  int64  `json:"vote_power"`                    // 투표 시점의 BLUEPRINT 보유량
	Direction  string `json:"direction" gorm:"not null"`     // for, against

	CreatedAt time.Time `json:"created_at"`

	// 관계 & 유니크 제약
	Proposal GovernanceProposal `json:"proposal,omitempty" gorm:"foreignKey:ProposalID"`
	User     User               `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

func (GovernanceVote) TableName() string {
	return "governance_votes"
}

// 🎁 BLUEPRINT 토큰 지급 내역
type BlueprintReward struct {
	ID       uint         `json:"id" gorm:"primaryKey"`
	UserID   uint         `json:"user_id" gorm:"not null;index"`
	Amount   int64        `json:"amount"`                     // 지급된 BLUEPRINT 양
	Reason   string       `json:"reason"`                     // 지급 사유
	Category RewardCategory `json:"category" gorm:"not null"` // 카테고리

	// 참조 ID (옵션)
	ProjectID   *uint `json:"project_id,omitempty"`   // 관련 프로젝트
	MilestoneID *uint `json:"milestone_id,omitempty"` // 관련 마일스톤

	CreatedAt time.Time `json:"created_at"`

	// 관계
	User      User       `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Project   *Project   `json:"project,omitempty" gorm:"foreignKey:ProjectID"`
	Milestone *Milestone `json:"milestone,omitempty" gorm:"foreignKey:MilestoneID"`
}

// 🏆 보상 카테고리
type RewardCategory string

const (
	RewardSignup         RewardCategory = "signup"          // 회원가입 보상
	RewardProjectCreate  RewardCategory = "project_create"  // 프로젝트 생성 보상
	RewardMilestoneSuccess RewardCategory = "milestone_success" // 마일스톤 달성 보상
	RewardMentoring      RewardCategory = "mentoring"       // 멘토링 활동 보상
	RewardCommunity      RewardCategory = "community"       // 커뮤니티 기여 보상
	RewardReferral       RewardCategory = "referral"        // 추천인 보상
	RewardDispute        RewardCategory = "dispute"         // 분쟁 해결 참여 보상
)

func (BlueprintReward) TableName() string {
	return "blueprint_rewards"
}

// 💸 거래 수수료 설정
type PlatformFeeConfig struct {
	ID                uint    `json:"id" gorm:"primaryKey"`
	TradingFeeRate    float64 `json:"trading_fee_rate" gorm:"default:0.05"`    // 5% 거래 수수료
	WithdrawFeeFlat   int64   `json:"withdraw_fee_flat" gorm:"default:100"`    // $1 출금 수수료 (센트)
	MinBetAmount      int64   `json:"min_bet_amount" gorm:"default:100"`       // $1 최소 베팅 (센트)
	MaxBetAmount      int64   `json:"max_bet_amount" gorm:"default:1000000"`   // $10,000 최대 베팅 (센트)

	// 스테이킹 보상 비율
	StakingRewardRate float64 `json:"staking_reward_rate" gorm:"default:0.70"` // 수수료의 70%를 스테이커에게 분배

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (PlatformFeeConfig) TableName() string {
	return "platform_fee_configs"
}

// 🔥 API Request/Response Models

// CreateOrderRequest 주문 생성 요청 (USDC 기준)
type CreateOrderRequest struct {
	ProjectID   uint      `json:"project_id" binding:"required"`
	MilestoneID uint      `json:"milestone_id" binding:"required"`
	OptionID    string    `json:"option_id" binding:"required"`
	Type        OrderType `json:"type" binding:"required"`
	Side        OrderSide `json:"side" binding:"required"`
	Quantity    int64     `json:"quantity" binding:"required,min=1"`              // 주식 수량
	Price       float64   `json:"price" binding:"required,min=0.01,max=0.99"`    // 확률 (0.01-0.99)
	Currency    CurrencyType `json:"currency" gorm:"default:'USDC'"`              // 화폐 타입 (항상 USDC)
}

// OrderResponse 주문 응답
type OrderResponse struct {
	Order  Order   `json:"order"`
	Trades []Trade `json:"trades,omitempty"`
}

// OrderBookLevel 호가창 레벨
type OrderBookLevel struct {
	Price    float64 `json:"price"`
	Quantity int64   `json:"quantity"`
	Count    int     `json:"count"` // 주문 개수
}

// OrderBook 호가창
type OrderBook struct {
	MilestoneID uint             `json:"milestone_id"`
	OptionID    string           `json:"option_id"`
	Bids        []OrderBookLevel `json:"bids"` // 매수 호가 (높은 가격부터)
	Asks        []OrderBookLevel `json:"asks"` // 매도 호가 (낮은 가격부터)
	Spread      float64          `json:"spread"`
	LastUpdate  time.Time        `json:"last_update"`
}

// OrderBookResponse 호가창 응답
type OrderBookResponse struct {
	OrderBook OrderBook `json:"orderbook"`
}

// TradeImpact 거래 영향도
type TradeImpact struct {
	Quantity       int64   `json:"quantity"`         // 주문 수량
	TotalCost      int64   `json:"total_cost"`       // 총 비용
	AvgPrice       float64 `json:"avg_price"`        // 평균 체결 가격
	PriceImpact    float64 `json:"price_impact"`     // 가격 영향도 (%)
	Fee            int64   `json:"fee"`              // 예상 수수료
	ExpectedPayout int64   `json:"expected_payout"`  // 예상 지급액
	ROI            float64 `json:"roi"`              // 예상 수익률 (%)
}

// MarketStatusResponse 마켓 상태 응답
type MarketStatusResponse struct {
	MarketData   []MarketData `json:"market_data"`
	ViewerCount  int          `json:"viewer_count"`
	TotalClients int          `json:"total_clients"`
}
