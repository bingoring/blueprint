package services

import (
	"blueprint-module/pkg/models"
	"fmt"
	"log"
	"time"

	"gorm.io/gorm"
)

// 🏆 멘토 평판 서비스 - 온체인 명예 시스템
type MentorReputationService struct {
	db         *gorm.DB
	sseService *SSEService
}

// NewMentorReputationService 평판 서비스 생성자
func NewMentorReputationService(db *gorm.DB, sseService *SSEService) *MentorReputationService {
	return &MentorReputationService{
		db:         db,
		sseService: sseService,
	}
}

// ReputationEventType 평판 이벤트 타입
const (
	EventSuccessfulMentoring = "successful_mentoring"  // 성공적인 멘토링 완료
	EventMilestoneSuccess    = "milestone_success"     // 마일스톤 성공 기여
	EventHighRating          = "high_rating"           // 높은 평점 획득
	EventRewardEarned        = "mentoring_reward_earned" // 멘토링 보상 획득
	EventFirstMentoring      = "first_mentoring"       // 첫 멘토링 완료
	EventLeadMentor          = "lead_mentor_achieved"   // 리드 멘토 달성
	EventTierUpgrade         = "tier_upgrade"          // 등급 승급
	EventLongTermCommitment  = "long_term_commitment"  // 장기 헌신 (6개월+ 멘토링)
	EventExceptionalFeedback = "exceptional_feedback"  // 탁월한 피드백 (평점 9.5+)
)

// MentorAchievement 멘토 성취
type MentorAchievement struct {
	ID          uint                     `json:"id"`
	Type        string                   `json:"type"`
	Title       string                   `json:"title"`
	Description string                   `json:"description"`
	Points      int                      `json:"points"`
	Multiplier  float64                  `json:"multiplier"`
	IconURL     string                   `json:"icon_url"`
	Rarity      string                   `json:"rarity"` // "common", "rare", "epic", "legendary"
	UnlockedAt  time.Time                `json:"unlocked_at"`
}

// MentorLeaderboard 멘토 리더보드 정보
type MentorLeaderboard struct {
	Rank                 int                     `json:"rank"`
	MentorID             uint                    `json:"mentor_id"`
	UserID               uint                    `json:"user_id"`
	Username             string                  `json:"username"`
	Tier                 models.MentorTier       `json:"tier"`
	ReputationScore      int                     `json:"reputation_score"`
	SuccessRate          float64                 `json:"success_rate"`
	TotalMentorings      int                     `json:"total_mentorings"`
	TotalEarned          int64                   `json:"total_earned"`
	RecentAchievements   []MentorAchievement     `json:"recent_achievements"`
	BadgeCount           map[string]int          `json:"badge_count"` // 배지 종류별 개수
}

// RecordReputationEvent 평판 이벤트 기록
func (mrs *MentorReputationService) RecordReputationEvent(mentorID uint, eventType string, points int, multiplier float64,
	milestoneID *uint, projectID *uint, sessionID *uint, description string) (*models.MentorReputation, error) {

	// 최종 점수 계산
	finalPoints := int(float64(points) * multiplier)

	reputation := models.MentorReputation{
		MentorID:    mentorID,
		EventType:   eventType,
		Points:      finalPoints,
		Multiplier:  multiplier,
		MilestoneID: milestoneID,
		ProjectID:   projectID,
		SessionID:   sessionID,
		Description: description,
	}

	if err := mrs.db.Create(&reputation).Error; err != nil {
		return nil, err
	}

	// 멘토 총 평판 점수 업데이트
	if err := mrs.updateMentorReputationScore(mentorID, finalPoints); err != nil {
		log.Printf("⚠️ Failed to update mentor reputation score: %v", err)
	}

	// 등급 승급 확인
	go mrs.checkTierUpgrade(mentorID)

	log.Printf("🏆 Reputation event recorded: %s (+%d points, x%.1f) for mentor %d",
		eventType, points, multiplier, mentorID)

	return &reputation, nil
}

// ProcessMentoringCompletion 멘토링 완료 시 평판 이벤트 처리
func (mrs *MentorReputationService) ProcessMentoringCompletion(sessionID uint, rating float64) error {
	// 세션 정보 조회
	var session models.MentoringSession
	if err := mrs.db.Where("id = ?", sessionID).First(&session).Error; err != nil {
		return fmt.Errorf("session not found: %v", err)
	}

	// 기본 완료 점수
	basePoints := 50
	multiplier := 1.0

	// 평점에 따른 보너스
	if rating >= 9.5 {
		// 탁월한 평점 (9.5+)
		multiplier = 2.0
		mrs.RecordReputationEvent(session.MentorID, EventExceptionalFeedback, 100, 1.0,
			&session.MilestoneID, &session.ProjectID, &sessionID,
			fmt.Sprintf("Received exceptional rating of %.1f", rating))
	} else if rating >= 8.0 {
		// 높은 평점 (8.0+)
		multiplier = 1.5
		mrs.RecordReputationEvent(session.MentorID, EventHighRating, 30, 1.0,
			&session.MilestoneID, &session.ProjectID, &sessionID,
			fmt.Sprintf("Received high rating of %.1f", rating))
	}

	// 첫 멘토링 확인
	var mentorCount int64
	if err := mrs.db.Model(&models.MentoringSession{}).
		Where("mentor_id = ? AND status = ?", session.MentorID, models.SessionStatusCompleted).
		Count(&mentorCount).Error; err == nil && mentorCount == 1 {

		mrs.RecordReputationEvent(session.MentorID, EventFirstMentoring, 100, 1.0,
			&session.MilestoneID, &session.ProjectID, &sessionID,
			"Completed first mentoring session successfully")
	}

	// 메인 완료 이벤트 기록
	_, err := mrs.RecordReputationEvent(session.MentorID, EventSuccessfulMentoring, basePoints, multiplier,
		&session.MilestoneID, &session.ProjectID, &sessionID,
		fmt.Sprintf("Successfully completed mentoring session (rating: %.1f)", rating))

	return err
}

// ProcessMilestoneSuccess 마일스톤 성공 시 멘토들에게 평판 점수 부여
func (mrs *MentorReputationService) ProcessMilestoneSuccess(milestoneID uint) error {
	// 해당 마일스톤의 활성 멘토들 조회
	var mentorMilestones []models.MentorMilestone
	if err := mrs.db.Where("milestone_id = ? AND is_active = ?", milestoneID, true).
		Find(&mentorMilestones).Error; err != nil {
		return err
	}

	// 마일스톤 정보 조회
	var milestone models.Milestone
	if err := mrs.db.Where("id = ?", milestoneID).First(&milestone).Error; err != nil {
		return fmt.Errorf("milestone not found: %v", err)
	}

	for _, mm := range mentorMilestones {
		points := 75 // 기본 성공 기여 점수
		multiplier := 1.0

		// 리드 멘토 보너스
		if mm.IsLeadMentor {
			multiplier = 1.5
		}

		// 베팅 비중에 따른 추가 점수
		if mm.BetSharePercentage >= 20 {
			points += 25 // 큰 베팅자 보너스
		}

		mrs.RecordReputationEvent(mm.MentorID, EventMilestoneSuccess, points, multiplier,
			&milestoneID, &milestone.ProjectID, nil,
			fmt.Sprintf("Contributed to milestone success (%.1f%% share)", mm.BetSharePercentage))

		// 리드 멘토 성취 기록
		if mm.IsLeadMentor {
			mrs.RecordReputationEvent(mm.MentorID, EventLeadMentor, 50, 1.0,
				&milestoneID, &milestone.ProjectID, nil,
				fmt.Sprintf("Achieved lead mentor status (rank %d)", mm.LeadMentorRank))
		}
	}

	log.Printf("🎯 Processed milestone success reputation for %d mentors", len(mentorMilestones))
	return nil
}

// updateMentorReputationScore 멘토 평판 점수 업데이트
func (mrs *MentorReputationService) updateMentorReputationScore(mentorID uint, points int) error {
	return mrs.db.Model(&models.Mentor{}).Where("id = ?", mentorID).
		Update("reputation_score", gorm.Expr("reputation_score + ?", points)).Error
}

// checkTierUpgrade 등급 승급 확인 및 처리
func (mrs *MentorReputationService) checkTierUpgrade(mentorID uint) {
	// 멘토 정보 조회
	var mentor models.Mentor
	if err := mrs.db.Where("id = ?", mentorID).First(&mentor).Error; err != nil {
		return
	}

	currentTier := mentor.Tier
	newTier := mrs.calculateMentorTier(&mentor)

	if newTier != currentTier {
		// 등급 승급
		mentor.Tier = newTier
		if err := mrs.db.Save(&mentor).Error; err != nil {
			log.Printf("❌ Failed to upgrade mentor tier: %v", err)
			return
		}

		// 승급 이벤트 기록
		tierPoints := mrs.getTierUpgradePoints(newTier)
		mrs.RecordReputationEvent(mentorID, EventTierUpgrade, tierPoints, 1.0, nil, nil, nil,
			fmt.Sprintf("Upgraded from %s to %s tier", currentTier, newTier))

		log.Printf("⬆️ Mentor %d upgraded from %s to %s", mentorID, currentTier, newTier)

		// 승급 알림
		go mrs.broadcastTierUpgrade(mentorID, currentTier, newTier)
	}
}

// calculateMentorTier 멘토 등급 계산
func (mrs *MentorReputationService) calculateMentorTier(mentor *models.Mentor) models.MentorTier {
	// 성공률 업데이트
	mentor.SuccessRate = mentor.CalculateSuccessRate()

	if mentor.IsQualifiedForTier(models.MentorTierLegend) {
		return models.MentorTierLegend
	} else if mentor.IsQualifiedForTier(models.MentorTierPlatinum) {
		return models.MentorTierPlatinum
	} else if mentor.IsQualifiedForTier(models.MentorTierGold) {
		return models.MentorTierGold
	} else if mentor.IsQualifiedForTier(models.MentorTierSilver) {
		return models.MentorTierSilver
	}
	return models.MentorTierBronze
}

// getTierUpgradePoints 등급별 승급 보너스 점수
func (mrs *MentorReputationService) getTierUpgradePoints(tier models.MentorTier) int {
	switch tier {
	case models.MentorTierSilver:
		return 200
	case models.MentorTierGold:
		return 500
	case models.MentorTierPlatinum:
		return 1000
	case models.MentorTierLegend:
		return 2000
	default:
		return 0
	}
}

// GetMentorAchievements 멘토 성취 목록 조회
func (mrs *MentorReputationService) GetMentorAchievements(mentorID uint) ([]MentorAchievement, error) {
	var reputations []models.MentorReputation
	if err := mrs.db.Where("mentor_id = ?", mentorID).
		Order("created_at DESC").Find(&reputations).Error; err != nil {
		return nil, err
	}

	achievements := make([]MentorAchievement, 0, len(reputations))
	for _, rep := range reputations {
		achievement := MentorAchievement{
			ID:          rep.ID,
			Type:        rep.EventType,
			Title:       mrs.getAchievementTitle(rep.EventType),
			Description: rep.Description,
			Points:      rep.Points,
			Multiplier:  rep.Multiplier,
			IconURL:     mrs.getAchievementIcon(rep.EventType),
			Rarity:      mrs.getAchievementRarity(rep.EventType),
			UnlockedAt:  rep.CreatedAt,
		}
		achievements = append(achievements, achievement)
	}

	return achievements, nil
}

// GetLeaderboard 멘토 리더보드 조회
func (mrs *MentorReputationService) GetLeaderboard(limit int) ([]MentorLeaderboard, error) {
	if limit <= 0 {
		limit = 50 // 기본값
	}

	// 상위 멘토들 조회
	var mentors []models.Mentor
	if err := mrs.db.Preload("User").
		Order("reputation_score DESC, success_rate DESC, total_mentorings DESC").
		Limit(limit).Find(&mentors).Error; err != nil {
		return nil, err
	}

	leaderboard := make([]MentorLeaderboard, 0, len(mentors))

	for i, mentor := range mentors {
		// 최근 성취들 조회 (최근 5개)
		recentAchievements, _ := mrs.getRecentAchievements(mentor.ID, 5)

		// 배지 개수 계산
		badgeCount := mrs.calculateBadgeCount(mentor.ID)

		leader := MentorLeaderboard{
			Rank:               i + 1,
			MentorID:           mentor.ID,
			UserID:             mentor.UserID,
			Username:           mentor.User.Username,
			Tier:               mentor.Tier,
			ReputationScore:    mentor.ReputationScore,
			SuccessRate:        mentor.SuccessRate,
			TotalMentorings:    mentor.TotalMentorings,
			TotalEarned:        mentor.TotalEarnedAmount,
			RecentAchievements: recentAchievements,
			BadgeCount:         badgeCount,
		}

		leaderboard = append(leaderboard, leader)
	}

	return leaderboard, nil
}

// getRecentAchievements 최근 성취들 조회
func (mrs *MentorReputationService) getRecentAchievements(mentorID uint, limit int) ([]MentorAchievement, error) {
	var reputations []models.MentorReputation
	if err := mrs.db.Where("mentor_id = ?", mentorID).
		Order("created_at DESC").Limit(limit).Find(&reputations).Error; err != nil {
		return []MentorAchievement{}, nil
	}

	achievements := make([]MentorAchievement, 0, len(reputations))
	for _, rep := range reputations {
		achievement := MentorAchievement{
			ID:          rep.ID,
			Type:        rep.EventType,
			Title:       mrs.getAchievementTitle(rep.EventType),
			Description: rep.Description,
			Points:      rep.Points,
			Rarity:      mrs.getAchievementRarity(rep.EventType),
			UnlockedAt:  rep.CreatedAt,
		}
		achievements = append(achievements, achievement)
	}

	return achievements, nil
}

// calculateBadgeCount 배지 개수 계산
func (mrs *MentorReputationService) calculateBadgeCount(mentorID uint) map[string]int {
	badgeCount := make(map[string]int)

	var counts []struct {
		EventType string `gorm:"column:event_type"`
		Count     int    `gorm:"column:count"`
	}

	if err := mrs.db.Model(&models.MentorReputation{}).
		Select("event_type, count(*) as count").
		Where("mentor_id = ?", mentorID).
		Group("event_type").Find(&counts).Error; err != nil {
		return badgeCount
	}

	for _, count := range counts {
		badgeCount[count.EventType] = count.Count
	}

	return badgeCount
}

// Helper 메서드들
func (mrs *MentorReputationService) getAchievementTitle(eventType string) string {
	titles := map[string]string{
		EventSuccessfulMentoring:  "Mentoring Master",
		EventMilestoneSuccess:     "Success Contributor",
		EventHighRating:          "Highly Rated",
		EventRewardEarned:        "Reward Earner",
		EventFirstMentoring:      "First Steps",
		EventLeadMentor:          "Lead Mentor",
		EventTierUpgrade:         "Tier Upgrade",
		EventLongTermCommitment:  "Long-term Commitment",
		EventExceptionalFeedback: "Exceptional Mentor",
	}
	if title, exists := titles[eventType]; exists {
		return title
	}
	return "Achievement"
}

func (mrs *MentorReputationService) getAchievementIcon(eventType string) string {
	// 실제 구현시 CDN URL 또는 아이콘 경로 반환
	return fmt.Sprintf("/icons/achievements/%s.svg", eventType)
}

func (mrs *MentorReputationService) getAchievementRarity(eventType string) string {
	rarities := map[string]string{
		EventSuccessfulMentoring:  "common",
		EventMilestoneSuccess:     "common",
		EventHighRating:          "rare",
		EventRewardEarned:        "common",
		EventFirstMentoring:      "rare",
		EventLeadMentor:          "epic",
		EventTierUpgrade:         "epic",
		EventLongTermCommitment:  "legendary",
		EventExceptionalFeedback: "legendary",
	}
	if rarity, exists := rarities[eventType]; exists {
		return rarity
	}
	return "common"
}

// broadcastTierUpgrade 등급 승급 알림
func (mrs *MentorReputationService) broadcastTierUpgrade(mentorID uint, oldTier, newTier models.MentorTier) {
	if mrs.sseService == nil {
		return
	}

	event := MarketUpdateEvent{
		MilestoneID: 0, // 전역 이벤트
		MarketData: map[string]interface{}{
			"event_type": "mentor_tier_upgrade",
			"data": map[string]interface{}{
				"mentor_id": mentorID,
				"old_tier":  oldTier,
				"new_tier":  newTier,
				"timestamp": time.Now().Unix(),
			},
		},
		Timestamp: time.Now().Unix(),
	}

	mrs.sseService.BroadcastMarketUpdate(event)
}
