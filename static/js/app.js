// DOM Elements
const chatContainer = document.getElementById('chat-container');
const loginContainer = document.getElementById('login-container');
const chatMessages = document.getElementById('chat-messages');
const userList = document.getElementById('user-list');
const messageForm = document.getElementById('message-form');
const messageInput = document.getElementById('message-input');
const usernameInput = document.getElementById('username-input');
const joinBtn = document.getElementById('join-btn');
const logoutBtn = document.getElementById('logout-btn');
const fileInput = document.getElementById('file-input');
const fileBtn = document.getElementById('file-btn');
const emojiBtn = document.getElementById('emoji-btn');
const emojiPicker = document.getElementById('emoji-picker');
const currentUser = document.getElementById('current-user');

// Global variables
let socket = null;
let username = '';
let reconnectAttempts = 0;
const maxReconnectAttempts = 5;
const reconnectInterval = 3000; // 3 seconds

// Initialize the application
function init() {
    console.log('Initializing application...');
    
    // Check if DOM elements exist
    if (!joinBtn) {
        console.error('Join button not found');
        return;
    }
    
    // Setup event listeners
    joinBtn.addEventListener('click', joinChat);
    
    if (logoutBtn) {
        logoutBtn.addEventListener('click', leaveChat);
    }
    
    if (messageForm) {
        messageForm.addEventListener('submit', sendMessage);
    }
    
    if (fileInput && fileBtn) {
        fileBtn.addEventListener('click', () => fileInput.click());
        fileInput.addEventListener('change', handleFileUpload);
    }
    
    if (emojiBtn) {
        emojiBtn.addEventListener('click', toggleEmojiPicker);
    }
    
    document.addEventListener('click', handleDocumentClick);
    
    // Create emoji picker if element exists
    if (emojiPicker) {
        createEmojiPicker();
    } else {
        console.warn('Emoji picker element not found');
    }
    
    // Check if user was previously logged in
    const savedUsername = localStorage.getItem('gochat_username');
    if (savedUsername) {
        console.log('Found saved username:', savedUsername);
        usernameInput.value = savedUsername;
        // Auto-join chat if username exists
        setTimeout(() => {
            joinChat();
        }, 500);
    }
}

// Join the chat
function joinChat() {
    if (!usernameInput) {
        console.error('Username input not found');
        return;
    }
    
    username = usernameInput.value.trim();
    
    if (!username) {
        showError('Please enter a username');
        return;
    }
    
    console.log('Joining chat as:', username);
    
    // Save username to localStorage
    localStorage.setItem('gochat_username', username);
    
    // Connect to WebSocket
    connectWebSocket();
    
    // Update current user display
    if (currentUser) {
        currentUser.textContent = username;
    }
    
    // Show chat container
    if (loginContainer && chatContainer) {
        loginContainer.style.display = 'none';
        chatContainer.style.display = 'flex';
    } else {
        console.error('Container elements not found');
    }
    
    // Focus message input
    if (messageInput) {
        messageInput.focus();
    }
}

// Leave the chat
function leaveChat() {
    // Close WebSocket connection
    if (socket && socket.readyState === WebSocket.OPEN) {
        socket.close();
    }
    
    // Clear chat messages
    if (chatMessages) {
        chatMessages.innerHTML = '';
    }
    
    if (userList) {
        userList.innerHTML = '';
    }
    
    // Show login container
    if (loginContainer && chatContainer) {
        chatContainer.style.display = 'none';
        loginContainer.style.display = 'flex';
    }
    
    // Clear username
    localStorage.removeItem('gochat_username');
    if (usernameInput) {
        usernameInput.value = '';
        usernameInput.focus();
    }
}

// Connect to WebSocket server
function connectWebSocket() {
    // Close existing connection if any
    if (socket) {
        socket.close();
    }
    
    // Create new WebSocket connection
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const wsUrl = `${protocol}//${window.location.host}/ws?username=${encodeURIComponent(username)}`;
    
    console.log('Connecting to WebSocket:', wsUrl);
    socket = new WebSocket(wsUrl);
    
    // Setup WebSocket event handlers
    socket.onopen = handleSocketOpen;
    socket.onmessage = handleSocketMessage;
    socket.onclose = handleSocketClose;
    socket.onerror = handleSocketError;
}

// Handle WebSocket open event
function handleSocketOpen() {
    console.log('WebSocket connection established');
    reconnectAttempts = 0;
    
    // Send a test message
    const testMsg = {
        type: 'message',
        content: 'Hello, I just connected!',
        sender: username
    };
    
    // Send the test message
    if (socket.readyState === WebSocket.OPEN) {
        socket.send(JSON.stringify(testMsg));
        console.log('Sent test message:', testMsg);
    }
}

// Handle WebSocket message event
function handleSocketMessage(event) {
    console.log('Received message:', event.data);
    
    try {
        // Check if the message is empty or not valid JSON
        if (!event.data || event.data.trim() === '') {
            console.warn('Received empty message from server');
            return;
        }
        
        const message = JSON.parse(event.data);
        
        // Validate message structure
        if (!message || typeof message !== 'object') {
            console.error('Invalid message format:', event.data);
            return;
        }
        
        // Handle different message types
        switch (message.type) {
            case 'message':
                addChatMessage(message);
                break;
            case 'system':
                if (message.content === 'userStatus') {
                    updateUserList(message.data);
                } else {
                    addSystemMessage(message.content);
                }
                break;
            case 'file':
                addFileMessage(message);
                break;
            case 'ping':
                // Respond to ping with pong
                if (socket && socket.readyState === WebSocket.OPEN) {
                    socket.send(JSON.stringify({ type: 'pong' }));
                }
                break;
            default:
                console.warn('Unknown message type:', message.type);
        }
    } catch (error) {
        console.error('Error parsing message:', error, 'Raw data:', event.data);
    }
}

// Handle WebSocket close event
function handleSocketClose(event) {
    console.log('WebSocket connection closed:', event.code, event.reason);
    
    // Attempt to reconnect if not a normal closure
    if (event.code !== 1000 && reconnectAttempts < maxReconnectAttempts) {
        reconnectAttempts++;
        const delay = reconnectInterval * reconnectAttempts;
        
        console.log(`Attempting to reconnect (${reconnectAttempts}/${maxReconnectAttempts}) in ${delay}ms...`);
        addSystemMessage(`Connection lost. Reconnecting in ${delay / 1000} seconds...`);
        
        setTimeout(connectWebSocket, delay);
    } else if (reconnectAttempts >= maxReconnectAttempts) {
        addSystemMessage('Failed to reconnect after multiple attempts. Please refresh the page.');
    }
}

// Handle WebSocket error event
function handleSocketError(error) {
    console.error('WebSocket error:', error);
    addSystemMessage('Connection error. Please check your network connection.');
}

// Send a message
function sendMessage(event) {
    event.preventDefault();
    
    if (!messageInput) {
        console.error('Message input not found');
        return;
    }
    
    const content = messageInput.value.trim();
    if (!content) return;
    
    // Create message object matching the ClientMessage struct on the server
    const message = {
        type: 'message',
        content: content,
        sender: username
    };
    
    // Send message to server
    if (socket && socket.readyState === WebSocket.OPEN) {
        try {
            const jsonMessage = JSON.stringify(message);
            socket.send(jsonMessage);
            console.log('Sent message:', message);
        } catch (error) {
            console.error('Error sending message:', error);
            addSystemMessage('Error sending message. Please try again.');
        }
    } else {
        console.error('WebSocket not connected');
        addSystemMessage('Not connected to server. Please refresh the page.');
    }
    
    // Clear input
    messageInput.value = '';
    messageInput.focus();
}

// Handle file upload
function handleFileUpload() {
    if (!fileInput || !fileInput.files || fileInput.files.length === 0) {
        console.error('No file selected');
        return;
    }
    
    const file = fileInput.files[0];
    console.log('Selected file:', file.name, file.size, file.type);
    
    // Check file size (10MB max)
    const maxSize = 10 * 1024 * 1024;
    if (file.size > maxSize) {
        showError('File too large. Maximum size is 10MB.');
        fileInput.value = '';
        return;
    }
    
    // Create form data
    const formData = new FormData();
    formData.append('file', file);
    formData.append('username', username);
    
    // Show upload progress
    addSystemMessage(`Uploading file: ${file.name}...`);
    
    // Upload file
    fetch('/upload', {
        method: 'POST',
        body: formData
    })
    .then(response => {
        if (!response.ok) {
            throw new Error('Upload failed');
        }
        return response.json();
    })
    .then(data => {
        console.log('Upload successful:', data);
        addSystemMessage(`File uploaded: ${file.name}`);
    })
    .catch(error => {
        console.error('Upload error:', error);
        addSystemMessage(`Upload failed: ${error.message}`);
    })
    .finally(() => {
        // Reset file input
        fileInput.value = '';
    });
}

// Add a chat message to the UI
function addChatMessage(message) {
    if (!chatMessages) {
        console.error('Chat messages container not found');
        return;
    }
    
    const isCurrentUser = message.sender === username;
    
    // Create message element
    const messageElement = document.createElement('div');
    messageElement.className = `message ${isCurrentUser ? 'user-message' : ''}`;
    
    // Format timestamp
    let timestamp = '';
    if (message.sentAt) {
        const date = new Date(message.sentAt);
        timestamp = date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
    } else {
        const now = new Date();
        timestamp = now.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
    }
    
    // Create message HTML
    messageElement.innerHTML = `
        <div class="message-info">
            <span class="message-sender">${message.sender}</span>
            <span class="message-time">${timestamp}</span>
        </div>
        <div class="message-content">${message.content}</div>
    `;
    
    // Add message to chat
    chatMessages.appendChild(messageElement);
    
    // Scroll to bottom
    chatMessages.scrollTop = chatMessages.scrollHeight;
}

// Add a system message to the UI
function addSystemMessage(content) {
    if (!chatMessages) {
        console.error('Chat messages container not found');
        return;
    }
    
    // Create message element
    const messageElement = document.createElement('div');
    messageElement.className = 'message system-message';
    
    // Create message HTML
    messageElement.innerHTML = `
        <div class="message-content">${content}</div>
    `;
    
    // Add message to chat
    chatMessages.appendChild(messageElement);
    
    // Scroll to bottom
    chatMessages.scrollTop = chatMessages.scrollHeight;
}

// Add a file message to the UI
function addFileMessage(message) {
    if (!chatMessages) {
        console.error('Chat messages container not found');
        return;
    }
    
    const isCurrentUser = message.sender === username;
    
    // Create message element
    const messageElement = document.createElement('div');
    messageElement.className = `message file-message ${isCurrentUser ? 'user-message' : ''}`;
    
    // Format timestamp
    let timestamp = '';
    if (message.sentAt) {
        const date = new Date(message.sentAt);
        timestamp = date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
    } else {
        const now = new Date();
        timestamp = now.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
    }
    
    // Get file information
    const attachment = message.attachments && message.attachments.length > 0 ? message.attachments[0] : null;
    
    // Create file HTML
    let fileHTML = '';
    if (attachment) {
        const fileSize = formatFileSize(attachment.fileSize);
        const fileURL = `/uploads/${attachment.filePath}`;
        const fileIcon = getFileIcon(attachment.fileType);
        
        fileHTML = `
            <div class="file-attachment">
                <i class="${fileIcon}"></i>
                <div class="file-info">
                    <div class="file-name">${attachment.fileName}</div>
                    <div class="file-size">${fileSize}</div>
                </div>
                <a href="${fileURL}" target="_blank" class="file-download">
                    <i class="fas fa-download"></i>
                </a>
            </div>
        `;
    }
    
    // Create message HTML
    messageElement.innerHTML = `
        <div class="message-info">
            <span class="message-sender">${message.sender}</span>
            <span class="message-time">${timestamp}</span>
        </div>
        <div class="message-content">
            ${message.content}
            ${fileHTML}
        </div>
    `;
    
    // Add message to chat
    chatMessages.appendChild(messageElement);
    
    // Scroll to bottom
    chatMessages.scrollTop = chatMessages.scrollHeight;
}

// Update the user list
function updateUserList(users) {
    if (!userList) {
        console.error('User list container not found');
        return;
    }
    
    // Clear user list
    userList.innerHTML = '';
    
    // Add each user to the list
    users.forEach(user => {
        const userElement = document.createElement('div');
        userElement.className = 'user-item';
        userElement.innerHTML = `
            <i class="fas fa-circle"></i>
            <span>${user}</span>
        `;
        userList.appendChild(userElement);
    });
}

// Toggle emoji picker
function toggleEmojiPicker() {
    if (!emojiPicker) {
        console.error('Emoji picker not found');
        return;
    }
    
    const isVisible = emojiPicker.style.display === 'grid';
    emojiPicker.style.display = isVisible ? 'none' : 'grid';
}

// Handle document click to close emoji picker
function handleDocumentClick(event) {
    if (!emojiPicker || !emojiBtn) return;
    
    if (event.target !== emojiBtn && !emojiPicker.contains(event.target)) {
        emojiPicker.style.display = 'none';
    }
}

// Show error message
function showError(message) {
    console.error(message);
    alert(message);
}

// Format file size
function formatFileSize(bytes) {
    if (bytes < 1024) return bytes + ' B';
    if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB';
    return (bytes / (1024 * 1024)).toFixed(1) + ' MB';
}

// Get file icon based on file type
function getFileIcon(fileType) {
    if (fileType.startsWith('image/')) return 'fas fa-file-image';
    if (fileType.startsWith('video/')) return 'fas fa-file-video';
    if (fileType.startsWith('audio/')) return 'fas fa-file-audio';
    if (fileType.includes('pdf')) return 'fas fa-file-pdf';
    if (fileType.includes('word') || fileType.includes('document')) return 'fas fa-file-word';
    if (fileType.includes('excel') || fileType.includes('spreadsheet')) return 'fas fa-file-excel';
    if (fileType.includes('powerpoint') || fileType.includes('presentation')) return 'fas fa-file-powerpoint';
    if (fileType.includes('zip') || fileType.includes('compressed')) return 'fas fa-file-archive';
    if (fileType.includes('text')) return 'fas fa-file-alt';
    if (fileType.includes('code') || fileType.includes('javascript') || fileType.includes('html') || fileType.includes('css')) return 'fas fa-file-code';
    return 'fas fa-file';
}

// Create emoji picker
function createEmojiPicker() {
    if (!emojiPicker) {
        console.error('Emoji picker element not found');
        return;
    }
    
    const emojis = ['😀', '😁', '😂', '🤣', '😃', '😄', '😅', '😆', '😉', '😊', '😋', '😎', '😍', '😘', '🥰', '😗', '😙', '😚', '🙂', '🤗', '🤩', '🤔', '🤨', '😐', '😑', '😶', '🙄', '😏', '😣', '😥', '😮', '🤐', '😯', '😪', '😫', '😴', '😌', '😛', '😜', '😝', '🤤', '😒', '😓', '😔', '😕', '🙃', '🤑', '😲', '☹️', '🙁', '😖', '😞', '😟', '😤', '😢', '😭', '😦', '😧', '😨', '😩', '🤯', '😬', '😰', '😱', '🥵', '🥶', '😳', '🤪', '😵', '😡', '😠', '🤬', '😷', '🤒', '🤕', '🤢', '🤮', '🤧', '😇', '🥳', '🥴', '🥺', '🤠', '🤡', '🤥', '🤫', '🤭', '🧐', '🤓', '😈', '👿', '👹', '👺', '💀', '👻', '👽', '🤖', '💩', '😺', '😸', '😹', '😻', '😼', '😽', '🙀', '😿', '😾'];
    
    // Clear emoji picker
    emojiPicker.innerHTML = '';
    
    // Add emojis to picker
    emojis.forEach(emoji => {
        const emojiElement = document.createElement('div');
        emojiElement.className = 'emoji';
        emojiElement.textContent = emoji;
        
        // Add click event
        emojiElement.addEventListener('click', () => {
            if (messageInput) {
                messageInput.value += emoji;
                messageInput.focus();
            }
            emojiPicker.style.display = 'none';
        });
        
        // Add emoji to picker
        emojiPicker.appendChild(emojiElement);
    });
}

// Initialize the application
document.addEventListener('DOMContentLoaded', init);
