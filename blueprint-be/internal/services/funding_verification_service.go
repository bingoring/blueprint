package services

import (
	"blueprint/internal/models"
	"fmt"
	"log"
	"strings"
	"time"

	"gorm.io/gorm"
)

// 🏛️ 마일스톤 시장성 검증 서비스 (Market Viability Verification)
type FundingVerificationService struct {
	db         *gorm.DB
	sseService *SSEService
}

// NewFundingVerificationService 펀딩 검증 서비스 생성자
func NewFundingVerificationService(db *gorm.DB, sseService *SSEService) *FundingVerificationService {
	return &FundingVerificationService{
		db:         db,
		sseService: sseService,
	}
}

// StartFundingPhase 마일스톤의 펀딩 단계 시작
func (fv *FundingVerificationService) StartFundingPhase(milestoneID uint) error {
	log.Printf("🚀 Starting funding phase for milestone %d", milestoneID)

	// 트랜잭션 시작
	tx := fv.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 마일스톤 조회
	var milestone models.Milestone
	if err := tx.Where("id = ?", milestoneID).First(&milestone).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("milestone not found: %v", err)
	}

	// 이미 펀딩 단계이거나 더 진행된 상태면 건너뜀
	if milestone.Status != models.MilestoneStatusProposal {
		tx.Rollback()
		return fmt.Errorf("milestone %d is not in proposal status (current: %s)", milestoneID, milestone.Status)
	}

	// 펀딩 단계 시작
	milestone.StartFundingPhase()

	// 카테고리별 최소 자본 요구액 설정
	milestone.MinViableCapital = fv.calculateMinViableCapital(&milestone)

	if err := tx.Save(&milestone).Error; err != nil {
		// 컬럼이 존재하지 않는 경우 로그만 남기고 넘어감
		if fv.isColumnNotExistsError(err) {
			tx.Rollback()
			log.Printf("📋 Funding columns not available - cannot start funding for milestone %d", milestoneID)
			return fmt.Errorf("funding system not available - database schema needs migration")
		}
		tx.Rollback()
		return fmt.Errorf("failed to update milestone: %v", err)
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	log.Printf("✅ Funding phase started for milestone %d (MVC: $%.2f, Duration: %d days)",
		milestoneID, float64(milestone.MinViableCapital)/100, milestone.FundingDuration)

	// 실시간 알림 브로드캐스트
	fv.broadcastFundingUpdate(milestoneID, "funding_started", map[string]interface{}{
		"milestone_id":         milestoneID,
		"min_viable_capital":   milestone.MinViableCapital,
		"funding_end_date":     milestone.FundingEndDate,
		"funding_duration":     milestone.FundingDuration,
	})

	return nil
}

// UpdateTVL 마일스톤의 총 베팅액 업데이트 (거래 발생 시 호출)
func (fv *FundingVerificationService) UpdateTVL(milestoneID uint, optionID string, additionalAmount int64) error {
	tx := fv.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var milestone models.Milestone
	if err := tx.Where("id = ?", milestoneID).First(&milestone).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("milestone not found: %v", err)
	}

	// TVL 업데이트 (새 컬럼이 없는 경우 gracefully 처리)
	milestone.CurrentTVL += additionalAmount
	milestone.FundingProgress = milestone.CalculateFundingProgress()

	if err := tx.Save(&milestone).Error; err != nil {
		// 컬럼이 존재하지 않는 경우 로그만 남기고 넘어감
		if fv.isColumnNotExistsError(err) {
			tx.Rollback()
			log.Printf("📋 Funding columns not available - skipping TVL update for milestone %d", milestoneID)
			return nil
		}
		tx.Rollback()
		return fmt.Errorf("failed to update milestone TVL: %v", err)
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	log.Printf("📊 TVL updated for milestone %d: $%.2f (+$%.2f)",
		milestoneID, float64(milestone.CurrentTVL)/100, float64(additionalAmount)/100)

	// 펀딩 목표 달성 확인
	if milestone.Status == models.MilestoneStatusFunding && milestone.HasReachedMinViableCapital() {
		log.Printf("🎉 Milestone %d has reached minimum viable capital!", milestoneID)
		fv.broadcastFundingUpdate(milestoneID, "funding_target_reached", map[string]interface{}{
			"milestone_id":    milestoneID,
			"current_tvl":     milestone.CurrentTVL,
			"funding_progress": milestone.FundingProgress,
		})
	}

	// 실시간 진행률 업데이트
	fv.broadcastFundingUpdate(milestoneID, "tvl_updated", map[string]interface{}{
		"milestone_id":     milestoneID,
		"current_tvl":      milestone.CurrentTVL,
		"funding_progress": milestone.FundingProgress,
		"additional_amount": additionalAmount,
	})

	return nil
}

// ProcessExpiredFunding 만료된 펀딩들 처리 (스케줄러가 주기적으로 호출)
func (fv *FundingVerificationService) ProcessExpiredFunding() error {
	log.Printf("🔄 Processing expired funding milestones...")

	// 펀딩 만료된 마일스톤들 조회
	var milestones []models.Milestone
	if err := fv.db.Where("status = ? AND funding_end_date <= ?",
		models.MilestoneStatusFunding, time.Now()).Find(&milestones).Error; err != nil {

		// 컬럼이 존재하지 않는 경우 (기존 데이터베이스) - 정상적인 상황
		if fv.isColumnNotExistsError(err) {
			log.Printf("📋 Funding columns not found - skipping expired funding processing")
			return nil
		}
		return fmt.Errorf("failed to query expired milestones: %v", err)
	}

	for _, milestone := range milestones {
		if err := fv.processSingleExpiredMilestone(&milestone); err != nil {
			log.Printf("❌ Failed to process expired milestone %d: %v", milestone.ID, err)
			continue
		}
	}

	log.Printf("✅ Processed %d expired funding milestones", len(milestones))
	return nil
}

// processSingleExpiredMilestone 개별 만료 마일스톤 처리
func (fv *FundingVerificationService) processSingleExpiredMilestone(milestone *models.Milestone) error {
	tx := fv.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if milestone.HasReachedMinViableCapital() {
		// ✅ 펀딩 성공: 활성화
		milestone.Status = models.MilestoneStatusActive
		log.Printf("✅ Milestone %d FUNDED successfully (TVL: $%.2f)",
			milestone.ID, float64(milestone.CurrentTVL)/100)

		// 실시간 알림
		fv.broadcastFundingUpdate(milestone.ID, "funding_successful", map[string]interface{}{
			"milestone_id": milestone.ID,
			"current_tvl":  milestone.CurrentTVL,
		})

	} else {
		// ❌ 펀딩 실패: 거부 및 자금 반환 처리
		milestone.Status = models.MilestoneStatusRejected
		log.Printf("❌ Milestone %d REJECTED due to insufficient funding (TVL: $%.2f, Required: $%.2f)",
			milestone.ID, float64(milestone.CurrentTVL)/100, float64(milestone.MinViableCapital)/100)

		// 자금 반환 처리 (비동기로 처리)
		go fv.refundFailedFunding(milestone.ID)

		// 실시간 알림
		fv.broadcastFundingUpdate(milestone.ID, "funding_failed", map[string]interface{}{
			"milestone_id":       milestone.ID,
			"current_tvl":        milestone.CurrentTVL,
			"min_viable_capital": milestone.MinViableCapital,
		})
	}

	if err := tx.Save(milestone).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update milestone status: %v", err)
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	return nil
}

// refundFailedFunding 실패한 펀딩의 자금 반환 처리
func (fv *FundingVerificationService) refundFailedFunding(milestoneID uint) {
	log.Printf("💰 Processing refunds for failed milestone %d", milestoneID)

	// 해당 마일스톤의 모든 주문 조회
	var orders []models.Order
	if err := fv.db.Where("milestone_id = ? AND status IN ?", milestoneID,
		[]models.OrderStatus{models.OrderStatusPending, models.OrderStatusPartial}).Find(&orders).Error; err != nil {
		log.Printf("❌ Failed to query orders for refund: %v", err)
		return
	}

	for _, order := range orders {
		// 각 사용자의 지갑에 자금 반환
		if err := fv.refundOrderAmount(&order); err != nil {
			log.Printf("❌ Failed to refund order %d: %v", order.ID, err)
			continue
		}
	}

	log.Printf("✅ Completed refunds for %d orders", len(orders))
}

// refundOrderAmount 개별 주문의 자금 반환
func (fv *FundingVerificationService) refundOrderAmount(order *models.Order) error {
	if order.Side != models.OrderSideBuy {
		return nil // 매도 주문은 자금이 잠겨있지 않음
	}

	refundAmount := int64(float64(order.Remaining) * order.Price * 100) // 미체결 부분만 반환

	// 지갑 업데이트
	tx := fv.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var wallet models.UserWallet
	if err := tx.Where("user_id = ?", order.UserID).First(&wallet).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("wallet not found for user %d: %v", order.UserID, err)
	}

	// 잠긴 잔액을 가용 잔액으로 이동
	wallet.USDCLockedBalance -= refundAmount
	wallet.USDCBalance += refundAmount

	if err := tx.Save(&wallet).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update wallet: %v", err)
	}

	// 주문 상태를 취소로 변경
	order.Status = models.OrderStatusCancelled
	if err := tx.Save(order).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to cancel order: %v", err)
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit refund transaction: %v", err)
	}

	log.Printf("💰 Refunded $%.2f to user %d for cancelled order %d",
		float64(refundAmount)/100, order.UserID, order.ID)

	return nil
}

// calculateMinViableCapital 카테고리별 최소 자본 요구액 계산
func (fv *FundingVerificationService) calculateMinViableCapital(milestone *models.Milestone) int64 {
	// 프로젝트 정보 로딩
	var project models.Project
	if err := fv.db.Where("id = ?", milestone.ProjectID).First(&project).Error; err != nil {
		log.Printf("❌ Failed to load project for milestone %d: %v", milestone.ID, err)
		return 100000 // 기본값: $1000
	}

	// 카테고리별 최소 자본 요구액 (센트 단위)
	switch project.Category {
	case models.CareerProject:
		return 200000 // $2000 - 커리어는 높은 투자 가치
	case models.BusinessProject:
		return 500000 // $5000 - 비즈니스는 가장 높은 투자 가치
	case models.EducationProject:
		return 150000 // $1500 - 교육은 중간 투자 가치
	case models.PersonalProject:
		return 100000 // $1000 - 개인은 기본 투자 가치
	case models.LifeProject:
		return 75000  // $750 - 라이프스타일은 가장 낮은 투자 가치
	default:
		return 100000 // 기본값
	}
}

// broadcastFundingUpdate 펀딩 상태 실시간 브로드캐스트
func (fv *FundingVerificationService) broadcastFundingUpdate(milestoneID uint, eventType string, data map[string]interface{}) {
	if fv.sseService == nil {
		return
	}

	// MarketUpdateEvent를 사용하여 펀딩 업데이트 브로드캐스트
	marketEvent := MarketUpdateEvent{
		MilestoneID: milestoneID,
		MarketData: map[string]interface{}{
			"event_type": eventType,
			"data":       data,
		},
		Timestamp: time.Now().Unix(),
	}

	fv.sseService.BroadcastMarketUpdate(marketEvent)
}

// GetFundingStats 펀딩 통계 조회
func (fv *FundingVerificationService) GetFundingStats(milestoneID uint) (*FundingStats, error) {
	var milestone models.Milestone
	if err := fv.db.Where("id = ?", milestoneID).First(&milestone).Error; err != nil {
		// 컬럼이 존재하지 않는 경우 기본값으로 응답
		if fv.isColumnNotExistsError(err) {
			return &FundingStats{
				MilestoneID:       milestoneID,
				Status:            models.MilestoneStatusPending, // 기본 상태
				CurrentTVL:        0,
				MinViableCapital:  100000, // 기본값: $1000
				FundingProgress:   0,
				IsActive:          false,
				IsExpired:         false,
				HasReachedTarget:  false,
			}, nil
		}
		return nil, fmt.Errorf("milestone not found: %v", err)
	}

	stats := &FundingStats{
		MilestoneID:       milestoneID,
		Status:            milestone.Status,
		CurrentTVL:        milestone.CurrentTVL,
		MinViableCapital:  milestone.MinViableCapital,
		FundingProgress:   milestone.FundingProgress,
		FundingStartDate:  milestone.FundingStartDate,
		FundingEndDate:    milestone.FundingEndDate,
		FundingDuration:   milestone.FundingDuration,
		IsActive:          milestone.IsFundingActive(),
		IsExpired:         milestone.IsFundingExpired(),
		HasReachedTarget:  milestone.HasReachedMinViableCapital(),
	}

	return stats, nil
}

// FundingStats 펀딩 통계 구조체
type FundingStats struct {
	MilestoneID       uint                `json:"milestone_id"`
	Status            models.MilestoneStatus `json:"status"`
	CurrentTVL        int64               `json:"current_tvl"`
	MinViableCapital  int64               `json:"min_viable_capital"`
	FundingProgress   float64             `json:"funding_progress"`
	FundingStartDate  *time.Time          `json:"funding_start_date,omitempty"`
	FundingEndDate    *time.Time          `json:"funding_end_date,omitempty"`
	FundingDuration   int                 `json:"funding_duration"`
	IsActive          bool                `json:"is_active"`
	IsExpired         bool                `json:"is_expired"`
	HasReachedTarget  bool                `json:"has_reached_target"`
}

// isColumnNotExistsError 컬럼이 존재하지 않는 오류인지 확인
func (fv *FundingVerificationService) isColumnNotExistsError(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()
	// PostgreSQL: column "funding_end_date" does not exist
	// MySQL: Unknown column 'funding_end_date' in 'where clause'
	// SQLite: no such column: funding_end_date
	return (errStr != "" &&
		   (strings.Contains(errStr, `column "funding_end_date" does not exist`) ||
			strings.Contains(errStr, `column "funding_start_date" does not exist`) ||
			strings.Contains(errStr, `column "min_viable_capital" does not exist`) ||
			strings.Contains(errStr, `column "current_tvl" does not exist`) ||
			strings.Contains(errStr, `Unknown column`) && strings.Contains(errStr, `funding_`) ||
			strings.Contains(errStr, `no such column`) && strings.Contains(errStr, `funding_`)))
}
