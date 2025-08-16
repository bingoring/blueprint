package services

import (
	"errors"
	"fmt"
	"math"
	"mime/multipart"
	"time"

	"blueprint-module/pkg/models"
	"gorm.io/gorm"
)

// VerificationService 마일스톤 증명 및 검증 서비스
type VerificationService struct {
	db          *gorm.DB
	fileService *FileService // 파일 업로드 서비스
}

// NewVerificationService 생성자
func NewVerificationService(db *gorm.DB, fileService *FileService) *VerificationService {
	return &VerificationService{
		db:          db,
		fileService: fileService,
	}
}

// UploadFile 파일 업로드 (FileService 래퍼)
func (s *VerificationService) UploadFile(file multipart.File, header *multipart.FileHeader, category string) (string, error) {
	return s.fileService.UploadFile(file, header, category)
}

// SubmitProof 증거 제출
func (s *VerificationService) SubmitProof(req *models.SubmitProofRequest, userID uint) (*models.MilestoneProof, error) {
	// 1. 마일스톤 조회 및 검증
	var milestone models.Milestone
	if err := s.db.First(&milestone, req.MilestoneID).Error; err != nil {
		return nil, fmt.Errorf("마일스톤을 찾을 수 없습니다: %w", err)
	}

	// 2. 증거 제출 권한 확인 (프로젝트 소유자인지 확인)
	var project models.Project
	if err := s.db.First(&project, milestone.ProjectID).Error; err != nil {
		return nil, fmt.Errorf("프로젝트를 찾을 수 없습니다: %w", err)
	}

	if project.UserID != userID {
		return nil, errors.New("마일스톤 증거는 프로젝트 소유자만 제출할 수 있습니다")
	}

	// 3. 마일스톤 상태 확인
	if !milestone.CanSubmitProof() {
		return nil, errors.New("현재 마일스톤 상태에서는 증거를 제출할 수 없습니다")
	}

	// 4. 이미 제출된 증거가 있는지 확인
	var existingProof models.MilestoneProof
	if err := s.db.Where("milestone_id = ? AND status != ?", req.MilestoneID, models.ProofStatusRejected).First(&existingProof).Error; err == nil {
		return nil, errors.New("이미 증거가 제출되었습니다")
	}

	// 5. 증거 생성
	proof := &models.MilestoneProof{
		MilestoneID:    req.MilestoneID,
		UserID:         userID,
		ProofType:      req.ProofType,
		Title:          req.Title,
		Description:    req.Description,
		ExternalURL:    req.ExternalURL,
		APIData:        req.APIData,
		Metadata:       req.Metadata,
		Status:         models.ProofStatusSubmitted,
		SubmittedAt:    time.Now(),
		ReviewDeadline: time.Now().Add(72 * time.Hour), // 72시간 후
	}

	// 6. 데이터베이스에 저장
	if err := s.db.Create(proof).Error; err != nil {
		return nil, fmt.Errorf("증거 저장 실패: %w", err)
	}

	// 7. 마일스톤 상태 업데이트
	milestone.Status = models.MilestoneStatusProofSubmitted
	if err := s.db.Save(&milestone).Error; err != nil {
		return nil, fmt.Errorf("마일스톤 상태 업데이트 실패: %w", err)
	}

	// 8. 검증 프로세스 시작
	if err := s.StartVerificationProcess(proof.ID); err != nil {
		return nil, fmt.Errorf("검증 프로세스 시작 실패: %w", err)
	}

	return proof, nil
}

// StartVerificationProcess 검증 프로세스 시작
func (s *VerificationService) StartVerificationProcess(proofID uint) error {
	// 1. 증거 조회
	var proof models.MilestoneProof
	if err := s.db.Preload("Milestone").First(&proof, proofID).Error; err != nil {
		return fmt.Errorf("증거를 찾을 수 없습니다: %w", err)
	}

	// 2. 검증 프로세스 생성
	verification := &models.MilestoneVerification{
		MilestoneID:       proof.MilestoneID,
		ProofID:           proof.ID,
		Status:            models.MilestoneVerificationStatusActive,
		StartedAt:         time.Now(),
		ReviewDeadline:    time.Now().Add(72 * time.Hour),
		AutoCompleteAfter: time.Now().Add(96 * time.Hour), // 96시간 후 자동 완료
		MinimumVotes:      proof.Milestone.MinValidators,
		WeightedScore:     0,
	}

	if err := s.db.Create(verification).Error; err != nil {
		return fmt.Errorf("검증 프로세스 생성 실패: %w", err)
	}

	// 3. 마일스톤 상태 업데이트
	proof.Milestone.StartVerificationProcess()
	if err := s.db.Save(&proof.Milestone).Error; err != nil {
		return fmt.Errorf("마일스톤 상태 업데이트 실패: %w", err)
	}

	// 4. 검증인들에게 알림 발송 (향후 구현)
	// TODO: 검증인들에게 이메일/푸시 알림 발송

	return nil
}

// ValidateProof 증거 검증 (검증인 투표)
func (s *VerificationService) ValidateProof(req *models.ValidateProofRequest, validatorID uint) (*models.ProofValidator, error) {
	// 1. 증거 조회
	var proof models.MilestoneProof
	if err := s.db.Preload("Milestone").First(&proof, req.ProofID).Error; err != nil {
		return nil, fmt.Errorf("증거를 찾을 수 없습니다: %w", err)
	}

	// 2. 검증인 자격 확인
	canValidate, qualification, err := s.CanUserValidate(validatorID, proof.MilestoneID)
	if err != nil {
		return nil, err
	}
	if !canValidate {
		return nil, errors.New("검증 권한이 없습니다")
	}

	// 3. 이미 투표했는지 확인
	var existingVote models.ProofValidator
	if err := s.db.Where("proof_id = ? AND user_id = ?", req.ProofID, validatorID).First(&existingVote).Error; err == nil {
		return nil, errors.New("이미 투표하셨습니다")
	}

	// 4. 검증 기간 확인
	if proof.Milestone.IsVerificationExpired() {
		return nil, errors.New("검증 기간이 만료되었습니다")
	}

	// 5. 투표 가중치 계산
	voteWeight := s.CalculateVoteWeight(qualification)

	// 6. 검증인 투표 생성
	validator := &models.ProofValidator{
		ProofID:           req.ProofID,
		UserID:            validatorID,
		ValidatorType:     s.getValidatorType(qualification),
		StakeAmount:       qualification.StakedAmount,
		QualificationScore: qualification.ReputationScore,
		Vote:              req.Vote,
		Confidence:        req.Confidence,
		Reasoning:         req.Reasoning,
		Evidence:          req.Evidence,
		VoteWeight:        voteWeight,
		VotedAt:          time.Now(),
	}

	if err := s.db.Create(validator).Error; err != nil {
		return nil, fmt.Errorf("투표 저장 실패: %w", err)
	}

	// 7. 검증 통계 업데이트
	if err := s.UpdateVerificationStats(req.ProofID); err != nil {
		return nil, fmt.Errorf("검증 통계 업데이트 실패: %w", err)
	}

	// 8. 검증 완료 조건 확인
	if err := s.CheckVerificationCompletion(req.ProofID); err != nil {
		return nil, fmt.Errorf("검증 완료 확인 실패: %w", err)
	}

	return validator, nil
}

// CanUserValidate 사용자의 검증 자격 확인
func (s *VerificationService) CanUserValidate(userID, milestoneID uint) (bool, *models.ValidatorQualification, error) {
	// 1. 검증인 자격 조회
	var qualification models.ValidatorQualification
	if err := s.db.Where("user_id = ?", userID).First(&qualification).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 자격이 없는 경우 기본 자격으로 생성
			qualification = models.ValidatorQualification{
				UserID:          userID,
				StakedAmount:    0,
				ReputationScore: 0.5, // 기본 평판 점수
			}
			if err := s.db.Create(&qualification).Error; err != nil {
				return false, nil, fmt.Errorf("검증인 자격 생성 실패: %w", err)
			}
		} else {
			return false, nil, fmt.Errorf("검증인 자격 조회 실패: %w", err)
		}
	}

	// 2. 제재 여부 확인
	if qualification.IsSuspended {
		if qualification.SuspendedUntil != nil && time.Now().Before(*qualification.SuspendedUntil) {
			return false, nil, errors.New("계정이 제재 중입니다")
		}
		// 제재 기간이 만료된 경우 제재 해제
		qualification.IsSuspended = false
		qualification.SuspendedUntil = nil
		s.db.Save(&qualification)
	}

	// 3. 최소 자격 요건 확인
	minStake := int64(1000) // 최소 1000 BLUEPRINT 스테이킹
	if qualification.StakedAmount < minStake {
		return false, nil, errors.New("검증에 필요한 최소 스테이킹 양이 부족합니다")
	}

	// 4. 마일스톤과의 이해충돌 확인
	var milestone models.Milestone
	if err := s.db.Preload("Project").First(&milestone, milestoneID).Error; err != nil {
		return false, nil, fmt.Errorf("마일스톤 조회 실패: %w", err)
	}

	// 프로젝트 소유자는 자신의 마일스톤을 검증할 수 없음
	if milestone.Project.UserID == userID {
		return false, nil, errors.New("자신의 프로젝트는 검증할 수 없습니다")
	}

	return true, &qualification, nil
}

// CalculateVoteWeight 투표 가중치 계산
func (s *VerificationService) CalculateVoteWeight(qualification *models.ValidatorQualification) float64 {
	// 기본 가중치 1.0
	weight := 1.0

	// 스테이킹 양에 따른 가중치 (로그 스케일)
	if qualification.StakedAmount > 0 {
		stakeWeight := math.Log10(float64(qualification.StakedAmount)/1000 + 1) // 1000 BLUEPRINT당 0.3 가중치
		weight += stakeWeight * 0.3
	}

	// 평판 점수에 따른 가중치
	reputationWeight := qualification.ReputationScore * 0.5
	weight += reputationWeight

	// 정확도에 따른 가중치
	if qualification.TotalVerifications > 10 { // 최소 10회 이상 검증 참여
		accuracyWeight := qualification.AccuracyRate * 0.3
		weight += accuracyWeight
	}

	// 최대 가중치 제한 (3.0)
	if weight > 3.0 {
		weight = 3.0
	}

	return weight
}

// getValidatorType 검증인 타입 결정
func (s *VerificationService) getValidatorType(qualification *models.ValidatorQualification) string {
	if qualification.IsExpert {
		return "expert"
	}
	if qualification.IsMentor {
		return "mentor"
	}
	return "stakeholder"
}

// UpdateVerificationStats 검증 통계 업데이트
func (s *VerificationService) UpdateVerificationStats(proofID uint) error {
	// 1. 모든 투표 조회
	var validators []models.ProofValidator
	if err := s.db.Where("proof_id = ?", proofID).Find(&validators).Error; err != nil {
		return fmt.Errorf("투표 조회 실패: %w", err)
	}

	// 2. 통계 계산
	var approvalVotes, rejectionVotes int
	var totalWeight, approvalWeight float64

	for _, validator := range validators {
		totalWeight += validator.VoteWeight
		switch validator.Vote {
		case "approve":
			approvalVotes++
			approvalWeight += validator.VoteWeight
		case "reject":
			rejectionVotes++
		}
	}

	// 3. 가중 승인률 계산
	var weightedApprovalRate float64
	if totalWeight > 0 {
		weightedApprovalRate = approvalWeight / totalWeight
	}

	// 4. 증거 통계 업데이트
	if err := s.db.Model(&models.MilestoneProof{}).
		Where("id = ?", proofID).
		Updates(map[string]interface{}{
			"total_validators": len(validators),
			"approval_votes":   approvalVotes,
			"rejection_votes":  rejectionVotes,
		}).Error; err != nil {
		return fmt.Errorf("증거 통계 업데이트 실패: %w", err)
	}

	// 5. 마일스톤 통계 업데이트
	var proof models.MilestoneProof
	if err := s.db.First(&proof, proofID).Error; err != nil {
		return fmt.Errorf("증거 조회 실패: %w", err)
	}

	if err := s.db.Model(&models.Milestone{}).
		Where("id = ?", proof.MilestoneID).
		Updates(map[string]interface{}{
			"total_validators":      len(validators),
			"approval_votes":        approvalVotes,
			"rejection_votes":       rejectionVotes,
			"current_approval_rate": weightedApprovalRate,
		}).Error; err != nil {
		return fmt.Errorf("마일스톤 통계 업데이트 실패: %w", err)
	}

	// 6. 검증 프로세스 통계 업데이트
	if err := s.db.Model(&models.MilestoneVerification{}).
		Where("proof_id = ?", proofID).
		Updates(map[string]interface{}{
			"total_votes":    len(validators),
			"approval_rate":  weightedApprovalRate,
			"weighted_score": approvalWeight,
		}).Error; err != nil {
		return fmt.Errorf("검증 프로세스 통계 업데이트 실패: %w", err)
	}

	return nil
}

// CheckVerificationCompletion 검증 완료 조건 확인
func (s *VerificationService) CheckVerificationCompletion(proofID uint) error {
	// 1. 검증 정보 조회
	var verification models.MilestoneVerification
	if err := s.db.Preload("Milestone").Preload("Proof").First(&verification, "proof_id = ?", proofID).Error; err != nil {
		return fmt.Errorf("검증 정보 조회 실패: %w", err)
	}

	// 2. 완료 조건 확인
	canComplete := verification.Milestone.CanCompleteVerification()
	isExpired := verification.Milestone.IsVerificationExpired()
	
	if !canComplete && !isExpired {
		return nil // 아직 완료 조건 미달성
	}

	// 3. 검증 결과 결정
	approved := verification.Milestone.HasReachedApprovalThreshold()
	
	// 4. 검증 완료 처리
	return s.CompleteVerification(proofID, approved)
}

// CompleteVerification 검증 완료 처리
func (s *VerificationService) CompleteVerification(proofID uint, approved bool) error {
	// 트랜잭션 시작
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 1. 검증 정보 조회
		var verification models.MilestoneVerification
		if err := tx.Preload("Milestone").Preload("Proof").First(&verification, "proof_id = ?", proofID).Error; err != nil {
			return fmt.Errorf("검증 정보 조회 실패: %w", err)
		}

		// 2. 검증 프로세스 완료
		now := time.Now()
		verification.Status = models.MilestoneVerificationStatusApproved
		if !approved {
			verification.Status = models.MilestoneVerificationStatusRejected
		}
		verification.CompletedAt = &now
		verification.FinalResult = "approved"
		if !approved {
			verification.FinalResult = "rejected"
		}

		if err := tx.Save(&verification).Error; err != nil {
			return fmt.Errorf("검증 프로세스 업데이트 실패: %w", err)
		}

		// 3. 증거 상태 업데이트
		verification.Proof.Status = models.ProofStatusApproved
		if !approved {
			verification.Proof.Status = models.ProofStatusRejected
		}

		if err := tx.Save(&verification.Proof).Error; err != nil {
			return fmt.Errorf("증거 상태 업데이트 실패: %w", err)
		}

		// 4. 마일스톤 완료 처리
		verification.Milestone.CompleteVerification(approved)
		if err := tx.Save(&verification.Milestone).Error; err != nil {
			return fmt.Errorf("마일스톤 상태 업데이트 실패: %w", err)
		}

		// 5. 검증인 보상 지급
		if err := s.DistributeValidatorRewards(tx, proofID, approved); err != nil {
			return fmt.Errorf("검증인 보상 지급 실패: %w", err)
		}

		// 6. 베팅 정산 (승인된 경우)
		if approved {
			// TODO: 베팅 정산 로직 구현
		}

		return nil
	})
}

// DistributeValidatorRewards 검증인 보상 지급
func (s *VerificationService) DistributeValidatorRewards(tx *gorm.DB, proofID uint, wasApproved bool) error {
	// 1. 모든 검증인 조회
	var validators []models.ProofValidator
	if err := tx.Where("proof_id = ?", proofID).Find(&validators).Error; err != nil {
		return fmt.Errorf("검증인 조회 실패: %w", err)
	}

	// 2. 각 검증인에게 보상 지급
	for _, validator := range validators {
		// 정확한 투표 여부 확인
		isCorrectVote := (validator.Vote == "approve" && wasApproved) || 
						 (validator.Vote == "reject" && !wasApproved)

		// 기본 보상 계산
		baseReward := int64(100) // 기본 100 BLUEPRINT
		amount := int64(float64(baseReward) * validator.VoteWeight)

		// 정확한 투표에 대한 보너스
		if isCorrectVote {
			amount = int64(float64(amount) * 1.5) // 50% 보너스
		}

		// 보상 레코드 생성
		reward := models.VerificationReward{
			ValidatorID:     validator.ID,
			UserID:          validator.UserID,
			ProofID:         proofID,
			RewardType:      "validation_fee",
			Amount:          amount,
			BonusMultiplier: 1.0,
			IsCorrectVote:   isCorrectVote,
			VoteWeight:      validator.VoteWeight,
			Status:          "pending",
		}

		if isCorrectVote {
			reward.BonusMultiplier = 1.5
		}

		if err := tx.Create(&reward).Error; err != nil {
			return fmt.Errorf("보상 레코드 생성 실패: %w", err)
		}

		// TODO: 실제 토큰 지급 로직 구현
	}

	return nil
}

// DisputeProof 증거 분쟁 제기
func (s *VerificationService) DisputeProof(req *models.DisputeProofRequest, disputerID uint) (*models.ProofDispute, error) {
	// 1. 증거 조회
	var proof models.MilestoneProof
	if err := s.db.First(&proof, req.ProofID).Error; err != nil {
		return nil, fmt.Errorf("증거를 찾을 수 없습니다: %w", err)
	}

	// 2. 분쟁 제기 자격 확인
	canDispute, _, err := s.CanUserValidate(disputerID, proof.MilestoneID)
	if err != nil {
		return nil, err
	}
	if !canDispute {
		return nil, errors.New("분쟁 제기 권한이 없습니다")
	}

	// 3. 스테이킹 확인 (분쟁 제기시 BLUEPRINT 스테이킹 필요)
	var userWallet models.UserWallet
	if err := s.db.Where("user_id = ?", disputerID).First(&userWallet).Error; err != nil {
		return nil, errors.New("지갑을 찾을 수 없습니다")
	}

	if userWallet.BlueprintBalance < req.StakeAmount {
		return nil, errors.New("분쟁 제기에 필요한 BLUEPRINT 잔액이 부족합니다")
	}

	// 트랜잭션 시작
	var dispute *models.ProofDispute
	err = s.db.Transaction(func(tx *gorm.DB) error {
		// 4. BLUEPRINT 스테이킹 (잠금)
		userWallet.BlueprintBalance -= req.StakeAmount
		userWallet.BlueprintLockedBalance += req.StakeAmount
		if err := tx.Save(&userWallet).Error; err != nil {
			return fmt.Errorf("스테이킹 처리 실패: %w", err)
		}

		// 5. 분쟁 레코드 생성
		dispute = &models.ProofDispute{
			ProofID:     req.ProofID,
			UserID:      disputerID,
			DisputeType: req.DisputeType,
			Title:       req.Title,
			Description: req.Description,
			Evidence:    req.Evidence,
			Status:      "open",
			StakeAmount: req.StakeAmount,
		}

		if err := tx.Create(dispute).Error; err != nil {
			return fmt.Errorf("분쟁 레코드 생성 실패: %w", err)
		}

		// 6. 증거 및 마일스톤 상태 업데이트
		proof.Status = models.ProofStatusDisputed
		if err := tx.Save(&proof).Error; err != nil {
			return fmt.Errorf("증거 상태 업데이트 실패: %w", err)
		}

		var milestone models.Milestone
		if err := tx.First(&milestone, proof.MilestoneID).Error; err != nil {
			return fmt.Errorf("마일스톤 조회 실패: %w", err)
		}

		milestone.SetDisputed()
		if err := tx.Save(&milestone).Error; err != nil {
			return fmt.Errorf("마일스톤 상태 업데이트 실패: %w", err)
		}

		return nil
	})
	
	if err != nil {
		return nil, err
	}
	
	return dispute, nil
}

// GetProofVerification 증거 검증 정보 조회
func (s *VerificationService) GetProofVerification(proofID uint, userID uint) (*models.ProofVerificationResponse, error) {
	// 1. 증거 정보 조회
	var proof models.MilestoneProof
	if err := s.db.Preload("Milestone").Preload("User").First(&proof, proofID).Error; err != nil {
		return nil, fmt.Errorf("증거를 찾을 수 없습니다: %w", err)
	}

	// 2. 검증 정보 조회
	var verification models.MilestoneVerification
	if err := s.db.First(&verification, "proof_id = ?", proofID).Error; err != nil {
		return nil, fmt.Errorf("검증 정보를 찾을 수 없습니다: %w", err)
	}

	// 3. 검증인 목록 조회
	var validators []models.ProofValidator
	s.db.Preload("User").Where("proof_id = ?", proofID).Find(&validators)

	// 4. 분쟁 목록 조회
	var disputes []models.ProofDispute
	s.db.Preload("User").Where("proof_id = ?", proofID).Find(&disputes)

	// 5. 현재 사용자의 투표 여부 확인
	canVote := false
	var userVote *models.ProofValidator

	if userID > 0 {
		canValidate, _, err := s.CanUserValidate(userID, proof.MilestoneID)
		if err == nil && canValidate {
			// 이미 투표했는지 확인
			var existingVote models.ProofValidator
			if err := s.db.Where("proof_id = ? AND user_id = ?", proofID, userID).First(&existingVote).Error; err != nil {
				canVote = true
			} else {
				userVote = &existingVote
			}
		}
	}

	return &models.ProofVerificationResponse{
		Proof:        proof,
		Verification: verification,
		Validators:   validators,
		Disputes:     disputes,
		CanVote:      canVote,
		UserVote:     userVote,
	}, nil
}

// GetValidatorDashboard 검증인 대시보드 정보 조회
func (s *VerificationService) GetValidatorDashboard(userID uint) (*models.ValidatorDashboardResponse, error) {
	// 1. 검증인 자격 조회
	var qualification models.ValidatorQualification
	if err := s.db.Where("user_id = ?", userID).First(&qualification).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 기본 자격 생성
			qualification = models.ValidatorQualification{
				UserID:          userID,
				ReputationScore: 0.5,
			}
			s.db.Create(&qualification)
		} else {
			return nil, fmt.Errorf("검증인 자격 조회 실패: %w", err)
		}
	}

	// 2. 대기 중인 증거 목록 조회
	var pendingProofs []models.MilestoneProof
	s.db.Preload("Milestone").Preload("User").
		Where("status = ? AND review_deadline > ?", models.ProofStatusUnderReview, time.Now()).
		Find(&pendingProofs)

	// 3. 최근 투표 내역 조회
	var recentVotes []models.ProofValidator
	s.db.Preload("Proof").Where("user_id = ?", userID).
		Order("voted_at DESC").Limit(10).Find(&recentVotes)

	// 4. 보상 내역 조회
	var rewards []models.VerificationReward
	s.db.Preload("Proof").Where("user_id = ?", userID).
		Order("created_at DESC").Limit(20).Find(&rewards)

	// 5. 통계 계산
	statistics := models.ValidatorStatistics{
		TotalVotes:      qualification.TotalVerifications,
		AccuracyRate:    qualification.AccuracyRate,
		ConsensusRate:   qualification.ConsensusRate,
		CurrentStake:    qualification.StakedAmount,
		ReputationScore: qualification.ReputationScore,
	}

	// 총 보상 계산
	for _, reward := range rewards {
		if reward.Status == "distributed" {
			statistics.TotalRewards += reward.Amount
		}
	}

	return &models.ValidatorDashboardResponse{
		Qualification: qualification,
		PendingProofs: pendingProofs,
		RecentVotes:   recentVotes,
		Rewards:       rewards,
		Statistics:    statistics,
	}, nil
}