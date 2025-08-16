package services

import (
	"blueprint-module/pkg/models"
	"fmt"
	"log"
	"time"

	"gorm.io/gorm"
)

// 🤝 멘토-진행자 매칭 서비스
type MentorMatchingService struct {
	db         *gorm.DB
	sseService *SSEService
}

// NewMentorMatchingService 매칭 서비스 생성자
func NewMentorMatchingService(db *gorm.DB, sseService *SSEService) *MentorMatchingService {
	return &MentorMatchingService{
		db:         db,
		sseService: sseService,
	}
}

// MentorCandidateInfo 멘토 후보 정보
type MentorCandidateInfo struct {
	Mentor           models.Mentor          `json:"mentor"`
	MentorMilestone  models.MentorMilestone `json:"mentor_milestone"`
	User             models.User            `json:"user"`
	TotalBetAmount   int64                  `json:"total_bet_amount"`
	BetSharePercent  float64                `json:"bet_share_percent"`
	IsLeadMentor     bool                   `json:"is_lead_mentor"`
	LeadMentorRank   int                    `json:"lead_mentor_rank"`
	SuccessRate      float64                `json:"success_rate"`
	ReputationScore  int                    `json:"reputation_score"`
	IsAvailable      bool                   `json:"is_available"`
	ActiveMentorings int                    `json:"active_mentorings"`
}

// MentorProjectInfo 멘토가 베팅한 프로젝트 정보
type MentorProjectInfo struct {
	Project         models.Project         `json:"project"`
	Milestone       models.Milestone       `json:"milestone"`
	MentorMilestone models.MentorMilestone `json:"mentor_milestone"`
	ProjectOwner    models.User            `json:"project_owner"`
	TotalBetAmount  int64                  `json:"total_bet_amount"`
	BetSharePercent float64                `json:"bet_share_percent"`
	IsLeadMentor    bool                   `json:"is_lead_mentor"`
	MentoringStatus string                 `json:"mentoring_status"` // "available", "requested", "active"
}

// MentoringRequest 멘토링 요청
type MentoringRequest struct {
	ID          uint `json:"id" gorm:"primaryKey"`
	MentorID    uint `json:"mentor_id" gorm:"not null;index"`
	MenteeID    uint `json:"mentee_id" gorm:"not null;index"`
	MilestoneID uint `json:"milestone_id" gorm:"not null;index"`
	ProjectID   uint `json:"project_id" gorm:"not null;index"`

	// 요청 정보
	RequestType string `json:"request_type"`                    // "mentor_initiated", "mentee_initiated"
	Status      string `json:"status" gorm:"default:'pending'"` // "pending", "accepted", "rejected", "expired"
	Message     string `json:"message" gorm:"type:text"`

	// 제안 조건 (멘토가 제시)
	ProposedDuration int `json:"proposed_duration"` // 예상 멘토링 기간 (주)
	ProposedMeetings int `json:"proposed_meetings"` // 예상 미팅 횟수
	ExpectedTime     int `json:"expected_time"`     // 예상 소요 시간 (시간/주)

	// 응답 정보
	ResponseMessage string     `json:"response_message" gorm:"type:text"`
	ResponsedAt     *time.Time `json:"responsed_at,omitempty"`
	ExpiresAt       time.Time  `json:"expires_at"` // 요청 만료일

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// 관계
	Mentor    models.Mentor    `json:"mentor,omitempty" gorm:"foreignKey:MentorID"`
	Mentee    models.User      `json:"mentee,omitempty" gorm:"foreignKey:MenteeID"`
	Milestone models.Milestone `json:"milestone,omitempty" gorm:"foreignKey:MilestoneID"`
	Project   models.Project   `json:"project,omitempty" gorm:"foreignKey:ProjectID"`
}

// GetMentorCandidatesForMilestone 특정 마일스톤의 멘토 후보들 조회 (진행자용)
func (mms *MentorMatchingService) GetMentorCandidatesForMilestone(milestoneID uint, menteeID uint) ([]MentorCandidateInfo, error) {
	// 1. 해당 마일스톤에 베팅한 멘토들 조회
	var mentorMilestones []models.MentorMilestone
	if err := mms.db.Where("milestone_id = ?", milestoneID).
		Preload("Mentor").Preload("Mentor.User").
		Order("is_lead_mentor DESC, total_bet_amount DESC").
		Find(&mentorMilestones).Error; err != nil {
		return nil, err
	}

	// 2. 각 멘토의 활성 멘토링 수 조회
	mentorIDs := make([]uint, 0, len(mentorMilestones))
	for _, mm := range mentorMilestones {
		mentorIDs = append(mentorIDs, mm.MentorID)
	}

	activeMentorings := make(map[uint]int)
	if len(mentorIDs) > 0 {
		var counts []struct {
			MentorID uint `gorm:"column:mentor_id"`
			Count    int  `gorm:"column:count"`
		}
		if err := mms.db.Model(&models.MentoringSession{}).
			Select("mentor_id, count(*) as count").
			Where("mentor_id IN ? AND status = ?", mentorIDs, models.SessionStatusActive).
			Group("mentor_id").Find(&counts).Error; err == nil {
			for _, count := range counts {
				activeMentorings[count.MentorID] = count.Count
			}
		}
	}

	// 3. 멘토 후보 정보 구성
	candidates := make([]MentorCandidateInfo, 0, len(mentorMilestones))
	for _, mm := range mentorMilestones {
		if mm.Mentor.ID == 0 {
			continue // 멘토 정보가 없으면 스킵
		}

		candidate := MentorCandidateInfo{
			Mentor:           mm.Mentor,
			MentorMilestone:  mm,
			User:             mm.Mentor.User,
			TotalBetAmount:   mm.TotalBetAmount,
			BetSharePercent:  mm.BetSharePercentage,
			IsLeadMentor:     mm.IsLeadMentor,
			LeadMentorRank:   mm.LeadMentorRank,
			SuccessRate:      mm.Mentor.SuccessRate,
			ReputationScore:  mm.Mentor.ReputationScore,
			IsAvailable:      mm.Mentor.CanTakeNewMentoring(),
			ActiveMentorings: activeMentorings[mm.MentorID],
		}

		candidates = append(candidates, candidate)
	}

	return candidates, nil
}

// GetMentorProjects 멘토가 베팅한 프로젝트들 조회 (멘토용)
func (mms *MentorMatchingService) GetMentorProjects(mentorID uint) ([]MentorProjectInfo, error) {
	// 멘토가 베팅한 마일스톤들 조회
	var mentorMilestones []models.MentorMilestone
	if err := mms.db.Where("mentor_id = ?", mentorID).
		Preload("Project").Preload("Project.User").
		Preload("Milestone").
		Order("total_bet_amount DESC").
		Find(&mentorMilestones).Error; err != nil {
		return nil, err
	}

	// 기존 멘토링 요청/세션 상태 조회
	milestoneIDs := make([]uint, 0, len(mentorMilestones))
	for _, mm := range mentorMilestones {
		milestoneIDs = append(milestoneIDs, mm.MilestoneID)
	}

	mentoringStatuses := make(map[uint]string)
	if len(milestoneIDs) > 0 {
		// 활성 세션 조회
		var activeSessions []models.MentoringSession
		if err := mms.db.Where("mentor_id = ? AND milestone_id IN ? AND status = ?",
			mentorID, milestoneIDs, models.SessionStatusActive).
			Find(&activeSessions).Error; err == nil {
			for _, session := range activeSessions {
				mentoringStatuses[session.MilestoneID] = "active"
			}
		}

		// 대기 중인 요청 조회
		var pendingRequests []MentoringRequest
		if err := mms.db.Where("mentor_id = ? AND milestone_id IN ? AND status = ?",
			mentorID, milestoneIDs, "pending").
			Find(&pendingRequests).Error; err == nil {
			for _, request := range pendingRequests {
				if _, exists := mentoringStatuses[request.MilestoneID]; !exists {
					mentoringStatuses[request.MilestoneID] = "requested"
				}
			}
		}
	}

	// 프로젝트 정보 구성
	projects := make([]MentorProjectInfo, 0, len(mentorMilestones))
	for _, mm := range mentorMilestones {
		status := "available"
		if existingStatus, exists := mentoringStatuses[mm.MilestoneID]; exists {
			status = existingStatus
		}

		project := MentorProjectInfo{
			Project:         mm.Project,
			Milestone:       mm.Milestone,
			MentorMilestone: mm,
			ProjectOwner:    mm.Project.User,
			TotalBetAmount:  mm.TotalBetAmount,
			BetSharePercent: mm.BetSharePercentage,
			IsLeadMentor:    mm.IsLeadMentor,
			MentoringStatus: status,
		}

		projects = append(projects, project)
	}

	return projects, nil
}

// RequestMentoring 멘토링 요청 생성 (진행자가 멘토에게 요청)
func (mms *MentorMatchingService) RequestMentoring(menteeID, mentorID, milestoneID uint, message string) (*MentoringRequest, error) {
	// 1. 유효성 검사
	if err := mms.validateMentoringRequest(menteeID, mentorID, milestoneID); err != nil {
		return nil, err
	}

	// 2. 멘토링 요청 생성
	request := MentoringRequest{
		MentorID:    mentorID,
		MenteeID:    menteeID,
		MilestoneID: milestoneID,
		RequestType: "mentee_initiated",
		Status:      "pending",
		Message:     message,
		ExpiresAt:   time.Now().AddDate(0, 0, 7), // 7일 후 만료
	}

	// 프로젝트 ID 조회
	var milestone models.Milestone
	if err := mms.db.Where("id = ?", milestoneID).First(&milestone).Error; err != nil {
		return nil, fmt.Errorf("milestone not found: %v", err)
	}
	request.ProjectID = milestone.ProjectID

	if err := mms.db.Create(&request).Error; err != nil {
		return nil, err
	}

	log.Printf("📨 Mentoring request created: mentee %d → mentor %d for milestone %d",
		menteeID, mentorID, milestoneID)

	// 3. 멘토에게 알림
	go mms.notifyMentoringRequest(&request, "request_received")

	return &request, nil
}

// ProposeMentoring 멘토링 제안 (멘토가 진행자에게 제안)
func (mms *MentorMatchingService) ProposeMentoring(mentorID, menteeID, milestoneID uint, message string,
	proposedDuration, proposedMeetings, expectedTime int) (*MentoringRequest, error) {

	// 1. 유효성 검사
	if err := mms.validateMentoringProposal(mentorID, menteeID, milestoneID); err != nil {
		return nil, err
	}

	// 2. 멘토링 제안 생성
	request := MentoringRequest{
		MentorID:         mentorID,
		MenteeID:         menteeID,
		MilestoneID:      milestoneID,
		RequestType:      "mentor_initiated",
		Status:           "pending",
		Message:          message,
		ProposedDuration: proposedDuration,
		ProposedMeetings: proposedMeetings,
		ExpectedTime:     expectedTime,
		ExpiresAt:        time.Now().AddDate(0, 0, 7), // 7일 후 만료
	}

	// 프로젝트 ID 조회
	var milestone models.Milestone
	if err := mms.db.Where("id = ?", milestoneID).First(&milestone).Error; err != nil {
		return nil, fmt.Errorf("milestone not found: %v", err)
	}
	request.ProjectID = milestone.ProjectID

	if err := mms.db.Create(&request).Error; err != nil {
		return nil, err
	}

	log.Printf("💡 Mentoring proposal created: mentor %d → mentee %d for milestone %d",
		mentorID, menteeID, milestoneID)

	// 3. 진행자에게 알림
	go mms.notifyMentoringRequest(&request, "proposal_received")

	return &request, nil
}

// AcceptMentoringRequest 멘토링 요청 수락
func (mms *MentorMatchingService) AcceptMentoringRequest(requestID uint, userID uint, responseMessage string) (*models.MentoringSession, error) {
	// 1. 요청 조회 및 권한 확인
	var request MentoringRequest
	if err := mms.db.Where("id = ?", requestID).First(&request).Error; err != nil {
		return nil, fmt.Errorf("request not found: %v", err)
	}

	// 권한 확인 (요청 받은 사람만 수락 가능)
	if (request.RequestType == "mentee_initiated" && request.MentorID != userID) ||
		(request.RequestType == "mentor_initiated" && request.MenteeID != userID) {
		return nil, fmt.Errorf("unauthorized to accept this request")
	}

	if request.Status != "pending" {
		return nil, fmt.Errorf("request is not pending (status: %s)", request.Status)
	}

	// 2. 트랜잭션으로 처리
	tx := mms.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 요청 상태 업데이트
	now := time.Now()
	request.Status = "accepted"
	request.ResponseMessage = responseMessage
	request.ResponsedAt = &now

	if err := tx.Save(&request).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// 3. 멘토링 세션 생성
	session := models.MentoringSession{
		MentorID:    request.MentorID,
		MenteeID:    request.MenteeID,
		MilestoneID: request.MilestoneID,
		ProjectID:   request.ProjectID,
		Status:      models.SessionStatusActive,
		Title:       fmt.Sprintf("Mentoring for Milestone"),
		Description: request.Message,
		StartedAt:   now,
	}

	if err := tx.Create(&session).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// 4. MentorMilestone 상태 업데이트 (활성화)
	if err := tx.Model(&models.MentorMilestone{}).
		Where("mentor_id = ? AND milestone_id = ?", request.MentorID, request.MilestoneID).
		Updates(map[string]interface{}{
			"is_active":        true,
			"started_at":       &now,
			"last_activity_at": &now,
		}).Error; err != nil {
		log.Printf("⚠️ Failed to update mentor milestone status: %v", err)
		// 이 오류는 치명적이지 않으므로 계속 진행
	}

	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %v", err)
	}

	log.Printf("🤝 Mentoring session started: mentor %d ↔ mentee %d for milestone %d",
		request.MentorID, request.MenteeID, request.MilestoneID)

	// 5. 양쪽에 알림
	go mms.notifyMentoringStarted(&session)

	return &session, nil
}

// RejectMentoringRequest 멘토링 요청 거절
func (mms *MentorMatchingService) RejectMentoringRequest(requestID uint, userID uint, responseMessage string) error {
	// 요청 조회 및 권한 확인
	var request MentoringRequest
	if err := mms.db.Where("id = ?", requestID).First(&request).Error; err != nil {
		return fmt.Errorf("request not found: %v", err)
	}

	// 권한 확인
	if (request.RequestType == "mentee_initiated" && request.MentorID != userID) ||
		(request.RequestType == "mentor_initiated" && request.MenteeID != userID) {
		return fmt.Errorf("unauthorized to reject this request")
	}

	if request.Status != "pending" {
		return fmt.Errorf("request is not pending (status: %s)", request.Status)
	}

	// 요청 상태 업데이트
	now := time.Now()
	request.Status = "rejected"
	request.ResponseMessage = responseMessage
	request.ResponsedAt = &now

	if err := mms.db.Save(&request).Error; err != nil {
		return err
	}

	log.Printf("❌ Mentoring request rejected: %d", requestID)

	// 상대방에게 거절 알림
	go mms.notifyMentoringRequest(&request, "request_rejected")

	return nil
}

// validateMentoringRequest 멘토링 요청 유효성 검사
func (mms *MentorMatchingService) validateMentoringRequest(menteeID, mentorID, milestoneID uint) error {
	// 1. 중복 요청 확인
	var existingRequest MentoringRequest
	if err := mms.db.Where("mentor_id = ? AND mentee_id = ? AND milestone_id = ? AND status = ?",
		mentorID, menteeID, milestoneID, "pending").First(&existingRequest).Error; err == nil {
		return fmt.Errorf("pending request already exists")
	}

	// 2. 활성 세션 확인
	var existingSession models.MentoringSession
	if err := mms.db.Where("mentor_id = ? AND mentee_id = ? AND milestone_id = ? AND status = ?",
		mentorID, menteeID, milestoneID, models.SessionStatusActive).First(&existingSession).Error; err == nil {
		return fmt.Errorf("active mentoring session already exists")
	}

	// 3. 멘토 자격 확인
	var mentorMilestone models.MentorMilestone
	if err := mms.db.Where("mentor_id = ? AND milestone_id = ?", mentorID, milestoneID).First(&mentorMilestone).Error; err != nil {
		return fmt.Errorf("mentor is not qualified for this milestone")
	}

	// 4. 멘토 가용성 확인
	var mentor models.Mentor
	if err := mms.db.Where("id = ?", mentorID).First(&mentor).Error; err != nil {
		return fmt.Errorf("mentor not found")
	}

	if !mentor.CanTakeNewMentoring() {
		return fmt.Errorf("mentor is not available for new mentoring")
	}

	return nil
}

// validateMentoringProposal 멘토링 제안 유효성 검사
func (mms *MentorMatchingService) validateMentoringProposal(mentorID, menteeID, milestoneID uint) error {
	// 멘토링 요청 검사와 동일한 로직
	return mms.validateMentoringRequest(menteeID, mentorID, milestoneID)
}

// notifyMentoringRequest 멘토링 요청 알림
func (mms *MentorMatchingService) notifyMentoringRequest(request *MentoringRequest, eventType string) {
	if mms.sseService == nil {
		return
	}

	event := MarketUpdateEvent{
		MilestoneID: request.MilestoneID,
		MarketData: map[string]interface{}{
			"event_type": eventType,
			"data":       request,
		},
		Timestamp: time.Now().Unix(),
	}

	mms.sseService.BroadcastMarketUpdate(event)
}

// notifyMentoringStarted 멘토링 시작 알림
func (mms *MentorMatchingService) notifyMentoringStarted(session *models.MentoringSession) {
	if mms.sseService == nil {
		return
	}

	event := MarketUpdateEvent{
		MilestoneID: session.MilestoneID,
		MarketData: map[string]interface{}{
			"event_type": "mentoring_started",
			"data":       session,
		},
		Timestamp: time.Now().Unix(),
	}

	mms.sseService.BroadcastMarketUpdate(event)
}

// TableName GORM 테이블명 설정
func (MentoringRequest) TableName() string { return "mentoring_requests" }
