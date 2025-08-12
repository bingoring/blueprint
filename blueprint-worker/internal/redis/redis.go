package redis

import (
	moduleConfig "blueprint-module/pkg/config"
	moduleRedis "blueprint-module/pkg/redis"
	"blueprint-worker/internal/config"
)

// InitRedis 워커용 Redis 초기화 (blueprint-module 재사용)
func InitRedis(cfg *config.Config) error {
	// 워커 설정을 모듈 설정으로 변환
	moduleCfg := &moduleConfig.Config{
		Redis: moduleConfig.RedisConfig{
			Host:     cfg.Redis.Host,
			Port:     cfg.Redis.Port,
			Password: cfg.Redis.Password,
			DB:       cfg.Redis.DB,
		},
	}

	return moduleRedis.InitRedis(moduleCfg)
}

// CloseRedis Redis 연결 종료
func CloseRedis() error {
	return moduleRedis.CloseRedis()
}

// GetClient Redis 클라이언트 반환
func GetClient() interface{} {
	return moduleRedis.GetClient()
}
