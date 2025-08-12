package handlers

import (
	"blueprint-module/pkg/config"
	"blueprint-module/pkg/models"
	"blueprint-module/pkg/queue"
	"crypto/rand"
	"fmt"
	"log"
	"math/big"
	"strings"
	"time"

	"blueprint/internal/database"
	"blueprint/internal/middleware"
	"blueprint/pkg/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// MagicLinkHandler 매직링크 전용 핸들러
type MagicLinkHandler struct {
	cfg *config.Config
}

func NewMagicLinkHandler(cfg *config.Config) *MagicLinkHandler {
	return &MagicLinkHandler{
		cfg: cfg,
	}
}

// generateRandomCode 6자리 랜덤 숫자 코드 생성
func generateRandomCode() (string, error) {
	max := big.NewInt(999999)
	min := big.NewInt(100000)

	n, err := rand.Int(rand.Reader, max.Sub(max, min).Add(max, big.NewInt(1)))
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%06d", n.Add(n, min).Int64()), nil
}

// CreateMagicLink 매직링크 생성 (이메일 발송)
func (h *MagicLinkHandler) CreateMagicLink(c *gin.Context) {
	var req models.CreateMagicLinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.BadRequest(c, err.Error())
		return
	}

	// 6자리 랜덤 코드 생성
	code, err := generateRandomCode()
	if err != nil {
		middleware.InternalServerError(c, "Failed to generate verification code")
		return
	}

	// 기존 미사용 매직링크 삭제 (동일 이메일)
	database.GetDB().Where("email = ? AND is_used = false", req.Email).Delete(&models.MagicLink{})

	// 새 매직링크 생성
	magicLink := models.MagicLink{
		Email:     req.Email,
		Code:      code,
		ExpiresAt: time.Now().Add(15 * time.Minute), // 15분 후 만료
		IsUsed:    false,
	}

	if err := database.GetDB().Create(&magicLink).Error; err != nil {
		middleware.InternalServerError(c, "Failed to create magic link")
		return
	}

	// 이메일 발송 (백그라운드)
	err = queue.PublishJob("email", map[string]interface{}{
		"type":       "magic_link",
		"email":      req.Email,
		"code":       code,
		"expires_at": magicLink.ExpiresAt,
	})
	if err != nil {
		log.Printf("❌ Failed to queue magic link email: %v", err)
	}

	middleware.Success(c, gin.H{
		"message":    "Magic link sent",
		"code":       code, // 개발/테스트용 - 프로덕션에서는 제거
		"expires_in": 900,  // 15분 (초)
	}, "Magic link created successfully")
}

// VerifyMagicLink 매직링크 인증 및 로그인
func (h *MagicLinkHandler) VerifyMagicLink(c *gin.Context) {
	var req models.VerifyMagicLinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.BadRequest(c, err.Error())
		return
	}

	// 매직링크 조회
	var magicLink models.MagicLink
	if err := database.GetDB().Where("code = ? AND is_used = false", req.Code).First(&magicLink).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			middleware.Unauthorized(c, "Invalid or expired verification code")
			return
		}
		middleware.InternalServerError(c, "Database error")
		return
	}

	// 만료 확인
	if time.Now().After(magicLink.ExpiresAt) {
		middleware.Unauthorized(c, "Verification code has expired")
		return
	}

	// 매직링크 사용 처리
	magicLink.IsUsed = true
	database.GetDB().Save(&magicLink)

	// 사용자 조회 또는 생성
	var user models.User
	err := database.GetDB().Where("email = ?", magicLink.Email).First(&user).Error

	if err == gorm.ErrRecordNotFound {
		// 새 사용자 생성 (매직링크 방식)
		// 이메일에서 사용자명 생성
		username := strings.Split(magicLink.Email, "@")[0]
		// 중복 방지를 위한 숫자 추가
		originalUsername := username
		counter := 1
		for {
			var existingUser models.User
			err := database.GetDB().Where("username = ?", username).First(&existingUser).Error
			if err == gorm.ErrRecordNotFound {
				break
			}
			counter++
			username = fmt.Sprintf("%s%d", originalUsername, counter)
		}

		user = models.User{
			Email:    magicLink.Email,
			Username: username,
			Provider: "magic_link",
			IsActive: true,
		}

		if err := database.GetDB().Create(&user).Error; err != nil {
			middleware.InternalServerError(c, "Failed to create user")
			return
		}

		// 기본 프로필 생성
		profile := models.UserProfile{
			UserID: user.ID,
		}
		database.GetDB().Create(&profile)

		// 후속 작업들을 큐로 비동기 처리
		publisher := queue.NewPublisher()
		err = publisher.EnqueueUserCreated(queue.UserCreatedEventData{
			UserID:   user.ID,
			Email:    user.Email,
			Username: user.Username,
			Provider: "magic_link",
		})
		if err != nil {
			log.Printf("❌ Failed to enqueue magic link user created tasks: %v", err)
		}
	} else if err != nil {
		middleware.InternalServerError(c, "Database error")
		return
	}

	// 매직링크와 사용자 연결
	magicLink.UserID = &user.ID
	database.GetDB().Save(&magicLink)

	// JWT 토큰 생성
	token, err := utils.GenerateToken(&user, h.cfg.JWT.Secret)
	if err != nil {
		middleware.InternalServerError(c, "Failed to generate token")
		return
	}

	middleware.Success(c, gin.H{
		"token": token,
		"user":  user,
	}, "Magic link verification successful")
}
