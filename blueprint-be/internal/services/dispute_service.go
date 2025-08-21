package services

import (
	"fmt"
	"log"
	"time"

	"blueprint-module/pkg/models"
	"blueprint/internal/database"

	"gorm.io/gorm"
)

type DisputeService struct {
	db          *gorm.DB
	sseService  *SSEService
	juryService *JuryService
}

func NewDisputeService(sseService *SSEService, juryService *JuryService) *DisputeService {
	return &DisputeService{
		db:          database.GetDB(),
		sseService:  sseService,
		juryService: juryService,
	}
}

// 🏛️ 마일스톤 결과 보고 (생성자가 성공/실패 선언)
func (ds *DisputeService) ReportMilestoneResult(
	milestoneID uint,
	reporterID uint,
	result bool,
	evidenceURL string,
	evidenceFiles []string,
	description string,
) error {
	log.Printf("🏛️ Blueprint Court: Reporting milestone %d result - %v", milestoneID, result)

	// 마일스톤 존재 확인
	var milestone models.Milestone
	if err := ds.db.Where("id = ?", milestoneID).First(&milestone).Error; err != nil {
		return fmt.Errorf("milestone not found: %w", err)
	}

	// 이미 보고된 경우 체크
	var existingResult models.MilestoneResult
	if err := ds.db.Where("milestone_id = ?", milestoneID).First(&existingResult).Error; err == nil {
		return fmt.Errorf("milestone result already reported")
	}

	// 트랜잭션으로 처리
	return ds.db.Transaction(func(tx *gorm.DB) error {
		// 결과 보고 생성
		evidenceFilesJSON := ""
		if len(evidenceFiles) > 0 {
			// JSON으로 직렬화 (실제로는 json.Marshal 사용)
			evidenceFilesJSON = fmt.Sprintf(`["%s"]`, evidenceFiles[0])
		}

		milestoneResult := models.MilestoneResult{
			MilestoneID:   milestoneID,
			ReporterID:    reporterID,
			Result:        result,
			EvidenceURL:   evidenceURL,
			EvidenceFiles: evidenceFilesJSON,
			Description:   description,
			IsDisputed:    false,
			IsFinal:       false,
		}

		if err := tx.Create(&milestoneResult).Error; err != nil {
			return fmt.Errorf("failed to create milestone result: %w", err)
		}

		// 마일스톤 상태 업데이트
		now := time.Now()
		if err := tx.Model(&milestone).Updates(map[string]interface{}{
			"result_reported":    true,
			"result_reported_at": now,
		}).Error; err != nil {
			return fmt.Errorf("failed to update milestone: %w", err)
		}

		// 48시간 이의 제기 창 시작
		ds.startChallengeWindow(milestoneID, now)

		// 투자자들에게 알림 전송
		ds.notifyInvestors(milestoneID, result)

		log.Printf("✅ Milestone %d result reported successfully", milestoneID)
		return nil
	})
}

// ⚔️ 이의 제기 생성
func (ds *DisputeService) CreateDispute(milestoneID uint, challengerID uint, reason string) error {
	log.Printf("⚔️ Blueprint Court: Creating dispute for milestone %d by user %d", milestoneID, challengerID)

	// 유효성 검사
	if len(reason) < 100 {
		return fmt.Errorf("dispute reason must be at least 100 characters")
	}

	// 마일스톤 결과 확인
	var milestoneResult models.MilestoneResult
	if err := ds.db.Where("milestone_id = ? AND is_final = false", milestoneID).First(&milestoneResult).Error; err != nil {
		return fmt.Errorf("no disputable milestone result found: %w", err)
	}

	// 이미 분쟁 중인지 확인
	if milestoneResult.IsDisputed {
		return fmt.Errorf("milestone is already in dispute")
	}

	// 투자자 자격 확인 (1 USDC 이상 투자) - 임시로 true로 설정
	// TODO: Investment 모델 구현 후 실제 투자 여부 확인
	hasInvestment := true
	if !hasInvestment {
		return fmt.Errorf("challenger must have invested at least 1 USDC in this milestone")
	}

	// 총 투자액 계산 (심급 결정용) - 임시 값 사용
	// TODO: Investment 모델 구현 후 실제 투자액 계산
	var totalInvestmentAmount int64 = 50000 // 임시로 $500 설정

	// 심급 결정 (10,000 USDC = 1,000,000 센트)
	tier := models.DisputeTierExpert
	if totalInvestmentAmount >= 1000000 {
		tier = models.DisputeTierGovernance
	}

	return ds.db.Transaction(func(tx *gorm.DB) error {
		// 분쟁 생성
		challengeWindowEnd := time.Now().Add(48 * time.Hour)
		dispute := models.Dispute{
			MilestoneID:           milestoneID,
			ChallengerID:          challengerID,
			OriginalResult:        milestoneResult.Result,
			DisputeReason:         reason,
			StakeAmount:           10000, // 100 $BLUEPRINT (센트 단위)
			Status:                models.DisputeStatusChallengeWindow,
			Tier:                  tier,
			TotalInvestmentAmount: totalInvestmentAmount,
			ChallengeWindowEnd:    challengeWindowEnd,
		}

		if err := tx.Create(&dispute).Error; err != nil {
			return fmt.Errorf("failed to create dispute: %w", err)
		}

		// 예치금 기록
		stake := models.DisputeStake{
			DisputeID: dispute.ID,
			UserID:    challengerID,
			Amount:    10000,
		}
		if err := tx.Create(&stake).Error; err != nil {
			return fmt.Errorf("failed to create stake: %w", err)
		}

		// 마일스톤 결과 및 마일스톤 상태 업데이트
		if err := tx.Model(&milestoneResult).Update("is_disputed", true).Error; err != nil {
			return err
		}

		var milestone models.Milestone
		if err := tx.Where("id = ?", milestoneID).First(&milestone).Error; err != nil {
			return err
		}

		if err := tx.Model(&milestone).Updates(map[string]interface{}{
			"is_in_dispute":  true,
			"dispute_count":  milestone.DisputeCount + 1,
		}).Error; err != nil {
			return err
		}

		// Tier별 후속 처리
		if tier == models.DisputeTierExpert {
			// 전문가 판결단 구성
			ds.juryService.FormExpertJury(dispute.ID, milestoneID)
		}

		// 투표 기간 시작 준비
		ds.prepareVotingPeriod(dispute.ID)

		log.Printf("✅ Dispute %d created successfully (Tier: %s)", dispute.ID, tier)
		return nil
	})
}

// 🗳️ 투표 제출
func (ds *DisputeService) SubmitVote(disputeID uint, voterID uint, choice models.VoteChoice) error {
	log.Printf("🗳️ Blueprint Court: User %d voting %s on dispute %d", voterID, choice, disputeID)

	// 분쟁 정보 확인
	var dispute models.Dispute
	if err := ds.db.Where("id = ? AND status = ?", disputeID, models.DisputeStatusVotingPeriod).First(&dispute).Error; err != nil {
		return fmt.Errorf("dispute not found or not in voting period: %w", err)
	}

	// 이미 투표했는지 확인
	var existingVote models.DisputeVote
	if err := ds.db.Where("dispute_id = ? AND voter_id = ?", disputeID, voterID).First(&existingVote).Error; err == nil {
		return fmt.Errorf("user has already voted on this dispute")
	}

	// 투표 자격 확인 (심급별로 다름)
	var vote models.DisputeVote
	if dispute.Tier == models.DisputeTierExpert {
		// 전문가 판결: 판결단 멤버인지 확인
		var juryMember models.DisputeJury
		if err := ds.db.Where("dispute_id = ? AND juror_id = ?", disputeID, voterID).First(&juryMember).Error; err != nil {
			return fmt.Errorf("user is not a jury member for this dispute")
		}

		vote = models.DisputeVote{
			DisputeID:        disputeID,
			VoterID:          voterID,
			Choice:           choice,
			InvestmentAmount: juryMember.InvestmentAmount,
		}

		// 판결단 투표 상태 업데이트
		ds.db.Model(&juryMember).Update("has_voted", true)

	} else {
		// DAO 거버넌스: 토큰 보유량 확인
		// TODO: 실제 토큰 잔액 확인 로직 구현
		tokenAmount := int64(1000) // 임시값

		vote = models.DisputeVote{
			DisputeID:   disputeID,
			VoterID:     voterID,
			Choice:      choice,
			TokenAmount: tokenAmount,
		}
	}

	// 투표 저장 및 집계 업데이트
	return ds.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&vote).Error; err != nil {
			return fmt.Errorf("failed to create vote: %w", err)
		}

		// 투표 집계 업데이트
		if choice == models.VoteChoiceMaintain {
			tx.Model(&dispute).Update("maintain_votes", gorm.Expr("maintain_votes + 1"))
		} else {
			tx.Model(&dispute).Update("overrule_votes", gorm.Expr("overrule_votes + 1"))
		}

		// SSE로 실시간 업데이트
		ds.broadcastVotingUpdate(disputeID)

		log.Printf("✅ Vote submitted successfully")
		return nil
	})
}

// ⚖️ 분쟁 해결 (투표 마감 후 자동 실행)
func (ds *DisputeService) ResolveDispute(disputeID uint) error {
	log.Printf("⚖️ Blueprint Court: Resolving dispute %d", disputeID)

	var dispute models.Dispute
	if err := ds.db.Preload("Stakes").Where("id = ?", disputeID).First(&dispute).Error; err != nil {
		return fmt.Errorf("dispute not found: %w", err)
	}

	// 투표 결과 집계
	maintainWins := dispute.MaintainVotes > dispute.OverruleVotes
	finalResult := dispute.OriginalResult // 기본값은 원래 결과
	status := models.DisputeStatusRejected

	if !maintainWins {
		// 분쟁 인용: 결과 뒤집기
		finalResult = !dispute.OriginalResult
		status = models.DisputeStatusUpheld
	}

	return ds.db.Transaction(func(tx *gorm.DB) error {
		// 분쟁 상태 업데이트
		if err := tx.Model(&dispute).Updates(map[string]interface{}{
			"status":       status,
			"final_result": finalResult,
		}).Error; err != nil {
			return err
		}

		// 마일스톤 결과 업데이트
		var milestoneResult models.MilestoneResult
		if err := tx.Where("milestone_id = ?", dispute.MilestoneID).First(&milestoneResult).Error; err != nil {
			return err
		}

		if err := tx.Model(&milestoneResult).Updates(map[string]interface{}{
			"result":      finalResult,
			"is_final":    true,
			"is_disputed": false,
		}).Error; err != nil {
			return err
		}

		// 마일스톤 상태 업데이트
		if err := tx.Model(&models.Milestone{}).Where("id = ?", dispute.MilestoneID).Updates(map[string]interface{}{
			"is_in_dispute":           false,
			"final_result_confirmed":  true,
		}).Error; err != nil {
			return err
		}

		// 예치금 처리
		if maintainWins {
			// 분쟁 기각: 제기자 예치금 몰수
			ds.forfeitStakes(tx, disputeID)
		} else {
			// 분쟁 인용: 제기자 예치금 반환
			ds.refundStakes(tx, disputeID)
		}

		// 보상 분배
		ds.distributeRewards(tx, disputeID, maintainWins)

		// 예측 시장 정산
		ds.settlePredictionMarket(dispute.MilestoneID, finalResult)

		// 알림 전송
		ds.notifyDisputeResolution(disputeID, status, finalResult)

		log.Printf("✅ Dispute %d resolved: %s (Final result: %v)", disputeID, status, finalResult)
		return nil
	})
}

// 🚨 투자자 알림
func (ds *DisputeService) notifyInvestors(milestoneID uint, result bool) {
	// TODO: 실제 알림 시스템 구현
	log.Printf("📢 Notifying investors about milestone %d result: %v", milestoneID, result)
}

// ⏰ 48시간 이의제기 창 시작
func (ds *DisputeService) startChallengeWindow(milestoneID uint, startTime time.Time) {
	// TODO: 타이머 서비스에 48시간 후 자동 확정 스케줄링
	endTime := startTime.Add(48 * time.Hour)
	log.Printf("⏰ Challenge window started for milestone %d (ends at: %v)", milestoneID, endTime)
}

// 🗳️ 투표 기간 준비
func (ds *DisputeService) prepareVotingPeriod(disputeID uint) {
	// TODO: 72시간 투표 기간 스케줄링
	log.Printf("🗳️ Preparing voting period for dispute %d", disputeID)
}

// 📊 실시간 투표 업데이트 전송
func (ds *DisputeService) broadcastVotingUpdate(disputeID uint) {
	if ds.sseService != nil {
		// TODO: SSE로 투표 현황 실시간 전송
		log.Printf("📊 Broadcasting voting update for dispute %d", disputeID)
	}
}

// 💰 예치금 몰수
func (ds *DisputeService) forfeitStakes(tx *gorm.DB, disputeID uint) error {
	return tx.Model(&models.DisputeStake{}).Where("dispute_id = ?", disputeID).Update("is_forfeited", true).Error
}

// 💸 예치금 반환
func (ds *DisputeService) refundStakes(tx *gorm.DB, disputeID uint) error {
	return tx.Model(&models.DisputeStake{}).Where("dispute_id = ?", disputeID).Update("is_refunded", true).Error
}

// 🎁 보상 분배
func (ds *DisputeService) distributeRewards(tx *gorm.DB, disputeID uint, maintainWins bool) {
	// TODO: 판결단/투표자들에게 보상 분배
	log.Printf("🎁 Distributing rewards for dispute %d (maintain wins: %v)", disputeID, maintainWins)
}

// 💱 예측 시장 정산
func (ds *DisputeService) settlePredictionMarket(milestoneID uint, finalResult bool) {
	// TODO: 매칭 엔진에 최종 결과 전달하여 시장 정산
	log.Printf("💱 Settling prediction market for milestone %d with result: %v", milestoneID, finalResult)
}

// 📢 분쟁 해결 알림
func (ds *DisputeService) notifyDisputeResolution(disputeID uint, status models.DisputeStatus, finalResult bool) {
	// TODO: 관련자들에게 해결 결과 알림
	log.Printf("📢 Notifying dispute resolution: %d (%s, result: %v)", disputeID, status, finalResult)
}

// 📈 분쟁 상세 정보 조회
func (ds *DisputeService) GetDisputeDetail(disputeID uint) (*models.DisputeDetailResponse, error) {
	var dispute models.Dispute
	if err := ds.db.Preload("Milestone").Preload("Challenger").Preload("Votes").Preload("JuryMembers").Where("id = ?", disputeID).First(&dispute).Error; err != nil {
		return nil, fmt.Errorf("dispute not found: %w", err)
	}

	var milestoneResult models.MilestoneResult
	if err := ds.db.Where("milestone_id = ?", dispute.MilestoneID).First(&milestoneResult).Error; err != nil {
		return nil, fmt.Errorf("milestone result not found: %w", err)
	}

	// 투표 통계 계산
	totalVoters := len(dispute.JuryMembers)
	if dispute.Tier == models.DisputeTierGovernance {
		// DAO 투표의 경우 토큰 보유자 수로 계산
		totalVoters = 1000 // TODO: 실제 토큰 보유자 수
	}

	votingStats := models.VotingStats{
		TotalVoters:    totalVoters,
		VotedCount:     len(dispute.Votes),
		MaintainVotes:  dispute.MaintainVotes,
		OverruleVotes:  dispute.OverruleVotes,
		VotingProgress: float64(len(dispute.Votes)) / float64(totalVoters),
	}

	// 남은 시간 계산
	timeRemaining := ds.calculateTimeRemaining(&dispute)

	return &models.DisputeDetailResponse{
		Dispute:         dispute,
		MilestoneResult: milestoneResult,
		JuryMembers:     dispute.JuryMembers,
		VotingStats:     votingStats,
		TimeRemaining:   timeRemaining,
	}, nil
}

// ⏰ 남은 시간 계산
func (ds *DisputeService) calculateTimeRemaining(dispute *models.Dispute) models.TimeRemaining {
	now := time.Now()
	var endTime time.Time
	var phase string

	switch dispute.Status {
	case models.DisputeStatusChallengeWindow:
		endTime = dispute.ChallengeWindowEnd
		phase = "challenge_window"
	case models.DisputeStatusVotingPeriod:
		if dispute.VotingPeriodEnd != nil {
			endTime = *dispute.VotingPeriodEnd
		}
		phase = "voting_period"
	default:
		return models.TimeRemaining{Phase: phase, IsExpired: true}
	}

	remaining := endTime.Sub(now)
	if remaining <= 0 {
		return models.TimeRemaining{Phase: phase, IsExpired: true}
	}

	hours := int(remaining.Hours())
	minutes := int(remaining.Minutes()) % 60
	seconds := int(remaining.Seconds()) % 60

	return models.TimeRemaining{
		Phase:     phase,
		Hours:     hours,
		Minutes:   minutes,
		Seconds:   seconds,
		IsExpired: false,
	}
}
