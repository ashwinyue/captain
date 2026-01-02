.PHONY: help up down logs build clean dev infra ps

# Default target
help:
	@echo "Captain Docker Commands:"
	@echo ""
	@echo "  make up        - Start all services"
	@echo "  make down      - Stop all services"
	@echo "  make logs      - View logs (follow mode)"
	@echo "  make build     - Build all services"
	@echo "  make clean     - Stop and remove all containers and volumes"
	@echo "  make infra     - Start only infrastructure (postgres, redis, wukongim)"
	@echo "  make ps        - Show running containers"
	@echo ""

# Start all services
up:
	docker-compose up -d

# Stop all services
down:
	docker-compose down

# View logs
logs:
	docker-compose logs -f

# Build all services
build:
	docker-compose build

# Clean everything
clean:
	docker-compose down -v --remove-orphans

# Start only infrastructure
infra:
	docker-compose up -d postgres redis wukongim adminer redis-commander

# Show running containers
ps:
	docker-compose ps

# Restart a specific service
restart-%:
	docker-compose restart $*

# View logs for a specific service
logs-%:
	docker-compose logs -f $*

# Build a specific service
build-%:
	docker-compose build $*

# Health check all services
health:
	@echo "Checking API Server..."
	@curl -s http://localhost:8000/health || echo "API Server not responding"
	@echo ""
	@echo "Checking AI Center..."
	@curl -s http://localhost:8081/health || echo "AI Center not responding"
	@echo ""
	@echo "Checking RAG Service..."
	@curl -s http://localhost:8082/health || echo "RAG Service not responding"
	@echo ""
	@echo "Checking Platform Service..."
	@curl -s http://localhost:8083/health || echo "Platform Service not responding"
	@echo ""
	@echo "Checking WuKongIM..."
	@curl -s http://localhost:5001/health || echo "WuKongIM not responding"

# Initialize database
init-db:
	docker-compose exec postgres psql -U captain -d captain -f /docker-entrypoint-initdb.d/01-init.sql

# Connect to database
db:
	docker-compose exec postgres psql -U captain -d captain

# Connect to redis
redis-cli:
	docker-compose exec redis redis-cli
