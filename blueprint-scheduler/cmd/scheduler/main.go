package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"blueprint-module/pkg/config"
	"blueprint-module/pkg/database"
	moduleModels "blueprint-module/pkg/models"
	localConfig "blueprint-scheduler/internal/config"
	"blueprint-scheduler/internal/jobs"
	"blueprint-scheduler/pkg/models"

	"github.com/go-co-op/gocron"
	"github.com/hibiken/asynq"
	"github.com/joho/godotenv"
)

func main() {
	// 환경변수 로드
	if err := godotenv.Load(); err != nil {
		log.Printf("⚠️ .env 파일 로드 실패 (무시): %v", err)
	}

	cfg := localConfig.Load()

	// 데이터베이스 연결
	moduleConfig := &config.Config{
		Database: config.DatabaseConfig{
			Host:     cfg.Database.Host,
			User:     cfg.Database.User,
			Password: cfg.Database.Password,
			Name:     cfg.Database.Name,
			Port:     cfg.Database.Port,
			SSLMode:  cfg.Database.SSLMode,
		},
	}

	if err := database.Connect(moduleConfig); err != nil {
		log.Fatalf("❌ 데이터베이스 연결 실패: %v", err)
	}

	// 통계 테이블 마이그레이션
	db := database.GetDB()
	if err := db.AutoMigrate(
		&models.UserStatsCache{},
		&models.ProjectStatsCache{}, 
		&models.GlobalStatsCache{},
		&models.DashboardCache{},
	); err != nil {
		log.Fatalf("❌ 통계 테이블 마이그레이션 실패: %v", err)
	}
	log.Println("✅ 통계 테이블 마이그레이션 완료")

	// Redis 클라이언트 설정
	redisAddr := cfg.Redis.Host + ":" + cfg.Redis.Port
	redisOpt := asynq.RedisClientOpt{
		Addr:     redisAddr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	}

	// Asynq 클라이언트 (작업 큐 전송용)
	client := asynq.NewClient(redisOpt)
	defer client.Close()

	// Asynq 서버 (작업 처리용)
	srv := asynq.NewServer(redisOpt, asynq.Config{
		Concurrency: cfg.Queue.Concurrency,
		Queues: map[string]int{
			cfg.Queue.StatsQueueName: 6, // 우선순위 6
			"critical":               10, // 우선순위 10 (높음)
			"default":               4,  // 우선순위 4
		},
	})

	// 작업 핸들러 등록
	statsCalculator := jobs.NewStatsCalculator()
	
	mux := asynq.NewServeMux()
	mux.HandleFunc("stats:user", func(ctx context.Context, t *asynq.Task) error {
		var userID uint
		if err := json.Unmarshal(t.Payload(), &userID); err != nil {
			return err
		}
		return statsCalculator.CalculateUserStats(ctx, userID)
	})

	mux.HandleFunc("stats:project", func(ctx context.Context, t *asynq.Task) error {
		var projectID uint
		if err := json.Unmarshal(t.Payload(), &projectID); err != nil {
			return err
		}
		return statsCalculator.CalculateProjectStats(ctx, projectID)
	})

	mux.HandleFunc("stats:dashboard", func(ctx context.Context, t *asynq.Task) error {
		var userID uint
		if err := json.Unmarshal(t.Payload(), &userID); err != nil {
			return err
		}
		return statsCalculator.CalculateDashboardCache(ctx, userID)
	})

	mux.HandleFunc("stats:global", func(ctx context.Context, t *asynq.Task) error {
		return statsCalculator.CalculateGlobalStats(ctx)
	})

	// 스케줄러 설정 (gocron 사용)
	scheduler := gocron.NewScheduler(time.UTC)

	// 매 시간 글로벌 통계 계산
	scheduler.Every(1).Hour().Do(func() {
		task := asynq.NewTask("stats:global", nil)
		client.Enqueue(task, asynq.Queue(cfg.Queue.StatsQueueName))
		log.Println("📊 글로벌 통계 계산 작업 예약됨")
	})

	// 매 30분 활성 사용자 통계 재계산  
	scheduler.Every(30).Minutes().Do(func() {
		// 최근 활성 사용자들의 통계 업데이트
		var activeUserIDs []uint
		db.Model(&moduleModels.User{}).
			Where("updated_at > ?", time.Now().Add(-24*time.Hour)).
			Select("id").
			Find(&activeUserIDs)

		for _, userID := range activeUserIDs {
			payload, _ := json.Marshal(userID)
			task := asynq.NewTask("stats:user", payload)
			client.Enqueue(task, asynq.Queue(cfg.Queue.StatsQueueName))
		}
		log.Printf("📊 활성 사용자 %d명 통계 계산 작업 예약됨", len(activeUserIDs))
	})

	// 매 15분 대시보드 캐시 업데이트
	scheduler.Every(15).Minutes().Do(func() {
		var activeUserIDs []uint
		db.Model(&moduleModels.User{}).
			Where("updated_at > ?", time.Now().Add(-6*time.Hour)).
			Select("id").
			Find(&activeUserIDs)

		for _, userID := range activeUserIDs {
			payload, _ := json.Marshal(userID)
			task := asynq.NewTask("stats:dashboard", payload)
			client.Enqueue(task, asynq.Queue(cfg.Queue.StatsQueueName))
		}
		log.Printf("📊 활성 사용자 %d명 대시보드 캐시 업데이트 예약됨", len(activeUserIDs))
	})

	// 스케줄러 시작
	scheduler.StartAsync()
	log.Println("🕐 스케줄러 시작됨")

	// Asynq 서버 시작
	go func() {
		log.Println("🔄 통계 계산 워커 시작됨")
		if err := srv.Run(mux); err != nil {
			log.Fatalf("❌ Asynq 서버 실행 실패: %v", err)
		}
	}()

	log.Println("🚀 Blueprint 통계 스케줄러 시작됨")

	// 종료 시그널 대기
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("🛑 종료 시그널 수신됨")

	// 정리 작업
	scheduler.Stop()
	srv.Shutdown()
	
	log.Println("✅ Blueprint 통계 스케줄러 종료됨")
}