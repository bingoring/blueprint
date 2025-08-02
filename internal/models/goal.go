package models

import (
	"time"

	"gorm.io/gorm"
)

// 프로젝트 카테고리
type ProjectCategory string

const (
	CareerProject   ProjectCategory = "career"
	BusinessProject ProjectCategory = "business"
	EducationProject ProjectCategory = "education"
	PersonalProject ProjectCategory = "personal"
	LifeProject     ProjectCategory = "life"
)

// 프로젝트 상태
type ProjectStatus string

const (
	ProjectDraft      ProjectStatus = "draft"      // 초안
	ProjectActive     ProjectStatus = "active"     // 활성
	ProjectCompleted  ProjectStatus = "completed"  // 완료
	ProjectCancelled  ProjectStatus = "cancelled"  // 취소
	ProjectOnHold     ProjectStatus = "on_hold"    // 보류
)

type Project struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	UserID      uint           `json:"user_id" gorm:"not null;index"`
	Title       string         `json:"title" gorm:"not null"`
	Description string         `json:"description" gorm:"type:text"`
	Category    ProjectCategory `json:"category" gorm:"type:varchar(20);not null"`
	Status      ProjectStatus  `json:"status" gorm:"type:varchar(20);default:'draft'"`
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
	Milestones  []Milestone  `json:"milestones,omitempty" gorm:"foreignKey:ProjectID"`
}

// TableName GORM 테이블명 설정
func (Project) TableName() string {
	return "projects"
}

// 프로젝트 생성 요청
type CreateProjectRequest struct {
	Title       string          `json:"title" binding:"required,min=3,max=200"`
	Description string          `json:"description"`
	Category    ProjectCategory `json:"category" binding:"required"`
	TargetDate  *time.Time      `json:"target_date"`
	Budget      int64           `json:"budget"`
	Priority    int             `json:"priority" binding:"min=1,max=5"`
	IsPublic    bool            `json:"is_public"`
	Tags        []string        `json:"tags"`
	Metrics     string          `json:"metrics"`
}

// 프로젝트 업데이트 요청
type UpdateProjectRequest struct {
	Title       string          `json:"title" binding:"min=3,max=200"`
	Description string          `json:"description"`
	Category    ProjectCategory `json:"category"`
	Status      ProjectStatus   `json:"status"`
	TargetDate  *time.Time      `json:"target_date"`
	Budget      int64           `json:"budget"`
	Priority    int             `json:"priority" binding:"min=1,max=5"`
	IsPublic    bool            `json:"is_public"`
	Tags        []string        `json:"tags"`
	Metrics     string          `json:"metrics"`
}

// 프로젝트와 함께 마일스톤을 생성하는 요청
type CreateProjectWithMilestonesRequest struct {
	CreateProjectRequest
	Milestones []CreateProjectMilestoneRequest `json:"milestones" binding:"max=5"`
}

// 프로젝트 마일스톤 생성 요청
type CreateProjectMilestoneRequest struct {
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

// Goal 관련 호환성 코드 제거 완료
// Path 모델도 제거 예정 (예전 워딩)
// 이제 Project -> Milestone 구조로 단순화
