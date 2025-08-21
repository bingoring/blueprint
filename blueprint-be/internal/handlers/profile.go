package handlers

import (
	"blueprint-module/pkg/models"
	"blueprint/internal/database"
	"blueprint/internal/middleware"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	// Import scheduler models for cache access
	schedulerModels "blueprint-scheduler/pkg/models"
)

// ProfileHandler 프로필 관련 핸들러
type ProfileHandler struct{}

// NewProfileHandler ProfileHandler 인스턴스 생성
func NewProfileHandler() *ProfileHandler {
	return &ProfileHandler{}
}

// ProfileStats 프로필 통계 데이터
type ProfileStats struct {
	ProjectSuccessRate   float64 `json:"projectSuccessRate"`   // 프로젝트 성공률
	MentoringSuccessRate float64 `json:"mentoringSuccessRate"` // 멘토링 성공률 (임시로 0)
	TotalInvestment      int64   `json:"totalInvestment"`      // 총 투자액 (USDC cents)
	SbtCount             int     `json:"sbtCount"`             // SBT 개수
}

// CurrentProject 현재 진행 프로젝트
type CurrentProject struct {
	ID       uint   `json:"id"`
	Title    string `json:"title"`
	Progress int    `json:"progress"`
	Category string `json:"category"`
	Status   string `json:"status"`
}

// FeaturedProject 대표 프로젝트
type FeaturedProject struct {
	ID          uint    `json:"id"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Status      string  `json:"status"`
	Investment  int64   `json:"investment"`  // 받은 투자액
	SuccessRate float64 `json:"successRate"` // 성공률
}

// RecentActivity 최근 활동
type RecentActivity struct {
	ID          uint   `json:"id"`
	Type        string `json:"type"`        // investment, milestone, project 등
	Description string `json:"description"` // 활동 설명
	Timestamp   string `json:"timestamp"`   // "2시간 전" 형태
}

// ProfileResponse 프로필 페이지 응답 데이터
type ProfileResponse struct {
	Username         string            `json:"username"`
	DisplayName      string            `json:"displayName"`
	Bio              string            `json:"bio"`
	Avatar           string            `json:"avatar"`
	JoinedDate       string            `json:"joinedDate"`
	Stats            ProfileStats      `json:"stats"`
	CurrentProjects  []CurrentProject  `json:"currentProjects"`
	FeaturedProjects []FeaturedProject `json:"featuredProjects"`
	RecentActivities []RecentActivity  `json:"recentActivities"`
}

// GetUserProfile 사용자 프로필 정보 조회 (목데이터와 동일한 구조)
func (h *ProfileHandler) GetUserProfile(c *gin.Context) {
	username := c.Param("username")
	if username == "" {
		middleware.BadRequest(c, "Username is required")
		return
	}

	db := database.GetDB()

	// 사용자 조회
	var user models.User
	if err := db.Preload("Profile").Where("username = ?", username).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			middleware.NotFound(c, "User not found")
		} else {
			middleware.InternalServerError(c, "Failed to fetch user")
		}
		return
	}

	// 프로필 통계 계산
	stats := h.calculateProfileStats(user.ID)

	// 현재 진행 프로젝트 조회 (활성 상태)
	currentProjects := h.getCurrentProjects(user.ID)

	// 대표 프로젝트 조회 (완료된 프로젝트 중 상위)
	featuredProjects := h.getFeaturedProjects(user.ID)

	// 최근 활동 조회
	recentActivities := h.getRecentActivities(user.ID)

	// 아바타 URL 생성 (항상 dicebear 사용)
	avatar := "https://api.dicebear.com/6.x/avataaars/svg?seed=" + user.Username

	// 응답 구성
	displayName := user.Username // 기본값은 username
	if user.Profile != nil && user.Profile.DisplayName != "" {
		displayName = user.Profile.DisplayName
	}

	response := ProfileResponse{
		Username:         user.Username,
		DisplayName:      displayName,
		Bio:              "",
		Avatar:           avatar,
		JoinedDate:       user.CreatedAt.Format("2006-01-02"),
		Stats:            stats,
		CurrentProjects:  currentProjects,
		FeaturedProjects: featuredProjects,
		RecentActivities: recentActivities,
	}

	// 프로필이 있으면 bio 설정
	if user.Profile != nil {
		response.Bio = user.Profile.Bio
	}

	middleware.Success(c, response, "Profile retrieved successfully")
}

// calculateProfileStats 프로필 통계 계산 (캐시 데이터 우선 사용)
func (h *ProfileHandler) calculateProfileStats(userID uint) ProfileStats {
	db := database.GetDB()

	// 캐시된 통계 조회
	var cachedStats schedulerModels.UserStatsCache
	err := db.Where("user_id = ?", userID).First(&cachedStats).Error

	if err == nil {
		// 캐시된 데이터 사용 (1시간 이내)
		if time.Since(cachedStats.LastCalculatedAt) < time.Hour {
			return ProfileStats{
				ProjectSuccessRate:   cachedStats.ProjectSuccessRate,
				MentoringSuccessRate: cachedStats.MentoringSuccessRate,
				TotalInvestment:      cachedStats.TotalInvestment,
				SbtCount:             cachedStats.SbtCount,
			}
		}
	}

	// 캐시가 없거나 오래된 경우 실시간 계산 (fallback)
	// 프로젝트 관련 통계
	var totalProjects int64
	var completedProjects int64

	db.Model(&models.Project{}).Where("user_id = ?", userID).Count(&totalProjects)
	db.Model(&models.Project{}).Where("user_id = ? AND status = ?", userID, "completed").Count(&completedProjects)

	projectSuccessRate := float64(0)
	if totalProjects > 0 {
		projectSuccessRate = (float64(completedProjects) / float64(totalProjects)) * 100
	}

	// 총 투자액 계산 (Position 모델에서)
	var totalInvestment int64
	db.Model(&models.Position{}).
		Where("user_id = ?", userID).
		Select("COALESCE(SUM(quantity * (avg_price * 100)), 0)").
		Scan(&totalInvestment)

	// SBT 개수 (임시로 사용자 ID 기반 계산)
	sbtCount := int(userID % 15) // 0-14 범위의 임시 값

	// 멘토링 성공률 (멘토링 시스템 구현 전까지 임시값)
	mentoringSuccessRate := float64(85) // 임시 고정값

	return ProfileStats{
		ProjectSuccessRate:   projectSuccessRate,
		MentoringSuccessRate: mentoringSuccessRate,
		TotalInvestment:      totalInvestment,
		SbtCount:             sbtCount,
	}
}

// getCurrentProjects 현재 진행 중인 프로젝트 조회
func (h *ProfileHandler) getCurrentProjects(userID uint) []CurrentProject {
	db := database.GetDB()
	var projects []models.Project

	// 활성 상태인 프로젝트들만 조회 (최대 5개)
	db.Where("user_id = ? AND status IN ?", userID, []string{"active", "in_progress"}).
		Order("updated_at DESC").
		Limit(5).
		Find(&projects)

	var result []CurrentProject
	for _, project := range projects {
		// 진행률 계산 (완료된 마일스톤 비율)
		progress := h.calculateProjectProgress(project.ID)

		result = append(result, CurrentProject{
			ID:       project.ID,
			Title:    project.Title,
			Progress: progress,
			Category: string(project.Category),
			Status:   "active",
		})
	}

	return result
}

// getFeaturedProjects 대표 프로젝트 조회 (완료된 프로젝트)
func (h *ProfileHandler) getFeaturedProjects(userID uint) []FeaturedProject {
	db := database.GetDB()
	var projects []models.Project

	// 완료된 프로젝트들 조회 (최대 4개)
	db.Where("user_id = ? AND status = ?", userID, "completed").
		Order("updated_at DESC").
		Limit(4).
		Find(&projects)

	var result []FeaturedProject
	for _, project := range projects {
		// 임시 데이터 (실제로는 투자 및 성과 데이터에서 계산)
		investment := int64((project.ID % 10) * 100000) // 임시 투자액
		successRate := float64(80 + (project.ID % 20))  // 80-99% 범위의 임시 성공률

		result = append(result, FeaturedProject{
			ID:          project.ID,
			Title:       project.Title,
			Description: project.Description,
			Status:      "completed",
			Investment:  investment,
			SuccessRate: successRate,
		})
	}

	return result
}

// getRecentActivities 최근 활동 조회
func (h *ProfileHandler) getRecentActivities(userID uint) []RecentActivity {
	db := database.GetDB()
	var activities []models.ActivityLog

	// 최근 활동 조회 (최대 5개)
	db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(5).
		Find(&activities)

	var result []RecentActivity
	for _, activity := range activities {
		description := h.formatActivityDescription(activity)
		timestamp := h.formatRelativeTime(activity.CreatedAt)

		result = append(result, RecentActivity{
			ID:          activity.ID,
			Type:        activity.ActivityType,
			Description: description,
			Timestamp:   timestamp,
		})
	}

	return result
}

// calculateProjectProgress 프로젝트 진행률 계산
func (h *ProfileHandler) calculateProjectProgress(projectID uint) int {
	db := database.GetDB()

	var totalMilestones int64
	var completedMilestones int64

	db.Model(&models.Milestone{}).Where("project_id = ?", projectID).Count(&totalMilestones)
	db.Model(&models.Milestone{}).Where("project_id = ? AND status = ?", projectID, "completed").Count(&completedMilestones)

	if totalMilestones == 0 {
		return 0
	}

	return int((float64(completedMilestones) / float64(totalMilestones)) * 100)
}

// formatActivityDescription 활동 설명 포맷팅
func (h *ProfileHandler) formatActivityDescription(activity models.ActivityLog) string {
	switch activity.ActivityType {
	case "project":
		switch activity.Action {
		case "create":
			return "새로운 프로젝트 '" + activity.Metadata.ProjectTitle + "'를 생성했습니다."
		case "complete":
			return "프로젝트 '" + activity.Metadata.ProjectTitle + "'를 성공적으로 완료했습니다."
		}
	case "milestone":
		switch activity.Action {
		case "complete":
			return "마일스톤 '" + activity.Metadata.MilestoneTitle + "'를 성공적으로 완료했습니다."
		}
	case "investment":
		switch activity.Action {
		case "create":
			return "프로젝트에 " + formatAmount(activity.Metadata.Amount) + " USDC를 투자했습니다."
		}
	case "account":
		switch activity.Action {
		case "register":
			return "Blueprint에 가입했습니다!"
		}
	}

	return activity.Description
}

// formatRelativeTime 상대 시간 포맷팅 ("2시간 전" 형태)
func (h *ProfileHandler) formatRelativeTime(t time.Time) string {
	now := time.Now()
	diff := now.Sub(t)

	if diff < time.Minute {
		return "방금 전"
	} else if diff < time.Hour {
		minutes := int(diff.Minutes())
		return formatPlural(minutes, "분") + " 전"
	} else if diff < 24*time.Hour {
		hours := int(diff.Hours())
		return formatPlural(hours, "시간") + " 전"
	} else if diff < 7*24*time.Hour {
		days := int(diff.Hours() / 24)
		return formatPlural(days, "일") + " 전"
	} else {
		return t.Format("2006-01-02")
	}
}

// 헬퍼 함수들
func formatAmount(amount float64) string {
	if amount >= 1000000 {
		return formatFloat(amount/1000000) + "M"
	} else if amount >= 1000 {
		return formatFloat(amount/1000) + "K"
	}
	return formatFloat(amount)
}

func formatFloat(f float64) string {
	if f == float64(int(f)) {
		return fmt.Sprintf("%.0f", f)
	}
	return fmt.Sprintf("%.1f", f)
}

func formatPlural(count int, unit string) string {
	return fmt.Sprintf("%d%s", count, unit)
}
