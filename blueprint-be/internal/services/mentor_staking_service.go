package services

import (
	"errors"
	"fmt"
	"math"
	"time"

	"blueprint-module/pkg/models"
	"gorm.io/gorm"
	"github.com/gin-gonic/gin"
)

// MentorStakingService 멘토 스테이킹 및 슬래싱 서비스
type MentorStakingService struct {
	db *gorm.DB
}

// NewMentorStakingService 생성자
func NewMentorStakingService(db *gorm.DB) *MentorStakingService {
	return &MentorStakingService{
		db: db,
	}
}

// StakeMentor 멘토 스테이킹
func (s *MentorStakingService) StakeMentor(req *models.StakeMentorRequest, userID uint) (*models.MentorStake, error) {
	// 1. 멘토 존재 확인
	var mentor models.Mentor
	if err := s.db.First(&mentor, req.MentorID).Error; err != nil {
		return nil, fmt.Errorf("멘토를 찾을 수 없습니다: %w", err)
	}

	// 2. 사용자 지갑 확인
	var userWallet models.UserWallet
	if err := s.db.Where("user_id = ?", userID).First(&userWallet).Error; err != nil {
		return nil, errors.New("지갑을 찾을 수 없습니다")
	}

	// 3. 잔액 확인
	if userWallet.BlueprintBalance < req.Amount {
		return nil, errors.New("스테이킹에 필요한 BLUEPRINT 잔액이 부족합니다")
	}

	// 4. 기존 스테이킹 확인 (중복 방지)
	var existingStake models.MentorStake
	if err := s.db.Where("mentor_id = ? AND user_id = ? AND status = ?", 
		req.MentorID, userID, models.MentorStakeStatusActive).First(&existingStake).Error; err == nil {
		return nil, errors.New("이미 해당 멘토에게 스테이킹하고 있습니다")
	}

	// 트랜잭션 시작
	var mentorStake *models.MentorStake
	err := s.db.Transaction(func(tx *gorm.DB) error {
		// 5. BLUEPRINT 차감
		userWallet.BlueprintBalance -= req.Amount
		if err := tx.Save(&userWallet).Error; err != nil {
			return fmt.Errorf("지갑 업데이트 실패: %w", err)
		}

		// 6. 스테이킹 생성
		unlockDate := time.Now().AddDate(0, 0, req.MinimumPeriod)
		mentorStake = &models.MentorStake{
			MentorID:        req.MentorID,
			UserID:          userID,
			Amount:          req.Amount,
			AvailableAmount: req.Amount,
			StakeType:       req.StakeType,
			Purpose:         req.Purpose,
			MinimumPeriod:   req.MinimumPeriod,
			UnlockDate:      unlockDate,
			Status:          models.MentorStakeStatusActive,
			IsAutoRenewal:   req.IsAutoRenewal,
		}

		if err := tx.Create(mentorStake).Error; err != nil {
			return fmt.Errorf("스테이킹 생성 실패: %w", err)
		}

		// 7. 멘토 총 스테이킹 업데이트
		if err := s.updateMentorTotalStake(tx, req.MentorID); err != nil {
			return fmt.Errorf("멘토 스테이킹 업데이트 실패: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return mentorStake, nil
}

// ReportMentor 멘토 신고 및 슬래싱 요청
func (s *MentorStakingService) ReportMentor(req *models.ReportMentorRequest, reporterID uint) (*models.MentorSlashEvent, error) {
	// 1. 멘토 존재 확인
	var mentor models.Mentor
	if err := s.db.First(&mentor, req.MentorID).Error; err != nil {
		return nil, fmt.Errorf("멘토를 찾을 수 없습니다: %w", err)
	}

	// 2. 신고자 자격 확인 (멘티이거나 관련 당사자여야 함)
	canReport, err := s.canUserReportMentor(reporterID, req.MentorID, req.MilestoneID, req.MentorshipID)
	if err != nil {
		return nil, err
	}
	if !canReport {
		return nil, errors.New("해당 멘토를 신고할 권한이 없습니다")
	}

	// 3. 중복 신고 확인
	var existingReport models.MentorSlashEvent
	if err := s.db.Where("mentor_id = ? AND reporter_id = ? AND status IN ?", 
		req.MentorID, reporterID, []models.SlashEventStatus{
			models.SlashEventStatusPending, 
			models.SlashEventStatusReviewing,
		}).First(&existingReport).Error; err == nil {
		return nil, errors.New("이미 해당 멘토에 대한 신고가 처리 중입니다")
	}

	// 4. 멘토의 활성 스테이킹 조회
	var activeStakes []models.MentorStake
	if err := s.db.Where("mentor_id = ? AND status = ?", req.MentorID, models.MentorStakeStatusActive).
		Find(&activeStakes).Error; err != nil {
		return nil, fmt.Errorf("멘토 스테이킹 조회 실패: %w", err)
	}

	if len(activeStakes) == 0 {
		return nil, errors.New("해당 멘토의 활성 스테이킹이 없습니다")
	}

	// 5. 슬래싱 이벤트 생성
	slashEvent := &models.MentorSlashEvent{
		MentorID:     req.MentorID,
		ReporterID:   &reporterID,
		SlashType:    req.SlashType,
		Severity:     req.Severity,
		Reason:       req.Reason,
		Description:  req.Description,
		Evidence:     req.Evidence,
		MilestoneID:  req.MilestoneID,
		MentorshipID: req.MentorshipID,
		Status:       models.SlashEventStatusPending,
		CanAppeal:    true,
		AppealDeadline: &[]time.Time{time.Now().Add(7 * 24 * time.Hour)}[0], // 7일 이의제기 기간
	}

	// 6. 예상 슬래싱 금액 계산
	slashRate := s.calculateSlashRate(req.SlashType, req.Severity)
	totalStaked := s.calculateTotalStaked(activeStakes)
	slashEvent.SlashedAmount = int64(float64(totalStaked) * slashRate)
	slashEvent.SlashRate = slashRate

	if err := s.db.Create(slashEvent).Error; err != nil {
		return nil, fmt.Errorf("슬래싱 이벤트 생성 실패: %w", err)
	}

	// 7. 자동 검토 시작 (비동기)
	go s.startSlashEventReview(slashEvent.ID)

	return slashEvent, nil
}

// ProcessSlashing 슬래싱 실행
func (s *MentorStakingService) ProcessSlashing(slashEventID uint, reviewerID uint, approved bool, comment string) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 1. 슬래싱 이벤트 조회
		var slashEvent models.MentorSlashEvent
		if err := tx.Preload("Mentor").First(&slashEvent, slashEventID).Error; err != nil {
			return fmt.Errorf("슬래싱 이벤트 조회 실패: %w", err)
		}

		// 2. 상태 확인
		if slashEvent.Status != models.SlashEventStatusReviewing {
			return errors.New("현재 검토 중인 슬래싱 이벤트가 아닙니다")
		}

		// 3. 검토 결과 업데이트
		now := time.Now()
		slashEvent.ReviewedBy = &reviewerID
		slashEvent.ReviewComment = comment
		slashEvent.ProcessedAt = &now

		if approved {
			slashEvent.Status = models.SlashEventStatusApproved
			
			// 4. 실제 슬래싱 실행
			if err := s.executeSlashing(tx, &slashEvent); err != nil {
				return fmt.Errorf("슬래싱 실행 실패: %w", err)
			}
		} else {
			slashEvent.Status = models.SlashEventStatusRejected
		}

		if err := tx.Save(&slashEvent).Error; err != nil {
			return fmt.Errorf("슬래싱 이벤트 업데이트 실패: %w", err)
		}

		// 5. 멘토 성과 지표 업데이트
		if approved {
			if err := s.updateMentorPerformanceAfterSlash(tx, slashEvent.MentorID, &slashEvent); err != nil {
				return fmt.Errorf("멘토 성과 지표 업데이트 실패: %w", err)
			}
		}

		return nil
	})
}

// ExecuteSlashing 실제 슬래싱 실행
func (s *MentorStakingService) executeSlashing(tx *gorm.DB, slashEvent *models.MentorSlashEvent) error {
	// 1. 멘토의 활성 스테이킹 조회
	var stakes []models.MentorStake
	if err := tx.Where("mentor_id = ? AND status = ?", slashEvent.MentorID, models.MentorStakeStatusActive).
		Find(&stakes).Error; err != nil {
		return fmt.Errorf("스테이킹 조회 실패: %w", err)
	}

	// 2. 각 스테이킹에서 비례적으로 슬래싱
	totalSlashAmount := slashEvent.SlashedAmount
	remainingSlash := totalSlashAmount

	for i, stake := range stakes {
		if remainingSlash <= 0 {
			break
		}

		// 비례 계산
		stakeRatio := float64(stake.AvailableAmount) / float64(s.calculateTotalStaked(stakes))
		slashFromThisStake := int64(float64(totalSlashAmount) * stakeRatio)
		
		// 마지막 스테이킹에서는 나머지 전부 처리
		if i == len(stakes)-1 {
			slashFromThisStake = remainingSlash
		}

		// 사용 가능 금액보다 많이 슬래싱할 수 없음
		if slashFromThisStake > stake.AvailableAmount {
			slashFromThisStake = stake.AvailableAmount
		}

		// 3. 스테이킹에서 차감
		stake.AvailableAmount -= slashFromThisStake
		stake.LockedAmount += slashFromThisStake

		// 모든 금액이 슬래싱되면 상태 변경
		if stake.AvailableAmount == 0 {
			stake.Status = models.MentorStakeStatusSlashed
		}

		if err := tx.Save(&stake).Error; err != nil {
			return fmt.Errorf("스테이킹 업데이트 실패: %w", err)
		}

		// 4. 슬래싱 이벤트와 스테이킹 연결
		slashEvent.StakeID = stake.ID

		remainingSlash -= slashFromThisStake
	}

	// 5. 슬래싱된 토큰을 플랫폼 보상 풀로 이동
	if err := s.transferSlashedTokensToRewardPool(tx, totalSlashAmount); err != nil {
		return fmt.Errorf("슬래싱 토큰 이동 실패: %w", err)
	}

	return nil
}

// CalculatePerformanceMetrics 멘토 성과 지표 계산
func (s *MentorStakingService) CalculatePerformanceMetrics(mentorID uint, periodType models.MetricPeriodType) (*models.MentorPerformanceMetric, error) {
	// 1. 기간 계산
	endDate := time.Now()
	var startDate time.Time
	
	switch periodType {
	case models.MetricPeriodWeekly:
		startDate = endDate.AddDate(0, 0, -7)
	case models.MetricPeriodMonthly:
		startDate = endDate.AddDate(0, -1, 0)
	case models.MetricPeriodQuarterly:
		startDate = endDate.AddDate(0, -3, 0)
	case models.MetricPeriodYearly:
		startDate = endDate.AddDate(-1, 0, 0)
	default:
		startDate = endDate.AddDate(0, -1, 0) // 기본 1개월
	}

	// 2. 멘토링 활동 통계
	mentorshipStats, err := s.calculateMentorshipStats(mentorID, startDate, endDate)
	if err != nil {
		return nil, err
	}

	// 3. 마일스톤 성과 통계
	milestoneStats, err := s.calculateMilestoneStats(mentorID, startDate, endDate)
	if err != nil {
		return nil, err
	}

	// 4. 참여도 통계
	participationStats, err := s.calculateParticipationStats(mentorID, startDate, endDate)
	if err != nil {
		return nil, err
	}

	// 5. 만족도 통계
	satisfactionStats, err := s.calculateSatisfactionStats(mentorID, startDate, endDate)
	if err != nil {
		return nil, err
	}

	// 6. 경제적 지표
	economicStats, err := s.calculateEconomicStats(mentorID, startDate, endDate)
	if err != nil {
		return nil, err
	}

	// 7. 위험 지표
	riskStats, err := s.calculateRiskStats(mentorID, startDate, endDate)
	if err != nil {
		return nil, err
	}

	// 8. 종합 점수 계산
	performanceScore := s.calculatePerformanceScore(mentorshipStats, milestoneStats, participationStats, satisfactionStats)
	riskScore := s.calculateRiskScore(riskStats)
	qualityScore := s.calculateQualityScore(satisfactionStats, participationStats)

	// 9. 성과 지표 생성
	metric := &models.MentorPerformanceMetric{
		MentorID:             mentorID,
		PeriodType:           periodType,
		StartDate:            startDate,
		EndDate:              endDate,
		TotalMentees:         mentorshipStats["total_mentees"].(int),
		ActiveMentees:        mentorshipStats["active_mentees"].(int),
		CompletedMentorships: mentorshipStats["completed"].(int),
		SuccessfulMilestones: milestoneStats["successful"].(int),
		TotalMilestones:      milestoneStats["total"].(int),
		SuccessRate:          milestoneStats["success_rate"].(float64),
		TotalSessions:        participationStats["total_sessions"].(int),
		AttendanceRate:       participationStats["attendance_rate"].(float64),
		ResponseTime:         participationStats["response_time"].(int),
		SessionRating:        participationStats["session_rating"].(float64),
		MenteeRating:         satisfactionStats["mentee_rating"].(float64),
		FeedbackScore:        satisfactionStats["feedback_score"].(float64),
		RetentionRate:        satisfactionStats["retention_rate"].(float64),
		ReferralRate:         satisfactionStats["referral_rate"].(float64),
		TotalRevenue:         economicStats["total_revenue"].(int64),
		AvgRevenuePerMentee:  economicStats["avg_revenue"].(int64),
		ProfitMargin:         economicStats["profit_margin"].(float64),
		ComplaintCount:       riskStats["complaints"].(int),
		DisputeCount:         riskStats["disputes"].(int),
		SlashCount:           riskStats["slashes"].(int),
		SlashedAmount:        riskStats["slashed_amount"].(int64),
		PerformanceScore:     performanceScore,
		RiskScore:            riskScore,
		QualityScore:         qualityScore,
	}

	// 10. 데이터베이스에 저장
	if err := s.db.Create(metric).Error; err != nil {
		return nil, fmt.Errorf("성과 지표 저장 실패: %w", err)
	}

	return metric, nil
}

// Helper functions

func (s *MentorStakingService) canUserReportMentor(userID, mentorID uint, milestoneID, mentorshipID *uint) (bool, error) {
	// 1. 직접적인 멘토링 관계 확인
	var mentorshipCount int64
	s.db.Model(&models.MentoringSession{}).
		Where("mentor_id = ? AND mentee_id = ?", mentorID, userID).
		Count(&mentorshipCount)
	
	if mentorshipCount > 0 {
		return true, nil
	}

	// 2. 마일스톤 관련 확인 (베팅 참여자 등)
	if milestoneID != nil {
		var tradeCount int64
		s.db.Model(&models.Trade{}).
			Where("milestone_id = ? AND (buyer_id = ? OR seller_id = ?)", *milestoneID, userID, userID).
			Count(&tradeCount)
		
		if tradeCount > 0 {
			return true, nil
		}
	}

	// 3. 특별 권한 확인 (배심원 자격이 있는 사용자는 신고 가능)
	var jurorQualification models.JurorQualification
	if err := s.db.Where("user_id = ? AND is_active = ?", userID, true).First(&jurorQualification).Error; err == nil {
		// 배심원 자격이 있는 사용자는 신고 가능
		return true, nil
	}

	return false, nil
}

func (s *MentorStakingService) calculateSlashRate(slashType models.MentorSlashType, severity models.SlashSeverity) float64 {
	baseRate := 0.0

	// 슬래싱 유형별 기본 비율
	switch slashType {
	case models.SlashTypeAbandonment:
		baseRate = 0.15 // 15%
	case models.SlashTypeMalpractice:
		baseRate = 0.25 // 25%
	case models.SlashTypeFraud:
		baseRate = 0.50 // 50%
	case models.SlashTypePoorPerformance:
		baseRate = 0.10 // 10%
	case models.SlashTypeEthicsViolation:
		baseRate = 0.20 // 20%
	case models.SlashTypeAbuse:
		baseRate = 0.30 // 30%
	case models.SlashTypeConflictOfInterest:
		baseRate = 0.15 // 15%
	case models.SlashTypeNoShow:
		baseRate = 0.05 // 5%
	}

	// 심각도별 조정
	switch severity {
	case models.SlashSeverityMinor:
		baseRate *= 0.5 // 50% 감소
	case models.SlashSeverityModerate:
		// 기본값 유지
	case models.SlashSeverityMajor:
		baseRate *= 1.5 // 50% 증가
	case models.SlashSeverityCritical:
		baseRate *= 2.0 // 100% 증가
	}

	// 최대 100% 제한
	if baseRate > 1.0 {
		baseRate = 1.0
	}

	return baseRate
}

func (s *MentorStakingService) calculateTotalStaked(stakes []models.MentorStake) int64 {
	total := int64(0)
	for _, stake := range stakes {
		total += stake.AvailableAmount
	}
	return total
}

func (s *MentorStakingService) updateMentorTotalStake(tx *gorm.DB, mentorID uint) error {
	var totalStake int64
	tx.Model(&models.MentorStake{}).
		Where("mentor_id = ? AND status = ?", mentorID, models.MentorStakeStatusActive).
		Select("COALESCE(SUM(available_amount), 0)").
		Scan(&totalStake)

	// Mentor 테이블의 total_stake 필드 업데이트 (필드가 있다고 가정)
	return tx.Model(&models.Mentor{}).
		Where("id = ?", mentorID).
		Update("total_staked", totalStake).Error
}

func (s *MentorStakingService) startSlashEventReview(slashEventID uint) {
	// 비동기 검토 프로세스 시작
	// 실제 구현에서는 더 복잡한 검토 로직이 필요
	time.Sleep(1 * time.Hour) // 1시간 후 자동 검토 시작
	
	var slashEvent models.MentorSlashEvent
	if err := s.db.First(&slashEvent, slashEventID).Error; err != nil {
		return
	}

	slashEvent.Status = models.SlashEventStatusReviewing
	s.db.Save(&slashEvent)
}

func (s *MentorStakingService) updateMentorPerformanceAfterSlash(tx *gorm.DB, mentorID uint, slashEvent *models.MentorSlashEvent) error {
	// 슬래싱 이후 멘토 성과 지표 업데이트
	// 신뢰도 하락, 위험 점수 상승 등
	return nil
}

func (s *MentorStakingService) transferSlashedTokensToRewardPool(tx *gorm.DB, amount int64) error {
	// 슬래싱된 토큰을 플랫폼 보상 풀로 이동
	// 실제 구현에서는 토큰 이동 로직 필요
	return nil
}

// 성과 지표 계산 관련 helper functions
func (s *MentorStakingService) calculateMentorshipStats(mentorID uint, startDate, endDate time.Time) (map[string]interface{}, error) {
	stats := make(map[string]interface{})
	
	// 총 멘티 수
	var totalMentees int64
	s.db.Model(&models.MentoringSession{}).
		Where("mentor_id = ? AND created_at BETWEEN ? AND ?", mentorID, startDate, endDate).
		Distinct("mentee_id").Count(&totalMentees)
	
	// 활성 멘티 수
	var activeMentees int64
	s.db.Model(&models.MentoringSession{}).
		Where("mentor_id = ? AND status = ? AND created_at BETWEEN ? AND ?", 
			mentorID, "active", startDate, endDate).
		Distinct("mentee_id").Count(&activeMentees)
	
	// 완료된 멘토링 수
	var completed int64
	s.db.Model(&models.MentoringSession{}).
		Where("mentor_id = ? AND status = ? AND updated_at BETWEEN ? AND ?", 
			mentorID, "completed", startDate, endDate).
		Count(&completed)
	
	stats["total_mentees"] = int(totalMentees)
	stats["active_mentees"] = int(activeMentees)
	stats["completed"] = int(completed)
	
	return stats, nil
}

func (s *MentorStakingService) calculateMilestoneStats(mentorID uint, startDate, endDate time.Time) (map[string]interface{}, error) {
	stats := make(map[string]interface{})
	
	// 멘토가 관련된 마일스톤들 조회 (멘토링을 통해)
	var totalMilestones int64
	var successfulMilestones int64
	
	// 실제 구현에서는 더 복잡한 쿼리가 필요
	s.db.Raw(`
		SELECT COUNT(*) as total,
			   SUM(CASE WHEN m.status = 'completed' THEN 1 ELSE 0 END) as successful
		FROM milestones m
		JOIN mentoring_sessions ms ON m.project_id = ms.project_id
		WHERE ms.mentor_id = ? AND m.created_at BETWEEN ? AND ?
	`, mentorID, startDate, endDate).
		Scan(&struct {
			Total      int64
			Successful int64
		}{Total: totalMilestones, Successful: successfulMilestones})
	
	successRate := 0.0
	if totalMilestones > 0 {
		successRate = float64(successfulMilestones) / float64(totalMilestones)
	}
	
	stats["total"] = int(totalMilestones)
	stats["successful"] = int(successfulMilestones)
	stats["success_rate"] = successRate
	
	return stats, nil
}

func (s *MentorStakingService) calculateParticipationStats(mentorID uint, startDate, endDate time.Time) (map[string]interface{}, error) {
	stats := make(map[string]interface{})
	
	// 기본값으로 채우기 (실제 구현에서는 정확한 데이터 필요)
	stats["total_sessions"] = 10
	stats["attendance_rate"] = 0.9
	stats["response_time"] = 4 // 4시간
	stats["session_rating"] = 4.5
	
	return stats, nil
}

func (s *MentorStakingService) calculateSatisfactionStats(mentorID uint, startDate, endDate time.Time) (map[string]interface{}, error) {
	stats := make(map[string]interface{})
	
	// 기본값으로 채우기 (실제 구현에서는 정확한 데이터 필요)
	stats["mentee_rating"] = 4.3
	stats["feedback_score"] = 4.2
	stats["retention_rate"] = 0.85
	stats["referral_rate"] = 0.3
	
	return stats, nil
}

func (s *MentorStakingService) calculateEconomicStats(mentorID uint, startDate, endDate time.Time) (map[string]interface{}, error) {
	stats := make(map[string]interface{})
	
	// 기본값으로 채우기 (실제 구현에서는 정확한 데이터 필요)
	stats["total_revenue"] = int64(50000)
	stats["avg_revenue"] = int64(5000)
	stats["profit_margin"] = 0.7
	
	return stats, nil
}

func (s *MentorStakingService) calculateRiskStats(mentorID uint, startDate, endDate time.Time) (map[string]interface{}, error) {
	stats := make(map[string]interface{})
	
	// 실제 슬래싱 데이터 조회
	var slashCount int64
	var slashedAmount int64
	
	s.db.Model(&models.MentorSlashEvent{}).
		Where("mentor_id = ? AND status = ? AND created_at BETWEEN ? AND ?", 
			mentorID, models.SlashEventStatusApproved, startDate, endDate).
		Count(&slashCount)
	
	s.db.Model(&models.MentorSlashEvent{}).
		Where("mentor_id = ? AND status = ? AND created_at BETWEEN ? AND ?", 
			mentorID, models.SlashEventStatusApproved, startDate, endDate).
		Select("COALESCE(SUM(slashed_amount), 0)").
		Scan(&slashedAmount)
	
	stats["complaints"] = 0      // 실제 구현 필요
	stats["disputes"] = 0        // 실제 구현 필요
	stats["slashes"] = int(slashCount)
	stats["slashed_amount"] = slashedAmount
	
	return stats, nil
}

func (s *MentorStakingService) calculatePerformanceScore(mentorship, milestone, participation, satisfaction map[string]interface{}) float64 {
	// 가중 평균으로 성과 점수 계산
	successRate := milestone["success_rate"].(float64)
	attendanceRate := participation["attendance_rate"].(float64)
	menteeRating := satisfaction["mentee_rating"].(float64) / 5.0 // 5점 만점을 1.0으로 정규화
	retentionRate := satisfaction["retention_rate"].(float64)
	
	score := (successRate*0.3 + attendanceRate*0.2 + menteeRating*0.3 + retentionRate*0.2) * 100
	return math.Min(score, 100.0)
}

func (s *MentorStakingService) calculateRiskScore(risk map[string]interface{}) float64 {
	slashCount := risk["slashes"].(int)
	complaints := risk["complaints"].(int)
	disputes := risk["disputes"].(int)
	
	// 위험 요소가 많을수록 높은 점수
	score := float64(slashCount*10 + complaints*5 + disputes*8)
	return math.Min(score, 100.0)
}

func (s *MentorStakingService) calculateQualityScore(satisfaction, participation map[string]interface{}) float64 {
	feedbackScore := satisfaction["feedback_score"].(float64) / 5.0
	sessionRating := participation["session_rating"].(float64) / 5.0
	
	score := (feedbackScore*0.5 + sessionRating*0.5) * 100
	return math.Min(score, 100.0)
}

// UnstakeMentor 멘토 스테이킹 해제
func (s *MentorStakingService) UnstakeMentor(stakeID uint, userID uint) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var stake models.MentorStake
		if err := tx.Where("id = ? AND user_id = ?", stakeID, userID).First(&stake).Error; err != nil {
			return fmt.Errorf("스테이킹을 찾을 수 없습니다: %w", err)
		}

		if stake.Status != models.MentorStakeStatusActive {
			return errors.New("활성화된 스테이킹이 아닙니다")
		}

		if time.Now().Before(stake.UnlockDate) {
			return errors.New("아직 잠금 해제 기간이 되지 않았습니다")
		}

		if stake.LockedAmount > 0 {
			return errors.New("슬래싱된 금액이 있어 전체 해제할 수 없습니다")
		}

		// 지갑으로 반환
		var userWallet models.UserWallet
		if err := tx.Where("user_id = ?", userID).First(&userWallet).Error; err != nil {
			return fmt.Errorf("지갑 조회 실패: %w", err)
		}

		userWallet.BlueprintBalance += stake.AvailableAmount
		if err := tx.Save(&userWallet).Error; err != nil {
			return fmt.Errorf("지갑 업데이트 실패: %w", err)
		}

		stake.Status = models.MentorStakeStatusWithdrawn
		now := time.Now()
		stake.UnstakedAt = &now
		stake.AvailableAmount = 0

		if err := tx.Save(&stake).Error; err != nil {
			return fmt.Errorf("스테이킹 상태 업데이트 실패: %w", err)
		}

		return s.updateMentorTotalStake(tx, stake.MentorID)
	})
}

// GetUserStakes 사용자 스테이킹 목록 조회
func (s *MentorStakingService) GetUserStakes(userID uint, page, limit int, status, stakeType string) (interface{}, error) {
	offset := (page - 1) * limit
	
	query := s.db.Model(&models.MentorStake{}).Where("user_id = ?", userID)

	if status != "" {
		query = query.Where("status = ?", status)
	}
	if stakeType != "" {
		query = query.Where("stake_type = ?", stakeType)
	}

	var stakes []models.MentorStake
	var total int64
	
	query.Count(&total)
	query.Offset(offset).Limit(limit).Preload("Mentor").Find(&stakes)

	return gin.H{
		"stakes": stakes,
		"pagination": gin.H{
			"page":        page,
			"limit":       limit,
			"total":       total,
			"total_pages": (total + int64(limit) - 1) / int64(limit),
		},
	}, nil
}

// GetMentorStakeInfo 멘토 스테이킹 정보 조회
func (s *MentorStakingService) GetMentorStakeInfo(mentorID uint) (*models.MentorStakeResponse, error) {
	var mentor models.Mentor
	if err := s.db.First(&mentor, mentorID).Error; err != nil {
		return nil, fmt.Errorf("멘토를 찾을 수 없습니다: %w", err)
	}

	var stakes []models.MentorStake
	s.db.Where("mentor_id = ? AND status = ?", mentorID, models.MentorStakeStatusActive).
		Preload("User").Find(&stakes)

	performance, _ := s.CalculatePerformanceMetrics(mentorID, models.MetricPeriodMonthly)

	var recentSlashes []models.MentorSlashEvent
	s.db.Where("mentor_id = ?", mentorID).
		Order("created_at DESC").Limit(5).Find(&recentSlashes)

	var pendingRewards []models.MentorStakeReward
	s.db.Where("mentor_id = ? AND status = ?", mentorID, "pending").Find(&pendingRewards)

	totalStaked := s.calculateTotalStaked(stakes)
	statistics := s.calculateMentorStatistics(mentorID, stakes)

	return &models.MentorStakeResponse{
		Stake:          stakes[0], // 첫 번째 스테이킹 (여러 개일 수 있음)
		Performance:    *performance,
		RecentSlashes:  recentSlashes,
		PendingRewards: pendingRewards,
		Statistics: models.MentorStakeStatistics{
			TotalStaked:     totalStaked,
			TotalSlashed:    statistics["total_slashed"].(int64),
			TotalRewards:    statistics["total_rewards"].(int64),
			CurrentAPY:      statistics["current_apy"].(float64),
			RiskScore:       statistics["risk_score"].(float64),
			SlashingHistory: statistics["slashing_history"].(int),
			StakingRank:     statistics["staking_rank"].(int),
		},
	}, nil
}

// GetMentorByUserID 사용자 ID로 멘토 정보 조회
func (s *MentorStakingService) GetMentorByUserID(userID uint) (*models.Mentor, error) {
	var mentor models.Mentor
	if err := s.db.Where("user_id = ?", userID).First(&mentor).Error; err != nil {
		return nil, fmt.Errorf("멘토 정보를 찾을 수 없습니다: %w", err)
	}
	return &mentor, nil
}

// GetMentorDashboard 멘토 대시보드 조회
func (s *MentorStakingService) GetMentorDashboard(mentorID uint) (*models.MentorDashboardResponse, error) {
	var stakes []models.MentorStake
	s.db.Where("mentor_id = ?", mentorID).Find(&stakes)

	performance, _ := s.CalculatePerformanceMetrics(mentorID, models.MetricPeriodMonthly)

	var slashEvents []models.MentorSlashEvent
	s.db.Where("mentor_id = ?", mentorID).Order("created_at DESC").Find(&slashEvents)

	var rewards []models.MentorStakeReward
	s.db.Where("mentor_id = ?", mentorID).Order("created_at DESC").Find(&rewards)

	statistics := s.calculateMentorStatistics(mentorID, stakes)

	recommendations := s.generateRecommendations(performance, slashEvents)

	return &models.MentorDashboardResponse{
		Stakes:      stakes,
		Performance: *performance,
		SlashEvents: slashEvents,
		Rewards:     rewards,
		Statistics: models.MentorStakeStatistics{
			TotalStaked:     statistics["total_staked"].(int64),
			TotalSlashed:    statistics["total_slashed"].(int64),
			TotalRewards:    statistics["total_rewards"].(int64),
			CurrentAPY:      statistics["current_apy"].(float64),
			RiskScore:       statistics["risk_score"].(float64),
			SlashingHistory: statistics["slashing_history"].(int),
			StakingRank:     statistics["staking_rank"].(int),
		},
		Recommendations: recommendations,
	}, nil
}

// GetMentorSlashEvents 멘토 슬래싱 이벤트 목록 조회
func (s *MentorStakingService) GetMentorSlashEvents(mentorID uint, page, limit int, status, slashType string) (interface{}, error) {
	offset := (page - 1) * limit
	
	query := s.db.Model(&models.MentorSlashEvent{}).Where("mentor_id = ?", mentorID)

	if status != "" {
		query = query.Where("status = ?", status)
	}
	if slashType != "" {
		query = query.Where("slash_type = ?", slashType)
	}

	var events []models.MentorSlashEvent
	var total int64
	
	query.Count(&total)
	query.Offset(offset).Limit(limit).Preload("Reporter").Preload("Reviewer").Find(&events)

	return gin.H{
		"slash_events": events,
		"pagination": gin.H{
			"page":        page,
			"limit":       limit,
			"total":       total,
			"total_pages": (total + int64(limit) - 1) / int64(limit),
		},
	}, nil
}

// GetStakingStats 스테이킹 통계 조회
func (s *MentorStakingService) GetStakingStats(period, mentorID string) (interface{}, error) {
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

	query := s.db.Model(&models.MentorStake{})
	if mentorID != "" {
		query = query.Where("mentor_id = ?", mentorID)
	}

	var totalStakes int64
	var totalAmount int64
	var avgAmount float64

	query.Where("created_at BETWEEN ? AND ?", startDate, endDate).Count(&totalStakes)
	query.Where("created_at BETWEEN ? AND ?", startDate, endDate).
		Select("COALESCE(SUM(amount), 0)").Scan(&totalAmount)
	query.Where("created_at BETWEEN ? AND ?", startDate, endDate).
		Select("COALESCE(AVG(amount), 0)").Scan(&avgAmount)

	return gin.H{
		"period":       period,
		"total_stakes": totalStakes,
		"total_amount": totalAmount,
		"avg_amount":   avgAmount,
		"updated_at":   endDate,
	}, nil
}

// GetTopMentors 상위 멘토 목록 조회
func (s *MentorStakingService) GetTopMentors(limit int, sortBy, category string) (interface{}, error) {
	query := s.db.Model(&models.Mentor{})

	if category != "" {
		query = query.Where("JSON_CONTAINS(expertise_areas, ?)", fmt.Sprintf(`"%s"`, category))
	}

	var mentors []models.Mentor
	
	switch sortBy {
	case "total_staked":
		query = query.Order("total_staked DESC")
	case "performance_score":
		query = query.Joins("LEFT JOIN mentor_performance_metrics ON mentors.id = mentor_performance_metrics.mentor_id").
			Order("mentor_performance_metrics.performance_score DESC")
	case "success_rate":
		query = query.Joins("LEFT JOIN mentor_performance_metrics ON mentors.id = mentor_performance_metrics.mentor_id").
			Order("mentor_performance_metrics.success_rate DESC")
	default:
		query = query.Order("total_staked DESC")
	}

	query.Limit(limit).Find(&mentors)

	return mentors, nil
}

// Helper methods
func (s *MentorStakingService) calculateMentorStatistics(mentorID uint, stakes []models.MentorStake) map[string]interface{} {
	stats := make(map[string]interface{})
	
	totalStaked := s.calculateTotalStaked(stakes)
	
	var totalSlashed int64
	s.db.Model(&models.MentorSlashEvent{}).
		Where("mentor_id = ? AND status = ?", mentorID, models.SlashEventStatusApproved).
		Select("COALESCE(SUM(slashed_amount), 0)").Scan(&totalSlashed)

	var totalRewards int64
	s.db.Model(&models.MentorStakeReward{}).
		Where("mentor_id = ? AND status = ?", mentorID, "distributed").
		Select("COALESCE(SUM(amount), 0)").Scan(&totalRewards)

	stats["total_staked"] = totalStaked
	stats["total_slashed"] = totalSlashed
	stats["total_rewards"] = totalRewards
	stats["current_apy"] = 12.5 // 임시값
	stats["risk_score"] = 25.0  // 임시값
	stats["slashing_history"] = 1 // 임시값
	stats["staking_rank"] = 10     // 임시값
	
	return stats
}

func (s *MentorStakingService) generateRecommendations(performance *models.MentorPerformanceMetric, slashEvents []models.MentorSlashEvent) []string {
	var recommendations []string
	
	if performance.SuccessRate < 0.7 {
		recommendations = append(recommendations, "성공률이 낮습니다. 멘토링 품질 개선이 필요합니다.")
	}
	
	if performance.AttendanceRate < 0.8 {
		recommendations = append(recommendations, "출석률을 높여 멘티와의 소통을 늘리세요.")
	}
	
	if len(slashEvents) > 0 {
		recommendations = append(recommendations, "최근 슬래싱 이벤트가 있었습니다. 멘토링 윤리를 재검토하세요.")
	}
	
	if len(recommendations) == 0 {
		recommendations = append(recommendations, "훌륭한 멘토링을 유지하고 계십니다!")
	}
	
	return recommendations
}