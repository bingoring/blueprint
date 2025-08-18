package handlers

import (
	"blueprint-module/pkg/config"
	"blueprint-module/pkg/models"
	"blueprint-module/pkg/queue"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"blueprint/internal/database"
	"blueprint/internal/middleware"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// UserSettingsHandler 사용자 설정 핸들러
type UserSettingsHandler struct {
	cfg *config.Config
}

func NewUserSettingsHandler(cfg *config.Config) *UserSettingsHandler {
	return &UserSettingsHandler{
		cfg: cfg,
	}
}

// GetMySettings 내 프로필/설정/검증 상태 조회
// GET /api/v1/users/me/settings
func (h *UserSettingsHandler) GetMySettings(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		middleware.Unauthorized(c, "User not authenticated")
		return
	}

	var user models.User
	if err := database.GetDB().Preload("Profile").Preload("Verification").First(&user, userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			middleware.NotFound(c, "User not found")
		} else {
			middleware.InternalServerError(c, "Failed to fetch user data")
		}
		return
	}

	// Profile이 없으면 기본값으로 생성
	if user.Profile == nil {
		profile := &models.UserProfile{
			UserID:                 userID.(uint),
			EmailNotifications:     true,
			PushNotifications:      false,
			MarketingNotifications: false,
			ProfilePublic:          true,
			InvestmentPublic:       false,
		}
		if err := database.GetDB().Create(profile).Error; err != nil {
			middleware.InternalServerError(c, "Failed to create default profile")
			return
		}
		user.Profile = profile
	}

	// Verification이 없으면 기본값으로 생성
	if user.Verification == nil {
		verification := &models.UserVerification{
			UserID:             userID.(uint),
			EmailVerified:      false,
			PhoneVerified:      false,
			LinkedInConnected:  false,
			GitHubConnected:    false,
			TwitterConnected:   false,
			WorkEmailVerified:  false,
			ProfessionalStatus: models.VerificationUnverified,
			EducationStatus:    models.VerificationUnverified,
		}
		if err := database.GetDB().Create(verification).Error; err != nil {
			middleware.InternalServerError(c, "Failed to create default verification")
			return
		}
		user.Verification = verification
	}

	middleware.Success(c, gin.H{
		"user": gin.H{
			"id":       user.ID,
			"email":    user.Email,
			"username": user.Username,
		},
		"profile":      user.Profile,
		"verification": user.Verification,
	}, "User settings fetched")
}

// UpdateProfile 내 기본 프로필(표시이름/아바타/바이오) 업데이트
// PUT /api/v1/users/me/profile
func (h *UserSettingsHandler) UpdateProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		middleware.Unauthorized(c, "User not authenticated")
		return
	}

	var req models.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.BadRequest(c, "Invalid request data: "+err.Error())
		return
	}

	// 유효성 검사
	if req.DisplayName != "" && len(req.DisplayName) > 100 {
		middleware.BadRequest(c, "Display name too long (max 100 characters)")
		return
	}
	if req.Bio != "" && len(req.Bio) > 500 {
		middleware.BadRequest(c, "Bio too long (max 500 characters)")
		return
	}

	db := database.GetDB()
	var profile models.UserProfile

	// 기존 프로필 조회 또는 새로 생성
	if err := db.Where("user_id = ?", userID).First(&profile).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			profile = models.UserProfile{
				UserID:                 userID.(uint),
				EmailNotifications:     true,
				PushNotifications:      false,
				MarketingNotifications: false,
				ProfilePublic:          true,
				InvestmentPublic:       false,
			}
		} else {
			middleware.InternalServerError(c, "Failed to query profile")
			return
		}
	}

	// 필드별 업데이트 (빈 값이 아닌 경우만)
	if req.DisplayName != "" {
		profile.DisplayName = req.DisplayName
	}
	if req.Avatar != "" {
		profile.Avatar = req.Avatar
	}
	// Bio는 빈 문자열도 허용 (삭제 가능)
	profile.Bio = req.Bio

	// 데이터베이스 저장
	if profile.ID == 0 {
		if err := db.Create(&profile).Error; err != nil {
			middleware.InternalServerError(c, "Failed to create profile")
			return
		}
	} else {
		if err := db.Save(&profile).Error; err != nil {
			middleware.InternalServerError(c, "Failed to update profile")
			return
		}
	}

	middleware.Success(c, profile, "Profile updated successfully")
}

// UpdatePreferences 알림/공개 범위 설정 업데이트
// PUT /api/v1/users/me/preferences
func (h *UserSettingsHandler) UpdatePreferences(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		middleware.Unauthorized(c, "User not authenticated")
		return
	}

	var req models.UpdatePreferencesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.BadRequest(c, "Invalid request data: "+err.Error())
		return
	}

	db := database.GetDB()
	var profile models.UserProfile

	// 기존 프로필 조회 또는 새로 생성
	if err := db.Where("user_id = ?", userID).First(&profile).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			profile = models.UserProfile{
				UserID:                 userID.(uint),
				EmailNotifications:     true,
				PushNotifications:      false,
				MarketingNotifications: false,
				ProfilePublic:          true,
				InvestmentPublic:       false,
			}
		} else {
			middleware.InternalServerError(c, "Failed to query profile")
			return
		}
	}

	// 각 설정 업데이트 (nil 체크로 전송된 필드만 업데이트)
	if req.EmailNotifications != nil {
		profile.EmailNotifications = *req.EmailNotifications
	}
	if req.PushNotifications != nil {
		profile.PushNotifications = *req.PushNotifications
	}
	if req.MarketingNotifications != nil {
		profile.MarketingNotifications = *req.MarketingNotifications
	}
	if req.ProfilePublic != nil {
		profile.ProfilePublic = *req.ProfilePublic
	}
	if req.InvestmentPublic != nil {
		profile.InvestmentPublic = *req.InvestmentPublic
	}

	// 데이터베이스 저장
	if profile.ID == 0 {
		if err := db.Create(&profile).Error; err != nil {
			middleware.InternalServerError(c, "Failed to create preferences")
			return
		}
	} else {
		if err := db.Save(&profile).Error; err != nil {
			middleware.InternalServerError(c, "Failed to update preferences")
			return
		}
	}

	middleware.Success(c, profile, "Preferences updated successfully")
}

// --- 검증 관련 핸들러들 ---

// RequestVerifyEmail 이메일 인증 요청
// POST /api/v1/users/me/verify/email
func (h *UserSettingsHandler) RequestVerifyEmail(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		middleware.Unauthorized(c, "User not authenticated")
		return
	}

	// 사용자 정보 조회
	var user models.User
	if err := database.GetDB().First(&user, userID).Error; err != nil {
		middleware.NotFound(c, "User not found")
		return
	}

	// 이미 인증된 경우 체크
	var verification models.UserVerification
	if err := database.GetDB().Where("user_id = ?", userID).First(&verification).Error; err == nil {
		if verification.EmailVerified {
			middleware.BadRequest(c, "Email is already verified")
			return
		}
	}

	// 인증 토큰 생성 (6자리 숫자)
	verificationCode := fmt.Sprintf("%06d", time.Now().UnixNano()%1000000)

	// Redis에 인증 코드 저장 (15분 만료)
	redisKey := fmt.Sprintf("email_verification:%d", userID)
	if err := queue.SetWithExpiry(redisKey, verificationCode, 15*time.Minute); err != nil {
		middleware.InternalServerError(c, "Failed to store verification code")
		return
	}

	// 워커 큐에 이메일 전송 작업 추가
	emailJob := map[string]interface{}{
		"type":     "send_email",
		"to":       user.Email,
		"template": "email_verification",
		"data": map[string]interface{}{
			"username": user.Username,
			"code":     verificationCode,
		},
		"user_id":   userID,
		"timestamp": time.Now().Unix(),
	}

	if err := queue.PublishJob("email_queue", emailJob); err != nil {
		middleware.InternalServerError(c, "Failed to queue email job")
		return
	}

	middleware.SuccessWithStatus(c, http.StatusAccepted, gin.H{
		"message":    "Verification email sent",
		"expires_in": 900, // 15분
	}, "Email verification requested")
}

// VerifyEmailCode 이메일 인증 코드 확인
// POST /api/v1/users/me/verify/email/confirm
func (h *UserSettingsHandler) VerifyEmailCode(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		middleware.Unauthorized(c, "User not authenticated")
		return
	}

	var req struct {
		Code string `json:"code" binding:"required,len=6"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.BadRequest(c, "Invalid verification code format")
		return
	}

	// Redis에서 저장된 코드 확인
	redisKey := fmt.Sprintf("email_verification:%d", userID)
	storedCode, err := queue.Get(redisKey)
	if err != nil || storedCode != req.Code {
		middleware.BadRequest(c, "Invalid or expired verification code")
		return
	}

	// 인증 상태 업데이트
	db := database.GetDB()
	now := time.Now()
	var verification models.UserVerification

	if err := db.Where("user_id = ?", userID).First(&verification).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			verification = models.UserVerification{UserID: userID.(uint)}
		} else {
			middleware.InternalServerError(c, "Failed to query verification")
			return
		}
	}

	verification.EmailVerified = true
	verification.EmailVerifiedAt = &now

	if verification.ID == 0 {
		if err := db.Create(&verification).Error; err != nil {
			middleware.InternalServerError(c, "Failed to create verification record")
			return
		}
	} else {
		if err := db.Save(&verification).Error; err != nil {
			middleware.InternalServerError(c, "Failed to update verification")
			return
		}
	}

	// Redis에서 인증 코드 삭제
	queue.Delete(redisKey)

	middleware.Success(c, verification, "Email verified successfully")
}

// RequestVerifyPhone 휴대폰 인증 요청
// POST /api/v1/users/me/verify/phone
func (h *UserSettingsHandler) RequestVerifyPhone(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		middleware.Unauthorized(c, "User not authenticated")
		return
	}

	var req struct {
		PhoneNumber string `json:"phone_number" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.BadRequest(c, "Phone number is required")
		return
	}

	// 휴대폰 번호 유효성 검사 (한국 번호 형식)
	// TODO: 더 정교한 유효성 검사 구현

	// 이미 인증된 경우 체크
	var verification models.UserVerification
	if err := database.GetDB().Where("user_id = ?", userID).First(&verification).Error; err == nil {
		if verification.PhoneVerified {
			middleware.BadRequest(c, "Phone number is already verified")
			return
		}
	}

	// 인증 코드 생성
	verificationCode := fmt.Sprintf("%06d", time.Now().UnixNano()%1000000)

	// Redis에 인증 코드 저장 (5분 만료)
	redisKey := fmt.Sprintf("phone_verification:%d", userID)
	if err := queue.SetWithExpiry(redisKey, verificationCode, 5*time.Minute); err != nil {
		middleware.InternalServerError(c, "Failed to store verification code")
		return
	}

	// 워커 큐에 SMS 전송 작업 추가
	smsJob := map[string]interface{}{
		"type":      "send_sms",
		"to":        req.PhoneNumber,
		"message":   fmt.Sprintf("[Blueprint] 인증번호: %s (5분간 유효)", verificationCode),
		"user_id":   userID,
		"timestamp": time.Now().Unix(),
	}

	if err := queue.PublishJob("sms_queue", smsJob); err != nil {
		middleware.InternalServerError(c, "Failed to queue SMS job")
		return
	}

	middleware.SuccessWithStatus(c, http.StatusAccepted, gin.H{
		"message":    "Verification SMS sent",
		"expires_in": 300, // 5분
	}, "Phone verification requested")
}

// ConnectProvider 소셜 미디어 연결
// POST /api/v1/users/me/connect/:provider (linkedin|github|twitter)
func (h *UserSettingsHandler) ConnectProvider(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		middleware.Unauthorized(c, "User not authenticated")
		return
	}

	provider := c.Param("provider")
	if provider != "linkedin" && provider != "github" && provider != "twitter" {
		middleware.BadRequest(c, "Unsupported provider. Use: linkedin, github, twitter")
		return
	}

	var req struct {
		AccessToken string `json:"access_token" binding:"required"`
		ProfileID   string `json:"profile_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.BadRequest(c, "Access token is required")
		return
	}

	// 외부 API를 통한 토큰 유효성 검사는 워커에서 처리
	verificationJob := map[string]interface{}{
		"type":         "verify_social_provider",
		"provider":     provider,
		"access_token": req.AccessToken,
		"profile_id":   req.ProfileID,
		"user_id":      userID,
		"timestamp":    time.Now().Unix(),
	}

	if err := queue.PublishJob("verification_queue", verificationJob); err != nil {
		middleware.InternalServerError(c, "Failed to queue verification job")
		return
	}

	middleware.SuccessWithStatus(c, http.StatusAccepted, gin.H{
		"message": fmt.Sprintf("%s connection verification started", provider),
		"status":  "pending",
	}, "Provider connection requested")
}

// VerifyWorkEmail 직장 이메일 인증
// POST /api/v1/users/me/verify/work-email
func (h *UserSettingsHandler) VerifyWorkEmail(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		middleware.Unauthorized(c, "User not authenticated")
		return
	}

	var req struct {
		WorkEmail string `json:"work_email" binding:"required,email"`
		Company   string `json:"company" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.BadRequest(c, "Work email and company are required")
		return
	}

	// 회사 도메인 검증 (간단한 형태)
	// TODO: 더 정교한 회사 도메인 검증 로직 구현

	// 인증 코드 생성
	verificationCode := fmt.Sprintf("%06d", time.Now().UnixNano()%1000000)

	// Redis에 저장
	redisKey := fmt.Sprintf("work_email_verification:%d", userID)
	verificationData := map[string]interface{}{
		"code":       verificationCode,
		"work_email": req.WorkEmail,
		"company":    req.Company,
	}

	dataBytes, _ := json.Marshal(verificationData)
	if err := queue.SetWithExpiry(redisKey, string(dataBytes), 15*time.Minute); err != nil {
		middleware.InternalServerError(c, "Failed to store verification data")
		return
	}

	// 워커 큐에 이메일 전송 작업 추가
	emailJob := map[string]interface{}{
		"type":     "send_email",
		"to":       req.WorkEmail,
		"template": "work_email_verification",
		"data": map[string]interface{}{
			"company": req.Company,
			"code":    verificationCode,
		},
		"user_id":   userID,
		"timestamp": time.Now().Unix(),
	}

	if err := queue.PublishJob("email_queue", emailJob); err != nil {
		middleware.InternalServerError(c, "Failed to queue email job")
		return
	}

	middleware.SuccessWithStatus(c, http.StatusAccepted, gin.H{
		"message":    "Work email verification sent",
		"expires_in": 900, // 15분
	}, "Work email verification requested")
}

// SubmitProfessionalDoc 전문 자격 서류 제출
// POST /api/v1/users/me/verify/professional
func (h *UserSettingsHandler) SubmitProfessionalDoc(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		middleware.Unauthorized(c, "User not authenticated")
		return
	}

	// 파일 업로드 처리
	file, header, err := c.Request.FormFile("document")
	if err != nil {
		middleware.BadRequest(c, "Document file is required")
		return
	}
	defer file.Close()

	professionalTitle := c.PostForm("professional_title")
	if professionalTitle == "" {
		middleware.BadRequest(c, "Professional title is required")
		return
	}

	// 파일 크기 제한 (10MB)
	if header.Size > 10*1024*1024 {
		middleware.BadRequest(c, "File size too large (max 10MB)")
		return
	}

	// 허용된 파일 형식 확인
	allowedTypes := map[string]bool{
		"image/jpeg":      true,
		"image/png":       true,
		"application/pdf": true,
	}

	contentType := header.Header.Get("Content-Type")
	if !allowedTypes[contentType] {
		middleware.BadRequest(c, "Invalid file type. Only JPEG, PNG, PDF allowed")
		return
	}

	// 파일 업로드 작업을 워커에 전달
	fileUploadJob := map[string]interface{}{
		"type":         "upload_verification_doc",
		"doc_type":     "professional",
		"user_id":      userID,
		"title":        professionalTitle,
		"filename":     header.Filename,
		"content_type": contentType,
		"size":         header.Size,
		"timestamp":    time.Now().Unix(),
	}

	if err := queue.PublishJob("file_processing_queue", fileUploadJob); err != nil {
		middleware.InternalServerError(c, "Failed to queue file processing job")
		return
	}

	// 검증 상태를 pending으로 업데이트
	db := database.GetDB()
	var verification models.UserVerification

	if err := db.Where("user_id = ?", userID).First(&verification).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			verification = models.UserVerification{UserID: userID.(uint)}
		} else {
			middleware.InternalServerError(c, "Failed to query verification")
			return
		}
	}

	verification.ProfessionalStatus = models.VerificationPending
	verification.ProfessionalTitle = professionalTitle

	if verification.ID == 0 {
		if err := db.Create(&verification).Error; err != nil {
			middleware.InternalServerError(c, "Failed to create verification record")
			return
		}
	} else {
		if err := db.Save(&verification).Error; err != nil {
			middleware.InternalServerError(c, "Failed to update verification")
			return
		}
	}

	middleware.SuccessWithStatus(c, http.StatusAccepted, gin.H{
		"status":  "pending",
		"message": "Professional document submitted for review",
	}, "Professional document submitted")
}

// SubmitEducationDoc 학력 서류 제출
// POST /api/v1/users/me/verify/education
func (h *UserSettingsHandler) SubmitEducationDoc(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		middleware.Unauthorized(c, "User not authenticated")
		return
	}

	// 파일 업로드 처리
	file, header, err := c.Request.FormFile("document")
	if err != nil {
		middleware.BadRequest(c, "Document file is required")
		return
	}
	defer file.Close()

	educationDegree := c.PostForm("education_degree")
	if educationDegree == "" {
		middleware.BadRequest(c, "Education degree is required")
		return
	}

	// 파일 크기 제한 (10MB)
	if header.Size > 10*1024*1024 {
		middleware.BadRequest(c, "File size too large (max 10MB)")
		return
	}

	// 허용된 파일 형식 확인
	allowedTypes := map[string]bool{
		"image/jpeg":      true,
		"image/png":       true,
		"application/pdf": true,
	}

	contentType := header.Header.Get("Content-Type")
	if !allowedTypes[contentType] {
		middleware.BadRequest(c, "Invalid file type. Only JPEG, PNG, PDF allowed")
		return
	}

	// 파일 업로드 작업을 워커에 전달
	fileUploadJob := map[string]interface{}{
		"type":         "upload_verification_doc",
		"doc_type":     "education",
		"user_id":      userID,
		"degree":       educationDegree,
		"filename":     header.Filename,
		"content_type": contentType,
		"size":         header.Size,
		"timestamp":    time.Now().Unix(),
	}

	if err := queue.PublishJob("file_processing_queue", fileUploadJob); err != nil {
		middleware.InternalServerError(c, "Failed to queue file processing job")
		return
	}

	// 검증 상태를 pending으로 업데이트
	db := database.GetDB()
	var verification models.UserVerification

	if err := db.Where("user_id = ?", userID).First(&verification).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			verification = models.UserVerification{UserID: userID.(uint)}
		} else {
			middleware.InternalServerError(c, "Failed to query verification")
			return
		}
	}

	verification.EducationStatus = models.VerificationPending
	verification.EducationDegree = educationDegree

	if verification.ID == 0 {
		if err := db.Create(&verification).Error; err != nil {
			middleware.InternalServerError(c, "Failed to create verification record")
			return
		}
	} else {
		if err := db.Save(&verification).Error; err != nil {
			middleware.InternalServerError(c, "Failed to update verification")
			return
		}
	}

	middleware.SuccessWithStatus(c, http.StatusAccepted, gin.H{
		"status":  "pending",
		"message": "Education document submitted for review",
	}, "Education document submitted")
}
