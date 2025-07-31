package models

import (
	"time"

	"gorm.io/gorm"
)

// 목표 카테고리
type GoalCategory string

const (
	CareerGoal   GoalCategory = "career"
	BusinessGoal GoalCategory = "business"
	EducationGoal GoalCategory = "education"
	PersonalGoal GoalCategory = "personal"
	LifeGoal     GoalCategory = "life"
)

// 목표 상태
type GoalStatus string

const (
	GoalDraft      GoalStatus = "draft"      // 초안
	GoalActive     GoalStatus = "active"     // 활성
	GoalCompleted  GoalStatus = "completed"  // 완료
	GoalCancelled  GoalStatus = "cancelled"  // 취소
	GoalOnHold     GoalStatus = "on_hold"    // 보류
)

type Goal struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	UserID      uint           `json:"user_id" gorm:"not null;index"`
	Title       string         `json:"title" gorm:"not null"`
	Description string         `json:"description" gorm:"type:text"`
	Category    GoalCategory   `json:"category" gorm:"type:varchar(20);not null"`
	Status      GoalStatus     `json:"status" gorm:"type:varchar(20);default:'draft'"`
	TargetDate  *time.Time     `json:"target_date"`
	Budget      int64          `json:"budget"`                         // 예산 (원 단위)
	Priority    int            `json:"priority" gorm:"default:1"`      // 1-5 (높을수록 우선순위 높음)
	IsPublic    bool           `json:"is_public" gorm:"default:false"` // 공개 여부
	Tags        string         `json:"tags" gorm:"type:text"`          // JSON 배열로 저장
	Metrics     string         `json:"metrics" gorm:"type:text"`       // 성공 지표 (JSON)
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`

	// 외래키 참조
	User User `json:"user,omitempty" gorm:"foreignKey:UserID"`

	// 관련 모델들
	Paths      []Path      `json:"paths,omitempty" gorm:"foreignKey:GoalID"`
	Milestones []Milestone `json:"milestones,omitempty" gorm:"foreignKey:GoalID"`
}

// 목표 생성 요청
type CreateGoalRequest struct {
	Title       string       `json:"title" binding:"required,min=3,max=200"`
	Description string       `json:"description"`
	Category    GoalCategory `json:"category" binding:"required"`
	TargetDate  *time.Time   `json:"target_date"`
	Budget      int64        `json:"budget"`
	Priority    int          `json:"priority" binding:"min=1,max=5"`
	IsPublic    bool         `json:"is_public"`
	Tags        []string     `json:"tags"`
	Metrics     string       `json:"metrics"`
}

// 목표 업데이트 요청
type UpdateGoalRequest struct {
	Title       string       `json:"title" binding:"min=3,max=200"`
	Description string       `json:"description"`
	Category    GoalCategory `json:"category"`
	Status      GoalStatus   `json:"status"`
	TargetDate  *time.Time   `json:"target_date"`
	Budget      int64        `json:"budget"`
	Priority    int          `json:"priority" binding:"min=1,max=5"`
	IsPublic    bool         `json:"is_public"`
	Tags        []string     `json:"tags"`
	Metrics     string       `json:"metrics"`
}

// 꿈과 함께 마일스톤을 생성하는 요청
type CreateGoalWithMilestonesRequest struct {
	CreateGoalRequest
	Milestones []CreateGoalMilestoneRequest `json:"milestones" binding:"max=5"`
}

// 꿈 마일스톤 생성 요청
type CreateGoalMilestoneRequest struct {
	Title       string     `json:"title" binding:"required,min=3,max=200"`
	Description string     `json:"description"`
	Order       int        `json:"order" binding:"required,min=1,max=5"`
	TargetDate  *time.Time `json:"target_date"`
}

// 마일스톤 업데이트 요청
type UpdateMilestoneRequest struct {
	Title       string     `json:"title" binding:"min=3,max=200"`
	Description string     `json:"description"`
	Status      string     `json:"status"`
	TargetDate  *time.Time `json:"target_date"`
	Evidence    string     `json:"evidence"`
	Notes       string     `json:"notes"`
}
