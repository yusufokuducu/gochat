import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../utils/AuthContext';
import { friendshipApi, userApi } from '../services/api';
import Navbar from '../components/Navbar';
import './Friends.css';

const Friends = () => {
  const { currentUser } = useAuth();
  const navigate = useNavigate();
  
  const [friends, setFriends] = useState([]);
  const [pendingRequests, setPendingRequests] = useState([]);
  const [sentRequests, setSentRequests] = useState([]);
  const [users, setUsers] = useState([]);
  const [searchTerm, setSearchTerm] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [successMessage, setSuccessMessage] = useState('');

  // Load friends, requests, and users when component mounts
  useEffect(() => {
    loadFriends();
    loadPendingRequests();
    loadSentRequests();
    loadUsers();
  }, []);

  // Load friends
  const loadFriends = async () => {
    try {
      setLoading(true);
      const response = await friendshipApi.getFriends('accepted');
      
      // Process friendship data
      const processedFriends = response.data.map(friendship => {
        return {
          id: friendship.friend_id,
          username: friendship.friend?.username || 'Unknown User',
          email: friendship.friend?.email || '',
          friendship_id: friendship.id,
          status: friendship.status
        };
      });
      
      setFriends(processedFriends);
    } catch (err) {
      setError('Failed to load friends');
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  // Load pending friend requests
  const loadPendingRequests = async () => {
    try {
      setLoading(true);
      const response = await friendshipApi.getFriends('pending');
      
      // Filter only incoming requests
      const incomingRequests = response.data.filter(
        req => req.friend_id === currentUser.id
      ).map(req => ({
        id: req.id,
        user_id: req.user_id,
        username: req.user?.username || 'Unknown User',
        email: req.user?.email || '',
        status: req.status,
        created_at: req.created_at
      }));
      
      setPendingRequests(incomingRequests);
    } catch (err) {
      console.error('Failed to load pending requests:', err);
    } finally {
      setLoading(false);
    }
  };

  // Load sent friend requests
  const loadSentRequests = async () => {
    try {
      setLoading(true);
      const response = await friendshipApi.getFriends('pending');
      
      // Filter only outgoing requests
      const outgoingRequests = response.data.filter(
        req => req.user_id === currentUser.id
      ).map(req => ({
        id: req.id,
        friend_id: req.friend_id,
        username: req.friend?.username || 'Unknown User',
        email: req.friend?.email || '',
        status: req.status,
        created_at: req.created_at
      }));
      
      setSentRequests(outgoingRequests);
    } catch (err) {
      console.error('Failed to load sent requests:', err);
    } finally {
      setLoading(false);
    }
  };

  // Load all users for friend suggestions
  const loadUsers = async () => {
    try {
      const response = await userApi.getUsers();
      
      // Filter out current user and existing friends/requests
      const filteredUsers = response.data.filter(user => 
        user.id !== currentUser.id &&
        !friends.some(friend => friend.id === user.id) &&
        !pendingRequests.some(req => req.user_id === user.id) &&
        !sentRequests.some(req => req.friend_id === user.id)
      );
      
      setUsers(filteredUsers);
    } catch (err) {
      console.error('Failed to load users:', err);
    }
  };

  // Send friend request
  const sendFriendRequest = async (userId) => {
    try {
      await friendshipApi.addFriend(userId);
      
      // Update UI
      setSuccessMessage('Friend request sent successfully!');
      setTimeout(() => setSuccessMessage(''), 3000);
      
      // Reload data
      loadSentRequests();
      loadUsers();
    } catch (err) {
      setError('Failed to send friend request');
      console.error(err);
      setTimeout(() => setError(''), 3000);
    }
  };

  // Accept friend request
  const acceptFriendRequest = async (requestId) => {
    try {
      await friendshipApi.updateFriendship(requestId, 'accepted');
      
      // Update UI
      setSuccessMessage('Friend request accepted!');
      setTimeout(() => setSuccessMessage(''), 3000);
      
      // Reload data
      loadFriends();
      loadPendingRequests();
    } catch (err) {
      setError('Failed to accept friend request');
      console.error(err);
      setTimeout(() => setError(''), 3000);
    }
  };

  // Reject friend request
  const rejectFriendRequest = async (requestId) => {
    try {
      await friendshipApi.updateFriendship(requestId, 'rejected');
      
      // Update UI
      setSuccessMessage('Friend request rejected');
      setTimeout(() => setSuccessMessage(''), 3000);
      
      // Reload data
      loadPendingRequests();
    } catch (err) {
      setError('Failed to reject friend request');
      console.error(err);
      setTimeout(() => setError(''), 3000);
    }
  };

  // Remove friend
  const removeFriend = async (friendshipId) => {
    try {
      await friendshipApi.removeFriend(friendshipId);
      
      // Update UI
      setSuccessMessage('Friend removed successfully');
      setTimeout(() => setSuccessMessage(''), 3000);
      
      // Reload data
      loadFriends();
      loadUsers();
    } catch (err) {
      setError('Failed to remove friend');
      console.error(err);
      setTimeout(() => setError(''), 3000);
    }
  };

  // Cancel sent friend request
  const cancelFriendRequest = async (requestId) => {
    try {
      await friendshipApi.removeFriend(requestId);
      
      // Update UI
      setSuccessMessage('Friend request cancelled');
      setTimeout(() => setSuccessMessage(''), 3000);
      
      // Reload data
      loadSentRequests();
      loadUsers();
    } catch (err) {
      setError('Failed to cancel friend request');
      console.error(err);
      setTimeout(() => setError(''), 3000);
    }
  };

  // Filter users based on search term
  const filteredUsers = users.filter(user => 
    user.username.toLowerCase().includes(searchTerm.toLowerCase()) ||
    user.email.toLowerCase().includes(searchTerm.toLowerCase())
  );

  return (
    <div className="friends-page">
      <Navbar />
      
      <div className="friends-container">
        <h2>Manage Friends</h2>
        
        {successMessage && (
          <div className="alert alert-success">{successMessage}</div>
        )}
        
        {error && (
          <div className="alert alert-danger">{error}</div>
        )}
        
        <div className="friends-sections">
          {/* Current Friends Section */}
          <div className="friends-section">
            <h3>Your Friends ({friends.length})</h3>
            {loading && friends.length === 0 ? (
              <p>Loading friends...</p>
            ) : friends.length === 0 ? (
              <p>You don't have any friends yet.</p>
            ) : (
              <ul className="friends-list">
                {friends.map(friend => (
                  <li key={friend.friendship_id} className="friend-item">
                    <div className="friend-avatar">
                      {friend.username.charAt(0).toUpperCase()}
                    </div>
                    <div className="friend-info">
                      <span className="friend-name">{friend.username}</span>
                      <span className="friend-email">{friend.email}</span>
                    </div>
                    <div className="friend-actions">
                      <button 
                        className="btn btn-danger btn-sm"
                        onClick={() => removeFriend(friend.friendship_id)}
                      >
                        Remove
                      </button>
                      <button 
                        className="btn btn-primary btn-sm"
                        onClick={() => navigate(`/chat?friend=${friend.id}`)}
                      >
                        Chat
                      </button>
                    </div>
                  </li>
                ))}
              </ul>
            )}
          </div>
          
          {/* Pending Requests Section */}
          {pendingRequests.length > 0 && (
            <div className="friends-section">
              <h3>Friend Requests ({pendingRequests.length})</h3>
              <ul className="friends-list">
                {pendingRequests.map(request => (
                  <li key={request.id} className="friend-item">
                    <div className="friend-avatar">
                      {request.username.charAt(0).toUpperCase()}
                    </div>
                    <div className="friend-info">
                      <span className="friend-name">{request.username}</span>
                      <span className="friend-email">{request.email}</span>
                    </div>
                    <div className="friend-actions">
                      <button 
                        className="btn btn-success btn-sm"
                        onClick={() => acceptFriendRequest(request.id)}
                      >
                        Accept
                      </button>
                      <button 
                        className="btn btn-danger btn-sm"
                        onClick={() => rejectFriendRequest(request.id)}
                      >
                        Reject
                      </button>
                    </div>
                  </li>
                ))}
              </ul>
            </div>
          )}
          
          {/* Sent Requests Section */}
          {sentRequests.length > 0 && (
            <div className="friends-section">
              <h3>Sent Requests ({sentRequests.length})</h3>
              <ul className="friends-list">
                {sentRequests.map(request => (
                  <li key={request.id} className="friend-item">
                    <div className="friend-avatar">
                      {request.username.charAt(0).toUpperCase()}
                    </div>
                    <div className="friend-info">
                      <span className="friend-name">{request.username}</span>
                      <span className="friend-email">{request.email}</span>
                      <span className="request-status">Pending</span>
                    </div>
                    <div className="friend-actions">
                      <button 
                        className="btn btn-warning btn-sm"
                        onClick={() => cancelFriendRequest(request.id)}
                      >
                        Cancel
                      </button>
                    </div>
                  </li>
                ))}
              </ul>
            </div>
          )}
          
          {/* Find Friends Section */}
          <div className="friends-section">
            <h3>Find Friends</h3>
            <div className="search-box">
              <input
                type="text"
                placeholder="Search by username or email"
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
                className="form-control"
              />
            </div>
            
            {filteredUsers.length === 0 ? (
              <p>No users found.</p>
            ) : (
              <ul className="friends-list">
                {filteredUsers.map(user => (
                  <li key={user.id} className="friend-item">
                    <div className="friend-avatar">
                      {user.username.charAt(0).toUpperCase()}
                    </div>
                    <div className="friend-info">
                      <span className="friend-name">{user.username}</span>
                      <span className="friend-email">{user.email}</span>
                    </div>
                    <div className="friend-actions">
                      <button 
                        className="btn btn-primary btn-sm"
                        onClick={() => sendFriendRequest(user.id)}
                      >
                        Add Friend
                      </button>
                    </div>
                  </li>
                ))}
              </ul>
            )}
          </div>
        </div>
        
        <button 
          className="btn btn-secondary back-btn"
          onClick={() => navigate('/chat')}
        >
          Back to Chat
        </button>
      </div>
    </div>
  );
};

export default Friends;
