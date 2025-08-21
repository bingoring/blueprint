package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"blueprint-module/pkg/models"
	"blueprint/internal/middleware"
	"blueprint/internal/services"

	"github.com/gin-gonic/gin"
)

// MentorStakingHandler 멘토 스테이킹 핸들러
type MentorStakingHandler struct {
	mentorStakingService *services.MentorStakingService
}

// NewMentorStakingHandler 생성자
func NewMentorStakingHandler(mentorStakingService *services.MentorStakingService) *MentorStakingHandler {
	return &MentorStakingHandler{
		mentorStakingService: mentorStakingService,
	}
}

// StakeMentor 멘토 스테이킹
// POST /api/v1/mentors/:id/stake
func (h *MentorStakingHandler) StakeMentor(c *gin.Context) {
	// 1. 멘토 ID 파라미터 추출
	mentorIDStr := c.Param("id")
	mentorID, err := strconv.ParseUint(mentorIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "잘못된 멘토 ID입니다"})
		return
	}

	// 2. 요청 바디 파싱
	var req models.StakeMentorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "잘못된 요청 데이터입니다: " + err.Error()})
		return
	}

	// 멘토 ID 설정
	req.MentorID = uint(mentorID)

	// 3. 사용자 ID 추출
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "로그인이 필요합니다"})
		return
	}

	// 4. 스테이킹 처리
	stake, err := h.mentorStakingService.StakeMentor(&req, userID.(uint))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 5. 성공 응답
	c.JSON(http.StatusCreated, gin.H{
		"message": "멘토 스테이킹이 성공적으로 완료되었습니다",
		"stake":   stake,
	})
}

// UnstakeMentor 멘토 스테이킹 해제
// POST /api/v1/stakes/:id/unstake
func (h *MentorStakingHandler) UnstakeMentor(c *gin.Context) {
	// 1. 스테이킹 ID 파라미터 추출
	stakeIDStr := c.Param("id")
	stakeID, err := strconv.ParseUint(stakeIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "잘못된 스테이킹 ID입니다"})
		return
	}

	// 2. 사용자 ID 추출
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "로그인이 필요합니다"})
		return
	}

	// 3. 스테이킹 해제 처리
	if err := h.mentorStakingService.UnstakeMentor(uint(stakeID), userID.(uint)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 4. 성공 응답
	middleware.Success(c, nil, "스테이킹 해제가 성공적으로 처리되었습니다")
}

// ReportMentor 멘토 신고
// POST /api/v1/mentors/:id/report
func (h *MentorStakingHandler) ReportMentor(c *gin.Context) {
	// 1. 멘토 ID 파라미터 추출
	mentorIDStr := c.Param("id")
	mentorID, err := strconv.ParseUint(mentorIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "잘못된 멘토 ID입니다"})
		return
	}

	// 2. 요청 바디 파싱
	var req models.ReportMentorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "잘못된 요청 데이터입니다: " + err.Error()})
		return
	}

	// 멘토 ID 설정
	req.MentorID = uint(mentorID)

	// 3. 사용자 ID 추출
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "로그인이 필요합니다"})
		return
	}

	// 4. 멘토 신고 처리
	slashEvent, err := h.mentorStakingService.ReportMentor(&req, userID.(uint))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 5. 성공 응답
	c.JSON(http.StatusCreated, gin.H{
		"message":     "멘토 신고가 성공적으로 제출되었습니다",
		"slash_event": slashEvent,
	})
}

// GetMyStakes 내 스테이킹 목록 조회
// GET /api/v1/stakes/my
func (h *MentorStakingHandler) GetMyStakes(c *gin.Context) {
	// 1. 사용자 ID 추출
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "로그인이 필요합니다"})
		return
	}

	// 2. 쿼리 파라미터 추출
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	status := c.Query("status")
	stakeType := c.Query("stake_type")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	// 3. 내 스테이킹 목록 조회
	response, err := h.mentorStakingService.GetUserStakes(userID.(uint), page, limit, status, stakeType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 4. 성공 응답
	c.JSON(http.StatusOK, response)
}

// GetMentorStakes 특정 멘토의 스테이킹 정보 조회
// GET /api/v1/mentors/:id/stakes
func (h *MentorStakingHandler) GetMentorStakes(c *gin.Context) {
	// 1. 멘토 ID 파라미터 추출
	mentorIDStr := c.Param("id")
	mentorID, err := strconv.ParseUint(mentorIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "잘못된 멘토 ID입니다"})
		return
	}

	// 2. 멘토 스테이킹 정보 조회
	response, err := h.mentorStakingService.GetMentorStakeInfo(uint(mentorID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// 3. 성공 응답
	c.JSON(http.StatusOK, response)
}

// GetMentorPerformance 멘토 성과 지표 조회
// GET /api/v1/mentors/:id/performance
func (h *MentorStakingHandler) GetMentorPerformance(c *gin.Context) {
	// 1. 멘토 ID 파라미터 추출
	mentorIDStr := c.Param("id")
	mentorID, err := strconv.ParseUint(mentorIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "잘못된 멘토 ID입니다"})
		return
	}

	// 2. 쿼리 파라미터 추출
	period := c.DefaultQuery("period", "monthly") // weekly, monthly, quarterly, yearly

	var periodType models.MetricPeriodType
	switch period {
	case "weekly":
		periodType = models.MetricPeriodWeekly
	case "monthly":
		periodType = models.MetricPeriodMonthly
	case "quarterly":
		periodType = models.MetricPeriodQuarterly
	case "yearly":
		periodType = models.MetricPeriodYearly
	default:
		periodType = models.MetricPeriodMonthly
	}

	// 3. 성과 지표 계산 및 조회
	metric, err := h.mentorStakingService.CalculatePerformanceMetrics(uint(mentorID), periodType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 4. 성공 응답
	middleware.Success(c, gin.H{
		"performance_metric": metric,
	}, "")
}

// GetMentorDashboard 멘토 대시보드 조회
// GET /api/v1/mentors/my/dashboard
func (h *MentorStakingHandler) GetMentorDashboard(c *gin.Context) {
	// 1. 사용자 ID 추출
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "로그인이 필요합니다"})
		return
	}

	// 2. 멘토 정보 확인
	mentor, err := h.mentorStakingService.GetMentorByUserID(userID.(uint))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "멘토 정보를 찾을 수 없습니다"})
		return
	}

	// 3. 대시보드 정보 조회
	response, err := h.mentorStakingService.GetMentorDashboard(mentor.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 4. 성공 응답
	c.JSON(http.StatusOK, response)
}

// GetSlashEvents 슬래싱 이벤트 목록 조회
// GET /api/v1/mentors/:id/slash-events
func (h *MentorStakingHandler) GetSlashEvents(c *gin.Context) {
	// 1. 멘토 ID 파라미터 추출
	mentorIDStr := c.Param("id")
	mentorID, err := strconv.ParseUint(mentorIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "잘못된 멘토 ID입니다"})
		return
	}

	// 2. 쿼리 파라미터 추출
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	status := c.Query("status")
	slashType := c.Query("slash_type")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	// 3. 슬래싱 이벤트 목록 조회
	response, err := h.mentorStakingService.GetMentorSlashEvents(uint(mentorID), page, limit, status, slashType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 4. 성공 응답
	c.JSON(http.StatusOK, response)
}

// ProcessSlashEvent 슬래싱 이벤트 처리 (관리자용)
// POST /api/v1/slash-events/:id/process
func (h *MentorStakingHandler) ProcessSlashEvent(c *gin.Context) {
	// 1. 슬래싱 이벤트 ID 파라미터 추출
	eventIDStr := c.Param("id")
	eventID, err := strconv.ParseUint(eventIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "잘못된 이벤트 ID입니다"})
		return
	}

	// 2. 요청 바디 파싱
	var req struct {
		Approved bool   `json:"approved" binding:"required"`
		Comment  string `json:"comment"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "잘못된 요청 데이터입니다: " + err.Error()})
		return
	}

	// 3. 사용자 ID 추출 (검토자)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "로그인이 필요합니다"})
		return
	}

	// 4. 슬래싱 이벤트 처리
	if err := h.mentorStakingService.ProcessSlashing(uint(eventID), userID.(uint), req.Approved, req.Comment); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 5. 성공 응답
	status := "거부"
	if req.Approved {
		status = "승인"
	}

	middleware.Success(c, nil, fmt.Sprintf("슬래싱 이벤트가 %s되었습니다", status))
}

// GetStakingStats 스테이킹 통계 조회
// GET /api/v1/staking/stats
func (h *MentorStakingHandler) GetStakingStats(c *gin.Context) {
	// 1. 쿼리 파라미터 추출
	period := c.DefaultQuery("period", "monthly")
	mentorID := c.Query("mentor_id")

	// 2. 통계 정보 조회
	stats, err := h.mentorStakingService.GetStakingStats(period, mentorID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 3. 성공 응답
	c.JSON(http.StatusOK, stats)
}

// GetTopMentors 상위 멘토 목록 조회 (스테이킹 기준)
// GET /api/v1/mentors/top
func (h *MentorStakingHandler) GetTopMentors(c *gin.Context) {
	// 1. 쿼리 파라미터 추출
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	sortBy := c.DefaultQuery("sort_by", "total_staked") // total_staked, performance_score, success_rate
	category := c.Query("category") // 전문 분야별 필터링

	if limit < 1 || limit > 100 {
		limit = 20
	}

	// 2. 상위 멘토 목록 조회
	mentors, err := h.mentorStakingService.GetTopMentors(limit, sortBy, category)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 3. 성공 응답
	middleware.Success(c, gin.H{
		"top_mentors": mentors,
		"criteria":    sortBy,
		"category":    category,
	}, "")
}
