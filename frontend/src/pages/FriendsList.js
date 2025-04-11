import React, { useState, useEffect } from 'react';
import { useAuth } from '../utils/AuthContext';
import { friendshipApi, userApi } from '../services/api';
import Navbar from '../components/Navbar';
import './FriendsList.css';

const FriendsList = () => {
  const { currentUser } = useAuth();
  
  const [friends, setFriends] = useState([]);
  const [friendRequests, setFriendRequests] = useState([]);
  const [searchTerm, setSearchTerm] = useState('');
  const [searchResults, setSearchResults] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');
  
  // Load friends and friend requests when component mounts
  useEffect(() => {
    loadFriends();
    loadFriendRequests();
  }, []);
  
  const loadFriends = async () => {
    try {
      setLoading(true);
      const response = await friendshipApi.getFriends('accepted');
      console.log('Friends data:', response.data);
      setFriends(response.data);
    } catch (err) {
      setError('Failed to load friends');
      console.error(err);
    } finally {
      setLoading(false);
    }
  };
  
  const loadFriendRequests = async () => {
    try {
      setLoading(true);
      const response = await friendshipApi.getFriendRequests();
      console.log('Friend requests data:', response.data);
      setFriendRequests(response.data);
    } catch (err) {
      setError('Failed to load friend requests');
      console.error(err);
    } finally {
      setLoading(false);
    }
  };
  
  const handleSearch = async () => {
    if (!searchTerm.trim()) return;
    
    try {
      setLoading(true);
      setError('');
      
      const response = await userApi.getUsers({ search: searchTerm });
      console.log('Search results:', response.data);
      
      // Filter out current user and existing friends
      const filteredResults = response.data.filter(user => 
        user.id !== currentUser.id && 
        !friends.some(friend => friend.friend_id === user.id)
      );
      
      setSearchResults(filteredResults);
    } catch (err) {
      setError('Failed to search for users');
      console.error(err);
    } finally {
      setLoading(false);
    }
  };
  
  const handleAddFriend = async (userId) => {
    try {
      setLoading(true);
      await friendshipApi.addFriend(userId);
      
      // Remove from search results
      setSearchResults(prev => prev.filter(user => user.id !== userId));
      
      setSuccess('Friend request sent successfully');
      
      // Clear success message after 3 seconds
      setTimeout(() => setSuccess(''), 3000);
    } catch (err) {
      setError('Failed to send friend request');
      console.error(err);
    } finally {
      setLoading(false);
    }
  };
  
  const handleRespondToRequest = async (requestId, status) => {
    try {
      setLoading(true);
      console.log(`Responding to request ${requestId} with status ${status}`);
      await friendshipApi.respondToRequest(requestId, status);
      
      // Refresh lists
      loadFriendRequests();
      if (status === 'accepted') {
        loadFriends();
      }
      
      setSuccess(`Friend request ${status}`);
      
      // Clear success message after 3 seconds
      setTimeout(() => setSuccess(''), 3000);
    } catch (err) {
      setError(`Failed to ${status} friend request`);
      console.error(err);
    } finally {
      setLoading(false);
    }
  };
  
  const handleRemoveFriend = async (friendId) => {
    try {
      setLoading(true);
      await friendshipApi.removeFriend(friendId);
      
      // Refresh friends list
      loadFriends();
      
      setSuccess('Friend removed successfully');
      
      // Clear success message after 3 seconds
      setTimeout(() => setSuccess(''), 3000);
    } catch (err) {
      setError('Failed to remove friend');
      console.error(err);
    } finally {
      setLoading(false);
    }
  };
  
  return (
    <div className="friends-page">
      <Navbar />
      
      <div className="container">
        <h2>Friends Management</h2>
        
        {error && <div className="alert alert-danger">{error}</div>}
        {success && <div className="alert alert-success">{success}</div>}
        
        <div className="friends-section">
          <h3>Find New Friends</h3>
          
          <div className="search-form">
            <input
              type="text"
              className="form-control search-input"
              placeholder="Search by username or email"
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              disabled={loading}
            />
            <button 
              className="btn btn-primary search-btn"
              onClick={handleSearch}
              disabled={loading || !searchTerm.trim()}
            >
              Search
            </button>
          </div>
          
          {searchResults.length > 0 && (
            <div className="search-results">
              <h4>Search Results</h4>
              <ul className="user-list">
                {searchResults.map(user => (
                  <li key={user.id} className="user-item">
                    <div className="user-info">
                      <div className="user-avatar">
                        {user.username.charAt(0).toUpperCase()}
                      </div>
                      <div>
                        <div className="user-name">{user.username}</div>
                        <div className="user-email">{user.email}</div>
                      </div>
                    </div>
                    <button
                      className="btn btn-primary btn-sm"
                      onClick={() => handleAddFriend(user.id)}
                      disabled={loading}
                    >
                      Add Friend
                    </button>
                  </li>
                ))}
              </ul>
            </div>
          )}
        </div>
        
        <div className="friends-section">
          <h3>Friend Requests ({friendRequests.length})</h3>
          
          {friendRequests.length === 0 ? (
            <p className="no-items">No pending friend requests</p>
          ) : (
            <ul className="user-list">
              {friendRequests.map(request => (
                <li key={request.id} className="user-item">
                  <div className="user-info">
                    <div className="user-avatar">
                      {request.user && request.user.username ? request.user.username.charAt(0).toUpperCase() : '?'}
                    </div>
                    <div>
                      <div className="user-name">{request.user ? request.user.username : 'Unknown User'}</div>
                      <div className="request-time">
                        Sent a friend request
                      </div>
                    </div>
                  </div>
                  <div className="request-actions">
                    <button
                      className="btn btn-primary btn-sm"
                      onClick={() => handleRespondToRequest(request.id, 'accepted')}
                      disabled={loading}
                    >
                      Accept
                    </button>
                    <button
                      className="btn btn-secondary btn-sm"
                      onClick={() => handleRespondToRequest(request.id, 'rejected')}
                      disabled={loading}
                    >
                      Decline
                    </button>
                  </div>
                </li>
              ))}
            </ul>
          )}
        </div>
        
        <div className="friends-section">
          <h3>My Friends ({friends.length})</h3>
          
          {friends.length === 0 ? (
            <p className="no-items">You don't have any friends yet</p>
          ) : (
            <ul className="user-list">
              {friends.map(friend => (
                <li key={friend.id} className="user-item">
                  <div className="user-info">
                    <div className="user-avatar">
                      {friend.friend && friend.friend.username ? friend.friend.username.charAt(0).toUpperCase() : '?'}
                    </div>
                    <div>
                      <div className="user-name">{friend.friend ? friend.friend.username : 'Unknown User'}</div>
                      <div className="friend-since">
                        Friends
                      </div>
                    </div>
                  </div>
                  <button
                    className="btn btn-danger btn-sm"
                    onClick={() => handleRemoveFriend(friend.friend_id)}
                    disabled={loading}
                  >
                    Remove
                  </button>
                </li>
              ))}
            </ul>
          )}
        </div>
      </div>
    </div>
  );
};

export default FriendsList;
