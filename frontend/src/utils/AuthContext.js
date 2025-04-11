import React, { createContext, useState, useEffect, useContext } from 'react';
import axios from 'axios';

const AuthContext = createContext(null);

export const useAuth = () => useContext(AuthContext);

export const AuthProvider = ({ children }) => {
  const [currentUser, setCurrentUser] = useState(null);
  const [token, setToken] = useState(localStorage.getItem('token'));
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    // If token exists, fetch user data
    if (token) {
      fetchUserData();
    } else {
      setLoading(false);
    }
  }, [token]);

  const fetchUserData = async () => {
    try {
      setLoading(true);
      const response = await axios.get(`http://localhost:8000/api/users/me`, {
        headers: {
          Authorization: `Bearer ${token}`
        }
      });
      setCurrentUser(response.data);
      setLoading(false);
    } catch (error) {
      console.error('Error fetching user data:', error);
      logout();
      setLoading(false);
    }
  };

  const login = async (username, password) => {
    try {
      const formData = new FormData();
      formData.append('username', username);
      formData.append('password', password);

      const response = await axios.post(`http://localhost:8000/api/auth/login`, formData);
      const { access_token } = response.data;
      
      localStorage.setItem('token', access_token);
      setToken(access_token);
      
      return true;
    } catch (error) {
      console.error('Login error:', error);
      return false;
    }
  };

  const register = async (username, email, password) => {
    try {
      await axios.post(`http://localhost:8000/api/auth/register`, {
        username,
        email,
        password
      });
      return true;
    } catch (error) {
      console.error('Registration error:', error);
      return false;
    }
  };

  const logout = () => {
    localStorage.removeItem('token');
    setToken(null);
    setCurrentUser(null);
  };

  const updateProfile = async (userData) => {
    try {
      const response = await axios.put(
        `http://localhost:8000/api/users/me`, 
        userData,
        {
          headers: {
            Authorization: `Bearer ${token}`
          }
        }
      );
      setCurrentUser(response.data);
      return true;
    } catch (error) {
      console.error('Update profile error:', error);
      return false;
    }
  };

  const value = {
    currentUser,
    token,
    loading,
    login,
    register,
    logout,
    updateProfile
  };

  return (
    <AuthContext.Provider value={value}>
      {children}
    </AuthContext.Provider>
  );
};

export default AuthContext;
