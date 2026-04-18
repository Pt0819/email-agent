import apiClient from './client';
import type { ApiResponse, AuthResponse, LoginRequest, RegisterRequest, User } from './types';

// ==================== 认证API ====================

export const authApi = {
  /** 用户注册 */
  register: (data: RegisterRequest) =>
    apiClient.post<ApiResponse<AuthResponse>>('/auth/register', data),

  /** 用户登录 */
  login: (data: LoginRequest) =>
    apiClient.post<ApiResponse<AuthResponse>>('/auth/login', data),

  /** 获取当前用户信息 */
  me: () =>
    apiClient.get<ApiResponse<User>>('/auth/me'),
};
