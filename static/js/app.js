class ChatApp {
    constructor() {
        this.ws = null;
        this.messageHistory = document.getElementById('messageHistory');
        this.messageForm = document.getElementById('messageForm');
        this.messageInput = document.getElementById('messageInput');
        this.fileInput = document.getElementById('fileInput');
        this.userList = document.getElementById('userList');
        this.reconnectAttempts = 0;
        this.maxReconnectAttempts = 5;
        this.reconnectDelay = 3000;
        this.username = new URLSearchParams(window.location.search).get('username') || 'Anonim';
        this.emojiButton = document.getElementById('emojiButton');
        this.emojiPicker = null;
        
        this.setupWebSocket();
        this.setupEventListeners();
        this.initEmojiPicker();
    }

    setupWebSocket() {
        if (this.reconnectAttempts >= this.maxReconnectAttempts) {
            this.showSystemMessage('Maksimum yeniden bağlanma denemesi aşıldı. Lütfen sayfayı yenileyin.');
            return;
        }

        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        const wsUrl = `${protocol}//${window.location.host}/ws?username=${encodeURIComponent(this.username)}`;
        
        try {
            this.ws = new WebSocket(wsUrl);
            
            this.ws.onopen = () => {
                console.log('WebSocket bağlantısı kuruldu');
                this.showSystemMessage('Sohbete bağlandınız!');
                this.reconnectAttempts = 0;
            };
            
            this.ws.onmessage = (event) => {
                try {
                    const message = JSON.parse(event.data);
                    this.handleMessage(message);
                } catch (error) {
                    console.error('Mesaj işleme hatası:', error);
                }
            };
            
            this.ws.onclose = (event) => {
                console.log('WebSocket bağlantısı kapandı:', event.code, event.reason);
                this.showSystemMessage('Bağlantı kesildi. Yeniden bağlanmaya çalışılıyor...');
                this.reconnectAttempts++;
                setTimeout(() => this.setupWebSocket(), this.reconnectDelay);
            };
            
            this.ws.onerror = (error) => {
                console.error('WebSocket hatası:', error);
                this.showSystemMessage('Bir bağlantı hatası oluştu!');
            };
        } catch (error) {
            console.error('WebSocket bağlantı hatası:', error);
            this.showSystemMessage('Bağlantı kurulamadı. Yeniden deneniyor...');
            this.reconnectAttempts++;
            setTimeout(() => this.setupWebSocket(), this.reconnectDelay);
        }
    }

    setupEventListeners() {
        this.messageForm.addEventListener('submit', (e) => {
            e.preventDefault();
            this.sendMessage();
        });

        this.messageInput.addEventListener('keypress', (e) => {
            if (e.key === 'Enter' && !e.shiftKey) {
                e.preventDefault();
                this.sendMessage();
            }
        });

        this.fileInput.addEventListener('change', (e) => {
            const file = e.target.files[0];
            if (file) {
                this.sendFile(file);
            }
        });
    }

    async sendFile(file) {
        if (!file || !this.ws || this.ws.readyState !== WebSocket.OPEN) return;

        // Dosya boyutu kontrolü (10MB)
        if (file.size > 10 * 1024 * 1024) {
            this.showSystemMessage('Dosya boyutu 10MB\'dan büyük olamaz!');
            return;
        }

        try {
            // Yükleniyor mesajı göster
            const loadingMsg = document.createElement('div');
            loadingMsg.className = 'text-center text-sm text-gray-500 my-2';
            loadingMsg.textContent = `${file.name} yükleniyor...`;
            this.messageHistory.appendChild(loadingMsg);
            this.scrollToBottom();

            // Dosyayı Base64'e çevir
            const base64Data = await this.fileToBase64(file);
            
            const messageData = {
                type: 'file',
                content: base64Data,
                sender: this.username,
                timestamp: new Date().toISOString(),
                fileInfo: {
                    fileName: file.name,
                    fileSize: file.size,
                    fileType: file.type
                }
            };

            // WebSocket üzerinden gönder
            this.ws.send(JSON.stringify(messageData));
            
            // Yükleniyor mesajını kaldır
            this.messageHistory.removeChild(loadingMsg);
            
            // Input'u temizle
            this.fileInput.value = '';
        } catch (error) {
            console.error('Dosya gönderme hatası:', error);
            this.showSystemMessage('Dosya gönderilemedi!');
        }
    }

    fileToBase64(file) {
        return new Promise((resolve, reject) => {
            const reader = new FileReader();
            reader.readAsDataURL(file);
            reader.onload = () => {
                const base64String = reader.result.split(',')[1];
                resolve(base64String);
            };
            reader.onerror = (error) => reject(error);
        });
    }

    sendMessage() {
        const content = this.messageInput.value.trim();
        if (!content || !this.ws || this.ws.readyState !== WebSocket.OPEN) return;

        const messageData = {
            type: 'message',
            content: content,
            sender: this.username,
            timestamp: new Date().toISOString()
        };

        try {
            this.ws.send(JSON.stringify(messageData));
            
            // Mesajı hemen göster
            this.displayMessage({
                ...messageData,
                isSelf: true
            });
            
            this.messageInput.value = '';
        } catch (error) {
            console.error('Mesaj gönderme hatası:', error);
            this.showSystemMessage('Mesaj gönderilemedi!');
        }
    }

    handleMessage(message) {
        console.log('Gelen mesaj:', message);
        switch (message.type) {
            case 'message':
                // Kendi gönderdiğimiz mesajları tekrar gösterme
                if (message.sender !== this.username) {
                    this.displayMessage({
                        ...message,
                        isSelf: false
                    });
                }
                break;
            case 'file':
                // Kendi gönderdiğimiz dosyaları tekrar gösterme
                if (message.sender !== this.username) {
                    this.displayMessage({
                        ...message,
                        isSelf: false
                    });
                }
                break;
            case 'user_list':
                this.updateUserList(message.users);
                break;
            case 'system':
                this.showSystemMessage(message.content);
                break;
            case 'typing':
                this.handleTypingStatus(message);
                break;
            default:
                console.warn('Bilinmeyen mesaj tipi:', message.type);
        }
    }

    displayMessage(message) {
        const messageElement = document.createElement('div');
        messageElement.className = `message ${message.isSelf ? 'sent' : 'received'}`;
        
        const sender = document.createElement('div');
        sender.className = 'message-sender text-xs text-gray-600 mb-1';
        sender.textContent = message.sender;
        
        const content = document.createElement('div');
        content.className = 'message-content';

        if (message.type === 'file') {
            const fileInfo = message.fileInfo;
            if (fileInfo.fileType.startsWith('image/')) {
                // Resim dosyası
                const img = document.createElement('img');
                img.src = fileInfo.fileURL;
                img.className = 'max-w-full rounded-lg cursor-pointer';
                img.onclick = () => window.open(fileInfo.fileURL, '_blank');
                content.appendChild(img);
            } else {
                // Diğer dosya türleri
                const fileLink = document.createElement('a');
                fileLink.href = fileInfo.fileURL;
                fileLink.target = '_blank';
                fileLink.className = 'flex items-center space-x-2 text-blue-500 hover:text-blue-700';
                
                const icon = document.createElement('i');
                icon.className = 'fas fa-file';
                fileLink.appendChild(icon);
                
                const fileName = document.createElement('span');
                fileName.textContent = fileInfo.fileName;
                fileLink.appendChild(fileName);
                
                content.appendChild(fileLink);
            }
        } else {
            content.textContent = message.content;
        }
        
        const timestamp = document.createElement('div');
        timestamp.className = 'message-timestamp text-xs text-gray-500 mt-1';
        timestamp.textContent = new Date(message.timestamp).toLocaleTimeString();
        
        messageElement.appendChild(sender);
        messageElement.appendChild(content);
        messageElement.appendChild(timestamp);
        
        this.messageHistory.appendChild(messageElement);
        this.scrollToBottom();
    }

    getCurrentUsername() {
        return this.username;
    }

    updateUserList(users) {
        this.userList.innerHTML = '';
        if (typeof users === 'string') {
            try {
                users = JSON.parse(users).users;
            } catch (error) {
                console.error('Kullanıcı listesi ayrıştırma hatası:', error);
                return;
            }
        }
        
        users.forEach(user => {
            const userElement = document.createElement('div');
            userElement.className = 'user-item flex items-center space-x-2';
            
            const status = document.createElement('span');
            status.className = `w-2 h-2 rounded-full ${user.online ? 'bg-green-500' : 'bg-gray-400'}`;
            
            const name = document.createElement('span');
            name.textContent = user.name;
            
            userElement.appendChild(status);
            userElement.appendChild(name);
            this.userList.appendChild(userElement);
        });
    }

    showSystemMessage(content) {
        const messageElement = document.createElement('div');
        messageElement.className = 'text-center text-sm text-gray-500 my-2';
        messageElement.textContent = content;
        this.messageHistory.appendChild(messageElement);
        this.scrollToBottom();
    }

    scrollToBottom() {
        this.messageHistory.scrollTop = this.messageHistory.scrollHeight;
    }

    handleTypingStatus(message) {
        const typingIndicator = document.getElementById('typingIndicator');
        if (!typingIndicator) {
            const indicator = document.createElement('div');
            indicator.id = 'typingIndicator';
            indicator.className = 'text-sm text-gray-500 italic ml-4 mb-2';
            document.querySelector('.messages').appendChild(indicator);
        }

        if (message.content && message.sender !== this.username) {
            typingIndicator.textContent = `${message.sender} yazıyor...`;
            typingIndicator.classList.remove('hidden');
        } else {
            typingIndicator.classList.add('hidden');
        }
    }

    initEmojiPicker() {
        // Emoji picker'ı oluştur
        this.emojiPicker = document.createElement('div');
        this.emojiPicker.style.position = 'absolute';
        this.emojiPicker.style.bottom = '80px';
        this.emojiPicker.style.left = '20px';
        this.emojiPicker.style.zIndex = '1000';
        this.emojiPicker.style.display = 'none';
        document.body.appendChild(this.emojiPicker);

        // Emoji picker'ı başlat
        const picker = new EmojiMart.Picker({
            data: emojiMartData,
            onEmojiSelect: (emoji) => {
                const cursorPos = this.messageInput.selectionStart;
                const text = this.messageInput.value;
                const newText = text.slice(0, cursorPos) + emoji.native + text.slice(cursorPos);
                this.messageInput.value = newText;
                this.messageInput.focus();
                const newCursorPos = cursorPos + emoji.native.length;
                this.messageInput.setSelectionRange(newCursorPos, newCursorPos);
                this.emojiPicker.style.display = 'none';
            },
            theme: 'light',
            set: 'native',
            showPreview: false,
            showSkinTones: false,
            autoFocus: false,
            maxFrequentRows: 4,
            perLine: 8
        });

        this.emojiPicker.appendChild(picker);

        // Emoji butonuna tıklama olayını ekle
        this.emojiButton.addEventListener('click', (e) => {
            e.stopPropagation();
            const isVisible = this.emojiPicker.style.display === 'block';
            this.emojiPicker.style.display = isVisible ? 'none' : 'block';
        });

        // Sayfa tıklamalarını dinle ve emoji picker dışında tıklanırsa kapat
        document.addEventListener('click', (e) => {
            if (!this.emojiPicker.contains(e.target) && e.target !== this.emojiButton) {
                this.emojiPicker.style.display = 'none';
            }
        });
    }
}

// Chat uygulamasını başlat
document.addEventListener('DOMContentLoaded', () => {
    window.chatApp = new ChatApp();

    // Yazıyor göstergesi
    let typingTimer;
    const TYPING_TIMEOUT = 1000;

    document.getElementById('messageInput').addEventListener('input', () => {
        clearTimeout(typingTimer);
        sendTypingStatus(true);

        typingTimer = setTimeout(() => {
            sendTypingStatus(false);
        }, TYPING_TIMEOUT);
    });

    function sendTypingStatus(isTyping) {
        if (window.chatApp.ws.readyState === WebSocket.OPEN) {
            window.chatApp.ws.send(JSON.stringify({
                type: 'typing',
                sender: window.chatApp.getCurrentUsername(),
                content: isTyping
            }));
        }
    }
}); 