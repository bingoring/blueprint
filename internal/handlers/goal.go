package handlers

import (
	"blueprint/internal/database"
	"blueprint/internal/middleware"
	"blueprint/internal/models"
	"encoding/json"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type GoalHandler struct{}

func NewGoalHandler() *GoalHandler {
	return &GoalHandler{}
}

// CreateGoal 목표 생성
func (h *GoalHandler) CreateGoal(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		middleware.Unauthorized(c, "User not authenticated")
		return
	}

	var req models.CreateGoalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.BadRequest(c, err.Error())
		return
	}

	// Tags JSON 변환
	tagsJSON := ""
	if len(req.Tags) > 0 {
		if tagsBytes, err := json.Marshal(req.Tags); err == nil {
			tagsJSON = string(tagsBytes)
		}
	}

	// Goal 생성
	goal := models.Goal{
		UserID:      userID.(uint),
		Title:       req.Title,
		Description: req.Description,
		Category:    req.Category,
		Status:      models.GoalDraft, // 기본값: 초안
		TargetDate:  req.TargetDate,
		Budget:      req.Budget,
		Priority:    req.Priority,
		IsPublic:    req.IsPublic,
		Tags:        tagsJSON,
		Metrics:     req.Metrics,
	}

	if err := database.GetDB().Create(&goal).Error; err != nil {
		middleware.InternalServerError(c, "Failed to create goal")
		return
	}

	middleware.SuccessWithStatus(c, 201, goal, "Goal created successfully")
}

// GetGoals 목표 목록 조회 (카테고리별 필터링, 페이지네이션 지원)
func (h *GoalHandler) GetGoals(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		middleware.Unauthorized(c, "User not authenticated")
		return
	}

	// 쿼리 파라미터 파싱
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	category := c.Query("category")
	status := c.Query("status")
	sortBy := c.DefaultQuery("sort", "created_at")
	order := c.DefaultQuery("order", "desc")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	offset := (page - 1) * limit

	// 쿼리 빌드
	query := database.GetDB().Where("user_id = ?", userID)

	if category != "" {
		query = query.Where("category = ?", category)
	}

	if status != "" {
		query = query.Where("status = ?", status)
	}

	// 정렬
	validSorts := map[string]bool{
		"created_at": true, "updated_at": true, "priority": true, "target_date": true,
	}
	if !validSorts[sortBy] {
		sortBy = "created_at"
	}
	if order != "asc" && order != "desc" {
		order = "desc"
	}

	var goals []models.Goal
	var total int64

	// 총 개수 조회
	query.Model(&models.Goal{}).Count(&total)

	// 데이터 조회
	err := query.
		Order(sortBy + " " + order).
		Offset(offset).
		Limit(limit).
		Find(&goals).Error

	if err != nil {
		middleware.InternalServerError(c, "Failed to fetch goals")
		return
	}

	result := gin.H{
		"goals": goals,
		"pagination": gin.H{
			"page":       page,
			"limit":      limit,
			"total":      total,
			"total_pages": (total + int64(limit) - 1) / int64(limit),
		},
	}

	middleware.Success(c, result, "Goals retrieved successfully")
}

// GetGoal 단일 목표 조회
func (h *GoalHandler) GetGoal(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		middleware.Unauthorized(c, "User not authenticated")
		return
	}

	goalID := c.Param("id")
	if goalID == "" {
		middleware.BadRequest(c, "Goal ID is required")
		return
	}

	var goal models.Goal
	err := database.GetDB().
		Where("id = ? AND user_id = ?", goalID, userID).
		Preload("Paths"). // 관련 경로도 함께 로드
		First(&goal).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			middleware.NotFound(c, "Goal not found")
			return
		}
		middleware.InternalServerError(c, "Failed to fetch goal")
		return
	}

	middleware.Success(c, goal, "Goal retrieved successfully")
}

// UpdateGoal 목표 수정
func (h *GoalHandler) UpdateGoal(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		middleware.Unauthorized(c, "User not authenticated")
		return
	}

	goalID := c.Param("id")
	if goalID == "" {
		middleware.BadRequest(c, "Goal ID is required")
		return
	}

	var req models.UpdateGoalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.BadRequest(c, err.Error())
		return
	}

	// 기존 목표 조회
	var goal models.Goal
	err := database.GetDB().
		Where("id = ? AND user_id = ?", goalID, userID).
		First(&goal).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			middleware.NotFound(c, "Goal not found")
			return
		}
		middleware.InternalServerError(c, "Failed to fetch goal")
		return
	}

	// 업데이트할 필드들
	updates := map[string]interface{}{}

	if req.Title != "" {
		updates["title"] = req.Title
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.Category != "" {
		updates["category"] = req.Category
	}
	if req.Status != "" {
		updates["status"] = req.Status
	}
	if req.TargetDate != nil {
		updates["target_date"] = req.TargetDate
	}
	if req.Budget > 0 {
		updates["budget"] = req.Budget
	}
	if req.Priority > 0 {
		updates["priority"] = req.Priority
	}
	updates["is_public"] = req.IsPublic

	// Tags 처리
	if len(req.Tags) > 0 {
		if tagsBytes, err := json.Marshal(req.Tags); err == nil {
			updates["tags"] = string(tagsBytes)
		}
	}

	if req.Metrics != "" {
		updates["metrics"] = req.Metrics
	}

	// 업데이트 실행
	if err := database.GetDB().Model(&goal).Updates(updates).Error; err != nil {
		middleware.InternalServerError(c, "Failed to update goal")
		return
	}

	// 업데이트된 목표 다시 조회
	database.GetDB().Where("id = ?", goalID).First(&goal)

	middleware.Success(c, goal, "Goal updated successfully")
}

// DeleteGoal 목표 삭제 (소프트 삭제)
func (h *GoalHandler) DeleteGoal(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		middleware.Unauthorized(c, "User not authenticated")
		return
	}

	goalID := c.Param("id")
	if goalID == "" {
		middleware.BadRequest(c, "Goal ID is required")
		return
	}

	// 목표 존재 확인
	var goal models.Goal
	err := database.GetDB().
		Where("id = ? AND user_id = ?", goalID, userID).
		First(&goal).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			middleware.NotFound(c, "Goal not found")
			return
		}
		middleware.InternalServerError(c, "Failed to fetch goal")
		return
	}

	// 소프트 삭제
	if err := database.GetDB().Delete(&goal).Error; err != nil {
		middleware.InternalServerError(c, "Failed to delete goal")
		return
	}

	middleware.Success(c, nil, "Goal deleted successfully")
}

// UpdateGoalStatus 목표 상태 변경
func (h *GoalHandler) UpdateGoalStatus(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		middleware.Unauthorized(c, "User not authenticated")
		return
	}

	goalID := c.Param("id")
	if goalID == "" {
		middleware.BadRequest(c, "Goal ID is required")
		return
	}

	var req struct {
		Status models.GoalStatus `json:"status" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.BadRequest(c, err.Error())
		return
	}

	// 목표 존재 확인 및 상태 업데이트
	result := database.GetDB().
		Model(&models.Goal{}).
		Where("id = ? AND user_id = ?", goalID, userID).
		Update("status", req.Status)

	if result.Error != nil {
		middleware.InternalServerError(c, "Failed to update goal status")
		return
	}

	if result.RowsAffected == 0 {
		middleware.NotFound(c, "Goal not found")
		return
	}

	middleware.Success(c, gin.H{"status": req.Status}, "Goal status updated successfully")
}

// GetGoalCategories 목표 카테고리 목록 조회
func (h *GoalHandler) GetGoalCategories(c *gin.Context) {
	categories := []gin.H{
		{"value": "career", "label": "커리어", "icon": "💼"},
		{"value": "business", "label": "비즈니스", "icon": "🚀"},
		{"value": "education", "label": "교육", "icon": "📚"},
		{"value": "personal", "label": "개인", "icon": "🌱"},
		{"value": "life", "label": "라이프", "icon": "🏡"},
	}

	middleware.Success(c, categories, "Goal categories retrieved successfully")
}

// GetGoalStatuses 목표 상태 목록 조회
func (h *GoalHandler) GetGoalStatuses(c *gin.Context) {
	statuses := []gin.H{
		{"value": "draft", "label": "초안", "color": "gray"},
		{"value": "active", "label": "활성", "color": "blue"},
		{"value": "completed", "label": "완료", "color": "green"},
		{"value": "cancelled", "label": "취소", "color": "red"},
		{"value": "on_hold", "label": "보류", "color": "yellow"},
	}

	middleware.Success(c, statuses, "Goal statuses retrieved successfully")
}
