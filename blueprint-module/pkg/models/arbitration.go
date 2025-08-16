package models

import (
	"time"
)

// 🏛️ 탈중앙화된 분쟁 해결 시스템 (Kleros/Aragon Court 스타일)

// ArbitrationCase 분쟁 사건
type ArbitrationCase struct {
	ID          uint   `json:"id" gorm:"primaryKey"`
	CaseNumber  string `json:"case_number" gorm:"unique;not null"` // ACC-2024-0001 형태
	
	// 분쟁 당사자
	PlaintiffID  uint `json:"plaintiff_id" gorm:"not null;index"`  // 신청인 (멘티 또는 베팅 참여자)
	DefendantID  uint `json:"defendant_id" gorm:"not null;index"`  // 피신청인 (멘토 또는 프로젝트 소유자)
	
	// 분쟁 대상
	DisputeType    ArbitrationDisputeType `json:"dispute_type" gorm:"not null"`
	MilestoneID    *uint                  `json:"milestone_id,omitempty" gorm:"index"`    // 마일스톤 관련 분쟁
	MentorshipID   *uint                  `json:"mentorship_id,omitempty" gorm:"index"`   // 멘토링 관련 분쟁
	TradeID        *uint                  `json:"trade_id,omitempty" gorm:"index"`        // 거래 관련 분쟁
	
	// 분쟁 내용
	Title       string `json:"title" gorm:"not null"`
	Description string `json:"description" gorm:"type:text;not null"`
	Evidence    string `json:"evidence" gorm:"type:text"`           // 증거 자료 (JSON 형태)
	ClaimedAmount int64 `json:"claimed_amount"`                     // 청구 금액 (BLUEPRINT/USDC)
	
	// 분쟁 상태
	Status      ArbitrationStatus `json:"status" gorm:"default:'submitted'"`
	Priority    ArbitrationPriority `json:"priority" gorm:"default:'normal'"`
	
	// 스테이킹 (분쟁 제기 비용)
	StakeAmount     int64 `json:"stake_amount" gorm:"not null"`      // 분쟁 제기시 스테이킹 금액
	StakeReturned   bool  `json:"stake_returned" gorm:"default:false"`
	
	// 배심원단 구성
	RequiredJurors    int       `json:"required_jurors" gorm:"default:5"`    // 필요한 배심원 수
	SelectedJurors    []uint    `json:"selected_jurors" gorm:"type:jsonb"`   // 선정된 배심원 ID 목록
	JuryFormationDeadline time.Time `json:"jury_formation_deadline"`          // 배심원단 구성 마감일
	
	// 심리 과정
	VotingStarted    bool       `json:"voting_started" gorm:"default:false"`
	VotingDeadline   *time.Time `json:"voting_deadline"`                     // 투표 마감일
	RevealDeadline   *time.Time `json:"reveal_deadline"`                     // 투표 공개 마감일
	
	// 최종 결과
	Decision        ArbitrationDecision `json:"decision"`                     // 최종 판결
	DecisionReason  string             `json:"decision_reason" gorm:"type:text"` // 판결 이유
	AwardAmount     int64              `json:"award_amount"`                 // 배상 금액
	
	// 타임스탬프
	SubmittedAt time.Time  `json:"submitted_at" gorm:"default:CURRENT_TIMESTAMP"`
	DecidedAt   *time.Time `json:"decided_at"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`

	// 관계
	Plaintiff User `json:"plaintiff,omitempty" gorm:"foreignKey:PlaintiffID"`
	Defendant User `json:"defendant,omitempty" gorm:"foreignKey:DefendantID"`
	Milestone *Milestone `json:"milestone,omitempty" gorm:"foreignKey:MilestoneID"`
	Votes     []ArbitrationVote `json:"votes,omitempty" gorm:"foreignKey:CaseID"`
}

func (ArbitrationCase) TableName() string {
	return "arbitration_cases"
}

// ArbitrationDisputeType 분쟁 유형
type ArbitrationDisputeType string

const (
	DisputeTypeMilestoneCompletion ArbitrationDisputeType = "milestone_completion" // 마일스톤 완료 여부
	DisputeTypeMentorMalpractice   ArbitrationDisputeType = "mentor_malpractice"   // 멘토 직무유기/부정행위
	DisputeTypeProjectFraud        ArbitrationDisputeType = "project_fraud"        // 프로젝트 사기
	DisputeTypePaymentIssue        ArbitrationDisputeType = "payment_issue"        // 결제 문제
	DisputeTypeIntellectualProperty ArbitrationDisputeType = "intellectual_property" // 지적재산권 침해
	DisputeTypeContractBreach      ArbitrationDisputeType = "contract_breach"      // 계약 위반
)

// ArbitrationStatus 분쟁 상태
type ArbitrationStatus string

const (
	ArbitrationStatusSubmitted     ArbitrationStatus = "submitted"      // 제출됨
	ArbitrationStatusUnderReview   ArbitrationStatus = "under_review"   // 검토 중
	ArbitrationStatusJurySelection ArbitrationStatus = "jury_selection" // 배심원 선정 중
	ArbitrationStatusEvidence      ArbitrationStatus = "evidence"       // 증거 제출 기간
	ArbitrationStatusVoting        ArbitrationStatus = "voting"         // 투표 진행 중
	ArbitrationStatusReveal        ArbitrationStatus = "reveal"         // 투표 공개 중
	ArbitrationStatusDecided       ArbitrationStatus = "decided"        // 판결 완료
	ArbitrationStatusAppealed      ArbitrationStatus = "appealed"       // 항소 중
	ArbitrationStatusClosed        ArbitrationStatus = "closed"         // 종료
	ArbitrationStatusRejected      ArbitrationStatus = "rejected"       // 기각됨
)

// ArbitrationPriority 분쟁 우선순위
type ArbitrationPriority string

const (
	ArbitrationPriorityLow    ArbitrationPriority = "low"
	ArbitrationPriorityNormal ArbitrationPriority = "normal"
	ArbitrationPriorityHigh   ArbitrationPriority = "high"
	ArbitrationPriorityUrgent ArbitrationPriority = "urgent"
)

// ArbitrationDecision 분쟁 판결
type ArbitrationDecision string

const (
	ArbitrationDecisionPlaintiffWins ArbitrationDecision = "plaintiff_wins" // 신청인 승리
	ArbitrationDecisionDefendantWins ArbitrationDecision = "defendant_wins" // 피신청인 승리
	ArbitrationDecisionPartialWin    ArbitrationDecision = "partial_win"    // 부분 승리
	ArbitrationDecisionDismissed     ArbitrationDecision = "dismissed"      // 기각
	ArbitrationDecisionSettled       ArbitrationDecision = "settled"        // 합의
)

// ArbitrationVote 배심원 투표
type ArbitrationVote struct {
	ID       uint `json:"id" gorm:"primaryKey"`
	CaseID   uint `json:"case_id" gorm:"not null;index"`
	JurorID  uint `json:"juror_id" gorm:"not null;index"`
	
	// 투표 내용 (Commit-Reveal 방식)
	CommitHash    string    `json:"commit_hash"`                    // SHA256(vote + salt)
	RevealedVote  *ArbitrationDecision `json:"revealed_vote"`      // 공개된 투표
	RevealedSalt  string    `json:"revealed_salt"`                 // 공개된 솔트
	VoteReason    string    `json:"vote_reason" gorm:"type:text"`  // 투표 이유
	
	// 배심원 자격
	JurorStake    int64     `json:"juror_stake"`                   // 배심원 스테이킹 금액
	QualificationScore float64 `json:"qualification_score"`        // 자격 점수
	
	// 투표 과정
	CommittedAt   *time.Time `json:"committed_at"`                 // 투표 제출 시간
	RevealedAt    *time.Time `json:"revealed_at"`                  // 투표 공개 시간
	IsValid       bool      `json:"is_valid" gorm:"default:true"`  // 유효한 투표인지
	
	// 보상/페널티
	RewardAmount  int64     `json:"reward_amount"`                 // 배심원 보상
	PenaltyAmount int64     `json:"penalty_amount"`               // 페널티 (불참/잘못된 투표)
	
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// 관계
	Case  ArbitrationCase `json:"case,omitempty" gorm:"foreignKey:CaseID"`
	Juror User           `json:"juror,omitempty" gorm:"foreignKey:JurorID"`
}

func (ArbitrationVote) TableName() string {
	return "arbitration_votes"
}

// JurorQualification 배심원 자격
type JurorQualification struct {
	ID     uint `json:"id" gorm:"primaryKey"`
	UserID uint `json:"user_id" gorm:"not null;uniqueIndex"`
	
	// 자격 요건
	MinStakeAmount     int64   `json:"min_stake_amount" gorm:"default:5000"`     // 최소 스테이킹 5,000 BLUEPRINT
	CurrentStake       int64   `json:"current_stake"`                           // 현재 스테이킹 양
	ReputationScore    float64 `json:"reputation_score" gorm:"default:0.5"`     // 평판 점수 (0-1)
	
	// 전문성
	ExpertiseAreas     []string `json:"expertise_areas" gorm:"type:jsonb"`      // 전문 분야
	LanguageSkills     []string `json:"language_skills" gorm:"type:jsonb"`      // 언어 능력
	LegalBackground    bool     `json:"legal_background" gorm:"default:false"`  // 법률 배경 지식
	
	// 배심원 히스토리
	TotalCases         int     `json:"total_cases" gorm:"default:0"`            // 총 참여 사건 수
	AccuracyRate       float64 `json:"accuracy_rate" gorm:"default:0"`          // 정확도 (다수 의견과 일치율)
	ParticipationRate  float64 `json:"participation_rate" gorm:"default:1"`     // 참여율
	AverageResponseTime int    `json:"avg_response_time" gorm:"default:0"`      // 평균 응답 시간 (시간)
	
	// 상태
	IsActive          bool       `json:"is_active" gorm:"default:true"`          // 활성 상태
	IsSuspended       bool       `json:"is_suspended" gorm:"default:false"`      // 정지 상태
	SuspendedUntil    *time.Time `json:"suspended_until"`                        // 정지 해제일
	SuspensionReason  string     `json:"suspension_reason"`                      // 정지 사유
	
	LastActiveAt time.Time `json:"last_active_at"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	// 관계
	User User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

func (JurorQualification) TableName() string {
	return "juror_qualifications"
}

// ArbitrationReward 배심원 보상
type ArbitrationReward struct {
	ID       uint `json:"id" gorm:"primaryKey"`
	CaseID   uint `json:"case_id" gorm:"not null;index"`
	JurorID  uint `json:"juror_id" gorm:"not null;index"`
	
	// 보상 정보
	BaseReward      int64   `json:"base_reward"`                        // 기본 보상 (참여비)
	PerformanceBonus int64  `json:"performance_bonus"`                  // 성과 보너스
	QualityBonus    int64   `json:"quality_bonus"`                      // 품질 보너스 (상세한 이유 제공 등)
	TotalReward     int64   `json:"total_reward"`                       // 총 보상
	
	// 보상 조건
	VotedWithMajority bool    `json:"voted_with_majority"`               // 다수 의견과 일치 여부
	ResponseTime      int     `json:"response_time"`                     // 응답 시간 (시간)
	QualityScore      float64 `json:"quality_score" gorm:"default:0.5"`  // 투표 품질 점수
	
	// 지급 상태
	Status        string     `json:"status" gorm:"default:'pending'"`    // pending, distributed, forfeited
	DistributedAt *time.Time `json:"distributed_at"`
	
	CreatedAt time.Time `json:"created_at"`

	// 관계
	Case  ArbitrationCase `json:"case,omitempty" gorm:"foreignKey:CaseID"`
	Juror User           `json:"juror,omitempty" gorm:"foreignKey:JurorID"`
}

func (ArbitrationReward) TableName() string {
	return "arbitration_rewards"
}

// 🔧 API Request/Response Models

// SubmitArbitrationRequest 분쟁 제기 요청
type SubmitArbitrationRequest struct {
	DefendantID   uint                   `json:"defendant_id" binding:"required"`
	DisputeType   ArbitrationDisputeType `json:"dispute_type" binding:"required"`
	MilestoneID   *uint                  `json:"milestone_id,omitempty"`
	MentorshipID  *uint                  `json:"mentorship_id,omitempty"`
	TradeID       *uint                  `json:"trade_id,omitempty"`
	Title         string                 `json:"title" binding:"required"`
	Description   string                 `json:"description" binding:"required"`
	Evidence      string                 `json:"evidence"`
	ClaimedAmount int64                  `json:"claimed_amount"`
	StakeAmount   int64                  `json:"stake_amount" binding:"min=1000"`  // 최소 1,000 BLUEPRINT
}

// JurorVoteRequest 배심원 투표 요청
type JurorVoteRequest struct {
	CaseID     uint   `json:"case_id" binding:"required"`
	CommitHash string `json:"commit_hash" binding:"required"`  // SHA256(vote + salt)
}

// RevealVoteRequest 투표 공개 요청
type RevealVoteRequest struct {
	CaseID       uint                `json:"case_id" binding:"required"`
	Vote         ArbitrationDecision `json:"vote" binding:"required"`
	Salt         string              `json:"salt" binding:"required"`
	VoteReason   string              `json:"vote_reason"`
}

// ArbitrationCaseResponse 분쟁 사건 응답
type ArbitrationCaseResponse struct {
	Case       ArbitrationCase   `json:"case"`
	Votes      []ArbitrationVote `json:"votes"`
	CanVote    bool              `json:"can_vote"`        // 현재 사용자가 배심원으로 투표 가능한지
	UserVote   *ArbitrationVote  `json:"user_vote"`       // 현재 사용자의 투표 (있다면)
	TimeLeft   int64             `json:"time_left"`       // 남은 시간 (초)
	Statistics CaseStatistics   `json:"statistics"`
}

// CaseStatistics 사건 통계
type CaseStatistics struct {
	TotalJurors      int     `json:"total_jurors"`
	VotesCommitted   int     `json:"votes_committed"`
	VotesRevealed    int     `json:"votes_revealed"`
	MajorityDecision *ArbitrationDecision `json:"majority_decision"`
	DecisionConfidence float64 `json:"decision_confidence"`  // 신뢰도 (0-1)
}

// JurorDashboardResponse 배심원 대시보드 응답
type JurorDashboardResponse struct {
	Qualification   JurorQualification  `json:"qualification"`
	PendingCases    []ArbitrationCase   `json:"pending_cases"`     // 참여 가능한 사건들
	ActiveCases     []ArbitrationCase   `json:"active_cases"`      // 현재 참여 중인 사건들
	CompletedCases  []ArbitrationCase   `json:"completed_cases"`   // 완료된 사건들
	TotalRewards    int64               `json:"total_rewards"`     // 총 보상
	Statistics      JurorStatistics     `json:"statistics"`
}

// JurorStatistics 배심원 통계
type JurorStatistics struct {
	TotalCases        int     `json:"total_cases"`
	AccuracyRate      float64 `json:"accuracy_rate"`
	ParticipationRate float64 `json:"participation_rate"`
	AverageResponseTime int   `json:"avg_response_time"`
	Rank              int     `json:"rank"`              // 전체 배심원 중 순위
	TotalEarnings     int64   `json:"total_earnings"`
}