import React from 'react';
import { Navigate } from 'react-router-dom';
import { useAuth } from '../utils/AuthContext';

const PrivateRoute = ({ children }) => {
  const { currentUser, loading } = useAuth();
  
  // Show loading state while checking authentication
  if (loading) {
    return <div className="container">Loading...</div>;
  }
  
  // Redirect to login if not authenticated
  if (!currentUser) {
    return <Navigate to="/login" />;
  }
  
  // Render the protected component
  return children;
};

export default PrivateRoute;
