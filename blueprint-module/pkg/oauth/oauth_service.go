package oauth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
	"time"

	"blueprint-module/pkg/config"
)

// OAuthService OAuth 서비스 매니저
type OAuthService struct {
	factory   *ProviderFactory
	config    config.OAuthConfig
	stateKeys map[string]StateInfo // state 토큰 저장 (실제로는 Redis 사용 권장)
}

// StateInfo state 토큰 정보
type StateInfo struct {
	UserID     uint      `json:"user_id"`
	Provider   string    `json:"provider"`
	Action     string    `json:"action"` // "login" or "connect"
	CreatedAt  time.Time `json:"created_at"`
	ExpiresAt  time.Time `json:"expires_at"`
}

// NewOAuthService OAuth 서비스 생성
func NewOAuthService(cfg config.OAuthConfig) *OAuthService {
	service := &OAuthService{
		factory:   NewProviderFactory(),
		config:    cfg,
		stateKeys: make(map[string]StateInfo),
	}

	// 지원되는 제공업체들 등록
	service.registerProviders()

	return service
}

// registerProviders 제공업체들 등록
func (s *OAuthService) registerProviders() {
	// LinkedIn 제공업체 등록
	if s.config.LinkedIn.ClientID != "" {
		linkedinProvider := NewLinkedInProvider(s.config.LinkedIn)
		s.factory.Register("linkedin", linkedinProvider)
	}

	// TODO: 다른 제공업체들도 여기에 추가
	// s.factory.Register("github", NewGitHubProvider(s.config.GitHub))
	// s.factory.Register("twitter", NewTwitterProvider(s.config.Twitter))
}

// GetAuthURL 인증 URL 생성 (연결용)
func (s *OAuthService) GetAuthURL(providerName string, userID uint, action string) (string, error) {
	provider, err := s.factory.Get(providerName)
	if err != nil {
		return "", err
	}

	if err := provider.ValidateConfig(); err != nil {
		return "", fmt.Errorf("provider config invalid: %w", err)
	}

	// state 토큰 생성
	state := s.generateState(userID, providerName, action)

	// 인증 URL 생성
	authURL := provider.GetAuthURL(state)

	return authURL, nil
}

// HandleCallback OAuth 콜백 처리
func (s *OAuthService) HandleCallback(ctx context.Context, providerName, code, state string) (*CallbackResult, error) {
	// state 토큰 검증
	stateInfo, err := s.validateState(state)
	if err != nil {
		return nil, fmt.Errorf("invalid state: %w", err)
	}

	// 제공업체 조회
	provider, err := s.factory.Get(providerName)
	if err != nil {
		return nil, err
	}

	// authorization code를 access token으로 교환
	tokenResp, err := provider.ExchangeCode(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}

	// 사용자 프로필 정보 조회
	profile, err := provider.GetUserProfile(ctx, tokenResp.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to get user profile: %w", err)
	}

	result := &CallbackResult{
		UserID:      stateInfo.UserID,
		Provider:    providerName,
		Action:      stateInfo.Action,
		Profile:     profile,
		TokenInfo:   tokenResp,
		IsNewUser:   false, // 연결 기능이므로 항상 기존 사용자
	}

	// state 토큰 삭제 (사용 완료)
	delete(s.stateKeys, state)

	return result, nil
}

// CallbackResult 콜백 처리 결과
type CallbackResult struct {
	UserID      uint          `json:"user_id"`
	Provider    string        `json:"provider"`
	Action      string        `json:"action"`
	Profile     *UserProfile  `json:"profile"`
	TokenInfo   *TokenResponse `json:"token_info"`
	IsNewUser   bool          `json:"is_new_user"`
}

// GetSupportedProviders 지원되는 제공업체 목록
func (s *OAuthService) GetSupportedProviders() []string {
	return s.factory.GetSupportedProviders()
}

// generateState state 토큰 생성
func (s *OAuthService) generateState(userID uint, provider, action string) string {
	// 랜덤 문자열 생성
	randomBytes := make([]byte, 16)
	rand.Read(randomBytes)
	randomStr := base64.URLEncoding.EncodeToString(randomBytes)

	// state 정보 생성
	now := time.Now()
	stateInfo := StateInfo{
		UserID:    userID,
		Provider:  provider,
		Action:    action,
		CreatedAt: now,
		ExpiresAt: now.Add(10 * time.Minute), // 10분 후 만료
	}

	// state 토큰 형식: userID:provider:action:timestamp:random
	stateToken := fmt.Sprintf("%d:%s:%s:%d:%s",
		userID, provider, action, now.Unix(), randomStr)

	// base64 인코딩
	encodedState := base64.URLEncoding.EncodeToString([]byte(stateToken))

	// 메모리에 저장 (실제로는 Redis 사용 권장)
	s.stateKeys[encodedState] = stateInfo

	return encodedState
}

// validateState state 토큰 검증
func (s *OAuthService) validateState(stateToken string) (*StateInfo, error) {
	// state 토큰 조회
	stateInfo, exists := s.stateKeys[stateToken]
	if !exists {
		return nil, fmt.Errorf("state token not found")
	}

	// 만료 시간 확인
	if time.Now().After(stateInfo.ExpiresAt) {
		delete(s.stateKeys, stateToken) // 만료된 토큰 삭제
		return nil, fmt.Errorf("state token expired")
	}

	// 토큰 디코딩 및 검증
	decodedBytes, err := base64.URLEncoding.DecodeString(stateToken)
	if err != nil {
		return nil, fmt.Errorf("invalid state format")
	}

	parts := strings.Split(string(decodedBytes), ":")
	if len(parts) != 5 {
		return nil, fmt.Errorf("invalid state format")
	}

	// userID 검증
	userID, err := strconv.ParseUint(parts[0], 10, 32)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID in state")
	}

	if uint(userID) != stateInfo.UserID {
		return nil, fmt.Errorf("user ID mismatch")
	}

	// provider, action 검증
	if parts[1] != stateInfo.Provider || parts[2] != stateInfo.Action {
		return nil, fmt.Errorf("provider or action mismatch")
	}

	return &stateInfo, nil
}

// CleanupExpiredStates 만료된 state 토큰들 정리 (주기적으로 호출)
func (s *OAuthService) CleanupExpiredStates() {
	now := time.Now()
	for token, stateInfo := range s.stateKeys {
		if now.After(stateInfo.ExpiresAt) {
			delete(s.stateKeys, token)
		}
	}
}
