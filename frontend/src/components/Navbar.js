import React from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { useAuth } from '../utils/AuthContext';
import './Navbar.css';

const Navbar = () => {
  const { currentUser, logout } = useAuth();
  const navigate = useNavigate();
  
  const handleLogout = () => {
    logout();
    navigate('/login');
  };
  
  return (
    <nav className="navbar">
      <div className="navbar-container">
        <Link to="/chat" className="navbar-logo">
          GoChat
        </Link>
        
        <div className="navbar-links">
          <Link to="/chat" className="nav-link">
            Chat
          </Link>
          <Link to="/friends" className="nav-link">
            Friends
          </Link>
          <Link to="/profile" className="nav-link">
            Profile
          </Link>
        </div>
        
        <div className="navbar-user">
          {currentUser && (
            <>
              <span className="user-greeting">
                Hello, {currentUser.username}
              </span>
              <button 
                className="btn btn-secondary logout-btn"
                onClick={handleLogout}
              >
                Logout
              </button>
            </>
          )}
        </div>
      </div>
    </nav>
  );
};

export default Navbar;
