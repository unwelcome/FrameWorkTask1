import { axiosInstance } from './api.ts';

export const authApi = {
  login: (credentials: { login: string; password: string }) => 
    axiosInstance.post<{id: number}>('/api/login', credentials),
  
  logout: () => 
    axiosInstance.post('/api/auth/logout'),
  
  refresh: () => 
    axiosInstance.post('/api/auth/refresh'),
  
  getUserInfo: () => 
    axiosInstance.get('/api/auth/user'),
};