package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	Email     string         `json:"email" gorm:"unique;not null"`
	Username  string         `json:"username" gorm:"unique;not null"`
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
	Profile       *UserProfile      `json:"profile,omitempty" gorm:"foreignKey:UserID"`
	Verification  *UserVerification `json:"verification,omitempty" gorm:"foreignKey:UserID"`
	Projects      []Project         `json:"projects,omitempty" gorm:"foreignKey:UserID"`
}

type UserProfile struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	UserID      uint      `json:"user_id" gorm:"uniqueIndex;not null"`
	DisplayName string    `json:"display_name" gorm:"size:100"`
	FirstName   string    `json:"first_name"`
	LastName    string    `json:"last_name"`
	Avatar      string    `json:"avatar"`
	Bio         string    `json:"bio"`
	Age         int       `json:"age"`
	Location    string    `json:"location"`
	Website     string    `json:"website"`
	Occupation  string    `json:"occupation"`
	Experience  string    `json:"experience" gorm:"type:text"` // JSON 형태로 저장
	Skills      string    `json:"skills" gorm:"type:text"`     // JSON 형태로 저장
	Interests   string    `json:"interests" gorm:"type:text"`  // JSON 형태로 저장
	Capital     int64     `json:"capital"`                      // 보유 자본 (원 단위)
	Constraints string    `json:"constraints" gorm:"type:text"` // JSON 형태로 저장

	// 소셜 미디어 링크
	GithubLink   string    `json:"github_link"`
	LinkedinLink string    `json:"linkedin_link"`
	TwitterLink  string    `json:"twitter_link"`

	// 설정 - 알림/공개 범위
	EmailNotifications     bool `json:"email_notifications" gorm:"default:true"`
	PushNotifications      bool `json:"push_notifications" gorm:"default:false"`
	MarketingNotifications bool `json:"marketing_notifications" gorm:"default:false"`
	ProfilePublic          bool `json:"profile_public" gorm:"default:true"`
	InvestmentPublic       bool `json:"investment_public" gorm:"default:false"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// 외래키 참조
	User User `json:"-" gorm:"foreignKey:UserID"`
}

// VerificationStatus 검증 상태
// unverified -> pending -> approved/rejected
type VerificationStatus string

const (
	VerificationUnverified VerificationStatus = "unverified"
	VerificationPending    VerificationStatus = "pending"
	VerificationApproved   VerificationStatus = "approved"
	VerificationRejected   VerificationStatus = "rejected"
)

type UserVerification struct {
	ID     uint `json:"id" gorm:"primaryKey"`
	UserID uint `json:"user_id" gorm:"uniqueIndex;not null"`

	// Level 1
	EmailVerified   bool       `json:"email_verified" gorm:"default:false"`
	EmailVerifiedAt *time.Time `json:"email_verified_at"`
	PhoneVerified   bool       `json:"phone_verified" gorm:"default:false"`
	PhoneVerifiedAt *time.Time `json:"phone_verified_at"`

	// Level 2 - Social & Career
	LinkedInConnected     bool       `json:"linkedin_connected" gorm:"default:false"`
	LinkedInProfileID     *string    `json:"linkedin_profile_id"`
	LinkedInProfileURL    *string    `json:"linkedin_profile_url"`
	LinkedInVerifiedAt    *time.Time `json:"linkedin_verified_at"`

	GitHubConnected       bool       `json:"github_connected" gorm:"default:false"`
	GitHubProfileID       *string    `json:"github_profile_id"`
	GitHubUsername        *string    `json:"github_username"`
	GitHubVerifiedAt      *time.Time `json:"github_verified_at"`

	TwitterConnected      bool       `json:"twitter_connected" gorm:"default:false"`
	TwitterProfileID      *string    `json:"twitter_profile_id"`
	TwitterUsername       *string    `json:"twitter_username"`
	TwitterVerifiedAt     *time.Time `json:"twitter_verified_at"`

	WorkEmailVerified     bool       `json:"work_email_verified" gorm:"default:false"`
	WorkEmailCompany      string     `json:"work_email_company" gorm:"size:120"`
	WorkEmailVerifiedAt   *time.Time `json:"work_email_verified_at"`

	// Level 3 - Professional & Education
	ProfessionalStatus   VerificationStatus `json:"professional_status" gorm:"default:'unverified'"`
	ProfessionalTitle    string             `json:"professional_title" gorm:"size:120"`
	ProfessionalDocPath  string             `json:"professional_doc_path"`
	ProfessionalVerifiedAt *time.Time       `json:"professional_verified_at"`

	EducationStatus   VerificationStatus `json:"education_status" gorm:"default:'unverified'"`
	EducationDegree   string             `json:"education_degree" gorm:"size:120"`
	EducationDocPath  string             `json:"education_doc_path"`
	EducationVerifiedAt *time.Time       `json:"education_verified_at"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// 외래키 참조
	User User `json:"-" gorm:"foreignKey:UserID"`
}

// 사용자 프로필 업데이트를 위한 요청 구조체
type UpdateProfileRequest struct {
	DisplayName  string `json:"display_name"`
	Avatar       string `json:"avatar"`
	Bio          string `json:"bio"`
	Location     string `json:"location"`
	Website      string `json:"website"`
	GithubLink   string `json:"github_link"`
	LinkedinLink string `json:"linkedin_link"`
	TwitterLink  string `json:"twitter_link"`
}

// 설정 업데이트 요청
type UpdatePreferencesRequest struct {
	EmailNotifications     *bool `json:"email_notifications"`
	PushNotifications      *bool `json:"push_notifications"`
	MarketingNotifications *bool `json:"marketing_notifications"`
	ProfilePublic          *bool `json:"profile_public"`
	InvestmentPublic       *bool `json:"investment_public"`
}

// JWT 페이로드에 포함될 사용자 정보
type UserClaims struct {
	UserID   uint   `json:"user_id"`
	Email    string `json:"email"`
	Username string `json:"username"`
}

// MagicLink 매직링크 인증 모델
type MagicLink struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Email     string    `json:"email" gorm:"not null;index"`
	Code      string    `json:"code" gorm:"not null;unique"`
	ExpiresAt time.Time `json:"expires_at" gorm:"not null"`
	IsUsed    bool      `json:"is_used" gorm:"default:false"`
	UserID    *uint     `json:"user_id"` // 연결된 사용자 ID (있다면)

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// TableName GORM 테이블명 설정
func (MagicLink) TableName() string {
	return "magic_links"
}

// 매직링크 생성 요청
type CreateMagicLinkRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// 매직링크 인증 요청
type VerifyMagicLinkRequest struct {
	Code string `json:"code" binding:"required,len=6"`
}
