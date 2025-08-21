package models

import (
	"time"
)

// 분쟁 상태
type DisputeStatus string

const (
	DisputeStatusChallengeWindow DisputeStatus = "challenge_window"  // 이의 제기 대기 (48시간)
	DisputeStatusVotingPeriod    DisputeStatus = "voting_period"     // 판결 투표 중 (72시간)
	DisputeStatusResolved        DisputeStatus = "resolved"          // 분쟁 해결 완료
	DisputeStatusRejected        DisputeStatus = "rejected"          // 분쟁 기각 (생성자 승소)
	DisputeStatusUpheld          DisputeStatus = "upheld"            // 분쟁 인용 (제기자 승소)
)

// 분쟁 심급 (투자액 기준)
type DisputeTier string

const (
	DisputeTierExpert    DisputeTier = "expert"     // Tier 1: 전문가 판결 (<10,000 USDC)
	DisputeTierGovernance DisputeTier = "governance" // Tier 2: DAO 거버넌스 (≥10,000 USDC)
)

// 투표 선택
type VoteChoice string

const (
	VoteChoiceMaintain VoteChoice = "maintain" // 생성자 결과 유지
	VoteChoiceOverrule VoteChoice = "overrule" // 분쟁 제기자 지지
)

// 분쟁 메인 모델
type Dispute struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	MilestoneID uint           `json:"milestone_id" gorm:"not null;index"`

	// 기본 정보
	ChallengerID        uint           `json:"challenger_id" gorm:"not null;index"`           // 분쟁 제기자
	OriginalResult      bool           `json:"original_result"`                               // 원래 결과 (true=성공)
	DisputeReason       string         `json:"dispute_reason" gorm:"type:text;not null"`     // 이의 제기 사유
	StakeAmount         int64          `json:"stake_amount" gorm:"not null;default:10000"`   // 예치금 (센트 단위)

	// 분쟁 처리 정보
	Status              DisputeStatus  `json:"status" gorm:"type:varchar(20);default:'challenge_window'"`
	Tier                DisputeTier    `json:"tier" gorm:"type:varchar(20)"`                 // 심급
	TotalInvestmentAmount int64        `json:"total_investment_amount"`                       // 총 투자액 (심급 결정용)

	// 타이밍
	ChallengeWindowEnd  time.Time      `json:"challenge_window_end"`  // 이의 제기 마감 (48시간)
	VotingPeriodEnd     *time.Time     `json:"voting_period_end"`     // 투표 마감 (72시간)

	// 투표 결과
	MaintainVotes       int            `json:"maintain_votes" gorm:"default:0"`       // 생성자 지지 투표
	OverruleVotes       int            `json:"overrule_votes" gorm:"default:0"`       // 제기자 지지 투표
	FinalResult         *bool          `json:"final_result"`                          // 최종 판결 (null=미결정)

	// 관계
	Milestone           Milestone      `json:"milestone" gorm:"foreignKey:MilestoneID"`
	Challenger          User           `json:"challenger" gorm:"foreignKey:ChallengerID"`
	Votes               []DisputeVote  `json:"votes" gorm:"foreignKey:DisputeID"`
	JuryMembers         []DisputeJury  `json:"jury_members" gorm:"foreignKey:DisputeID"`
	Stakes              []DisputeStake `json:"stakes" gorm:"foreignKey:DisputeID"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// 분쟁 투표
type DisputeVote struct {
	ID        uint       `json:"id" gorm:"primaryKey"`
	DisputeID uint       `json:"dispute_id" gorm:"not null;index"`
	VoterID   uint       `json:"voter_id" gorm:"not null;index"`

	Choice    VoteChoice `json:"choice" gorm:"type:varchar(20);not null"`      // 투표 선택
	TokenAmount int64    `json:"token_amount"`                                 // 토큰 가중치 (DAO 투표용)
	InvestmentAmount int64 `json:"investment_amount"`                          // 투자액 (전문가 투표용)

	// 관계
	Dispute   Dispute    `json:"dispute" gorm:"foreignKey:DisputeID"`
	Voter     User       `json:"voter" gorm:"foreignKey:VoterID"`

	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

// 판결단 구성 (Tier 1 전문가 판결용)
type DisputeJury struct {
	ID        uint    `json:"id" gorm:"primaryKey"`
	DisputeID uint    `json:"dispute_id" gorm:"not null;index"`
	JurorID   uint    `json:"juror_id" gorm:"not null;index"`

	Position  string  `json:"position" gorm:"type:varchar(20)"`       // "success_investor" 또는 "fail_investor"
	InvestmentAmount int64 `json:"investment_amount"`               // 투자액 (선정 기준)
	HasVoted  bool    `json:"has_voted" gorm:"default:false"`      // 투표 참여 여부

	// 관계
	Dispute   Dispute `json:"dispute" gorm:"foreignKey:DisputeID"`
	Juror     User    `json:"juror" gorm:"foreignKey:JurorID"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// 예치금 관리
type DisputeStake struct {
	ID          uint     `json:"id" gorm:"primaryKey"`
	DisputeID   uint     `json:"dispute_id" gorm:"not null;index"`
	UserID      uint     `json:"user_id" gorm:"not null;index"`

	Amount      int64    `json:"amount" gorm:"not null"`                    // 예치금액 (센트)
	IsRefunded  bool     `json:"is_refunded" gorm:"default:false"`         // 반환 여부
	IsForfeited bool     `json:"is_forfeited" gorm:"default:false"`        // 몰수 여부

	// 관계
	Dispute     Dispute  `json:"dispute" gorm:"foreignKey:DisputeID"`
	User        User     `json:"user" gorm:"foreignKey:UserID"`

	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// 마일스톤 결과 보고
type MilestoneResult struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	MilestoneID uint      `json:"milestone_id" gorm:"not null;unique;index"`
	ReporterID  uint      `json:"reporter_id" gorm:"not null;index"`           // 보고자 (보통 프로젝트 생성자)

	Result      bool      `json:"result"`                                      // 보고된 결과 (true=성공)
	EvidenceURL string    `json:"evidence_url" gorm:"type:text"`               // 증거 URL
	EvidenceFiles string  `json:"evidence_files" gorm:"type:json"`             // 증거 파일들 (JSON 배열)
	Description string    `json:"description" gorm:"type:text"`                // 설명

	IsDisputed  bool      `json:"is_disputed" gorm:"default:false"`            // 분쟁 중 여부
	IsFinal     bool      `json:"is_final" gorm:"default:false"`               // 최종 확정 여부

	// 관계
	Milestone   Milestone `json:"milestone" gorm:"foreignKey:MilestoneID"`
	Reporter    User      `json:"reporter" gorm:"foreignKey:ReporterID"`
	Dispute     *Dispute  `json:"dispute,omitempty" gorm:"foreignKey:MilestoneID;references:MilestoneID"`

	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// 분쟁 관련 요청/응답 구조체들

// 이의 제기 요청
type CreateDisputeRequest struct {
	MilestoneID   uint   `json:"milestone_id" binding:"required"`
	DisputeReason string `json:"dispute_reason" binding:"required,min=100"`  // 최소 100자
}

// 투표 요청
type SubmitVoteRequest struct {
	DisputeID uint       `json:"dispute_id" binding:"required"`
	Choice    VoteChoice `json:"choice" binding:"required"`
}

// 분쟁 상세 응답
type DisputeDetailResponse struct {
	Dispute         Dispute           `json:"dispute"`
	MilestoneResult MilestoneResult   `json:"milestone_result"`
	JuryMembers     []DisputeJury     `json:"jury_members"`
	VotingStats     VotingStats       `json:"voting_stats"`
	TimeRemaining   TimeRemaining     `json:"time_remaining"`
}

// 투표 통계
type VotingStats struct {
	TotalVoters    int `json:"total_voters"`
	VotedCount     int `json:"voted_count"`
	MaintainVotes  int `json:"maintain_votes"`
	OverruleVotes  int `json:"overrule_votes"`
	VotingProgress float64 `json:"voting_progress"` // 0-1
}

// 시간 정보
type TimeRemaining struct {
	Phase       string `json:"phase"`        // "challenge_window" 또는 "voting_period"
	Hours       int    `json:"hours"`
	Minutes     int    `json:"minutes"`
	Seconds     int    `json:"seconds"`
	IsExpired   bool   `json:"is_expired"`
}
