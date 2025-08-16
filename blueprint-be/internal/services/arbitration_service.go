package services

import (
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"fmt"
	"math"
	"time"

	"blueprint-module/pkg/models"
	"gorm.io/gorm"
	"github.com/gin-gonic/gin"
)

// ArbitrationService 탈중앙화된 분쟁 해결 서비스
type ArbitrationService struct {
	db *gorm.DB
}

// NewArbitrationService 생성자
func NewArbitrationService(db *gorm.DB) *ArbitrationService {
	return &ArbitrationService{
		db: db,
	}
}

// SubmitCase 분쟁 사건 제기
func (s *ArbitrationService) SubmitCase(req *models.SubmitArbitrationRequest, plaintiffID uint) (*models.ArbitrationCase, error) {
	// 1. 사용자 지갑 확인
	var userWallet models.UserWallet
	if err := s.db.Where("user_id = ?", plaintiffID).First(&userWallet).Error; err != nil {
		return nil, errors.New("지갑을 찾을 수 없습니다")
	}

	// 2. 스테이킹 금액 확인
	if userWallet.BlueprintBalance < req.StakeAmount {
		return nil, errors.New("분쟁 제기에 필요한 BLUEPRINT 잔액이 부족합니다")
	}

	// 3. 분쟁 대상 유효성 검증
	if err := s.validateDisputeTarget(req); err != nil {
		return nil, err
	}

	// 4. 사건 번호 생성
	caseNumber, err := s.generateCaseNumber()
	if err != nil {
		return nil, fmt.Errorf("사건 번호 생성 실패: %w", err)
	}

	// 트랜잭션 시작
	var arbitrationCase *models.ArbitrationCase
	err = s.db.Transaction(func(tx *gorm.DB) error {
		// 5. BLUEPRINT 스테이킹 (잠금)
		userWallet.BlueprintBalance -= req.StakeAmount
		userWallet.BlueprintLockedBalance += req.StakeAmount
		if err := tx.Save(&userWallet).Error; err != nil {
			return fmt.Errorf("스테이킹 처리 실패: %w", err)
		}

		// 6. 분쟁 사건 생성
		requiredJurors := s.calculateRequiredJurors(req.DisputeType, req.ClaimedAmount)
		formationDeadline := time.Now().Add(48 * time.Hour) // 48시간 내 배심원단 구성

		arbitrationCase = &models.ArbitrationCase{
			CaseNumber:            caseNumber,
			PlaintiffID:           plaintiffID,
			DefendantID:           req.DefendantID,
			DisputeType:           req.DisputeType,
			MilestoneID:           req.MilestoneID,
			MentorshipID:          req.MentorshipID,
			TradeID:               req.TradeID,
			Title:                 req.Title,
			Description:           req.Description,
			Evidence:              req.Evidence,
			ClaimedAmount:         req.ClaimedAmount,
			Status:                models.ArbitrationStatusSubmitted,
			Priority:              s.calculatePriority(req.DisputeType, req.ClaimedAmount),
			StakeAmount:           req.StakeAmount,
			RequiredJurors:        requiredJurors,
			JuryFormationDeadline: formationDeadline,
		}

		if err := tx.Create(arbitrationCase).Error; err != nil {
			return fmt.Errorf("분쟁 사건 생성 실패: %w", err)
		}

		// 7. 초기 검토 시작
		if err := s.startInitialReview(tx, arbitrationCase.ID); err != nil {
			return fmt.Errorf("초기 검토 시작 실패: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// 8. 배심원단 선정 프로세스 시작 (비동기)
	go s.startJurySelection(arbitrationCase.ID)

	return arbitrationCase, nil
}

// StartJurySelection 배심원단 선정 프로세스
func (s *ArbitrationService) startJurySelection(caseID uint) {
	// 1. 사건 정보 조회
	var arbitrationCase models.ArbitrationCase
	if err := s.db.First(&arbitrationCase, caseID).Error; err != nil {
		return
	}

	// 2. 자격을 갖춘 배심원 후보 조회
	candidates, err := s.getEligibleJurors(arbitrationCase.DisputeType, arbitrationCase.PlaintiffID, arbitrationCase.DefendantID)
	if err != nil {
		return
	}

	// 3. 무작위로 배심원 선정
	selectedJurors, err := s.selectJurors(candidates, arbitrationCase.RequiredJurors)
	if err != nil {
		return
	}

	// 4. 선정된 배심원들에게 알림 및 스테이킹 요구
	s.db.Transaction(func(tx *gorm.DB) error {
		// 배심원 목록 업데이트
		arbitrationCase.SelectedJurors = selectedJurors
		arbitrationCase.Status = models.ArbitrationStatusJurySelection
		tx.Save(&arbitrationCase)

		// 배심원들에게 알림 발송 및 스테이킹 요구
		for _, jurorID := range selectedJurors {
			s.notifyJurorSelection(jurorID, caseID)
		}

		return nil
	})
}

// GetEligibleJurors 자격을 갖춘 배심원 후보 조회
func (s *ArbitrationService) getEligibleJurors(disputeType models.ArbitrationDisputeType, plaintiffID, defendantID uint) ([]models.JurorQualification, error) {
	var candidates []models.JurorQualification

	// 기본 자격 요건: 충분한 스테이킹, 활성 상태, 이해충돌 없음
	query := s.db.Where("is_active = ? AND is_suspended = ? AND current_stake >= min_stake_amount", true, false).
		Where("user_id != ? AND user_id != ?", plaintiffID, defendantID) // 이해충돌 방지

	// 분쟁 유형별 전문성 고려
	switch disputeType {
	case models.DisputeTypeMentorMalpractice:
		query = query.Where("JSON_CONTAINS(expertise_areas, '\"mentoring\"') OR legal_background = ?", true)
	case models.DisputeTypeProjectFraud:
		query = query.Where("JSON_CONTAINS(expertise_areas, '\"technical\"') OR legal_background = ?", true)
	case models.DisputeTypeIntellectualProperty:
		query = query.Where("legal_background = ? OR JSON_CONTAINS(expertise_areas, '\"legal\"')", true)
	}

	// 평판 점수와 정확도로 정렬
	query = query.Order("reputation_score DESC, accuracy_rate DESC")

	if err := query.Find(&candidates).Error; err != nil {
		return nil, fmt.Errorf("배심원 후보 조회 실패: %w", err)
	}

	return candidates, nil
}

// SelectJurors 무작위 배심원 선정 (가중 확률)
func (s *ArbitrationService) selectJurors(candidates []models.JurorQualification, requiredCount int) ([]uint, error) {
	if len(candidates) < requiredCount {
		return nil, errors.New("충분한 배심원 후보가 없습니다")
	}

	// 가중 확률 계산 (평판 점수 + 정확도 기반)
	type weightedCandidate struct {
		UserID uint
		Weight float64
	}

	var weightedCandidates []weightedCandidate
	totalWeight := 0.0

	for _, candidate := range candidates {
		// 가중치 = 평판점수 * 정확도 * 스테이킹비율
		stakeRatio := math.Min(float64(candidate.CurrentStake)/float64(candidate.MinStakeAmount), 2.0) // 최대 2배
		weight := candidate.ReputationScore * candidate.AccuracyRate * stakeRatio
		
		weightedCandidates = append(weightedCandidates, weightedCandidate{
			UserID: candidate.UserID,
			Weight: weight,
		})
		totalWeight += weight
	}

	// 가중 무작위 선정
	selected := make([]uint, 0, requiredCount)
	used := make(map[uint]bool)

	for len(selected) < requiredCount {
		// 무작위 숫자 생성
		randBytes := make([]byte, 8)
		rand.Read(randBytes)
		randFloat := float64(randBytes[0]) / 255.0 * totalWeight

		// 가중 선택
		currentWeight := 0.0
		for _, candidate := range weightedCandidates {
			if used[candidate.UserID] {
				continue
			}
			
			currentWeight += candidate.Weight
			if currentWeight >= randFloat {
				selected = append(selected, candidate.UserID)
				used[candidate.UserID] = true
				break
			}
		}

		// 무한 루프 방지
		if len(used) >= len(weightedCandidates) {
			break
		}
	}

	return selected, nil
}

// CommitVote 배심원 투표 제출 (Commit phase)
func (s *ArbitrationService) CommitVote(req *models.JurorVoteRequest, jurorID uint) (*models.ArbitrationVote, error) {
	// 1. 사건 조회 및 상태 확인
	var arbitrationCase models.ArbitrationCase
	if err := s.db.First(&arbitrationCase, req.CaseID).Error; err != nil {
		return nil, fmt.Errorf("사건을 찾을 수 없습니다: %w", err)
	}

	if arbitrationCase.Status != models.ArbitrationStatusVoting {
		return nil, errors.New("현재 투표 기간이 아닙니다")
	}

	// 2. 배심원 자격 확인
	isEligible := false
	for _, selectedJurorID := range arbitrationCase.SelectedJurors {
		if selectedJurorID == jurorID {
			isEligible = true
			break
		}
	}
	if !isEligible {
		return nil, errors.New("이 사건의 배심원이 아닙니다")
	}

	// 3. 이미 투표했는지 확인
	var existingVote models.ArbitrationVote
	if err := s.db.Where("case_id = ? AND juror_id = ?", req.CaseID, jurorID).First(&existingVote).Error; err == nil {
		return nil, errors.New("이미 투표하셨습니다")
	}

	// 4. 배심원 자격 정보 조회
	var jurorQualification models.JurorQualification
	if err := s.db.Where("user_id = ?", jurorID).First(&jurorQualification).Error; err != nil {
		return nil, errors.New("배심원 자격을 찾을 수 없습니다")
	}

	// 5. 투표 생성
	vote := &models.ArbitrationVote{
		CaseID:             req.CaseID,
		JurorID:            jurorID,
		CommitHash:         req.CommitHash,
		JurorStake:         jurorQualification.CurrentStake,
		QualificationScore: jurorQualification.ReputationScore,
		CommittedAt:        &[]time.Time{time.Now()}[0],
	}

	if err := s.db.Create(vote).Error; err != nil {
		return nil, fmt.Errorf("투표 저장 실패: %w", err)
	}

	// 6. 모든 배심원이 투표했는지 확인
	s.checkVotingCompletion(req.CaseID)

	return vote, nil
}

// RevealVote 투표 공개 (Reveal phase)
func (s *ArbitrationService) RevealVote(req *models.RevealVoteRequest, jurorID uint) error {
	// 1. 투표 조회
	var vote models.ArbitrationVote
	if err := s.db.Where("case_id = ? AND juror_id = ?", req.CaseID, jurorID).First(&vote).Error; err != nil {
		return fmt.Errorf("투표를 찾을 수 없습니다: %w", err)
	}

	// 2. 사건 상태 확인
	var arbitrationCase models.ArbitrationCase
	if err := s.db.First(&arbitrationCase, req.CaseID).Error; err != nil {
		return fmt.Errorf("사건을 찾을 수 없습니다: %w", err)
	}

	if arbitrationCase.Status != models.ArbitrationStatusReveal {
		return errors.New("현재 투표 공개 기간이 아닙니다")
	}

	// 3. 해시 검증
	expectedHash := s.generateCommitHash(string(req.Vote), req.Salt)
	if vote.CommitHash != expectedHash {
		return errors.New("투표 해시가 일치하지 않습니다")
	}

	// 4. 투표 공개
	vote.RevealedVote = &req.Vote
	vote.RevealedSalt = req.Salt
	vote.VoteReason = req.VoteReason
	vote.RevealedAt = &[]time.Time{time.Now()}[0]

	if err := s.db.Save(&vote).Error; err != nil {
		return fmt.Errorf("투표 공개 실패: %w", err)
	}

	// 5. 모든 투표가 공개되었는지 확인
	s.checkRevealCompletion(req.CaseID)

	return nil
}

// FinalizeCase 사건 최종 판결
func (s *ArbitrationService) FinalizeCase(caseID uint) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 1. 사건 및 투표 조회
		var arbitrationCase models.ArbitrationCase
		if err := tx.Preload("Votes").First(&arbitrationCase, caseID).Error; err != nil {
			return fmt.Errorf("사건 조회 실패: %w", err)
		}

		// 2. 투표 집계 및 결과 결정
		decision, confidence := s.calculateDecision(arbitrationCase.Votes)
		
		// 3. 사건 결과 업데이트
		now := time.Now()
		arbitrationCase.Decision = decision
		arbitrationCase.DecisionReason = s.generateDecisionReason(arbitrationCase.Votes, decision)
		arbitrationCase.Status = models.ArbitrationStatusDecided
		arbitrationCase.DecidedAt = &now

		// 4. 배상 금액 결정
		if decision == models.ArbitrationDecisionPlaintiffWins {
			arbitrationCase.AwardAmount = arbitrationCase.ClaimedAmount
		} else if decision == models.ArbitrationDecisionPartialWin {
			arbitrationCase.AwardAmount = arbitrationCase.ClaimedAmount / 2 // 50% 배상
		}

		if err := tx.Save(&arbitrationCase).Error; err != nil {
			return fmt.Errorf("사건 업데이트 실패: %w", err)
		}

		// 5. 배심원 보상 지급
		if err := s.distributeJurorRewards(tx, caseID, decision, confidence); err != nil {
			return fmt.Errorf("배심원 보상 지급 실패: %w", err)
		}

		// 6. 당사자들에게 배상/환급 처리
		if err := s.processSettlement(tx, &arbitrationCase); err != nil {
			return fmt.Errorf("배상 처리 실패: %w", err)
		}

		return nil
	})
}

// Helper functions

func (s *ArbitrationService) generateCaseNumber() (string, error) {
	year := time.Now().Year()
	
	// 해당 연도의 사건 수 조회
	var count int64
	s.db.Model(&models.ArbitrationCase{}).
		Where("EXTRACT(YEAR FROM created_at) = ?", year).
		Count(&count)

	return fmt.Sprintf("ACC-%d-%04d", year, count+1), nil
}

func (s *ArbitrationService) calculateRequiredJurors(disputeType models.ArbitrationDisputeType, claimedAmount int64) int {
	// 기본 5명, 금액이나 중요도에 따라 증가
	baseJurors := 5
	
	if claimedAmount > 100000 { // 10만 이상
		baseJurors = 7
	}
	if claimedAmount > 500000 { // 50만 이상
		baseJurors = 9
	}
	
	// 특정 분쟁 유형은 더 많은 배심원 필요
	switch disputeType {
	case models.DisputeTypeProjectFraud, models.DisputeTypeIntellectualProperty:
		baseJurors += 2
	}
	
	return baseJurors
}

func (s *ArbitrationService) calculatePriority(disputeType models.ArbitrationDisputeType, claimedAmount int64) models.ArbitrationPriority {
	if claimedAmount > 500000 {
		return models.ArbitrationPriorityUrgent
	}
	if claimedAmount > 100000 {
		return models.ArbitrationPriorityHigh
	}
	
	switch disputeType {
	case models.DisputeTypeProjectFraud:
		return models.ArbitrationPriorityHigh
	case models.DisputeTypeMentorMalpractice:
		return models.ArbitrationPriorityNormal
	default:
		return models.ArbitrationPriorityNormal
	}
}

func (s *ArbitrationService) generateCommitHash(vote, salt string) string {
	hash := sha256.Sum256([]byte(vote + salt))
	return fmt.Sprintf("%x", hash)
}

func (s *ArbitrationService) calculateDecision(votes []models.ArbitrationVote) (models.ArbitrationDecision, float64) {
	if len(votes) == 0 {
		return models.ArbitrationDecisionDismissed, 0.0
	}

	// 가중 투표 집계
	voteCount := make(map[models.ArbitrationDecision]float64)
	totalWeight := 0.0

	for _, vote := range votes {
		if vote.RevealedVote != nil && vote.IsValid {
			weight := vote.QualificationScore * (1.0 + float64(vote.JurorStake)/10000.0) // 스테이킹 가중치
			voteCount[*vote.RevealedVote] += weight
			totalWeight += weight
		}
	}

	// 최다 득표 결정 찾기
	maxVotes := 0.0
	var winningDecision models.ArbitrationDecision
	
	for decision, count := range voteCount {
		if count > maxVotes {
			maxVotes = count
			winningDecision = decision
		}
	}

	// 신뢰도 계산 (최다 득표 비율)
	confidence := 0.0
	if totalWeight > 0 {
		confidence = maxVotes / totalWeight
	}

	return winningDecision, confidence
}

func (s *ArbitrationService) generateDecisionReason(votes []models.ArbitrationVote, decision models.ArbitrationDecision) string {
	validVotes := 0
	for _, vote := range votes {
		if vote.RevealedVote != nil && *vote.RevealedVote == decision {
			validVotes++
		}
	}
	
	return fmt.Sprintf("배심원 %d명 중 %d명이 %s에 투표했습니다.", len(votes), validVotes, decision)
}

// Additional helper methods would be implemented here...
func (s *ArbitrationService) validateDisputeTarget(req *models.SubmitArbitrationRequest) error {
	// Implementation for validating dispute targets
	return nil
}

func (s *ArbitrationService) startInitialReview(tx *gorm.DB, caseID uint) error {
	// Implementation for initial case review
	return nil
}

func (s *ArbitrationService) notifyJurorSelection(jurorID uint, caseID uint) {
	// Implementation for notifying selected jurors
}

func (s *ArbitrationService) checkVotingCompletion(caseID uint) {
	// Implementation for checking if all jurors have voted
}

func (s *ArbitrationService) checkRevealCompletion(caseID uint) {
	// Implementation for checking if all votes are revealed
}

func (s *ArbitrationService) distributeJurorRewards(tx *gorm.DB, caseID uint, decision models.ArbitrationDecision, confidence float64) error {
	// Implementation for distributing rewards to jurors
	return nil
}

func (s *ArbitrationService) processSettlement(tx *gorm.DB, arbitrationCase *models.ArbitrationCase) error {
	// Implementation for processing settlement between parties
	return nil
}

// GetCaseDetails 분쟁 사건 상세 정보 조회
func (s *ArbitrationService) GetCaseDetails(caseID uint, userID uint) (*models.ArbitrationCaseResponse, error) {
	var arbitrationCase models.ArbitrationCase
	if err := s.db.Preload("Plaintiff").Preload("Defendant").Preload("Votes").First(&arbitrationCase, caseID).Error; err != nil {
		return nil, fmt.Errorf("사건을 찾을 수 없습니다: %w", err)
	}

	var votes []models.ArbitrationVote
	s.db.Preload("Juror").Where("case_id = ?", caseID).Find(&votes)

	canVote := false
	var userVote *models.ArbitrationVote
	if userID > 0 {
		for _, jurorID := range arbitrationCase.SelectedJurors {
			if jurorID == userID {
				canVote = true
				break
			}
		}
		
		for i := range votes {
			if votes[i].JurorID == userID {
				userVote = &votes[i]
				canVote = false
				break
			}
		}
	}

	statistics := s.calculateCaseStatistics(votes, arbitrationCase.RequiredJurors)
	
	return &models.ArbitrationCaseResponse{
		Case:       arbitrationCase,
		Votes:      votes,
		CanVote:    canVote,
		UserVote:   userVote,
		TimeLeft:   int64(time.Until(arbitrationCase.JuryFormationDeadline).Seconds()),
		Statistics: statistics,
	}, nil
}

// GetJurorDashboard 배심원 대시보드 조회
func (s *ArbitrationService) GetJurorDashboard(userID uint) (*models.JurorDashboardResponse, error) {
	var qualification models.JurorQualification
	if err := s.db.Where("user_id = ?", userID).First(&qualification).Error; err != nil {
		return nil, fmt.Errorf("배심원 자격을 찾을 수 없습니다: %w", err)
	}

	var pendingCases []models.ArbitrationCase
	s.db.Where("status = ?", models.ArbitrationStatusJurySelection).Find(&pendingCases)

	var activeCases []models.ArbitrationCase
	s.db.Where("JSON_CONTAINS(selected_jurors, ?)", fmt.Sprintf(`"%d"`, userID)).
		Where("status IN ?", []models.ArbitrationStatus{
			models.ArbitrationStatusVoting,
			models.ArbitrationStatusReveal,
		}).Find(&activeCases)

	var completedCases []models.ArbitrationCase
	s.db.Joins("JOIN arbitration_votes ON arbitration_cases.id = arbitration_votes.case_id").
		Where("arbitration_votes.juror_id = ? AND arbitration_cases.status = ?", userID, models.ArbitrationStatusDecided).
		Find(&completedCases)

	var totalRewards int64
	s.db.Model(&models.ArbitrationReward{}).
		Where("juror_id = ? AND status = ?", userID, "distributed").
		Select("COALESCE(SUM(total_reward), 0)").Scan(&totalRewards)

	statistics := models.JurorStatistics{
		TotalCases:        qualification.TotalCases,
		AccuracyRate:      qualification.AccuracyRate,
		ParticipationRate: qualification.ParticipationRate,
		AverageResponseTime: qualification.AverageResponseTime,
		Rank:              1, // TODO: 실제 순위 계산
		TotalEarnings:     totalRewards,
	}

	return &models.JurorDashboardResponse{
		Qualification:   qualification,
		PendingCases:    pendingCases,
		ActiveCases:     activeCases,
		CompletedCases:  completedCases,
		TotalRewards:    totalRewards,
		Statistics:      statistics,
	}, nil
}

// GetPendingCases 대기 중인 분쟁 사건 목록 조회
func (s *ArbitrationService) GetPendingCases(userID uint, page, limit int, disputeType, priority string) (interface{}, error) {
	offset := (page - 1) * limit
	
	query := s.db.Model(&models.ArbitrationCase{}).
		Where("status IN ?", []models.ArbitrationStatus{
			models.ArbitrationStatusSubmitted,
			models.ArbitrationStatusJurySelection,
			models.ArbitrationStatusVoting,
		})

	if disputeType != "" {
		query = query.Where("dispute_type = ?", disputeType)
	}
	if priority != "" {
		query = query.Where("priority = ?", priority)
	}

	var cases []models.ArbitrationCase
	var total int64
	
	query.Count(&total)
	query.Offset(offset).Limit(limit).Preload("Plaintiff").Preload("Defendant").Find(&cases)

	return gin.H{
		"cases": cases,
		"pagination": gin.H{
			"page":        page,
			"limit":       limit,
			"total":       total,
			"total_pages": (total + int64(limit) - 1) / int64(limit),
		},
	}, nil
}

// GetUserCases 사용자의 분쟁 사건 목록 조회
func (s *ArbitrationService) GetUserCases(userID uint, page, limit int, status, role string) (interface{}, error) {
	offset := (page - 1) * limit
	
	query := s.db.Model(&models.ArbitrationCase{})

	switch role {
	case "plaintiff":
		query = query.Where("plaintiff_id = ?", userID)
	case "defendant":
		query = query.Where("defendant_id = ?", userID)
	case "juror":
		query = query.Where("JSON_CONTAINS(selected_jurors, ?)", fmt.Sprintf(`"%d"`, userID))
	default:
		query = query.Where("plaintiff_id = ? OR defendant_id = ? OR JSON_CONTAINS(selected_jurors, ?)", 
			userID, userID, fmt.Sprintf(`"%d"`, userID))
	}

	if status != "" {
		query = query.Where("status = ?", status)
	}

	var cases []models.ArbitrationCase
	var total int64
	
	query.Count(&total)
	query.Offset(offset).Limit(limit).Preload("Plaintiff").Preload("Defendant").Find(&cases)

	return gin.H{
		"cases": cases,
		"pagination": gin.H{
			"page":        page,
			"limit":       limit,
			"total":       total,
			"total_pages": (total + int64(limit) - 1) / int64(limit),
		},
	}, nil
}

// RegisterJuror 배심원 등록
func (s *ArbitrationService) RegisterJuror(userID uint, req interface{}) (*models.JurorQualification, error) {
	reqData := req.(*struct {
		MinStakeAmount  int64    `json:"min_stake_amount"`
		ExpertiseAreas  []string `json:"expertise_areas"`
		LanguageSkills  []string `json:"language_skills"`
		LegalBackground bool     `json:"legal_background"`
	})

	var userWallet models.UserWallet
	if err := s.db.Where("user_id = ?", userID).First(&userWallet).Error; err != nil {
		return nil, errors.New("지갑을 찾을 수 없습니다")
	}

	if userWallet.BlueprintBalance < reqData.MinStakeAmount {
		return nil, errors.New("배심원 등록에 필요한 BLUEPRINT 잔액이 부족합니다")
	}

	qualification := &models.JurorQualification{
		UserID:          userID,
		MinStakeAmount:  reqData.MinStakeAmount,
		CurrentStake:    reqData.MinStakeAmount,
		ReputationScore: 0.5,
		ExpertiseAreas:  reqData.ExpertiseAreas,
		LanguageSkills:  reqData.LanguageSkills,
		LegalBackground: reqData.LegalBackground,
		IsActive:        true,
		ParticipationRate: 1.0,
	}

	if err := s.db.Create(qualification).Error; err != nil {
		return nil, fmt.Errorf("배심원 등록 실패: %w", err)
	}

	return qualification, nil
}

// GetArbitrationStats 분쟁 해결 통계 조회
func (s *ArbitrationService) GetArbitrationStats(period string) (interface{}, error) {
	var startDate time.Time
	endDate := time.Now()

	switch period {
	case "daily":
		startDate = endDate.AddDate(0, 0, -1)
	case "weekly":
		startDate = endDate.AddDate(0, 0, -7)
	case "monthly":
		startDate = endDate.AddDate(0, -1, 0)
	case "yearly":
		startDate = endDate.AddDate(-1, 0, 0)
	default:
		startDate = endDate.AddDate(0, -1, 0)
	}

	var totalCases int64
	var resolvedCases int64
	var pendingCases int64
	var avgResolutionTime float64

	s.db.Model(&models.ArbitrationCase{}).
		Where("created_at BETWEEN ? AND ?", startDate, endDate).
		Count(&totalCases)

	s.db.Model(&models.ArbitrationCase{}).
		Where("status = ? AND created_at BETWEEN ? AND ?", models.ArbitrationStatusDecided, startDate, endDate).
		Count(&resolvedCases)

	s.db.Model(&models.ArbitrationCase{}).
		Where("status NOT IN ? AND created_at BETWEEN ? AND ?", 
			[]models.ArbitrationStatus{models.ArbitrationStatusDecided, models.ArbitrationStatusClosed}, 
			startDate, endDate).
		Count(&pendingCases)

	s.db.Model(&models.ArbitrationCase{}).
		Where("status = ? AND decided_at IS NOT NULL AND created_at BETWEEN ? AND ?", 
			models.ArbitrationStatusDecided, startDate, endDate).
		Select("AVG(TIMESTAMPDIFF(HOUR, created_at, decided_at))").
		Scan(&avgResolutionTime)

	return gin.H{
		"period":              period,
		"total_cases":         totalCases,
		"resolved_cases":      resolvedCases,
		"pending_cases":       pendingCases,
		"resolution_rate":     float64(resolvedCases) / float64(totalCases),
		"avg_resolution_time": avgResolutionTime,
		"updated_at":          endDate,
	}, nil
}

// AppealCase 판결 이의제기
func (s *ArbitrationService) AppealCase(caseID uint, userID uint, reason, evidence string, stakeAmount int64) (interface{}, error) {
	var arbitrationCase models.ArbitrationCase
	if err := s.db.First(&arbitrationCase, caseID).Error; err != nil {
		return nil, fmt.Errorf("사건을 찾을 수 없습니다: %w", err)
	}

	if arbitrationCase.Status != models.ArbitrationStatusDecided {
		return nil, errors.New("아직 판결이 나지 않은 사건입니다")
	}

	if arbitrationCase.PlaintiffID != userID && arbitrationCase.DefendantID != userID {
		return nil, errors.New("해당 사건의 당사자가 아닙니다")
	}

	appealCase := &models.ArbitrationCase{
		CaseNumber:    arbitrationCase.CaseNumber + "-APPEAL",
		PlaintiffID:   userID,
		DefendantID:   arbitrationCase.DefendantID,
		DisputeType:   arbitrationCase.DisputeType,
		Title:         "항소: " + arbitrationCase.Title,
		Description:   reason,
		Evidence:      evidence,
		ClaimedAmount: arbitrationCase.ClaimedAmount,
		StakeAmount:   stakeAmount,
		Status:        models.ArbitrationStatusSubmitted,
		Priority:      models.ArbitrationPriorityHigh,
	}

	if arbitrationCase.DefendantID == userID {
		appealCase.DefendantID = arbitrationCase.PlaintiffID
	}

	if err := s.db.Create(appealCase).Error; err != nil {
		return nil, fmt.Errorf("이의제기 사건 생성 실패: %w", err)
	}

	arbitrationCase.Status = models.ArbitrationStatusAppealed
	s.db.Save(&arbitrationCase)

	return appealCase, nil
}

// Helper method
func (s *ArbitrationService) calculateCaseStatistics(votes []models.ArbitrationVote, requiredJurors int) models.CaseStatistics {
	votesCommitted := 0
	votesRevealed := 0
	decisionCount := make(map[models.ArbitrationDecision]int)

	for _, vote := range votes {
		if vote.CommittedAt != nil {
			votesCommitted++
		}
		if vote.RevealedVote != nil {
			votesRevealed++
			decisionCount[*vote.RevealedVote]++
		}
	}

	var majorityDecision *models.ArbitrationDecision
	maxVotes := 0
	for decision, count := range decisionCount {
		if count > maxVotes {
			maxVotes = count
			decision := decision
			majorityDecision = &decision
		}
	}

	confidence := 0.0
	if votesRevealed > 0 {
		confidence = float64(maxVotes) / float64(votesRevealed)
	}

	return models.CaseStatistics{
		TotalJurors:        requiredJurors,
		VotesCommitted:     votesCommitted,
		VotesRevealed:      votesRevealed,
		MajorityDecision:   majorityDecision,
		DecisionConfidence: confidence,
	}
}