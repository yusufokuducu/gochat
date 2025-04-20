# GoChat - Real-time Chat Application

GoChat is a modern, real-time chat application built with Go for the backend and React with Material-UI for the frontend.

## Features

- User authentication with JWT
- User profile management
- Friend system (add, accept, reject, remove)
- Real-time messaging using WebSockets
- Responsive UI with Material-UI components
- PostgreSQL database for data persistence
- Docker and Docker Compose for easy deployment

## Project Structure

```
gochat/
├── backend-go/               # Go backend
│   ├── main.go               # Main application file
│   ├── go.mod                # Go module definition
│   ├── go.sum                # Dependencies versions
│   └── Dockerfile            # Backend Docker configuration
├── frontend/                 # React frontend
│   ├── public/               # Static assets
│   ├── src/                  # Source code
│   │   ├── components/       # React components
│   │   ├── pages/            # Page components
│   │   ├── services/         # API services
│   │   └── utils/            # Utility functions
│   ├── Dockerfile            # Frontend Docker configuration
│   └── package.json          # Node.js dependencies
├── docker-compose.yml        # Docker Compose configuration
└── README.md                 # Project documentation
```

## Getting Started

### Prerequisites

- Docker and Docker Compose
- Node.js and npm (for local development)
- Go (for local development)

### Running with Docker

1. Clone the repository:
   ```
   git clone https://github.com/faust-lvii/gochat.git
   cd gochat
   ```

2. Start the application with Docker Compose:
   ```
   docker-compose up -d
   ```

3. Access the application:
   - Frontend: http://localhost:3000
   - Backend API: http://localhost:8000
   - API Documentation: http://localhost:8000/docs

### Default Login Credentials

The application is initialized with a default admin user:
- Username: `admin1`
- Password: `admin1234`

### Local Development Setup

#### Backend

1. Navigate to the backend directory:
   ```
   cd backend-go
   ```

2. Initialize the Go module:
   ```
   go mod init
   ```

3. Install dependencies:
   ```
   go get
   ```

4. Run the Go server:
   ```
   go run main.go
   ```

#### Frontend

1. Navigate to the frontend directory:
   ```
   cd frontend
   ```

2. Install dependencies:
   ```
   npm install
   ```

3. Run the React development server:
   ```
   npm start
   ```

## API Documentation

The API documentation is automatically generated using Swagger UI and is available at:
http://localhost:8000/docs

## Database Migrations

To create a new database migration:

```
cd backend-go
go run migrations/main.go create -m "Description of changes"
```

To apply migrations:

```
go run migrations/main.go up
```

## Troubleshooting

### CORS Issues
If you encounter CORS issues when the frontend tries to communicate with the backend, ensure that:
1. The backend CORS settings in `main.go` include your frontend origin
2. The frontend API URL in `docker-compose.yml` is set to `http://localhost:8000/api`

### Database Connection
If the backend cannot connect to the database, check:
1. The database connection string in `main.go`
2. Ensure the PostgreSQL container is running with `docker-compose ps`

## License

This project is licensed under the MIT License - see the LICENSE file for details.
