package handlers

import (
	"blueprint-module/pkg/models"
	"blueprint/internal/database"
	"blueprint/internal/middleware"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// ActivityHandler 활동 로그 핸들러
type ActivityHandler struct{}

// NewActivityHandler ActivityHandler 인스턴스 생성
func NewActivityHandler() *ActivityHandler {
	return &ActivityHandler{}
}

// GetUserActivities 사용자의 활동 로그 조회
func (h *ActivityHandler) GetUserActivities(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		middleware.Unauthorized(c, "User not authenticated")
		return
	}

	// 쿼리 파라미터 파싱
	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")
	activityTypes := c.QueryArray("types") // ?types=project&types=trade

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 100 {
		limit = 20
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	// 날짜 범위 파라미터
	var startDate, endDate *time.Time
	if startStr := c.Query("start_date"); startStr != "" {
		if parsed, err := time.Parse("2006-01-02", startStr); err == nil {
			startDate = &parsed
		}
	}
	if endStr := c.Query("end_date"); endStr != "" {
		if parsed, err := time.Parse("2006-01-02", endStr); err == nil {
			endDate = &parsed
		}
	}

	// 데이터베이스 쿼리 구성
	db := database.GetDB()
	query := db.Model(&models.ActivityLog{}).
		Where("user_id = ?", userID).
		Preload("Project").
		Preload("Milestone").
		Order("created_at DESC")

	// 활동 타입 필터
	if len(activityTypes) > 0 {
		query = query.Where("activity_type IN ?", activityTypes)
	}

	// 날짜 범위 필터
	if startDate != nil {
		query = query.Where("created_at >= ?", startDate)
	}
	if endDate != nil {
		query = query.Where("created_at <= ?", endDate.Add(24*time.Hour-time.Second))
	}

	// 총 개수 조회
	var total int64
	if err := query.Count(&total).Error; err != nil {
		middleware.InternalServerError(c, "Failed to count activities")
		return
	}

	// 페이지네이션 적용하여 데이터 조회
	var activities []models.ActivityLog
	if err := query.Limit(limit).Offset(offset).Find(&activities).Error; err != nil {
		middleware.InternalServerError(c, "Failed to retrieve activities")
		return
	}

	// 응답 데이터 구성
	response := map[string]interface{}{
		"activities": activities,
		"pagination": map[string]interface{}{
			"total":  total,
			"limit":  limit,
			"offset": offset,
			"pages":  (total + int64(limit) - 1) / int64(limit),
		},
	}

	middleware.Success(c, response, "Activities retrieved successfully")
}

// GetActivitySummary 사용자의 활동 요약 정보 조회 (대시보드용)
func (h *ActivityHandler) GetActivitySummary(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		middleware.Unauthorized(c, "User not authenticated")
		return
	}

	db := database.GetDB()

	// 최근 30일간의 활동 요약
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)

	// 활동 타입별 개수
	var activityCounts []struct {
		ActivityType string `json:"activity_type"`
		Count        int64  `json:"count"`
	}

	if err := db.Model(&models.ActivityLog{}).
		Select("activity_type, COUNT(*) as count").
		Where("user_id = ? AND created_at >= ?", userID, thirtyDaysAgo).
		Group("activity_type").
		Find(&activityCounts).Error; err != nil {
		middleware.InternalServerError(c, "Failed to get activity summary")
		return
	}

	// 최근 활동 (최대 5개)
	var recentActivities []models.ActivityLog
	if err := db.Model(&models.ActivityLog{}).
		Where("user_id = ?", userID).
		Preload("Project").
		Preload("Milestone").
		Order("created_at DESC").
		Limit(5).
		Find(&recentActivities).Error; err != nil {
		middleware.InternalServerError(c, "Failed to get recent activities")
		return
	}

	// 응답 구성
	response := map[string]interface{}{
		"activity_counts":   activityCounts,
		"recent_activities": recentActivities,
		"summary_period":    "최근 30일",
	}

	middleware.Success(c, response, "Activity summary retrieved successfully")
}
