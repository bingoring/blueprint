package services

import (
	"blueprint-module/pkg/models"
	"fmt"
	"log"
	"sort"
	"time"

	"gorm.io/gorm"
)

// 🧭 멘토 자격 증명 서비스 - "Proof of Confidence"
type MentorQualificationService struct {
	db         *gorm.DB
	sseService *SSEService
}

// NewMentorQualificationService 멘토 자격 증명 서비스 생성자
func NewMentorQualificationService(db *gorm.DB, sseService *SSEService) *MentorQualificationService {
	return &MentorQualificationService{
		db:         db,
		sseService: sseService,
	}
}

// BettorInfo 베팅자 정보 (내부 계산용)
type BettorInfo struct {
	UserID          uint    `json:"user_id"`
	TotalBetAmount  int64   `json:"total_bet_amount"`
	SharePercentage float64 `json:"share_percentage"`
	OrderCount      int     `json:"order_count"`
	LatestBetTime   time.Time `json:"latest_bet_time"`
}

// MentorQualificationResult 멘토 자격 증명 결과
type MentorQualificationResult struct {
	MilestoneID      uint   `json:"milestone_id"`
	ProjectID        uint   `json:"project_id"`
	TotalBettors     int    `json:"total_bettors"`
	LeadMentorsCount int    `json:"lead_mentors_count"`
	TotalBetAmount   int64  `json:"total_bet_amount"`
	NewMentors       []uint `json:"new_mentors"`        // 새로 생성된 멘토 ID들
	UpdatedMentors   []uint `json:"updated_mentors"`    // 업데이트된 멘토 ID들
	ProcessedAt      time.Time `json:"processed_at"`
}

// ProcessMilestoneBetting 특정 마일스톤의 베팅 정보를 처리하여 멘토 자격 부여
func (mqs *MentorQualificationService) ProcessMilestoneBetting(milestoneID uint) (*MentorQualificationResult, error) {
	log.Printf("🎯 Processing mentor qualification for milestone %d", milestoneID)

	// 트랜잭션 시작
	tx := mqs.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 1. 마일스톤 정보 조회
	var milestone models.Milestone
	if err := tx.Where("id = ?", milestoneID).First(&milestone).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("milestone not found: %v", err)
	}

	// 2. 해당 마일스톤의 '성공' 베팅자들 분석
	bettors, totalBetAmount, err := mqs.analyzeMilestoneBettors(tx, milestoneID, "success")
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to analyze bettors: %v", err)
	}

	if len(bettors) == 0 {
		log.Printf("📋 No bettors found for milestone %d", milestoneID)
		return &MentorQualificationResult{
			MilestoneID:    milestoneID,
			ProjectID:      milestone.ProjectID,
			TotalBettors:   0,
			ProcessedAt:    time.Now(),
		}, nil
	}

	// 3. 리드 멘토 수 계산 (상위 10% 또는 최소 3명, 최대 10명)
	leadMentorCount := mqs.calculateLeadMentorCount(len(bettors))

	// 4. 멘토 프로필 생성/업데이트 및 MentorMilestone 처리
	newMentors := []uint{}
	updatedMentors := []uint{}

	for i, bettor := range bettors {
		// 멘토 프로필 확인/생성
		mentorID, isNew, err := mqs.ensureMentorProfile(tx, bettor.UserID)
		if err != nil {
			log.Printf("❌ Failed to ensure mentor profile for user %d: %v", bettor.UserID, err)
			continue
		}

		if isNew {
			newMentors = append(newMentors, mentorID)
		} else {
			updatedMentors = append(updatedMentors, mentorID)
		}

		// MentorMilestone 생성/업데이트
		isLeadMentor := i < leadMentorCount
		leadMentorRank := 0
		if isLeadMentor {
			leadMentorRank = i + 1
		}

		if err := mqs.updateMentorMilestone(tx, mentorID, milestoneID, milestone.ProjectID, &bettor, isLeadMentor, leadMentorRank); err != nil {
			log.Printf("❌ Failed to update mentor milestone for mentor %d: %v", mentorID, err)
			continue
		}
	}

	// 5. 멘토 풀 생성
	if err := mqs.ensureMentorPool(tx, milestoneID, milestone.ProjectID); err != nil {
		log.Printf("⚠️ Failed to create mentor pool: %v", err)
		// 풀 생성 실패는 치명적이지 않으므로 계속 진행
	}

	// 6. 트랜잭션 커밋
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %v", err)
	}

	result := &MentorQualificationResult{
		MilestoneID:      milestoneID,
		ProjectID:        milestone.ProjectID,
		TotalBettors:     len(bettors),
		LeadMentorsCount: leadMentorCount,
		TotalBetAmount:   totalBetAmount,
		NewMentors:       newMentors,
		UpdatedMentors:   updatedMentors,
		ProcessedAt:      time.Now(),
	}

	log.Printf("✅ Mentor qualification completed for milestone %d: %d bettors, %d lead mentors",
		milestoneID, result.TotalBettors, result.LeadMentorsCount)

	// 7. 실시간 알림 브로드캐스트
	go mqs.broadcastQualificationUpdate(result)

	return result, nil
}

// analyzeMilestoneBettors 마일스톤의 베팅자들 분석 (베팅액 큰 순으로 정렬)
func (mqs *MentorQualificationService) analyzeMilestoneBettors(tx *gorm.DB, milestoneID uint, optionID string) ([]BettorInfo, int64, error) {
	// 해당 마일스톤의 성공 베팅 주문들 조회
	var orders []models.Order
	if err := tx.Where("milestone_id = ? AND option_id = ? AND side = ? AND (status = ? OR status = ? OR filled > 0)",
		milestoneID, optionID, models.OrderSideBuy, models.OrderStatusFilled, models.OrderStatusPartial).
		Find(&orders).Error; err != nil {
		return nil, 0, err
	}

	// 사용자별 베팅 정보 집계
	userBets := make(map[uint]*BettorInfo)
	var totalBetAmount int64

	for _, order := range orders {
		betAmount := int64(float64(order.Filled) * order.Price * 100) // 실제 체결된 금액만

		if existing, exists := userBets[order.UserID]; exists {
			existing.TotalBetAmount += betAmount
			existing.OrderCount++
			if order.CreatedAt.After(existing.LatestBetTime) {
				existing.LatestBetTime = order.CreatedAt
			}
		} else {
			userBets[order.UserID] = &BettorInfo{
				UserID:         order.UserID,
				TotalBetAmount: betAmount,
				OrderCount:     1,
				LatestBetTime:  order.CreatedAt,
			}
		}
		totalBetAmount += betAmount
	}

	// 베팅 비중 계산
	bettors := make([]BettorInfo, 0, len(userBets))
	for _, bettor := range userBets {
		if totalBetAmount > 0 {
			bettor.SharePercentage = (float64(bettor.TotalBetAmount) / float64(totalBetAmount)) * 100
		}
		bettors = append(bettors, *bettor)
	}

	// 베팅액 큰 순으로 정렬 (같으면 일찍 베팅한 순)
	sort.Slice(bettors, func(i, j int) bool {
		if bettors[i].TotalBetAmount == bettors[j].TotalBetAmount {
			return bettors[i].LatestBetTime.Before(bettors[j].LatestBetTime)
		}
		return bettors[i].TotalBetAmount > bettors[j].TotalBetAmount
	})

	return bettors, totalBetAmount, nil
}

// calculateLeadMentorCount 리드 멘토 수 계산
func (mqs *MentorQualificationService) calculateLeadMentorCount(totalBettors int) int {
	// 상위 10% 또는 최소 3명, 최대 10명
	leadCount := totalBettors / 10
	if leadCount < 3 {
		leadCount = 3
	}
	if leadCount > 10 {
		leadCount = 10
	}
	if leadCount > totalBettors {
		leadCount = totalBettors
	}
	return leadCount
}

// ensureMentorProfile 멘토 프로필 확인/생성
func (mqs *MentorQualificationService) ensureMentorProfile(tx *gorm.DB, userID uint) (uint, bool, error) {
	var mentor models.Mentor
	err := tx.Where("user_id = ?", userID).First(&mentor).Error

	if err == nil {
		// 기존 멘토 프로필 존재
		return mentor.ID, false, nil
	}

	if err != gorm.ErrRecordNotFound {
		return 0, false, err
	}

	// 새 멘토 프로필 생성
	// 사용자 정보 조회
	var user models.User
	if err := tx.Where("id = ?", userID).First(&user).Error; err != nil {
		return 0, false, fmt.Errorf("user not found: %v", err)
	}

	// 기본 멘토 프로필 생성
	mentor = models.Mentor{
		UserID:              userID,
		Status:              models.MentorStatusActive,
		Tier:                models.MentorTierBronze,
		Bio:                 fmt.Sprintf("Mentor qualified through betting on milestone success"),
		IsAvailable:         true,
		MaxActiveMentorings: 5,
		ReputationScore:     10, // 초기 점수
		TrustScore:          5.0, // 초기 신뢰도
	}

	if err := tx.Create(&mentor).Error; err != nil {
		return 0, false, err
	}

	log.Printf("✨ Created new mentor profile for user %d (mentor ID: %d)", userID, mentor.ID)
	return mentor.ID, true, nil
}

// updateMentorMilestone MentorMilestone 생성/업데이트
func (mqs *MentorQualificationService) updateMentorMilestone(tx *gorm.DB, mentorID, milestoneID, projectID uint, bettor *BettorInfo, isLeadMentor bool, leadMentorRank int) error {
	var mentorMilestone models.MentorMilestone
	err := tx.Where("mentor_id = ? AND milestone_id = ?", mentorID, milestoneID).First(&mentorMilestone).Error

	now := time.Now()

	if err == gorm.ErrRecordNotFound {
		// 새 MentorMilestone 생성
		mentorMilestone = models.MentorMilestone{
			MentorID:           mentorID,
			MilestoneID:        milestoneID,
			ProjectID:          projectID,
			TotalBetAmount:     bettor.TotalBetAmount,
			BetSharePercentage: bettor.SharePercentage,
			IsLeadMentor:       isLeadMentor,
			LeadMentorRank:     leadMentorRank,
			IsActive:           false, // 아직 멘토링 시작 전
			CreatedAt:          now,
			UpdatedAt:          now,
		}

		if err := tx.Create(&mentorMilestone).Error; err != nil {
			return err
		}

		log.Printf("🆕 Created MentorMilestone: mentor %d, milestone %d, amount $%.2f, lead: %v (rank: %d)",
			mentorID, milestoneID, float64(bettor.TotalBetAmount)/100, isLeadMentor, leadMentorRank)
	} else if err != nil {
		return err
	} else {
		// 기존 MentorMilestone 업데이트
		mentorMilestone.TotalBetAmount = bettor.TotalBetAmount
		mentorMilestone.BetSharePercentage = bettor.SharePercentage
		mentorMilestone.IsLeadMentor = isLeadMentor
		mentorMilestone.LeadMentorRank = leadMentorRank
		mentorMilestone.UpdatedAt = now

		if err := tx.Save(&mentorMilestone).Error; err != nil {
			return err
		}

		log.Printf("🔄 Updated MentorMilestone: mentor %d, milestone %d, amount $%.2f, lead: %v (rank: %d)",
			mentorID, milestoneID, float64(bettor.TotalBetAmount)/100, isLeadMentor, leadMentorRank)
	}

	return nil
}

// ensureMentorPool 멘토 풀 생성 확인
func (mqs *MentorQualificationService) ensureMentorPool(tx *gorm.DB, milestoneID, projectID uint) error {
	var mentorPool models.MentorPool
	err := tx.Where("milestone_id = ?", milestoneID).First(&mentorPool).Error

	if err == gorm.ErrRecordNotFound {
		// 새 멘토 풀 생성
		mentorPool = models.MentorPool{
			MilestoneID:         milestoneID,
			ProjectID:           projectID,
			FeePercentage:       50.0, // 거래 수수료의 50%
			PerformanceWeighted: true,
			MentorRatingWeight:  30.0,
			BettingAmountWeight: 70.0,
		}

		if err := tx.Create(&mentorPool).Error; err != nil {
			return err
		}

		log.Printf("💰 Created mentor pool for milestone %d", milestoneID)
	} else if err != nil {
		return err
	}
	// 이미 존재하면 그대로 둠

	return nil
}

// ProcessAllActiveMilestones 모든 활성 마일스톤의 멘토 자격 처리
func (mqs *MentorQualificationService) ProcessAllActiveMilestones() error {
	log.Printf("🔄 Processing mentor qualification for all active milestones...")

	// 활성 마일스톤들 조회 (펀딩 성공한 것들)
	var milestones []models.Milestone
	if err := mqs.db.Where("status IN ?", []models.MilestoneStatus{
		models.MilestoneStatusActive,
		models.MilestoneStatusPending, // 구버전 호환
	}).Find(&milestones).Error; err != nil {
		return fmt.Errorf("failed to query active milestones: %v", err)
	}

	processed := 0
	errors := 0

	for _, milestone := range milestones {
		if _, err := mqs.ProcessMilestoneBetting(milestone.ID); err != nil {
			log.Printf("❌ Failed to process milestone %d: %v", milestone.ID, err)
			errors++
		} else {
			processed++
		}
	}

	log.Printf("✅ Mentor qualification batch completed: %d processed, %d errors", processed, errors)
	return nil
}

// GetMentorCandidates 특정 마일스톤의 멘토 후보들 조회
func (mqs *MentorQualificationService) GetMentorCandidates(milestoneID uint) ([]models.MentorMilestone, error) {
	var mentorMilestones []models.MentorMilestone
	if err := mqs.db.Where("milestone_id = ?", milestoneID).
		Preload("Mentor").Preload("Mentor.User").
		Order("total_bet_amount DESC, is_lead_mentor DESC").
		Find(&mentorMilestones).Error; err != nil {
		return nil, err
	}

	return mentorMilestones, nil
}

// broadcastQualificationUpdate 멘토 자격 증명 결과 실시간 브로드캐스트
func (mqs *MentorQualificationService) broadcastQualificationUpdate(result *MentorQualificationResult) {
	if mqs.sseService == nil {
		return
	}

	// 마일스톤별 채널에 브로드캐스트
	event := MarketUpdateEvent{
		MilestoneID: result.MilestoneID,
		MarketData: map[string]interface{}{
			"event_type": "mentor_qualification_update",
			"data":       result,
		},
		Timestamp: time.Now().Unix(),
	}

	mqs.sseService.BroadcastMarketUpdate(event)
}
