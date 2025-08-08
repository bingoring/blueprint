package services

import (
	"blueprint/internal/models"
	"blueprint/internal/queue"
	"fmt"
	"log"
	"time"

	"gorm.io/gorm"
)

// TradingService P2P 거래 서비스 (매칭 엔진 기반)
type TradingService struct {
	db             *gorm.DB
	sseService     *SSEService
	queuePublisher *queue.Publisher
	matchingEngine *MatchingEngine
}

// NewTradingService 거래 서비스 생성자
func NewTradingService(db *gorm.DB, sseService *SSEService, matchingEngine *MatchingEngine) *TradingService {
	return &TradingService{
		db:             db,
		sseService:     sseService,
		queuePublisher: queue.NewPublisher(),
		matchingEngine: matchingEngine,
	}
}

// CreateOrder 주문 생성 및 매칭 실행
func (s *TradingService) CreateOrder(userID uint, req models.CreateOrderRequest, ipAddress, userAgent string) (*models.OrderResponse, error) {
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 1. 매수 주문인 경우 지갑 잠금 처리
	if req.Side == models.OrderSideBuy {
		requiredUSDC := int64(float64(req.Quantity) * req.Price * 100) // 확률을 센트로 변환

		var wallet models.UserWallet
		if err := tx.Where("user_id = ?", userID).First(&wallet).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("지갑 조회 실패: %v", err)
		}

		// 잔액에서 잠긴 잔액으로 이동
		if wallet.USDCBalance < requiredUSDC {
			tx.Rollback()
			return nil, fmt.Errorf("USDC 잔액 부족: 필요 $%.2f, 보유 $%.2f",
				float64(requiredUSDC)/100, float64(wallet.USDCBalance)/100)
		}

		wallet.USDCBalance -= requiredUSDC
		wallet.USDCLockedBalance += requiredUSDC

		if err := tx.Save(&wallet).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("지갑 업데이트 실패: %v", err)
		}

		log.Printf("🔒 Locked %d USDC for user %d order", requiredUSDC, userID)
	}

	// 2. 주문 생성
	order := models.Order{
		ProjectID:   req.ProjectID,
		MilestoneID: req.MilestoneID,
		OptionID:    req.OptionID,
		UserID:      userID,
		Type:        req.Type,
		Side:        req.Side,
		Quantity:    req.Quantity,
		Price:       req.Price,
		Remaining:   req.Quantity,
		Status:      models.OrderStatusPending,
		IPAddress:   ipAddress,
		UserAgent:   userAgent,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := tx.Create(&order).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create order: %v", err)
	}

	// 3. 고성능 매칭 엔진으로 매칭 실행
	result, err := s.matchingEngine.SubmitOrder(&order)
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("matching failed: %v", err)
	}

	// 4. 결과 저장 및 브로드캐스트
	var trades []models.Trade
	if result.Executed && len(result.Trades) > 0 {
		trades = result.Trades

		// 실시간 브로드캐스트는 매칭 엔진에서 처리됨
		log.Printf("✅ Order %d executed with %d trades", order.ID, len(trades))
	}

	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %v", err)
	}

	return &models.OrderResponse{
		Order:  order,
		Trades: trades,
	}, nil
}

// GetOrderBook 호가창 조회 (매칭 엔진에서 직접 조회)
func (s *TradingService) GetOrderBook(milestoneID uint, optionID string) (*models.OrderBook, error) {
	return s.matchingEngine.GetOrderBook(milestoneID, optionID), nil
}

// GetMyOrders 내 주문 목록 조회
func (s *TradingService) GetMyOrders(userID uint, status string, limit, offset int) ([]models.Order, error) {
	var orders []models.Order
	query := s.db.Where("user_id = ?", userID)

	if status != "" {
		query = query.Where("status = ?", status)
	}

	err := query.Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&orders).Error

	return orders, err
}

// GetMyTrades 내 거래 내역 조회
func (s *TradingService) GetMyTrades(userID uint, limit, offset int) ([]models.Trade, error) {
	var trades []models.Trade
	err := s.db.Where("buyer_id = ? OR seller_id = ?", userID, userID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&trades).Error

	return trades, err
}

// GetMyPositions 내 포지션 조회
func (s *TradingService) GetMyPositions(userID uint) ([]models.Position, error) {
	var positions []models.Position
	err := s.db.Where("user_id = ? AND quantity != 0", userID).
		Find(&positions).Error

	return positions, err
}

// GetPosition 특정 마일스톤의 포지션 조회
func (s *TradingService) GetPosition(userID uint, milestoneID uint, optionID string) (*models.Position, error) {
	var position models.Position
	err := s.db.Where("user_id = ? AND milestone_id = ? AND option_id = ?",
		userID, milestoneID, optionID).First(&position).Error

	if err == gorm.ErrRecordNotFound {
		return &models.Position{
			UserID:      userID,
			MilestoneID: milestoneID,
			OptionID:    optionID,
			Quantity:    0,
		}, nil
	}

	return &position, err
}

// CancelOrder 주문 취소
func (s *TradingService) CancelOrder(userID uint, orderID uint) error {
	var order models.Order
	err := s.db.Where("id = ? AND user_id = ?", orderID, userID).First(&order).Error
	if err != nil {
		return err
	}

	if order.Status != models.OrderStatusPending && order.Status != models.OrderStatusPartial {
		return fmt.Errorf("cannot cancel order with status: %s", order.Status)
	}

	// 주문 상태 업데이트
	order.Status = models.OrderStatusCancelled
	return s.db.Save(&order).Error
}

// GetRecentTrades 최근 거래 내역 조회
func (s *TradingService) GetRecentTrades(milestoneID uint, optionID string, limit int) ([]models.Trade, error) {
	var trades []models.Trade
	err := s.db.Where("milestone_id = ? AND option_id = ?", milestoneID, optionID).
		Order("created_at DESC").
		Limit(limit).
		Find(&trades).Error

	return trades, err
}

// GetDB 데이터베이스 인스턴스 반환 (핸들러에서 직접 쿼리용)
func (s *TradingService) GetDB() *gorm.DB {
	return s.db
}

// GetOrderTrades 특정 주문의 거래 내역 조회
func (s *TradingService) GetOrderTrades(orderID uint) ([]models.Trade, error) {
	var trades []models.Trade
	err := s.db.Where("buy_order_id = ? OR sell_order_id = ?", orderID, orderID).
		Order("created_at DESC").
		Find(&trades).Error

	return trades, err
}

// GetStats 거래 통계 조회
func (s *TradingService) GetStats() map[string]interface{} {
	stats := make(map[string]interface{})

	// 총 거래 수
	var totalTrades int64
	s.db.Model(&models.Trade{}).Count(&totalTrades)

	// 총 거래량
	var totalVolume int64
	s.db.Model(&models.Trade{}).Select("COALESCE(SUM(total_amount), 0)").Scan(&totalVolume)

	// 활성 주문 수
	var activeOrders int64
	s.db.Model(&models.Order{}).Where("status IN ?", []string{"pending", "partial"}).Count(&activeOrders)

	// 매칭 엔진 통계
	matchingStats := s.matchingEngine.GetStats()

	stats["total_trades"] = totalTrades
	stats["total_volume"] = totalVolume
	stats["active_orders"] = activeOrders
	stats["matching_engine"] = matchingStats

	return stats
}
