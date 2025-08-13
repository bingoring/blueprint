package services

import (
	"blueprint/internal/models"
	"fmt"
	"log"
	"time"

	"gorm.io/gorm"
)

// 💰 멘토 보상 서비스 - 성과 기반 보상 분배 시스템
type MentorRewardService struct {
	db         *gorm.DB
	sseService *SSEService
}

// NewMentorRewardService 보상 서비스 생성자
func NewMentorRewardService(db *gorm.DB, sseService *SSEService) *MentorRewardService {
	return &MentorRewardService{
		db:         db,
		sseService: sseService,
	}
}

// MentorRewardInfo 멘토 보상 정보
type MentorRewardInfo struct {
	MentorID            uint    `json:"mentor_id"`
	UserID              uint    `json:"user_id"`
	Username            string  `json:"username"`
	TotalBetAmount      int64   `json:"total_bet_amount"`      // 베팅 금액
	BetSharePercentage  float64 `json:"bet_share_percentage"`  // 베팅 비중 (%)
	MentorRating        float64 `json:"mentor_rating"`         // 멘토 평점
	ActionsCount        int     `json:"actions_count"`         // 수행한 액션 수
	IsActive            bool    `json:"is_active"`             // 활성 멘토링 여부
	IsLeadMentor        bool    `json:"is_lead_mentor"`        // 리드 멘토 여부

	// 보상 계산
	BetWeightScore      float64 `json:"bet_weight_score"`      // 베팅 가중치 점수
	PerformanceScore    float64 `json:"performance_score"`     // 성과 점수
	TotalScore          float64 `json:"total_score"`           // 총 점수
	RewardAmount        int64   `json:"reward_amount"`         // 보상 금액 (센트)
	RewardPercentage    float64 `json:"reward_percentage"`     // 보상 비중 (%)
}

// RewardDistributionResult 보상 분배 결과
type RewardDistributionResult struct {
	MilestoneID          uint                `json:"milestone_id"`
	ProjectID            uint                `json:"project_id"`
	TotalPoolAmount      int64               `json:"total_pool_amount"`
	DistributedAmount    int64               `json:"distributed_amount"`
	EligibleMentorCount  int                 `json:"eligible_mentor_count"`
	BettingAmountWeight  float64             `json:"betting_amount_weight"`  // 베팅액 가중치
	MentorRatingWeight   float64             `json:"mentor_rating_weight"`   // 평점 가중치
	MentorRewards        []MentorRewardInfo  `json:"mentor_rewards"`
	DistributedAt        time.Time           `json:"distributed_at"`
}

// DistributeMentorPoolRewards 멘토 풀 보상 분배 (마일스톤 성공 시 호출)
func (mrs *MentorRewardService) DistributeMentorPoolRewards(milestoneID uint) (*RewardDistributionResult, error) {
	log.Printf("💰 Starting mentor pool reward distribution for milestone %d", milestoneID)

	// 트랜잭션 시작
	tx := mrs.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 1. 멘토 풀 조회
	var mentorPool models.MentorPool
	if err := tx.Where("milestone_id = ?", milestoneID).First(&mentorPool).Error; err != nil {
		tx.Rollback()
		if err == gorm.ErrRecordNotFound {
			log.Printf("📋 No mentor pool found for milestone %d", milestoneID)
			return nil, nil // 멘토 풀이 없으면 그냥 넘어감
		}
		return nil, fmt.Errorf("failed to query mentor pool: %v", err)
	}

	// 이미 분배된 경우 확인
	if mentorPool.IsDistributed {
		return nil, fmt.Errorf("rewards already distributed for milestone %d", milestoneID)
	}

	if mentorPool.TotalPoolAmount <= 0 {
		log.Printf("📋 No funds in mentor pool for milestone %d", milestoneID)
		return nil, nil
	}

	// 2. 자격 있는 멘토들 조회 및 보상 정보 계산
	mentorRewards, err := mrs.calculateMentorRewards(tx, milestoneID, &mentorPool)
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to calculate mentor rewards: %v", err)
	}

	if len(mentorRewards) == 0 {
		log.Printf("📋 No eligible mentors found for milestone %d", milestoneID)
		// 풀 상태만 업데이트하고 종료
		mentorPool.IsDistributed = true
		mentorPool.EligibleMentorsCount = 0
		now := time.Now()
		mentorPool.DistributedAt = &now
		tx.Save(&mentorPool)
		tx.Commit()
		return nil, nil
	}

	// 3. 보상 분배 실행
	totalDistributed := int64(0)
	for i, reward := range mentorRewards {
		if err := mrs.distributeSingleReward(tx, milestoneID, &reward); err != nil {
			log.Printf("❌ Failed to distribute reward to mentor %d: %v", reward.MentorID, err)
			continue
		}
		totalDistributed += reward.RewardAmount
		mentorRewards[i] = reward // 업데이트된 정보 반영
	}

	// 4. 멘토 풀 상태 업데이트
	now := time.Now()
	mentorPool.IsDistributed = true
	mentorPool.DistributedAmount = totalDistributed
	mentorPool.DistributedAt = &now
	mentorPool.EligibleMentorsCount = len(mentorRewards)

	if err := tx.Save(&mentorPool).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to update mentor pool: %v", err)
	}

	// 5. 트랜잭션 커밋
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %v", err)
	}

	result := &RewardDistributionResult{
		MilestoneID:         milestoneID,
		ProjectID:           mentorPool.ProjectID,
		TotalPoolAmount:     mentorPool.TotalPoolAmount,
		DistributedAmount:   totalDistributed,
		EligibleMentorCount: len(mentorRewards),
		BettingAmountWeight: mentorPool.BettingAmountWeight,
		MentorRatingWeight:  mentorPool.MentorRatingWeight,
		MentorRewards:       mentorRewards,
		DistributedAt:       now,
	}

	log.Printf("✅ Mentor pool rewards distributed: $%.2f to %d mentors for milestone %d",
		float64(totalDistributed)/100, len(mentorRewards), milestoneID)

	// 6. 실시간 알림
	go mrs.broadcastRewardDistribution(result)

	return result, nil
}

// calculateMentorRewards 멘토 보상 정보 계산
func (mrs *MentorRewardService) calculateMentorRewards(tx *gorm.DB, milestoneID uint, pool *models.MentorPool) ([]MentorRewardInfo, error) {
	// 자격 있는 멘토들 조회 (활성 멘토링을 한 멘토들만)
	var mentorMilestones []models.MentorMilestone
	if err := tx.Where("milestone_id = ? AND is_active = ? AND actions_count > 0",
		milestoneID, true).
		Preload("Mentor").Preload("Mentor.User").
		Find(&mentorMilestones).Error; err != nil {
		return nil, err
	}

	if len(mentorMilestones) == 0 {
		return []MentorRewardInfo{}, nil
	}

	// 총 베팅 금액 및 최대 평점 계산
	totalBetAmount := int64(0)
	maxRating := 0.0
	for _, mm := range mentorMilestones {
		totalBetAmount += mm.TotalBetAmount
		if mm.MenteeRating > maxRating {
			maxRating = mm.MenteeRating
		}
	}

	// 최대 평점이 0이면 기본값 설정
	if maxRating == 0 {
		maxRating = 5.0 // 기본 최대 평점
	}

	// 각 멘토의 보상 정보 계산
	rewards := make([]MentorRewardInfo, 0, len(mentorMilestones))
	totalScore := 0.0

	for _, mm := range mentorMilestones {
		// 베팅 가중치 점수 (베팅 비중 기반)
		betWeightScore := 0.0
		if totalBetAmount > 0 {
			betWeightScore = float64(mm.TotalBetAmount) / float64(totalBetAmount)
		}

		// 성과 점수 (평점 기반, 0점도 참여 점수 부여)
		performanceScore := 0.1 // 최소 참여 점수
		if mm.MenteeRating > 0 {
			performanceScore = mm.MenteeRating / maxRating
		}

		// 총 점수 계산 (가중 평균)
		score := (betWeightScore * pool.BettingAmountWeight / 100) +
		         (performanceScore * pool.MentorRatingWeight / 100)

		reward := MentorRewardInfo{
			MentorID:           mm.MentorID,
			UserID:             mm.Mentor.UserID,
			Username:           mm.Mentor.User.Username,
			TotalBetAmount:     mm.TotalBetAmount,
			BetSharePercentage: mm.BetSharePercentage,
			MentorRating:       mm.MenteeRating,
			ActionsCount:       mm.ActionsCount,
			IsActive:           mm.IsActive,
			IsLeadMentor:       mm.IsLeadMentor,
			BetWeightScore:     betWeightScore,
			PerformanceScore:   performanceScore,
			TotalScore:         score,
		}

		rewards = append(rewards, reward)
		totalScore += score
	}

	// 보상 금액 계산
	if totalScore > 0 {
		for i := range rewards {
			rewards[i].RewardPercentage = (rewards[i].TotalScore / totalScore) * 100
			rewards[i].RewardAmount = int64(float64(pool.TotalPoolAmount) * rewards[i].TotalScore / totalScore)
		}
	}

	return rewards, nil
}

// distributeSingleReward 개별 멘토에게 보상 지급
func (mrs *MentorRewardService) distributeSingleReward(tx *gorm.DB, milestoneID uint, reward *MentorRewardInfo) error {
	// 1. MentorMilestone 보상 정보 업데이트
	if err := tx.Model(&models.MentorMilestone{}).
		Where("mentor_id = ? AND milestone_id = ?", reward.MentorID, milestoneID).
		Update("earned_from_mentoring", reward.RewardAmount).Error; err != nil {
		return fmt.Errorf("failed to update mentor milestone reward: %v", err)
	}

	// 2. 멘토 총 수익 업데이트
	if err := tx.Model(&models.Mentor{}).Where("id = ?", reward.MentorID).
		Update("total_earned_amount", gorm.Expr("total_earned_amount + ?", reward.RewardAmount)).Error; err != nil {
		return fmt.Errorf("failed to update mentor total earnings: %v", err)
	}

	// 3. 멘토 평판 점수 업데이트 (보상 금액에 비례)
	reputationPoints := int(reward.RewardAmount / 100) // $1당 1점
	if reputationPoints > 0 {
		if err := tx.Model(&models.Mentor{}).Where("id = ?", reward.MentorID).
			Update("reputation_score", gorm.Expr("reputation_score + ?", reputationPoints)).Error; err != nil {
			log.Printf("⚠️ Failed to update mentor reputation: %v", err)
			// 평판 업데이트 실패는 치명적이지 않으므로 계속 진행
		}
	}

	// 4. 평판 기록 생성 (온체인 준비)
	reputation := models.MentorReputation{
		MentorID:     reward.MentorID,
		EventType:    "mentoring_reward_earned",
		Points:       reputationPoints,
		Multiplier:   1.0,
		MilestoneID:  &milestoneID,
		Description:  fmt.Sprintf("Earned $%.2f from mentoring milestone success", float64(reward.RewardAmount)/100),
	}

	if err := tx.Create(&reputation).Error; err != nil {
		log.Printf("⚠️ Failed to create reputation record: %v", err)
		// 평판 기록 실패는 치명적이지 않으므로 계속 진행
	}

	log.Printf("💰 Distributed $%.2f to mentor %d (user: %s)",
		float64(reward.RewardAmount)/100, reward.MentorID, reward.Username)

	return nil
}

// GetRewardDistributionHistory 보상 분배 이력 조회
func (mrs *MentorRewardService) GetRewardDistributionHistory(mentorID uint) ([]models.MentorMilestone, error) {
	var mentorMilestones []models.MentorMilestone
	if err := mrs.db.Where("mentor_id = ? AND earned_from_mentoring > 0", mentorID).
		Preload("Milestone").Preload("Project").
		Order("updated_at DESC").
		Find(&mentorMilestones).Error; err != nil {
		return nil, err
	}

	return mentorMilestones, nil
}

// GetMilestoneRewardDistribution 특정 마일스톤의 보상 분배 결과 조회
func (mrs *MentorRewardService) GetMilestoneRewardDistribution(milestoneID uint) (*RewardDistributionResult, error) {
	// 멘토 풀 조회
	var mentorPool models.MentorPool
	if err := mrs.db.Where("milestone_id = ?", milestoneID).First(&mentorPool).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	if !mentorPool.IsDistributed {
		return nil, fmt.Errorf("rewards not yet distributed for milestone %d", milestoneID)
	}

	// 보상 받은 멘토들 조회
	var mentorMilestones []models.MentorMilestone
	if err := mrs.db.Where("milestone_id = ? AND earned_from_mentoring > 0", milestoneID).
		Preload("Mentor").Preload("Mentor.User").
		Order("earned_from_mentoring DESC").
		Find(&mentorMilestones).Error; err != nil {
		return nil, err
	}

	// 결과 구성
	mentorRewards := make([]MentorRewardInfo, 0, len(mentorMilestones))
	for _, mm := range mentorMilestones {
		reward := MentorRewardInfo{
			MentorID:           mm.MentorID,
			UserID:             mm.Mentor.UserID,
			Username:           mm.Mentor.User.Username,
			TotalBetAmount:     mm.TotalBetAmount,
			BetSharePercentage: mm.BetSharePercentage,
			MentorRating:       mm.MenteeRating,
			ActionsCount:       mm.ActionsCount,
			IsLeadMentor:       mm.IsLeadMentor,
			RewardAmount:       mm.EarnedFromMentoring,
		}

		// 보상 비중 계산
		if mentorPool.DistributedAmount > 0 {
			reward.RewardPercentage = (float64(reward.RewardAmount) / float64(mentorPool.DistributedAmount)) * 100
		}

		mentorRewards = append(mentorRewards, reward)
	}

	result := &RewardDistributionResult{
		MilestoneID:         milestoneID,
		ProjectID:           mentorPool.ProjectID,
		TotalPoolAmount:     mentorPool.TotalPoolAmount,
		DistributedAmount:   mentorPool.DistributedAmount,
		EligibleMentorCount: mentorPool.EligibleMentorsCount,
		BettingAmountWeight: mentorPool.BettingAmountWeight,
		MentorRatingWeight:  mentorPool.MentorRatingWeight,
		MentorRewards:       mentorRewards,
		DistributedAt:       *mentorPool.DistributedAt,
	}

	return result, nil
}

// ProcessExpiredMilestonePools 실패한 마일스톤의 풀 처리 (환불)
func (mrs *MentorRewardService) ProcessExpiredMilestonePools(milestoneID uint) error {
	// 멘토 풀 조회
	var mentorPool models.MentorPool
	if err := mrs.db.Where("milestone_id = ?", milestoneID).First(&mentorPool).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil // 풀이 없으면 처리할 것 없음
		}
		return err
	}

	if mentorPool.IsDistributed {
		return nil // 이미 처리됨
	}

	// 실패한 마일스톤의 경우 환불 처리
	// 여기서는 단순히 분배 완료로 마킹 (실제 환불 로직은 별도 구현 필요)
	now := time.Now()
	mentorPool.IsDistributed = true
	mentorPool.DistributedAmount = 0 // 환불이므로 분배 금액은 0
	mentorPool.DistributedAt = &now
	mentorPool.EligibleMentorsCount = 0

	if err := mrs.db.Save(&mentorPool).Error; err != nil {
		return err
	}

	log.Printf("💸 Processed expired mentor pool for milestone %d (amount: $%.2f refunded)",
		milestoneID, float64(mentorPool.TotalPoolAmount)/100)

	return nil
}

// broadcastRewardDistribution 보상 분배 결과 브로드캐스트
func (mrs *MentorRewardService) broadcastRewardDistribution(result *RewardDistributionResult) {
	if mrs.sseService == nil {
		return
	}

	event := MarketUpdateEvent{
		MilestoneID: result.MilestoneID,
		MarketData: map[string]interface{}{
			"event_type": "mentor_rewards_distributed",
			"data":       result,
		},
		Timestamp: time.Now().Unix(),
	}

	mrs.sseService.BroadcastMarketUpdate(event)
}
