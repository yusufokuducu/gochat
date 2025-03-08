let ws;
let username = localStorage.getItem('username');
let messageHistory = [];
let onlineUsers = new Set();
const messageForm = document.getElementById('message-form');
const messageInput = document.getElementById('message-input');
const fileInput = document.getElementById('file-input');
const emojiButton = document.querySelector('.emoji-toggle');
const emojiPicker = document.querySelector('.emoji-picker');
const usernameModal = document.getElementById('username-modal');
const usernameInput = document.getElementById('username-input');
const usernameSubmit = document.getElementById('username-submit');
const userList = document.getElementById('user-list');

// Emoji listesi
const emojis = ['😀', '😃', '😄', '😁', '😅', '😂', '🤣', '😊', '😇', '🙂', '🙃', '😉', '😌', '😍', '🥰', '😘',
                '😎', '🤓', '🧐', '🤔', '🤨', '😐', '😑', '😶', '😏', '😒', '🙄', '😬', '🤥', '😌', '😔', '😪',
                '❤️', '🧡', '💛', '💚', '💙', '💜', '🖤', '💔', '❣️', '💕', '💞', '💓', '💗', '💖', '💘', '💝',
                '👍', '👎', '👊', '✊', '🤛', '🤜', '🤞', '✌️', '🤟', '🤘', '👌', '👈', '👉', '👆', '👇', '☝️'];

// Emoji picker'ı oluştur
function initEmojiPicker() {
    emojiPicker.innerHTML = '';
    emojis.forEach(emoji => {
        const button = document.createElement('button');
        button.textContent = emoji;
        button.className = 'emoji-button';
        button.addEventListener('click', (e) => {
            e.preventDefault();
            e.stopPropagation();
            messageInput.value += emoji;
            toggleEmojiPicker(false);
        });
        emojiPicker.appendChild(button);
    });
}

// Emoji picker'ı toggle yapma fonksiyonu
function toggleEmojiPicker(show) {
    if (show) {
        emojiPicker.style.display = 'grid';
    } else {
        emojiPicker.style.display = 'none';
    }
}

// Emoji picker'ı başlat
initEmojiPicker();

// Kullanıcı adı modal'ını göster
function showUsernameModal() {
    usernameModal.style.display = 'flex';
    if (username) {
        usernameInput.value = username;
    }
    usernameInput.focus();
}

// Kullanıcı adı modal'ını gizle
function hideUsernameModal() {
    usernameModal.style.display = 'none';
}

// Kullanıcı adı girişi
usernameSubmit.addEventListener('click', (e) => {
    e.preventDefault();
    const newUsername = usernameInput.value.trim();
    if (newUsername) {
        if (newUsername.length > 50) {
            alert('Kullanıcı adı 50 karakterden kısa olmalıdır!');
            return;
        }
        username = newUsername;
        localStorage.setItem('username', username);
        hideUsernameModal();
        // Eğer WebSocket bağlantısı varsa kapat ve yeniden bağlan
        if (ws) {
            ws.close();
        }
        connectWebSocket();
    } else {
        alert('Lütfen geçerli bir kullanıcı adı girin!');
    }
});

// Enter tuşu ile kullanıcı adı girişi
usernameInput.addEventListener('keypress', (e) => {
    if (e.key === 'Enter') {
        e.preventDefault();
        usernameSubmit.click();
    }
});

// Sayfa yüklendiğinde kullanıcı adı modal'ını göster
window.addEventListener('load', () => {
    if (!username) {
        showUsernameModal();
    } else {
        connectWebSocket();
    }
});

// Kullanıcı listesini güncelle
function updateUserList(users) {
    userList.innerHTML = '';
    
    // Kullanıcıları alfabetik sıraya göre sırala
    const sortedUsers = Array.from(users).sort();
    
    sortedUsers.forEach(user => {
        const li = document.createElement('li');
        li.className = 'user-item online';
        li.innerHTML = `
            <div class="user-status"></div>
            <span>${user}</span>
        `;
        userList.appendChild(li);
    });
}

// WebSocket bağlantısı
function connectWebSocket() {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const wsUrl = `${protocol}//${window.location.host}/ws?username=${encodeURIComponent(username)}`;
    console.log(`WebSocket bağlantısı kuruluyor: ${wsUrl}`);
    
    ws = new WebSocket(wsUrl);

    ws.onopen = () => {
        console.log('WebSocket bağlantısı kuruldu');
        // Bağlantı kurulduğunda geçmiş mesajları iste
        const historyRequest = { 
            type: 'get_history',
            sender: username,
            sent_at: new Date().toISOString()
        };
        console.log('Mesaj geçmişi isteniyor:', historyRequest);
        ws.send(JSON.stringify(historyRequest));
    };

    ws.onmessage = (event) => {
        try {
            console.log('Mesaj alındı:', event.data.substring(0, 100) + (event.data.length > 100 ? '...' : ''));
            const message = JSON.parse(event.data);
            if (Array.isArray(message)) {
                // Geçmiş mesajları göster
                console.log(`${message.length} adet geçmiş mesaj alındı`);
                messageHistory = [];
                const chatMessages = document.getElementById('chat-messages');
                chatMessages.innerHTML = '';
                message.forEach(msg => {
                    messageHistory.push(msg);
                    displayMessage(msg);
                });
            } else if (message.type === 'user_list') {
                // Kullanıcı listesini güncelle
                console.log('Kullanıcı listesi alındı:', message.users);
                onlineUsers = new Set(message.users);
                updateUserList(onlineUsers);
            } else {
                console.log('Tekil mesaj alındı:', message.type, message.sender);
                
                // Kullanıcı giriş/çıkış mesajlarını işle
                if (message.type === 'user_joined') {
                    onlineUsers.add(message.sender);
                    updateUserList(onlineUsers);
                } else if (message.type === 'user_left') {
                    onlineUsers.delete(message.sender);
                    updateUserList(onlineUsers);
                }
                
                messageHistory.push(message);
                displayMessage(message);
            }
        } catch (error) {
            console.error('Mesaj işleme hatası:', error);
        }
    };

    ws.onclose = (event) => {
        console.log(`WebSocket bağlantısı kapandı. Kod: ${event.code}, Neden: ${event.reason}`);
        setTimeout(connectWebSocket, 2000);
    };

    ws.onerror = (err) => {
        console.error('WebSocket hatası:', err);
    };
}

// Emoji picker toggle işlevi
let isEmojiPickerVisible = false;
emojiButton.addEventListener('click', (e) => {
    e.preventDefault();
    e.stopPropagation();
    isEmojiPickerVisible = !isEmojiPickerVisible;
    toggleEmojiPicker(isEmojiPickerVisible);
});

// Sayfanın başka bir yerine tıklandığında emoji picker'ı gizle
document.addEventListener('click', () => {
    if (isEmojiPickerVisible) {
        isEmojiPickerVisible = false;
        toggleEmojiPicker(false);
    }
});

// Emoji picker'ın içine tıklandığında kapanmasını engelle
emojiPicker.addEventListener('click', (e) => {
    e.stopPropagation();
});

// Dosya yükleme işlevi
fileInput.addEventListener('change', async (e) => {
    const file = e.target.files[0];
    if (!file) return;

    const formData = new FormData();
    formData.append('file', file);

    try {
        const response = await fetch('/upload', {
            method: 'POST',
            body: formData
        });

        if (!response.ok) throw new Error('Dosya yükleme hatası');

        const result = await response.json();
        const message = {
            type: 'message',
            content: `Dosya gönderildi: ${file.name}`,
            sender: username,
            sent_at: new Date().toISOString(),
            attachments: [{
                file_url: result.url,
                file_name: file.name,
                file_type: file.type
            }]
        };

        ws.send(JSON.stringify(message));
        fileInput.value = ''; // Input'u temizle
    } catch (error) {
        console.error('Dosya yükleme hatası:', error);
        alert('Dosya yüklenirken bir hata oluştu');
    }
});

// Enter tuşu ile emoji ekleme sorununu çöz
messageInput.addEventListener('keypress', (e) => {
    if (e.key === 'Enter' && !e.shiftKey) {
        e.preventDefault();
        messageForm.dispatchEvent(new Event('submit'));
    }
});

// Mesaj gönderme
messageForm.addEventListener('submit', (e) => {
    e.preventDefault();
    const content = messageInput.value.trim();
    if (!content) return;

    const message = {
        type: 'message',
        content: content,
        sender: username,
        sent_at: new Date().toISOString(),
        attachments: []
    };

    if (ws && ws.readyState === WebSocket.OPEN) {
        // Mesajı WebSocket üzerinden gönder
        ws.send(JSON.stringify(message));
        // Input'u temizle
        messageInput.value = '';
        // Emoji picker'ı kapat
        isEmojiPickerVisible = false;
        toggleEmojiPicker(false);
    } else {
        console.error('WebSocket bağlantısı kapalı');
        alert('Bağlantı hatası! Sayfa yenileniyor...');
        window.location.reload();
    }
});

// Mesajları görüntüleme
function displayMessage(message) {
    const chatMessages = document.getElementById('chat-messages');
    const messageDiv = document.createElement('div');

    if (message.type === 'system') {
        messageDiv.className = 'message system-message';
        messageDiv.innerHTML = `
            <div class="message-content">
                <p>${message.content}</p>
                <span class="timestamp">${new Date(message.sent_at).toLocaleTimeString()}</span>
            </div>
        `;
    } else {
        const isOwnMessage = message.sender === username;
        messageDiv.className = `message ${isOwnMessage ? 'sent' : 'received'}`;
        
        let attachmentHtml = '';
        if (message.attachments && message.attachments.length > 0) {
            attachmentHtml = message.attachments.map(attachment => {
                if (attachment.file_type.startsWith('image/')) {
                    return `<img src="${attachment.file_url}" alt="${attachment.file_name}" class="message-image">`;
                } else {
                    return `<a href="${attachment.file_url}" target="_blank" class="file-attachment">
                        <i class="fas fa-file"></i> ${attachment.file_name}
                    </a>`;
                }
            }).join('');
        }

        messageDiv.innerHTML = `
            <div class="message-header">
                <span class="username">${message.sender}</span>
                <span class="timestamp">${new Date(message.sent_at).toLocaleTimeString()}</span>
            </div>
            <div class="message-content">
                <p>${message.content}</p>
                ${attachmentHtml}
            </div>
        `;
    }

    chatMessages.appendChild(messageDiv);
    chatMessages.scrollTop = chatMessages.scrollHeight;
}