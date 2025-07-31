package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/sashabaranov/go-openai"
)

// OpenAIModel OpenAI API 구현체
type OpenAIModel struct {
	client *openai.Client
	config OpenAIConfig
}

// OpenAIConfig OpenAI 설정
type OpenAIConfig struct {
	APIKey string
	Model  string
}

// NewOpenAIModel OpenAI 모델 생성자
func NewOpenAIModel(config OpenAIConfig) *OpenAIModel {
	client := openai.NewClient(config.APIKey)
	return &OpenAIModel{
		client: client,
		config: config,
	}
}

// GenerateMilestones OpenAI를 사용하여 마일스톤 생성
func (m *OpenAIModel) GenerateMilestones(ctx context.Context, request AIRequest) (*AIResponse, error) {
	startTime := time.Now()

	prompt := m.buildPrompt(request)

	req := openai.ChatCompletionRequest{
		Model: m.config.Model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: m.getSystemPrompt(),
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
		Temperature: 0.7,
		MaxTokens:   2000,
	}

	resp, err := m.client.CreateChatCompletion(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("OpenAI API 호출 실패: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("OpenAI 응답이 비어있습니다")
	}

	// 기존 AIMilestoneResponse 구조를 파싱
	var legacyResponse AIMilestoneResponse
	if err := json.Unmarshal([]byte(resp.Choices[0].Message.Content), &legacyResponse); err != nil {
		return nil, fmt.Errorf("OpenAI 응답 파싱 실패: %w", err)
	}

	// 마일스톤 순서 정렬
	for i := range legacyResponse.Milestones {
		legacyResponse.Milestones[i].Order = i + 1
	}

	// 새로운 AIResponse 구조로 변환
	response := &AIResponse{
		Milestones: legacyResponse.Milestones,
		Tips:       legacyResponse.Tips,
		Warnings:   legacyResponse.Warnings,
		Metadata: AIMetadata{
			Provider:     ProviderOpenAI,
			Model:        m.config.Model,
			ResponseTime: time.Since(startTime).Milliseconds(),
			TokensUsed:   resp.Usage.TotalTokens,
			RequestID:    resp.ID,
			GeneratedAt:  time.Now().Format(time.RFC3339),
		},
	}

	return response, nil
}

// ValidateConnection OpenAI API 연결 상태 확인
func (m *OpenAIModel) ValidateConnection(ctx context.Context) error {
	req := openai.ChatCompletionRequest{
		Model: m.config.Model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: "테스트",
			},
		},
		MaxTokens: 10,
	}

	_, err := m.client.CreateChatCompletion(ctx, req)
	return err
}

// GetProviderInfo OpenAI 제공업체 정보 반환
func (m *OpenAIModel) GetProviderInfo() AIProviderInfo {
	return AIProviderInfo{
		Name:        "OpenAI",
		Provider:    ProviderOpenAI,
		Model:       m.config.Model,
		Description: "OpenAI의 GPT 모델을 사용한 AI 마일스톤 생성",
		Features: []string{
			"자연어 처리",
			"창의적 제안",
			"단계별 마일스톤",
			"난이도 분석",
		},
		Limits: AILimits{
			MaxTokens:            2000,
			MaxRequestsPerMinute: 60,
			MaxRequestsPerDay:    1000,
		},
	}
}

// buildPrompt 요청을 바탕으로 프롬프트 생성
func (m *OpenAIModel) buildPrompt(request AIRequest) string {
	categoryNames := map[string]string{
		"career":    "커리어 성장",
		"business":  "창업/사업",
		"education": "교육/학습",
		"personal":  "개인 발전",
		"life":      "라이프스타일",
	}

	categoryName := categoryNames[request.Category]
	if categoryName == "" {
		categoryName = request.Category
	}

	prompt := fmt.Sprintf(`꿈 분석 요청:

제목: %s
설명: %s
카테고리: %s
예산: %d만원
우선순위: %d/5`,
		request.Title,
		request.Description,
		categoryName,
		request.Budget,
		request.Priority,
	)

	// 목표 날짜가 있는 경우 추가
	if request.TargetDate != "" {
		if parsedDate, err := time.Parse(time.RFC3339, request.TargetDate); err == nil {
			prompt += fmt.Sprintf("\n목표 날짜: %s", parsedDate.Format("2006년 1월 2일"))
		}
	}

	// 태그가 있는 경우 추가
	if len(request.Tags) > 0 {
		tagsStr := ""
		for i, tag := range request.Tags {
			if i > 0 {
				tagsStr += ", "
			}
			tagsStr += tag
		}
		prompt += fmt.Sprintf("\n관심 분야: %s", tagsStr)
	}

	prompt += "\n\n위 꿈을 실현하기 위한 구체적이고 실행 가능한 마일스톤을 제안해주세요."

	return prompt
}

// getSystemPrompt 시스템 프롬프트 반환
func (m *OpenAIModel) getSystemPrompt() string {
	return `당신은 한국의 전문 라이프 코치이자 목표 달성 전문가입니다.
사용자의 꿈을 분석하여 실현 가능하고 구체적인 마일스톤을 제안해주세요.

응답 규칙:
1. 반드시 JSON 형식으로 응답하세요
2. 마일스톤은 3-5개, 논리적 순서로 배열
3. 각 마일스톤은 구체적인 액션 아이템이어야 함
4. 한국 상황에 맞는 현실적인 제안
5. 예상 기간은 정확하고 실현 가능해야 함

JSON 구조:
{
  "milestones": [
    {
      "title": "구체적인 마일스톤 제목",
      "description": "상세한 실행 방법과 팁",
      "duration": "예상 소요 기간",
      "difficulty": "쉬움|보통|어려움",
      "category": "준비|실행|완성"
    }
  ],
  "tips": ["성공을 위한 추가 팁들"],
  "warnings": ["주의해야 할 점들"]
}`
}
