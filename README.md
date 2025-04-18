# CChat - Modern Chat Application

A real-time chat application with a Go backend and web frontend.

## Features

- Secure user authentication with email verification
- Multiple chat spaces for different topics or groups
- Direct messaging between users
- Group messaging in spaces
- Real-time communication using WebSockets

## Tech Stack

### Backend
- Go
- PostgreSQL
- WebSockets for real-time communication
- JWT for authentication

### Frontend
- React
- TypeScript
- Modern UI libraries

## Project Structure

```
cchat/
├── server/              # Go backend server
│   ├── cmd/             # Main applications and entry points
│   └── internal/        # Private application code
│       ├── api/         # API handlers
│       ├── auth/        # Authentication logic
│       ├── db/          # Database connection and queries
│       ├── models/      # Data models
│       ├── middleware/  # HTTP middleware
│       ├── spaces/      # Chat spaces logic
│       ├── users/       # User management
│       └── messages/    # Messaging functionality
└── client/              # Web frontend
    ├── public/          # Static assets
    └── src/             # Source code
        ├── components/  # React components
        ├── pages/       # Page components
        ├── hooks/       # Custom React hooks
        ├── services/    # API services
        └── styles/      # CSS and styling
```

## Quick Start

### Option 1: Using the Run Script (Easiest)

```bash
# One command to start everything
./run.sh
```

This script will automatically detect if you have Docker or Make installed and use the appropriate method to run the application. If neither is available, it will start everything manually.

### Option 2: Using Make

```bash
# Initial setup (database, dependencies)
make setup

# Run both server and client
make run-all

# Or run them separately
make run-server
make run-client

# See all available commands
make help
```

### Option 3: Using Docker (Recommended for Contributors)

```bash
# Start all services with Docker Compose
docker-compose up

# Start in detached mode
docker-compose up -d

# Stop all services
docker-compose down

# Rebuild containers after changes to Dockerfiles
docker-compose up --build
```

Docker is the recommended approach for contributors as it provides a consistent development environment across all platforms.

### Option 4: Manual Setup

#### Prerequisites
- Go 1.24+
- Node.js 18+
- PostgreSQL

#### Backend Setup
```bash
cd server
go mod tidy
go run cmd/server/main.go
```

#### Frontend Setup
```bash
cd client
npm install
npm start
```

## VS Code Integration

For VS Code users, we provide built-in tasks for running the application:

1. Press `Cmd+Shift+B` (macOS) or `Ctrl+Shift+B` (Windows/Linux) to run the application
2. Or open the Command Palette (`Cmd+Shift+P` / `Ctrl+Shift+P`) and select "Tasks: Run Build Task"

This will execute the run.sh script which handles all the setup and startup automatically.

## API Endpoints

### Authentication
- `POST /api/auth/register` - Register a new user
- `POST /api/auth/login` - Login and get JWT token
- `GET /api/user/profile` - Get current user profile (protected)

## Development

See [TODO.md](./TODO.md) for the current development status and upcoming tasks.

## License

This project is licensed under the MIT License - see the LICENSE file for details. 