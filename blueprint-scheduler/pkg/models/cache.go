package models

import (
	"time"
)

// UserStatsCache 사용자별 통계 캐시 테이블
type UserStatsCache struct {
	ID                   uint      `json:"id" gorm:"primaryKey"`
	UserID               uint      `json:"user_id" gorm:"unique;not null;index"`
	ProjectSuccessRate   float64   `json:"project_success_rate"`
	MentoringSuccessRate float64   `json:"mentoring_success_rate"`
	TotalInvestment      int64     `json:"total_investment"` // USDC cents
	SbtCount             int       `json:"sbt_count"`
	LastCalculatedAt     time.Time `json:"last_calculated_at"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

// ProjectStatsCache 프로젝트별 통계 캐시 테이블  
type ProjectStatsCache struct {
	ID               uint      `json:"id" gorm:"primaryKey"`
	ProjectID        uint      `json:"project_id" gorm:"unique;not null;index"`
	TotalInvestment  int64     `json:"total_investment"`
	InvestorCount    int       `json:"investor_count"`
	SuccessProbability float64 `json:"success_probability"`
	CompletionRate   float64   `json:"completion_rate"`
	LastCalculatedAt time.Time `json:"last_calculated_at"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// GlobalStatsCache 전체 플랫폼 통계 캐시
type GlobalStatsCache struct {
	ID                   uint      `json:"id" gorm:"primaryKey"`
	StatKey              string    `json:"stat_key" gorm:"unique;not null;index"`
	StatValue            float64   `json:"stat_value"`
	StatMeta             string    `json:"stat_meta" gorm:"type:jsonb"` // JSON 메타데이터
	LastCalculatedAt     time.Time `json:"last_calculated_at"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

// 글로벌 통계 키 상수들
const (
	GlobalStatActiveUsers     = "active_users"
	GlobalStatActiveProjects  = "active_projects" 
	GlobalStatActiveDisputes  = "active_disputes"
	GlobalStatTotalTVL       = "total_tvl"
	GlobalStatAvgSuccessRate = "avg_success_rate"
	GlobalStatTotalTrades    = "total_trades"
)

// DashboardCache 대시보드 데이터 캐시
type DashboardCache struct {
	ID               uint      `json:"id" gorm:"primaryKey"`
	UserID           uint      `json:"user_id" gorm:"unique;not null;index"`
	FeaturedProjects string    `json:"featured_projects" gorm:"type:jsonb"` // JSON 배열
	ActivityFeed     string    `json:"activity_feed" gorm:"type:jsonb"`     // JSON 배열
	Portfolio        string    `json:"portfolio" gorm:"type:jsonb"`         // JSON 객체
	NextMilestone    string    `json:"next_milestone" gorm:"type:jsonb"`    // JSON 객체
	LastCalculatedAt time.Time `json:"last_calculated_at"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}