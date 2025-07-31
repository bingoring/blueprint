package services

import (
	"blueprint/internal/config"
	"blueprint/internal/models"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/sashabaranov/go-openai"
	"gorm.io/gorm"
)

type AIService struct {
	client *openai.Client
	config *config.Config
	db     *gorm.DB
}

func NewAIService(cfg *config.Config, db *gorm.DB) *AIService {
	client := openai.NewClient(cfg.OpenAI.APIKey)
	return &AIService{
		client: client,
		config: cfg,
		db:     db,
	}
}

// AI가 제안하는 마일스톤 구조
type AIMilestoneResponse struct {
	Milestones []AIMilestone `json:"milestones"`
	Tips       []string      `json:"tips"`       // 추가 팁
	Warnings   []string      `json:"warnings"`   // 주의사항
}

type AIMilestone struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Order       int    `json:"order"`
	Duration    string `json:"duration"`    // 예상 소요 기간 (예: "2-3개월")
	Difficulty  string `json:"difficulty"`  // 난이도 (쉬움/보통/어려움)
	Category    string `json:"category"`    // 카테고리 (준비/실행/완성 등)
}

// GenerateMilestones AI를 사용해서 마일스톤을 생성합니다 🤖
func (s *AIService) GenerateMilestones(dream models.CreateGoalRequest) (*AIMilestoneResponse, error) {
	prompt := s.buildMilestonePrompt(dream)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req := openai.ChatCompletionRequest{
		Model: s.config.OpenAI.Model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: s.getSystemPrompt(),
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
		Temperature: 0.7, // 창의적이지만 일관성 있는 응답
		MaxTokens:   2000,
	}

	resp, err := s.client.CreateChatCompletion(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("OpenAI API 호출 실패: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("AI 응답이 비어있습니다")
	}

	// JSON 응답 파싱
	var aiResponse AIMilestoneResponse
	if err := json.Unmarshal([]byte(resp.Choices[0].Message.Content), &aiResponse); err != nil {
		return nil, fmt.Errorf("AI 응답 파싱 실패: %w", err)
	}

	// 마일스톤 순서 정렬
	for i := range aiResponse.Milestones {
		aiResponse.Milestones[i].Order = i + 1
	}

	return &aiResponse, nil
}

// 시스템 프롬프트 - AI의 역할과 응답 형식을 정의
func (s *AIService) getSystemPrompt() string {
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

// 사용자 꿈 정보를 바탕으로 프롬프트 생성
func (s *AIService) buildMilestonePrompt(dream models.CreateGoalRequest) string {
	categoryNames := map[string]string{
		"career":    "커리어 성장",
		"business":  "창업/사업",
		"education": "교육/학습",
		"personal":  "개인 발전",
		"life":      "라이프스타일",
	}

	categoryName := categoryNames[string(dream.Category)]
	if categoryName == "" {
		categoryName = string(dream.Category)
	}

	prompt := fmt.Sprintf(`꿈 분석 요청:

제목: %s
설명: %s
카테고리: %s
예산: %d만원
우선순위: %d/5`,
		dream.Title,
		dream.Description,
		categoryName,
		dream.Budget,
		dream.Priority,
	)

	// 목표 날짜가 있는 경우 추가
	if dream.TargetDate != nil {
		prompt += fmt.Sprintf("\n목표 날짜: %s", dream.TargetDate.Format("2006년 1월 2일"))
	}

	// 태그가 있는 경우 추가
	if len(dream.Tags) > 0 {
		tagsStr := ""
		for i, tag := range dream.Tags {
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

// ValidateAPIKey OpenAI API 키가 유효한지 확인
func (s *AIService) ValidateAPIKey() error {
	if s.config.OpenAI.APIKey == "" {
		return fmt.Errorf("OpenAI API 키가 설정되지 않았습니다")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 간단한 테스트 요청
	req := openai.ChatCompletionRequest{
		Model: s.config.OpenAI.Model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: "테스트",
			},
		},
		MaxTokens: 10,
	}

	_, err := s.client.CreateChatCompletion(ctx, req)
	return err
}

// CheckAIUsageLimit 사용자의 AI 사용 횟수를 체크합니다 🚫
func (s *AIService) CheckAIUsageLimit(userID uint) (bool, int, error) {
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
func (s *AIService) IncrementAIUsage(userID uint) error {
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
func (s *AIService) GetAIUsageInfo(userID uint) (*AIUsageInfo, error) {
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

// AI 사용 정보 구조체
type AIUsageInfo struct {
	Used      int  `json:"used"`      // 사용한 횟수
	Limit     int  `json:"limit"`     // 최대 사용 가능 횟수
	Remaining int  `json:"remaining"` // 남은 횟수
	CanUse    bool `json:"can_use"`   // 사용 가능 여부
}
