package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"blueprint-module/pkg/models"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	// 환경변수에서 설정 읽기
	dbType := getEnv("DB_TYPE", "sqlite")
	dbURL := getEnv("DATABASE_URL", "test.db")
	
	// PostgreSQL 연결을 위한 개별 환경변수 읽기
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "postgres")
	dbPassword := getEnv("DB_PASSWORD", "password")
	dbName := getEnv("DB_NAME", "blueprint")
	dbSSLMode := getEnv("DB_SSLMODE", "disable")
	numUsers := getEnvInt("NUM_USERS", 100)
	usdcBalance := getEnvInt64("USDC_BALANCE", 100000000) // $1,000,000 기본값

	fmt.Printf("🚀 테스트 계정 생성 스크립트 시작\n")
	fmt.Printf("   - DB 타입: %s\n", dbType)
	if dbType == "postgres" {
		fmt.Printf("   - PostgreSQL 호스트: %s:%s\n", dbHost, dbPort)
		fmt.Printf("   - 데이터베이스: %s\n", dbName)
		fmt.Printf("   - 사용자: %s\n", dbUser)
	} else {
		fmt.Printf("   - DB URL: %s\n", dbURL)
	}
	fmt.Printf("   - 생성할 사용자 수: %d\n", numUsers)
	fmt.Printf("   - 기본 USDC 잔액: $%.2f\n", float64(usdcBalance)/100)

	// 데이터베이스 연결
	var db *gorm.DB
	var err error

	switch dbType {
	case "postgres":
		// 개별 환경변수로 PostgreSQL DSN 구성
		if dbURL == "test.db" { // 기본값이면 환경변수로 DSN 구성
			dbURL = fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=UTC",
				dbHost, dbUser, dbPassword, dbName, dbPort, dbSSLMode)
		}
		db, err = gorm.Open(postgres.Open(dbURL), &gorm.Config{})
	case "sqlite":
		db, err = gorm.Open(sqlite.Open(dbURL), &gorm.Config{})
	default:
		log.Fatalf("❌ 지원하지 않는 DB 타입: %s", dbType)
	}

	if err != nil {
		log.Fatalf("❌ 데이터베이스 연결 실패: %v", err)
	}

	fmt.Println("✅ 데이터베이스 연결 성공")

	// 테이블 마이그레이션
	err = db.AutoMigrate(
		&models.User{},
		&models.UserWallet{},
		&models.Project{},
		&models.Milestone{},
		&models.Order{},
		&models.Trade{},
		&models.Position{},
		&models.MarketData{},
		&models.PriceHistory{},
	)
	if err != nil {
		log.Fatalf("❌ 테이블 마이그레이션 실패: %v", err)
	}

	fmt.Println("✅ 테이블 마이그레이션 완료")

	// 기존 테스트 계정 확인 및 삭제
	var existingCount int64
	db.Model(&models.User{}).Where("username LIKE ? OR username LIKE ?", "testuser_%", "loadtest_user_%").Count(&existingCount)

	if existingCount > 0 {
		fmt.Printf("🔍 기존 테스트 계정 %d개 발견\n", existingCount)

		if getEnv("CLEAN_EXISTING", "false") == "true" {
			fmt.Println("🧹 기존 테스트 계정 삭제 중...")

			// 관련된 지갑 먼저 완전 삭제 (외래키 제약 때문에)
			db.Unscoped().Where("user_id IN (SELECT id FROM users WHERE username LIKE ? OR username LIKE ?)",
				"testuser_%", "loadtest_user_%").Delete(&models.UserWallet{})

			// 테스트 계정 완전 삭제 (soft delete 무시)
			db.Unscoped().Where("username LIKE ? OR username LIKE ?", "testuser_%", "loadtest_user_%").Delete(&models.User{})

			fmt.Println("✅ 기존 테스트 계정 삭제 완료")
		} else {
			fmt.Println("💡 기존 계정을 삭제하려면 CLEAN_EXISTING=true 환경변수를 설정하세요")
			fmt.Println("   또는 'make recreate-test-accounts' 명령어를 사용하세요")
			fmt.Println("🔄 기존 계정을 건너뛰고 새 계정만 생성합니다...")
		}
	}

	// 테스트 사용자 생성
	fmt.Printf("👥 %d명의 테스트 사용자 생성 중...\n", numUsers)

	createdUsers := 0
	createdWallets := 0

	for i := 1; i <= numUsers; i++ {
		username := fmt.Sprintf("testuser_%d", i)
		email := fmt.Sprintf("testuser%d@example.com", i)

		// 기존 사용자 확인 (soft delete 무시하고 실제 존재 여부 확인)
		var existingUser models.User
		err := db.Unscoped().Where("username = ?", username).First(&existingUser).Error

		if err == nil {
			// 사용자가 이미 존재함 - 지갑만 확인/생성
			fmt.Printf("   사용자 %d 이미 존재 (ID: %d), 지갑 확인 중...\n", i, existingUser.ID)

			var existingWallet models.UserWallet
			walletErr := db.Where("user_id = ?", existingUser.ID).First(&existingWallet).Error

			if walletErr != nil {
				// 지갑이 없으면 생성
				wallet := models.UserWallet{
					UserID:      existingUser.ID,
					USDCBalance: usdcBalance,
				}
				if walletCreateErr := db.Create(&wallet).Error; walletCreateErr == nil {
					createdWallets++
					fmt.Printf("   ✅ 사용자 %d 지갑 생성 완료\n", i)
				} else {
					fmt.Printf("   ⚠️  사용자 %d 지갑 생성 실패: %v\n", i, walletCreateErr)
				}
			} else {
				fmt.Printf("   ℹ️  사용자 %d 지갑 이미 존재 (잔액: $%.2f)\n", i, float64(existingWallet.USDCBalance)/100)
			}
			continue
		}

		// 새 사용자 생성
		user := models.User{
			Username: username,
			Email:    email,
		}

		err = db.Create(&user).Error
		if err != nil {
			fmt.Printf("⚠️  사용자 %d 생성 실패: %v\n", i, err)
			continue
		}
		createdUsers++

		// 지갑 생성
		wallet := models.UserWallet{
			UserID:      user.ID,
			USDCBalance: usdcBalance,
		}

		err = db.Create(&wallet).Error
		if err != nil {
			fmt.Printf("⚠️  사용자 %d 지갑 생성 실패: %v\n", i, err)
			continue
		}
		createdWallets++

		if i%10 == 0 {
			fmt.Printf("   진행률: %d/%d (%.1f%%)\n", i, numUsers, float64(i)/float64(numUsers)*100)
		}
	}

	// 테스트 프로젝트 및 마일스톤 생성
	fmt.Println("📁 테스트 프로젝트 생성 중...")

	project := models.Project{
		Title:       "Test Project",
		Description: "테스트용 프로젝트",
		UserID:      1, // 첫 번째 사용자가 소유
		Status:      "active",
	}
	db.Create(&project)

	milestone := models.Milestone{
		ProjectID:   project.ID,
		Title:       "Test Milestone",
		Description: "테스트용 마일스톤",
		Status:      "funding",
		Order:       1,
	}
	db.Create(&milestone)

	// 플랫폼 수수료 설정
	feeConfig := models.PlatformFeeConfig{
		TradingFeeRate:    0.02,    // 2%
		WithdrawFeeFlat:   200,     // $2
		MinBetAmount:      1000,    // $10
		MaxBetAmount:      5000000, // $50,000
		StakingRewardRate: 0.70,
	}
	db.Create(&feeConfig)

	// 최종 통계 확인
	var totalUsers, totalWallets int64
	db.Model(&models.User{}).Where("username LIKE ?", "testuser_%").Count(&totalUsers)
	db.Model(&models.UserWallet{}).Joins("JOIN users ON users.id = user_wallets.user_id").
		Where("users.username LIKE ?", "testuser_%").Count(&totalWallets)

	// 결과 출력
	fmt.Printf("\n🎉 테스트 계정 생성 완료!\n")
	fmt.Printf("   🆕 새로 생성된 사용자: %d\n", createdUsers)
	fmt.Printf("   🆕 새로 생성된 지갑: %d\n", createdWallets)
	fmt.Printf("   📊 총 테스트 사용자: %d\n", totalUsers)
	fmt.Printf("   📊 총 테스트 지갑: %d\n", totalWallets)
	fmt.Printf("   ✅ 테스트 프로젝트 ID: %d\n", project.ID)
	fmt.Printf("   ✅ 테스트 마일스톤 ID: %d\n", milestone.ID)

	// 생성된 계정 샘플 출력
	var sampleUsers []models.User
	db.Where("username LIKE ?", "testuser_%").Limit(5).Find(&sampleUsers)

	fmt.Printf("\n📋 생성된 계정 샘플:\n")
	for _, user := range sampleUsers {
		var wallet models.UserWallet
		db.Where("user_id = ?", user.ID).First(&wallet)
		fmt.Printf("   - ID: %d, Username: %s, Email: %s, USDC: $%.2f\n",
			user.ID, user.Username, user.Email, float64(wallet.USDCBalance)/100)
	}

	fmt.Printf("\n💡 사용 방법:\n")
	fmt.Printf("   - API 테스트: curl -X POST /api/v1/trading/orders -d '{\"user_id\": 1, ...}'\n")
	fmt.Printf("   - 부하 테스트: go test ./tests/load -v\n")
	fmt.Printf("   - 계정 삭제: CLEAN_EXISTING=true go run scripts/create_test_accounts.go\n")
}

// getEnv 환경변수 읽기 (기본값 지원)
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvInt 환경변수를 정수로 읽기
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getEnvInt64 환경변수를 int64로 읽기
func getEnvInt64(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.ParseInt(value, 10, 64); err == nil {
			return intValue
		}
	}
	return defaultValue
}
