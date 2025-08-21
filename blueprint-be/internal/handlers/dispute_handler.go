package handlers

import (
	"fmt"
	"strconv"

	"blueprint-module/pkg/models"
	"blueprint/internal/middleware"
	"blueprint/internal/services"

	"github.com/gin-gonic/gin"
)

type DisputeHandler struct {
	disputeService *services.DisputeService
}

func NewDisputeHandler(disputeService *services.DisputeService) *DisputeHandler {
	return &DisputeHandler{
		disputeService: disputeService,
	}
}

// 🏛️ 마일스톤 결과 보고
func (dh *DisputeHandler) ReportMilestoneResult(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		middleware.Unauthorized(c, "User not authenticated")
		return
	}

	milestoneIDStr := c.Param("id")
	milestoneID, err := strconv.ParseUint(milestoneIDStr, 10, 32)
	if err != nil {
		middleware.BadRequest(c, "Invalid milestone ID")
		return
	}

	var req struct {
		Result        bool     `json:"result" binding:"required"`
		EvidenceURL   string   `json:"evidence_url"`
		EvidenceFiles []string `json:"evidence_files"`
		Description   string   `json:"description" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.BadRequest(c, "Invalid request body")
		return
	}

	// 결과 보고
	err = dh.disputeService.ReportMilestoneResult(
		uint(milestoneID),
		userID.(uint),
		req.Result,
		req.EvidenceURL,
		req.EvidenceFiles,
		req.Description,
	)

	if err != nil {
		middleware.InternalServerError(c, err.Error())
		return
	}

	middleware.Success(c, gin.H{
		"milestone_id": milestoneID,
		"result":       req.Result,
		"challenge_window_hours": 48,
	}, "Milestone result reported successfully. Challenge window is now open for 48 hours.")
}

// ⚔️ 이의 제기
func (dh *DisputeHandler) CreateDispute(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		middleware.Unauthorized(c, "User not authenticated")
		return
	}

	var req models.CreateDisputeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.BadRequest(c, "Invalid request body")
		return
	}

	err := dh.disputeService.CreateDispute(req.MilestoneID, userID.(uint), req.DisputeReason)
	if err != nil {
		middleware.InternalServerError(c, err.Error())
		return
	}

	middleware.Success(c, gin.H{
		"milestone_id": req.MilestoneID,
		"stake_amount": 100, // $BLUEPRINT
	}, "Dispute created successfully. Voting period will begin soon.")
}

// 🗳️ 분쟁 투표
func (dh *DisputeHandler) SubmitVote(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		middleware.Unauthorized(c, "User not authenticated")
		return
	}

	var req models.SubmitVoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.BadRequest(c, "Invalid request body")
		return
	}

	err := dh.disputeService.SubmitVote(req.DisputeID, userID.(uint), req.Choice)
	if err != nil {
		middleware.InternalServerError(c, err.Error())
		return
	}

	middleware.Success(c, gin.H{
		"dispute_id": req.DisputeID,
		"choice":     req.Choice,
	}, "Vote submitted successfully")
}

// 📊 분쟁 상세 정보 조회
func (dh *DisputeHandler) GetDisputeDetail(c *gin.Context) {
	disputeIDStr := c.Param("id")
	disputeID, err := strconv.ParseUint(disputeIDStr, 10, 32)
	if err != nil {
		middleware.BadRequest(c, "Invalid dispute ID")
		return
	}

	disputeDetail, err := dh.disputeService.GetDisputeDetail(uint(disputeID))
	if err != nil {
		middleware.NotFound(c, "Dispute not found")
		return
	}

	middleware.Success(c, disputeDetail, "")
}

// 📋 마일스톤별 분쟁 목록 조회
func (dh *DisputeHandler) GetMilestoneDisputes(c *gin.Context) {
	milestoneIDStr := c.Param("id")
	milestoneID, err := strconv.ParseUint(milestoneIDStr, 10, 32)
	if err != nil {
		middleware.BadRequest(c, "Invalid milestone ID")
		return
	}

	// TODO: 분쟁 목록 조회 로직 구현
	middleware.Success(c, gin.H{
		"milestone_id": milestoneID,
		"disputes":     []gin.H{}, // 임시 빈 배열
	}, "")
}

// 🏛️ 현재 진행중인 분쟁 목록 (거버넌스 탭용)
func (dh *DisputeHandler) GetActiveDisputes(c *gin.Context) {
	// DisputeService를 사용하여 실제 데이터 조회
	disputes, err := dh.disputeService.GetActiveDisputes()
	if err != nil {
		middleware.InternalServerError(c, fmt.Sprintf("Failed to get active disputes: %v", err))
		return
	}

	middleware.Success(c, disputes, "")
}

// ⏰ 분쟁 타이머 상태 조회 (실시간 업데이트용)
func (dh *DisputeHandler) GetDisputeTimer(c *gin.Context) {
	disputeIDStr := c.Param("id")
	disputeID, err := strconv.ParseUint(disputeIDStr, 10, 32)
	if err != nil {
		middleware.BadRequest(c, "Invalid dispute ID")
		return
	}

	// TODO: 실제 타이머 계산 로직
	middleware.Success(c, gin.H{
		"dispute_id":     disputeID,
		"phase":          "voting_period",
		"time_remaining": gin.H{
			"hours":      47,
			"minutes":    23,
			"seconds":    15,
			"is_expired": false,
		},
	}, "")
}
