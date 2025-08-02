# 🚀 LifePathDAO Development Makefile

# 🔧 Environment Setup
.PHONY: setup
setup:
	@echo "🔧 Setting up development environment..."
	@if [ ! -f .env ]; then \
		echo "📁 Copying .env.example to .env..."; \
		cp .env.example .env; \
		echo "⚠️  Please edit .env file with your actual configuration values"; \
	else \
		echo "✅ .env file already exists"; \
	fi

# 🐳 Docker Commands
.PHONY: build up down logs clean install run-dev run-backend run-frontend run-backend-with-env

# Build all containers
build:
	@echo "🏗️  Building Docker containers..."
	docker-compose build

# Start all services (with database)
up:
	@echo "🚀 Starting all services..."
	docker-compose up -d

# Stop all services
down:
	@echo "🛑 Stopping all services..."
	docker-compose down

# Show logs for all services
logs:
	@echo "📋 Showing logs..."
	docker-compose logs -f

# Clean up everything (containers, volumes, images)
clean:
	@echo "🧹 Cleaning up Docker resources..."
	docker-compose down -v --remove-orphans
	docker system prune -f

# Build and start all services
install: build up

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

# 🚀 Development Commands (Local)
# Start only database and Redis for local development
run-dev:
	@echo "🗄️  Starting database and Redis only..."
	docker-compose -f docker-compose.dev.yml up -d

# Run backend locally (requires Go)
run-backend:
	@echo "🔙 Starting backend server locally..."
	@if [ -f .env ]; then \
		echo "📁 Loading environment from .env file..."; \
		set -a && . ./.env && set +a && go run cmd/server/main.go; \
	else \
		echo "❌ .env file not found. Run 'make setup' first."; \
		exit 1; \
	fi

# Run backend with explicit environment loading (alternative method)
run-backend-with-env:
	@echo "🔙 Starting backend server with environment..."
	@if [ -f .env ]; then \
		echo "📁 Loading .env and starting server..."; \
		env $$(cat .env | grep -v '^#' | xargs) go run cmd/server/main.go; \
	else \
		echo "❌ .env file not found. Run 'make setup' first."; \
		exit 1; \
	fi

# Run frontend locally (requires Node.js)
run-frontend:
	@echo "🎨 Starting frontend server locally..."
	cd web && npm run dev

# 🔍 Utility Commands
.PHONY: status db-logs redis-logs

# Show status of all containers
status:
	@echo "📊 Container status:"
	docker-compose ps

# Show database logs
db-logs:
	@echo "🗄️  Database logs:"
	docker-compose logs -f postgres

# Show Redis logs
redis-logs:
	@echo "🔴 Redis logs:"
	docker-compose logs -f redis

# 📋 Help
.PHONY: help
help:
	@echo "🚀 LifePathDAO Development Commands:"
	@echo ""
	@echo "🔧 Setup:"
	@echo "  make setup           - Initialize .env file from template"
	@echo ""
	@echo "🐳 Docker (Full Stack):"
	@echo "  make build          - Build all Docker containers"
	@echo "  make install        - Build and start all services"
	@echo "  make up             - Start all services"
	@echo "  make down           - Stop all services"
	@echo "  make logs           - Show logs for all services"
	@echo "  make clean          - Clean up Docker resources"
	@echo ""
	@echo "🚀 Local Development:"
	@echo "  make run-dev        - Start only database and Redis"
	@echo "  make run-backend    - Run backend server locally"
	@echo "  make run-frontend   - Run frontend server locally"
	@echo ""
	@echo "🔍 Monitoring:"
	@echo "  make status         - Show container status"
	@echo "  make db-logs        - Show database logs"
	@echo "  make redis-logs     - Show Redis logs"
	@echo ""
	@echo "🆘 Example workflow:"
	@echo "  1. make setup       # Setup environment"
	@echo "  2. make run-dev     # Start database"
	@echo "  3. make run-backend # Start backend in another terminal"
	@echo "  4. make run-frontend # Start frontend in another terminal"
