package models

import (
	"time"

	"gorm.io/gorm"
)

// 🧭 멘토링 시스템 - "Wisdom Market" 데이터 모델들

// MentorStatus 멘토 상태
type MentorStatus string

const (
	MentorStatusActive    MentorStatus = "active"    // 활성 멘토
	MentorStatusInactive  MentorStatus = "inactive"  // 비활성 (휴면)
	MentorStatusSuspended MentorStatus = "suspended" // 정지됨
	MentorStatusVerified  MentorStatus = "verified"  // 검증된 멘토
)

// MentorTier 멘토 등급
type MentorTier string

const (
	MentorTierBronze   MentorTier = "bronze"   // 초급 멘토
	MentorTierSilver   MentorTier = "silver"   // 중급 멘토
	MentorTierGold     MentorTier = "gold"     // 고급 멘토
	MentorTierPlatinum MentorTier = "platinum" // 최상급 멘토
	MentorTierLegend   MentorTier = "legend"   // 전설적 멘토
)

// Mentor 멘토 프로필 및 평판
type Mentor struct {
	ID           uint         `json:"id" gorm:"primaryKey"`
	UserID       uint         `json:"user_id" gorm:"not null;index"`
	Status       MentorStatus `json:"status" gorm:"type:varchar(20);default:'active'"`
	Tier         MentorTier   `json:"tier" gorm:"type:varchar(20);default:'bronze'"`

	// 전문 분야 및 경력
	Expertise      []string `json:"expertise" gorm:"type:text;serializer:json"`       // 전문 분야
	Industries     []string `json:"industries" gorm:"type:text;serializer:json"`      // 산업 분야
	YearsExperience int     `json:"years_experience" gorm:"default:0"`                // 경력 연수
	Bio            string   `json:"bio" gorm:"type:text"`                             // 자기 소개
	LinkedInURL    string   `json:"linkedin_url"`                                     // LinkedIn 프로필
	PersonalURL    string   `json:"personal_url"`                                     // 개인 웹사이트

	// 멘토링 통계 (실시간 계산됨)
	TotalMentorings     int     `json:"total_mentorings" gorm:"default:0"`         // 총 멘토링 횟수
	SuccessfulMentorings int    `json:"successful_mentorings" gorm:"default:0"`    // 성공한 멘토링
	SuccessRate         float64 `json:"success_rate" gorm:"default:0"`             // 성공률 (%)
	TotalBettingAmount  int64   `json:"total_betting_amount" gorm:"default:0"`     // 총 베팅 금액 (센트)
	TotalEarnedAmount   int64   `json:"total_earned_amount" gorm:"default:0"`      // 총 획득 금액 (센트)
	AverageRating       float64 `json:"average_rating" gorm:"default:0"`           // 평균 평점

	// 평판 점수 (온체인 기록용)
	ReputationScore     int     `json:"reputation_score" gorm:"default:0"`         // 평판 점수
	TrustScore          float64 `json:"trust_score" gorm:"default:0"`              // 신뢰도 점수

	// 설정
	IsAvailable         bool    `json:"is_available" gorm:"default:true"`          // 멘토링 가능 여부
	MaxActiveMentorings int     `json:"max_active_mentorings" gorm:"default:5"`    // 최대 동시 멘토링 수
	PreferredCategories []ProjectCategory `json:"preferred_categories" gorm:"type:text;serializer:json"` // 선호 프로젝트 카테고리

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// 관계
	User            User              `json:"user,omitempty" gorm:"foreignKey:UserID"`
	MentorMilestones []MentorMilestone `json:"mentor_milestones,omitempty" gorm:"foreignKey:MentorID"`
	MentoringSessions []MentoringSession `json:"mentoring_sessions,omitempty" gorm:"foreignKey:MentorID"`
}

// MentorMilestone 특정 마일스톤에 대한 멘토의 베팅 및 자격 정보
type MentorMilestone struct {
	ID          uint `json:"id" gorm:"primaryKey"`
	MentorID    uint `json:"mentor_id" gorm:"not null;index"`
	MilestoneID uint `json:"milestone_id" gorm:"not null;index"`
	ProjectID   uint `json:"project_id" gorm:"not null;index"`

	// 베팅 정보 (Proof of Confidence)
	TotalBetAmount    int64   `json:"total_bet_amount" gorm:"not null"`           // 총 베팅 금액
	BetSharePercentage float64 `json:"bet_share_percentage" gorm:"not null"`      // 해당 마일스톤에서의 베팅 비중 (%)
	IsLeadMentor      bool    `json:"is_lead_mentor" gorm:"default:false"`       // 리드 멘토 여부
	LeadMentorRank    int     `json:"lead_mentor_rank" gorm:"default:0"`         // 리드 멘토 순위 (1,2,3...)

	// 멘토링 상태
	IsActive           bool      `json:"is_active" gorm:"default:false"`           // 활성 멘토링 여부
	StartedAt          *time.Time `json:"started_at,omitempty"`                   // 멘토링 시작일
	LastActivityAt     *time.Time `json:"last_activity_at,omitempty"`             // 마지막 활동일
	MentoringEndedAt   *time.Time `json:"mentoring_ended_at,omitempty"`           // 멘토링 종료일

	// 성과 및 보상
	ActionsCount       int     `json:"actions_count" gorm:"default:0"`            // 수행한 멘토링 액션 수
	MenteeRating       float64 `json:"mentee_rating" gorm:"default:0"`            // 멘티(진행자) 평점
	EarnedFromBetting  int64   `json:"earned_from_betting" gorm:"default:0"`      // 베팅 수익
	EarnedFromMentoring int64  `json:"earned_from_mentoring" gorm:"default:0"`    // 멘토 풀 보상

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// 관계
	Mentor    Mentor    `json:"mentor,omitempty" gorm:"foreignKey:MentorID"`
	Milestone Milestone `json:"milestone,omitempty" gorm:"foreignKey:MilestoneID"`
	Project   Project   `json:"project,omitempty" gorm:"foreignKey:ProjectID"`
}

// MentoringSessionStatus 멘토링 세션 상태
type MentoringSessionStatus string

const (
	SessionStatusActive    MentoringSessionStatus = "active"    // 진행 중
	SessionStatusCompleted MentoringSessionStatus = "completed" // 완료
	SessionStatusCancelled MentoringSessionStatus = "cancelled" // 취소
	SessionStatusPaused    MentoringSessionStatus = "paused"    // 일시정지
)

// MentoringSession 멘토-진행자 간의 멘토링 세션
type MentoringSession struct {
	ID          uint                   `json:"id" gorm:"primaryKey"`
	MentorID    uint                   `json:"mentor_id" gorm:"not null;index"`
	MenteeID    uint                   `json:"mentee_id" gorm:"not null;index"` // 프로젝트 진행자
	MilestoneID uint                   `json:"milestone_id" gorm:"not null;index"`
	ProjectID   uint                   `json:"project_id" gorm:"not null;index"`
	Status      MentoringSessionStatus `json:"status" gorm:"type:varchar(20);default:'active'"`

	// 세션 정보
	Title           string     `json:"title" gorm:"not null"`
	Description     string     `json:"description" gorm:"type:text"`
	StartedAt       time.Time  `json:"started_at"`
	CompletedAt     *time.Time `json:"completed_at,omitempty"`
	LastMessageAt   *time.Time `json:"last_message_at,omitempty"`

	// 성과 지표
	MessagesCount   int     `json:"messages_count" gorm:"default:0"`        // 메시지 수
	ActionsCount    int     `json:"actions_count" gorm:"default:0"`         // 액션 수
	FilesShared     int     `json:"files_shared" gorm:"default:0"`          // 공유된 파일 수
	MeetingsHeld    int     `json:"meetings_held" gorm:"default:0"`         // 진행된 미팅 수

	// 평가
	MenteeRating    float64 `json:"mentee_rating" gorm:"default:0"`         // 멘티의 멘토 평가
	MentorRating    float64 `json:"mentor_rating" gorm:"default:0"`         // 멘토의 멘티 평가
	MenteeReview    string  `json:"mentee_review" gorm:"type:text"`         // 멘티 후기
	MentorReview    string  `json:"mentor_review" gorm:"type:text"`         // 멘토 후기

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// 관계
	Mentor    Mentor    `json:"mentor,omitempty" gorm:"foreignKey:MentorID"`
	Mentee    User      `json:"mentee,omitempty" gorm:"foreignKey:MenteeID"`
	Milestone Milestone `json:"milestone,omitempty" gorm:"foreignKey:MilestoneID"`
	Project   Project   `json:"project,omitempty" gorm:"foreignKey:ProjectID"`
	Actions   []MentorAction `json:"actions,omitempty" gorm:"foreignKey:SessionID"`
}

// MentorActionType 멘토링 액션 타입
type MentorActionType string

const (
	ActionTypeTaskProposal   MentorActionType = "task_proposal"   // 핵심 과제 제안
	ActionTypeFeedback       MentorActionType = "feedback"        // 피드백 제출
	ActionTypeAdvice         MentorActionType = "advice"          // 조언 제공
	ActionTypeResourceShare  MentorActionType = "resource_share"  // 리소스 공유
	ActionTypeMeetingRequest MentorActionType = "meeting_request" // 미팅 요청
	ActionTypeProgressCheck  MentorActionType = "progress_check"  // 진행상황 점검
)

// MentorActionStatus 멘토링 액션 상태
type MentorActionStatus string

const (
	ActionStatusProposed   MentorActionStatus = "proposed"   // 제안됨
	ActionStatusAccepted   MentorActionStatus = "accepted"   // 수락됨
	ActionStatusRejected   MentorActionStatus = "rejected"   // 거절됨
	ActionStatusInProgress MentorActionStatus = "in_progress" // 진행 중
	ActionStatusCompleted  MentorActionStatus = "completed"  // 완료됨
)

// MentorAction 구체적인 멘토링 액션들
type MentorAction struct {
	ID        uint               `json:"id" gorm:"primaryKey"`
	SessionID uint               `json:"session_id" gorm:"not null;index"`
	MentorID  uint               `json:"mentor_id" gorm:"not null;index"`
	MenteeID  uint               `json:"mentee_id" gorm:"not null;index"`
	Type      MentorActionType   `json:"type" gorm:"not null"`
	Status    MentorActionStatus `json:"status" gorm:"type:varchar(20);default:'proposed'"`

	// 액션 내용
	Title         string    `json:"title" gorm:"not null"`
	Description   string    `json:"description" gorm:"type:text"`
	Content       string    `json:"content" gorm:"type:text"`              // JSON 형태의 추가 데이터
	DueDate       *time.Time `json:"due_date,omitempty"`                   // 마감일 (과제의 경우)
	Priority      int       `json:"priority" gorm:"default:3"`             // 우선순위 (1-5)

	// 응답 및 결과
	MenteeResponse string    `json:"mentee_response" gorm:"type:text"`      // 멘티 응답
	ResultFiles    []string  `json:"result_files" gorm:"type:text;serializer:json"` // 결과 파일들
	CompletedAt    *time.Time `json:"completed_at,omitempty"`               // 완료일

	// 평가
	Effectiveness  float64   `json:"effectiveness" gorm:"default:0"`        // 효과성 평가
	MenteeRating   float64   `json:"mentee_rating" gorm:"default:0"`        // 멘티 평가

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// 관계
	Session MentoringSession `json:"session,omitempty" gorm:"foreignKey:SessionID"`
	Mentor  Mentor           `json:"mentor,omitempty" gorm:"foreignKey:MentorID"`
	Mentee  User             `json:"mentee,omitempty" gorm:"foreignKey:MenteeID"`
}

// MentorPool 마일스톤별 멘토 보상 풀
type MentorPool struct {
	ID          uint `json:"id" gorm:"primaryKey"`
	MilestoneID uint `json:"milestone_id" gorm:"not null;uniqueIndex"`
	ProjectID   uint `json:"project_id" gorm:"not null;index"`

	// 풀 정보
	TotalPoolAmount       int64   `json:"total_pool_amount" gorm:"default:0"`       // 총 풀 금액 (센트)
	AccumulatedFees       int64   `json:"accumulated_fees" gorm:"default:0"`        // 누적 수수료
	FeePercentage         float64 `json:"fee_percentage" gorm:"default:50"`         // 거래 수수료 중 풀로 이동하는 비율 (%)

	// 분배 정보
	IsDistributed         bool      `json:"is_distributed" gorm:"default:false"`     // 분배 완료 여부
	DistributedAmount     int64     `json:"distributed_amount" gorm:"default:0"`     // 분배된 금액
	DistributedAt         *time.Time `json:"distributed_at,omitempty"`               // 분배 완료일
	EligibleMentorsCount  int       `json:"eligible_mentors_count" gorm:"default:0"` // 자격있는 멘토 수

	// 분배 방식 설정
	SimpleDistribution    bool    `json:"simple_distribution" gorm:"default:false"`  // 단순 분배 (베팅액 비례)
	PerformanceWeighted   bool    `json:"performance_weighted" gorm:"default:true"`  // 성과 기반 분배
	MentorRatingWeight    float64 `json:"mentor_rating_weight" gorm:"default:30"`    // 멘토 평점 가중치 (%)
	BettingAmountWeight   float64 `json:"betting_amount_weight" gorm:"default:70"`   // 베팅액 가중치 (%)

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// 관계
	Milestone Milestone `json:"milestone,omitempty" gorm:"foreignKey:MilestoneID"`
	Project   Project   `json:"project,omitempty" gorm:"foreignKey:ProjectID"`
}

// MentorReputation 온체인 평판 기록
type MentorReputation struct {
	ID         uint   `json:"id" gorm:"primaryKey"`
	MentorID   uint   `json:"mentor_id" gorm:"not null;index"`

	// 평판 이벤트
	EventType  string `json:"event_type" gorm:"not null"`        // "successful_mentoring", "milestone_success", "high_rating"
	Points     int    `json:"points" gorm:"not null"`            // 획득/차감 점수
	Multiplier float64 `json:"multiplier" gorm:"default:1"`      // 점수 배율

	// 관련 정보
	MilestoneID  *uint   `json:"milestone_id,omitempty" gorm:"index"`
	ProjectID    *uint   `json:"project_id,omitempty" gorm:"index"`
	SessionID    *uint   `json:"session_id,omitempty" gorm:"index"`
	Description  string  `json:"description" gorm:"type:text"`

	// 블록체인 기록 (추후 구현)
	TxHash       string  `json:"tx_hash"`                       // 트랜잭션 해시
	BlockNumber  uint64  `json:"block_number" gorm:"default:0"` // 블록 번호
	IsOnChain    bool    `json:"is_on_chain" gorm:"default:false"` // 온체인 기록 여부

	CreatedAt time.Time `json:"created_at"`

	// 관계
	Mentor Mentor `json:"mentor,omitempty" gorm:"foreignKey:MentorID"`
}

// TableName GORM 테이블명 설정들
func (Mentor) TableName() string           { return "mentors" }
func (MentorMilestone) TableName() string { return "mentor_milestones" }
func (MentoringSession) TableName() string { return "mentoring_sessions" }
func (MentorAction) TableName() string     { return "mentor_actions" }
func (MentorPool) TableName() string       { return "mentor_pools" }
func (MentorReputation) TableName() string { return "mentor_reputations" }

// 🚀 Helper 메서드들

// CalculateSuccessRate 성공률 계산
func (m *Mentor) CalculateSuccessRate() float64 {
	if m.TotalMentorings <= 0 {
		return 0
	}
	return (float64(m.SuccessfulMentorings) / float64(m.TotalMentorings)) * 100
}

// IsQualifiedForTier 특정 등급 자격 확인
func (m *Mentor) IsQualifiedForTier(tier MentorTier) bool {
	switch tier {
	case MentorTierSilver:
		return m.SuccessfulMentorings >= 5 && m.SuccessRate >= 70 && m.ReputationScore >= 100
	case MentorTierGold:
		return m.SuccessfulMentorings >= 15 && m.SuccessRate >= 80 && m.ReputationScore >= 500
	case MentorTierPlatinum:
		return m.SuccessfulMentorings >= 30 && m.SuccessRate >= 90 && m.ReputationScore >= 1500
	case MentorTierLegend:
		return m.SuccessfulMentorings >= 100 && m.SuccessRate >= 95 && m.ReputationScore >= 5000
	default:
		return true // Bronze는 누구나
	}
}

// CanTakeNewMentoring 새로운 멘토링 가능 여부
func (m *Mentor) CanTakeNewMentoring() bool {
	return m.IsAvailable && m.Status == MentorStatusActive
}

// CalculateLeadMentorRank 리드 멘토 순위 계산 (베팅액 기준)
func (mm *MentorMilestone) CalculateLeadMentorRank() int {
	// 이 로직은 서비스 레이어에서 구현될 예정
	return mm.LeadMentorRank
}

// IsEligibleForReward 보상 자격 확인
func (mm *MentorMilestone) IsEligibleForReward() bool {
	return mm.IsActive && mm.ActionsCount > 0 && mm.MenteeRating > 0
}