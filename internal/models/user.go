package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	Email     string         `json:"email" gorm:"unique;not null"`
	Username  string         `json:"username" gorm:"unique;not null"`
	Password  string         `json:"-" gorm:"not null"` // JSON에서 제외
	Provider  string         `json:"provider" gorm:"default:'local'"`
	GoogleID  *string        `json:"google_id" gorm:"unique"`
	IsActive  bool           `json:"is_active" gorm:"default:true"`

	// AI 사용 횟수 추적 🤖
	AIUsageCount int `json:"ai_usage_count" gorm:"default:0"` // 사용한 횟수
	AIUsageLimit int `json:"ai_usage_limit" gorm:"default:5"` // 최대 사용 가능 횟수

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// 관계 (순환 참조 방지를 위해 포인터 사용)
	Profile *UserProfile `json:"profile,omitempty" gorm:"foreignKey:UserID"`
	Goals   []Goal       `json:"goals,omitempty" gorm:"foreignKey:UserID"`
}

type UserProfile struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	UserID      uint      `json:"user_id" gorm:"uniqueIndex;not null"`
	FirstName   string    `json:"first_name"`
	LastName    string    `json:"last_name"`
	Avatar      string    `json:"avatar"`
	Bio         string    `json:"bio"`
	Age         int       `json:"age"`
	Location    string    `json:"location"`
	Occupation  string    `json:"occupation"`
	Experience  string    `json:"experience" gorm:"type:text"` // JSON 형태로 저장
	Skills      string    `json:"skills" gorm:"type:text"`     // JSON 형태로 저장
	Interests   string    `json:"interests" gorm:"type:text"`  // JSON 형태로 저장
	Capital     int64     `json:"capital"`                     // 보유 자본 (원 단위)
	Constraints string    `json:"constraints" gorm:"type:text"` // JSON 형태로 저장
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// 외래키 참조
	User User `json:"-" gorm:"foreignKey:UserID"`
}

// 사용자 생성을 위한 요청 구조체
type CreateUserRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Username string `json:"username" binding:"required,min=3,max=30"`
	Password string `json:"password" binding:"required,min=6"`
}

// 사용자 로그인을 위한 요청 구조체
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// 사용자 프로필 업데이트를 위한 요청 구조체
type UpdateProfileRequest struct {
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	Avatar      string `json:"avatar"`
	Bio         string `json:"bio"`
	Age         int    `json:"age"`
	Location    string `json:"location"`
	Occupation  string `json:"occupation"`
	Experience  string `json:"experience"`
	Skills      string `json:"skills"`
	Interests   string `json:"interests"`
	Capital     int64  `json:"capital"`
	Constraints string `json:"constraints"`
}

// JWT 페이로드에 포함될 사용자 정보
type UserClaims struct {
	UserID   uint   `json:"user_id"`
	Email    string `json:"email"`
	Username string `json:"username"`
}
