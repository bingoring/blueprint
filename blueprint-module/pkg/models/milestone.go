package models

import (
	"time"

	"gorm.io/gorm"
)

// 마일스톤 상태 - 시장성 검증 시스템 지원
type MilestoneStatus string

const (
	// 🆕 Proposal & Funding Phase
	MilestoneStatusProposal  MilestoneStatus = "proposal"  // 제안 단계
	MilestoneStatusFunding   MilestoneStatus = "funding"   // 펀딩 진행 중
	MilestoneStatusActive    MilestoneStatus = "active"    // 펀딩 성공, 활성화됨
	MilestoneStatusRejected  MilestoneStatus = "rejected"  // 펀딩 실패, 자동 폐기

	// 기존 진행 상태들
	MilestoneStatusPending   MilestoneStatus = "pending"   // 대기중 (구버전 호환)
	MilestoneStatusCompleted MilestoneStatus = "completed" // 완료
	MilestoneStatusFailed    MilestoneStatus = "failed"    // 실패
	MilestoneStatusCancelled MilestoneStatus = "cancelled" // 취소
)

// 마일스톤 모델 (Project와 직접 연결, Path 제거)
type Milestone struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	ProjectID   uint           `json:"project_id" gorm:"not null;index"`

	// 마일스톤 정보
	Title       string         `json:"title" gorm:"not null;size:255"`
	Description string         `json:"description" gorm:"type:text"`
	Order       int            `json:"order" gorm:"not null;default:1"`   // 순서 (1-5)

	// 날짜 정보
	TargetDate  *time.Time     `json:"target_date"`
	CompletedAt *time.Time     `json:"completed_at"`

	// 🆕 펀딩 및 시장성 검증 관련
	FundingStartDate  *time.Time `json:"funding_start_date,omitempty"`   // 펀딩 시작일
	FundingEndDate    *time.Time `json:"funding_end_date,omitempty"`     // 펀딩 마감일
	FundingDuration   int        `json:"funding_duration" gorm:"default:5"` // 펀딩 기간 (일수)
	MinViableCapital  int64      `json:"min_viable_capital" gorm:"default:100000"` // 최소 목표 금액 (센트)
	CurrentTVL        int64      `json:"current_tvl" gorm:"default:0"`    // 현재 총 베팅액 (센트)
	FundingProgress   float64    `json:"funding_progress" gorm:"default:0"` // 펀딩 진행률 (0-1)

	// 상태 정보 (기본값을 proposal로 변경)
	Status      MilestoneStatus `json:"status" gorm:"type:varchar(20);default:'proposal'"`
	IsCompleted bool           `json:"is_completed" gorm:"default:false"`

	// 베팅 관련 (새로 추가)
	BettingType    string   `json:"betting_type" gorm:"type:varchar(20);default:'simple'"` // simple, custom
	BettingOptions []string `json:"betting_options" gorm:"type:text;serializer:json"`      // JSON 배열

	// 응원 (베팅) 관련
	TotalSupport       int64   `json:"total_support" gorm:"default:0"`
	SupporterCount     int     `json:"supporter_count" gorm:"default:0"`
	SuccessProbability float64 `json:"success_probability" gorm:"default:0"`

	// 증빙 및 노트
	Evidence    string         `json:"evidence" gorm:"type:text"`
	Notes       string         `json:"notes" gorm:"type:text"`

	// 알림 관련
	EmailSent    bool          `json:"email_sent" gorm:"default:false"`
	ReminderSent bool          `json:"reminder_sent" gorm:"default:false"`

	// 메타데이터
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`

	// 외래키 참조
	Project Project `json:"project,omitempty" gorm:"foreignKey:ProjectID"`
}

// 🆕 펀딩 검증 관련 메서드들
func (m *Milestone) IsFundingActive() bool {
	return m.Status == MilestoneStatusFunding &&
		   m.FundingEndDate != nil &&
		   time.Now().Before(*m.FundingEndDate)
}

func (m *Milestone) HasReachedMinViableCapital() bool {
	return m.CurrentTVL >= m.MinViableCapital
}

func (m *Milestone) IsFundingExpired() bool {
	return m.Status == MilestoneStatusFunding &&
		   m.FundingEndDate != nil &&
		   time.Now().After(*m.FundingEndDate)
}

func (m *Milestone) CalculateFundingProgress() float64 {
	if m.MinViableCapital <= 0 {
		return 0
	}
	progress := float64(m.CurrentTVL) / float64(m.MinViableCapital)
	if progress > 1.0 {
		progress = 1.0
	}
	return progress
}

// StartFundingPhase 펀딩 단계 시작
func (m *Milestone) StartFundingPhase() {
	m.Status = MilestoneStatusFunding
	now := time.Now()
	m.FundingStartDate = &now
	fundingEnd := now.AddDate(0, 0, m.FundingDuration)
	m.FundingEndDate = &fundingEnd
}

// TableName GORM 테이블명 설정
func (Milestone) TableName() string {
	return "milestones"
}
