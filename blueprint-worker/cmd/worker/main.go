package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	moduleConfig "blueprint-module/pkg/config"
	"blueprint-module/pkg/database"
	moduleRedis "blueprint-module/pkg/redis"
	"blueprint-worker/internal/config"
	"blueprint-worker/internal/handlers"
)

func main() {
	log.Println("🚀 Blueprint Worker Server Starting...")

	// 설정 로드
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 데이터베이스 연결
	dbConfig := &moduleConfig.Config{
		Database: moduleConfig.DatabaseConfig{
			Host:     cfg.Database.Host,
			User:     cfg.Database.User,
			Password: cfg.Database.Password,
			Name:     cfg.Database.Name,
			Port:     cfg.Database.Port,
			SSLMode:  cfg.Database.SSLMode,
		},
	}
	if err := database.Connect(dbConfig); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Redis 연결
	redisConfig := &moduleConfig.Config{
		Redis: moduleConfig.RedisConfig{
			Host:     cfg.Redis.Host,
			Port:     cfg.Redis.Port,
			Password: cfg.Redis.Password,
			DB:       cfg.Redis.DB,
		},
	}
	if err := moduleRedis.InitRedis(redisConfig); err != nil {
		log.Fatalf("Failed to initialize Redis: %v", err)
	}
	defer moduleRedis.CloseRedis()

	// 워커 핸들러 초기화
	emailHandler := handlers.NewEmailHandler(cfg)
	smsHandler := handlers.NewSMSHandler(cfg)
	fileHandler := handlers.NewFileHandler(cfg)
	verificationHandler := handlers.NewVerificationHandler(cfg)
	activityHandler := handlers.NewActivityHandler() // 활동 로그 핸들러 추가

	// Graceful shutdown을 위한 context 생성
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 워커 시작
	var wg sync.WaitGroup

	// 이메일 큐 워커
	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Println("📧 Starting Email Queue Worker...")
		if err := emailHandler.StartEmailWorker(ctx); err != nil {
			log.Printf("Email worker error: %v", err)
		}
	}()

	// SMS 큐 워커 (기존 버전 유지)
	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Println("📱 Starting SMS Queue Worker...")
		if err := smsHandler.StartSMSWorker(); err != nil {
			log.Printf("SMS worker error: %v", err)
		}
	}()

	// 파일 처리 큐 워커 (기존 버전 유지)
	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Println("📁 Starting File Processing Worker...")
		if err := fileHandler.StartFileWorker(); err != nil {
			log.Printf("File worker error: %v", err)
		}
	}()

	// 검증 큐 워커 (기존 버전 유지)
	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Println("🔍 Starting Verification Worker...")
		if err := verificationHandler.StartVerificationWorker(); err != nil {
			log.Printf("Verification worker error: %v", err)
		}
	}()

	// 활동 로그 큐 워커
	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Println("📝 Starting Activity Log Worker...")
		if err := activityHandler.StartActivityWorker(ctx); err != nil {
			log.Printf("Activity worker error: %v", err)
		}
	}()

	log.Println("✅ All workers started successfully")

	// Graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	log.Println("🛑 Shutting down worker server...")

	// Context 취소로 모든 워커에 종료 신호 전송
	cancel()

	// 모든 워커가 종료될 때까지 대기
	wg.Wait()
	log.Println("✅ Worker server shutdown complete")
}
