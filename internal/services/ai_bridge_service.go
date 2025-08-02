package services

import (
	"blueprint/internal/config"
	"blueprint/internal/models"
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// BridgeAIService 브릿지 패턴을 적용한 AI 서비스
type BridgeAIService struct {
	aiModel   AIModelInterface
	factory   AIModelFactory
	provider  AIProvider
	config    *config.Config
	db        *gorm.DB
}

// NewBridgeAIService 새로운 브릿지 AI 서비스 생성
func NewBridgeAIService(cfg *config.Config, db *gorm.DB) *BridgeAIService {
	factory := NewAIModelFactory()

		// 환경변수에서 설정된 AI 제공업체 사용
	provider := AIProvider(cfg.AI.Provider)
	var modelConfig map[string]string

	switch provider {
	case ProviderOpenAI:
		if cfg.AI.OpenAI.APIKey != "" && cfg.AI.OpenAI.APIKey != "your-openai-api-key" {
			modelConfig = CreateOpenAIConfig(cfg.AI.OpenAI.APIKey, cfg.AI.OpenAI.Model)
		} else {
			// API 키가 없으면 Mock으로 폴백
			provider = ProviderMock
			modelConfig = CreateMockConfig(100, 0.0)
		}
	case ProviderMock:
		modelConfig = CreateMockConfig(100, 0.0) // 100ms 지연, 실패율 0%
	default:
		// 지원되지 않는 제공업체는 Mock으로 폴백
		provider = ProviderMock
		modelConfig = CreateMockConfig(100, 0.0)
	}

	aiModel, err := factory.CreateModel(provider, modelConfig)
	if err != nil {
		// OpenAI 실패 시 Mock으로 폴백
		provider = ProviderMock
		modelConfig = CreateMockConfig(100, 0.0)
		aiModel, _ = factory.CreateModel(provider, modelConfig)
	}

	return &BridgeAIService{
		aiModel:  aiModel,
		factory:  factory,
		provider: provider,
		config:   cfg,
		db:       db,
	}
}

// SwitchProvider AI 제공업체 변경
func (s *BridgeAIService) SwitchProvider(provider AIProvider) error {
	var modelConfig map[string]string

	switch provider {
	case ProviderOpenAI:
		if s.config.AI.OpenAI.APIKey == "" || s.config.AI.OpenAI.APIKey == "your-openai-api-key" {
			return fmt.Errorf("OpenAI API 키가 설정되지 않았습니다")
		}
		modelConfig = CreateOpenAIConfig(s.config.AI.OpenAI.APIKey, s.config.AI.OpenAI.Model)
	case ProviderMock:
		modelConfig = CreateMockConfig(100, 0.0)
	default:
		return fmt.Errorf("지원되지 않는 제공업체입니다: %s", provider)
	}

	aiModel, err := s.factory.CreateModel(provider, modelConfig)
	if err != nil {
		return fmt.Errorf("AI 모델 생성 실패: %w", err)
	}

	// 연결 테스트
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := aiModel.ValidateConnection(ctx); err != nil {
		return fmt.Errorf("AI 모델 연결 실패: %w", err)
	}

	s.aiModel = aiModel
	s.provider = provider

	fmt.Printf("✅ AI 제공업체를 %s로 변경했습니다\n", provider)
	return nil
}

// GetCurrentProvider 현재 사용 중인 제공업체 반환
func (s *BridgeAIService) GetCurrentProvider() AIProvider {
	return s.provider
}

// GetProviderInfo 현재 제공업체 정보 반환
func (s *BridgeAIService) GetProviderInfo() AIProviderInfo {
	return s.aiModel.GetProviderInfo()
}

// GetSupportedProviders 지원되는 제공업체 목록 반환
func (s *BridgeAIService) GetSupportedProviders() []AIProvider {
	return s.factory.GetSupportedProviders()
}

// GenerateMilestones AI를 사용해서 마일스톤을 생성합니다 🤖
func (s *BridgeAIService) GenerateMilestones(project models.CreateProjectRequest) (*AIMilestoneResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// CreateProjectRequest를 AIRequest로 변환
	aiRequest := s.convertToAIRequest(project)

	// AI 모델을 통해 마일스톤 생성
	aiResponse, err := s.aiModel.GenerateMilestones(ctx, aiRequest)
	if err != nil {
		// OpenAI 실패 시 자동으로 Mock으로 전환
		if s.provider == ProviderOpenAI {
			fmt.Printf("⚠️ OpenAI 실패, Mock 모델로 자동 전환: %v\n", err)
			if switchErr := s.SwitchProvider(ProviderMock); switchErr == nil {
				aiResponse, err = s.aiModel.GenerateMilestones(ctx, aiRequest)
			}
		}

		if err != nil {
			return nil, fmt.Errorf("AI 마일스톤 생성 실패: %w", err)
		}
	}

	// AIResponse를 기존 AIMilestoneResponse 형태로 변환 (하위 호환성)
	return s.convertToLegacyResponse(aiResponse), nil
}

// convertToAIRequest CreateProjectRequest를 AIRequest로 변환
func (s *BridgeAIService) convertToAIRequest(project models.CreateProjectRequest) AIRequest {
	var targetDateStr string
	if project.TargetDate != nil {
		targetDateStr = project.TargetDate.Format(time.RFC3339)
	}

	return AIRequest{
		Title:       project.Title,
		Description: project.Description,
		Category:    string(project.Category),
		TargetDate:  targetDateStr,
		Budget:      project.Budget,
		Priority:    project.Priority,
		Tags:        project.Tags,
		Context: map[string]string{
			"provider": string(s.provider),
			"model":    s.aiModel.GetProviderInfo().Model,
		},
	}
}

// convertToLegacyResponse AIResponse를 기존 AIMilestoneResponse로 변환
func (s *BridgeAIService) convertToLegacyResponse(response *AIResponse) *AIMilestoneResponse {
	return &AIMilestoneResponse{
		Milestones: response.Milestones,
		Tips:       response.Tips,
		Warnings:   response.Warnings,
	}
}

// ValidateAPIKey 현재 AI 모델의 연결 상태 확인
func (s *BridgeAIService) ValidateAPIKey() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return s.aiModel.ValidateConnection(ctx)
}

// 기존 AIService 메서드들과의 호환성을 위한 메서드들

// CheckAIUsageLimit 사용자의 AI 사용 횟수를 체크합니다 🚫
func (s *BridgeAIService) CheckAIUsageLimit(userID uint) (bool, int, error) {
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return false, 0, fmt.Errorf("사용자 정보를 찾을 수 없습니다: %w", err)
	}

	canUse := user.AIUsageCount < user.AIUsageLimit
	remaining := user.AIUsageLimit - user.AIUsageCount
	if remaining < 0 {
		remaining = 0
	}

	return canUse, remaining, nil
}

// IncrementAIUsage 사용자의 AI 사용 횟수를 증가시킵니다 📈
func (s *BridgeAIService) IncrementAIUsage(userID uint) error {
	result := s.db.Model(&models.User{}).
		Where("id = ?", userID).
		Update("ai_usage_count", gorm.Expr("ai_usage_count + 1"))

	if result.Error != nil {
		return fmt.Errorf("AI 사용 횟수 업데이트 실패: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("사용자를 찾을 수 없습니다")
	}

	return nil
}

// GetAIUsageInfo 사용자의 AI 사용 정보를 반환합니다 📊
func (s *BridgeAIService) GetAIUsageInfo(userID uint) (*AIUsageInfo, error) {
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return nil, fmt.Errorf("사용자 정보를 찾을 수 없습니다: %w", err)
	}

	remaining := user.AIUsageLimit - user.AIUsageCount
	if remaining < 0 {
		remaining = 0
	}

	return &AIUsageInfo{
		Used:      user.AIUsageCount,
		Limit:     user.AIUsageLimit,
		Remaining: remaining,
		CanUse:    user.AIUsageCount < user.AIUsageLimit,
	}, nil
}
