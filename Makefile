# LifePathDAO Makefile

.PHONY: help build up down restart logs clean test

# 기본 타겟
help: ## 사용 가능한 명령어 목록 표시
	@echo "LifePathDAO Docker Commands:"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

build: ## Docker 이미지 빌드
	@echo "Building LifePathDAO Docker image..."
	docker-compose build --no-cache

up: ## 모든 서비스 시작 (백그라운드)
	@echo "Starting LifePathDAO services..."
	docker-compose up -d

up-logs: ## 모든 서비스 시작 (로그 표시)
	@echo "Starting LifePathDAO services with logs..."
	docker-compose up

down: ## 모든 서비스 중지 및 제거
	@echo "Stopping LifePathDAO services..."
	docker-compose down

restart: ## 모든 서비스 재시작
	@echo "Restarting LifePathDAO services..."
	docker-compose restart

logs: ## 모든 서비스 로그 표시
	docker-compose logs -f

logs-app: ## 애플리케이션 로그만 표시
	docker-compose logs -f app

logs-db: ## 데이터베이스 로그만 표시
	docker-compose logs -f postgres

status: ## 서비스 상태 확인
	docker-compose ps

clean: ## 모든 컨테이너, 볼륨, 네트워크 제거
	@echo "Cleaning up LifePathDAO Docker resources..."
	docker-compose down -v --remove-orphans
	docker system prune -f

clean-all: ## 모든 Docker 리소스 제거 (이미지 포함)
	@echo "Cleaning up all LifePathDAO Docker resources..."
	docker-compose down -v --remove-orphans --rmi all
	docker system prune -af

dev: ## 개발 모드로 시작 (rebuild + logs)
	@echo "Starting development environment..."
	docker-compose up --build

test: ## 애플리케이션 테스트 실행
	@echo "Running tests..."
	docker-compose exec app go test ./...

shell-app: ## 애플리케이션 컨테이너 셸 접속
	docker-compose exec app /bin/sh

shell-db: ## 데이터베이스 컨테이너 셸 접속
	docker-compose exec postgres psql -U postgres -d lifepath_dao

backup-db: ## 데이터베이스 백업
	@echo "Creating database backup..."
	docker-compose exec postgres pg_dump -U postgres lifepath_dao > backup_$(shell date +%Y%m%d_%H%M%S).sql

install: ## 첫 실행을 위한 전체 설정
	@echo "Setting up LifePathDAO for the first time..."
	@echo "1. Building images..."
	docker-compose build
	@echo "2. Starting services..."
	docker-compose up -d
	@echo "3. Waiting for services to be ready..."
	sleep 10
	@echo "4. Checking status..."
	docker-compose ps
	@echo ""
	@echo "🚀 LifePathDAO is now running!"
	@echo "📡 API Server: http://localhost:8080"
	@echo "🗄️  PostgreSQL: localhost:5432"
	@echo "🔴 Redis: localhost:6379"
	@echo ""
	@echo "Use 'make logs' to see the logs"
	@echo "Use 'make down' to stop all services"
