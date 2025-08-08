package services

import (
	"blueprint/internal/models"
	"fmt"
	"math"
)

// 🎯 Polymarket-style Probability Validation

// ProbabilityValidator 확률 검증 서비스
type ProbabilityValidator struct{}

// NewProbabilityValidator 생성자
func NewProbabilityValidator() *ProbabilityValidator {
	return &ProbabilityValidator{}
}

// ValidateProbabilitySum 확률 합계 검증 (폴리마켓 스타일)
// 모든 옵션의 확률 합이 100% (1.0)가 되어야 함
func (pv *ProbabilityValidator) ValidateProbabilitySum(prices []float64) error {
	if len(prices) == 0 {
		return fmt.Errorf("no prices provided")
	}

	sum := 0.0
	for _, price := range prices {
		if price < 0.01 || price > 0.99 {
			return fmt.Errorf("price %.4f is out of valid range (0.01-0.99)", price)
		}
		sum += price
	}

	// 부동소수점 오차 허용 (±0.01)
	tolerance := 0.01
	if math.Abs(sum-1.0) > tolerance {
		return fmt.Errorf("probability sum %.4f must equal 1.0 (±%.2f)", sum, tolerance)
	}

	return nil
}

// ValidateBinaryMarket 이진 마켓 검증 (success/fail)
func (pv *ProbabilityValidator) ValidateBinaryMarket(successPrice, failPrice float64) error {
	return pv.ValidateProbabilitySum([]float64{successPrice, failPrice})
}

// ValidateMarketPrices 마켓 가격 검증 (다중 옵션)
func (pv *ProbabilityValidator) ValidateMarketPrices(milestoneID uint, optionPrices map[string]float64) error {
	if len(optionPrices) < 2 {
		return fmt.Errorf("market must have at least 2 options")
	}

	var prices []float64
	for option, price := range optionPrices {
		if option == "" {
			return fmt.Errorf("option name cannot be empty")
		}
		prices = append(prices, price)
	}

	return pv.ValidateProbabilitySum(prices)
}

// CalculateImpliedProbability 주문장 기반 내재 확률 계산
func (pv *ProbabilityValidator) CalculateImpliedProbability(orderBook *models.OrderBook) (float64, error) {
	if orderBook == nil || len(orderBook.Bids) == 0 || len(orderBook.Asks) == 0 {
		return 0.5, nil // 유동성이 없으면 50% 기본값
	}

	// 최고 매수가와 최저 매도가의 중간값
	bestBid := orderBook.Bids[0].Price
	bestAsk := orderBook.Asks[0].Price

	if bestBid <= 0 || bestAsk <= 0 {
		return 0.5, nil
	}

	// 스프레드가 클 경우 중간값 사용
	if bestAsk > bestBid {
		midPrice := (bestBid + bestAsk) / 2.0
		return midPrice, nil
	}

	// 크로스된 경우 (비정상) - 보수적으로 처리
	return bestBid, nil
}

// RebalanceMarketPrices 마켓 가격 재조정 (확률 합 = 1.0)
func (pv *ProbabilityValidator) RebalanceMarketPrices(prices map[string]float64) (map[string]float64, error) {
	if len(prices) == 0 {
		return prices, fmt.Errorf("no prices to rebalance")
	}

	// 현재 합계 계산
	sum := 0.0
	for _, price := range prices {
		sum += price
	}

	if sum <= 0 {
		return prices, fmt.Errorf("total price sum must be positive")
	}

	// 비례적으로 조정
	rebalanced := make(map[string]float64)
	for option, price := range prices {
		adjusted := price / sum

		// 최소/최대 값 제한
		if adjusted < 0.01 {
			adjusted = 0.01
		} else if adjusted > 0.99 {
			adjusted = 0.99
		}

		rebalanced[option] = adjusted
	}

	// 재검증
	if err := pv.ValidateMarketPrices(0, rebalanced); err != nil {
		return prices, fmt.Errorf("rebalancing failed: %v", err)
	}

	return rebalanced, nil
}

// ValidateOrderPrice 주문 가격 검증
func (pv *ProbabilityValidator) ValidateOrderPrice(price float64, orderType models.OrderType) error {
	if price < 0.01 || price > 0.99 {
		return fmt.Errorf("order price %.4f must be between 0.01 and 0.99", price)
	}

	// 시장가 주문은 추가 검증 없음
	if orderType == models.OrderTypeMarket {
		return nil
	}

	// 지정가 주문은 합리적인 가격 범위 검증
	if price <= 0.05 || price >= 0.95 {
		return fmt.Errorf("limit order price %.4f is in extreme range, please confirm", price)
	}

	return nil
}

// CalculateArbitrageOpportunity 차익거래 기회 분석
func (pv *ProbabilityValidator) CalculateArbitrageOpportunity(marketPrices map[string]float64) *ArbitrageOpportunity {
	if len(marketPrices) < 2 {
		return nil
	}

	sum := 0.0
	for _, price := range marketPrices {
		sum += price
	}

	// 확률 합이 1.0이 아니면 차익거래 기회
	arbitrageValue := math.Abs(sum - 1.0)

	if arbitrageValue < 0.01 {
		return nil // 차익거래 기회 없음
	}

	return &ArbitrageOpportunity{
		Type:           determineArbitrageType(sum),
		Value:          arbitrageValue,
		ProbabilitySum: sum,
		Explanation:    generateArbitrageExplanation(sum),
		Severity:       calculateArbitrageSeverity(arbitrageValue),
	}
}

// ArbitrageOpportunity 차익거래 기회
type ArbitrageOpportunity struct {
	Type           string  `json:"type"`            // "underpriced", "overpriced"
	Value          float64 `json:"value"`           // 차익거래 가치
	ProbabilitySum float64 `json:"probability_sum"` // 확률 합계
	Explanation    string  `json:"explanation"`     // 설명
	Severity       string  `json:"severity"`        // "low", "medium", "high"
}

// Helper functions

func determineArbitrageType(sum float64) string {
	if sum < 1.0 {
		return "underpriced" // 전체적으로 저평가됨
	}
	return "overpriced" // 전체적으로 고평가됨
}

func generateArbitrageExplanation(sum float64) string {
	if sum < 1.0 {
		return fmt.Sprintf("Market is underpriced (%.2f%% total). Consider buying all options.", sum*100)
	}
	return fmt.Sprintf("Market is overpriced (%.2f%% total). Consider selling all options.", sum*100)
}

func calculateArbitrageSeverity(value float64) string {
	if value < 0.05 {
		return "low"
	} else if value < 0.15 {
		return "medium"
	}
	return "high"
}
