// API客户端配置
export const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api/v1';
export const AGENT_API_URL = import.meta.env.VITE_AGENT_API_URL || 'http://localhost:8001';

// API超时配置
export const API_TIMEOUT = 30000;

// 请求头配置
export const headers = {
  'Content-Type': 'application/json',
};