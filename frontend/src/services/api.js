import axios from 'axios';

const API_URL = process.env.REACT_APP_API_URL || 'http://localhost:8000/api';

// Create axios instance with base URL
const api = axios.create({
  baseURL: API_URL,
});

// Add request interceptor to add auth token to all requests
api.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('token');
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => Promise.reject(error)
);

// User API
export const userApi = {
  getCurrentUser: () => api.get('/users/me'),
  updateProfile: (userData) => api.put('/users/me', userData),
  getUser: (userId) => api.get(`/users/${userId}`),
  getUsers: (params) => api.get('/users', { params }),
};

// Auth API
export const authApi = {
  login: (username, password) => {
    const formData = new FormData();
    formData.append('username', username);
    formData.append('password', password);
    return api.post('/auth/login', formData);
  },
  register: (userData) => api.post('/auth/register', userData),
};

// Friendship API
export const friendshipApi = {
  getFriends: (status) => api.get('/friendships', { params: { status } }),
  addFriend: (friendId) => api.post('/friendships', { friend_id: friendId }),
  updateFriendship: (friendshipId, status) => api.put(`/friendships/${friendshipId}`, { status }),
  removeFriend: (friendshipId) => api.delete(`/friendships/${friendshipId}`),
};

// Message API
export const messageApi = {
  getMessagesWithUser: (userId, skip, limit) => 
    api.get(`/messages/with/${userId}`, { params: { skip, limit } }),
  getUnreadMessages: () => api.get('/messages/unread'),
  sendMessage: (messageData) => api.post('/messages', messageData),
  markAsRead: (messageId) => api.put(`/messages/${messageId}/read`),
};

export default {
  userApi,
  authApi,
  friendshipApi,
  messageApi,
};
