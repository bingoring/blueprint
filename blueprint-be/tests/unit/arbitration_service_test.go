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
	"github.com/gin-gonic/gin"
)

type ArbitrationServiceTestSuite struct {
	suite.Suite
	db                 *gorm.DB
	arbitrationService *services.ArbitrationService
}

func (suite *ArbitrationServiceTestSuite) SetupSuite() {
	// 테스트용 인메모리 SQLite 데이터베이스 설정
	db, err := gorm.Open(postgres.Open("host=localhost user=test password=test dbname=test_blueprint port=5432 sslmode=disable"), &gorm.Config{})
	if err != nil {
		// 포스트그레스 연결 실패시 테스트 스킵
		suite.T().Skip("PostgreSQL 테스트 데이터베이스에 연결할 수 없습니다")
		return
	}
	
	suite.db = db
	suite.arbitrationService = services.NewArbitrationService(suite.db)
	
	// 테스트에 필요한 테이블 생성
	suite.db.AutoMigrate(
		&models.User{},
		&models.UserWallet{},
		&models.ArbitrationCase{},
		&models.ArbitrationVote{},
		&models.JurorQualification{},
		&models.ArbitrationReward{},
	)
}

func (suite *ArbitrationServiceTestSuite) TearDownSuite() {
	if suite.db != nil {
		// 테스트 테이블 정리
		suite.db.Exec("DROP TABLE IF EXISTS arbitration_rewards CASCADE")
		suite.db.Exec("DROP TABLE IF EXISTS arbitration_votes CASCADE")
		suite.db.Exec("DROP TABLE IF EXISTS arbitration_cases CASCADE")
		suite.db.Exec("DROP TABLE IF EXISTS juror_qualifications CASCADE")
		suite.db.Exec("DROP TABLE IF EXISTS user_wallets CASCADE")
		suite.db.Exec("DROP TABLE IF EXISTS users CASCADE")
	}
}

func (suite *ArbitrationServiceTestSuite) SetupTest() {
	if suite.db == nil {
		return
	}
	
	// 각 테스트 전에 테이블 데이터 초기화
	suite.db.Exec("TRUNCATE TABLE arbitration_rewards, arbitration_votes, arbitration_cases, juror_qualifications, user_wallets, users RESTART IDENTITY CASCADE")
	
	// 테스트 데이터 시드
	user1 := &models.User{ID: 1, Email: "plaintiff@test.com"}
	user2 := &models.User{ID: 2, Email: "defendant@test.com"}
	suite.db.Create(user1)
	suite.db.Create(user2)
}

func (suite *ArbitrationServiceTestSuite) TestSubmitCase() {
	// 1. 테스트 데이터 준비
	plaintiffID := uint(1)
	defendantID := uint(2)
	
	// 사용자 지갑 생성
	userWallet := &models.UserWallet{
		UserID:            plaintiffID,
		BlueprintBalance:  100000, // 충분한 잔액
		USDCBalance:       50000,
	}
	suite.db.Create(userWallet)

	// 2. 분쟁 제기 요청 생성
	req := &models.SubmitArbitrationRequest{
		DefendantID:   defendantID,
		DisputeType:   models.DisputeTypeMilestoneCompletion,
		Title:         "마일스톤 완료 분쟁",
		Description:   "마일스톤이 완료되지 않았습니다",
		Evidence:      "증거 자료",
		ClaimedAmount: 50000,
		StakeAmount:   5000,
	}

	// 3. 분쟁 사건 제기
	arbitrationCase, err := suite.arbitrationService.SubmitCase(req, plaintiffID)

	// 4. 검증
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), arbitrationCase)
	assert.Equal(suite.T(), plaintiffID, arbitrationCase.PlaintiffID)
	assert.Equal(suite.T(), defendantID, arbitrationCase.DefendantID)
	assert.Equal(suite.T(), models.ArbitrationStatusSubmitted, arbitrationCase.Status)
	assert.Equal(suite.T(), req.StakeAmount, arbitrationCase.StakeAmount)
	assert.True(suite.T(), arbitrationCase.RequiredJurors >= 5)
	
	// 사용자 지갑에서 스테이킹 금액이 차감되었는지 확인
	var updatedWallet models.UserWallet
	suite.db.Where("user_id = ?", plaintiffID).First(&updatedWallet)
	assert.Equal(suite.T(), int64(95000), updatedWallet.BlueprintBalance)
	assert.Equal(suite.T(), int64(5000), updatedWallet.BlueprintLockedBalance)
}

func (suite *ArbitrationServiceTestSuite) TestSubmitCaseInsufficientBalance() {
	// 1. 테스트 데이터 준비 (잔액 부족)
	plaintiffID := uint(1)
	defendantID := uint(2)
	
	userWallet := &models.UserWallet{
		UserID:            plaintiffID,
		BlueprintBalance:  1000, // 부족한 잔액
		USDCBalance:       0,
	}
	suite.db.Create(userWallet)

	// 2. 분쟁 제기 요청 생성
	req := &models.SubmitArbitrationRequest{
		DefendantID:   defendantID,
		DisputeType:   models.DisputeTypeMilestoneCompletion,
		Title:         "테스트 분쟁",
		Description:   "테스트 설명",
		StakeAmount:   5000, // 잔액보다 많은 스테이킹
	}

	// 3. 분쟁 사건 제기 (실패해야 함)
	arbitrationCase, err := suite.arbitrationService.SubmitCase(req, plaintiffID)

	// 4. 검증
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), arbitrationCase)
	assert.Contains(suite.T(), err.Error(), "BLUEPRINT 잔액이 부족합니다")
}

func (suite *ArbitrationServiceTestSuite) TestRegisterJuror() {
	// 1. 테스트 데이터 준비
	userID := uint(1)
	
	userWallet := &models.UserWallet{
		UserID:            userID,
		BlueprintBalance:  100000,
		USDCBalance:       0,
	}
	suite.db.Create(userWallet)

	// 2. 배심원 등록 요청
	req := &struct {
		MinStakeAmount  int64    `json:"min_stake_amount"`
		ExpertiseAreas  []string `json:"expertise_areas"`
		LanguageSkills  []string `json:"language_skills"`
		LegalBackground bool     `json:"legal_background"`
	}{
		MinStakeAmount:  10000,
		ExpertiseAreas:  []string{"technology", "business"},
		LanguageSkills:  []string{"korean", "english"},
		LegalBackground: false,
	}

	// 3. 배심원 등록
	qualification, err := suite.arbitrationService.RegisterJuror(userID, req)

	// 4. 검증
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), qualification)
	assert.Equal(suite.T(), userID, qualification.UserID)
	assert.Equal(suite.T(), req.MinStakeAmount, qualification.MinStakeAmount)
	assert.Equal(suite.T(), req.MinStakeAmount, qualification.CurrentStake)
	assert.Equal(suite.T(), req.ExpertiseAreas, qualification.ExpertiseAreas)
	assert.Equal(suite.T(), req.LanguageSkills, qualification.LanguageSkills)
	assert.True(suite.T(), qualification.IsActive)
	assert.Equal(suite.T(), 0.5, qualification.ReputationScore)
}

func (suite *ArbitrationServiceTestSuite) TestCommitVote() {
	// 1. 테스트 데이터 준비
	jurorID := uint(1)
	
	// 분쟁 사건 생성
	arbitrationCase := &models.ArbitrationCase{
		CaseNumber:      "ACC-2024-0001",
		PlaintiffID:     2,
		DefendantID:     3,
		DisputeType:     models.DisputeTypeMilestoneCompletion,
		Title:           "테스트 분쟁",
		Description:     "테스트 설명",
		Status:          models.ArbitrationStatusVoting,
		RequiredJurors:  5,
		SelectedJurors:  []uint{jurorID, 4, 5, 6, 7},
		VotingDeadline:  &[]time.Time{time.Now().Add(24 * time.Hour)}[0],
	}
	suite.db.Create(arbitrationCase)

	// 배심원 자격 생성
	qualification := &models.JurorQualification{
		UserID:          jurorID,
		CurrentStake:    10000,
		ReputationScore: 0.8,
		IsActive:        true,
	}
	suite.db.Create(qualification)

	// 2. 투표 요청 생성
	req := &models.JurorVoteRequest{
		CaseID:     arbitrationCase.ID,
		CommitHash: "abc123def456", // 테스트용 해시
	}

	// 3. 투표 제출
	vote, err := suite.arbitrationService.CommitVote(req, jurorID)

	// 4. 검증
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), vote)
	assert.Equal(suite.T(), arbitrationCase.ID, vote.CaseID)
	assert.Equal(suite.T(), jurorID, vote.JurorID)
	assert.Equal(suite.T(), req.CommitHash, vote.CommitHash)
	assert.Equal(suite.T(), qualification.CurrentStake, vote.JurorStake)
	assert.Equal(suite.T(), qualification.ReputationScore, vote.QualificationScore)
	assert.NotNil(suite.T(), vote.CommittedAt)
}

func (suite *ArbitrationServiceTestSuite) TestRevealVote() {
	// 1. 테스트 데이터 준비
	jurorID := uint(1)
	
	// 분쟁 사건 생성
	arbitrationCase := &models.ArbitrationCase{
		ID:              1,
		Status:          models.ArbitrationStatusReveal,
		RequiredJurors:  5,
		SelectedJurors:  []uint{jurorID},
		RevealDeadline:  &[]time.Time{time.Now().Add(24 * time.Hour)}[0],
	}
	suite.db.Create(arbitrationCase)

	// 기존 투표 생성
	vote := &models.ArbitrationVote{
		CaseID:     arbitrationCase.ID,
		JurorID:    jurorID,
		CommitHash: "a665a45920422f9d417e4867efdc4fb8a04a1f3fff1fa07e998e86f7f7a27ae3", // SHA256("hello" + "salt123")
		CommittedAt: &[]time.Time{time.Now()}[0],
	}
	suite.db.Create(vote)

	// 2. 투표 공개 요청
	req := &models.RevealVoteRequest{
		CaseID:     arbitrationCase.ID,
		Vote:       models.ArbitrationDecisionPlaintiffWins,
		Salt:       "salt123",
		VoteReason: "증거가 충분합니다",
	}

	// 3. 투표 공개
	err := suite.arbitrationService.RevealVote(req, jurorID)

	// 4. 검증
	assert.NoError(suite.T(), err)
	
	// 업데이트된 투표 조회
	var updatedVote models.ArbitrationVote
	suite.db.First(&updatedVote, vote.ID)
	assert.Equal(suite.T(), req.Vote, *updatedVote.RevealedVote)
	assert.Equal(suite.T(), req.Salt, updatedVote.RevealedSalt)
	assert.Equal(suite.T(), req.VoteReason, updatedVote.VoteReason)
	assert.NotNil(suite.T(), updatedVote.RevealedAt)
}

func (suite *ArbitrationServiceTestSuite) TestGetArbitrationStats() {
	// 1. 테스트 데이터 준비
	now := time.Now()
	
	// 몇 개의 분쟁 사건 생성
	cases := []models.ArbitrationCase{
		{
			CaseNumber: "ACC-2024-0001",
			Status:     models.ArbitrationStatusDecided,
			CreatedAt:  now.AddDate(0, 0, -5),
			DecidedAt:  &[]time.Time{now.AddDate(0, 0, -3)}[0],
		},
		{
			CaseNumber: "ACC-2024-0002",
			Status:     models.ArbitrationStatusVoting,
			CreatedAt:  now.AddDate(0, 0, -3),
		},
		{
			CaseNumber: "ACC-2024-0003",
			Status:     models.ArbitrationStatusDecided,
			CreatedAt:  now.AddDate(0, 0, -1),
			DecidedAt:  &[]time.Time{now}[0],
		},
	}
	
	for _, c := range cases {
		suite.db.Create(&c)
	}

	// 2. 통계 조회
	stats, err := suite.arbitrationService.GetArbitrationStats("weekly")

	// 3. 검증
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), stats)
	
	statsMap := stats.(gin.H)
	assert.Equal(suite.T(), "weekly", statsMap["period"])
	assert.Equal(suite.T(), int64(3), statsMap["total_cases"])
	assert.Equal(suite.T(), int64(2), statsMap["resolved_cases"])
	assert.Equal(suite.T(), int64(1), statsMap["pending_cases"])
}

func TestArbitrationServiceTestSuite(t *testing.T) {
	suite.Run(t, new(ArbitrationServiceTestSuite))
}