.PHONY: all setup run-server run-client run-all db-start db-create db-schema clean help

# Colors for console output
GREEN=\033[0;32m
YELLOW=\033[0;33m
RED=\033[0;31m
NC=\033[0m # No Color

# Default target
all: help

# Database settings
DB_NAME=gotext
DB_USER=$(shell whoami)

# Help message
help:
	@echo "${YELLOW}GoText Application Makefile${NC}"
	@echo ""
	@echo "Available commands:"
	@echo "${GREEN}make setup${NC}      - Initialize the development environment"
	@echo "${GREEN}make run-all${NC}    - Run both server and client in parallel"
	@echo "${GREEN}make run-server${NC} - Run the Go server only"
	@echo "${GREEN}make run-client${NC} - Run the React client only"
	@echo "${GREEN}make db-start${NC}   - Start the PostgreSQL database"
	@echo "${GREEN}make db-create${NC}  - Create the database"
	@echo "${GREEN}make db-schema${NC}  - Apply database schema"
	@echo "${GREEN}make clean${NC}      - Clean up build artifacts"

# Setup the development environment
setup: db-start db-create db-schema
	@echo "${GREEN}Installing Go dependencies...${NC}"
	@cd server && go mod tidy
	@echo "${GREEN}Installing frontend dependencies...${NC}"
	@cd client && npm install
	@echo "${GREEN}Setup complete!${NC}"

# Run the server
run-server:
	@echo "${GREEN}Starting Go server...${NC}"
	@cd server && go run cmd/server/main.go

# Run the client
run-client:
	@echo "${GREEN}Starting React client...${NC}"
	@cd client && npm start

# Run both server and client
run-all:
	@echo "${GREEN}Starting server and client...${NC}"
	@$(MAKE) -j2 run-server run-client

# Start PostgreSQL database
db-start:
	@echo "${GREEN}Starting PostgreSQL...${NC}"
	@if pgrep -x postgres > /dev/null; then \
		echo "PostgreSQL is already running"; \
	else \
		brew services start postgresql || echo "${RED}Failed to start PostgreSQL. Is it installed?${NC}"; \
	fi

# Create database
db-create:
	@echo "${GREEN}Creating database...${NC}"
	@createdb $(DB_NAME) 2>/dev/null || echo "${YELLOW}Database $(DB_NAME) already exists.${NC}"

# Apply database schema
db-schema:
	@echo "${GREEN}Applying database schema...${NC}"
	@psql -d $(DB_NAME) -f server/internal/db/schema.sql

# Clean up
clean:
	@echo "${GREEN}Cleaning up...${NC}"
	@rm -rf server/bin client/build
	@echo "${GREEN}Cleanup complete!${NC}" 