package handlers

import (
	"blueprint-module/pkg/database"
	"blueprint-module/pkg/models"
	"blueprint-module/pkg/redis"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	redislib "github.com/redis/go-redis/v9"
)

// ActivityHandler 활동 로그 처리 핸들러
type ActivityHandler struct{}

// NewActivityHandler ActivityHandler 인스턴스 생성
func NewActivityHandler() *ActivityHandler {
	return &ActivityHandler{}
}

// HandleActivityLogJob 활동 로그 작업 처리
func (h *ActivityHandler) HandleActivityLogJob(jobData map[string]interface{}) error {
	log.Printf("📝 활동 로그 작업 처리 시작: %+v", jobData)

	// 작업 타입 확인
	jobType, ok := jobData["type"].(string)
	if !ok {
		return fmt.Errorf("invalid job type")
	}

	switch jobType {
	case "create_activity_log":
		return h.createActivityLog(jobData)
	default:
		return fmt.Errorf("unknown activity job type: %s", jobType)
	}
}

// createActivityLog 활동 로그를 데이터베이스에 저장
func (h *ActivityHandler) createActivityLog(jobData map[string]interface{}) error {
	// 필수 필드 추출
	userID, ok := jobData["user_id"].(float64)
	if !ok {
		return fmt.Errorf("invalid user_id")
	}

	activityType, ok := jobData["activity_type"].(string)
	if !ok {
		return fmt.Errorf("invalid activity_type")
	}

	action, ok := jobData["action"].(string)
	if !ok {
		return fmt.Errorf("invalid action")
	}

	description, _ := jobData["description"].(string)

	// 선택적 필드 추출 (nil 가능)
	var projectID, milestoneID, orderID, tradeID *uint

	if pid, exists := jobData["project_id"]; exists && pid != nil {
		if pidFloat, ok := pid.(float64); ok {
			pidUint := uint(pidFloat)
			projectID = &pidUint
		}
	}

	if mid, exists := jobData["milestone_id"]; exists && mid != nil {
		if midFloat, ok := mid.(float64); ok {
			midUint := uint(midFloat)
			milestoneID = &midUint
		}
	}

	if oid, exists := jobData["order_id"]; exists && oid != nil {
		if oidFloat, ok := oid.(float64); ok {
			oidUint := uint(oidFloat)
			orderID = &oidUint
		}
	}

	if tid, exists := jobData["trade_id"]; exists && tid != nil {
		if tidFloat, ok := tid.(float64); ok {
			tidUint := uint(tidFloat)
			tradeID = &tidUint
		}
	}

	// 메타데이터 추출 및 변환
	var metadata models.ActivityMetadata
	if metaData, exists := jobData["metadata"]; exists && metaData != nil {
		// 메타데이터를 JSON으로 변환 후 다시 구조체로 파싱
		metaBytes, err := json.Marshal(metaData)
		if err != nil {
			log.Printf("⚠️ 메타데이터 직렬화 실패: %v", err)
		} else {
			if err := json.Unmarshal(metaBytes, &metadata); err != nil {
				log.Printf("⚠️ 메타데이터 파싱 실패: %v", err)
			}
		}
	}

	// ActivityLog 인스턴스 생성
	activityLog := models.ActivityLog{
		UserID:       uint(userID),
		ActivityType: activityType,
		Action:       action,
		Description:  description,
		ProjectID:    projectID,
		MilestoneID:  milestoneID,
		OrderID:      orderID,
		TradeID:      tradeID,
		Metadata:     metadata,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// 데이터베이스에 저장
	db := database.GetDB()
	if err := db.Create(&activityLog).Error; err != nil {
		log.Printf("❌ 활동 로그 저장 실패: %v", err)
		return fmt.Errorf("failed to save activity log: %w", err)
	}

	log.Printf("✅ 활동 로그 저장 성공 (ID: %d, Type: %s, Action: %s, UserID: %d)",
		activityLog.ID, activityLog.ActivityType, activityLog.Action, activityLog.UserID)

	return nil
}

// StartActivityWorker 활동 로그 큐 워커 시작
func (h *ActivityHandler) StartActivityWorker(ctx context.Context) error {
	queueName := "activity_logs"
	consumerGroup := "activity_workers"
	consumerName := "worker-1"

	log.Printf("📝 활동 로그 워커 시작 (큐: %s)", queueName)

	// Consumer Group 생성 (이미 존재하면 무시)
	client := redis.GetClient()
	_, err := client.XGroupCreateMkStream(context.Background(), queueName, consumerGroup, "0").Result()
	if err != nil && err.Error() != "BUSYGROUP Consumer Group name already exists" {
		log.Printf("⚠️ Consumer Group 생성 실패 (무시하고 계속): %v", err)
	} else {
		log.Printf("✅ Consumer Group 생성 또는 확인됨: %s", consumerGroup)
	}

	for {
		// Context 취소 확인
		select {
		case <-ctx.Done():
			log.Printf("📝 Activity worker gracefully shutting down...")
			return nil
		default:
		}

		// Redis Stream에서 메시지 읽기
		result, err := client.XReadGroup(ctx, &redislib.XReadGroupArgs{
			Group:    consumerGroup,
			Consumer: consumerName,
			Streams:  []string{queueName, ">"},
			Count:    1,
			Block:    time.Second * 5,
		}).Result()

		if err != nil {
			// Context가 취소된 경우
			if err == context.Canceled {
				log.Printf("📝 Activity worker context cancelled, shutting down...")
				return nil
			}
			if err.Error() == "redis: nil" {
				continue // 타임아웃, 계속 대기
			}
			log.Printf("❌ 큐 읽기 오류: %v", err)
			time.Sleep(time.Second * 5)
			continue
		}

		// 메시지 처리
		for _, stream := range result {
			for _, message := range stream.Messages {
				if err := h.processActivityMessage(message); err != nil {
					log.Printf("❌ 활동 로그 메시지 처리 실패: %v", err)
				} else {
					// 메시지 처리 완료 확인
					client.XAck(ctx, queueName, consumerGroup, message.ID)
				}
			}
		}
	}
}

// processActivityMessage 개별 활동 로그 메시지 처리
func (h *ActivityHandler) processActivityMessage(message redislib.XMessage) error {
	log.Printf("📝 활동 로그 메시지 처리: %s", message.ID)

	// job_data 필드에서 JSON 데이터 추출
	jobDataStr, exists := message.Values["job_data"].(string)
	if !exists {
		return fmt.Errorf("job_data field not found")
	}

	// JSON 파싱
	var jobData map[string]interface{}
	if err := json.Unmarshal([]byte(jobDataStr), &jobData); err != nil {
		return fmt.Errorf("failed to parse job data: %w", err)
	}

	// 활동 로그 처리
	return h.HandleActivityLogJob(jobData)
}
