#!/bin/bash

# Colors for console output
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Print header
echo -e "${GREEN}====================================${NC}"
echo -e "${GREEN}  GoText Application - Development   ${NC}"
echo -e "${GREEN}====================================${NC}"

# Function to check if a command exists
command_exists() {
  command -v "$1" >/dev/null 2>&1
}

# Handle stop command
if [ "$1" = "stop" ]; then
  echo -e "${YELLOW}Stopping all Docker containers...${NC}"
  if command_exists docker && command_exists docker-compose; then
    docker-compose down
    echo -e "${GREEN}Docker containers stopped successfully.${NC}"
  else
    echo -e "${RED}Docker or docker-compose not found. Cannot stop containers.${NC}"
  fi
  exit 0
fi

# Check if Docker is installed
if command_exists docker && command_exists docker-compose; then
  echo -e "${YELLOW}Docker detected! Running with Docker...${NC}"
  
  # Shut down any running Docker containers
  echo -e "${GREEN}Shutting down any running Docker containers...${NC}"
  docker-compose down 2>/dev/null
  
  # Build and start Docker containers
  echo -e "${GREEN}Starting Docker containers...${NC}"
  docker-compose up --build
  
  exit 0
fi

# Check if Make is installed
if command_exists make; then
  echo -e "${YELLOW}Make detected! Running with Makefile...${NC}"
  
  # Run the application with Make
  make run-all
  
  exit 0
fi

# Fallback to manual setup if neither Docker nor Make are available
echo -e "${YELLOW}Neither Docker nor Make detected. Running manually...${NC}"

# Start PostgreSQL if it's not running
if ! pgrep -x postgres > /dev/null; then
  echo -e "${GREEN}Starting PostgreSQL...${NC}"
  brew services start postgresql || { 
    echo -e "${RED}Failed to start PostgreSQL. Is it installed?${NC}"; 
    exit 1;
  }
fi

# Create database if it doesn't exist
echo -e "${GREEN}Creating database (if it doesn't exist)...${NC}"
createdb gotext 2>/dev/null || echo -e "${YELLOW}Database gotext already exists.${NC}"

# Apply schema
echo -e "${GREEN}Applying database schema...${NC}"
psql -d gotext -f server/internal/db/schema.sql

# Start server in background
echo -e "${GREEN}Starting Go server...${NC}"
cd server && go run cmd/server/main.go &
SERVER_PID=$!
cd ..

# Wait a moment for server to start
sleep 2

# Start client
echo -e "${GREEN}Starting React client...${NC}"
cd client && npm start &
CLIENT_PID=$!
cd ..

# Setup trap to kill processes on script exit
trap "kill $SERVER_PID $CLIENT_PID 2>/dev/null" EXIT

# Wait for processes
wait 