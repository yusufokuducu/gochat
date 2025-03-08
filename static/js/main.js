// GoChat Application JavaScript
document.addEventListener('DOMContentLoaded', () => {
    // DOM Elements
    const messageForm = document.getElementById('message-form');
    const messageInput = document.getElementById('message-input');
    const fileInput = document.getElementById('file-input');
    const chatMessages = document.getElementById('chat-messages');
    const userList = document.getElementById('user-list');
    const loginModal = document.getElementById('login-modal');
    const usernameInput = document.getElementById('username-input');
    const loginForm = document.getElementById('login-form');
    const emojiToggle = document.getElementById('emoji-toggle');
    const emojiPicker = document.getElementById('emoji-picker');
    
    // App state
    let socket = null;
    let username = '';
    let reconnectAttempts = 0;
    let reconnectInterval = null;
    let lastActivityTime = Date.now();
    const maxReconnectAttempts = 5;
    const reconnectDelay = 3000; // 3 seconds
    const inactivityTimeout = 30 * 60 * 1000; // 30 minutes
    
    // Common emojis
    const commonEmojis = ['😀', '😂', '😊', '❤️', '👍', '🎉', '🔥', '😎', '🤔', '😢', '😡', '🥳', '👋', '🙏', '🤝', '👏'];
    
    // Initialize emoji picker
    function initEmojiPicker() {
        emojiPicker.innerHTML = '';
        commonEmojis.forEach(emoji => {
            const button = document.createElement('button');
            button.className = 'emoji-button';
            button.textContent = emoji;
            button.addEventListener('click', () => {
                messageInput.value += emoji;
                emojiPicker.style.display = 'none';
                messageInput.focus();
            });
            emojiPicker.appendChild(button);
        });
        
        emojiToggle.addEventListener('click', () => {
            emojiPicker.style.display = emojiPicker.style.display === 'grid' ? 'none' : 'grid';
        });
        
        // Close emoji picker when clicking outside
        document.addEventListener('click', (e) => {
            if (!emojiToggle.contains(e.target) && !emojiPicker.contains(e.target)) {
                emojiPicker.style.display = 'none';
            }
        });
    }
    
    // Format timestamp
    function formatTimestamp(timestamp) {
        const date = new Date(timestamp);
        return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
    }
    
    // Create message element
    function createMessageElement(message) {
        const messageElement = document.createElement('div');
        const isCurrentUser = message.sender === username;
        const isSystemMessage = message.type === 'system';
        
        if (isSystemMessage) {
            messageElement.className = 'message system-message';
            messageElement.innerHTML = `
                <div class="message-content">${sanitizeHTML(message.content)}</div>
            `;
        } else {
            messageElement.className = `message ${isCurrentUser ? 'sent' : 'received'}`;
            messageElement.innerHTML = `
                ${!isCurrentUser ? `<div class="message-sender">${sanitizeHTML(message.sender)}</div>` : ''}
                <div class="message-content">${sanitizeHTML(message.content)}</div>
                ${message.attachments && message.attachments.length > 0 ? createAttachmentHTML(message.attachments) : ''}
                <div class="message-time">${formatTimestamp(message.sent_at)}</div>
            `;
        }
        
        return messageElement;
    }
    
    // Create attachment HTML
    function createAttachmentHTML(attachments) {
        let html = '<div class="message-attachment">';
        
        attachments.forEach(attachment => {
            const fileExtension = attachment.filename.split('.').pop().toLowerCase();
            const isImage = ['jpg', 'jpeg', 'png', 'gif', 'webp'].includes(fileExtension);
            
            if (isImage) {
                html += `
                    <img class="attachment-preview" src="/uploads/${attachment.filename}" alt="Attachment" />
                `;
            }
            
            html += `
                <a class="attachment-download" href="/uploads/${attachment.filename}" target="_blank" download>
                    ${isImage ? 'Download Image' : 'Download File: ' + attachment.filename}
                </a>
            `;
        });
        
        html += '</div>';
        return html;
    }
    
    // Sanitize HTML to prevent XSS
    function sanitizeHTML(text) {
        const element = document.createElement('div');
        element.textContent = text;
        return element.innerHTML;
    }
    
    // Update user list
    function updateUserList(users) {
        userList.innerHTML = '';
        users.forEach(user => {
            const userElement = document.createElement('li');
            userElement.className = 'user-item';
            userElement.innerHTML = `
                <div class="user-status"></div>
                ${sanitizeHTML(user)}
            `;
            userList.appendChild(userElement);
        });
    }
    
    // Connect to WebSocket server
    function connectWebSocket() {
        if (socket) {
            socket.close();
        }
        
        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        const wsUrl = `${protocol}//${window.location.host}/ws?username=${encodeURIComponent(username)}`;
        
        socket = new WebSocket(wsUrl);
        
        socket.onopen = () => {
            console.log('WebSocket connection established');
            reconnectAttempts = 0;
            if (reconnectInterval) {
                clearInterval(reconnectInterval);
                reconnectInterval = null;
            }
            
            // Request message history
            sendMessage({
                type: 'history_request'
            });
        };
        
        socket.onmessage = (event) => {
            handleMessage(event.data);
        };
        
        socket.onclose = (event) => {
            console.log('WebSocket connection closed', event);
            
            if (!event.wasClean && reconnectAttempts < maxReconnectAttempts) {
                console.log(`Attempting to reconnect (${reconnectAttempts + 1}/${maxReconnectAttempts})...`);
                reconnectAttempts++;
                
                if (!reconnectInterval) {
                    reconnectInterval = setTimeout(() => {
                        connectWebSocket();
                    }, reconnectDelay);
                }
            }
        };
        
        socket.onerror = (error) => {
            console.error('WebSocket error:', error);
        };
    }
    
    // Handle incoming messages
    function handleMessage(data) {
        try {
            const message = JSON.parse(data);
            
            // Update activity timestamp
            lastActivityTime = Date.now();
            
            // Handle different message types
            if (Array.isArray(message)) {
                // Message history
                chatMessages.innerHTML = '';
                message.forEach(msg => {
                    chatMessages.appendChild(createMessageElement(msg));
                });
                scrollToBottom();
            } else if (message.type === 'user_list') {
                // User list update
                updateUserList(message.users);
            } else {
                // Regular message
                chatMessages.appendChild(createMessageElement(message));
                scrollToBottom();
            }
        } catch (error) {
            console.error('Error parsing message:', error);
        }
    }
    
    // Send message to server
    function sendMessage(message) {
        if (socket && socket.readyState === WebSocket.OPEN) {
            socket.send(JSON.stringify(message));
            return true;
        }
        return false;
    }
    
    // Scroll chat to bottom
    function scrollToBottom() {
        chatMessages.scrollTop = chatMessages.scrollHeight;
    }
    
    // Handle file selection
    fileInput.addEventListener('change', (e) => {
        if (fileInput.files.length > 0) {
            const file = fileInput.files[0];
            const maxSize = 5 * 1024 * 1024; // 5MB
            
            if (file.size > maxSize) {
                alert('File is too large. Maximum size is 5MB.');
                fileInput.value = '';
                return;
            }
            
            // Show file name in message input
            messageInput.value = `Uploading: ${file.name}`;
            messageInput.disabled = true;
            
            // Upload file
            const formData = new FormData();
            formData.append('file', file);
            formData.append('username', username);
            
            fetch('/upload', {
                method: 'POST',
                body: formData
            })
            .then(response => response.json())
            .then(data => {
                if (data.success) {
                    // Send message with attachment
                    sendMessage({
                        type: 'message',
                        content: messageInput.value !== `Uploading: ${file.name}` ? messageInput.value : 'Shared a file',
                        attachment_id: data.attachment_id
                    });
                    
                    // Reset form
                    messageInput.value = '';
                    messageInput.disabled = false;
                    fileInput.value = '';
                    messageInput.focus();
                } else {
                    alert('Failed to upload file: ' + data.error);
                    messageInput.disabled = false;
                    fileInput.value = '';
                }
            })
            .catch(error => {
                console.error('Error uploading file:', error);
                alert('Failed to upload file. Please try again.');
                messageInput.disabled = false;
                fileInput.value = '';
            });
        }
    });
    
    // Handle message form submission
    messageForm.addEventListener('submit', (e) => {
        e.preventDefault();
        
        const content = messageInput.value.trim();
        if (content) {
            const success = sendMessage({
                type: 'message',
                content: content
            });
            
            if (success) {
                messageInput.value = '';
                lastActivityTime = Date.now();
            }
        }
        
        messageInput.focus();
    });
    
    // Handle login form submission
    loginForm.addEventListener('submit', (e) => {
        e.preventDefault();
        
        username = usernameInput.value.trim();
        if (username) {
            loginModal.style.display = 'none';
            connectWebSocket();
            
            // Update header with username
            const headerUsername = document.querySelector('.header .username');
            if (headerUsername) {
                headerUsername.textContent = username;
            }
            
            // Start inactivity checker
            setInterval(checkInactivity, 60000); // Check every minute
        }
    });
    
    // Check for user inactivity
    function checkInactivity() {
        const now = Date.now();
        if (now - lastActivityTime > inactivityTimeout) {
            // User has been inactive for too long
            if (socket && socket.readyState === WebSocket.OPEN) {
                socket.close();
                showLoginModal('You were disconnected due to inactivity. Please log in again.');
            }
        }
    }
    
    // Show login modal with optional message
    function showLoginModal(message) {
        if (message) {
            const messageElement = document.getElementById('login-message');
            if (messageElement) {
                messageElement.textContent = message;
                messageElement.style.display = 'block';
            }
        }
        
        loginModal.style.display = 'flex';
        usernameInput.focus();
    }
    
    // Initialize application
    function init() {
        initEmojiPicker();
        showLoginModal();
        
        // Track user activity
        ['click', 'keypress', 'scroll', 'mousemove'].forEach(event => {
            document.addEventListener(event, () => {
                lastActivityTime = Date.now();
            });
        });
        
        // Handle window focus/blur
        window.addEventListener('focus', () => {
            if (socket && socket.readyState !== WebSocket.OPEN && username) {
                connectWebSocket();
            }
        });
    }
    
    // Start the application
    init();
});
