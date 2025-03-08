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

## 🔄 How It Works

### Server-Side Flow
1. The server initializes and connects to the PostgreSQL database
2. Database tables are automatically created or migrated if needed
3. A WebSocket hub is created to manage all client connections
4. The Gin web server starts and sets up routes for:
   - Static file serving
   - WebSocket connections
   - File uploads
5. The server listens for incoming connections

### Client-Side Flow
1. User opens the application in a browser
2. User enters a username to join the chat
3. WebSocket connection is established with the server
4. User receives message history and current online users list
5. User can send text messages, emojis, and file attachments
6. All connected users receive messages in real-time

### WebSocket Communication
1. Client establishes WebSocket connection to `/ws` endpoint
2. Server registers the client in the central hub
3. Two goroutines are started for each client:
   - `ReadPump`: Reads messages from the client and broadcasts them
   - `WritePump`: Writes messages from the hub to the client
4. Messages are JSON-encoded with the following structure:
   - `type`: Message type (message, system, file, etc.)
   - `content`: Message content
   - `sender`: Username of the sender
   - `sent_at`: Timestamp of when the message was sent
   - `attachments`: Optional file attachments

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

## 🚀 Running the Application

### Prerequisites
- Go 1.16 or higher
- PostgreSQL 12 or higher
- Git

### Setup and Installation
1. Clone the repository:
   ```bash
   git clone https://github.com/yourusername/gochat.git
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

## 🔒 Security Considerations

- All file uploads are validated for size and type
- WebSocket connections include basic authentication
- Database credentials should be secured using environment variables
- Content Security Policy headers are implemented
- Input validation is performed on all user inputs

## 🛠️ Future Improvements

- Add user authentication with password protection
- Implement private messaging between users
- Add chat rooms/channels functionality
- Enhance UI with themes and customization options
- Implement end-to-end encryption for messages
- Add message reactions and threading
- Develop mobile applications using the same backend

## 📜 License

This project is licensed under the MIT License - see the LICENSE file for details.

## 👏 Acknowledgements

- The Go Team for the amazing language
- The Gin and Gorilla WebSocket teams for their excellent libraries
- The GORM team for simplifying database operations
- The PostgreSQL team for a robust and reliable database
