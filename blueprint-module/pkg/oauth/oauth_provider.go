package oauth

import (
	"context"
	"fmt"
	"net/url"
)

// OAuthProvider 인터페이스 - 모든 OAuth 제공업체가 구현해야 하는 메서드들
type OAuthProvider interface {
	// GetAuthURL 인증 URL 생성
	GetAuthURL(state string) string

	// ExchangeCode authorization code를 access token으로 교환
	ExchangeCode(ctx context.Context, code string) (*TokenResponse, error)

	// GetUserProfile access token으로 사용자 프로필 정보 조회
	GetUserProfile(ctx context.Context, accessToken string) (*UserProfile, error)

	// GetProviderName 제공업체 이름 반환
	GetProviderName() string

	// ValidateConfig 설정 유효성 검사
	ValidateConfig() error
}

// TokenResponse OAuth token 교환 응답
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
	Scope        string `json:"scope,omitempty"`
}

// UserProfile 사용자 프로필 정보 (표준화된 형태)
type UserProfile struct {
	ID          string `json:"id"`
	Email       string `json:"email"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	DisplayName string `json:"display_name"`
	ProfileURL  string `json:"profile_url,omitempty"`
	Avatar      string `json:"avatar,omitempty"`
	Provider    string `json:"provider"`
	RawData     map[string]interface{} `json:"raw_data,omitempty"`
}

// ProviderFactory OAuth 제공업체 팩토리
type ProviderFactory struct {
	providers map[string]OAuthProvider
}

// NewProviderFactory 팩토리 생성
func NewProviderFactory() *ProviderFactory {
	return &ProviderFactory{
		providers: make(map[string]OAuthProvider),
	}
}

// Register 제공업체 등록
func (f *ProviderFactory) Register(name string, provider OAuthProvider) {
	f.providers[name] = provider
}

// Get 제공업체 조회
func (f *ProviderFactory) Get(name string) (OAuthProvider, error) {
	provider, exists := f.providers[name]
	if !exists {
		return nil, fmt.Errorf("oauth provider '%s' not found", name)
	}
	return provider, nil
}

// GetSupportedProviders 지원되는 제공업체 목록
func (f *ProviderFactory) GetSupportedProviders() []string {
	var names []string
	for name := range f.providers {
		names = append(names, name)
	}
	return names
}

// Helper functions

// BuildURL URL 생성 헬퍼
func BuildURL(baseURL string, params map[string]string) string {
	u, _ := url.Parse(baseURL)
	q := u.Query()
	for key, value := range params {
		q.Set(key, value)
	}
	u.RawQuery = q.Encode()
	return u.String()
}
