package logger

import (
	"blueprint-module/pkg/models"
	"blueprint-module/pkg/queue"
	"context"
	"fmt"
	"log"
	"time"
)

// ActivityLogger 활동 로그 전송을 담당하는 구조체
type ActivityLogger struct {
	queueName string
}

// NewActivityLogger 새로운 ActivityLogger 인스턴스 생성
func NewActivityLogger() *ActivityLogger {
	return &ActivityLogger{
		queueName: "activity_logs",
	}
}

// LogActivity 활동 로그를 큐로 전송 (비동기)
func (a *ActivityLogger) LogActivity(ctx context.Context, req models.CreateActivityLogRequest) error {
	// 컨텍스트 정보 추가
	if req.Metadata.Platform == "" {
		req.Metadata.Platform = "web" // 기본값
	}

	// 큐로 전송할 job 맵 생성
	job := map[string]interface{}{
		"type":           "create_activity_log",
		"user_id":        req.UserID,
		"activity_type":  req.ActivityType,
		"action":         req.Action,
		"description":    req.Description,
		"project_id":     req.ProjectID,
		"milestone_id":   req.MilestoneID,
		"order_id":       req.OrderID,
		"trade_id":       req.TradeID,
		"metadata":       req.Metadata,
		"created_at":     time.Now().Unix(),
	}

	// 큐로 전송
	err := queue.PublishJob(a.queueName, job)
	if err != nil {
		log.Printf("❌ 활동 로그 큐 전송 실패: %v", err)
		return fmt.Errorf("failed to publish activity log: %w", err)
	}

	log.Printf("✅ 활동 로그 큐 전송 성공 (Type: %s, Action: %s)",
		req.ActivityType, req.Action)
	return nil
}

// LogProjectActivity 프로젝트 관련 활동 로그
func (a *ActivityLogger) LogProjectActivity(ctx context.Context, userID uint, action string, projectID uint, projectTitle string, description string) error {
	return a.LogActivity(ctx, models.CreateActivityLogRequest{
		UserID:       userID,
		ActivityType: models.ActivityTypeProject,
		Action:       action,
		Description:  description,
		ProjectID:    &projectID,
		Metadata: models.ActivityMetadata{
			ProjectTitle: projectTitle,
			Platform:     "web",
		},
	})
}

// LogMilestoneActivity 마일스톤 관련 활동 로그
func (a *ActivityLogger) LogMilestoneActivity(ctx context.Context, userID uint, action string, projectID, milestoneID uint, projectTitle, milestoneTitle, description string) error {
	return a.LogActivity(ctx, models.CreateActivityLogRequest{
		UserID:       userID,
		ActivityType: models.ActivityTypeMilestone,
		Action:       action,
		Description:  description,
		ProjectID:    &projectID,
		MilestoneID:  &milestoneID,
		Metadata: models.ActivityMetadata{
			ProjectTitle:   projectTitle,
			MilestoneTitle: milestoneTitle,
			Platform:       "web",
		},
	})
}

// LogTradeActivity 거래 관련 활동 로그
func (a *ActivityLogger) LogTradeActivity(ctx context.Context, userID uint, action string, orderID *uint, tradeID *uint, amount, price float64, orderType, description string) error {
	return a.LogActivity(ctx, models.CreateActivityLogRequest{
		UserID:       userID,
		ActivityType: models.ActivityTypeTrade,
		Action:       action,
		Description:  description,
		OrderID:      orderID,
		TradeID:      tradeID,
		Metadata: models.ActivityMetadata{
			Amount:    amount,
			Price:     price,
			Currency:  "USDC",
			OrderType: orderType,
			Platform:  "web",
		},
	})
}

// LogMentoringActivity 멘토링 관련 활동 로그
func (a *ActivityLogger) LogMentoringActivity(ctx context.Context, userID uint, action string, mentorUsername string, sessionDuration int, rating int, description string) error {
	return a.LogActivity(ctx, models.CreateActivityLogRequest{
		UserID:       userID,
		ActivityType: models.ActivityTypeMentoring,
		Action:       action,
		Description:  description,
		Metadata: models.ActivityMetadata{
			MentorUsername:  mentorUsername,
			SessionDuration: sessionDuration,
			Rating:          rating,
			Platform:        "web",
		},
	})
}

// LogAccountActivity 계정 관련 활동 로그
func (a *ActivityLogger) LogAccountActivity(ctx context.Context, userID uint, action string, description string, ipAddress, userAgent string) error {
	return a.LogActivity(ctx, models.CreateActivityLogRequest{
		UserID:       userID,
		ActivityType: models.ActivityTypeAccount,
		Action:       action,
		Description:  description,
		Metadata: models.ActivityMetadata{
			IPAddress: ipAddress,
			UserAgent: userAgent,
			Platform:  "web",
		},
	})
}

// LogInvestmentActivity 투자 관련 활동 로그
func (a *ActivityLogger) LogInvestmentActivity(ctx context.Context, userID uint, action string, amount float64, profitLoss float64, description string) error {
	return a.LogActivity(ctx, models.CreateActivityLogRequest{
		UserID:       userID,
		ActivityType: models.ActivityTypeInvestment,
		Action:       action,
		Description:  description,
		Metadata: models.ActivityMetadata{
			Amount:     amount,
			Currency:   "USDC",
			ProfitLoss: profitLoss,
			Platform:   "web",
		},
	})
}

// 전역 ActivityLogger 인스턴스
var GlobalActivityLogger = NewActivityLogger()

// 편의 함수들 (전역 인스턴스 사용)
func LogProjectActivity(ctx context.Context, userID uint, action string, projectID uint, projectTitle string, description string) error {
	return GlobalActivityLogger.LogProjectActivity(ctx, userID, action, projectID, projectTitle, description)
}

func LogMilestoneActivity(ctx context.Context, userID uint, action string, projectID, milestoneID uint, projectTitle, milestoneTitle, description string) error {
	return GlobalActivityLogger.LogMilestoneActivity(ctx, userID, action, projectID, milestoneID, projectTitle, milestoneTitle, description)
}

func LogTradeActivity(ctx context.Context, userID uint, action string, orderID *uint, tradeID *uint, amount, price float64, orderType, description string) error {
	return GlobalActivityLogger.LogTradeActivity(ctx, userID, action, orderID, tradeID, amount, price, orderType, description)
}

func LogAccountActivity(ctx context.Context, userID uint, action string, description string, ipAddress, userAgent string) error {
	return GlobalActivityLogger.LogAccountActivity(ctx, userID, action, description, ipAddress, userAgent)
}
