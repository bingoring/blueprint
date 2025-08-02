package models

import (
	"time"

	"gorm.io/gorm"
)

// 경로 상태
type PathStatus string

const (
	PathPending   PathStatus = "pending"   // 대기중
	PathActive    PathStatus = "active"    // 활성
	PathCompleted PathStatus = "completed" // 완료
	PathFailed    PathStatus = "failed"    // 실패
)

// 마일스톤 상태
type MilestoneStatus string

const (
	MilestoneStatusPending   MilestoneStatus = "pending"   // 대기중
	MilestoneStatusCompleted MilestoneStatus = "completed" // 완료
	MilestoneStatusFailed    MilestoneStatus = "failed"    // 실패
	MilestoneStatusCancelled MilestoneStatus = "cancelled" // 취소
)

type Path struct {
	ID              uint           `json:"id" gorm:"primaryKey"`
	ProjectID       uint           `json:"project_id" gorm:"not null;index"`
	Title           string         `json:"title" gorm:"not null"`
	Description     string         `json:"description" gorm:"type:text"`
	Status          PathStatus     `json:"status" gorm:"type:varchar(20);default:'pending'"`
	EstimatedDays   int            `json:"estimated_days"`                        // 예상 소요일
	EstimatedCost   int64          `json:"estimated_cost"`                        // 예상 비용
	DifficultyLevel int            `json:"difficulty_level" gorm:"default:1"`     // 1-5 (어려움 정도)
	SuccessRate     float64        `json:"success_rate" gorm:"default:0.5"`       // 성공 확률 (0.0-1.0)
	Requirements    string         `json:"requirements" gorm:"type:text"`         // 필요 조건 (JSON)
	Steps           string         `json:"steps" gorm:"type:text"`                // 단계별 계획 (JSON)
	Resources       string         `json:"resources" gorm:"type:text"`            // 필요 자원 (JSON)
	Risks           string         `json:"risks" gorm:"type:text"`                // 위험 요소 (JSON)
	CreatedBy       uint           `json:"created_by" gorm:"index"`               // 경로를 제안한 사용자 ID
	IsSelected      bool           `json:"is_selected" gorm:"default:false"`      // 선택된 경로인지
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `json:"-" gorm:"index"`

	// 외래키 참조
	Project   Project `json:"project,omitempty" gorm:"foreignKey:ProjectID"`
	Creator   User    `json:"creator,omitempty" gorm:"foreignKey:CreatedBy"`

	// 관련 모델들
	Predictions []PathPrediction `json:"predictions,omitempty" gorm:"foreignKey:PathID"`
	Milestones  []Milestone      `json:"milestones,omitempty" gorm:"foreignKey:PathID"`
}

// 경로 예측 (전문가들이 하는 베팅)
type PathPrediction struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	PathID       uint      `json:"path_id" gorm:"not null;index"`
	ExpertID     uint      `json:"expert_id" gorm:"not null;index"`
	Probability  float64   `json:"probability" gorm:"not null"`       // 성공 확률 예측 (0.0-1.0)
	StakeAmount  int64     `json:"stake_amount" gorm:"not null"`      // 베팅 금액 (토큰)
	Reasoning    string    `json:"reasoning" gorm:"type:text"`        // 예측 근거
	IsCorrect    *bool     `json:"is_correct"`                        // 예측이 맞았는지 (결과 확정 후)
	RewardAmount int64     `json:"reward_amount" gorm:"default:0"`    // 받은 보상 (토큰)
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	// 외래키 참조
	Path   Path `json:"path,omitempty" gorm:"foreignKey:PathID"`
	Expert User `json:"expert,omitempty" gorm:"foreignKey:ExpertID"`
}

// 마일스톤 (프로젝트의 중간 단계 또는 경로의 체크포인트)
type Milestone struct {
	ID          uint           `json:"id" gorm:"primaryKey"`

	// 연결 관계 (Project 직접 연결 또는 Path를 통한 연결)
	ProjectID   *uint          `json:"project_id" gorm:"index"`      // 프로젝트에 직접 연결된 마일스톤
	PathID      *uint          `json:"path_id" gorm:"index"`         // 경로를 통한 마일스톤

	// 마일스톤 정보
	Title       string         `json:"title" gorm:"not null;size:255"`
	Description string         `json:"description" gorm:"type:text"`
	Order       int            `json:"order" gorm:"not null;default:1"`   // 순서 (1-5)

	// 날짜 정보
	TargetDate  *time.Time     `json:"target_date"`               // 목표 날짜
	CompletedAt *time.Time     `json:"completed_at"`

	// 상태 정보
	Status      string         `json:"status" gorm:"default:'pending'"` // pending, completed, failed, cancelled
	IsCompleted bool           `json:"is_completed" gorm:"default:false"`

	// 응원 (베팅) 관련
	TotalSupport       int64   `json:"total_support" gorm:"default:0"`        // 총 응원금
	SupporterCount     int     `json:"supporter_count" gorm:"default:0"`      // 응원자 수
	SuccessProbability float64 `json:"success_probability" gorm:"default:0"`   // 성공 확률 (0-1)

	// 증빙 및 노트
	Evidence    string         `json:"evidence" gorm:"type:text"`         // 완료 증빙 (JSON)
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
	Path    Path    `json:"path,omitempty" gorm:"foreignKey:PathID"`
}

// TableName GORM 테이블명 설정
func (Milestone) TableName() string {
	return "milestones"
}

// 경로 생성 요청
type CreatePathRequest struct {
	Title           string  `json:"title" binding:"required,min=3,max=200"`
	Description     string  `json:"description"`
	EstimatedDays   int     `json:"estimated_days" binding:"min=1"`
	EstimatedCost   int64   `json:"estimated_cost"`
	DifficultyLevel int     `json:"difficulty_level" binding:"min=1,max=5"`
	Requirements    string  `json:"requirements"`
	Steps           string  `json:"steps"`
	Resources       string  `json:"resources"`
	Risks           string  `json:"risks"`
}

// 경로 예측 요청
type CreatePredictionRequest struct {
	Probability float64 `json:"probability" binding:"required,min=0,max=1"`
	StakeAmount int64   `json:"stake_amount" binding:"required,min=1"`
	Reasoning   string  `json:"reasoning" binding:"required,min=10"`
}

// 마일스톤 생성 요청
type CreateMilestoneRequest struct {
	Title       string     `json:"title" binding:"required,min=3,max=200"`
	Description string     `json:"description"`
	Order       int        `json:"order" binding:"required,min=1"`
	DueDate     *time.Time `json:"due_date"`
}

// 마일스톤 완료 요청
type CompleteMilestoneRequest struct {
	Evidence string `json:"evidence"`
	Notes    string `json:"notes"`
}
