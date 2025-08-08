package handlers

import (
	"blueprint/internal/config"
	"blueprint/internal/database"
	"blueprint/internal/middleware"
	"blueprint/internal/models"
	"blueprint/pkg/utils"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"gorm.io/gorm"

	"blueprint/internal/queue"
)

// Google 사용자 정보 구조체
type GoogleUserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
}

// generateUsernameFromGoogleName Google 이름을 기반으로 사용 가능한 username 생성
func generateUsernameFromGoogleName(name string, googleID string) string {
	// 1. 이름을 소문자로 변환하고 공백을 언더스코어로 변경
	username := strings.ToLower(strings.TrimSpace(name))
	username = strings.ReplaceAll(username, " ", "_")

	// 2. 영문, 숫자, 언더스코어만 남기기 (한글 등 특수문자 제거)
	reg := regexp.MustCompile(`[^a-z0-9_]`)
	username = reg.ReplaceAllString(username, "")

	// 3. 빈 문자열이거나 너무 짧으면 Google ID 사용
	if len(username) < 2 {
		username = fmt.Sprintf("user_%s", googleID[:8])
		return username
	}

	// 4. 너무 길면 잘라내기 (최대 20자)
	if len(username) > 20 {
		username = username[:20]
	}

	// 5. 중복 확인 및 고유한 username 생성
	originalUsername := username
	counter := 1

	for {
		var existingUser models.User
		err := database.GetDB().Where("username = ?", username).First(&existingUser).Error

		if err == gorm.ErrRecordNotFound {
			// 사용 가능한 username 발견
			break
		}

		// 중복이면 숫자 추가
		counter++
		username = fmt.Sprintf("%s_%d", originalUsername, counter)

		// 안전장치: 너무 많이 시도하면 Google ID 사용
		if counter > 999 {
			username = fmt.Sprintf("user_%s", googleID[:8])
			break
		}
	}

	return username
}

type AuthHandler struct {
	cfg         *config.Config
	googleOAuth *oauth2.Config
}

func NewAuthHandler(cfg *config.Config) *AuthHandler {
	googleConfig := &oauth2.Config{
		ClientID:     cfg.Google.ClientID,
		ClientSecret: cfg.Google.ClientSecret,
		RedirectURL:  cfg.Google.RedirectURL,
		Scopes:       []string{"openid", "email", "profile"},
		Endpoint:     google.Endpoint,
	}

	return &AuthHandler{
		cfg:         cfg,
		googleOAuth: googleConfig,
	}
}

// 회원가입
func (h *AuthHandler) Register(c *gin.Context) {
	var req models.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.BadRequest(c, err.Error())
		return
	}

	// 이메일 중복 확인
	var existingUser models.User
	if err := database.GetDB().Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		middleware.Conflict(c, "Email already exists")
		return
	}

	// 사용자명 중복 확인
	if err := database.GetDB().Where("username = ?", req.Username).First(&existingUser).Error; err == nil {
		middleware.Conflict(c, "Username already exists")
		return
	}

	// 비밀번호 해시화
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		middleware.InternalServerError(c, "Failed to hash password")
		return
	}

	// 사용자 생성
	user := models.User{
		Email:    req.Email,
		Username: req.Username,
		Password: hashedPassword,
		Provider: "local",
		IsActive: true,
	}

	if err := database.GetDB().Create(&user).Error; err != nil {
		middleware.InternalServerError(c, "Failed to create user")
		return
	}

	// 기본 프로필 생성 (즉시 처리 - 중요한 메인 동작)
	profile := models.UserProfile{
		UserID: user.ID,
	}
	database.GetDB().Create(&profile)

	// 🆕 후속 작업들을 큐로 비동기 처리
	publisher := queue.NewPublisher()
	err = publisher.EnqueueUserCreated(queue.UserCreatedEventData{
		UserID:   user.ID,
		Email:    user.Email,
		Username: user.Username,
		Provider: "local",
	})
	if err != nil {
		log.Printf("❌ Failed to enqueue user created tasks: %v", err)
		// 에러가 나도 회원가입은 성공으로 처리 (백그라운드 작업은 나중에 재시도 가능)
	}

	// JWT 토큰 생성
	token, err := utils.GenerateToken(&user, h.cfg.JWT.Secret)
	if err != nil {
		middleware.InternalServerError(c, "Failed to generate token")
		return
	}

	middleware.SuccessWithStatus(c, http.StatusCreated, gin.H{
		"token": token,
		"user":  user,
	}, "User created successfully")
}

// 로그인
func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.BadRequest(c, err.Error())
		return
	}

	// 사용자 조회
	var user models.User
	if err := database.GetDB().Where("email = ?", req.Email).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			middleware.Unauthorized(c, "Invalid credentials")
			return
		}
		middleware.InternalServerError(c, "Database error")
		return
	}

	// 비밀번호 확인
	if !utils.CheckPassword(req.Password, user.Password) {
		middleware.Unauthorized(c, "Invalid credentials")
		return
	}

	// 계정 활성화 확인
	if !user.IsActive {
		middleware.Unauthorized(c, "Account is disabled")
		return
	}

	// JWT 토큰 생성
	token, err := utils.GenerateToken(&user, h.cfg.JWT.Secret)
	if err != nil {
		middleware.InternalServerError(c, "Failed to generate token")
		return
	}

	middleware.Success(c, gin.H{
		"token": token,
		"user":  user,
	}, "Login successful")
}

// Google OAuth 로그인 시작
func (h *AuthHandler) GoogleLogin(c *gin.Context) {
	url := h.googleOAuth.AuthCodeURL("state", oauth2.AccessTypeOffline)
	middleware.Success(c, gin.H{"auth_url": url}, "Google auth URL generated successfully")
}

// Google OAuth 콜백
func (h *AuthHandler) GoogleCallback(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Authorization code not provided"})
		return
	}

	// 토큰 교환
	token, err := h.googleOAuth.Exchange(context.Background(), code)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to exchange token"})
		return
	}

	// 사용자 정보 가져오기 (HTTP 요청 사용)
	client := h.googleOAuth.Client(context.Background(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user info"})
		return
	}
	defer resp.Body.Close()

	var userinfo GoogleUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userinfo); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode user info"})
		return
	}

	// 기존 사용자 확인
	var user models.User
	err = database.GetDB().Where("google_id = ? OR email = ?", userinfo.ID, userinfo.Email).First(&user).Error

	if err == gorm.ErrRecordNotFound {
		// 새 사용자 생성
		googleID := userinfo.ID // 포인터를 위한 변수
		user = models.User{
			Email:    userinfo.Email,
			Username: generateUsernameFromGoogleName(userinfo.Name, userinfo.ID),
			Provider: "google",
			GoogleID: &googleID, // 포인터로 설정
			IsActive: true,
		}

		if err := database.GetDB().Create(&user).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
			return
		}

		// 기본 프로필 생성
		profile := models.UserProfile{
			UserID:    user.ID,
			FirstName: userinfo.GivenName,
			LastName:  userinfo.FamilyName,
			Avatar:    userinfo.Picture,
		}
		database.GetDB().Create(&profile)

		// 🆕 Google 회원가입 후속 작업들을 큐로 비동기 처리
		publisher := queue.NewPublisher()
		err = publisher.EnqueueUserCreated(queue.UserCreatedEventData{
			UserID:   user.ID,
			Email:    user.Email,
			Username: user.Username,
			Provider: "google",
		})
		if err != nil {
			log.Printf("❌ Failed to enqueue Google user created tasks: %v", err)
		}
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// JWT 토큰 생성
	jwtToken, err := utils.GenerateToken(&user, h.cfg.JWT.Secret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	// 프론트엔드로 JWT 토큰과 함께 리다이렉트
	frontendURL := fmt.Sprintf("http://localhost:3000?token=%s&user_id=%d", jwtToken, user.ID)
	c.Redirect(http.StatusFound, frontendURL)
}

// 현재 사용자 정보 조회
func (h *AuthHandler) Me(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		middleware.Unauthorized(c, "User not authenticated")
		return
	}

	var user models.User
	if err := database.GetDB().Preload("Profile").First(&user, userID).Error; err != nil {
		middleware.NotFound(c, "User not found")
		return
	}

	middleware.Success(c, user, "User information retrieved successfully")
}

// Logout 로그아웃 처리 🚪
func (h *AuthHandler) Logout(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		middleware.Unauthorized(c, "User not authenticated")
		return
	}

	// 현재 JWT 기반이므로 클라이언트에서 토큰 삭제하도록 안내
	// 향후 Redis 기반 블랙리스트나 세션 관리로 확장 가능
	middleware.Success(c, gin.H{
		"message": "로그아웃이 완료되었습니다",
		"user_id": userID,
		"logout_time": time.Now(),
		"instructions": "클라이언트에서 토큰을 삭제해주세요",
	}, "로그아웃이 성공적으로 처리되었습니다")
}

// RefreshToken JWT 토큰 갱신 🔄
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		middleware.Unauthorized(c, "User not authenticated")
		return
	}

	// 사용자 정보 조회
	var user models.User
	if err := database.GetDB().First(&user, userID).Error; err != nil {
		middleware.NotFound(c, "사용자를 찾을 수 없습니다")
		return
	}

	// 새로운 토큰 생성
	token, err := utils.GenerateToken(&user, h.cfg.JWT.Secret)
	if err != nil {
		middleware.InternalServerError(c, "토큰 생성에 실패했습니다")
		return
	}

	middleware.Success(c, gin.H{
		"token": token,
		"user":  user,
		"expires_in": 24 * 60 * 60, // 24시간 (초 단위)
		"refresh_time": time.Now(),
	}, "토큰이 성공적으로 갱신되었습니다")
}

// CheckTokenExpiry 토큰 만료 확인 ⏰
func (h *AuthHandler) CheckTokenExpiry(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		middleware.Unauthorized(c, "User not authenticated")
		return
	}

	// Authorization 헤더에서 토큰 추출
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		middleware.Unauthorized(c, "Authorization header missing")
		return
	}

	tokenString := ""
	if strings.HasPrefix(authHeader, "Bearer ") {
		tokenString = authHeader[7:]
	} else {
		middleware.Unauthorized(c, "Invalid authorization format")
		return
	}

	// 토큰 만료 시간 확인
	expirationTime, err := utils.GetTokenExpirationTime(tokenString, h.cfg.JWT.Secret)
	if err != nil {
		middleware.Unauthorized(c, "Invalid token")
		return
	}

	// 남은 시간 계산
	remaining, err := utils.GetTokenRemainingTime(tokenString, h.cfg.JWT.Secret)
	if err != nil {
		middleware.Unauthorized(c, "Token has expired")
		return
	}

	// 만료 여부 확인
	isExpired := utils.IsTokenExpired(tokenString, h.cfg.JWT.Secret)

	middleware.Success(c, gin.H{
		"user_id":          userID,
		"expiration_time":  expirationTime,
		"remaining_seconds": int(remaining.Seconds()),
		"remaining_minutes": int(remaining.Minutes()),
		"remaining_hours":   int(remaining.Hours()),
		"is_expired":       isExpired,
		"should_refresh":   remaining.Minutes() < 30, // 30분 이하일 때 갱신 권장
		"checked_at":       time.Now(),
	}, "토큰 만료 정보를 성공적으로 조회했습니다")
}
