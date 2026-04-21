import apiClient from './client';
import type { ApiResponse, PageData, SteamGame, SteamDeal, SteamStats } from './types';

// ==================== Steam API ====================

export const steamApi = {
  /** 获取游戏列表 */
  listGames: (params?: { page?: number; page_size?: number; keyword?: string }) =>
    apiClient.get<ApiResponse<PageData<SteamGame>>>('/steam/games', { params }),

  /** 获取促销列表 */
  listDeals: (params?: { page?: number; page_size?: number; sort?: string; active?: string }) =>
    apiClient.get<ApiResponse<PageData<SteamDeal>>>('/steam/deals', { params }),

  /** 获取促销详情 */
  getDeal: (id: number) =>
    apiClient.get<ApiResponse<SteamDeal>>(`/steam/deals/${id}`),

  /** 获取Steam统计 */
  getStats: () =>
    apiClient.get<ApiResponse<SteamStats>>('/steam/stats'),

  /** 提取Steam邮件信息 */
  extractFromEmail: (emailId: number) =>
    apiClient.post<ApiResponse<{ message: string }>>(`/steam/emails/${emailId}/extract`),
};
