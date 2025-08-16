package services

import (
	"blueprint-module/pkg/models"
	"blueprint/internal/config"
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
	client := openai.NewClient(cfg.AI.OpenAI.APIKey)
	return &AIService{
		client: client,
		config: cfg,
		db:     db,
	}
}

// AI가 제안하는 마일스톤 구조
type AIMilestoneResponse struct {
	Milestones []AIMilestone `json:"milestones"`
	Tips       []string      `json:"tips"`     // 추가 팁
	Warnings   []string      `json:"warnings"` // 주의사항
}

type AIMilestone struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Order       int    `json:"order"`
	Duration    string `json:"duration"`   // 예상 소요 기간 (예: "2-3개월")
	Difficulty  string `json:"difficulty"` // 난이도 (쉬움/보통/어려움)
	Category    string `json:"category"`   // 카테고리 (준비/실행/완성 등)
}

// GenerateMilestones AI를 사용해서 마일스톤을 생성합니다 🤖
func (s *AIService) GenerateMilestones(dream models.CreateProjectRequest) (*AIMilestoneResponse, error) {
	// OpenAI API 호출 시도
	aiResponse, err := s.generateMilestonesWithOpenAI(dream)
	if err != nil {
		// OpenAI API 실패 시 Mock 데이터 반환
		fmt.Printf("⚠️ OpenAI API 실패, Mock 데이터 사용: %v\n", err)
		return s.generateMockMilestones(dream), nil
	}

	return aiResponse, nil
}

// generateMilestonesWithOpenAI OpenAI API를 사용하여 마일스톤 생성
func (s *AIService) generateMilestonesWithOpenAI(dream models.CreateProjectRequest) (*AIMilestoneResponse, error) {
	prompt := s.buildMilestonePrompt(dream)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req := openai.ChatCompletionRequest{
		Model: s.config.AI.OpenAI.Model,
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

// generateMockMilestones 카테고리별 Mock 마일스톤 데이터 생성 🎭
func (s *AIService) generateMockMilestones(dream models.CreateProjectRequest) *AIMilestoneResponse {
	categoryMilestones := map[string][]AIMilestone{
		"career": {
			{Title: "현재 스킬 분석 및 부족한 부분 파악", Description: "현재 보유한 기술과 목표 직무에 필요한 기술을 비교 분석하여 학습 로드맵을 세워보세요.", Order: 1, Duration: "2-3주", Difficulty: "쉬움"},
			{Title: "관련 자격증 취득 또는 교육 과정 수료", Description: "목표 분야의 핵심 자격증을 취득하거나 온라인/오프라인 교육을 통해 전문성을 높여보세요.", Order: 2, Duration: "3-6개월", Difficulty: "보통"},
			{Title: "포트폴리오 및 이력서 업데이트", Description: "새로 습득한 기술과 경험을 바탕으로 경쟁력 있는 포트폴리오와 이력서를 작성해보세요.", Order: 3, Duration: "2-4주", Difficulty: "보통"},
			{Title: "네트워킹 및 업계 정보 수집", Description: "해당 분야 전문가들과 네트워킹을 형성하고 업계 트렌드와 채용 정보를 지속적으로 수집하세요.", Order: 4, Duration: "지속적", Difficulty: "어려움"},
		},
		"business": {
			{Title: "사업 아이디어 구체화 및 시장 조사", Description: "사업 아이디어를 명확히 하고 타겟 시장의 규모, 경쟁사, 고객 니즈를 철저히 분석해보세요.", Order: 1, Duration: "1-2개월", Difficulty: "보통"},
			{Title: "사업계획서 작성 및 자금 계획 수립", Description: "상세한 사업계획서를 작성하고 초기 운영자금, 투자 유치 계획을 구체적으로 세워보세요.", Order: 2, Duration: "1-2개월", Difficulty: "어려움"},
			{Title: "법인 설립 및 필요 허가 취득", Description: "사업자 등록, 법인 설립, 업종별 필요한 허가나 신고를 완료하여 합법적인 사업 기반을 마련하세요.", Order: 3, Duration: "2-4주", Difficulty: "보통"},
			{Title: "MVP 개발 및 테스트 마케팅", Description: "최소기능제품(MVP)을 개발하고 소규모 테스트 마케팅을 통해 시장 반응을 확인해보세요.", Order: 4, Duration: "3-6개월", Difficulty: "어려움"},
		},
		"education": {
			{Title: "학습 목표 및 커리큘럼 설정", Description: "명확한 학습 목표를 설정하고 단계별 커리큘럼을 계획하여 체계적인 학습 방향을 정해보세요.", Order: 1, Duration: "1-2주", Difficulty: "쉬움"},
			{Title: "기초 이론 학습 및 이해", Description: "해당 분야의 핵심 이론과 기초 개념을 탄탄히 익혀 향후 심화 학습의 토대를 만들어보세요.", Order: 2, Duration: "2-4개월", Difficulty: "보통"},
			{Title: "실습 프로젝트 진행", Description: "이론으로 학습한 내용을 실제 프로젝트나 과제를 통해 적용하며 실무 경험을 쌓아보세요.", Order: 3, Duration: "2-3개월", Difficulty: "보통"},
			{Title: "성과 평가 및 자격 취득", Description: "학습 성과를 객관적으로 평가받고 관련 자격증이나 수료증을 취득하여 실력을 인증받으세요.", Order: 4, Duration: "1-2개월", Difficulty: "어려움"},
		},
		"personal": {
			{Title: "현재 상태 분석 및 목표 구체화", Description: "현재 자신의 상태를 객관적으로 분석하고 달성하고자 하는 구체적인 목표를 명확히 설정해보세요.", Order: 1, Duration: "1-2주", Difficulty: "쉬움"},
			{Title: "단계별 실행 계획 수립", Description: "목표 달성을 위한 구체적이고 실현 가능한 단계별 계획을 수립하고 일일/주간 실행 스케줄을 만들어보세요.", Order: 2, Duration: "1-2주", Difficulty: "보통"},
			{Title: "꾸준한 실행 및 습관 형성", Description: "계획한 내용을 꾸준히 실행하며 목표 달성에 필요한 긍정적인 습관을 형성해나가세요.", Order: 3, Duration: "3-6개월", Difficulty: "어려움"},
			{Title: "진행 상황 점검 및 조정", Description: "정기적으로 진행 상황을 점검하고 필요시 계획을 조정하여 목표 달성 확률을 높여보세요.", Order: 4, Duration: "지속적", Difficulty: "보통"},
		},
		"life": {
			{Title: "생활 패턴 분석 및 개선점 파악", Description: "현재 생활 패턴을 분석하고 목표 달성을 위해 개선해야 할 부분들을 구체적으로 파악해보세요.", Order: 1, Duration: "1-2주", Difficulty: "쉬움"},
			{Title: "환경 조성 및 준비 작업", Description: "목표 달성에 필요한 물리적, 정신적 환경을 조성하고 필요한 도구나 자원을 준비해보세요.", Order: 2, Duration: "2-4주", Difficulty: "보통"},
			{Title: "점진적 변화 실행", Description: "급격한 변화보다는 점진적이고 지속 가능한 변화를 실행하여 안정적인 라이프스타일을 구축해보세요.", Order: 3, Duration: "3-6개월", Difficulty: "보통"},
			{Title: "새로운 라이프스타일 정착", Description: "변화된 생활 방식이 자연스럽게 몸에 베도록 하고 장기적으로 유지할 수 있는 시스템을 만들어보세요.", Order: 4, Duration: "6-12개월", Difficulty: "어려움"},
		},
	}

	// 카테고리에 맞는 마일스톤 선택 (기본값: career)
	milestones := categoryMilestones["career"]
	if categoryMilestones[string(dream.Category)] != nil {
		milestones = categoryMilestones[string(dream.Category)]
	}

	// 일반적인 팁과 주의사항
	tips := []string{
		"작은 목표부터 시작하여 성취감을 느끼며 동기를 유지하세요",
		"정기적으로 진행 상황을 점검하고 기록하는 습관을 만드세요",
		"혼자보다는 동료나 멘토와 함께 하면 성공 확률이 높아집니다",
		"완벽함보다는 꾸준함이 더 중요합니다",
		"실패를 두려워하지 말고 실패에서 배우는 자세를 가지세요",
	}

	warnings := []string{
		"너무 많은 목표를 동시에 추진하면 집중력이 분산될 수 있습니다",
		"단기간에 큰 변화를 기대하면 실망할 수 있으니 인내심을 가지세요",
		"외부 요인에 의해 계획이 변경될 수 있으니 유연성을 유지하세요",
		"번아웃을 방지하기 위해 적절한 휴식과 재충전 시간을 확보하세요",
	}

	return &AIMilestoneResponse{
		Milestones: milestones,
		Tips:       tips,
		Warnings:   warnings,
	}
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
func (s *AIService) buildMilestonePrompt(dream models.CreateProjectRequest) string {
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
	if s.config.AI.OpenAI.APIKey == "" {
		return fmt.Errorf("OpenAI API 키가 설정되지 않았습니다")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 간단한 테스트 요청
	req := openai.ChatCompletionRequest{
		Model: s.config.AI.OpenAI.Model,
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
