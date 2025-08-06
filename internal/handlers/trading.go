package handlers

import (
	"blueprint/internal/middleware"
	"blueprint/internal/models"
	"blueprint/internal/services"
	"fmt"
	"io"
	"log"
	"strconv"
	"time"

	"gorm.io/gorm"

	"blueprint/internal/queue"

	"github.com/gin-gonic/gin"
)

// TradingHandler P2P 거래 핸들러 (폴리마켓 스타일)
type TradingHandler struct {
	tradingService       *services.TradingService
	probabilityValidator *services.ProbabilityValidator
}

// NewTradingHandler 거래 핸들러 생성자
func NewTradingHandler(tradingService *services.TradingService) *TradingHandler {
	return &TradingHandler{
		tradingService:       tradingService,
		probabilityValidator: services.NewProbabilityValidator(),
	}
}

// CreateOrder 주문 생성 (매수/매도)
// POST /api/v1/orders
func (h *TradingHandler) CreateOrder(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		middleware.Unauthorized(c, "User not authenticated")
		return
	}

	var req models.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.BadRequest(c, "Invalid request format")
		return
	}

	// 🎯 폴리마켓 스타일 확률 검증
	if err := h.probabilityValidator.ValidateOrderPrice(req.Price, req.Type); err != nil {
		middleware.BadRequest(c, fmt.Sprintf("Invalid order price: %v", err))
		return
	}

	// 💰 USDC 잔액 검증 (매수 주문만)
	if req.Side == models.OrderSideBuy {
		var wallet models.UserWallet
		if err := h.tradingService.GetDB().Where("user_id = ?", userID).First(&wallet).Error; err != nil {
			middleware.InternalServerError(c, "지갑 조회 실패")
			return
		}

		// 필요 USDC 계산: 수량 × 가격 (센트 단위)
		requiredUSDC := int64(float64(req.Quantity) * req.Price * 100) // 확률을 센트로 변환
		if wallet.USDCBalance < requiredUSDC {
			middleware.BadRequest(c, fmt.Sprintf("USDC 잔액 부족: 필요 $%.2f, 보유 $%.2f",
				float64(requiredUSDC)/100, float64(wallet.USDCBalance)/100))
			return
		}
	}

	// IP와 User-Agent 추출
	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	response, err := h.tradingService.CreateOrder(
		userID.(uint),
		req,
		ipAddress,
		userAgent,
	)
	if err != nil {
		middleware.InternalServerError(c, err.Error())
		return
	}

	middleware.Success(c, response, "주문이 성공적으로 생성되었습니다")
}

// GetOrderBook 호가창 조회
// GET /api/v1/milestones/:id/orderbook/:option
func (h *TradingHandler) GetOrderBook(c *gin.Context) {
	milestoneIDStr := c.Param("id")
	milestoneID, err := strconv.ParseUint(milestoneIDStr, 10, 32)
	if err != nil {
		middleware.BadRequest(c, "Invalid milestone ID")
		return
	}

	optionID := c.Param("option")
	if optionID == "" {
		middleware.BadRequest(c, "Option ID is required")
		return
	}

	orderBook, err := h.tradingService.GetOrderBook(uint(milestoneID), optionID)
	if err != nil {
		middleware.InternalServerError(c, err.Error())
		return
	}

	result := gin.H{
		"order_book": *orderBook,
	}

	middleware.Success(c, result, "호가창 조회 성공")
}

// GetMyOrders 내 주문 내역 조회
// GET /api/v1/orders/my
func (h *TradingHandler) GetMyOrders(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		middleware.Unauthorized(c, "User not authenticated")
		return
	}

	// 쿼리 파라미터
	status := c.Query("status")
	milestoneIDStr := c.Query("milestone_id")
	limit := c.DefaultQuery("limit", "50")

	// 필터 조건 구성
	conditions := map[string]interface{}{
		"user_id": userID,
	}

	if status != "" {
		conditions["status"] = status
	}

	if milestoneIDStr != "" {
		milestoneID, err := strconv.ParseUint(milestoneIDStr, 10, 32)
		if err == nil {
			conditions["milestone_id"] = uint(milestoneID)
		}
	}

	var orders []models.Order
	query := h.tradingService.GetDB().Where(conditions).Order("created_at DESC")

	limitInt, err := strconv.Atoi(limit)
	if err == nil && limitInt > 0 {
		query = query.Limit(limitInt)
	}

	if err := query.Find(&orders).Error; err != nil {
		middleware.InternalServerError(c, err.Error())
		return
	}

	middleware.Success(c, orders, "내 주문 내역 조회 성공")
}

// GetMyTrades 내 거래 내역 조회
// GET /api/v1/trades/my
func (h *TradingHandler) GetMyTrades(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		middleware.Unauthorized(c, "User not authenticated")
		return
	}

	milestoneIDStr := c.Query("milestone_id")
	limit := c.DefaultQuery("limit", "50")

	var trades []models.Trade
	query := h.tradingService.GetDB().Where("buyer_id = ? OR seller_id = ?", userID, userID).
		Order("created_at DESC")

	if milestoneIDStr != "" {
		milestoneID, err := strconv.ParseUint(milestoneIDStr, 10, 32)
		if err == nil {
			query = query.Where("milestone_id = ?", uint(milestoneID))
		}
	}

	limitInt, err := strconv.Atoi(limit)
	if err == nil && limitInt > 0 {
		query = query.Limit(limitInt)
	}

	if err := query.Find(&trades).Error; err != nil {
		middleware.InternalServerError(c, err.Error())
		return
	}

	middleware.Success(c, trades, "내 거래 내역 조회 성공")
}

// GetMyPositions 내 포지션 조회
// GET /api/v1/positions/my
func (h *TradingHandler) GetMyPositions(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		middleware.Unauthorized(c, "User not authenticated")
		return
	}

	milestoneIDStr := c.Query("milestone_id")

	var positions []models.Position
	query := h.tradingService.GetDB().Where("user_id = ?", userID)

	if milestoneIDStr != "" {
		milestoneID, err := strconv.ParseUint(milestoneIDStr, 10, 32)
		if err == nil {
			query = query.Where("milestone_id = ?", uint(milestoneID))
		}
	}

	if err := query.Find(&positions).Error; err != nil {
		middleware.InternalServerError(c, err.Error())
		return
	}

	// 각 포지션의 미실현 손익 계산
	for i := range positions {
		position, err := h.tradingService.GetPosition(userID.(uint), positions[i].MilestoneID, positions[i].OptionID)
		if err == nil {
			positions[i] = *position
		}
	}

	middleware.Success(c, positions, "내 포지션 조회 성공")
}

// GetMilestonePosition 특정 마일스톤의 포지션 조회
// GET /api/v1/milestones/:id/position/:option
func (h *TradingHandler) GetMilestonePosition(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		middleware.Unauthorized(c, "User not authenticated")
		return
	}

	milestoneIDStr := c.Param("id")
	milestoneID, err := strconv.ParseUint(milestoneIDStr, 10, 32)
	if err != nil {
		middleware.BadRequest(c, "Invalid milestone ID")
		return
	}

	optionID := c.Param("option")
	if optionID == "" {
		middleware.BadRequest(c, "Option ID is required")
		return
	}

	position, err := h.tradingService.GetPosition(userID.(uint), uint(milestoneID), optionID)
	if err != nil {
		middleware.InternalServerError(c, err.Error())
		return
	}

	middleware.Success(c, position, "포지션 조회 성공")
}

// CancelOrder 주문 취소
// DELETE /api/v1/orders/:id
func (h *TradingHandler) CancelOrder(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		middleware.Unauthorized(c, "User not authenticated")
		return
	}

	orderIDStr := c.Param("id")
	orderID, err := strconv.ParseUint(orderIDStr, 10, 32)
	if err != nil {
		middleware.BadRequest(c, "Invalid order ID")
		return
	}

	// 주문 조회 및 권한 확인
	var order models.Order
	if err := h.tradingService.GetDB().Where("id = ? AND user_id = ?", uint(orderID), userID).First(&order).Error; err != nil {
		middleware.NotFound(c, "주문을 찾을 수 없습니다")
		return
	}

	// 취소 가능한 상태 확인
	if order.Status == models.OrderStatusFilled || order.Status == models.OrderStatusCancelled {
		middleware.BadRequest(c, "취소할 수 없는 주문입니다")
		return
	}

	// 주문 취소
	order.Status = models.OrderStatusCancelled
	if err := h.tradingService.GetDB().Save(&order).Error; err != nil {
		middleware.InternalServerError(c, "주문 취소 중 오류가 발생했습니다")
		return
	}

	middleware.Success(c, order, "주문이 성공적으로 취소되었습니다")
}

// GetRecentTrades 최근 거래 내역 조회 (공개)
// GET /api/v1/milestones/:id/trades/:option
func (h *TradingHandler) GetRecentTrades(c *gin.Context) {
	milestoneIDStr := c.Param("id")
	milestoneID, err := strconv.ParseUint(milestoneIDStr, 10, 32)
	if err != nil {
		middleware.BadRequest(c, "Invalid milestone ID")
		return
	}

	optionID := c.Param("option")
	if optionID == "" {
		middleware.BadRequest(c, "Option ID is required")
		return
	}

	limit := c.DefaultQuery("limit", "50")
	limitInt, err := strconv.Atoi(limit)
	if err != nil || limitInt <= 0 {
		limitInt = 50
	}

	// TradingService 메서드 사용
	trades, err := h.tradingService.GetRecentTrades(uint(milestoneID), optionID, limitInt)
	if err != nil {
		middleware.InternalServerError(c, err.Error())
		return
	}

	middleware.Success(c, gin.H{
		"trades": trades,
		"count":  len(trades),
	}, "최근 거래 조회 성공")
}

// GetUserWallet 사용자 지갑 조회
// GET /api/v1/wallet
func (h *TradingHandler) GetUserWallet(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)

	var wallet models.UserWallet
	err := h.tradingService.GetDB().Where("user_id = ?", userID).First(&wallet).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// 🆕 지갑이 없으면 큐로 비동기 생성 요청
			publisher := queue.NewPublisher()
			err := publisher.EnqueueWalletCreate(queue.WalletCreateEventData{
				UserID:        userID,
				InitialAmount: 10000,
			})
			if err != nil {
				log.Printf("❌ Failed to enqueue wallet creation: %v", err)
			}

			// 임시 응답 (프론트엔드에서 잠시 후 재시도 필요)
			middleware.Success(c, gin.H{
				"wallet_creating": true,
				"message": "지갑을 생성하고 있습니다. 잠시 후 다시 시도해주세요.",
				"retry_after": 3, // 3초 후 재시도 권장
			}, "지갑 생성 중")
			return
		}
		middleware.InternalServerError(c, "지갑 조회 실패")
		return
	}

	middleware.Success(c, wallet, "지갑 조회 성공")
}

// GetMilestoneMarket 마일스톤 마켓 정보 조회
// GET /api/v1/milestones/:id/market
func (h *TradingHandler) GetMilestoneMarket(c *gin.Context) {
	milestoneID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		middleware.BadRequest(c, "Invalid milestone ID")
		return
	}

	// 마켓 데이터 조회
	var marketData []models.MarketData
	if err := h.tradingService.GetDB().Where("milestone_id = ?", milestoneID).Find(&marketData).Error; err != nil {
		middleware.InternalServerError(c, "Failed to get market data")
		return
	}

	// 뷰어 수 조회 (SSE 서비스에서)
	viewerCount := 0
	totalClients := 0

	middleware.Success(c, gin.H{
		"market_data":   marketData,
		"viewer_count":  viewerCount,
		"total_clients": totalClients,
	}, "마켓 정보 조회 성공")
}

// InitializeMarket 마켓 초기화
// POST /api/v1/milestones/:id/market/init
func (h *TradingHandler) InitializeMarket(c *gin.Context) {
	milestoneID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		middleware.BadRequest(c, "Invalid milestone ID")
		return
	}

	var req struct {
		Options []string `json:"options"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.BadRequest(c, "Invalid request body")
		return
	}

	// 마일스톤 조회
	var milestone models.Milestone
	if err := h.tradingService.GetDB().First(&milestone, milestoneID).Error; err != nil {
		middleware.NotFound(c, "Milestone not found")
		return
	}

	// 옵션이 없으면 마일스톤의 betting_options 사용
	options := req.Options
	if len(options) == 0 {
		options = milestone.BettingOptions
	}

	// 마켓 초기화는 매칭 엔진에서 동적으로 처리됩니다
	// 첫 주문이 들어올 때 자동으로 마켓이 생성됩니다

	middleware.Success(c, gin.H{
		"message": "Market ready for trading",
		"options": options,
	}, "마켓 초기화 완료")
}

// HandleSSEConnection SSE 연결 처리
// GET /api/v1/milestones/:id/stream
func (h *TradingHandler) HandleSSEConnection(c *gin.Context) {
	milestoneID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		middleware.BadRequest(c, "Invalid milestone ID")
		return
	}

	log.Printf("🔗 SSE connection request for milestone %d from %s", milestoneID, c.ClientIP())

	// SSE 헤더 설정
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Headers", "Cache-Control")

	// 마일스톤 존재 확인
	var milestone models.Milestone
	if err := h.tradingService.GetDB().First(&milestone, milestoneID).Error; err != nil {
		log.Printf("❌ Milestone %d not found: %v", milestoneID, err)
		c.Data(200, "text/event-stream", []byte("data: {\"type\":\"error\",\"message\":\"Milestone not found\"}\n\n"))
		return
	}

	// 클라이언트가 연결을 종료했는지 확인하기 위한 채널
	clientGone := c.Writer.CloseNotify()

	log.Printf("✅ SSE connection established for milestone %d", milestoneID)

	// SSE 스트림 시작
	c.Stream(func(w io.Writer) bool {
		select {
		case <-clientGone:
			log.Printf("🔌 SSE client disconnected for milestone %d", milestoneID)
			return false
		default:
		}

		// 초기 연결 성공 메시지
		connectMsg := fmt.Sprintf("data: {\"type\":\"connection\",\"milestone_id\":%d,\"status\":\"connected\",\"timestamp\":%d}\n\n",
			milestoneID, time.Now().Unix())
		fmt.Fprint(w, connectMsg)

		// Keep-alive 메시지 (30초마다)
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-clientGone:
				log.Printf("🔌 SSE client disconnected for milestone %d", milestoneID)
				return false
			case <-ticker.C:
				// Keep-alive ping
				pingMsg := fmt.Sprintf("data: {\"type\":\"ping\",\"milestone_id\":%d,\"timestamp\":%d}\n\n",
					milestoneID, time.Now().Unix())
				if _, err := fmt.Fprint(w, pingMsg); err != nil {
					log.Printf("❌ SSE write error for milestone %d: %v", milestoneID, err)
					return false
				}
				log.Printf("📡 SSE ping sent for milestone %d", milestoneID)
			default:
				// 실제 거래/마켓 데이터는 Redis Pub/Sub으로 받아서 처리할 예정
				// 현재는 기본 연결 유지만 구현
				time.Sleep(1 * time.Second)
			}
		}
	})
}
