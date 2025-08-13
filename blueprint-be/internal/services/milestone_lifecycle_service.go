package services

import (
	"blueprint/internal/models"
	"context"
	"log"
	"sync"
	"time"

	"gorm.io/gorm"
)

// 🔄 마일스톤 라이프사이클 자동 관리 서비스
type MilestoneLifecycleService struct {
	db                      *gorm.DB
	fundingVerificationSvc  *FundingVerificationService

	// 스케줄러 관련
	isRunning               bool
	stopChan                chan struct{}
	ticker                  *time.Ticker
	mutex                   sync.RWMutex

	// 설정
	checkInterval           time.Duration    // 체크 주기 (기본: 1분)
	autoStartFundingDelay   time.Duration    // 제안 생성 후 펀딩 시작까지 대기 시간 (기본: 1시간)
}

// NewMilestoneLifecycleService 라이프사이클 서비스 생성자
func NewMilestoneLifecycleService(db *gorm.DB, fundingVerificationSvc *FundingVerificationService) *MilestoneLifecycleService {
	return &MilestoneLifecycleService{
		db:                      db,
		fundingVerificationSvc:  fundingVerificationSvc,
		isRunning:              false,
		stopChan:               make(chan struct{}),
		checkInterval:          time.Minute,          // 1분마다 체크
		autoStartFundingDelay:  30 * time.Minute,    // 30분 후 자동 펀딩 시작
	}
}

// Start 라이프사이클 관리 시작
func (mls *MilestoneLifecycleService) Start() error {
	mls.mutex.Lock()
	defer mls.mutex.Unlock()

	if mls.isRunning {
		return nil // 이미 실행 중
	}

	mls.ticker = time.NewTicker(mls.checkInterval)
	mls.isRunning = true

	// 백그라운드에서 실행
	go mls.run()

	log.Printf("✅ Milestone lifecycle service started (check interval: %v)", mls.checkInterval)
	return nil
}

// Stop 라이프사이클 관리 중지
func (mls *MilestoneLifecycleService) Stop() error {
	mls.mutex.Lock()
	defer mls.mutex.Unlock()

	if !mls.isRunning {
		return nil // 이미 중지됨
	}

	close(mls.stopChan)
	mls.ticker.Stop()
	mls.isRunning = false

	log.Printf("🛑 Milestone lifecycle service stopped")
	return nil
}

// IsRunning 실행 상태 확인
func (mls *MilestoneLifecycleService) IsRunning() bool {
	mls.mutex.RLock()
	defer mls.mutex.RUnlock()
	return mls.isRunning
}

// run 메인 루프 실행
func (mls *MilestoneLifecycleService) run() {
	log.Printf("🔄 Starting milestone lifecycle management loop...")

	for {
		select {
		case <-mls.stopChan:
			log.Printf("📴 Lifecycle management loop stopped")
			return

		case <-mls.ticker.C:
			// 모든 라이프사이클 단계 처리
			mls.processAllLifecycleStages()
		}
	}
}

// processAllLifecycleStages 모든 라이프사이클 단계들을 순차적으로 처리
func (mls *MilestoneLifecycleService) processAllLifecycleStages() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 1단계: 제안(Proposal) → 펀딩(Funding) 자동 전환
	if err := mls.processProposalToFunding(ctx); err != nil {
		log.Printf("❌ Error processing proposal to funding: %v", err)
	}

	// 2단계: 만료된 펀딩 처리 (펀딩→활성화 또는 펀딩→거부)
	if err := mls.processExpiredFunding(ctx); err != nil {
		log.Printf("❌ Error processing expired funding: %v", err)
	}

	// 3단계: 펀딩이 조기 달성된 경우 즉시 활성화
	if err := mls.processEarlyFundingSuccess(ctx); err != nil {
		log.Printf("❌ Error processing early funding success: %v", err)
	}
}

// processProposalToFunding 제안 상태의 마일스톤들을 펀딩 단계로 전환
func (mls *MilestoneLifecycleService) processProposalToFunding(ctx context.Context) error {
	// 제안 상태이면서 생성된 지 일정 시간이 지난 마일스톤들 조회
	cutoffTime := time.Now().Add(-mls.autoStartFundingDelay)

	var milestones []models.Milestone
	if err := mls.db.WithContext(ctx).Where("status = ? AND created_at <= ?",
		models.MilestoneStatusProposal, cutoffTime).Find(&milestones).Error; err != nil {
		return err
	}

	if len(milestones) == 0 {
		return nil
	}

	log.Printf("🚀 Processing %d milestones ready for funding phase", len(milestones))

	for _, milestone := range milestones {
		if err := mls.fundingVerificationSvc.StartFundingPhase(milestone.ID); err != nil {
			log.Printf("❌ Failed to start funding for milestone %d: %v", milestone.ID, err)
			continue
		}

		log.Printf("✅ Started funding phase for milestone %d (%s)", milestone.ID, milestone.Title)

		// 너무 빠른 처리를 방지하기 위해 잠시 대기
		time.Sleep(100 * time.Millisecond)
	}

	return nil
}

// processExpiredFunding 만료된 펀딩들 처리
func (mls *MilestoneLifecycleService) processExpiredFunding(ctx context.Context) error {
	return mls.fundingVerificationSvc.ProcessExpiredFunding()
}

// processEarlyFundingSuccess 펀딩 목표를 조기 달성한 마일스톤들 즉시 활성화
func (mls *MilestoneLifecycleService) processEarlyFundingSuccess(ctx context.Context) error {
	// 펀딩 중이면서 목표를 달성한 마일스톤들 조회
	var milestones []models.Milestone
	if err := mls.db.WithContext(ctx).Where("status = ?",
		models.MilestoneStatusFunding).Find(&milestones).Error; err != nil {
		return err
	}

	var activatedCount int
	for _, milestone := range milestones {
		// 목표 달성 및 최소 펀딩 기간 경과 확인
		if milestone.HasReachedMinViableCapital() && mls.hasMinFundingPeriodPassed(&milestone) {
			// 즉시 활성화
			milestone.Status = models.MilestoneStatusActive

			if err := mls.db.WithContext(ctx).Save(&milestone).Error; err != nil {
				log.Printf("❌ Failed to activate milestone %d early: %v", milestone.ID, err)
				continue
			}

			activatedCount++
			log.Printf("🎉 Early activated milestone %d after reaching funding target", milestone.ID)

			// 실시간 알림 (fundingVerificationSvc를 통해)
			go func(milestoneID uint) {
				if mls.fundingVerificationSvc != nil {
					mls.fundingVerificationSvc.broadcastFundingUpdate(milestoneID, "early_activation", map[string]interface{}{
						"milestone_id": milestoneID,
						"reason":       "funding_target_reached_early",
					})
				}
			}(milestone.ID)
		}
	}

	if activatedCount > 0 {
		log.Printf("✅ Early activated %d milestones", activatedCount)
	}

	return nil
}

// hasMinFundingPeriodPassed 최소 펀딩 기간이 지났는지 확인 (조기 활성화 남용 방지)
func (mls *MilestoneLifecycleService) hasMinFundingPeriodPassed(milestone *models.Milestone) bool {
	if milestone.FundingStartDate == nil {
		return false
	}

	// 최소 2시간은 펀딩을 진행해야 함 (너무 빠른 활성화 방지)
	minFundingDuration := 2 * time.Hour
	return time.Now().Sub(*milestone.FundingStartDate) >= minFundingDuration
}

// GetLifecycleStats 라이프사이클 통계 조회
func (mls *MilestoneLifecycleService) GetLifecycleStats() (*LifecycleStats, error) {
	stats := &LifecycleStats{
		IsRunning:     mls.IsRunning(),
		CheckInterval: mls.checkInterval,
	}

	// 상태별 마일스톤 수 조회
	statusCounts := make(map[models.MilestoneStatus]int)

	var results []struct {
		Status models.MilestoneStatus `gorm:"column:status"`
		Count  int                    `gorm:"column:count"`
	}

	if err := mls.db.Model(&models.Milestone{}).
		Select("status, count(*) as count").
		Group("status").
		Find(&results).Error; err != nil {
		return nil, err
	}

	for _, result := range results {
		statusCounts[result.Status] = result.Count
	}

	stats.ProposalCount = statusCounts[models.MilestoneStatusProposal]
	stats.FundingCount = statusCounts[models.MilestoneStatusFunding]
	stats.ActiveCount = statusCounts[models.MilestoneStatusActive]
	stats.RejectedCount = statusCounts[models.MilestoneStatusRejected]
	stats.CompletedCount = statusCounts[models.MilestoneStatusCompleted]

	return stats, nil
}

// ForceStartFunding 특정 마일스톤의 펀딩을 강제로 시작 (관리자용)
func (mls *MilestoneLifecycleService) ForceStartFunding(milestoneID uint) error {
	log.Printf("🔧 Force starting funding for milestone %d", milestoneID)
	return mls.fundingVerificationSvc.StartFundingPhase(milestoneID)
}

// ForceProcessExpired 만료된 펀딩들 강제 처리 (관리자용)
func (mls *MilestoneLifecycleService) ForceProcessExpired() error {
	log.Printf("🔧 Force processing expired funding milestones")
	return mls.fundingVerificationSvc.ProcessExpiredFunding()
}

// UpdateSettings 설정 업데이트
func (mls *MilestoneLifecycleService) UpdateSettings(checkInterval time.Duration, autoStartDelay time.Duration) {
	mls.mutex.Lock()
	defer mls.mutex.Unlock()

	mls.checkInterval = checkInterval
	mls.autoStartFundingDelay = autoStartDelay

	// 실행 중인 경우 ticker 업데이트
	if mls.isRunning && mls.ticker != nil {
		mls.ticker.Reset(mls.checkInterval)
	}

	log.Printf("⚙️ Updated lifecycle settings: check_interval=%v, auto_start_delay=%v",
		checkInterval, autoStartDelay)
}

// LifecycleStats 라이프사이클 통계 구조체
type LifecycleStats struct {
	IsRunning        bool          `json:"is_running"`
	CheckInterval    time.Duration `json:"check_interval"`
	ProposalCount    int           `json:"proposal_count"`
	FundingCount     int           `json:"funding_count"`
	ActiveCount      int           `json:"active_count"`
	RejectedCount    int           `json:"rejected_count"`
	CompletedCount   int           `json:"completed_count"`
}
