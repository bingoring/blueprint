package jobs

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"blueprint-module/pkg/database"
	moduleModels "blueprint-module/pkg/models"
	"blueprint-scheduler/pkg/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type StatsCalculator struct {
	db *gorm.DB
}

func NewStatsCalculator() *StatsCalculator {
	return &StatsCalculator{
		db: database.GetDB(),
	}
}

// CalculateUserStats 사용자별 통계 계산 작업
func (sc *StatsCalculator) CalculateUserStats(ctx context.Context, userID uint) error {
	log.Printf("📊 사용자 %d 통계 계산 시작", userID)

	// 1. 프로젝트 성공률 계산
	var totalProjects, completedProjects int64
	sc.db.Model(&moduleModels.Project{}).Where("user_id = ?", userID).Count(&totalProjects)
	sc.db.Model(&moduleModels.Project{}).Where("user_id = ? AND status = ?", userID, "completed").Count(&completedProjects)
	
	projectSuccessRate := float64(0)
	if totalProjects > 0 {
		projectSuccessRate = (float64(completedProjects) / float64(totalProjects)) * 100
	}

	// 2. 총 투자액 계산 (실제로는 Position 모델에서)
	var totalInvestment int64
	sc.db.Model(&moduleModels.Position{}).
		Where("user_id = ?", userID).
		Select("COALESCE(SUM(quantity * (avg_price * 100)), 0)").
		Scan(&totalInvestment)

	// 3. 멘토링 성공률 계산 (임시로 85% 고정)
	mentoringSuccessRate := float64(85)

	// 4. SBT 개수 계산 (임시로 user_id % 15)
	sbtCount := int(userID % 15)

	// 5. 캐시에 저장
	userStats := models.UserStatsCache{
		UserID:               userID,
		ProjectSuccessRate:   projectSuccessRate,
		MentoringSuccessRate: mentoringSuccessRate,
		TotalInvestment:      totalInvestment,
		SbtCount:             sbtCount,
		LastCalculatedAt:     time.Now(),
	}

	err := sc.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"project_success_rate", "mentoring_success_rate", "total_investment", "sbt_count", "last_calculated_at", "updated_at"}),
	}).Create(&userStats).Error

	if err != nil {
		return fmt.Errorf("사용자 통계 저장 실패: %w", err)
	}

	log.Printf("✅ 사용자 %d 통계 계산 완료", userID)
	return nil
}

// CalculateProjectStats 프로젝트별 통계 계산 작업
func (sc *StatsCalculator) CalculateProjectStats(ctx context.Context, projectID uint) error {
	log.Printf("📊 프로젝트 %d 통계 계산 시작", projectID)

	// 1. 총 투자액 계산
	var totalInvestment int64
	var investorCount int64

	sc.db.Model(&moduleModels.Position{}).
		Joins("JOIN milestones ON positions.milestone_id = milestones.id").
		Where("milestones.project_id = ?", projectID).
		Select("COALESCE(SUM(quantity * (avg_price * 100)), 0)").
		Scan(&totalInvestment)

	sc.db.Model(&moduleModels.Position{}).
		Joins("JOIN milestones ON positions.milestone_id = milestones.id").
		Where("milestones.project_id = ?", projectID).
		Distinct("user_id").
		Count(&investorCount)

	// 2. 완료율 계산
	var totalMilestones, completedMilestones int64
	sc.db.Model(&moduleModels.Milestone{}).Where("project_id = ?", projectID).Count(&totalMilestones)
	sc.db.Model(&moduleModels.Milestone{}).Where("project_id = ? AND status = ?", projectID, "completed").Count(&completedMilestones)

	completionRate := float64(0)
	if totalMilestones > 0 {
		completionRate = (float64(completedMilestones) / float64(totalMilestones)) * 100
	}

	// 3. 성공 확률 계산 (임시 로직)
	successProbability := completionRate * 0.8 + float64(investorCount)*0.1

	// 4. 캐시에 저장
	projectStats := models.ProjectStatsCache{
		ProjectID:          projectID,
		TotalInvestment:    totalInvestment,
		InvestorCount:      int(investorCount),
		SuccessProbability: successProbability,
		CompletionRate:     completionRate,
		LastCalculatedAt:   time.Now(),
	}

	err := sc.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "project_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"total_investment", "investor_count", "success_probability", "completion_rate", "last_calculated_at", "updated_at"}),
	}).Create(&projectStats).Error

	if err != nil {
		return fmt.Errorf("프로젝트 통계 저장 실패: %w", err)
	}

	log.Printf("✅ 프로젝트 %d 통계 계산 완료", projectID)
	return nil
}

// CalculateDashboardCache 대시보드 캐시 계산 작업
func (sc *StatsCalculator) CalculateDashboardCache(ctx context.Context, userID uint) error {
	log.Printf("📊 사용자 %d 대시보드 캐시 계산 시작", userID)

	// 1. 추천 프로젝트 조회 (최근 업데이트된 프로젝트들)
	var featuredProjects []moduleModels.Project
	sc.db.Order("updated_at DESC").Limit(6).Find(&featuredProjects)

	// 2. 활동 피드 조회 (사용자의 최근 활동)
	var activityFeed []moduleModels.ActivityLog
	sc.db.Where("user_id = ?", userID).Order("created_at DESC").Limit(5).Find(&activityFeed)

	// 3. 포트폴리오 계산
	var positions []moduleModels.Position
	sc.db.Where("user_id = ?", userID).Find(&positions)
	
	totalValue := int64(0)
	for _, pos := range positions {
		totalValue += pos.Quantity * int64(pos.AvgPrice * 100) // Convert price to cents
	}

	portfolio := map[string]interface{}{
		"totalInvested":  totalValue,
		"currentValue":   totalValue * 115 / 100, // 임시 15% 수익률
		"profit":         totalValue * 15 / 100,
		"profitPercent":  15.0,
		"blueprintTokens": 12500,
	}

	// 4. 다음 마일스톤
	nextMilestone := map[string]interface{}{
		"title":       "다음 마일스톤",
		"daysLeft":    30,
		"progress":    60,
		"mentorName":  "시스템",
		"mentorAvatar": "https://api.dicebear.com/6.x/avataaars/svg?seed=system",
	}

	// JSON 변환
	featuredProjectsJSON, _ := json.Marshal(featuredProjects)
	activityFeedJSON, _ := json.Marshal(activityFeed)
	portfolioJSON, _ := json.Marshal(portfolio)
	nextMilestoneJSON, _ := json.Marshal(nextMilestone)

	// 5. 캐시에 저장
	dashboardCache := models.DashboardCache{
		UserID:           userID,
		FeaturedProjects: string(featuredProjectsJSON),
		ActivityFeed:     string(activityFeedJSON),
		Portfolio:        string(portfolioJSON),
		NextMilestone:    string(nextMilestoneJSON),
		LastCalculatedAt: time.Now(),
	}

	err := sc.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"featured_projects", "activity_feed", "portfolio", "next_milestone", "last_calculated_at", "updated_at"}),
	}).Create(&dashboardCache).Error

	if err != nil {
		return fmt.Errorf("대시보드 캐시 저장 실패: %w", err)
	}

	log.Printf("✅ 사용자 %d 대시보드 캐시 계산 완료", userID)
	return nil
}

// CalculateGlobalStats 글로벌 통계 계산 작업
func (sc *StatsCalculator) CalculateGlobalStats(ctx context.Context) error {
	log.Printf("📊 글로벌 통계 계산 시작")

	// 활성 사용자 수
	var activeUsers int64
	sc.db.Model(&moduleModels.User{}).Where("updated_at > ?", time.Now().AddDate(0, 0, -30)).Count(&activeUsers)
	sc.updateGlobalStat(models.GlobalStatActiveUsers, float64(activeUsers))

	// 활성 프로젝트 수
	var activeProjects int64 
	sc.db.Model(&moduleModels.Project{}).Where("status IN ?", []string{"active", "in_progress"}).Count(&activeProjects)
	sc.updateGlobalStat(models.GlobalStatActiveProjects, float64(activeProjects))

	// 활성 분쟁 수
	var activeDisputes int64
	sc.db.Model(&moduleModels.Dispute{}).Where("status IN ?", []string{"challenge_window", "voting_period"}).Count(&activeDisputes)
	sc.updateGlobalStat(models.GlobalStatActiveDisputes, float64(activeDisputes))

	log.Printf("✅ 글로벌 통계 계산 완료")
	return nil
}

func (sc *StatsCalculator) updateGlobalStat(key string, value float64) {
	globalStat := models.GlobalStatsCache{
		StatKey:          key,
		StatValue:        value,
		LastCalculatedAt: time.Now(),
	}

	sc.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "stat_key"}},
		DoUpdates: clause.AssignmentColumns([]string{"stat_value", "last_calculated_at", "updated_at"}),
	}).Create(&globalStat)
}