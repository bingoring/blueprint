package models

import (
	"time"

	"gorm.io/gorm"
)

// ActivityLog 사용자 활동 로그
type ActivityLog struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// 기본 정보
	UserID      uint   `json:"user_id" gorm:"not null;index"`
	ActivityType string `json:"activity_type" gorm:"not null;index"` // 활동 타입
	Action      string `json:"action" gorm:"not null"`               // 구체적인 액션
	Description string `json:"description"`                          // 활동 설명

	// 관련 엔티티 정보 (nullable)
	ProjectID   *uint `json:"project_id,omitempty" gorm:"index"`
	MilestoneID *uint `json:"milestone_id,omitempty" gorm:"index"`
	OrderID     *uint `json:"order_id,omitempty" gorm:"index"`
	TradeID     *uint `json:"trade_id,omitempty" gorm:"index"`

	// 메타데이터 (JSON)
	Metadata ActivityMetadata `json:"metadata" gorm:"type:jsonb"`

	// 관계
	User      User       `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Project   *Project   `json:"project,omitempty" gorm:"foreignKey:ProjectID"`
	Milestone *Milestone `json:"milestone,omitempty" gorm:"foreignKey:MilestoneID"`
}

// ActivityMetadata 활동의 상세 메타데이터
type ActivityMetadata struct {
	// 프로젝트 관련
	ProjectTitle    string  `json:"project_title,omitempty"`
	MilestoneTitle  string  `json:"milestone_title,omitempty"`

	// 거래 관련
	Amount          float64 `json:"amount,omitempty"`
	Price           float64 `json:"price,omitempty"`
	Currency        string  `json:"currency,omitempty"`
	OrderType       string  `json:"order_type,omitempty"` // "buy", "sell"

	// 성과 관련
	SuccessRate     float64 `json:"success_rate,omitempty"`
	ProfitLoss      float64 `json:"profit_loss,omitempty"`

	// 멘토링 관련
	MentorUsername  string  `json:"mentor_username,omitempty"`
	SessionDuration int     `json:"session_duration,omitempty"` // 분 단위
	Rating          int     `json:"rating,omitempty"`

	// 추가 컨텍스트
	IPAddress       string  `json:"ip_address,omitempty"`
	UserAgent       string  `json:"user_agent,omitempty"`
	Platform        string  `json:"platform,omitempty"` // "web", "mobile"
}

// ActivityType 상수 정의
const (
	// 프로젝트 관련
	ActivityTypeProject = "project"
	ActionProjectCreate = "create"
	ActionProjectUpdate = "update"
	ActionProjectDelete = "delete"
	ActionProjectPublish = "publish"
	ActionProjectComplete = "complete"

	// 마일스톤 관련
	ActivityTypeMilestone = "milestone"
	ActionMilestoneCreate = "create"
	ActionMilestoneUpdate = "update"
	ActionMilestoneComplete = "complete"
	ActionMilestoneValidate = "validate"

	// 거래 관련
	ActivityTypeTrade = "trade"
	ActionTradeBuy = "buy"
	ActionTradeSell = "sell"
	ActionTradeCancel = "cancel"
	ActionTradeExecute = "execute"

	// 멘토링 관련
	ActivityTypeMentoring = "mentoring"
	ActionMentoringStart = "start"
	ActionMentoringEnd = "end"
	ActionMentoringRate = "rate"
	ActionMentoringRequest = "request"

	// 계정 관련
	ActivityTypeAccount = "account"
	ActionAccountLogin = "login"
	ActionAccountLogout = "logout"
	ActionAccountRegister = "register"
	ActionAccountVerify = "verify"
	ActionAccountUpdate = "update"

	// 투자 관련
	ActivityTypeInvestment = "investment"
	ActionInvestmentCreate = "create"
	ActionInvestmentWithdraw = "withdraw"
	ActionInvestmentPayout = "payout"
)

// CreateActivityLogRequest 활동 로그 생성 요청
type CreateActivityLogRequest struct {
	UserID       uint             `json:"user_id" binding:"required"`
	ActivityType string           `json:"activity_type" binding:"required"`
	Action       string           `json:"action" binding:"required"`
	Description  string           `json:"description"`
	ProjectID    *uint            `json:"project_id,omitempty"`
	MilestoneID  *uint            `json:"milestone_id,omitempty"`
	OrderID      *uint            `json:"order_id,omitempty"`
	TradeID      *uint            `json:"trade_id,omitempty"`
	Metadata     ActivityMetadata `json:"metadata"`
}

// ActivityLogResponse 활동 로그 응답
type ActivityLogResponse struct {
	ID           uint             `json:"id"`
	CreatedAt    time.Time        `json:"created_at"`
	ActivityType string           `json:"activity_type"`
	Action       string           `json:"action"`
	Description  string           `json:"description"`
	ProjectID    *uint            `json:"project_id,omitempty"`
	MilestoneID  *uint            `json:"milestone_id,omitempty"`
	Metadata     ActivityMetadata `json:"metadata"`

	// 관련 엔티티 정보 (로딩된 경우)
	Project   *Project   `json:"project,omitempty"`
	Milestone *Milestone `json:"milestone,omitempty"`
}

// GetActivityLogsRequest 활동 로그 조회 요청
type GetActivityLogsRequest struct {
	UserID       uint     `json:"user_id"`
	ActivityTypes []string `json:"activity_types,omitempty"` // 필터: 특정 활동 타입들
	Limit        int      `json:"limit,omitempty"`           // 기본값: 20
	Offset       int      `json:"offset,omitempty"`          // 페이지네이션
	StartDate    *time.Time `json:"start_date,omitempty"`    // 시작 날짜
	EndDate      *time.Time `json:"end_date,omitempty"`      // 종료 날짜
}
