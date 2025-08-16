package models

import (
	"time"
)

// 💎 멘토 스테이킹 및 슬래싱 시스템

// MentorStake 멘토 스테이킹
type MentorStake struct {
	ID       uint `json:"id" gorm:"primaryKey"`
	MentorID uint `json:"mentor_id" gorm:"not null;index"`
	UserID   uint `json:"user_id" gorm:"not null;index"`  // 스테이킹한 사용자 (보통 멘토 본인)
	
	// 스테이킹 정보
	Amount          int64                `json:"amount" gorm:"not null"`                    // 스테이킹 금액 (BLUEPRINT)
	LockedAmount    int64                `json:"locked_amount" gorm:"default:0"`            // 잠긴 금액 (슬래싱 대상)
	AvailableAmount int64                `json:"available_amount"`                          // 사용 가능 금액
	StakeType       MentorStakeType      `json:"stake_type" gorm:"default:'self'"`          // 스테이킹 유형
	Purpose         MentorStakePurpose   `json:"purpose" gorm:"default:'qualification'"`   // 스테이킹 목적
	
	// 스테이킹 조건
	MinimumPeriod   int       `json:"minimum_period" gorm:"default:30"`              // 최소 잠금 기간 (일)
	UnlockDate      time.Time `json:"unlock_date"`                                   // 잠금 해제일
	
	// 성과 관련
	ExpectedROI     float64   `json:"expected_roi" gorm:"default:0"`                 // 예상 수익률
	ActualROI       float64   `json:"actual_roi" gorm:"default:0"`                  // 실제 수익률
	PerformanceBonus int64    `json:"performance_bonus" gorm:"default:0"`           // 성과 보너스
	
	// 상태
	Status          MentorStakeStatus `json:"status" gorm:"default:'active'"`
	IsAutoRenewal   bool             `json:"is_auto_renewal" gorm:"default:false"`    // 자동 갱신 여부
	
	// 타임스탬프
	StakedAt        time.Time  `json:"staked_at" gorm:"default:CURRENT_TIMESTAMP"`
	UnstakedAt      *time.Time `json:"unstaked_at"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`

	// 관계
	Mentor      Mentor                `json:"mentor,omitempty" gorm:"foreignKey:MentorID"`
	User        User                  `json:"user,omitempty" gorm:"foreignKey:UserID"`
	SlashEvents []MentorSlashEvent    `json:"slash_events,omitempty" gorm:"foreignKey:StakeID"`
	Rewards     []MentorStakeReward   `json:"rewards,omitempty" gorm:"foreignKey:StakeID"`
}

func (MentorStake) TableName() string {
	return "mentor_stakes"
}

// MentorStakeType 스테이킹 유형
type MentorStakeType string

const (
	MentorStakeTypeSelf       MentorStakeType = "self"        // 자기 스테이킹
	MentorStakeTypeDelegated  MentorStakeType = "delegated"   // 위임 스테이킹
	MentorStakeTypePool       MentorStakeType = "pool"        // 풀 스테이킹
	MentorStakeTypeInsurance  MentorStakeType = "insurance"   // 보험 스테이킹
)

// MentorStakePurpose 스테이킹 목적
type MentorStakePurpose string

const (
	MentorStakePurposeQualification MentorStakePurpose = "qualification" // 자격 증명
	MentorStakePurposePerformance   MentorStakePurpose = "performance"   // 성과 보장
	MentorStakePurposeInsurance     MentorStakePurpose = "insurance"     // 보험/보상
	MentorStakePurposeGovernance    MentorStakePurpose = "governance"    // 거버넌스 참여
)

// MentorStakeStatus 스테이킹 상태
type MentorStakeStatus string

const (
	MentorStakeStatusActive    MentorStakeStatus = "active"     // 활성
	MentorStakeStatusUnlocking MentorStakeStatus = "unlocking"  // 잠금 해제 중
	MentorStakeStatusSlashed   MentorStakeStatus = "slashed"    // 슬래싱됨
	MentorStakeStatusWithdrawn MentorStakeStatus = "withdrawn"  // 인출됨
	MentorStakeStatusFrozen    MentorStakeStatus = "frozen"     // 동결됨
)

// MentorSlashEvent 멘토 슬래싱 이벤트
type MentorSlashEvent struct {
	ID          uint `json:"id" gorm:"primaryKey"`
	StakeID     uint `json:"stake_id" gorm:"not null;index"`
	MentorID    uint `json:"mentor_id" gorm:"not null;index"`
	ReporterID  *uint `json:"reporter_id,omitempty" gorm:"index"`  // 신고자 (없으면 시스템 자동)
	
	// 슬래싱 정보
	SlashType      MentorSlashType   `json:"slash_type" gorm:"not null"`
	Severity       SlashSeverity     `json:"severity" gorm:"not null"`
	SlashedAmount  int64             `json:"slashed_amount" gorm:"not null"`      // 슬래싱된 금액
	SlashRate      float64           `json:"slash_rate"`                          // 슬래싱 비율 (0-1)
	
	// 사유 및 증거
	Reason         string            `json:"reason" gorm:"not null"`
	Description    string            `json:"description" gorm:"type:text"`
	Evidence       string            `json:"evidence" gorm:"type:text"`           // JSON 형태 증거
	
	// 관련 정보
	MilestoneID    *uint             `json:"milestone_id,omitempty" gorm:"index"`
	MentorshipID   *uint             `json:"mentorship_id,omitempty" gorm:"index"`
	ProofID        *uint             `json:"proof_id,omitempty" gorm:"index"`     // 검증 관련
	
	// 처리 과정
	Status         SlashEventStatus  `json:"status" gorm:"default:'pending'"`
	ReviewedBy     *uint             `json:"reviewed_by,omitempty" gorm:"index"`  // 검토자
	ReviewComment  string            `json:"review_comment" gorm:"type:text"`
	
	// 이의제기 및 복구
	CanAppeal      bool              `json:"can_appeal" gorm:"default:true"`
	AppealDeadline *time.Time        `json:"appeal_deadline"`
	IsAppealed     bool              `json:"is_appealed" gorm:"default:false"`
	AppealCase     *uint             `json:"appeal_case,omitempty"`               // 분쟁 해결 사건 ID
	
	// 타임스탬프
	DetectedAt     time.Time         `json:"detected_at" gorm:"default:CURRENT_TIMESTAMP"`
	ProcessedAt    *time.Time        `json:"processed_at"`
	CreatedAt      time.Time         `json:"created_at"`
	UpdatedAt      time.Time         `json:"updated_at"`

	// 관계
	Stake       MentorStake       `json:"stake,omitempty" gorm:"foreignKey:StakeID"`
	Mentor      Mentor            `json:"mentor,omitempty" gorm:"foreignKey:MentorID"`
	Reporter    *User             `json:"reporter,omitempty" gorm:"foreignKey:ReporterID"`
	Reviewer    *User             `json:"reviewer,omitempty" gorm:"foreignKey:ReviewedBy"`
	Milestone   *Milestone        `json:"milestone,omitempty" gorm:"foreignKey:MilestoneID"`
}

func (MentorSlashEvent) TableName() string {
	return "mentor_slash_events"
}

// MentorSlashType 슬래싱 유형
type MentorSlashType string

const (
	SlashTypeAbandonment      MentorSlashType = "abandonment"       // 멘토링 포기/방치
	SlashTypeMalpractice      MentorSlashType = "malpractice"       // 직무유기/부정행위
	SlashTypeFraud            MentorSlashType = "fraud"             // 사기
	SlashTypePoorPerformance  MentorSlashType = "poor_performance"  // 저조한 성과
	SlashTypeEthicsViolation  MentorSlashType = "ethics_violation"  // 윤리 위반
	SlashTypeAbuse            MentorSlashType = "abuse"             // 괴롭힘/남용
	SlashTypeConflictOfInterest MentorSlashType = "conflict_interest" // 이해충돌
	SlashTypeNoShow           MentorSlashType = "no_show"           // 무단 불참
)

// SlashSeverity 슬래싱 심각도
type SlashSeverity string

const (
	SlashSeverityMinor    SlashSeverity = "minor"     // 경미 (5-10% 슬래싱)
	SlashSeverityModerate SlashSeverity = "moderate"  // 보통 (10-25% 슬래싱)
	SlashSeverityMajor    SlashSeverity = "major"     // 심각 (25-50% 슬래싱)
	SlashSeverityCritical SlashSeverity = "critical"  // 극심 (50-100% 슬래싱)
)

// SlashEventStatus 슬래싱 이벤트 상태
type SlashEventStatus string

const (
	SlashEventStatusPending   SlashEventStatus = "pending"    // 검토 대기
	SlashEventStatusReviewing SlashEventStatus = "reviewing"  // 검토 중
	SlashEventStatusApproved  SlashEventStatus = "approved"   // 승인됨 (슬래싱 실행)
	SlashEventStatusRejected  SlashEventStatus = "rejected"   // 거부됨
	SlashEventStatusAppealed  SlashEventStatus = "appealed"   // 이의제기됨
	SlashEventStatusReversed  SlashEventStatus = "reversed"   // 취소됨 (복구)
)

// MentorPerformanceMetric 멘토 성과 지표
type MentorPerformanceMetric struct {
	ID       uint `json:"id" gorm:"primaryKey"`
	MentorID uint `json:"mentor_id" gorm:"not null;index"`
	
	// 시간 범위
	PeriodType  MetricPeriodType `json:"period_type" gorm:"not null"`  // weekly, monthly, quarterly
	StartDate   time.Time        `json:"start_date" gorm:"not null"`
	EndDate     time.Time        `json:"end_date" gorm:"not null"`
	
	// 멘토링 활동 지표
	TotalMentees           int     `json:"total_mentees"`              // 총 멘티 수
	ActiveMentees          int     `json:"active_mentees"`             // 활성 멘티 수
	CompletedMentorships   int     `json:"completed_mentorships"`      // 완료된 멘토링
	SuccessfulMilestones   int     `json:"successful_milestones"`      // 성공한 마일스톤
	TotalMilestones        int     `json:"total_milestones"`           // 총 마일스톤
	SuccessRate            float64 `json:"success_rate"`               // 성공률
	
	// 참여도 지표
	TotalSessions          int     `json:"total_sessions"`             // 총 세션 수
	AttendanceRate         float64 `json:"attendance_rate"`            // 출석률
	ResponseTime           int     `json:"avg_response_time"`          // 평균 응답 시간 (시간)
	SessionRating          float64 `json:"avg_session_rating"`         // 평균 세션 평점
	
	// 만족도 지표
	MenteeRating           float64 `json:"avg_mentee_rating"`          // 멘티 평가 평균
	FeedbackScore          float64 `json:"avg_feedback_score"`         // 피드백 점수
	RetentionRate          float64 `json:"mentee_retention_rate"`      // 멘티 유지율
	ReferralRate           float64 `json:"referral_rate"`              // 추천율
	
	// 경제적 지표
	TotalRevenue           int64   `json:"total_revenue"`              // 총 수익
	AvgRevenuePerMentee    int64   `json:"avg_revenue_per_mentee"`     // 멘티당 평균 수익
	ProfitMargin           float64 `json:"profit_margin"`              // 수익률
	
	// 위험 지표
	ComplaintCount         int     `json:"complaint_count"`            // 불만 접수 수
	DisputeCount           int     `json:"dispute_count"`              // 분쟁 건수
	SlashCount             int     `json:"slash_count"`                // 슬래싱 횟수
	SlashedAmount          int64   `json:"slashed_amount"`             // 슬래싱된 총 금액
	
	// 종합 점수
	PerformanceScore       float64 `json:"performance_score"`          // 종합 성과 점수 (0-100)
	RiskScore              float64 `json:"risk_score"`                 // 위험 점수 (0-100)
	QualityScore           float64 `json:"quality_score"`              // 품질 점수 (0-100)
	
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// 관계
	Mentor Mentor `json:"mentor,omitempty" gorm:"foreignKey:MentorID"`
}

func (MentorPerformanceMetric) TableName() string {
	return "mentor_performance_metrics"
}

// MetricPeriodType 지표 기간 유형
type MetricPeriodType string

const (
	MetricPeriodWeekly    MetricPeriodType = "weekly"
	MetricPeriodMonthly   MetricPeriodType = "monthly"
	MetricPeriodQuarterly MetricPeriodType = "quarterly"
	MetricPeriodYearly    MetricPeriodType = "yearly"
)

// MentorStakeReward 스테이킹 보상
type MentorStakeReward struct {
	ID      uint `json:"id" gorm:"primaryKey"`
	StakeID uint `json:"stake_id" gorm:"not null;index"`
	MentorID uint `json:"mentor_id" gorm:"not null;index"`
	
	// 보상 정보
	RewardType    MentorRewardType `json:"reward_type" gorm:"not null"`
	Amount        int64            `json:"amount" gorm:"not null"`               // 보상 금액
	BonusMultiplier float64        `json:"bonus_multiplier" gorm:"default:1"`    // 보너스 배율
	
	// 보상 조건
	PeriodStart   time.Time        `json:"period_start"`                         // 보상 기간 시작
	PeriodEnd     time.Time        `json:"period_end"`                           // 보상 기간 종료
	TriggerEvent  string           `json:"trigger_event"`                        // 트리거 이벤트
	
	// 성과 기반 조건
	MilestonesCompleted int           `json:"milestones_completed"`               // 완료된 마일스톤 수
	SuccessRate         float64       `json:"success_rate"`                       // 성공률
	SatisfactionScore   float64       `json:"satisfaction_score"`                 // 만족도 점수
	
	// 지급 상태
	Status        string           `json:"status" gorm:"default:'pending'"`      // pending, distributed, forfeited
	DistributedAt *time.Time       `json:"distributed_at"`
	
	CreatedAt time.Time `json:"created_at"`

	// 관계
	Stake  MentorStake `json:"stake,omitempty" gorm:"foreignKey:StakeID"`
	Mentor Mentor      `json:"mentor,omitempty" gorm:"foreignKey:MentorID"`
}

func (MentorStakeReward) TableName() string {
	return "mentor_stake_rewards"
}

// MentorRewardType 멘토 보상 유형
type MentorRewardType string

const (
	MentorRewardTypeStaking      MentorRewardType = "staking"        // 스테이킹 보상
	MentorRewardTypePerformance  MentorRewardType = "performance"    // 성과 보상
	MentorRewardTypeCompletion   MentorRewardType = "completion"     // 완료 보상
	MentorRewardTypeLoyalty      MentorRewardType = "loyalty"        // 충성도 보상
	MentorRewardTypeBonus        MentorRewardType = "bonus"          // 특별 보너스
)

// 🔧 API Request/Response Models

// StakeMentorRequest 멘토 스테이킹 요청
type StakeMentorRequest struct {
	MentorID      uint               `json:"mentor_id" binding:"required"`
	Amount        int64              `json:"amount" binding:"min=1000"`                // 최소 1,000 BLUEPRINT
	StakeType     MentorStakeType    `json:"stake_type"`
	Purpose       MentorStakePurpose `json:"purpose"`
	MinimumPeriod int                `json:"minimum_period" binding:"min=7,max=365"`   // 7일-1년
	IsAutoRenewal bool               `json:"is_auto_renewal"`
}

// ReportMentorRequest 멘토 신고 요청
type ReportMentorRequest struct {
	MentorID     uint            `json:"mentor_id" binding:"required"`
	SlashType    MentorSlashType `json:"slash_type" binding:"required"`
	Severity     SlashSeverity   `json:"severity" binding:"required"`
	Reason       string          `json:"reason" binding:"required"`
	Description  string          `json:"description" binding:"required"`
	Evidence     string          `json:"evidence"`
	MilestoneID  *uint           `json:"milestone_id,omitempty"`
	MentorshipID *uint           `json:"mentorship_id,omitempty"`
}

// MentorStakeResponse 멘토 스테이킹 응답
type MentorStakeResponse struct {
	Stake          MentorStake             `json:"stake"`
	Performance    MentorPerformanceMetric `json:"performance"`
	RecentSlashes  []MentorSlashEvent      `json:"recent_slashes"`
	PendingRewards []MentorStakeReward     `json:"pending_rewards"`
	Statistics     MentorStakeStatistics   `json:"statistics"`
}

// MentorStakeStatistics 멘토 스테이킹 통계
type MentorStakeStatistics struct {
	TotalStaked        int64   `json:"total_staked"`
	TotalSlashed       int64   `json:"total_slashed"`
	TotalRewards       int64   `json:"total_rewards"`
	CurrentAPY         float64 `json:"current_apy"`          // 연간 수익률
	RiskScore          float64 `json:"risk_score"`           // 위험 점수
	SlashingHistory    int     `json:"slashing_history"`     // 슬래싱 이력 수
	StakingRank        int     `json:"staking_rank"`         // 스테이킹 순위
}

// MentorDashboardResponse 멘토 대시보드 응답  
type MentorDashboardResponse struct {
	Stakes         []MentorStake           `json:"stakes"`
	Performance    MentorPerformanceMetric `json:"performance"`
	SlashEvents    []MentorSlashEvent      `json:"slash_events"`
	Rewards        []MentorStakeReward     `json:"rewards"`
	Statistics     MentorStakeStatistics   `json:"statistics"`
	Recommendations []string               `json:"recommendations"`   // 개선 제안
}