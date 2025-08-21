package services

import (
	"log"
	"sort"

	"blueprint-module/pkg/models"
	"blueprint/internal/database"

	"gorm.io/gorm"
)

type JuryService struct {
	db *gorm.DB
}

func NewJuryService() *JuryService {
	return &JuryService{
		db: database.GetDB(),
	}
}

// 👥 전문가 판결단 구성 (Tier 1)
func (js *JuryService) FormExpertJury(disputeID uint, milestoneID uint) error {
	log.Printf("👥 Forming expert jury for dispute %d (milestone %d)", disputeID, milestoneID)

	// 해당 마일스톤에 투자한 사용자들 조회
	// TODO: Investment 모델이 없으므로 임시로 mock 데이터 사용
	successInvestors := js.getTopInvestorsMock("success", 5)
	failInvestors := js.getTopInvestorsMock("fail", 5)

	// 판결단 구성
	var juryMembers []models.DisputeJury

	// 성공에 투자한 상위 5명
	for _, investor := range successInvestors {
		juryMember := models.DisputeJury{
			DisputeID:        disputeID,
			JurorID:          investor.UserID,
			Position:         "success_investor",
			InvestmentAmount: investor.Amount,
			HasVoted:         false,
		}
		juryMembers = append(juryMembers, juryMember)
	}

	// 실패에 투자한 상위 5명
	for _, investor := range failInvestors {
		juryMember := models.DisputeJury{
			DisputeID:        disputeID,
			JurorID:          investor.UserID,
			Position:         "fail_investor",
			InvestmentAmount: investor.Amount,
			HasVoted:         false,
		}
		juryMembers = append(juryMembers, juryMember)
	}

	// 판결단 저장
	for _, member := range juryMembers {
		if err := js.db.Create(&member).Error; err != nil {
			log.Printf("❌ Failed to create jury member: %v", err)
			return err
		}
	}

	log.Printf("✅ Expert jury formed: %d members", len(juryMembers))
	return nil
}

// Mock 투자자 데이터 (실제로는 Investment 테이블에서 조회)
type MockInvestor struct {
	UserID uint
	Amount int64
}

func (js *JuryService) getTopInvestorsMock(option string, limit int) []MockInvestor {
	// TODO: 실제 Investment 모델에서 조회하도록 변경
	investors := []MockInvestor{
		{UserID: 1, Amount: 50000},
		{UserID: 2, Amount: 40000},
		{UserID: 3, Amount: 30000},
		{UserID: 4, Amount: 20000},
		{UserID: 5, Amount: 10000},
	}

	// 투자액으로 정렬
	sort.Slice(investors, func(i, j int) bool {
		return investors[i].Amount > investors[j].Amount
	})

	if len(investors) > limit {
		return investors[:limit]
	}
	return investors
}
