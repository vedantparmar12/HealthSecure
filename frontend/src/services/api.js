import axios from 'axios';

// Create axios instance with default config
const api = axios.create({
  baseURL: process.env.REACT_APP_API_URL || '/api',
  headers: {
    'Content-Type': 'application/json',
  },
});

// Token management
export const tokenManager = {
  getAccessToken: () => localStorage.getItem('accessToken'),
  getRefreshToken: () => localStorage.getItem('refreshToken'),
  setTokens: (accessToken, refreshToken) => {
    localStorage.setItem('accessToken', accessToken);
    if (refreshToken) {
      localStorage.setItem('refreshToken', refreshToken);
    }
  },
  clearTokens: () => {
    localStorage.removeItem('accessToken');
    localStorage.removeItem('refreshToken');
  }
};

// Request interceptor to add auth token
api.interceptors.request.use(
  (config) => {
    const token = tokenManager.getAccessToken();
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

// Response interceptor to handle token refresh
let isRefreshing = false;
let failedQueue = [];

const processQueue = (error, token = null) => {
  failedQueue.forEach(({ resolve, reject }) => {
    if (error) {
      reject(error);
    } else {
      resolve(token);
    }
  });
  
  failedQueue = [];
};

api.interceptors.response.use(
  (response) => response,
  async (error) => {
    const originalRequest = error.config;

    if (error.response?.status === 401 && !originalRequest._retry) {
      // Skip refresh for auth endpoints to prevent infinite loop
      if (originalRequest.url?.includes('/auth/')) {
        tokenManager.clearTokens();
        window.location.href = '/login';
        return Promise.reject(error);
      }

      if (isRefreshing) {
        // If already refreshing, queue this request
        return new Promise((resolve, reject) => {
          failedQueue.push({ resolve, reject });
        }).then(token => {
          originalRequest.headers.Authorization = `Bearer ${token}`;
          return api(originalRequest);
        }).catch(err => {
          return Promise.reject(err);
        });
      }

      originalRequest._retry = true;
      isRefreshing = true;

      try {
        const refreshToken = tokenManager.getRefreshToken();
        if (refreshToken) {
          const response = await api.post('/auth/refresh', {
            refresh_token: refreshToken
          });

          const { access_token, refresh_token: newRefreshToken } = response.data;
          tokenManager.setTokens(access_token, newRefreshToken);
          
          processQueue(null, access_token);
          
          // Retry original request
          originalRequest.headers.Authorization = `Bearer ${access_token}`;
          return api(originalRequest);
        } else {
          throw new Error('No refresh token available');
        }
      } catch (refreshError) {
        processQueue(refreshError, null);
        tokenManager.clearTokens();
        window.location.href = '/login';
        return Promise.reject(refreshError);
      } finally {
        isRefreshing = false;
      }
    }

    return Promise.reject(error);
  }
);

// Auth API
export const authAPI = {
  login: (credentials) => api.post('/auth/login', credentials),
  logout: () => api.post('/auth/logout'),
  getCurrentUser: () => api.get('/auth/me'),
  refreshToken: (refreshToken) => api.post('/auth/refresh', { refresh_token: refreshToken }),
  changePassword: (data) => api.post('/profile/change-password', data),
};

// Patients API
export const patientsAPI = {
  getPatients: (params = {}) => api.get('/patients', { params }),
  getPatient: (id, emergencyToken = null) => {
    const headers = emergencyToken ? { 'X-Emergency-Access-Token': emergencyToken } : {};
    return api.get(`/patients/${id}`, { headers });
  },
  createPatient: (data) => api.post('/patients', data),
  updatePatient: (id, data) => api.put(`/patients/${id}`, data),
  deletePatient: (id) => api.delete(`/patients/${id}`),
  searchPatients: (name) => api.get('/patients/search', { params: { name } }),
};

// Medical Records API
export const medicalRecordsAPI = {
  getPatientRecords: (patientId, params = {}, emergencyToken = null) => {
    const headers = emergencyToken ? { 'X-Emergency-Access-Token': emergencyToken } : {};
    return api.get(`/patients/${patientId}/records`, { params, headers });
  },
  getRecord: (id, emergencyToken = null) => {
    const headers = emergencyToken ? { 'X-Emergency-Access-Token': emergencyToken } : {};
    return api.get(`/records/${id}`, { headers });
  },
  createRecord: (patientId, data) => api.post(`/patients/${patientId}/records`, data),
  updateRecord: (id, data) => api.put(`/records/${id}`, data),
};

// Emergency Access API
export const emergencyAPI = {
  requestAccess: (data) => api.post('/emergency/request', data),
  activateAccess: (id) => api.post(`/emergency/activate/${id}`),
  revokeAccess: (id) => api.post(`/emergency/revoke/${id}`),
  getActiveAccess: () => api.get('/emergency/active'),
  getUserAccess: (userId, params = {}) => api.get(`/emergency/user/${userId}`, { params }),
};

// Audit API
export const auditAPI = {
  getLogs: (params = {}) => api.get('/audit/logs', { params }),
  getUserHistory: (userId, params = {}) => api.get(`/audit/users/${userId}`, { params }),
  getPatientHistory: (patientId, params = {}) => api.get(`/audit/patients/${patientId}`, { params }),
  getSecurityEvents: (params = {}) => api.get('/audit/security-events', { params }),
  resolveSecurityEvent: (id) => api.post(`/audit/security-events/${id}/resolve`),
  getStatistics: (params = {}) => api.get('/audit/statistics', { params }),
};

// Admin API
export const adminAPI = {
  getUsers: (params = {}) => api.get('/admin/users', { params }),
  getUser: (id) => api.get(`/admin/users/${id}`),
  createUser: (data) => api.post('/admin/users', data),
  updateUser: (id, data) => api.put(`/admin/users/${id}`, data),
  deactivateUser: (id) => api.post(`/admin/users/${id}/deactivate`),
  getUserSessions: (id) => api.get(`/admin/users/${id}/sessions`),
  getDashboardStats: () => api.get('/admin/dashboard/stats'),
};

// Chat/AI Assistant API
export const chatAPI = {
  // Send message to LangGraph AI
  sendMessage: (data) => api.post('/chat/langgraph', data),
  
  // Thread management
  createThread: (data) => api.post('/chat/threads', data),
  getThreads: (params = {}) => api.get('/chat/threads', { params }),
  getThreadMessages: (threadId, params = {}) => api.get(`/chat/threads/${threadId}/messages`, { params }),
  
  // Feedback and monitoring
  submitFeedback: (data) => api.post('/chat/feedback', data),
  getFeedbackStats: (params = {}) => api.get('/chat/feedback/stats', { params }),
};

// Error handler utility
export const handleAPIError = (error) => {
  if (error.response) {
    // Server responded with error status
    return error.response.data?.error || 'An error occurred';
  } else if (error.request) {
    // Request was made but no response received
    return 'Network error - please check your connection';
  } else {
    // Something else happened
    return error.message || 'An unexpected error occurred';
  }
};

export default api;