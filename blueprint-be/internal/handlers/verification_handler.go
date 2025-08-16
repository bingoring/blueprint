package handlers

import (
	"net/http"
	"strconv"

	"blueprint-module/pkg/models"
	"blueprint/internal/services"

	"github.com/gin-gonic/gin"
)

// VerificationHandler 검증 관련 핸들러
type VerificationHandler struct {
	verificationService *services.VerificationService
}

// NewVerificationHandler 생성자
func NewVerificationHandler(verificationService *services.VerificationService) *VerificationHandler {
	return &VerificationHandler{
		verificationService: verificationService,
	}
}

// SubmitProof 증거 제출
// POST /api/v1/milestones/:id/proof
func (h *VerificationHandler) SubmitProof(c *gin.Context) {
	// 1. 마일스톤 ID 파라미터 추출
	milestoneIDStr := c.Param("id")
	milestoneID, err := strconv.ParseUint(milestoneIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "잘못된 마일스톤 ID입니다"})
		return
	}

	// 2. 요청 바디 파싱
	var req models.SubmitProofRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "잘못된 요청 데이터입니다: " + err.Error()})
		return
	}

	// 마일스톤 ID 설정 (URL에서 가져온 값으로 덮어쓰기)
	req.MilestoneID = uint(milestoneID)

	// 3. 사용자 ID 추출 (JWT 미들웨어에서 설정)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "로그인이 필요합니다"})
		return
	}

	// 4. 증거 제출 처리
	proof, err := h.verificationService.SubmitProof(&req, userID.(uint))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 5. 성공 응답
	c.JSON(http.StatusCreated, gin.H{
		"message": "증거가 성공적으로 제출되었습니다",
		"proof":   proof,
	})
}

// ValidateProof 증거 검증 (검증인 투표)
// POST /api/v1/proofs/:id/validate
func (h *VerificationHandler) ValidateProof(c *gin.Context) {
	// 1. 증거 ID 파라미터 추출
	proofIDStr := c.Param("id")
	proofID, err := strconv.ParseUint(proofIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "잘못된 증거 ID입니다"})
		return
	}

	// 2. 요청 바디 파싱
	var req models.ValidateProofRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "잘못된 요청 데이터입니다: " + err.Error()})
		return
	}

	// 증거 ID 설정
	req.ProofID = uint(proofID)

	// 3. 사용자 ID 추출
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "로그인이 필요합니다"})
		return
	}

	// 4. 검증 처리
	validator, err := h.verificationService.ValidateProof(&req, userID.(uint))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 5. 성공 응답
	c.JSON(http.StatusCreated, gin.H{
		"message":   "검증 투표가 성공적으로 제출되었습니다",
		"validator": validator,
	})
}

// DisputeProof 증거 분쟁 제기
// POST /api/v1/proofs/:id/dispute
func (h *VerificationHandler) DisputeProof(c *gin.Context) {
	// 1. 증거 ID 파라미터 추출
	proofIDStr := c.Param("id")
	proofID, err := strconv.ParseUint(proofIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "잘못된 증거 ID입니다"})
		return
	}

	// 2. 요청 바디 파싱
	var req models.DisputeProofRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "잘못된 요청 데이터입니다: " + err.Error()})
		return
	}

	// 증거 ID 설정
	req.ProofID = uint(proofID)

	// 3. 사용자 ID 추출
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "로그인이 필요합니다"})
		return
	}

	// 4. 분쟁 제기 처리
	dispute, err := h.verificationService.DisputeProof(&req, userID.(uint))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 5. 성공 응답
	c.JSON(http.StatusCreated, gin.H{
		"message": "분쟁이 성공적으로 제기되었습니다",
		"dispute": dispute,
	})
}

// GetProofVerification 증거 검증 정보 조회
// GET /api/v1/proofs/:id/verification
func (h *VerificationHandler) GetProofVerification(c *gin.Context) {
	// 1. 증거 ID 파라미터 추출
	proofIDStr := c.Param("id")
	proofID, err := strconv.ParseUint(proofIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "잘못된 증거 ID입니다"})
		return
	}

	// 2. 사용자 ID 추출 (선택적)
	var userID uint
	if userIDVal, exists := c.Get("user_id"); exists {
		userID = userIDVal.(uint)
	}

	// 3. 검증 정보 조회
	response, err := h.verificationService.GetProofVerification(uint(proofID), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// 4. 성공 응답
	c.JSON(http.StatusOK, response)
}

// GetValidatorDashboard 검증인 대시보드 정보 조회
// GET /api/v1/verification/dashboard
func (h *VerificationHandler) GetValidatorDashboard(c *gin.Context) {
	// 1. 사용자 ID 추출
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "로그인이 필요합니다"})
		return
	}

	// 2. 대시보드 정보 조회
	response, err := h.verificationService.GetValidatorDashboard(userID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 3. 성공 응답
	c.JSON(http.StatusOK, response)
}

// GetPendingProofs 검증 대기 중인 증거 목록 조회
// GET /api/v1/verification/pending
func (h *VerificationHandler) GetPendingProofs(c *gin.Context) {
	// 1. 쿼리 파라미터 추출
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	
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

	// 3. 대기 중인 증거 목록 조회 (간단한 구현)
	response, err := h.verificationService.GetValidatorDashboard(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 4. 페이징 처리 (간단한 구현)
	offset := (page - 1) * limit
	proofs := response.PendingProofs
	
	var paginatedProofs []models.MilestoneProof
	if offset < len(proofs) {
		end := offset + limit
		if end > len(proofs) {
			end = len(proofs)
		}
		paginatedProofs = proofs[offset:end]
	}

	// 5. 성공 응답
	c.JSON(http.StatusOK, gin.H{
		"proofs": paginatedProofs,
		"pagination": gin.H{
			"page":        page,
			"limit":       limit,
			"total":       len(proofs),
			"total_pages": (len(proofs) + limit - 1) / limit,
		},
	})
}

// UploadProofFile 증거 파일 업로드
// POST /api/v1/verification/upload
func (h *VerificationHandler) UploadProofFile(c *gin.Context) {
	// 1. 사용자 인증 확인
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "로그인이 필요합니다"})
		return
	}

	// 2. 파일 추출
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "파일이 제공되지 않았습니다"})
		return
	}
	defer file.Close()

	// 3. 파일 크기 제한 (10MB)
	if header.Size > 10*1024*1024 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "파일 크기는 10MB를 초과할 수 없습니다"})
		return
	}

	// 4. 파일 타입 확인 (기본적인 확장자 검사)
	allowedExtensions := map[string]bool{
		".jpg":  true, ".jpeg": true, ".png": true, ".gif": true,
		".pdf":  true, ".doc": true, ".docx": true, ".txt": true,
		".mp4":  true, ".mov": true, ".avi": true,
		".zip":  true, ".rar": true,
	}

	ext := ""
	for i := len(header.Filename) - 1; i >= 0; i-- {
		if header.Filename[i] == '.' {
			ext = header.Filename[i:]
			break
		}
	}

	if !allowedExtensions[ext] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "지원하지 않는 파일 형식입니다"})
		return
	}

	// 5. 파일 업로드 (VerificationService를 통한 FileService 사용)
	fileURL, err := h.verificationService.UploadFile(file, header, "proofs")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "파일 업로드 실패: " + err.Error()})
		return
	}

	// 6. 성공 응답
	c.JSON(http.StatusOK, gin.H{
		"message":  "파일이 성공적으로 업로드되었습니다",
		"file_url": fileURL,
		"user_id":  userID,
	})
}

// GetMilestoneProofs 마일스톤의 증거 목록 조회
// GET /api/v1/milestones/:id/proofs
func (h *VerificationHandler) GetMilestoneProofs(c *gin.Context) {
	// 1. 마일스톤 ID 파라미터 추출
	milestoneIDStr := c.Param("id")
	milestoneID, err := strconv.ParseUint(milestoneIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "잘못된 마일스톤 ID입니다"})
		return
	}

	// 2. 증거 목록 조회 (간단한 구현)
	// TODO: VerificationService에 GetMilestoneProofs 메서드 추가
	c.JSON(http.StatusOK, gin.H{
		"milestone_id": milestoneID,
		"proofs":       []models.MilestoneProof{},
		"message":      "증거 목록 조회 기능이 구현 중입니다",
	})
}

// GetVerificationStats 검증 통계 조회
// GET /api/v1/verification/stats
func (h *VerificationHandler) GetVerificationStats(c *gin.Context) {
	// 1. 사용자 ID 추출 (선택적)
	var userID uint
	if userIDVal, exists := c.Get("user_id"); exists {
		userID = userIDVal.(uint)
	}

	// 2. 통계 정보 조회
	response, err := h.verificationService.GetValidatorDashboard(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 3. 성공 응답
	c.JSON(http.StatusOK, gin.H{
		"statistics": response.Statistics,
		"qualification": gin.H{
			"reputation_score": response.Qualification.ReputationScore,
			"staked_amount":    response.Qualification.StakedAmount,
			"total_verifications": response.Qualification.TotalVerifications,
			"accuracy_rate":    response.Qualification.AccuracyRate,
		},
	})
}