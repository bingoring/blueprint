# Blueprint Makefile

.PHONY: help build up down restart logs clean test

# 기본 타겟
help: ## 사용 가능한 명령어 목록 표시
	@echo "Blueprint Docker Commands:"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

build: ## Docker 이미지 빌드
	@echo "Building Blueprint Docker image..."
	docker-compose build --no-cache

up: ## 모든 서비스 시작 (백그라운드)
	@echo "Starting Blueprint services..."
	docker-compose up -d

up-logs: ## 모든 서비스 시작 (로그 표시)
	@echo "Starting Blueprint services with logs..."
	docker-compose up

down: ## 모든 서비스 중지 및 제거
	@echo "Stopping Blueprint services..."
	docker-compose down

restart: ## 모든 서비스 재시작
	@echo "Restarting Blueprint services..."
	docker-compose restart

logs: ## 모든 서비스 로그 표시
	docker-compose logs -f

logs-app: ## 백엔드 애플리케이션 로그만 표시
	docker-compose logs -f app

logs-web: ## 프론트엔드 애플리케이션 로그만 표시
	docker-compose logs -f web

logs-db: ## 데이터베이스 로그만 표시
	docker-compose logs -f postgres

status: ## 서비스 상태 확인
	docker-compose ps

clean: ## 모든 컨테이너, 볼륨, 네트워크 제거
	@echo "Cleaning up Blueprint Docker resources..."
	docker-compose down -v --remove-orphans
	docker system prune -f

clean-all: ## 모든 Docker 리소스 제거 (이미지 포함)
	@echo "Cleaning up all Blueprint Docker resources..."
	docker-compose down -v --remove-orphans --rmi all
	docker system prune -af

dev: ## 개발 모드로 시작 (rebuild + logs)
	@echo "Starting development environment..."
	docker-compose up --build

dev-db: ## 데이터베이스만 시작 (로컬 개발용)
	@echo "Starting development databases..."
	docker-compose -f docker-compose.dev.yml up -d
	@echo "✅ Development databases started!"
	@echo "🗄️  PostgreSQL: localhost:5432"
	@echo "🔴 Redis: localhost:6379"
	@echo ""
	@echo "이제 백엔드와 프론트엔드를 로컬에서 실행하세요:"
	@echo "  Backend:  make run-backend"
	@echo "  Frontend: make run-frontend"

dev-db-down: ## 개발 데이터베이스 중지
	docker-compose -f docker-compose.dev.yml down

run-backend: ## 로컬에서 백엔드 실행
	@echo "🚀 Starting backend locally..."
	export DB_HOST=localhost && \
	export DB_PORT=5432 && \
	export DB_USER=postgres && \
	export DB_PASSWORD=password && \
	export DB_NAME=blueprint_db && \
	export DB_SSLMODE=disable && \
	export JWT_SECRET=your-super-secret-jwt-key && \
	export GOOGLE_PROJECT_ID=blueprint-467515 && \
	export GOOGLE_CLIENT_ID=$${GOOGLE_CLIENT_ID:-your-google-client-id} && \
	export GOOGLE_CLIENT_SECRET=$${GOOGLE_CLIENT_SECRET:-your-google-client-secret} && \
	export GOOGLE_REDIRECT_URL=http://localhost:8080/api/v1/auth/google/callback && \
	export PORT=8080 && \
	export GIN_MODE=debug && \
	go run cmd/server/main.go

setup-env: ## 환경변수 설정 (현재 셸에 적용)
	@echo "🔧 환경변수 설정 스크립트 실행..."
	@echo "💡 사용법: source scripts/setup-env.sh"
	@chmod +x scripts/setup-env.sh

run-backend-with-env: setup-env ## 환경변수 설정 후 백엔드 실행
	@echo "🚀 환경변수 설정 후 백엔드 시작..."
	@source scripts/setup-env.sh && go run cmd/server/main.go

run-frontend: ## 로컬에서 프론트엔드 실행
	@echo "🌐 Starting frontend locally..."
	cd web && npm run dev

test: ## 애플리케이션 테스트 실행
	@echo "Running tests..."
	docker-compose exec app go test ./...

shell-app: ## 백엔드 애플리케이션 컨테이너 셸 접속
	docker-compose exec app /bin/sh

shell-web: ## 프론트엔드 애플리케이션 컨테이너 셸 접속
	docker-compose exec web /bin/sh

shell-db: ## 데이터베이스 컨테이너 셸 접속
	docker-compose exec postgres psql -U postgres -d blueprint_db

backup-db: ## 데이터베이스 백업
	@echo "Creating database backup..."
	docker-compose exec postgres pg_dump -U postgres blueprint_db > backup_$(shell date +%Y%m%d_%H%M%S).sql

install: ## 첫 실행을 위한 전체 설정
	@echo "Setting up Blueprint for the first time..."
	@echo "1. Building images..."
	docker-compose build
	@echo "2. Starting services..."
	docker-compose up -d
	@echo "3. Waiting for services to be ready..."
	sleep 10
	@echo "4. Checking status..."
	docker-compose ps
	@echo ""
	@echo "🚀 Blueprint is now running!"
	@echo "🌐 Frontend: http://localhost:3000"
	@echo "📡 API Server: http://localhost:8080"
	@echo "🗄️  PostgreSQL: localhost:5432"
	@echo "🔴 Redis: localhost:6379"
	@echo ""
	@echo "Use 'make logs' to see the logs"
	@echo "Use 'make down' to stop all services"
