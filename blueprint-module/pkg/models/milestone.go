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

	// 🔍 증명 및 검증 단계
	MilestoneStatusProofSubmitted    MilestoneStatus = "proof_submitted"    // 증거 제출됨
	MilestoneStatusUnderVerification MilestoneStatus = "under_verification" // 검증 진행 중
	MilestoneStatusProofApproved     MilestoneStatus = "proof_approved"     // 증거 승인됨
	MilestoneStatusProofRejected     MilestoneStatus = "proof_rejected"     // 증거 거부됨
	MilestoneStatusDisputed          MilestoneStatus = "disputed"           // 분쟁 중

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

	// 🔍 증명 및 검증 관련 필드
	RequiresProof         bool      `json:"requires_proof" gorm:"default:true"`          // 증거 제출 필요 여부
	ProofDeadline         *time.Time `json:"proof_deadline,omitempty"`                   // 증거 제출 마감일
	VerificationDeadline  *time.Time `json:"verification_deadline,omitempty"`            // 검증 완료 마감일
	MinValidators         int       `json:"min_validators" gorm:"default:3"`             // 최소 검증인 수
	MinApprovalRate       float64   `json:"min_approval_rate" gorm:"default:0.6"`        // 최소 승인률 (60%)
	
	// 검증 통계
	TotalValidators       int       `json:"total_validators" gorm:"default:0"`           // 총 검증인 수
	ApprovalVotes         int       `json:"approval_votes" gorm:"default:0"`             // 승인 투표 수
	RejectionVotes        int       `json:"rejection_votes" gorm:"default:0"`            // 거부 투표 수
	CurrentApprovalRate   float64   `json:"current_approval_rate" gorm:"default:0"`      // 현재 승인률

	// 알림 관련
	EmailSent    bool          `json:"email_sent" gorm:"default:false"`
	ReminderSent bool          `json:"reminder_sent" gorm:"default:false"`

	// 메타데이터
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`

	// 외래키 참조
	Project Project `json:"project,omitempty" gorm:"foreignKey:ProjectID"`
	
	// 🔍 검증 관련 관계 (circular import 방지를 위해 interface{} 사용)
	// 실제 사용시에는 적절한 타입 캐스팅 필요
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

// 🔍 증명 및 검증 관련 메서드들

// CanSubmitProof 증거 제출 가능 여부
func (m *Milestone) CanSubmitProof() bool {
	return m.RequiresProof && 
		   m.Status == MilestoneStatusActive &&
		   (m.ProofDeadline == nil || time.Now().Before(*m.ProofDeadline))
}

// IsProofSubmissionExpired 증거 제출 기간 만료 여부
func (m *Milestone) IsProofSubmissionExpired() bool {
	return m.ProofDeadline != nil && time.Now().After(*m.ProofDeadline)
}

// IsVerificationExpired 검증 기간 만료 여부
func (m *Milestone) IsVerificationExpired() bool {
	return m.VerificationDeadline != nil && time.Now().After(*m.VerificationDeadline)
}

// HasSufficientValidators 충분한 검증인 수 확인
func (m *Milestone) HasSufficientValidators() bool {
	return m.TotalValidators >= m.MinValidators
}

// HasReachedApprovalThreshold 승인 임계값 도달 여부
func (m *Milestone) HasReachedApprovalThreshold() bool {
	return m.CurrentApprovalRate >= m.MinApprovalRate
}

// CanCompleteVerification 검증 완료 가능 여부
func (m *Milestone) CanCompleteVerification() bool {
	return m.HasSufficientValidators() && 
		   (m.HasReachedApprovalThreshold() || m.IsVerificationExpired())
}

// UpdateVerificationStats 검증 통계 업데이트
func (m *Milestone) UpdateVerificationStats(approvalVotes, rejectionVotes int) {
	m.ApprovalVotes = approvalVotes
	m.RejectionVotes = rejectionVotes
	m.TotalValidators = approvalVotes + rejectionVotes
	
	if m.TotalValidators > 0 {
		m.CurrentApprovalRate = float64(approvalVotes) / float64(m.TotalValidators)
	} else {
		m.CurrentApprovalRate = 0
	}
}

// StartVerificationProcess 검증 프로세스 시작
func (m *Milestone) StartVerificationProcess() {
	m.Status = MilestoneStatusUnderVerification
	if m.VerificationDeadline == nil {
		deadline := time.Now().Add(72 * time.Hour) // 72시간 후
		m.VerificationDeadline = &deadline
	}
}

// CompleteVerification 검증 완료 처리
func (m *Milestone) CompleteVerification(approved bool) {
	if approved {
		m.Status = MilestoneStatusProofApproved
		now := time.Now()
		m.CompletedAt = &now
		m.IsCompleted = true
	} else {
		m.Status = MilestoneStatusProofRejected
	}
}

// SetDisputed 분쟁 상태로 변경
func (m *Milestone) SetDisputed() {
	m.Status = MilestoneStatusDisputed
}

// SetProofDeadline 증거 제출 마감일 설정
func (m *Milestone) SetProofDeadline(days int) {
	if days > 0 {
		deadline := time.Now().AddDate(0, 0, days)
		m.ProofDeadline = &deadline
	}
}

// TableName GORM 테이블명 설정
func (Milestone) TableName() string {
	return "milestones"
}
