package database

import (
	"blueprint/internal/config"
	"blueprint/internal/models"
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func Connect(cfg *config.Config) error {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		cfg.Database.Host,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Name,
		cfg.Database.Port,
		cfg.Database.SSLMode,
	)

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	log.Println("Database connected successfully")
	return nil
}

func AutoMigrate() error {
	if DB == nil {
		return fmt.Errorf("database connection is nil")
	}

	// 테이블 마이그레이션: projects→projects, phases→milestones
	if err := migrateTableName(); err != nil {
		log.Printf("Warning: Table migration failed: %v", err)
	}

	err := DB.AutoMigrate(
		&models.User{},
		&models.Project{},
		&models.Milestone{},
		&models.Order{},                  // P2P 거래 주문
		&models.Trade{},                  // 체결된 거래
		&models.Position{},               // 사용자 포지션
		&models.MarketData{},             // 시장 데이터
		&models.UserWallet{},             // 사용자 지갑 (하이브리드)
		&models.PriceHistory{},           // 가격 히스토리
		// 🆕 하이브리드 화폐 시스템
		&models.StakingPool{},            // 스테이킹 풀
		&models.RevenueDistribution{},    // 수수료 분배
		&models.StakingReward{},          // 스테이커 보상
		&models.GovernanceProposal{},     // 거버넌스 제안
		&models.GovernanceVote{},         // 거버넌스 투표
		&models.BlueprintReward{},        // BLUEPRINT 토큰 보상
		&models.PlatformFeeConfig{},      // 플랫폼 수수료 설정
	)

	if err != nil {
		return fmt.Errorf("failed to auto migrate: %w", err)
	}

	log.Println("Database migration completed successfully")
	return nil
}

// migrateTableName projects 테이블을 projects로, phases 테이블을 milestones로 변경
func migrateTableName() error {
	// projects 테이블을 projects로 변경
	var projectsCount int64
	DB.Raw("SELECT COUNT(*) FROM information_schema.tables WHERE table_name = 'projects' AND table_schema = CURRENT_SCHEMA()").Scan(&projectsCount)

	if projectsCount > 0 {
		log.Println("Found projects table, renaming to projects...")

		// projects 테이블이 이미 있는지 확인
		var projectsCount int64
		DB.Raw("SELECT COUNT(*) FROM information_schema.tables WHERE table_name = 'projects' AND table_schema = CURRENT_SCHEMA()").Scan(&projectsCount)

		if projectsCount == 0 {
			// projects 테이블을 projects로 이름 변경
			if err := DB.Exec("ALTER TABLE projects RENAME TO projects").Error; err != nil {
				log.Printf("Warning: failed to rename projects table to projects: %v", err)
			} else {
				log.Println("Successfully renamed projects table to projects")
			}
		} else {
			log.Println("projects table already exists, skipping projects migration")
		}
	}

	// phases 테이블을 milestones로 변경
	var phasesCount int64
	DB.Raw("SELECT COUNT(*) FROM information_schema.tables WHERE table_name = 'phases' AND table_schema = CURRENT_SCHEMA()").Scan(&phasesCount)

	if phasesCount > 0 {
		log.Println("Found phases table, renaming back to milestones...")

		// milestones 테이블이 이미 있는지 확인
		var milestonesCount int64
		DB.Raw("SELECT COUNT(*) FROM information_schema.tables WHERE table_name = 'milestones' AND table_schema = CURRENT_SCHEMA()").Scan(&milestonesCount)

		if milestonesCount == 0 {
			// phases 테이블을 milestones로 이름 변경
			if err := DB.Exec("ALTER TABLE phases RENAME TO milestones").Error; err != nil {
				log.Printf("Warning: failed to rename phases table to milestones: %v", err)
			} else {
				log.Println("Successfully renamed phases table to milestones")
			}
		} else {
			log.Println("milestones table already exists, skipping phases migration")
		}
	}

	return nil
}

func GetDB() *gorm.DB {
	return DB
}
