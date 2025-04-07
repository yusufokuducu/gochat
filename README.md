# GoChat - Real-Time Chat Application

GoChat is a modern, real-time chat application built with Go, featuring WebSocket communication, file sharing capabilities, and a robust PostgreSQL backend.

## 📋 Project Overview

GoChat is a full-featured chat application that enables real-time communication between users through a web interface. The application is built using Go (Golang) for the backend, PostgreSQL for data storage, and a simple but functional HTML/CSS/JavaScript frontend.

## 🚀 Features

- **Real-time messaging** using WebSocket protocol
- **User authentication** with username-based identification
- **Message history** with persistent storage in PostgreSQL
- **File sharing** with support for various file types
- **Emoji support** for expressive messaging
- **Online user status** updates in real-time
- **Responsive web interface** that works on desktop and mobile devices
- **Automatic reconnection** if connection is lost
- **System notifications** for user join/leave events

## 🛠 Technology Stack

### Backend
- **Go (Golang)** - Primary programming language
- **Gin Web Framework** - HTTP web framework for routing and middleware
- **Gorilla WebSocket** - WebSocket implementation for real-time communication
- **GORM** - Object-Relational Mapping library for database operations
- **PostgreSQL** - Relational database for persistent storage
- **Godotenv** - Environment variable management

### Frontend
- **HTML5** - Structure of the web interface
- **CSS3** - Styling and responsive design
- **JavaScript** - Client-side functionality and WebSocket communication
- **Font Awesome** - Icons for improved user experience

## 🚀 Running the Application

### Prerequisites
- Go 1.16 or higher
- PostgreSQL 12 or higher
- Git

### Setup and Installation
1. Clone the repository:
   ```bash
   git clone https://github.com/faust-lvii/gochat.git
   cd gochat
   ```

2. Install dependencies:
   ```bash
   go mod download
   ```

3. Set up PostgreSQL database:
   - Create a database named `gochat`
   - Update database connection details in `.env` file or environment variables

4. Start the server:
   ```bash
   go run cmd/server/main.go
   ```

5. Access the application at `http://localhost:8080`

## 🏗️ Project Structure

```
gochat/
├── cmd/                    # Application entry points
│   └── server/             # Main server application
│       └── main.go         # Server initialization and configuration
├── internal/               # Private application code
│   ├── handlers/           # HTTP and WebSocket request handlers
│   └── websocket/          # WebSocket implementation
│       ├── client.go       # Client connection management
│       ├── hub.go          # Central message hub
│       └── models.go       # Data models for messages and attachments
├── pkg/                    # Public libraries that can be used by external applications
│   └── handlers/           # Shared handlers
├── static/                 # Static web files
│   ├── css/                # CSS stylesheets
│   ├── js/                 # JavaScript files
│   └── index.html          # Main HTML page
├── uploads/                # File upload storage directory
└── go.mod                  # Go module dependencies
```

## 👥 User Capabilities

Users of the GoChat application can:

1. **Join the chat** by entering a username
2. **Send text messages** that are delivered in real-time to all users
3. **Use emojis** to express emotions in messages
4. **Upload and share files** with other users
5. **View message history** from previous conversations
6. **See who is online** in the user sidebar
7. **Receive notifications** when users join or leave
8. **Download shared files** from other users
9. **Automatically reconnect** if the connection is lost

## 🔒 Security Considerations

- All file uploads are validated for size and type
- WebSocket connections include basic authentication
- Database credentials should be secured using environment variables
- Content Security Policy headers are implemented
- Input validation is performed on all user inputs

## 📜 License

This project is licensed under the MIT License - see the LICENSE file for details.

-_-_-_-_-