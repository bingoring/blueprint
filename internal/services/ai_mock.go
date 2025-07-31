package services

import (
	"context"
	"fmt"
	"time"
)

// MockModel 개발/테스트용 Mock AI 모델
type MockModel struct {
	config MockConfig
}

// MockConfig Mock 모델 설정
type MockConfig struct {
	ResponseDelay time.Duration // 응답 지연 시뮬레이션
	FailRate      float64       // 실패율 (0.0-1.0)
}

// NewMockModel Mock 모델 생성자
func NewMockModel(config MockConfig) *MockModel {
	return &MockModel{
		config: config,
	}
}

// GenerateMilestones Mock 마일스톤 생성
func (m *MockModel) GenerateMilestones(ctx context.Context, request AIRequest) (*AIResponse, error) {
	startTime := time.Now()

	// 응답 지연 시뮬레이션
	if m.config.ResponseDelay > 0 {
		time.Sleep(m.config.ResponseDelay)
	}

	// 실패율 시뮬레이션
	if m.config.FailRate > 0 && time.Now().UnixNano()%100 < int64(m.config.FailRate*100) {
		return nil, fmt.Errorf("Mock API 실패 시뮬레이션 (실패율: %.1f%%)", m.config.FailRate*100)
	}

	milestones := m.generateMockMilestones(request)

	response := &AIResponse{
		Milestones: milestones,
		Tips:       m.generateMockTips(),
		Warnings:   m.generateMockWarnings(),
		Metadata: AIMetadata{
			Provider:     ProviderMock,
			Model:        "mock-v1",
			ResponseTime: time.Since(startTime).Milliseconds(),
			TokensUsed:   len(request.Title) + len(request.Description), // 간단한 토큰 계산
			RequestID:    fmt.Sprintf("mock-%d", time.Now().UnixNano()),
			GeneratedAt:  time.Now().Format(time.RFC3339),
		},
	}

	return response, nil
}

// ValidateConnection Mock API 연결 확인 (항상 성공)
func (m *MockModel) ValidateConnection(ctx context.Context) error {
	// Mock은 항상 연결 성공
	return nil
}

// GetProviderInfo Mock 제공업체 정보 반환
func (m *MockModel) GetProviderInfo() AIProviderInfo {
	return AIProviderInfo{
		Name:        "Mock AI",
		Provider:    ProviderMock,
		Model:       "mock-v1",
		Description: "개발 및 테스트용 Mock AI 모델",
		Features: []string{
			"빠른 응답",
			"실패 시뮬레이션",
			"카테고리별 맞춤형 데이터",
			"완전 오프라인",
		},
		Limits: AILimits{
			MaxTokens:            10000,
			MaxRequestsPerMinute: 1000,
			MaxRequestsPerDay:    999999,
		},
	}
}

// generateMockMilestones 카테고리별 Mock 마일스톤 생성
func (m *MockModel) generateMockMilestones(request AIRequest) []AIMilestone {
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
	if categoryMilestones[request.Category] != nil {
		milestones = categoryMilestones[request.Category]
	}

	return milestones
}

// generateMockTips Mock 팁 생성
func (m *MockModel) generateMockTips() []string {
	return []string{
		"작은 목표부터 시작하여 성취감을 느끼며 동기를 유지하세요",
		"정기적으로 진행 상황을 점검하고 기록하는 습관을 만드세요",
		"혼자보다는 동료나 멘토와 함께 하면 성공 확률이 높아집니다",
		"완벽함보다는 꾸준함이 더 중요합니다",
		"실패를 두려워하지 말고 실패에서 배우는 자세를 가지세요",
	}
}

// generateMockWarnings Mock 주의사항 생성
func (m *MockModel) generateMockWarnings() []string {
	return []string{
		"너무 많은 목표를 동시에 추진하면 집중력이 분산될 수 있습니다",
		"단기간에 큰 변화를 기대하면 실망할 수 있으니 인내심을 가지세요",
		"외부 요인에 의해 계획이 변경될 수 있으니 유연성을 유지하세요",
		"번아웃을 방지하기 위해 적절한 휴식과 재충전 시간을 확보하세요",
	}
}
