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

  /** 更新用户资料 */
  updateProfile: (data: { username: string }) =>
    apiClient.put<ApiResponse<User>>('/auth/profile', data),

  /** 修改密码 */
  changePassword: (data: { old_password: string; new_password: string; confirm_password: string }) =>
    apiClient.put<ApiResponse<{ message: string }>>('/auth/password', data),

  /** 上传头像 */
  uploadAvatar: (file: File) => {
    const formData = new FormData();
    formData.append('avatar', file);
    return apiClient.post<ApiResponse<User>>('/auth/avatar', formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
    });
  },
};
