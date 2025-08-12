package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"blueprint-module/pkg/config"

	"github.com/redis/go-redis/v9"
)

var (
	Client *redis.Client
	ctx    = context.Background()
)

// InitRedis Redis 클라이언트 초기화
func InitRedis(cfg *config.Config) error {
	Client = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	// 연결 테스트
	pong, err := Client.Ping(ctx).Result()
	if err != nil {
		return fmt.Errorf("failed to connect to Redis: %v", err)
	}

	fmt.Printf("✅ Redis connected: %s\n", pong)
	return nil
}

// CloseRedis Redis 연결 종료
func CloseRedis() error {
	if Client != nil {
		return Client.Close()
	}
	return nil
}

// GetClient Redis 클라이언트 반환 (다른 패키지에서 사용)
func GetClient() *redis.Client {
	return Client
}

// 🔥 High-Performance Caching Functions

// SetOrderBook 호가창 데이터 캐싱 (초고속 조회용)
func SetOrderBook(milestoneID uint, optionID string, data interface{}) error {
	key := fmt.Sprintf("orderbook:%d:%s", milestoneID, optionID)
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	return Client.Set(ctx, key, jsonData, 30*time.Second).Err()
}

// GetOrderBook 호가창 데이터 조회
func GetOrderBook(milestoneID uint, optionID string, result interface{}) error {
	key := fmt.Sprintf("orderbook:%d:%s", milestoneID, optionID)
	val, err := Client.Get(ctx, key).Result()
	if err != nil {
		return err
	}

	return json.Unmarshal([]byte(val), result)
}

// SetMarketPrice 현재 시장 가격 캐싱
func SetMarketPrice(milestoneID uint, optionID string, price float64) error {
	key := fmt.Sprintf("price:%d:%s", milestoneID, optionID)
	return Client.Set(ctx, key, price, 10*time.Second).Err()
}

// GetMarketPrice 현재 시장 가격 조회
func GetMarketPrice(milestoneID uint, optionID string) (float64, error) {
	key := fmt.Sprintf("price:%d:%s", milestoneID, optionID)
	return Client.Get(ctx, key).Float64()
}

// SetRecentTrades 최근 거래 내역 캐싱
func SetRecentTrades(milestoneID uint, optionID string, trades interface{}) error {
	key := fmt.Sprintf("trades:%d:%s", milestoneID, optionID)
	jsonData, err := json.Marshal(trades)
	if err != nil {
		return err
	}

	return Client.Set(ctx, key, jsonData, 60*time.Second).Err()
}

// GetRecentTrades 최근 거래 내역 조회
func GetRecentTrades(milestoneID uint, optionID string, result interface{}) error {
	key := fmt.Sprintf("trades:%d:%s", milestoneID, optionID)
	val, err := Client.Get(ctx, key).Result()
	if err != nil {
		return err
	}

	return json.Unmarshal([]byte(val), result)
}

// 🚀 Real-time Broadcasting

// BroadcastRealtimeUpdate 실시간 업데이트 브로드캐스트 (기존 PublishRealtimeNotification)
func BroadcastRealtimeUpdate(channel string, event interface{}) error {
	jsonData, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return Client.Publish(ctx, channel, jsonData).Err()
}

// BroadcastTradeUpdate 거래 완료 실시간 브로드캐스트 (기존 PublishTradeNotification)
func BroadcastTradeUpdate(milestoneID uint, optionID string, event interface{}) error {
	channel := fmt.Sprintf("trade_events:%d:%s", milestoneID, optionID)
	return BroadcastRealtimeUpdate(channel, event)
}

// BroadcastPriceChange 가격 변동 실시간 브로드캐스트 (기존 PublishPriceUpdate)
func BroadcastPriceChange(milestoneID uint, optionID string, price float64) error {
	channel := fmt.Sprintf("price_updates:%d:%s", milestoneID, optionID)
	data := map[string]interface{}{
		"milestone_id": milestoneID,
		"option_id":    optionID,
		"price":        price,
		"timestamp":    time.Now().Unix(),
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	return Client.Publish(ctx, channel, jsonData).Err()
}

// 💾 Session Management

// SetUserSession 사용자 세션 저장
func SetUserSession(sessionID string, userID uint) error {
	key := fmt.Sprintf("session:%s", sessionID)
	return Client.Set(ctx, key, userID, 24*time.Hour).Err()
}

// GetUserSession 사용자 세션 조회
func GetUserSession(sessionID string) (uint, error) {
	key := fmt.Sprintf("session:%s", sessionID)
	val, err := Client.Get(ctx, key).Result()
	if err != nil {
		return 0, err
	}

	var userID uint
	if err := json.Unmarshal([]byte(val), &userID); err != nil {
		return 0, err
	}

	return userID, nil
}

// DeleteUserSession 사용자 세션 삭제
func DeleteUserSession(sessionID string) error {
	key := fmt.Sprintf("session:%s", sessionID)
	return Client.Del(ctx, key).Err()
}

// 🛡️ Rate Limiting

// CheckRateLimit API 요청 제한 확인
func CheckRateLimit(userID uint, endpoint string, maxRequests int, window time.Duration) (bool, error) {
	key := fmt.Sprintf("rate_limit:%d:%s", userID, endpoint)

	count, err := Client.Incr(ctx, key).Result()
	if err != nil {
		return false, err
	}

	if count == 1 {
		// 첫 요청시 TTL 설정
		Client.Expire(ctx, key, window)
	}

	return count <= int64(maxRequests), nil
}

// 📊 Analytics & Metrics

// IncrementMarketViews 시장 조회수 증가
func IncrementMarketViews(milestoneID uint) error {
	key := fmt.Sprintf("views:%d", milestoneID)
	return Client.Incr(ctx, key).Err()
}

// GetMarketViews 시장 조회수 조회
func GetMarketViews(milestoneID uint) (int64, error) {
	key := fmt.Sprintf("views:%d", milestoneID)
	return Client.Get(ctx, key).Int64()
}

// SetActiveUsers 현재 활성 사용자 수 설정
func SetActiveUsers(milestoneID uint, count int) error {
	key := fmt.Sprintf("active_users:%d", milestoneID)
	return Client.Set(ctx, key, count, 30*time.Second).Err()
}

// GetActiveUsers 현재 활성 사용자 수 조회
func GetActiveUsers(milestoneID uint) (int, error) {
	key := fmt.Sprintf("active_users:%d", milestoneID)
	return Client.Get(ctx, key).Int()
}

// 🧹 Utility Functions

// FlushMarketData 특정 시장의 모든 캐시 데이터 삭제
func FlushMarketData(milestoneID uint) error {
	pattern := fmt.Sprintf("*:%d:*", milestoneID)
	keys, err := Client.Keys(ctx, pattern).Result()
	if err != nil {
		return err
	}

	if len(keys) > 0 {
		return Client.Del(ctx, keys...).Err()
	}

	return nil
}

// HealthCheck Redis 상태 확인
func HealthCheck() error {
	_, err := Client.Ping(ctx).Result()
	return err
}
