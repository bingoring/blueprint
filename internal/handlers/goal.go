package handlers

import (
	"blueprint/internal/database"
	"blueprint/internal/middleware"
	"blueprint/internal/models"
	"blueprint/internal/services"
	"encoding/json"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ProjectHandler struct{
	aiService services.AIServiceInterface
}

func NewProjectHandler(aiService services.AIServiceInterface) *ProjectHandler {
	return &ProjectHandler{
		aiService: aiService,
	}
}

// CreateProject 프로젝트 생성
func (h *ProjectHandler) CreateProject(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		middleware.Unauthorized(c, "User not authenticated")
		return
	}

	var req models.CreateProjectRequest
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

	// Project 생성
	project := models.Project{
		UserID:      userID.(uint),
		Title:       req.Title,
		Description: req.Description,
		Category:    req.Category,
		Status:      models.ProjectDraft, // 기본값: 초안
		TargetDate:  req.TargetDate,
		Budget:      req.Budget,
		Priority:    req.Priority,
		IsPublic:    req.IsPublic,
		Tags:        tagsJSON,
		Metrics:     req.Metrics,
	}

	if err := database.GetDB().Create(&project).Error; err != nil {
		middleware.InternalServerError(c, "Failed to create project")
		return
	}

	middleware.SuccessWithStatus(c, 201, project, "Project created successfully")
}

// CreateProjectWithMilestones 프로젝트와 마일스톤을 함께 생성 ✨
func (h *ProjectHandler) CreateProjectWithMilestones(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		middleware.Unauthorized(c, "User not authenticated")
		return
	}

	var req models.CreateProjectWithMilestonesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.BadRequest(c, err.Error())
		return
	}

	// 마일스톤 검증 (최대 5개)
	if len(req.Milestones) > 5 {
		middleware.BadRequest(c, "최대 5개의 마일스톤만 설정할 수 있습니다")
		return
	}

	// 트랜잭션으로 처리
	tx := database.GetDB().Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Tags JSON 변환
	tagsJSON := ""
	if len(req.Tags) > 0 {
		if tagsBytes, err := json.Marshal(req.Tags); err == nil {
			tagsJSON = string(tagsBytes)
		}
	}

	// 프로젝트 생성
	project := models.Project{
		UserID:      userID.(uint),
		Title:       req.Title,
		Description: req.Description,
		Category:    req.Category,
		Status:      models.ProjectDraft,
		TargetDate:  req.TargetDate,
		Budget:      req.Budget,
		Priority:    req.Priority,
		IsPublic:    req.IsPublic,
		Tags:        tagsJSON,
		Metrics:     req.Metrics,
	}

	if err := tx.Create(&project).Error; err != nil {
		tx.Rollback()
		middleware.InternalServerError(c, "프로젝트 생성에 실패했습니다")
		return
	}

	// 마일스톤들 생성
	var milestones []models.Milestone
	for _, milestoneReq := range req.Milestones {
		milestone := models.Milestone{
			ProjectID:   &project.ID,
			Title:       milestoneReq.Title,
			Description: milestoneReq.Description,
			Order:       milestoneReq.Order,
			TargetDate:  milestoneReq.TargetDate,
			Status:      string(models.MilestoneStatusPending),
		}

		if err := tx.Create(&milestone).Error; err != nil {
			tx.Rollback()
			middleware.InternalServerError(c, "마일스톤 생성에 실패했습니다")
			return
		}

		milestones = append(milestones, milestone)
	}

	// 트랜잭션 커밋
	if err := tx.Commit().Error; err != nil {
		middleware.InternalServerError(c, "데이터 저장에 실패했습니다")
		return
	}

	// 생성된 프로젝트와 마일스톤들을 함께 반환
	project.Milestones = milestones

	middleware.SuccessWithStatus(c, 201, project, "프로젝트와 마일스톤이 성공적으로 등록되었습니다! ✨")
}

// GetProjects 목표 목록 조회 (카테고리별 필터링, 페이지네이션 지원)
func (h *ProjectHandler) GetProjects(c *gin.Context) {
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

	var projects []models.Project
	var total int64

	// 총 개수 조회
	query.Model(&models.Project{}).Count(&total)

	// 데이터 조회
	err := query.
		Order(sortBy + " " + order).
		Offset(offset).
		Limit(limit).
		Find(&projects).Error

	if err != nil {
		middleware.InternalServerError(c, "Failed to fetch projects")
		return
	}

	result := gin.H{
		"projects": projects,
		"pagination": gin.H{
			"page":       page,
			"limit":      limit,
			"total":      total,
			"total_pages": (total + int64(limit) - 1) / int64(limit),
		},
	}

	middleware.Success(c, result, "Projects retrieved successfully")
}

// GetProject 단일 목표 조회
func (h *ProjectHandler) GetProject(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		middleware.Unauthorized(c, "User not authenticated")
		return
	}

	projectID := c.Param("id")
	if projectID == "" {
		middleware.BadRequest(c, "Project ID is required")
		return
	}

	var project models.Project
	err := database.GetDB().
		Where("id = ? AND user_id = ?", projectID, userID).
		Preload("Paths").      // 관련 경로도 함께 로드
		Preload("Milestones"). // 마일스톤들도 함께 로드
		First(&project).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			middleware.NotFound(c, "Project not found")
			return
		}
		middleware.InternalServerError(c, "Failed to fetch project")
		return
	}

	middleware.Success(c, project, "Project retrieved successfully")
}

// UpdateProject 목표 수정
func (h *ProjectHandler) UpdateProject(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		middleware.Unauthorized(c, "User not authenticated")
		return
	}

	projectID := c.Param("id")
	if projectID == "" {
		middleware.BadRequest(c, "Project ID is required")
		return
	}

	var req models.UpdateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.BadRequest(c, err.Error())
		return
	}

	// 기존 목표 조회
	var project models.Project
	err := database.GetDB().
		Where("id = ? AND user_id = ?", projectID, userID).
		First(&project).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			middleware.NotFound(c, "Project not found")
			return
		}
		middleware.InternalServerError(c, "Failed to fetch project")
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
	if err := database.GetDB().Model(&project).Updates(updates).Error; err != nil {
		middleware.InternalServerError(c, "Failed to update project")
		return
	}

	// 업데이트된 목표 다시 조회
	database.GetDB().Where("id = ?", projectID).First(&project)

	middleware.Success(c, project, "Project updated successfully")
}

// DeleteProject 목표 삭제 (소프트 삭제)
func (h *ProjectHandler) DeleteProject(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		middleware.Unauthorized(c, "User not authenticated")
		return
	}

	projectID := c.Param("id")
	if projectID == "" {
		middleware.BadRequest(c, "Project ID is required")
		return
	}

	// 목표 존재 확인
	var project models.Project
	err := database.GetDB().
		Where("id = ? AND user_id = ?", projectID, userID).
		First(&project).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			middleware.NotFound(c, "Project not found")
			return
		}
		middleware.InternalServerError(c, "Failed to fetch project")
		return
	}

	// 소프트 삭제
	if err := database.GetDB().Delete(&project).Error; err != nil {
		middleware.InternalServerError(c, "Failed to delete project")
		return
	}

	middleware.Success(c, nil, "Project deleted successfully")
}

// UpdateProjectStatus 목표 상태 변경
func (h *ProjectHandler) UpdateProjectStatus(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		middleware.Unauthorized(c, "User not authenticated")
		return
	}

	projectID := c.Param("id")
	if projectID == "" {
		middleware.BadRequest(c, "Project ID is required")
		return
	}

	var req struct {
		Status models.ProjectStatus `json:"status" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.BadRequest(c, err.Error())
		return
	}

	// 목표 존재 확인 및 상태 업데이트
	result := database.GetDB().
		Model(&models.Project{}).
		Where("id = ? AND user_id = ?", projectID, userID).
		Update("status", req.Status)

	if result.Error != nil {
		middleware.InternalServerError(c, "Failed to update project status")
		return
	}

	if result.RowsAffected == 0 {
		middleware.NotFound(c, "Project not found")
		return
	}

	middleware.Success(c, gin.H{"status": req.Status}, "Project status updated successfully")
}

// GetProjectCategories 꿈 카테고리 목록 조회 ✨
func (h *ProjectHandler) GetProjectCategories(c *gin.Context) {
	categories := []gin.H{
		{"value": "career", "label": "💼 커리어 성장", "icon": "💼", "description": "새로운 직장, 승진, 전직의 꿈"},
		{"value": "business", "label": "🚀 창업 도전", "icon": "🚀", "description": "사업 시작, 회사 확장의 꿈"},
		{"value": "education", "label": "📚 배움의 여정", "icon": "📚", "description": "새로운 지식, 자격증, 학위의 꿈"},
		{"value": "personal", "label": "🌱 자기계발", "icon": "🌱", "description": "취미, 건강, 인간관계의 꿈"},
		{"value": "life", "label": "🏡 인생 전환", "icon": "🏡", "description": "이민, 이사, 라이프스타일의 꿈"},
	}

	middleware.Success(c, categories, "꿈 카테고리를 성공적으로 가져왔습니다")
}

// GetProjectStatuses 꿈 상태 목록 조회 ✨
func (h *ProjectHandler) GetProjectStatuses(c *gin.Context) {
	statuses := []gin.H{
		{"value": "draft", "label": "💭 구상 중", "color": "gray", "description": "아직 꿈을 다듬고 있어요"},
		{"value": "active", "label": "🔥 도전 중", "color": "blue", "description": "꿈을 향해 달려가고 있어요"},
		{"value": "completed", "label": "🎉 꿈 달성", "color": "green", "description": "축하합니다! 꿈을 이루었어요"},
		{"value": "cancelled", "label": "😔 포기", "color": "red", "description": "다른 꿈을 찾아보세요"},
		{"value": "on_hold", "label": "⏸️ 잠시 휴식", "color": "yellow", "description": "언젠가 다시 시작할 거예요"},
	}

	middleware.Success(c, statuses, "꿈 상태를 성공적으로 가져왔습니다")
}

// GenerateAIMilestones AI를 사용해서 마일스톤을 제안합니다 🤖
func (h *ProjectHandler) GenerateAIMilestones(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		middleware.Unauthorized(c, "User not authenticated")
		return
	}

	var req models.CreateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.BadRequest(c, err.Error())
		return
	}

	// 필수 필드 검증
	if req.Title == "" {
		middleware.BadRequest(c, "프로젝트 제목이 필요합니다")
		return
	}

	// AI 사용 횟수 제한 체크 🚫
	canUse, remaining, err := h.aiService.CheckAIUsageLimit(userID.(uint))
	if err != nil {
		middleware.InternalServerError(c, "사용자 정보 확인에 실패했습니다")
		return
	}

	if !canUse {
		middleware.BadRequest(c, "AI 사용 횟수를 초과했습니다 (최대 5회)")
		return
	}

	// AI 마일스톤 생성
	aiResponse, err := h.aiService.GenerateMilestones(req)
	if err != nil {
		middleware.InternalServerError(c, "AI 마일스톤 생성에 실패했습니다: "+err.Error())
		return
	}

	// 사용 횟수 업데이트 📈
	if err := h.aiService.IncrementAIUsage(userID.(uint)); err != nil {
		// 로그만 남기고 응답은 정상적으로 반환 (이미 AI 호출은 성공)
		middleware.InternalServerError(c, "AI 사용 횟수 업데이트에 실패했습니다")
		return
	}

	middleware.Success(c, gin.H{
		"milestones": aiResponse.Milestones,
		"tips":       aiResponse.Tips,
		"warnings":   aiResponse.Warnings,
		"usage": gin.H{
			"remaining": remaining - 1, // 방금 사용했으므로 -1
			"total":     5,
		},
		"meta": gin.H{
			"model":        "GPT-4o-mini",
			"generated_at": "now",
			"user_id":      userID,
		},
	}, "🤖 AI 마일스톤 제안이 완성되었습니다!")
}

// GetAIUsageInfo 사용자의 AI 사용 정보를 반환합니다 📊
func (h *ProjectHandler) GetAIUsageInfo(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		middleware.Unauthorized(c, "User not authenticated")
		return
	}

	usageInfo, err := h.aiService.GetAIUsageInfo(userID.(uint))
	if err != nil {
		middleware.InternalServerError(c, "AI 사용 정보 조회에 실패했습니다")
		return
	}

	middleware.Success(c, usageInfo, "AI 사용 정보를 성공적으로 가져왔습니다")
}
