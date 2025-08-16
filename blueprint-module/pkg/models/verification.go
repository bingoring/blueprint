package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

// 🔍 마일스톤 증명 및 검증 시스템 모델들

// ProofType 증거 타입
type ProofType string

const (
	ProofTypeFile        ProofType = "file"        // 파일 업로드 (이미지, PDF, 문서 등)
	ProofTypeURL         ProofType = "url"         // 웹 링크 (GitHub, 블로그, 포트폴리오 등)
	ProofTypeAPI         ProofType = "api"         // API 연동 데이터 (GitHub, 헬스앱 등)
	ProofTypeText        ProofType = "text"        // 텍스트 설명
	ProofTypeVideo       ProofType = "video"       // 영상 업로드/링크
	ProofTypeScreenshot  ProofType = "screenshot"  // 스크린샷
	ProofTypeCertificate ProofType = "certificate" // 인증서/성적표
)

// ProofStatus 증거 상태
type ProofStatus string

const (
	ProofStatusSubmitted ProofStatus = "submitted" // 제출됨
	ProofStatusUnderReview ProofStatus = "under_review" // 검증 중
	ProofStatusApproved  ProofStatus = "approved"  // 승인됨
	ProofStatusRejected  ProofStatus = "rejected"  // 거부됨
	ProofStatusDisputed  ProofStatus = "disputed"  // 분쟁 중
)

// MilestoneVerificationStatus 마일스톤 검증 상태
type MilestoneVerificationStatus string

const (
	MilestoneVerificationStatusPending   MilestoneVerificationStatus = "pending"   // 검증 대기
	MilestoneVerificationStatusActive    MilestoneVerificationStatus = "active"    // 검증 진행 중
	MilestoneVerificationStatusApproved  MilestoneVerificationStatus = "approved"  // 검증 완료 (승인)
	MilestoneVerificationStatusRejected  MilestoneVerificationStatus = "rejected"  // 검증 완료 (거부)
	MilestoneVerificationStatusDisputed  MilestoneVerificationStatus = "disputed"  // 분쟁 중
	MilestoneVerificationStatusExpired   MilestoneVerificationStatus = "expired"   // 검증 기간 만료
)

// ProofMetadata 증거 메타데이터 (JSON 형태로 저장)
type ProofMetadata map[string]interface{}

// Value implements driver.Valuer for database storage
func (pm ProofMetadata) Value() (driver.Value, error) {
	return json.Marshal(pm)
}

// Scan implements sql.Scanner for database retrieval
func (pm *ProofMetadata) Scan(value interface{}) error {
	if value == nil {
		*pm = make(ProofMetadata)
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(bytes, pm)
}

// MilestoneProof 마일스톤 증거 제출
type MilestoneProof struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	MilestoneID uint      `json:"milestone_id" gorm:"not null;index"`
	UserID      uint      `json:"user_id" gorm:"not null;index"` // 멘티 (증거 제출자)
	
	// 증거 정보
	ProofType   ProofType     `json:"proof_type" gorm:"not null"`
	Title       string        `json:"title" gorm:"not null"`
	Description string        `json:"description" gorm:"type:text"`
	
	// 증거 데이터
	FileURL     string        `json:"file_url,omitempty"`      // 업로드된 파일 URL
	ExternalURL string        `json:"external_url,omitempty"`  // 외부 링크 (GitHub, 블로그 등)
	APIData     ProofMetadata `json:"api_data,omitempty" gorm:"type:jsonb"` // API 연동 데이터
	Metadata    ProofMetadata `json:"metadata,omitempty" gorm:"type:jsonb"` // 추가 메타데이터
	
	// 상태 관리
	Status       ProofStatus `json:"status" gorm:"default:'submitted'"`
	SubmittedAt  time.Time   `json:"submitted_at" gorm:"default:CURRENT_TIMESTAMP"`
	ReviewDeadline time.Time `json:"review_deadline"` // 검증 마감일 (제출 후 72시간)
	
	// 통계
	TotalValidators int `json:"total_validators" gorm:"default:0"` // 총 검증인 수
	ApprovalVotes   int `json:"approval_votes" gorm:"default:0"`   // 승인 투표 수
	RejectionVotes  int `json:"rejection_votes" gorm:"default:0"`  // 거부 투표 수
	
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// 관계
	Milestone Milestone `json:"milestone,omitempty" gorm:"foreignKey:MilestoneID"`
	User      User      `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Validators []ProofValidator `json:"validators,omitempty" gorm:"foreignKey:ProofID"`
	Disputes   []ProofDispute   `json:"disputes,omitempty" gorm:"foreignKey:ProofID"`
}

func (MilestoneProof) TableName() string {
	return "milestone_proofs"
}

// ProofValidator 증거 검증인 투표
type ProofValidator struct {
	ID      uint `json:"id" gorm:"primaryKey"`
	ProofID uint `json:"proof_id" gorm:"not null;index"`
	UserID  uint `json:"user_id" gorm:"not null;index"` // 검증인
	
	// 검증인 자격
	ValidatorType   string `json:"validator_type"`    // "mentor", "stakeholder", "expert"
	StakeAmount     int64  `json:"stake_amount"`      // 스테이킹한 BLUEPRINT 양
	QualificationScore float64 `json:"qualification_score"` // 자격 점수
	
	// 투표 정보
	Vote        string    `json:"vote"`         // "approve", "reject", "abstain"
	Confidence  float64   `json:"confidence"`   // 확신도 (0.0 - 1.0)
	Reasoning   string    `json:"reasoning" gorm:"type:text"` // 투표 이유
	Evidence    string    `json:"evidence" gorm:"type:text"`  // 추가 증거/의견
	
	// 투표 가중치
	VoteWeight  float64   `json:"vote_weight"`  // 투표 가중치 (스테이킹 양, 전문성 등에 따라)
	
	VotedAt   time.Time `json:"voted_at" gorm:"default:CURRENT_TIMESTAMP"`
	CreatedAt time.Time `json:"created_at"`

	// 관계
	Proof MilestoneProof `json:"proof,omitempty" gorm:"foreignKey:ProofID"`
	User  User           `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

func (ProofValidator) TableName() string {
	return "proof_validators"
}

// ProofDispute 증거 분쟁
type ProofDispute struct {
	ID      uint `json:"id" gorm:"primaryKey"`
	ProofID uint `json:"proof_id" gorm:"not null;index"`
	UserID  uint `json:"user_id" gorm:"not null;index"` // 분쟁 제기자
	
	// 분쟁 정보
	DisputeType   string `json:"dispute_type"`   // "fraud", "insufficient_proof", "technical_error"
	Title         string `json:"title" gorm:"not null"`
	Description   string `json:"description" gorm:"type:text;not null"`
	Evidence      string `json:"evidence" gorm:"type:text"`
	
	// 분쟁 해결
	Status        string    `json:"status" gorm:"default:'open'"` // "open", "investigating", "resolved", "dismissed"
	Resolution    string    `json:"resolution" gorm:"type:text"`  // 해결 결과
	ResolvedBy    *uint     `json:"resolved_by"`                  // 해결한 관리자/중재자
	ResolvedAt    *time.Time `json:"resolved_at"`
	
	// 스테이킹 (분쟁 제기 시 일정량 스테이킹 필요)
	StakeAmount   int64     `json:"stake_amount"`   // 분쟁 제기 시 스테이킹한 BLUEPRINT
	StakeReturned bool      `json:"stake_returned"` // 스테이킹 반환 여부
	
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// 관계
	Proof     MilestoneProof `json:"proof,omitempty" gorm:"foreignKey:ProofID"`
	User      User           `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Resolver  *User          `json:"resolver,omitempty" gorm:"foreignKey:ResolvedBy"`
}

func (ProofDispute) TableName() string {
	return "proof_disputes"
}

// MilestoneVerification 마일스톤 검증 프로세스
type MilestoneVerification struct {
	ID          uint `json:"id" gorm:"primaryKey"`
	MilestoneID uint `json:"milestone_id" gorm:"not null;uniqueIndex"`
	ProofID     uint `json:"proof_id" gorm:"not null;index"`
	
	// 검증 프로세스 상태
	Status           MilestoneVerificationStatus `json:"status" gorm:"default:'pending'"`
	StartedAt        time.Time          `json:"started_at" gorm:"default:CURRENT_TIMESTAMP"`
	ReviewDeadline   time.Time          `json:"review_deadline"`   // 72시간 후
	CompletedAt      *time.Time         `json:"completed_at"`
	
	// 검증 결과
	FinalResult      string    `json:"final_result"`      // "approved", "rejected"
	ApprovalRate     float64   `json:"approval_rate"`     // 승인률 (0.0 - 1.0)
	TotalVotes       int       `json:"total_votes"`       // 총 투표 수
	WeightedScore    float64   `json:"weighted_score"`    // 가중 점수
	MinimumVotes     int       `json:"minimum_votes"`     // 최소 필요 투표 수
	
	// 자동 완료 설정
	AutoCompleteAfter time.Time `json:"auto_complete_after"` // 자동 완료 시간
	AutoCompleted     bool      `json:"auto_completed"`      // 자동 완료 여부
	
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// 관계
	Milestone Milestone      `json:"milestone,omitempty" gorm:"foreignKey:MilestoneID"`
	Proof     MilestoneProof `json:"proof,omitempty" gorm:"foreignKey:ProofID"`
}

func (MilestoneVerification) TableName() string {
	return "milestone_verifications"
}

// ValidatorQualification 검증인 자격 관리
type ValidatorQualification struct {
	ID     uint `json:"id" gorm:"primaryKey"`
	UserID uint `json:"user_id" gorm:"not null;uniqueIndex"`
	
	// 자격 정보
	IsMentor           bool    `json:"is_mentor"`            // 멘토 여부
	IsExpert           bool    `json:"is_expert"`            // 전문가 여부
	StakedAmount       int64   `json:"staked_amount"`        // 현재 스테이킹 양
	ReputationScore    float64 `json:"reputation_score"`     // 평판 점수
	
	// 검증 히스토리
	TotalVerifications int     `json:"total_verifications"`  // 총 검증 참여 수
	AccuracyRate       float64 `json:"accuracy_rate"`        // 정확도 (0.0 - 1.0)
	ConsensusRate      float64 `json:"consensus_rate"`       // 다수 의견과의 일치율
	
	// 전문 분야
	ExpertiseAreas     []string `json:"expertise_areas" gorm:"type:jsonb"` // 전문 분야 목록
	
	// 제재 정보
	IsSuspended        bool       `json:"is_suspended"`        // 제재 여부
	SuspendedUntil     *time.Time `json:"suspended_until"`     // 제재 해제일
	SuspensionReason   string     `json:"suspension_reason"`   // 제재 사유
	
	LastActiveAt time.Time `json:"last_active_at"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	// 관계
	User User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

func (ValidatorQualification) TableName() string {
	return "validator_qualifications"
}

// VerificationReward 검증 참여 보상
type VerificationReward struct {
	ID           uint `json:"id" gorm:"primaryKey"`
	ValidatorID  uint `json:"validator_id" gorm:"not null;index"`  // ProofValidator ID
	UserID       uint `json:"user_id" gorm:"not null;index"`
	ProofID      uint `json:"proof_id" gorm:"not null;index"`
	
	// 보상 정보
	RewardType     string  `json:"reward_type"`    // "validation_fee", "accuracy_bonus", "consensus_bonus"
	Amount         int64   `json:"amount"`         // BLUEPRINT 토큰 보상량
	USDCAmount     int64   `json:"usdc_amount"`    // USDC 보상량 (수수료 분배)
	BonusMultiplier float64 `json:"bonus_multiplier"` // 보너스 배율
	
	// 보상 조건
	IsCorrectVote  bool    `json:"is_correct_vote"`  // 올바른 투표 여부
	VoteWeight     float64 `json:"vote_weight"`      // 투표 가중치
	
	// 지급 상태
	Status       string     `json:"status" gorm:"default:'pending'"` // "pending", "distributed", "forfeited"
	DistributedAt *time.Time `json:"distributed_at"`
	
	CreatedAt time.Time `json:"created_at"`

	// 관계
	Validator ProofValidator `json:"validator,omitempty" gorm:"foreignKey:ValidatorID"`
	User      User           `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Proof     MilestoneProof `json:"proof,omitempty" gorm:"foreignKey:ProofID"`
}

func (VerificationReward) TableName() string {
	return "verification_rewards"
}

// 🔧 API Request/Response Models

// SubmitProofRequest 증거 제출 요청
type SubmitProofRequest struct {
	MilestoneID uint      `json:"milestone_id" binding:"required"`
	ProofType   ProofType `json:"proof_type" binding:"required"`
	Title       string    `json:"title" binding:"required"`
	Description string    `json:"description"`
	ExternalURL string    `json:"external_url,omitempty"`
	APIData     ProofMetadata `json:"api_data,omitempty"`
	Metadata    ProofMetadata `json:"metadata,omitempty"`
}

// ValidateProofRequest 증거 검증 요청
type ValidateProofRequest struct {
	ProofID    uint    `json:"proof_id" binding:"required"`
	Vote       string  `json:"vote" binding:"required,oneof=approve reject abstain"`
	Confidence float64 `json:"confidence" binding:"min=0,max=1"`
	Reasoning  string  `json:"reasoning"`
	Evidence   string  `json:"evidence,omitempty"`
}

// DisputeProofRequest 증거 분쟁 제기 요청
type DisputeProofRequest struct {
	ProofID     uint   `json:"proof_id" binding:"required"`
	DisputeType string `json:"dispute_type" binding:"required,oneof=fraud insufficient_proof technical_error"`
	Title       string `json:"title" binding:"required"`
	Description string `json:"description" binding:"required"`
	Evidence    string `json:"evidence,omitempty"`
	StakeAmount int64  `json:"stake_amount" binding:"min=1000"` // 최소 1000 BLUEPRINT 스테이킹
}

// ProofVerificationResponse 증거 검증 응답
type ProofVerificationResponse struct {
	Proof        MilestoneProof        `json:"proof"`
	Verification MilestoneVerification `json:"verification"`
	Validators   []ProofValidator      `json:"validators"`
	Disputes     []ProofDispute        `json:"disputes"`
	CanVote      bool                  `json:"can_vote"`      // 현재 사용자가 투표 가능한지
	UserVote     *ProofValidator       `json:"user_vote"`     // 현재 사용자의 투표 (있다면)
}

// ValidatorDashboardResponse 검증인 대시보드 응답
type ValidatorDashboardResponse struct {
	Qualification ValidatorQualification `json:"qualification"`
	PendingProofs []MilestoneProof       `json:"pending_proofs"`
	RecentVotes   []ProofValidator       `json:"recent_votes"`
	Rewards       []VerificationReward   `json:"rewards"`
	Statistics    ValidatorStatistics    `json:"statistics"`
}

// ValidatorStatistics 검증인 통계
type ValidatorStatistics struct {
	TotalVotes       int     `json:"total_votes"`
	AccuracyRate     float64 `json:"accuracy_rate"`
	ConsensusRate    float64 `json:"consensus_rate"`
	TotalRewards     int64   `json:"total_rewards"`
	CurrentStake     int64   `json:"current_stake"`
	ReputationScore  float64 `json:"reputation_score"`
	Rank             int     `json:"rank"`
}