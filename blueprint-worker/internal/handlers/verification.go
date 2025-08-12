package handlers

import (
	"blueprint-module/pkg/queue"
	"blueprint-worker/internal/config"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

type VerificationHandler struct {
	config *config.Config
}

type LinkedInProfile struct {
	ID        string `json:"id"`
	FirstName string `json:"localizedFirstName"`
	LastName  string `json:"localizedLastName"`
	Email     string `json:"emailAddress"`
}

type GitHubProfile struct {
	ID      int    `json:"id"`
	Login   string `json:"login"`
	Name    string `json:"name"`
	Email   string `json:"email"`
	Company string `json:"company"`
}

type TwitterProfile struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Name     string `json:"name"`
	Verified bool   `json:"verified"`
}

func NewVerificationHandler(cfg *config.Config) *VerificationHandler {
	return &VerificationHandler{
		config: cfg,
	}
}

func (h *VerificationHandler) StartVerificationWorker() error {
	log.Println("🔍 Verification worker started")

	return queue.ConsumeJobs("verification_queue", "verification_workers", "verification_worker_1", h.handleVerificationJob)
}

func (h *VerificationHandler) handleVerificationJob(jobData map[string]interface{}) error {
	jobType, ok := jobData["type"].(string)
	if !ok {
		return fmt.Errorf("missing job type")
	}

	switch jobType {
	case "verify_social_provider":
		return h.verifySocialProvider(jobData)
	case "verify_domain":
		return h.verifyDomain(jobData)
	default:
		return fmt.Errorf("unknown verification job type: %s", jobType)
	}
}

func (h *VerificationHandler) verifySocialProvider(jobData map[string]interface{}) error {
	provider, ok := jobData["provider"].(string)
	if !ok {
		return fmt.Errorf("missing provider")
	}

	accessToken, ok := jobData["access_token"].(string)
	if !ok {
		return fmt.Errorf("missing access_token")
	}

	userID := jobData["user_id"]

	switch provider {
	case "linkedin":
		return h.verifyLinkedIn(accessToken, userID)
	case "github":
		return h.verifyGitHub(accessToken, userID)
	case "twitter":
		return h.verifyTwitter(accessToken, userID)
	default:
		return fmt.Errorf("unsupported provider: %s", provider)
	}
}

func (h *VerificationHandler) verifyLinkedIn(accessToken string, userID interface{}) error {
	// LinkedIn API로 프로필 정보 확인
	apiURL := "https://api.linkedin.com/v2/people/~?projection=(id,localizedFirstName,localizedLastName,emailAddress)"

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create LinkedIn request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to call LinkedIn API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("LinkedIn API error %d: %s", resp.StatusCode, string(body))
	}

	// 응답 파싱
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read LinkedIn response: %w", err)
	}

	var profile LinkedInProfile
	if err := json.Unmarshal(body, &profile); err != nil {
		return fmt.Errorf("failed to parse LinkedIn profile: %w", err)
	}

	// 프로필 정보 검증
	if profile.ID == "" {
		return fmt.Errorf("invalid LinkedIn profile")
	}

	log.Printf("✅ LinkedIn verified for user %v: %s %s (%s)", userID, profile.FirstName, profile.LastName, profile.ID)

	// TODO: 데이터베이스 업데이트
	// - user_verification 테이블의 linkedin_connected = true
	// - linkedin_profile_id 저장
	// - 검증 완료 시간 기록

	return nil
}

func (h *VerificationHandler) verifyGitHub(accessToken string, userID interface{}) error {
	// GitHub API로 프로필 정보 확인
	apiURL := "https://api.github.com/user"

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create GitHub request: %w", err)
	}

	req.Header.Set("Authorization", "token "+accessToken)
	req.Header.Set("User-Agent", "Blueprint-App")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to call GitHub API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("GitHub API error %d: %s", resp.StatusCode, string(body))
	}

	// 응답 파싱
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read GitHub response: %w", err)
	}

	var profile GitHubProfile
	if err := json.Unmarshal(body, &profile); err != nil {
		return fmt.Errorf("failed to parse GitHub profile: %w", err)
	}

	// 프로필 정보 검증
	if profile.ID == 0 || profile.Login == "" {
		return fmt.Errorf("invalid GitHub profile")
	}

	log.Printf("✅ GitHub verified for user %v: %s (%s)", userID, profile.Login, profile.Name)

	// TODO: 데이터베이스 업데이트
	// - user_verification 테이블의 github_connected = true
	// - github_username 저장
	// - 검증 완료 시간 기록

	return nil
}

func (h *VerificationHandler) verifyTwitter(accessToken string, userID interface{}) error {
	// Twitter API v2로 프로필 정보 확인
	apiURL := "https://api.twitter.com/2/users/me"

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create Twitter request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to call Twitter API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("Twitter API error %d: %s", resp.StatusCode, string(body))
	}

	// 응답 파싱
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read Twitter response: %w", err)
	}

	var response struct {
		Data TwitterProfile `json:"data"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return fmt.Errorf("failed to parse Twitter profile: %w", err)
	}

	profile := response.Data
	if profile.ID == "" || profile.Username == "" {
		return fmt.Errorf("invalid Twitter profile")
	}

	log.Printf("✅ Twitter verified for user %v: @%s (%s)", userID, profile.Username, profile.Name)

	// TODO: 데이터베이스 업데이트
	// - user_verification 테이블의 twitter_connected = true
	// - twitter_username 저장
	// - 검증 완료 시간 기록

	return nil
}

func (h *VerificationHandler) verifyDomain(jobData map[string]interface{}) error {
	domain, ok := jobData["domain"].(string)
	if !ok {
		return fmt.Errorf("missing domain")
	}

	email, ok := jobData["email"].(string)
	if !ok {
		return fmt.Errorf("missing email")
	}

	// 이메일 도메인과 회사 도메인 일치 확인
	if !strings.HasSuffix(email, "@"+domain) {
		return fmt.Errorf("email domain mismatch: %s vs %s", email, domain)
	}

	// 도메인 유효성 확인
	if err := h.checkDomainValidity(domain); err != nil {
		return fmt.Errorf("invalid domain: %w", err)
	}

	log.Printf("✅ Domain verified: %s for email %s", domain, email)
	return nil
}

func (h *VerificationHandler) checkDomainValidity(domain string) error {
	// 도메인이 실제로 존재하는지 확인
	// 실제 환경에서는 더 정교한 도메인 검증 로직 구현

	// 공개 이메일 도메인 차단
	publicDomains := map[string]bool{
		"gmail.com":    true,
		"yahoo.com":    true,
		"hotmail.com":  true,
		"outlook.com":  true,
		"naver.com":    true,
		"kakao.com":    true,
		"daum.net":     true,
	}

	if publicDomains[domain] {
		return fmt.Errorf("public email domain not allowed: %s", domain)
	}

	// DNS 조회로 도메인 존재 확인
	// TODO: net.LookupMX() 등을 사용한 실제 DNS 조회

	return nil
}

// 추가: 회사 정보 검증 (선택사항)
func (h *VerificationHandler) verifyCompanyInfo(company, domain string) error {
	// 외부 API를 통한 회사 정보 검증
	// 예: Clearbit, FullContact 등의 API 사용

	log.Printf("🔍 Company verification: %s (%s)", company, domain)

	// TODO: 외부 API 연동
	// - 회사명과 도메인 일치 확인
	// - 회사 규모, 업종 등 추가 정보 수집
	// - 신뢰도 점수 계산

	return nil
}
