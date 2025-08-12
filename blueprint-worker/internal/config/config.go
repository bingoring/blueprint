package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	// 데이터베이스 설정
	Database DatabaseConfig `json:"database"`

	// Redis 설정
	Redis RedisConfig `json:"redis"`

	// 이메일 서비스 설정
	Email EmailConfig `json:"email"`

	// SMS 서비스 설정
	SMS SMSConfig `json:"sms"`

	// 파일 저장소 설정
	Storage StorageConfig `json:"storage"`

	// 소셜 미디어 API 설정
	Social SocialConfig `json:"social"`
}

type DatabaseConfig struct {
	Host     string `json:"host"`
	User     string `json:"user"`
	Password string `json:"password"`
	Name     string `json:"name"`
	Port     string `json:"port"`
	SSLMode  string `json:"ssl_mode"`
}

type RedisConfig struct {
	Host     string `json:"host"`
	Port     string `json:"port"`
	Password string `json:"password"`
	DB       int    `json:"db"`
}

type EmailConfig struct {
	SMTPHost     string `json:"smtp_host"`
	SMTPPort     string `json:"smtp_port"`
	SMTPUsername string `json:"smtp_username"`
	SMTPPassword string `json:"smtp_password"`
	FromEmail    string `json:"from_email"`
	FromName     string `json:"from_name"`
}

type SMSConfig struct {
	Provider   string `json:"provider"`   // "twilio", "aligo", "solapi"
	APIKey     string `json:"api_key"`
	APISecret  string `json:"api_secret"`
	FromNumber string `json:"from_number"`
}

type StorageConfig struct {
	Provider        string `json:"provider"`         // "s3", "r2", "local"
	Bucket          string `json:"bucket"`
	Region          string `json:"region"`
	AccessKeyID     string `json:"access_key_id"`
	SecretAccessKey string `json:"secret_access_key"`
	Endpoint        string `json:"endpoint"`         // For R2 or custom S3 endpoint
	LocalPath       string `json:"local_path"`       // For local storage
}

type SocialConfig struct {
	LinkedIn LinkedInConfig `json:"linkedin"`
	GitHub   GitHubConfig   `json:"github"`
	Twitter  TwitterConfig  `json:"twitter"`
}

type LinkedInConfig struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

type GitHubConfig struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

type TwitterConfig struct {
	APIKey    string `json:"api_key"`
	APISecret string `json:"api_secret"`
}

func LoadConfig() (*Config, error) {
	// .env 파일 로드 (선택적)
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	config := &Config{
		Database: DatabaseConfig{
			Host:     getEnv("DATABASE_HOST", "localhost"),
			User:     getEnv("DATABASE_USER", "postgres"),
			Password: getEnv("DATABASE_PASSWORD", ""),
			Name:     getEnv("DATABASE_NAME", "blueprint"),
			Port:     getEnv("DATABASE_PORT", "5432"),
			SSLMode:  getEnv("DATABASE_SSL_MODE", "disable"),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       0,
		},
		Email: EmailConfig{
			SMTPHost:     getEnv("SMTP_HOST", "smtp.gmail.com"),
			SMTPPort:     getEnv("SMTP_PORT", "587"),
			SMTPUsername: getEnv("SMTP_USERNAME", ""),
			SMTPPassword: getEnv("SMTP_PASSWORD", ""),
			FromEmail:    getEnv("FROM_EMAIL", "noreply@blueprint.io"),
			FromName:     getEnv("FROM_NAME", "Blueprint"),
		},
		SMS: SMSConfig{
			Provider:   getEnv("SMS_PROVIDER", "aligo"),
			APIKey:     getEnv("SMS_API_KEY", ""),
			APISecret:  getEnv("SMS_API_SECRET", ""),
			FromNumber: getEnv("SMS_FROM_NUMBER", ""),
		},
		Storage: StorageConfig{
			Provider:        getEnv("STORAGE_PROVIDER", "local"),
			Bucket:          getEnv("STORAGE_BUCKET", "blueprint-files"),
			Region:          getEnv("STORAGE_REGION", "us-east-1"),
			AccessKeyID:     getEnv("STORAGE_ACCESS_KEY_ID", ""),
			SecretAccessKey: getEnv("STORAGE_SECRET_ACCESS_KEY", ""),
			Endpoint:        getEnv("STORAGE_ENDPOINT", ""),
			LocalPath:       getEnv("STORAGE_LOCAL_PATH", "./uploads"),
		},
		Social: SocialConfig{
			LinkedIn: LinkedInConfig{
				ClientID:     getEnv("LINKEDIN_CLIENT_ID", ""),
				ClientSecret: getEnv("LINKEDIN_CLIENT_SECRET", ""),
			},
			GitHub: GitHubConfig{
				ClientID:     getEnv("GITHUB_CLIENT_ID", ""),
				ClientSecret: getEnv("GITHUB_CLIENT_SECRET", ""),
			},
			Twitter: TwitterConfig{
				APIKey:    getEnv("TWITTER_API_KEY", ""),
				APISecret: getEnv("TWITTER_API_SECRET", ""),
			},
		},
	}

	return config, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
