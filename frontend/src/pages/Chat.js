import React, { useState, useEffect, useRef } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../utils/AuthContext';
import { friendshipApi, messageApi } from '../services/api';
import websocketService from '../services/websocket';
import Navbar from '../components/Navbar';
import './Chat.css';

const Chat = () => {
  const { currentUser, token } = useAuth();
  const navigate = useNavigate();
  
  const [friends, setFriends] = useState([]);
  const [selectedFriend, setSelectedFriend] = useState(null);
  const [messages, setMessages] = useState([]);
  const [newMessage, setNewMessage] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [wsConnected, setWsConnected] = useState(false);
  
  const messagesEndRef = useRef(null);
  
  // Connect to WebSocket when component mounts
  useEffect(() => {
    if (token) {
      websocketService.connect(token);
      
      const connectionHandler = (connected) => {
        setWsConnected(connected);
      };
      
      const messageHandler = (data) => {
        // If we receive a message and it's from the selected friend, add it to the messages
        if (selectedFriend && data.sender_id === selectedFriend.id) {
          setMessages(prev => [...prev, data]);
          scrollToBottom();
        }
      };
      
      const unsubscribeConnection = websocketService.onConnectionChange(connectionHandler);
      const unsubscribeMessage = websocketService.onMessage(messageHandler);
      
      return () => {
        unsubscribeConnection();
        unsubscribeMessage();
        websocketService.disconnect();
      };
    }
  }, [token, selectedFriend]);
  
  // Load friends when component mounts
  useEffect(() => {
    loadFriends();
  }, []);
  
  // Load messages when a friend is selected
  useEffect(() => {
    if (selectedFriend) {
      loadMessages(selectedFriend.id);
    }
  }, [selectedFriend]);
  
  // Scroll to bottom when messages change
  useEffect(() => {
    scrollToBottom();
  }, [messages]);
  
  const loadFriends = async () => {
    try {
      setLoading(true);
      const response = await friendshipApi.getFriends('accepted');
      console.log('Friends loaded:', response.data);
      
      // Arkadaşlık verilerini işle ve kullanılabilir hale getir
      const processedFriends = response.data.map(friendship => {
        // Arkadaş bilgilerini friendship.friend'den al
        if (friendship.friend) {
          return {
            id: friendship.friend_id,
            username: friendship.friend.username,
            email: friendship.friend.email,
            friendship_id: friendship.id,
            status: friendship.status
          };
        }
        // Eğer friend bilgisi yoksa, boş bir nesne döndür
        return {
          id: friendship.friend_id,
          username: 'Unknown User',
          email: '',
          friendship_id: friendship.id,
          status: friendship.status
        };
      });
      
      setFriends(processedFriends);
      
      // Select first friend by default if available
      if (processedFriends.length > 0 && !selectedFriend) {
        setSelectedFriend(processedFriends[0]);
      }
    } catch (err) {
      setError('Failed to load friends');
      console.error(err);
    } finally {
      setLoading(false);
    }
  };
  
  const loadMessages = async (friendId) => {
    try {
      setLoading(true);
      const response = await messageApi.getMessagesWithUser(friendId);
      setMessages(response.data);
    } catch (err) {
      setError('Failed to load messages');
      console.error(err);
    } finally {
      setLoading(false);
    }
  };
  
  const handleSelectFriend = (friend) => {
    setSelectedFriend(friend);
  };
  
  const handleSendMessage = async (e) => {
    e.preventDefault();
    
    if (!newMessage.trim() || !selectedFriend) return;
    
    try {
      // Send message via WebSocket
      const message = {
        sender_id: currentUser.id,
        receiver_id: selectedFriend.id,
        content: newMessage,
        timestamp: new Date().toISOString()
      };
      
      websocketService.sendMessage(message);
      
      // Optimistically add message to UI
      setMessages(prev => [...prev, {
        ...message,
        id: `temp-${Date.now()}`,
        is_read: false
      }]);
      
      // Clear input
      setNewMessage('');
    } catch (err) {
      setError('Failed to send message');
      console.error(err);
    }
  };
  
  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  };
  
  return (
    <div className="chat-page">
      <Navbar />
      
      <div className="chat-container">
        <div className="friends-sidebar">
          <h3>Friends</h3>
          
          {loading && friends.length === 0 && <p>Loading friends...</p>}
          {error && <p className="error">{error}</p>}
          
          <ul className="friends-list">
            {friends.map(friend => (
              <li 
                key={friend.friendship_id || friend.id} 
                className={`friend-item ${selectedFriend?.id === friend.id ? 'active' : ''}`}
                onClick={() => handleSelectFriend(friend)}
              >
                <div className="friend-avatar">
                  {friend.username ? friend.username.charAt(0).toUpperCase() : '?'}
                </div>
                <div className="friend-info">
                  <span className="friend-name">{friend.username || 'Unknown User'}</span>
                </div>
              </li>
            ))}
          </ul>
          
          {friends.length === 0 && !loading && (
            <p className="no-friends">No friends yet. Add some friends to start chatting!</p>
          )}
          
          <button 
            className="btn btn-primary add-friend-btn"
            onClick={() => navigate('/friends')}
          >
            Manage Friends
          </button>
        </div>
        
        <div className="chat-area">
          {selectedFriend ? (
            <>
              <div className="chat-header">
                <h3>{selectedFriend.username || 'Unknown User'}</h3>
                <div className="connection-status">
                  {wsConnected ? (
                    <span className="status-connected">Connected</span>
                  ) : (
                    <span className="status-disconnected">Disconnected</span>
                  )}
                </div>
              </div>
              
              <div className="messages-container">
                {messages.length === 0 && !loading && (
                  <div className="no-messages">
                    No messages yet. Start the conversation!
                  </div>
                )}
                
                {loading && <div className="loading">Loading messages...</div>}
                
                <div className="messages-list">
                  {messages.map(message => (
                    <div 
                      key={message.id} 
                      className={`message ${message.sender_id === currentUser.id ? 'sent' : 'received'}`}
                    >
                      <div className="message-content">{message.content}</div>
                      <div className="message-time">
                        {new Date(message.timestamp).toLocaleTimeString()}
                      </div>
                    </div>
                  ))}
                  <div ref={messagesEndRef} />
                </div>
              </div>
              
              <form className="message-form" onSubmit={handleSendMessage}>
                <input
                  type="text"
                  className="message-input"
                  placeholder="Type a message..."
                  value={newMessage}
                  onChange={(e) => setNewMessage(e.target.value)}
                />
                <button 
                  type="submit" 
                  className="send-button"
                  disabled={!newMessage.trim()}
                >
                  Send
                </button>
              </form>
            </>
          ) : (
            <div className="no-chat-selected">
              <p>Select a friend to start chatting</p>
            </div>
          )}
        </div>
      </div>
    </div>
  );
};

export default Chat;
