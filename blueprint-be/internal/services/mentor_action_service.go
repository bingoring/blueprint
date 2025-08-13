package services

import (
	"blueprint/internal/models"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"gorm.io/gorm"
)

// 🎯 멘토 액션 서비스 - 구체적인 가치 창출을 위한 액션 기반 멘토링
type MentorActionService struct {
	db         *gorm.DB
	sseService *SSEService
}

// NewMentorActionService 액션 서비스 생성자
func NewMentorActionService(db *gorm.DB, sseService *SSEService) *MentorActionService {
	return &MentorActionService{
		db:         db,
		sseService: sseService,
	}
}

// TaskProposalData 과제 제안 데이터
type TaskProposalData struct {
	Goals       []string `json:"goals"`        // 목표들
	Resources   []string `json:"resources"`    // 필요 리소스
	Deliverable string   `json:"deliverable"`  // 산출물
	CheckPoints []string `json:"checkpoints"`  // 체크포인트들
}

// FeedbackData 피드백 데이터
type FeedbackData struct {
	Rating      float64           `json:"rating"`       // 평점 (1-10)
	Strengths   []string          `json:"strengths"`    // 강점들
	Improvements []string         `json:"improvements"` // 개선점들
	Suggestions []string          `json:"suggestions"`  // 제안사항들
	NextSteps   []string          `json:"next_steps"`   // 다음 단계
	Categories  map[string]float64 `json:"categories"`  // 카테고리별 평점
}

// ResourceShareData 리소스 공유 데이터
type ResourceShareData struct {
	Type        string   `json:"type"`         // "link", "file", "book", "tool", "contact"
	Name        string   `json:"name"`         // 리소스 이름
	URL         string   `json:"url"`          // URL (해당시)
	Description string   `json:"description"`  // 설명
	Tags        []string `json:"tags"`         // 태그들
	IsEssential bool     `json:"is_essential"` // 필수 리소스 여부
}

// MeetingRequestData 미팅 요청 데이터
type MeetingRequestData struct {
	Type           string    `json:"type"`            // "video", "phone", "in_person"
	Duration       int       `json:"duration"`        // 예상 시간 (분)
	ProposedTimes  []string  `json:"proposed_times"`  // 제안 시간들 (ISO format)
	Agenda         []string  `json:"agenda"`          // 아젠다
	PreparationNeeds []string `json:"preparation_needs"` // 사전 준비사항
	Platform       string    `json:"platform"`        // 플랫폼 (Zoom, Google Meet 등)
}

// ProgressCheckData 진행상황 점검 데이터
type ProgressCheckData struct {
	CheckItems    []CheckItem `json:"check_items"`     // 점검 항목들
	OverallStatus string      `json:"overall_status"`  // 전반적 상태
	Blockers      []string    `json:"blockers"`        // 장애요인들
	Achievements  []string    `json:"achievements"`    // 성취한 것들
	NextMilestone string      `json:"next_milestone"`  // 다음 목표
}

type CheckItem struct {
	Item       string  `json:"item"`       // 점검 항목
	Status     string  `json:"status"`     // "completed", "in_progress", "pending", "blocked"
	Progress   float64 `json:"progress"`   // 진행률 (0-100)
	Comment    string  `json:"comment"`    // 코멘트
}

// CreateTaskProposal 핵심 과제 제안
func (mas *MentorActionService) CreateTaskProposal(sessionID, mentorID, menteeID uint, title, description string,
	dueDate *time.Time, priority int, taskData TaskProposalData) (*models.MentorAction, error) {

	// 데이터 JSON 인코딩
	contentBytes, err := json.Marshal(taskData)
	if err != nil {
		return nil, fmt.Errorf("failed to encode task data: %v", err)
	}

	action := models.MentorAction{
		SessionID:   sessionID,
		MentorID:    mentorID,
		MenteeID:    menteeID,
		Type:        models.ActionTypeTaskProposal,
		Status:      models.ActionStatusProposed,
		Title:       title,
		Description: description,
		Content:     string(contentBytes),
		DueDate:     dueDate,
		Priority:    priority,
	}

	if err := mas.db.Create(&action).Error; err != nil {
		return nil, err
	}

	log.Printf("📋 Task proposal created: %s (mentor %d → mentee %d)", title, mentorID, menteeID)

	// 멘티에게 알림
	go mas.notifyActionCreated(&action, "task_proposed")

	// 세션 통계 업데이트
	go mas.updateSessionStats(sessionID, "action_created")

	return &action, nil
}

// SubmitFeedback 피드백 제출
func (mas *MentorActionService) SubmitFeedback(sessionID, mentorID, menteeID uint, title, description string,
	feedbackData FeedbackData) (*models.MentorAction, error) {

	// 데이터 JSON 인코딩
	contentBytes, err := json.Marshal(feedbackData)
	if err != nil {
		return nil, fmt.Errorf("failed to encode feedback data: %v", err)
	}

	action := models.MentorAction{
		SessionID:   sessionID,
		MentorID:    mentorID,
		MenteeID:    menteeID,
		Type:        models.ActionTypeFeedback,
		Status:      models.ActionStatusCompleted, // 피드백은 즉시 완료
		Title:       title,
		Description: description,
		Content:     string(contentBytes),
		Priority:    3, // 기본 우선순위
		CompletedAt: &[]time.Time{time.Now()}[0],
	}

	if err := mas.db.Create(&action).Error; err != nil {
		return nil, err
	}

	log.Printf("💬 Feedback submitted: %s (rating: %.1f)", title, feedbackData.Rating)

	// 멘티에게 알림
	go mas.notifyActionCreated(&action, "feedback_received")

	// 세션 통계 업데이트
	go mas.updateSessionStats(sessionID, "feedback_given")

	return &action, nil
}

// ShareResource 리소스 공유
func (mas *MentorActionService) ShareResource(sessionID, mentorID, menteeID uint, resourceData ResourceShareData) (*models.MentorAction, error) {
	// 데이터 JSON 인코딩
	contentBytes, err := json.Marshal(resourceData)
	if err != nil {
		return nil, fmt.Errorf("failed to encode resource data: %v", err)
	}

	priority := 3
	if resourceData.IsEssential {
		priority = 5 // 필수 리소스는 높은 우선순위
	}

	action := models.MentorAction{
		SessionID:   sessionID,
		MentorID:    mentorID,
		MenteeID:    menteeID,
		Type:        models.ActionTypeResourceShare,
		Status:      models.ActionStatusCompleted, // 리소스 공유는 즉시 완료
		Title:       fmt.Sprintf("Shared: %s", resourceData.Name),
		Description: resourceData.Description,
		Content:     string(contentBytes),
		Priority:    priority,
		CompletedAt: &[]time.Time{time.Now()}[0],
	}

	if err := mas.db.Create(&action).Error; err != nil {
		return nil, err
	}

	log.Printf("📚 Resource shared: %s (%s)", resourceData.Name, resourceData.Type)

	// 멘티에게 알림
	go mas.notifyActionCreated(&action, "resource_shared")

	// 세션 통계 업데이트
	go mas.updateSessionStats(sessionID, "resource_shared")

	return &action, nil
}

// RequestMeeting 미팅 요청
func (mas *MentorActionService) RequestMeeting(sessionID, mentorID, menteeID uint, meetingData MeetingRequestData) (*models.MentorAction, error) {
	// 데이터 JSON 인코딩
	contentBytes, err := json.Marshal(meetingData)
	if err != nil {
		return nil, fmt.Errorf("failed to encode meeting data: %v", err)
	}

	action := models.MentorAction{
		SessionID:   sessionID,
		MentorID:    mentorID,
		MenteeID:    menteeID,
		Type:        models.ActionTypeMeetingRequest,
		Status:      models.ActionStatusProposed,
		Title:       fmt.Sprintf("%s Meeting Request", meetingData.Type),
		Description: fmt.Sprintf("%d-minute %s meeting", meetingData.Duration, meetingData.Type),
		Content:     string(contentBytes),
		Priority:    4, // 미팅은 높은 우선순위
	}

	if err := mas.db.Create(&action).Error; err != nil {
		return nil, err
	}

	log.Printf("📅 Meeting requested: %s (%d minutes)", meetingData.Type, meetingData.Duration)

	// 멘티에게 알림
	go mas.notifyActionCreated(&action, "meeting_requested")

	return &action, nil
}

// CheckProgress 진행상황 점검
func (mas *MentorActionService) CheckProgress(sessionID, mentorID, menteeID uint, progressData ProgressCheckData) (*models.MentorAction, error) {
	// 데이터 JSON 인코딩
	contentBytes, err := json.Marshal(progressData)
	if err != nil {
		return nil, fmt.Errorf("failed to encode progress data: %v", err)
	}

	// 전체 진행률 계산
	totalProgress := 0.0
	if len(progressData.CheckItems) > 0 {
		for _, item := range progressData.CheckItems {
			totalProgress += item.Progress
		}
		totalProgress /= float64(len(progressData.CheckItems))
	}

	action := models.MentorAction{
		SessionID:   sessionID,
		MentorID:    mentorID,
		MenteeID:    menteeID,
		Type:        models.ActionTypeProgressCheck,
		Status:      models.ActionStatusCompleted,
		Title:       fmt.Sprintf("Progress Check - %.1f%% Complete", totalProgress),
		Description: progressData.OverallStatus,
		Content:     string(contentBytes),
		Priority:    3,
		CompletedAt: &[]time.Time{time.Now()}[0],
	}

	if err := mas.db.Create(&action).Error; err != nil {
		return nil, err
	}

	log.Printf("📈 Progress check completed: %.1f%% overall progress", totalProgress)

	// 멘티에게 알림
	go mas.notifyActionCreated(&action, "progress_checked")

	// 세션 통계 업데이트
	go mas.updateSessionStats(sessionID, "progress_checked")

	return &action, nil
}

// AcceptAction 액션 수락 (멘티가 과제나 미팅 요청 수락)
func (mas *MentorActionService) AcceptAction(actionID, userID uint, response string) error {
	// 액션 조회 및 권한 확인
	var action models.MentorAction
	if err := mas.db.Where("id = ?", actionID).First(&action).Error; err != nil {
		return fmt.Errorf("action not found: %v", err)
	}

	if action.MenteeID != userID {
		return fmt.Errorf("unauthorized to accept this action")
	}

	if action.Status != models.ActionStatusProposed {
		return fmt.Errorf("action is not in proposed status")
	}

	// 상태 업데이트
	action.Status = models.ActionStatusAccepted
	action.MenteeResponse = response

	if err := mas.db.Save(&action).Error; err != nil {
		return err
	}

	log.Printf("✅ Action accepted: %s (ID: %d)", action.Title, actionID)

	// 멘토에게 알림
	go mas.notifyActionStatusChanged(&action, "action_accepted")

	return nil
}

// CompleteAction 액션 완료 (멘티가 과제 완료, 결과 제출)
func (mas *MentorActionService) CompleteAction(actionID, userID uint, response string, resultFiles []string) error {
	// 액션 조회 및 권한 확인
	var action models.MentorAction
	if err := mas.db.Where("id = ?", actionID).First(&action).Error; err != nil {
		return fmt.Errorf("action not found: %v", err)
	}

	if action.MenteeID != userID {
		return fmt.Errorf("unauthorized to complete this action")
	}

	if action.Status != models.ActionStatusAccepted && action.Status != models.ActionStatusInProgress {
		return fmt.Errorf("action is not in progress")
	}

	// 상태 업데이트
	now := time.Now()
	action.Status = models.ActionStatusCompleted
	action.MenteeResponse = response
	action.ResultFiles = resultFiles
	action.CompletedAt = &now

	if err := mas.db.Save(&action).Error; err != nil {
		return err
	}

	log.Printf("🎉 Action completed: %s (ID: %d)", action.Title, actionID)

	// 멘토에게 알림 및 세션 통계 업데이트
	go mas.notifyActionStatusChanged(&action, "action_completed")
	go mas.updateSessionStats(action.SessionID, "action_completed")

	return nil
}

// RateAction 액션 평가 (멘티가 멘토 액션 평가)
func (mas *MentorActionService) RateAction(actionID, userID uint, rating float64) error {
	// 액션 조회 및 권한 확인
	var action models.MentorAction
	if err := mas.db.Where("id = ?", actionID).First(&action).Error; err != nil {
		return fmt.Errorf("action not found: %v", err)
	}

	if action.MenteeID != userID {
		return fmt.Errorf("unauthorized to rate this action")
	}

	if action.Status != models.ActionStatusCompleted {
		return fmt.Errorf("action is not completed")
	}

	// 평점 업데이트
	action.MenteeRating = rating

	if err := mas.db.Save(&action).Error; err != nil {
		return err
	}

	log.Printf("⭐ Action rated: %s (rating: %.1f)", action.Title, rating)

	// 멘토 통계 업데이트 (비동기)
	go mas.updateMentorRatingStats(action.MentorID, rating)

	return nil
}

// GetSessionActions 세션의 모든 액션들 조회
func (mas *MentorActionService) GetSessionActions(sessionID uint, userID uint) ([]models.MentorAction, error) {
	// 권한 확인 (멘토 또는 멘티만 조회 가능)
	var session models.MentoringSession
	if err := mas.db.Where("id = ?", sessionID).First(&session).Error; err != nil {
		return nil, fmt.Errorf("session not found: %v", err)
	}

	if session.MentorID != userID && session.MenteeID != userID {
		return nil, fmt.Errorf("unauthorized to access this session")
	}

	// 액션들 조회
	var actions []models.MentorAction
	if err := mas.db.Where("session_id = ?", sessionID).
		Order("priority DESC, created_at DESC").
		Find(&actions).Error; err != nil {
		return nil, err
	}

	return actions, nil
}

// GetActionsByType 타입별 액션 조회
func (mas *MentorActionService) GetActionsByType(sessionID uint, actionType models.MentorActionType, userID uint) ([]models.MentorAction, error) {
	// 권한 확인
	var session models.MentoringSession
	if err := mas.db.Where("id = ?", sessionID).First(&session).Error; err != nil {
		return nil, fmt.Errorf("session not found: %v", err)
	}

	if session.MentorID != userID && session.MenteeID != userID {
		return nil, fmt.Errorf("unauthorized to access this session")
	}

	// 타입별 액션들 조회
	var actions []models.MentorAction
	if err := mas.db.Where("session_id = ? AND type = ?", sessionID, actionType).
		Order("created_at DESC").
		Find(&actions).Error; err != nil {
		return nil, err
	}

	return actions, nil
}

// GetPendingActions 대기 중인 액션들 조회 (멘티용)
func (mas *MentorActionService) GetPendingActions(menteeID uint) ([]models.MentorAction, error) {
	var actions []models.MentorAction
	if err := mas.db.Where("mentee_id = ? AND status IN ?", menteeID,
		[]models.MentorActionStatus{models.ActionStatusProposed, models.ActionStatusAccepted}).
		Preload("Session").Preload("Mentor").
		Order("priority DESC, created_at DESC").
		Find(&actions).Error; err != nil {
		return nil, err
	}

	return actions, nil
}

// updateSessionStats 세션 통계 업데이트
func (mas *MentorActionService) updateSessionStats(sessionID uint, eventType string) {
	// 세션 조회
	var session models.MentoringSession
	if err := mas.db.Where("id = ?", sessionID).First(&session).Error; err != nil {
		log.Printf("❌ Failed to find session %d for stats update: %v", sessionID, err)
		return
	}

	// 이벤트 타입에 따른 통계 업데이트
	updates := make(map[string]interface{})
	now := time.Now()

	switch eventType {
	case "action_created":
		updates["actions_count"] = gorm.Expr("actions_count + 1")
		updates["last_message_at"] = now
	case "feedback_given":
		updates["actions_count"] = gorm.Expr("actions_count + 1")
		updates["last_message_at"] = now
	case "resource_shared":
		updates["files_shared"] = gorm.Expr("files_shared + 1")
		updates["last_message_at"] = now
	case "progress_checked":
		updates["actions_count"] = gorm.Expr("actions_count + 1")
		updates["last_message_at"] = now
	case "action_completed":
		updates["last_message_at"] = now
	}

	if len(updates) > 0 {
		if err := mas.db.Model(&session).Updates(updates).Error; err != nil {
			log.Printf("❌ Failed to update session stats: %v", err)
		}
	}
}

// updateMentorRatingStats 멘토 평점 통계 업데이트
func (mas *MentorActionService) updateMentorRatingStats(mentorID uint, rating float64) {
	// 해당 멘토의 모든 평점 다시 계산
	var avgRating float64
	if err := mas.db.Model(&models.MentorAction{}).
		Where("mentor_id = ? AND mentee_rating > 0", mentorID).
		Select("AVG(mentee_rating)").Scan(&avgRating).Error; err != nil {
		log.Printf("❌ Failed to calculate average rating for mentor %d: %v", mentorID, err)
		return
	}

	// 멘토 평점 업데이트
	if err := mas.db.Model(&models.Mentor{}).Where("id = ?", mentorID).
		Update("average_rating", avgRating).Error; err != nil {
		log.Printf("❌ Failed to update mentor average rating: %v", err)
	}
}

// notifyActionCreated 액션 생성 알림
func (mas *MentorActionService) notifyActionCreated(action *models.MentorAction, eventType string) {
	if mas.sseService == nil {
		return
	}

	// 세션 정보 조회
	var session models.MentoringSession
	if err := mas.db.Where("id = ?", action.SessionID).First(&session).Error; err != nil {
		return
	}

	event := MarketUpdateEvent{
		MilestoneID: session.MilestoneID,
		MarketData: map[string]interface{}{
			"event_type": eventType,
			"data":       action,
			"session_id": session.ID,
		},
		Timestamp: time.Now().Unix(),
	}

	mas.sseService.BroadcastMarketUpdate(event)
}

// notifyActionStatusChanged 액션 상태 변경 알림
func (mas *MentorActionService) notifyActionStatusChanged(action *models.MentorAction, eventType string) {
	if mas.sseService == nil {
		return
	}

	// 세션 정보 조회
	var session models.MentoringSession
	if err := mas.db.Where("id = ?", action.SessionID).First(&session).Error; err != nil {
		return
	}

	event := MarketUpdateEvent{
		MilestoneID: session.MilestoneID,
		MarketData: map[string]interface{}{
			"event_type": eventType,
			"data":       action,
			"session_id": session.ID,
		},
		Timestamp: time.Now().Unix(),
	}

	mas.sseService.BroadcastMarketUpdate(event)
}
