# GoChat Backend (Go Version)

Bu proje, GoChat uygulamasının Go dilinde yazılmış backend kısmıdır. Önceki Python (FastAPI) tabanlı backend'in Go diline çevrilmiş halidir.

## Özellikler

- JWT tabanlı kimlik doğrulama
- Kullanıcı yönetimi
- Arkadaşlık sistemi (ekleme, kabul etme, reddetme, silme)
- WebSocket ile gerçek zamanlı mesajlaşma
- PostgreSQL veritabanı entegrasyonu
- Docker ile kolay dağıtım

## Teknolojiler

- [Go](https://golang.org/) - Programlama dili
- [Gin](https://github.com/gin-gonic/gin) - Web framework
- [GORM](https://gorm.io/) - ORM kütüphanesi
- [JWT](https://github.com/golang-jwt/jwt) - JSON Web Token kütüphanesi
- [WebSocket](https://github.com/gorilla/websocket) - WebSocket kütüphanesi
- [PostgreSQL](https://www.postgresql.org/) - Veritabanı

## Proje Yapısı

```
backend-go/
├── main.go                  # Ana uygulama dosyası
├── go.mod                   # Go modül tanımlaması
├── go.sum                   # Bağımlılık sürümleri
└── Dockerfile               # Docker yapılandırması
```

## Başlangıç

### Gereksinimler

- Go 1.16+
- PostgreSQL
- Docker ve Docker Compose (opsiyonel)

### Docker ile Çalıştırma

1. Projeyi klonlayın:
   ```
   git clone https://github.com/faust-lvii/gochat.git
   cd gochat
   ```

2. Docker Compose ile uygulamayı başlatın:
   ```
   docker-compose up -d
   ```

3. Uygulamaya erişin:
   - Frontend: http://localhost:3000
   - Backend API: http://localhost:8000
   - API Belgeleri: http://localhost:8000/swagger/index.html (ileride eklenecek)

### Varsayılan Giriş Bilgileri

Uygulama, varsayılan bir admin kullanıcısı ile başlatılır:
- Kullanıcı adı: `admin1`
- Şifre: `admin1234`

### Yerel Geliştirme

1. Backend dizinine gidin:
   ```
   cd backend-go
   ```

2. Bağımlılıkları yükleyin:
   ```
   go mod download
   ```

3. Uygulamayı çalıştırın:
   ```
   go run main.go
   ```

## API Endpoints

### Kimlik Doğrulama
- `POST /api/auth/login` - Giriş yapma
- `POST /api/auth/register` - Kayıt olma

### Kullanıcılar
- `GET /api/users` - Tüm kullanıcıları listeleme
- `GET /api/users/:id` - Belirli bir kullanıcıyı görüntüleme
- `PUT /api/users/:id` - Kullanıcı bilgilerini güncelleme

### Arkadaşlıklar
- `GET /api/friendships` - Arkadaşlıkları listeleme
- `POST /api/friendships` - Arkadaşlık isteği gönderme
- `PUT /api/friendships/:id` - Arkadaşlık isteğini güncelleme (kabul/red)
- `DELETE /api/friendships/:id` - Arkadaşlığı silme

### Mesajlar
- `GET /api/messages` - Mesajları listeleme
- `POST /api/messages` - Yeni mesaj gönderme
- `PUT /api/messages/:id/read` - Mesajı okundu olarak işaretleme

### WebSocket
- `GET /api/ws/:token` - WebSocket bağlantısı

## Sorun Giderme

### CORS Sorunları
Frontend, backend ile iletişim kurarken CORS sorunlarıyla karşılaşırsanız:
1. Backend'deki CORS ayarlarının frontend kaynağınızı içerdiğinden emin olun
2. Frontend API URL'sinin `http://localhost:8000/api` olarak ayarlandığından emin olun

### Veritabanı Bağlantısı
Backend veritabanına bağlanamazsa:
1. Veritabanı bağlantı dizesini kontrol edin
2. PostgreSQL konteynerinin çalıştığından emin olun: `docker-compose ps`
