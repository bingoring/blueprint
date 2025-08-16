package services

import (
	"context"
)

// AIProvider AI 제공업체 타입
type AIProvider string

const (
	ProviderOpenAI AIProvider = "openai"
	ProviderClaude AIProvider = "claude"
	ProviderGemini AIProvider = "gemini"
	ProviderMock   AIProvider = "mock" // 개발/테스트용
)

// AIModelInterface 모든 AI 모델이 구현해야 하는 인터페이스
type AIModelInterface interface {
	// GenerateMilestones 마일스톤 생성
	GenerateMilestones(ctx context.Context, request AIRequest) (*AIResponse, error)

	// ValidateConnection API 연결 상태 확인
	ValidateConnection(ctx context.Context) error

	// GetProviderInfo 제공업체 정보 반환
	GetProviderInfo() AIProviderInfo
}

// AIRequest 모든 AI 모델에 공통으로 사용되는 요청 구조
type AIRequest struct {
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Category    string            `json:"category"`
	TargetDate  string            `json:"target_date,omitempty"`
	Budget      int64             `json:"budget"`
	Priority    int               `json:"priority"`
	Tags        []string          `json:"tags,omitempty"`
	Context     map[string]string `json:"context,omitempty"` // 추가 컨텍스트
}

// AIResponse 모든 AI 모델에서 반환하는 공통 응답 구조
type AIResponse struct {
	Milestones []AIMilestone `json:"milestones"`
	Tips       []string      `json:"tips"`
	Warnings   []string      `json:"warnings"`
	Metadata   AIMetadata    `json:"metadata"`
}

// AIMetadata AI 응답에 대한 메타데이터
type AIMetadata struct {
	Provider     AIProvider `json:"provider"`
	Model        string     `json:"model"`
	ResponseTime int64      `json:"response_time_ms"`
	TokensUsed   int        `json:"tokens_used,omitempty"`
	RequestID    string     `json:"request_id,omitempty"`
	GeneratedAt  string     `json:"generated_at"`
}

// AIProviderInfo AI 제공업체 정보
type AIProviderInfo struct {
	Name        string     `json:"name"`
	Provider    AIProvider `json:"provider"`
	Model       string     `json:"model"`
	Description string     `json:"description"`
	Features    []string   `json:"features"`
	Limits      AILimits   `json:"limits"`
}

// AILimits AI 모델의 제한사항
type AILimits struct {
	MaxTokens            int `json:"max_tokens"`
	MaxRequestsPerMinute int `json:"max_requests_per_minute"`
	MaxRequestsPerDay    int `json:"max_requests_per_day,omitempty"`
}

// AIModelFactory AI 모델 팩토리 인터페이스
type AIModelFactory interface {
	CreateModel(provider AIProvider, config map[string]string) (AIModelInterface, error)
	GetSupportedProviders() []AIProvider
}
