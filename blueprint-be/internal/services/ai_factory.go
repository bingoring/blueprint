package services

import (
	"fmt"
	"strconv"
	"time"
)

// DefaultAIModelFactory 기본 AI 모델 팩토리 구현
type DefaultAIModelFactory struct{}

// NewAIModelFactory 팩토리 인스턴스 생성
func NewAIModelFactory() AIModelFactory {
	return &DefaultAIModelFactory{}
}

// CreateModel 지정된 제공업체와 설정으로 AI 모델 생성
func (f *DefaultAIModelFactory) CreateModel(provider AIProvider, config map[string]string) (AIModelInterface, error) {
	switch provider {
	case ProviderOpenAI:
		return f.createOpenAIModel(config)
	case ProviderMock:
		return f.createMockModel(config)
	case ProviderClaude:
		return nil, fmt.Errorf("Claude 모델은 아직 구현되지 않았습니다")
	case ProviderGemini:
		return nil, fmt.Errorf("Gemini 모델은 아직 구현되지 않았습니다")
	default:
		return nil, fmt.Errorf("지원되지 않는 AI 제공업체입니다: %s", provider)
	}
}

// GetSupportedProviders 지원되는 제공업체 목록 반환
func (f *DefaultAIModelFactory) GetSupportedProviders() []AIProvider {
	return []AIProvider{
		ProviderOpenAI,
		ProviderMock,
		// ProviderClaude,  // 향후 구현 예정
		// ProviderGemini,  // 향후 구현 예정
	}
}

// createOpenAIModel OpenAI 모델 생성
func (f *DefaultAIModelFactory) createOpenAIModel(config map[string]string) (AIModelInterface, error) {
	apiKey := config["api_key"]
	if apiKey == "" {
		return nil, fmt.Errorf("OpenAI API 키가 필요합니다")
	}

	model := config["model"]
	if model == "" {
		model = "gpt-4o-mini" // 기본 모델
	}

	openaiConfig := OpenAIConfig{
		APIKey: apiKey,
		Model:  model,
	}

	return NewOpenAIModel(openaiConfig), nil
}

// createMockModel Mock 모델 생성
func (f *DefaultAIModelFactory) createMockModel(config map[string]string) (AIModelInterface, error) {
	mockConfig := MockConfig{
		ResponseDelay: 0,
		FailRate:      0.0,
	}

	// 응답 지연 설정
	if delayStr := config["response_delay"]; delayStr != "" {
		if delayMs, err := strconv.Atoi(delayStr); err == nil {
			mockConfig.ResponseDelay = time.Duration(delayMs) * time.Millisecond
		}
	}

	// 실패율 설정
	if failRateStr := config["fail_rate"]; failRateStr != "" {
		if failRate, err := strconv.ParseFloat(failRateStr, 64); err == nil {
			mockConfig.FailRate = failRate
		}
	}

	return NewMockModel(mockConfig), nil
}

// AI 모델 설정을 위한 헬퍼 함수들

// CreateOpenAIConfig OpenAI 설정 생성
func CreateOpenAIConfig(apiKey, model string) map[string]string {
	config := map[string]string{
		"api_key": apiKey,
	}
	if model != "" {
		config["model"] = model
	}
	return config
}

// CreateMockConfig Mock 설정 생성
func CreateMockConfig(responseDelayMs int, failRate float64) map[string]string {
	return map[string]string{
		"response_delay": strconv.Itoa(responseDelayMs),
		"fail_rate":      strconv.FormatFloat(failRate, 'f', 2, 64),
	}
}

// 환경변수로부터 설정 생성
func CreateConfigFromEnv(provider AIProvider, apiKey, model string) map[string]string {
	switch provider {
	case ProviderOpenAI:
		return CreateOpenAIConfig(apiKey, model)
	case ProviderMock:
		return CreateMockConfig(0, 0.0) // 기본값
	default:
		return make(map[string]string)
	}
}
