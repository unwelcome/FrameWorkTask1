import { API_TIMEOUT, API_URL, LOG_REQUESTS, LOG_TOKEN_RESRESH } from '@/helpers/constants';
import axios, { type AxiosInstance, type AxiosResponse } from 'axios';

// Конфигурация axios
const createAxiosInstance = (): AxiosInstance => {
  const instance = axios.create({
    baseURL: API_URL,
    timeout: API_TIMEOUT,
    headers: {
      'Content-Type': 'application/json',
    },
    withCredentials: true,
  });

  // Response interceptor - обработка 403 ошибок
  instance.interceptors.response.use(
    (response: AxiosResponse) => {
      if (LOG_REQUESTS) console.log(`[API Response] ${response.status} ${response.config.url}`, {
        data: response.data,
        status: response.status,
        headers: response.headers
      });

      return response;
    },
    async (error) => {
      const originalRequest = error.config;

      if (LOG_REQUESTS) console.error(`[API Error] ${error.response?.status || 'No Status'} ${error.config?.url}`, {
        error: error.response?.data,
        status: error.response?.status,
        headers: error.response?.headers
      });

      // Если ошибка 403 и это не запрос обновления токена
      if (error.response?.status === 403 && !originalRequest._retry) {
        originalRequest._retry = true;

        try {
          if (LOG_TOKEN_RESRESH) console.log('Attempting token refresh...');

          // Отправляем запрос на обновление токена
          await axios.post(`${instance.defaults.baseURL}/api/auth/refresh`, {}, { withCredentials: true });

          if (LOG_TOKEN_RESRESH) console.log('Success, retrying original request...');
          // Повторяем запрос
          return instance(originalRequest);

        } catch (refreshError) {
          
          // Если не удалось обновить токен, очищаем хранилище и редиректим на логин
          if (LOG_TOKEN_RESRESH) console.error('Token Refresh Failed', refreshError);
          window.location.href = '/login';
          return Promise.reject(refreshError);
        }
      }

      return Promise.reject(error);
    }
  );

  return instance;
};

// Создаем экземпляр axios с middleware
export const axiosInstance = createAxiosInstance();

// Экспортируем все API
export { authApi } from './auth.api';
