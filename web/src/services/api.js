import axios from 'axios';
import { getAuthToken, removeAuthToken, cleanupStorage } from '../utils/storage';

const API_BASE = '/api';

const api = axios.create({
  baseURL: API_BASE,
  timeout: 30000,
});

// Adicionar interceptor para adicionar token à requisição
api.interceptors.request.use(
  (config) => {
    const token = getAuthToken();
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => Promise.reject(error)
);

// Adicionar interceptor para tratamento de erros
api.interceptors.response.use(
  (response) => response,
  (error) => {
    // Se erro 401, limpar token e redirecionar para login
    if (error.response?.status === 401) {
      removeAuthToken();
      window.location.href = '/?auth_error=unauthorized';
    }

    // Tentar limpar storage se houver erro de quota
    if (error.message?.includes('QuotaExceededError') ||
        error.message?.includes('kQuotaBytesPerItem')) {
      console.warn('Storage quota error, attempting cleanup...');
      cleanupStorage();
    }

    console.error('API Error:', error);
    return Promise.reject(error);
  }
);

export const apiService = {
  // Autenticação
  loginWithIMAP: async (email, password) => {
    const response = await api.post('/auth/login', { email, password });
    // Salvar token no localStorage
    if (response.data.token) {
      localStorage.setItem('auth_token', response.data.token);
    }
    return response.data;
  },

  login: () => {
    // Manter por compatibilidade (redireciona para /login)
    window.location.href = '/login';
  },

  logout: async () => {
    try {
      await api.post('/auth/logout');
    } catch (error) {
      console.error('Logout error:', error);
    } finally {
      removeAuthToken();
      window.location.href = '/';
    }
  },

  // Saúde
  health: async () => {
    const response = await api.get('/health');
    return response.data;
  },

  // Varredura
  startScan: async (folders = ['INBOX']) => {
    const response = await api.post('/scan', { folders });
    return response.data;
  },

  getScanStatus: async () => {
    const response = await api.get('/scan-status');
    return response.data;
  },

  getScanProgress: async () => {
    const response = await api.get('/scan-progress');
    return response.data;
  },

  cancelScan: async () => {
    const response = await api.post('/scan-cancel');
    return response.data;
  },

  getFolders: async () => {
    const response = await api.get('/folders');
    return response.data;
  },

  // Mensagens
  getMessages: async (page = 1, query = '') => {
    const params = new URLSearchParams();
    params.append('page', page);
    if (query) {
      params.append('q', query);
    }
    const response = await api.get(`/messages?${params.toString()}`);
    return response.data;
  },

  deleteMessage: async (messageId) => {
    const response = await api.delete(`/messages/${messageId}`);
    return response.data;
  },

  // Estatísticas
  getStats: async () => {
    const response = await api.get('/stats');
    return response.data;
  },
};

export default api;
