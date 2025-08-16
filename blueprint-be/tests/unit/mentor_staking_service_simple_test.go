package unit

import (
	"testing"
	"time"

	"blueprint-module/pkg/models"
	"blueprint/internal/services"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type MentorStakingSimpleTestSuite struct {
	suite.Suite
	db                   *gorm.DB
	mentorStakingService *services.MentorStakingService
}

func (suite *MentorStakingSimpleTestSuite) SetupSuite() {
	// 테스트용 PostgreSQL 데이터베이스 설정
	db, err := gorm.Open(postgres.Open("host=localhost user=test password=test dbname=test_blueprint port=5432 sslmode=disable"), &gorm.Config{})
	if err != nil {
		// 포스트그레스 연결 실패시 테스트 스킵
		suite.T().Skip("PostgreSQL 테스트 데이터베이스에 연결할 수 없습니다")
		return
	}
	
	suite.db = db
	suite.mentorStakingService = services.NewMentorStakingService(suite.db)
	
	// 테스트에 필요한 테이블 생성
	suite.db.AutoMigrate(
		&models.User{},
		&models.UserWallet{},
		&models.Mentor{},
		&models.MentorStake{},
		&models.MentorSlashEvent{},
		&models.MentorPerformanceMetric{},
		&models.MentorStakeReward{},
	)
}

func (suite *MentorStakingSimpleTestSuite) TearDownSuite() {
	if suite.db != nil {
		// 테스트 테이블 정리
		suite.db.Exec("DROP TABLE IF EXISTS mentor_stake_rewards CASCADE")
		suite.db.Exec("DROP TABLE IF EXISTS mentor_performance_metrics CASCADE")
		suite.db.Exec("DROP TABLE IF EXISTS mentor_slash_events CASCADE")
		suite.db.Exec("DROP TABLE IF EXISTS mentor_stakes CASCADE")
		suite.db.Exec("DROP TABLE IF EXISTS mentors CASCADE")
		suite.db.Exec("DROP TABLE IF EXISTS user_wallets CASCADE")
		suite.db.Exec("DROP TABLE IF EXISTS users CASCADE")
	}
}

func (suite *MentorStakingSimpleTestSuite) SetupTest() {
	if suite.db == nil {
		return
	}
	
	// 각 테스트 전에 테이블 데이터 초기화
	suite.db.Exec("TRUNCATE TABLE mentor_stake_rewards, mentor_performance_metrics, mentor_slash_events, mentor_stakes, mentors, user_wallets, users RESTART IDENTITY CASCADE")
	
	// 테스트 데이터 시드
	user1 := &models.User{ID: 1, Email: "mentor@test.com"}
	user2 := &models.User{ID: 2, Email: "staker@test.com"}
	suite.db.Create(user1)
	suite.db.Create(user2)
	
	// 멘토 생성
	mentor := &models.Mentor{
		ID:     1,
		UserID: 1,
	}
	suite.db.Create(mentor)
}

func (suite *MentorStakingSimpleTestSuite) TestStakeMentorSuccess() {
	// 1. 테스트 데이터 준비
	mentorID := uint(1)
	stakerID := uint(2)
	
	// 스테이커 지갑 생성
	stakerWallet := &models.UserWallet{
		UserID:           stakerID,
		BlueprintBalance: 100000, // 충분한 잔액
		USDCBalance:      50000,
	}
	suite.db.Create(stakerWallet)

	// 2. 스테이킹 요청 생성
	req := &models.StakeMentorRequest{
		MentorID:      mentorID,
		Amount:        10000,
		StakeType:     models.MentorStakeTypeSelf,
		Purpose:       models.MentorStakePurposeQualification,
		MinimumPeriod: 30, // 30일
		IsAutoRenewal: false,
	}

	// 3. 멘토 스테이킹
	stake, err := suite.mentorStakingService.StakeMentor(req, stakerID)

	// 4. 검증
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), stake)
	assert.Equal(suite.T(), mentorID, stake.MentorID)
	assert.Equal(suite.T(), stakerID, stake.UserID)
	assert.Equal(suite.T(), req.Amount, stake.Amount)
	assert.Equal(suite.T(), models.MentorStakeStatusActive, stake.Status)
	
	// 스테이커 지갑에서 스테이킹 금액이 차감되었는지 확인
	var updatedWallet models.UserWallet
	suite.db.Where("user_id = ?", stakerID).First(&updatedWallet)
	assert.Equal(suite.T(), int64(90000), updatedWallet.BlueprintBalance)
}

func (suite *MentorStakingSimpleTestSuite) TestUnstakeMentorSuccess() {
	// 1. 테스트 데이터 준비
	mentorID := uint(1)
	stakerID := uint(2)
	
	// 기존 스테이킹 생성 (잠금 해제된 상태)
	stake := &models.MentorStake{
		ID:              1,
		MentorID:        mentorID,
		UserID:          stakerID,
		Amount:          10000,
		AvailableAmount: 10000,
		Status:          models.MentorStakeStatusActive,
		UnlockDate:      time.Now().AddDate(0, 0, -1), // 어제 해제
	}
	suite.db.Create(stake)

	// 스테이커 지갑 생성
	stakerWallet := &models.UserWallet{
		UserID:           stakerID,
		BlueprintBalance: 90000,
	}
	suite.db.Create(stakerWallet)

	// 2. 스테이킹 해제
	err := suite.mentorStakingService.UnstakeMentor(stake.ID, stakerID)

	// 3. 검증
	assert.NoError(suite.T(), err)
	
	// 스테이킹 상태가 변경되었는지 확인
	var updatedStake models.MentorStake
	suite.db.First(&updatedStake, stake.ID)
	assert.Equal(suite.T(), models.MentorStakeStatusWithdrawn, updatedStake.Status)
	assert.Equal(suite.T(), int64(0), updatedStake.AvailableAmount)
	
	// 지갑에 금액이 반환되었는지 확인
	var updatedWallet models.UserWallet
	suite.db.Where("user_id = ?", stakerID).First(&updatedWallet)
	assert.Equal(suite.T(), int64(100000), updatedWallet.BlueprintBalance)
}

func (suite *MentorStakingSimpleTestSuite) TestGetUserStakesSuccess() {
	// 1. 테스트 데이터 준비
	mentorID := uint(1)
	stakerID := uint(2)
	
	stake := &models.MentorStake{
		MentorID:        mentorID,
		UserID:          stakerID,
		Amount:          10000,
		AvailableAmount: 10000,
		Status:          models.MentorStakeStatusActive,
	}
	suite.db.Create(stake)

	// 2. 사용자 스테이킹 목록 조회
	result, err := suite.mentorStakingService.GetUserStakes(stakerID, 1, 10, "", "")

	// 3. 검증
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
}

func TestMentorStakingSimpleTestSuite(t *testing.T) {
	suite.Run(t, new(MentorStakingSimpleTestSuite))
}