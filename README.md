# GoChat - Modern Real-time Chat Application

<p align="center">
  <img src="frontend/public/logo192.png" alt="GoChat Logo" width="100" />
</p>

<p align="center">
  <a href="#features">Features</a> •
  <a href="#architecture">Architecture</a> •
  <a href="#tech-stack">Tech Stack</a> •
  <a href="#getting-started">Getting Started</a> •
  <a href="#api-documentation">API Documentation</a> •
  <a href="#development">Development</a> •
  <a href="#deployment">Deployment</a> •
  <a href="#contributing">Contributing</a> •
  <a href="#license">License</a>
</p>

GoChat is a high-performance, scalable real-time chat application built with a modern tech stack. It features a Go-based backend with a modular architecture and a responsive React frontend, providing a seamless messaging experience with robust security features.

## Features

### User Management
- Secure JWT-based authentication
- User registration and login
- Profile management and customization
- Password hashing with bcrypt

### Messaging System
- Real-time message delivery via WebSockets
- Message history persistence
- Read receipts
- Typing indicators
- Message encryption in transit

### Friendship System
- Send and receive friend requests
- Accept or reject incoming requests
- View pending requests
- Remove existing friends
- Search for users

### UI/UX
- Responsive design for all devices
- Material-UI components
- Dark/light theme support
- Intuitive navigation
- Real-time status indicators

### Security
- JWT token authentication
- HTTPS/WSS secure connections
- Input validation and sanitization
- Protection against common web vulnerabilities
- Rate limiting for API endpoints

## Architecture

GoChat follows a clean, modular architecture that separates concerns and promotes maintainability:

```
gochat/
├── backend/                 # Go backend with modular architecture
│   ├── cmd/                 # Application entry points
│   │   └── server/          # Main server application
│   ├── internal/            # Internal packages (not importable)
│   │   ├── database/        # Database connection and operations
│   │   ├── handlers/        # HTTP request handlers
│   │   ├── middleware/      # HTTP middleware components
│   │   ├── models/          # Data models and structures
│   │   ├── utils/           # Utility functions
│   │   └── websocket/       # WebSocket connection handling
│   ├── Dockerfile           # Backend Docker configuration
│   └── go.mod               # Go module definition
├── backend-python/          # Legacy Python backend (FastAPI)
├── frontend/                # React frontend
│   ├── public/              # Static assets
│   ├── src/                 # Source code
│   │   ├── components/      # Reusable UI components
│   │   ├── pages/           # Page components
│   │   ├── services/        # API service integrations
│   │   ├── utils/           # Utility functions and hooks
│   │   └── App.js           # Main application component
│   ├── Dockerfile           # Frontend Docker configuration
│   └── package.json         # Node.js dependencies
├── docker-compose.yml       # Docker Compose configuration
└── README.md                # Project documentation
```

## Tech Stack

### Backend
- **Go (Golang)**: High-performance, statically typed language
- **Gin**: Lightweight web framework with excellent performance
- **GORM**: ORM library for database operations
- **JWT**: Authentication using JSON Web Tokens
- **Gorilla WebSocket**: WebSocket implementation for real-time communication
- **PostgreSQL**: Relational database for data persistence

### Frontend
- **React**: JavaScript library for building user interfaces
- **Material-UI**: React component library implementing Google's Material Design
- **React Router**: Declarative routing for React applications
- **Axios**: Promise-based HTTP client for API requests
- **Context API**: State management for user authentication and app state

### DevOps
- **Docker**: Containerization of application components
- **Docker Compose**: Multi-container application orchestration
- **GitHub Actions**: CI/CD pipeline for automated testing and deployment

## Getting Started

### Prerequisites
- Docker and Docker Compose
- Git

### Installation and Setup

1. Clone the repository:
   ```bash
   git clone https://github.com/faust-lvii/gochat.git
   cd gochat
   ```

2. Start the application with Docker Compose:
   ```bash
   docker-compose up -d
   ```

3. Access the application:
   - Frontend: http://localhost:3000
   - Backend API: http://localhost:8000

### Default Credentials
The application is initialized with a default admin user:
- Username: `admin`
- Password: `admin`

## API Documentation

### Authentication Endpoints
- `POST /api/auth/register`: Register a new user
- `POST /api/auth/login`: Authenticate a user and get JWT token

### User Endpoints
- `GET /api/users/me`: Get current user information
- `GET /api/users`: List all users
- `GET /api/users/:id`: Get specific user information
- `PUT /api/users/:id`: Update user information

### Friendship Endpoints
- `GET /api/friendships`: List friendships (filter by status)
- `POST /api/friendships`: Create a new friendship request
- `PUT /api/friendships/:id`: Update friendship status
- `DELETE /api/friendships/:id`: Delete a friendship

### Message Endpoints
- `GET /api/messages`: List messages with a specific friend
- `POST /api/messages`: Create a new message
- `PUT /api/messages/:id/read`: Mark a message as read

### WebSocket Endpoint
- `WS /api/ws`: WebSocket connection for real-time messaging

## Development

### Backend Development

1. Navigate to the backend directory:
   ```bash
   cd backend
   ```

2. Install Go dependencies:
   ```bash
   go mod download
   ```

3. Run the Go server:
   ```bash
   go run cmd/server/main.go
   ```

4. Environment variables:
   - `PORT`: Server port (default: 8000)
   - `DATABASE_URL`: PostgreSQL connection string
   - `SECRET_KEY`: JWT secret key
   - `ENVIRONMENT`: Development/production environment

### Frontend Development

1. Navigate to the frontend directory:
   ```bash
   cd frontend
   ```

2. Install dependencies:
   ```bash
   npm install
   ```

3. Start the development server:
   ```bash
   npm start
   ```

4. Environment variables:
   - `REACT_APP_API_URL`: Backend API URL
   - `REACT_APP_WS_URL`: WebSocket URL

## Deployment

### Docker Deployment
The application is containerized and can be deployed using Docker Compose:

```bash
docker-compose -f docker-compose.prod.yml up -d
```

### Manual Deployment

#### Backend
1. Build the Go binary:
   ```bash
   cd backend
   go build -o gochat-server cmd/server/main.go
   ```

2. Run the server:
   ```bash
   ./gochat-server
   ```

#### Frontend
1. Build the React application:
   ```bash
   cd frontend
   npm run build
   ```

2. Serve the static files using a web server like Nginx.

## Contributing

We welcome contributions to GoChat! Please follow these steps:

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/your-feature-name`
3. Commit your changes: `git commit -m 'Add some feature'`
4. Push to the branch: `git push origin feature/your-feature-name`
5. Open a pull request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Acknowledgements

- Thanks to all the open-source libraries and frameworks that made this project possible
- Special thanks to the Go and React communities for their excellent documentation and support
