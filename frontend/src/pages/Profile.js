import React, { useState } from 'react';
import { useAuth } from '../utils/AuthContext';
import Navbar from '../components/Navbar';
import './Profile.css';

const Profile = () => {
  const { currentUser, updateProfile } = useAuth();
  
  const [username, setUsername] = useState(currentUser?.username || '');
  const [email, setEmail] = useState(currentUser?.email || '');
  const [password, setPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');
  const [loading, setLoading] = useState(false);
  
  const handleSubmit = async (e) => {
    e.preventDefault();
    
    // Reset messages
    setError('');
    setSuccess('');
    
    // Validation
    if (password && password !== confirmPassword) {
      setError('Passwords do not match');
      return;
    }
    
    if (password && password.length < 6) {
      setError('Password must be at least 6 characters long');
      return;
    }
    
    try {
      setLoading(true);
      
      // Only include fields that have changed
      const updateData = {};
      if (username !== currentUser.username) updateData.username = username;
      if (email !== currentUser.email) updateData.email = email;
      if (password) updateData.password = password;
      
      // Only make API call if there are changes
      if (Object.keys(updateData).length > 0) {
        const success = await updateProfile(updateData);
        
        if (success) {
          setSuccess('Profile updated successfully');
          setPassword('');
          setConfirmPassword('');
        } else {
          setError('Failed to update profile');
        }
      } else {
        setSuccess('No changes to save');
      }
    } catch (err) {
      setError('An error occurred while updating profile');
      console.error(err);
    } finally {
      setLoading(false);
    }
  };
  
  return (
    <div className="profile-page">
      <Navbar />
      
      <div className="container">
        <div className="profile-card">
          <h2>My Profile</h2>
          
          {error && <div className="alert alert-danger">{error}</div>}
          {success && <div className="alert alert-success">{success}</div>}
          
          <form onSubmit={handleSubmit}>
            <div className="form-group">
              <label htmlFor="username">Username</label>
              <input
                type="text"
                id="username"
                className="form-control"
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                disabled={loading}
              />
            </div>
            
            <div className="form-group">
              <label htmlFor="email">Email</label>
              <input
                type="email"
                id="email"
                className="form-control"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                disabled={loading}
              />
            </div>
            
            <div className="form-group">
              <label htmlFor="password">New Password (leave blank to keep current)</label>
              <input
                type="password"
                id="password"
                className="form-control"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                disabled={loading}
              />
            </div>
            
            <div className="form-group">
              <label htmlFor="confirmPassword">Confirm New Password</label>
              <input
                type="password"
                id="confirmPassword"
                className="form-control"
                value={confirmPassword}
                onChange={(e) => setConfirmPassword(e.target.value)}
                disabled={loading}
              />
            </div>
            
            <div className="profile-info">
              <p><strong>Member since:</strong> {new Date(currentUser?.created_at).toLocaleDateString()}</p>
              <p><strong>User ID:</strong> {currentUser?.id}</p>
            </div>
            
            <button 
              type="submit" 
              className="btn btn-primary save-btn" 
              disabled={loading}
            >
              {loading ? 'Saving...' : 'Save Changes'}
            </button>
          </form>
        </div>
      </div>
    </div>
  );
};

export default Profile;
