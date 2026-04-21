import axios, { type AxiosInstance, type AxiosError } from 'axios';
import { API_BASE_URL, API_TIMEOUT, headers } from './config';
import type {
  ApiResponse,
  PageData,
  Email,
  EmailListParams,
  EmailAccount,
  CreateAccountRequest,
  ClassifyResponse,
  SyncStatusData,
  SyncResponse,
  SchedulerStatus,
  SetIntervalRequest,
  DailySummary,
} from './types';

// 创建axios实例
const apiClient: AxiosInstance = axios.create({
  baseURL: API_BASE_URL,
  timeout: API_TIMEOUT,
  headers,
});

// 请求拦截器
apiClient.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('token');
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => Promise.reject(error)
);

// 响应拦截器
apiClient.interceptors.response.use(
  (response) => response.data,
  (error: AxiosError<ApiResponse>) => {
    // 401 未授权 - 清除token并跳转登录页
    if (error.response?.status === 401) {
      localStorage.removeItem('token');
      localStorage.removeItem('user');
      // 如果不在登录页，则跳转
      if (!window.location.pathname.includes('/login')) {
        window.location.href = '/login';
      }
    }
    const message = error.response?.data?.message || error.message || '请求失败';
    console.error('API Error:', message);
    return Promise.reject(new Error(message));
  }
);

// ==================== 邮件API ====================

export const emailApi = {
  /** 获取邮件列表 */
  list: (params?: EmailListParams) =>
    apiClient.get<ApiResponse<PageData<Email>>>('/emails', { params }),

  /** 获取邮件详情 */
  getById: (id: number) =>
    apiClient.get<ApiResponse<Email>>(`/emails/${id}`),

  /** 分类邮件 */
  classify: (id: number) =>
    apiClient.post<ApiResponse<ClassifyResponse>>(`/emails/${id}/classify`),

  /** 更新邮件状态 */
  updateStatus: (id: number, status: string) =>
    apiClient.put<ApiResponse<{ id: number; status: string }>>(`/emails/${id}/status`, { status }),

  /** 删除邮件 */
  delete: (id: number) =>
    apiClient.delete<ApiResponse<null>>(`/emails/${id}`),
};

// ==================== 账户API ====================

export const accountApi = {
  /** 获取账户列表 */
  list: () =>
    apiClient.get<ApiResponse<{ list: EmailAccount[] }>>('/accounts'),

  /** 创建账户 */
  create: (data: CreateAccountRequest) =>
    apiClient.post<ApiResponse<EmailAccount>>('/accounts', data),

  /** 删除账户 */
  delete: (id: number) =>
    apiClient.delete<ApiResponse<null>>(`/accounts/${id}`),

  /** 测试账户连接 */
  test: (id: number) =>
    apiClient.post<ApiResponse<{ id: number; success: boolean; message: string }>>(`/accounts/${id}/test`),
};

// ==================== 同步API ====================

export const syncApi = {
  /** 触发同步 */
  trigger: (accountId?: number) =>
    apiClient.post<ApiResponse<SyncResponse>>('/sync', { account_id: accountId }),

  /** 获取同步状态 */
  status: () =>
    apiClient.get<ApiResponse<SyncStatusData>>('/sync/status'),

  // ========== 调度器控制 ==========

  /** 启动调度器 */
  startScheduler: () =>
    apiClient.post<ApiResponse<{ message: string }>>('/sync/scheduler/start'),

  /** 停止调度器 */
  stopScheduler: () =>
    apiClient.post<ApiResponse<{ message: string }>>('/sync/scheduler/stop'),

  /** 获取调度器状态 */
  schedulerStatus: () =>
    apiClient.get<ApiResponse<SchedulerStatus>>('/sync/scheduler/status'),

  /** 设置同步间隔 */
  setInterval: (data: SetIntervalRequest) =>
    apiClient.put<ApiResponse<{ message: string; interval: number }>>('/sync/scheduler/interval', data),
};

// ==================== 摘要API ====================

export const summaryApi = {
  /** 获取每日摘要 */
  daily: (date?: string) =>
    apiClient.get<ApiResponse<DailySummary>>('/summary/daily', { params: { date } }),
};

export default apiClient;
