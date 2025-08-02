package models

import (
	"time"

	"gorm.io/gorm"
)

// 마일스톤 상태
type MilestoneStatus string

const (
	MilestoneStatusPending   MilestoneStatus = "pending"   // 대기중
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

	// 상태 정보
	Status      MilestoneStatus `json:"status" gorm:"type:varchar(20);default:'pending'"`
	IsCompleted bool           `json:"is_completed" gorm:"default:false"`

	// 베팅 관련 (새로 추가)
	BettingType    string         `json:"betting_type" gorm:"type:varchar(20);default:'simple'"` // simple, custom
	BettingOptions string         `json:"betting_options" gorm:"type:text"`                      // JSON 배열

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

// TableName GORM 테이블명 설정
func (Milestone) TableName() string {
	return "milestones"
}
