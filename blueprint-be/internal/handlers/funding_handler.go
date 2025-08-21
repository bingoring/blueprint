package handlers

import (
	"blueprint/internal/middleware"
	"blueprint/internal/services"
	"strconv"

	"github.com/gin-gonic/gin"
)

// FundingHandler 펀딩 관련 API 핸들러
type FundingHandler struct {
	fundingService   *services.FundingVerificationService
	lifecycleService *services.MilestoneLifecycleService
}

// NewFundingHandler 펀딩 핸들러 생성자
func NewFundingHandler(fundingService *services.FundingVerificationService, lifecycleService *services.MilestoneLifecycleService) *FundingHandler {
	return &FundingHandler{
		fundingService:   fundingService,
		lifecycleService: lifecycleService,
	}
}

// GetFundingStats 마일스톤 펀딩 통계 조회
// GET /api/v1/milestones/:id/funding/stats
func (h *FundingHandler) GetFundingStats(c *gin.Context) {
	milestoneIDStr := c.Param("id")
	milestoneID, err := strconv.ParseUint(milestoneIDStr, 10, 32)
	if err != nil {
		middleware.BadRequest(c, "Invalid milestone ID")
		return
	}

	stats, err := h.fundingService.GetFundingStats(uint(milestoneID))
	if err != nil {
		middleware.InternalServerError(c, "Failed to get funding stats")
		return
	}

	middleware.Success(c, stats, "")
}

// StartFundingPhase 펀딩 단계 강제 시작 (관리자용)
// POST /api/v1/milestones/:id/funding/start
func (h *FundingHandler) StartFundingPhase(c *gin.Context) {
	// 관리자 권한 확인 (추후 구현)
	// userID, exists := c.Get("user_id")
	// if !exists {
	//     middleware.Unauthorized(c, "User not authenticated")
	//     return
	// }

	milestoneIDStr := c.Param("id")
	milestoneID, err := strconv.ParseUint(milestoneIDStr, 10, 32)
	if err != nil {
		middleware.BadRequest(c, "Invalid milestone ID")
		return
	}

	if err := h.lifecycleService.ForceStartFunding(uint(milestoneID)); err != nil {
		middleware.InternalServerError(c, "Failed to start funding phase: "+err.Error())
		return
	}

	middleware.Success(c, nil, "Funding phase started successfully")
}

// ProcessExpiredFunding 만료된 펀딩들 강제 처리 (관리자용)
// POST /api/v1/funding/process-expired
func (h *FundingHandler) ProcessExpiredFunding(c *gin.Context) {
	if err := h.lifecycleService.ForceProcessExpired(); err != nil {
		middleware.InternalServerError(c, "Failed to process expired funding: "+err.Error())
		return
	}

	middleware.Success(c, nil, "Expired funding processed successfully")
}

// GetLifecycleStats 전체 라이프사이클 통계 조회
// GET /api/v1/funding/lifecycle-stats
func (h *FundingHandler) GetLifecycleStats(c *gin.Context) {
	stats, err := h.lifecycleService.GetLifecycleStats()
	if err != nil {
		middleware.InternalServerError(c, "Failed to get lifecycle stats")
		return
	}

	middleware.Success(c, stats, "")
}

// GetFundingMilestones 펀딩 중인 마일스톤 목록 조회
// GET /api/v1/funding/active
func (h *FundingHandler) GetFundingMilestones(c *gin.Context) {
	// 쿼리 파라미터
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	category := c.Query("category")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 50 {
		limit = 10
	}

	offset := (page - 1) * limit

	// 데이터베이스 쿼리 (기본 구현)
	// TODO: 실제 서비스 메서드로 분리
	middleware.Success(c, gin.H{
		"page":       page,
		"limit":      limit,
		"category":   category,
		"offset":     offset,
		"milestones": []gin.H{}, // 실제 데이터는 추후 구현
	}, "펀딩 중인 마일스톤 목록")
}
