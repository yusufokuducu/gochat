import React from 'react';
import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import './App.css';

// Pages
import Login from './pages/Login';
import Register from './pages/Register';
import Chat from './pages/Chat';
import Profile from './pages/Profile';
import FriendsList from './pages/FriendsList';

// Context
import { AuthProvider } from './utils/AuthContext';
import PrivateRoute from './components/PrivateRoute';

function App() {
  return (
    <AuthProvider>
      <Router>
        <div className="App">
          <Routes>
            <Route path="/login" element={<Login />} />
            <Route path="/register" element={<Register />} />
            <Route 
              path="/chat" 
              element={
                <PrivateRoute>
                  <Chat />
                </PrivateRoute>
              } 
            />
            <Route 
              path="/profile" 
              element={
                <PrivateRoute>
                  <Profile />
                </PrivateRoute>
              } 
            />
            <Route 
              path="/friends" 
              element={
                <PrivateRoute>
                  <FriendsList />
                </PrivateRoute>
              } 
            />
            <Route path="/" element={<Navigate to="/login" replace />} />
          </Routes>
        </div>
      </Router>
    </AuthProvider>
  );
}

export default App;
