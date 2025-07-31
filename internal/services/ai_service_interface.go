package services

import "blueprint/internal/models"

// AIServiceInterface AI 서비스의 공통 인터페이스
type AIServiceInterface interface {
	// GenerateMilestones AI를 사용해서 마일스톤을 생성합니다
	GenerateMilestones(dream models.CreateGoalRequest) (*AIMilestoneResponse, error)

	// CheckAIUsageLimit 사용자의 AI 사용 횟수를 체크합니다
	CheckAIUsageLimit(userID uint) (bool, int, error)

	// IncrementAIUsage 사용자의 AI 사용 횟수를 증가시킵니다
	IncrementAIUsage(userID uint) error

	// GetAIUsageInfo 사용자의 AI 사용 정보를 반환합니다
	GetAIUsageInfo(userID uint) (*AIUsageInfo, error)

	// ValidateAPIKey AI API 연결 상태를 확인합니다
	ValidateAPIKey() error
}
