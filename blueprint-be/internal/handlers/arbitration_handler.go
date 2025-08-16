package handlers

import (
	"net/http"
	"strconv"

	"blueprint-module/pkg/models"
	"blueprint/internal/services"

	"github.com/gin-gonic/gin"
)

// ArbitrationHandler 분쟁 해결 핸들러
type ArbitrationHandler struct {
	arbitrationService *services.ArbitrationService
}

// NewArbitrationHandler 생성자
func NewArbitrationHandler(arbitrationService *services.ArbitrationService) *ArbitrationHandler {
	return &ArbitrationHandler{
		arbitrationService: arbitrationService,
	}
}

// SubmitCase 분쟁 사건 제기
// POST /api/v1/arbitration/cases
func (h *ArbitrationHandler) SubmitCase(c *gin.Context) {
	// 1. 요청 바디 파싱
	var req models.SubmitArbitrationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "잘못된 요청 데이터입니다: " + err.Error()})
		return
	}

	// 2. 사용자 ID 추출
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "로그인이 필요합니다"})
		return
	}

	// 3. 분쟁 사건 제기 처리
	arbitrationCase, err := h.arbitrationService.SubmitCase(&req, userID.(uint))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 4. 성공 응답
	c.JSON(http.StatusCreated, gin.H{
		"message": "분쟁 사건이 성공적으로 제기되었습니다",
		"case":    arbitrationCase,
	})
}

// GetCase 분쟁 사건 조회
// GET /api/v1/arbitration/cases/:id
func (h *ArbitrationHandler) GetCase(c *gin.Context) {
	// 1. 사건 ID 파라미터 추출
	caseIDStr := c.Param("id")
	caseID, err := strconv.ParseUint(caseIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "잘못된 사건 ID입니다"})
		return
	}

	// 2. 사용자 ID 추출 (선택적)
	var userID uint
	if userIDVal, exists := c.Get("user_id"); exists {
		userID = userIDVal.(uint)
	}

	// 3. 사건 정보 조회
	response, err := h.arbitrationService.GetCaseDetails(uint(caseID), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// 4. 성공 응답
	c.JSON(http.StatusOK, response)
}

// CommitVote 배심원 투표 제출 (Commit phase)
// POST /api/v1/arbitration/cases/:id/vote
func (h *ArbitrationHandler) CommitVote(c *gin.Context) {
	// 1. 사건 ID 파라미터 추출
	caseIDStr := c.Param("id")
	caseID, err := strconv.ParseUint(caseIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "잘못된 사건 ID입니다"})
		return
	}

	// 2. 요청 바디 파싱
	var req models.JurorVoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "잘못된 요청 데이터입니다: " + err.Error()})
		return
	}

	// 사건 ID 설정
	req.CaseID = uint(caseID)

	// 3. 사용자 ID 추출
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "로그인이 필요합니다"})
		return
	}

	// 4. 투표 제출 처리
	vote, err := h.arbitrationService.CommitVote(&req, userID.(uint))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 5. 성공 응답
	c.JSON(http.StatusCreated, gin.H{
		"message": "투표가 성공적으로 제출되었습니다",
		"vote":    vote,
	})
}

// RevealVote 투표 공개 (Reveal phase)
// POST /api/v1/arbitration/cases/:id/reveal
func (h *ArbitrationHandler) RevealVote(c *gin.Context) {
	// 1. 사건 ID 파라미터 추출
	caseIDStr := c.Param("id")
	caseID, err := strconv.ParseUint(caseIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "잘못된 사건 ID입니다"})
		return
	}

	// 2. 요청 바디 파싱
	var req models.RevealVoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "잘못된 요청 데이터입니다: " + err.Error()})
		return
	}

	// 사건 ID 설정
	req.CaseID = uint(caseID)

	// 3. 사용자 ID 추출
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "로그인이 필요합니다"})
		return
	}

	// 4. 투표 공개 처리
	if err := h.arbitrationService.RevealVote(&req, userID.(uint)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 5. 성공 응답
	c.JSON(http.StatusOK, gin.H{
		"message": "투표가 성공적으로 공개되었습니다",
	})
}

// GetJurorDashboard 배심원 대시보드 조회
// GET /api/v1/arbitration/juror/dashboard
func (h *ArbitrationHandler) GetJurorDashboard(c *gin.Context) {
	// 1. 사용자 ID 추출
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "로그인이 필요합니다"})
		return
	}

	// 2. 대시보드 정보 조회
	response, err := h.arbitrationService.GetJurorDashboard(userID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 3. 성공 응답
	c.JSON(http.StatusOK, response)
}

// GetPendingCases 대기 중인 분쟁 사건 목록 조회
// GET /api/v1/arbitration/cases/pending
func (h *ArbitrationHandler) GetPendingCases(c *gin.Context) {
	// 1. 쿼리 파라미터 추출
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	disputeType := c.Query("dispute_type")
	priority := c.Query("priority")
	
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	// 2. 사용자 ID 추출 (선택적)
	var userID uint
	if userIDVal, exists := c.Get("user_id"); exists {
		userID = userIDVal.(uint)
	}

	// 3. 대기 중인 사건 목록 조회
	response, err := h.arbitrationService.GetPendingCases(userID, page, limit, disputeType, priority)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 4. 성공 응답
	c.JSON(http.StatusOK, response)
}

// GetMyCases 내 분쟁 사건 목록 조회
// GET /api/v1/arbitration/cases/my
func (h *ArbitrationHandler) GetMyCases(c *gin.Context) {
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
	role := c.DefaultQuery("role", "all") // "plaintiff", "defendant", "juror", "all"
	
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	// 3. 내 사건 목록 조회
	response, err := h.arbitrationService.GetUserCases(userID.(uint), page, limit, status, role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 4. 성공 응답
	c.JSON(http.StatusOK, response)
}

// BecomeJuror 배심원 등록
// POST /api/v1/arbitration/juror/register
func (h *ArbitrationHandler) BecomeJuror(c *gin.Context) {
	// 1. 요청 바디 파싱
	var req struct {
		MinStakeAmount  int64    `json:"min_stake_amount" binding:"min=5000"`   // 최소 5,000 BLUEPRINT
		ExpertiseAreas  []string `json:"expertise_areas"`
		LanguageSkills  []string `json:"language_skills"`
		LegalBackground bool     `json:"legal_background"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "잘못된 요청 데이터입니다: " + err.Error()})
		return
	}

	// 2. 사용자 ID 추출
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "로그인이 필요합니다"})
		return
	}

	// 3. 배심원 등록 처리
	qualification, err := h.arbitrationService.RegisterJuror(userID.(uint), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 4. 성공 응답
	c.JSON(http.StatusCreated, gin.H{
		"message":       "배심원으로 성공적으로 등록되었습니다",
		"qualification": qualification,
	})
}

// GetArbitrationStats 분쟁 해결 통계 조회
// GET /api/v1/arbitration/stats
func (h *ArbitrationHandler) GetArbitrationStats(c *gin.Context) {
	// 1. 쿼리 파라미터 추출
	period := c.DefaultQuery("period", "monthly") // daily, weekly, monthly, yearly
	
	// 2. 통계 정보 조회
	stats, err := h.arbitrationService.GetArbitrationStats(period)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 3. 성공 응답
	c.JSON(http.StatusOK, stats)
}

// AppealCase 판결 이의제기
// POST /api/v1/arbitration/cases/:id/appeal
func (h *ArbitrationHandler) AppealCase(c *gin.Context) {
	// 1. 사건 ID 파라미터 추출
	caseIDStr := c.Param("id")
	caseID, err := strconv.ParseUint(caseIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "잘못된 사건 ID입니다"})
		return
	}

	// 2. 요청 바디 파싱
	var req struct {
		Reason      string `json:"reason" binding:"required"`
		Evidence    string `json:"evidence"`
		StakeAmount int64  `json:"stake_amount" binding:"min=2000"` // 이의제기 스테이킹
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "잘못된 요청 데이터입니다: " + err.Error()})
		return
	}

	// 3. 사용자 ID 추출
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "로그인이 필요합니다"})
		return
	}

	// 4. 이의제기 처리
	appeal, err := h.arbitrationService.AppealCase(uint(caseID), userID.(uint), req.Reason, req.Evidence, req.StakeAmount)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 5. 성공 응답
	c.JSON(http.StatusCreated, gin.H{
		"message": "이의제기가 성공적으로 제출되었습니다",
		"appeal":  appeal,
	})
}