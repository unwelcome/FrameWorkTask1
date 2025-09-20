export const LOG_REQUESTS = true;
export const LOG_TOKEN_RESRESH = true;

export const API_URL = import.meta.env.VUE_APP_API_BASE_URL || 'http://localhost:8080';
export const API_TIMEOUT = 10000;

export interface IValidator<T>{
  value: T,
  error: string,
}