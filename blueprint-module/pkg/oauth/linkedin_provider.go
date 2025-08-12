package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"blueprint-module/pkg/config"
)

// LinkedInProvider LinkedIn OAuth 제공업체
type LinkedInProvider struct {
	config config.LinkedInOAuthConfig
	client *http.Client
}

// NewLinkedInProvider LinkedIn 제공업체 생성
func NewLinkedInProvider(cfg config.LinkedInOAuthConfig) *LinkedInProvider {
	return &LinkedInProvider{
		config: cfg,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetProviderName 제공업체 이름 반환
func (p *LinkedInProvider) GetProviderName() string {
	return "linkedin"
}

// ValidateConfig 설정 유효성 검사
func (p *LinkedInProvider) ValidateConfig() error {
	if p.config.ClientID == "" {
		return fmt.Errorf("linkedin client_id is required")
	}
	if p.config.ClientSecret == "" {
		return fmt.Errorf("linkedin client_secret is required")
	}
	if p.config.RedirectURL == "" {
		return fmt.Errorf("linkedin redirect_url is required")
	}
	return nil
}

// GetAuthURL 인증 URL 생성
func (p *LinkedInProvider) GetAuthURL(state string) string {
	scopes := p.config.Scopes
	if scopes == "" {
		scopes = "r_liteprofile r_emailaddress"
	}

	params := map[string]string{
		"response_type": "code",
		"client_id":     p.config.ClientID,
		"redirect_uri":  p.config.RedirectURL,
		"scope":         scopes,
		"state":         state,
	}

	return BuildURL("https://www.linkedin.com/oauth/v2/authorization", params)
}

// ExchangeCode authorization code를 access token으로 교환
func (p *LinkedInProvider) ExchangeCode(ctx context.Context, code string) (*TokenResponse, error) {
	tokenURL := "https://www.linkedin.com/oauth/v2/accessToken"

	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", p.config.RedirectURL)
	data.Set("client_id", p.config.ClientID)
	data.Set("client_secret", p.config.ClientSecret)

	req, err := http.NewRequestWithContext(ctx, "POST", tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("linkedin token exchange failed %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("failed to parse token response: %w", err)
	}

	return &tokenResp, nil
}

// GetUserProfile access token으로 사용자 프로필 정보 조회
func (p *LinkedInProvider) GetUserProfile(ctx context.Context, accessToken string) (*UserProfile, error) {
	// LinkedIn API v2를 사용하여 프로필 정보 조회
	profileURL := "https://api.linkedin.com/v2/people/~?projection=(id,localizedFirstName,localizedLastName,profilePicture(displayImage~:playableStreams))"
	emailURL := "https://api.linkedin.com/v2/emailAddress?q=members&projection=(elements*(handle~))"

	// 프로필 정보 조회
	profile, err := p.fetchLinkedInProfile(ctx, profileURL, accessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch profile: %w", err)
	}

	// 이메일 정보 조회
	email, err := p.fetchLinkedInEmail(ctx, emailURL, accessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch email: %w", err)
	}

	userProfile := &UserProfile{
		ID:          profile.ID,
		Email:       email,
		FirstName:   profile.LocalizedFirstName,
		LastName:    profile.LocalizedLastName,
		DisplayName: fmt.Sprintf("%s %s", profile.LocalizedFirstName, profile.LocalizedLastName),
		ProfileURL:  fmt.Sprintf("https://www.linkedin.com/in/%s", profile.ID),
		Provider:    "linkedin",
		RawData:     map[string]interface{}{"profile": profile},
	}

	// 프로필 이미지 URL 추출
	if profile.ProfilePicture.DisplayImage.Elements != nil && len(profile.ProfilePicture.DisplayImage.Elements) > 0 {
		userProfile.Avatar = profile.ProfilePicture.DisplayImage.Elements[0].Identifiers[0].Identifier
	}

	return userProfile, nil
}

// LinkedIn API 응답 구조체들
type linkedInProfile struct {
	ID                   string `json:"id"`
	LocalizedFirstName   string `json:"localizedFirstName"`
	LocalizedLastName    string `json:"localizedLastName"`
	ProfilePicture       struct {
		DisplayImage struct {
			Elements []struct {
				Identifiers []struct {
					Identifier string `json:"identifier"`
				} `json:"identifiers"`
			} `json:"elements"`
		} `json:"displayImage~"`
	} `json:"profilePicture"`
}

type linkedInEmailResponse struct {
	Elements []struct {
		Handle struct {
			EmailAddress string `json:"emailAddress"`
		} `json:"handle~"`
	} `json:"elements"`
}

// fetchLinkedInProfile LinkedIn 프로필 정보 조회
func (p *LinkedInProvider) fetchLinkedInProfile(ctx context.Context, url, accessToken string) (*linkedInProfile, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("linkedin profile API error %d: %s", resp.StatusCode, string(body))
	}

	var profile linkedInProfile
	if err := json.NewDecoder(resp.Body).Decode(&profile); err != nil {
		return nil, err
	}

	return &profile, nil
}

// fetchLinkedInEmail LinkedIn 이메일 정보 조회
func (p *LinkedInProvider) fetchLinkedInEmail(ctx context.Context, url, accessToken string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("linkedin email API error %d: %s", resp.StatusCode, string(body))
	}

	var emailResp linkedInEmailResponse
	if err := json.NewDecoder(resp.Body).Decode(&emailResp); err != nil {
		return "", err
	}

	if len(emailResp.Elements) == 0 {
		return "", fmt.Errorf("no email found in linkedin response")
	}

	return emailResp.Elements[0].Handle.EmailAddress, nil
}
