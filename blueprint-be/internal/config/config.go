package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Database DatabaseConfig
	JWT      JWTConfig
	Google   GoogleConfig
	LinkedIn LinkedInConfig
	Twitter  TwitterConfig
	GitHub   GitHubConfig
	Server   ServerConfig
	AI       AIConfig
	Redis    RedisConfig
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

type JWTConfig struct {
	Secret string
}

type GoogleConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

type ServerConfig struct {
	Port        string
	Mode        string
	FrontendURL string
}

// OpenAIConfig OpenAI 설정
type OpenAIConfig struct {
	APIKey string
	Model  string
}

// AIConfig AI 전반적인 설정
type AIConfig struct {
	Provider string // openai, mock, claude, gemini
	OpenAI   OpenAIConfig
}

// RedisConfig Redis 설정
type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

type LinkedInConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

type TwitterConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

type GitHubConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

// LoadConfig .env 파일을 로드하고 설정을 반환합니다 🔧
func LoadConfig() *Config {
	// .env 파일 로드 (파일이 없어도 오류 없이 진행)
	if err := godotenv.Load(); err != nil {
		log.Println("📁 .env 파일을 찾을 수 없습니다. 시스템 환경변수를 사용합니다.")
	} else {
		log.Println("✅ .env 파일을 성공적으로 로드했습니다.")
	}

	return &Config{
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "password"),
			Name:     getEnv("DB_NAME", "blueprint"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		JWT: JWTConfig{
			Secret: getEnv("JWT_SECRET", "your-super-secret-jwt-key-change-this-in-production"),
		},
		Google: GoogleConfig{
			ClientID:     getEnv("GOOGLE_CLIENT_ID", ""),
			ClientSecret: getEnv("GOOGLE_CLIENT_SECRET", ""),
			RedirectURL:  getEnv("GOOGLE_REDIRECT_URL", "http://localhost:8080/api/v1/auth/google/callback"),
		},
		LinkedIn: LinkedInConfig{
			ClientID:     getEnv("LINKEDIN_CLIENT_ID", ""),
			ClientSecret: getEnv("LINKEDIN_CLIENT_SECRET", ""),
			RedirectURL:  getEnv("LINKEDIN_REDIRECT_URL", "http://localhost:8080/api/v1/auth/linkedin/callback"),
		},
		Twitter: TwitterConfig{
			ClientID:     getEnv("TWITTER_CLIENT_ID", ""),
			ClientSecret: getEnv("TWITTER_CLIENT_SECRET", ""),
			RedirectURL:  getEnv("TWITTER_REDIRECT_URL", "http://localhost:8080/api/v1/auth/twitter/callback"),
		},
		GitHub: GitHubConfig{
			ClientID:     getEnv("GITHUB_CLIENT_ID", ""),
			ClientSecret: getEnv("GITHUB_CLIENT_SECRET", ""),
			RedirectURL:  getEnv("GITHUB_REDIRECT_URL", "http://localhost:8080/api/v1/auth/github/callback"),
		},
		Server: ServerConfig{
			Port:        getEnv("PORT", "8080"),
			Mode:        getEnv("GIN_MODE", "debug"),
			FrontendURL: getEnv("FRONTEND_URL", "http://localhost:3000"),
		},
		AI: AIConfig{
			Provider: getEnv("AI_PROVIDER", "mock"),
			OpenAI: OpenAIConfig{
				APIKey: getEnv("OPENAI_API_KEY", ""),
				Model:  getEnv("OPENAI_MODEL", "gpt-4o-mini"),
			},
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvAsInt("REDIS_DB", 0),
		},
	}
}

// getEnv 환경변수를 가져오거나 기본값을 반환합니다
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt 환경변수를 정수로 가져오거나 기본값을 반환합니다
func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
