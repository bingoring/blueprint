package handlers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"blueprint-module/pkg/config"
	"blueprint-module/pkg/models"
	"blueprint-module/pkg/oauth"
	"blueprint/internal/database"
	"blueprint/internal/middleware"

	"github.com/gin-gonic/gin"
)

// OAuthHandler OAuth 관련 핸들러
type OAuthHandler struct {
	oauthService *oauth.OAuthService
	config       *config.Config
}

// NewOAuthHandler OAuth 핸들러 생성
func NewOAuthHandler(cfg *config.Config) *OAuthHandler {
	return &OAuthHandler{
		oauthService: oauth.NewOAuthService(cfg.OAuth),
		config:       cfg,
	}
}

// StartOAuthConnect 소셜 미디어 연결 시작 (신원 증명용)
// GET /api/v1/auth/:provider/connect
func (h *OAuthHandler) StartOAuthConnect(c *gin.Context) {
	provider := c.Param("provider")

	// 사용자 인증 확인 (연결 기능은 로그인된 사용자만 가능)
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		middleware.Unauthorized(c, "User authentication required for social media connection")
		return
	}

	userID, ok := userIDInterface.(uint)
	if !ok {
		middleware.InternalServerError(c, "Invalid user ID format")
		return
	}

	// OAuth 인증 URL 생성
	authURL, err := h.oauthService.GetAuthURL(provider, userID, "connect")
	if err != nil {
		if err.Error() == fmt.Sprintf("oauth provider '%s' not found", provider) {
			middleware.BadRequest(c, fmt.Sprintf("Unsupported provider: %s", provider))
			return
		}
		middleware.InternalServerError(c, "Failed to generate OAuth URL")
		return
	}

	middleware.Success(c, gin.H{
		"auth_url": authURL,
		"provider": provider,
		"action":   "connect",
	}, "OAuth URL generated successfully")
}

// OAuthCallback OAuth 콜백 처리
// GET /api/v1/auth/:provider/callback
func (h *OAuthHandler) OAuthCallback(c *gin.Context) {
	provider := c.Param("provider")
	code := c.Query("code")
	state := c.Query("state")
	errorParam := c.Query("error")

	// 에러 처리
	if errorParam != "" {
		errorDescription := c.Query("error_description")
		redirectURL := fmt.Sprintf("%s/settings?error=%s&description=%s",
			h.config.Server.FrontendURL, errorParam, errorDescription)
		c.Redirect(http.StatusFound, redirectURL)
		return
	}

	// code 검증
	if code == "" {
		redirectURL := fmt.Sprintf("%s/settings?error=no_code", h.config.Server.FrontendURL)
		c.Redirect(http.StatusFound, redirectURL)
		return
	}

	// OAuth 콜백 처리
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := h.oauthService.HandleCallback(ctx, provider, code, state)
	if err != nil {
		redirectURL := fmt.Sprintf("%s/settings?error=oauth_failed&provider=%s",
			h.config.Server.FrontendURL, provider)
		c.Redirect(http.StatusFound, redirectURL)
		return
	}

	// 연결 액션 처리
	if result.Action == "connect" {
		err = h.handleSocialConnection(result)
		if err != nil {
			redirectURL := fmt.Sprintf("%s/settings?error=connection_failed&provider=%s",
				h.config.Server.FrontendURL, provider)
			c.Redirect(http.StatusFound, redirectURL)
			return
		}

		// 성공 리다이렉트
		redirectURL := fmt.Sprintf("%s/settings?connected=%s&name=%s",
			h.config.Server.FrontendURL, provider, result.Profile.DisplayName)
		c.Redirect(http.StatusFound, redirectURL)
		return
	}

	// 기타 액션 (로그인 등) - 향후 확장
	redirectURL := fmt.Sprintf("%s/settings?error=unsupported_action", h.config.Server.FrontendURL)
	c.Redirect(http.StatusFound, redirectURL)
}

// handleSocialConnection 소셜 미디어 연결 처리
func (h *OAuthHandler) handleSocialConnection(result *oauth.CallbackResult) error {
	db := database.GetDB()

	// 사용자 조회
	var user models.User
	if err := db.First(&user, result.UserID).Error; err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// 기존 연결 여부 확인
	var verification models.UserVerification
	err := db.Where("user_id = ?", result.UserID).First(&verification).Error
	if err != nil {
		// 새로운 verification 레코드 생성
		verification = models.UserVerification{
			UserID: result.UserID,
		}
	}

	// 제공업체별 연결 정보 업데이트
	switch result.Provider {
	case "linkedin":
		verification.LinkedInConnected = true
		verification.LinkedInProfileID = &result.Profile.ID
		verification.LinkedInProfileURL = &result.Profile.ProfileURL
		verification.LinkedInVerifiedAt = &[]time.Time{time.Now()}[0]
	case "github":
		verification.GitHubConnected = true
		verification.GitHubProfileID = &result.Profile.ID
		verification.GitHubUsername = &result.Profile.DisplayName
		verification.GitHubVerifiedAt = &[]time.Time{time.Now()}[0]
	case "twitter":
		verification.TwitterConnected = true
		verification.TwitterProfileID = &result.Profile.ID
		verification.TwitterUsername = &result.Profile.DisplayName
		verification.TwitterVerifiedAt = &[]time.Time{time.Now()}[0]
	default:
		return fmt.Errorf("unsupported provider for connection: %s", result.Provider)
	}

	// 데이터베이스에 저장
	if verification.ID == 0 {
		// 새로운 레코드 생성
		if err := db.Create(&verification).Error; err != nil {
			return fmt.Errorf("failed to create verification record: %w", err)
		}
	} else {
		// 기존 레코드 업데이트
		if err := db.Save(&verification).Error; err != nil {
			return fmt.Errorf("failed to update verification record: %w", err)
		}
	}

	return nil
}

// GetSupportedProviders 지원되는 OAuth 제공업체 목록 조회
// GET /api/v1/auth/providers
func (h *OAuthHandler) GetSupportedProviders(c *gin.Context) {
	providers := h.oauthService.GetSupportedProviders()

	middleware.Success(c, gin.H{
		"providers": providers,
		"count":     len(providers),
	}, "Supported OAuth providers retrieved successfully")
}
