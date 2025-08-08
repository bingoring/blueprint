# 🚀 LifePathDAO Development Makefile

# 🔧 Environment Setup
.PHONY: setup
setup:
	@echo "🔧 Setting up development environment..."
	@if [ ! -f blueprint-be/.env ]; then \
		echo "📁 Creating .env file in blueprint-be/..."; \
		if [ -f blueprint-be/.env.example ]; then \
			cp blueprint-be/.env.example blueprint-be/.env; \
			echo "✅ Copied .env.example to .env"; \
		else \
			echo "# Blueprint Backend Environment Variables" > blueprint-be/.env; \
			echo "DB_HOST=localhost" >> blueprint-be/.env; \
			echo "DB_PORT=5432" >> blueprint-be/.env; \
			echo "DB_USER=postgres" >> blueprint-be/.env; \
			echo "DB_PASSWORD=password" >> blueprint-be/.env; \
			echo "DB_NAME=blueprint" >> blueprint-be/.env; \
			echo "JWT_SECRET=your-secret-key-here" >> blueprint-be/.env; \
			echo "REDIS_HOST=localhost" >> blueprint-be/.env; \
			echo "REDIS_PORT=6379" >> blueprint-be/.env; \
			echo "✅ Created default .env file"; \
		fi; \
		echo "⚠️  Please edit blueprint-be/.env file with your actual configuration values"; \
	else \
		echo "✅ .env file already exists in blueprint-be/"; \
	fi

# 🐳 Docker Commands
.PHONY: build up down logs clean install run-dev stop-backend backend-status run-backend run-frontend run-backend-with-env nuke-all nuke-db fresh-start build-backend build-frontend install-frontend

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
	docker-compose up -d postgres redis timescaledb
	@echo "✅ Development databases started!"
	@echo "🗄️  PostgreSQL: localhost:5432"
	@echo "🔴 Redis: localhost:6379"
	@echo ""
	@echo "이제 백엔드와 프론트엔드를 로컬에서 실행하세요:"
	@echo "  Backend:  make run-backend"
	@echo "  Frontend: make run-frontend"

dev-db-down: ## 개발 데이터베이스 중지
	docker-compose down

# 🚀 Development Commands (Local)
# Start only database and Redis for local development
run-dev:
	@echo "🗄️  Starting database and Redis only..."
	docker-compose -f docker-compose.dev.yml up -d

# Stop backend processes
stop-backend:
	@echo "🛑 Stopping backend processes..."
	@pkill -f "go run blueprint-be/cmd/server/main.go" 2>/dev/null && echo "✅ Go backend stopped" || echo "ℹ️  No Go backend running"
	@pkill -f "./blueprint-be/server" 2>/dev/null && echo "✅ Binary backend stopped" || echo "ℹ️  No binary backend running"

# Check backend process status
backend-status:
	@echo "📊 Backend Process Status:"
	@echo ""
	@echo "🔍 Go processes:"
	@pgrep -fl "go run blueprint-be/cmd/server/main.go" || echo "   No Go backend running"
	@echo ""
	@echo "🔍 Binary processes:"
	@pgrep -fl "./blueprint-be/server" || echo "   No binary backend running"
	@echo ""
	@echo "🔍 Port 8080 usage:"
	@lsof -i :8080 || echo "   Port 8080 is free"

# Run backend locally (requires Go)
run-backend:
	@echo "🔄 Checking for existing backend processes..."
	@pkill -f "go run blueprint-be/cmd/server/main.go" 2>/dev/null || true
	@lsof -ti:8080 | xargs kill -9 || true
	@pkill -f "./blueprint-be/server" 2>/dev/null || true
	@sleep 1
	@echo "🔙 Starting backend server locally..."
	@if [ -f blueprint-be/.env ]; then \
		echo "📁 Loading environment from .env file..."; \
		cd blueprint-be && set -a && . ./.env && set +a && go run cmd/server/main.go; \
	else \
		echo "❌ .env file not found in blueprint-be/. Run 'make setup' first."; \
		exit 1; \
	fi

# Run backend with explicit environment loading (alternative method)
run-backend-with-env:
	@echo "🔄 Checking for existing backend processes..."
	@pkill -f "go run blueprint-be/cmd/server/main.go" 2>/dev/null || true
	@pkill -f "./blueprint-be/server" 2>/dev/null || true
	@sleep 1
	@echo "🔙 Starting backend server with environment..."
	@if [ -f blueprint-be/.env ]; then \
		echo "📁 Loading .env and starting server..."; \
		cd blueprint-be && env $$(cat .env | grep -v '^#' | xargs) go run cmd/server/main.go; \
	else \
		echo "❌ .env file not found in blueprint-be/. Run 'make setup' first."; \
		exit 1; \
	fi

# Run frontend locally (requires Node.js)
run-frontend:
	@echo "🎨 Starting frontend server locally..."
	cd blueprint-fe && npm run dev

# Install frontend dependencies
install-frontend:
	@echo "📦 Installing frontend dependencies..."
	cd blueprint-fe && npm install
	@echo "✅ Frontend dependencies installed!"

# Build backend binary
build-backend:
	@echo "🏗️  Building backend binary..."
	cd blueprint-be && go build -o server cmd/server/main.go
	@echo "✅ Backend binary built!"

# Build frontend for production
build-frontend:
	@echo "🏗️  Building frontend for production..."
	cd blueprint-fe && npm run build
	@echo "✅ Frontend built!"

# 🔍 Utility Commands
.PHONY: status db-logs redis-logs redis-cli redis-info timescale-shell timescale-logs timescale-status timescale-tables timescale-info db-shell db-connect db-admin db-reset db-backup db-import db-info

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

# Connect to Redis CLI
redis-cli:
	@echo "🔴 Connecting to Redis CLI..."
	@echo "💡 Tip: Use 'info' to see server info, 'keys *' to list all keys, 'quit' to exit"
	@echo ""
	docker exec -it blueprint-redis redis-cli

# Show Redis server info
redis-info:
	@echo "🔴 Redis server information:"
	docker exec blueprint-redis-dev redis-cli info server

# 📊 TimescaleDB Commands
# Connect to TimescaleDB
timescale-shell:
	@echo "📊 Connecting to TimescaleDB..."
	@echo "📋 Database: timeseries | User: postgres | Container: blueprint-timescaledb"
	@echo "💡 Tip: Use \dt to list tables, \q to quit"
	@echo ""
	docker exec -it blueprint-timescaledb psql -U postgres -d timeseries

# Show TimescaleDB logs
timescale-logs:
	@echo "📊 TimescaleDB logs:"
	docker-compose logs -f timescaledb

# Show TimescaleDB status
timescale-status:
	@echo "📊 TimescaleDB container status:"
	docker ps --filter "name=blueprint-timescaledb" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"

# Query TimescaleDB hypertables
timescale-tables:
	@echo "📊 TimescaleDB hypertables:"
	docker exec blueprint-timescaledb psql -U postgres -d timeseries -c "SELECT hypertable_name, owner, num_dimensions FROM timescaledb_information.hypertables;"

# Check TimescaleDB extension
timescale-info:
	@echo "📊 TimescaleDB extension info:"
	docker exec blueprint-timescaledb psql -U postgres -d timeseries -c "SELECT * FROM pg_extension WHERE extname = 'timescaledb';"

# 🐘 PostgreSQL Commands
# Connect to PostgreSQL database
db-shell:
	@echo "🐘 Connecting to PostgreSQL database..."
	@echo "📋 Database: blueprint | User: postgres | Container: blueprint-postgres-dev"
	@echo "💡 Tip: Use \dt to list tables, \q to quit"
	@echo ""
	docker exec -it blueprint-postgres psql -U postgres -d blueprint

# Alternative connection command (shorter alias)
db-connect: db-shell

# Connect as root to postgres (for admin tasks)
db-admin:
	@echo "🔧 Connecting to PostgreSQL as admin..."
	docker exec -it blueprint-postgres-dev psql -U postgres

# Reset database (drop and recreate)
db-reset:
	@echo "⚠️  Are you sure you want to reset the database? This will DELETE ALL DATA!"
	@echo "Press Ctrl+C to cancel, or Enter to continue..."
	@read confirm
	@echo "🗑️  Resetting database..."
	docker exec -it blueprint-postgres-dev psql -U postgres -c "DROP DATABASE IF EXISTS blueprint;"
	docker exec -it blueprint-postgres-dev psql -U postgres -c "CREATE DATABASE blueprint;"
	@echo "✅ Database reset complete!"

# Import init.sql to database
db-import:
	@echo "📥 Importing init.sql to database..."
	docker exec -i blueprint-postgres psql -U postgres -d blueprint < blueprint-be/init.sql
	@echo "✅ Database import complete!"

# Create database backup
db-backup:
	@echo "💾 Creating database backup..."
	@mkdir -p backups
	docker exec blueprint-postgres-dev pg_dump -U postgres blueprint > backups/backup_$$(date +%Y%m%d_%H%M%S).sql
	@echo "✅ Backup created in backups/ directory!"

# 💀 NUCLEAR OPTIONS - Complete Data Destruction 💀

# Nuke all data (PostgreSQL + Redis + all volumes)
nuke-all:
	@echo "💀 ⚠️  NUCLEAR OPTION: This will COMPLETELY DESTROY ALL DATA ⚠️  💀"
	@echo "   - All PostgreSQL databases"
	@echo "   - All Redis data"
	@echo "   - All Docker volumes"
	@echo "   - All containers"
	@echo ""
	@echo "🚨 THIS CANNOT BE UNDONE! 🚨"
	@echo ""
	@echo "Type 'YES I WANT TO DESTROY EVERYTHING' to continue:"
	@read confirm && [ "$$confirm" = "YES I WANT TO DESTROY EVERYTHING" ] || (echo "❌ Cancelled." && exit 1)
	@echo ""
	@echo "💥 Destroying everything in 3 seconds..."
	@sleep 1 && echo "💥 3..."
	@sleep 1 && echo "💥 2..."
	@sleep 1 && echo "💥 1..."
	@echo "💀 NUKING ALL DATA..."
	docker-compose down -v --remove-orphans
	docker system prune -f --volumes
	docker volume prune -f
	@echo "💀 ✅ Everything has been destroyed!"
	@echo "🔄 Run 'make fresh-start' to rebuild from scratch"

# Nuke only database data (PostgreSQL + Redis, keep other containers)
nuke-db:
	@echo "💀 ⚠️  NUCLEAR DB OPTION: This will DESTROY ALL DATABASE DATA ⚠️  💀"
	@echo "   - All PostgreSQL databases and volumes"
	@echo "   - All Redis data and volumes"
	@echo "   - TimescaleDB data"
	@echo ""
	@echo "🚨 THIS CANNOT BE UNDONE! 🚨"
	@echo ""
	@echo "Type 'NUKE DATABASE' to continue:"
	@read confirm && [ "$$confirm" = "NUKE DATABASE" ] || (echo "❌ Cancelled." && exit 1)
	@echo ""
	@echo "💥 Destroying database data in 3 seconds..."
	@sleep 1 && echo "💥 3..."
	@sleep 1 && echo "💥 2..."
	@sleep 1 && echo "💥 1..."
	@echo "💀 NUKING DATABASE DATA..."
	docker-compose down -v
	docker volume rm blueprint_postgres_data blueprint_redis_data blueprint_timescale_data 2>/dev/null || true
	@echo "💀 ✅ All database data has been destroyed!"
	@echo "🔄 Run 'make dev-db' to restart clean databases"

# Fresh start - nuke everything and rebuild
fresh-start:
	@echo "🔄 FRESH START: Complete rebuild from scratch"
	@echo ""
	@$(MAKE) nuke-all
	@echo ""
	@echo "🏗️  Rebuilding everything..."
	@$(MAKE) build
	@echo ""
	@echo "🚀 Starting fresh development environment..."
	@$(MAKE) dev-db
	@echo ""
	@echo "✅ Fresh start complete! 🎉"
	@echo ""
	@echo "Next steps:"
	@echo "  1. make run-backend"
	@echo "  2. make run-frontend"

# Show database info
db-info:
	@echo "🐘 PostgreSQL Database Information:"
	@echo "📋 Database: blueprint"
	@echo "👤 User: postgres"
	@echo "🔗 Host: localhost:5432"
	@echo "🐳 Container: blueprint-postgres-dev"
	@echo ""
	@echo "📊 Database Size:"
	@docker exec blueprint-postgres-dev psql -U postgres -d blueprint -c "SELECT pg_size_pretty(pg_database_size('blueprint')) as size;"

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
	@echo "  make dev-db         - Start only database and Redis"
	@echo "  make run-backend    - Run backend server locally"
	@echo "  make run-frontend   - Run frontend server locally"
	@echo "  make stop-backend   - Stop all backend processes"
	@echo "  make backend-status - Check backend process status"
	@echo ""
	@echo "📦 Build & Install:"
	@echo "  make install-frontend - Install frontend dependencies"
	@echo "  make build-backend    - Build backend binary"
	@echo "  make build-frontend   - Build frontend for production"
	@echo ""
	@echo "🔍 Monitoring:"
	@echo "  make status         - Show container status"
	@echo "  make db-logs        - Show database logs"
	@echo "  make redis-logs     - Show Redis logs"
	@echo "  make redis-cli      - Connect to Redis CLI"
	@echo "  make redis-info     - Show Redis server info"
	@echo ""
	@echo "📊 TimescaleDB:"
	@echo "  make timescale-shell   - Connect to TimescaleDB shell"
	@echo "  make timescale-logs    - Show TimescaleDB logs"
	@echo "  make timescale-status  - Show TimescaleDB status"
	@echo "  make timescale-tables  - Show hypertables"
	@echo "  make timescale-info    - Show TimescaleDB extension info"
	@echo ""
	@echo "🐘 PostgreSQL:"
	@echo "  make db-shell       - Connect to PostgreSQL database"
	@echo "  make db-connect     - Same as db-shell (shorter alias)"
	@echo "  make db-admin       - Connect as PostgreSQL admin"
	@echo "  make db-info        - Show database information"
	@echo "  make db-reset       - Reset database (⚠️ DELETES ALL DATA)"
	@echo "  make db-backup      - Create database backup"
	@echo "  make db-import      - Import init.sql file"
	@echo ""
	@echo "💀 NUCLEAR OPTIONS (⚠️ DANGER ZONE ⚠️):"
	@echo "  make nuke-db        - 💀 DESTROY all database data (PostgreSQL + Redis)"
	@echo "  make nuke-all       - 💀 DESTROY EVERYTHING (all containers + volumes)"
	@echo "  make fresh-start    - 🔄 Complete reset + rebuild from scratch"
	@echo ""
	@echo "🆘 Example workflow:"
	@echo "  1. make setup            # Setup environment"
	@echo "  2. make install-frontend # Install frontend dependencies"
	@echo "  3. make dev-db           # Start database"
	@echo "  4. make run-backend      # Start backend in another terminal"
	@echo "  5. make run-frontend     # Start frontend in another terminal"
