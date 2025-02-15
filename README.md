# GoChat - Real-Time Chat Application

GoChat is a modern, real-time chat application built with Go, featuring WebSocket communication, file sharing capabilities, and a robust PostgreSQL backend.

## 🚀 Features

- Real-time messaging using WebSocket
- File sharing and attachment support
- Persistent message storage with PostgreSQL
- Clean and modern web interface
- Robust error handling and retry mechanisms
- Auto-migration of database schemas

## 🛠 Tech Stack

- **Backend:**
  - Go (Golang)
  - Gin Web Framework
  - GORM (Object Relational Mapper)
  - Gorilla WebSocket

- **Database:**
  - PostgreSQL

- **Frontend:**
  - HTML5
  - JavaScript
  - Static file serving

## 📋 Prerequisites

Before running the application, make sure you have the following installed:

- Go 1.16 or higher
- PostgreSQL 12 or higher
- Git

## ⚙️ Configuration

1. Clone the repository:
```bash
git clone https://github.com/yourusername/gochat.git
cd gochat
```

2. Set up your PostgreSQL database and update the configuration in `cmd/server/main.go`:
```go
dbConfig := database.Config{
    Host:     "localhost",
    Port:     "5432",
    User:     "postgres",
    Password: "your_db_password",
    DBName:   "gochat",
    SSLMode:  "disable",
}
```

## 🚀 Running the Application

1. Install dependencies:
```bash
go mod download
```

2. Start the server:
```bash
go run cmd/server/main.go
```

3. Access the application at `http://localhost:8080`

## 📁 Project Structure

```
gochat/
├── cmd/
│   └── server/
│       └── main.go         # Application entry point
├── internal/
│   ├── database/          # Database operations and models
│   └── websocket/         # WebSocket implementation
├── pkg/
│   └── database/          # Shared database utilities
├── static/                # Static web files
├── uploads/              # File upload directory
└── data/                 # Application data directory
```

## 🔒 Security

- All file uploads are validated and stored securely
- Database credentials should be properly secured in production
- WebSocket connections are handled with proper error checking

## 🤝 Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## ✨ Acknowledgments

- Thanks to all contributors who have helped shape GoChat
- Built with ❤️ using Go and modern web technologies

## 📞 Support

For support, please open an issue in the GitHub repository or contact the maintainers.

---
Made with ❤️ by [Your Name]
