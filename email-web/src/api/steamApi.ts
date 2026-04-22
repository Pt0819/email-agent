import apiClient from './client';
import type { ApiResponse, PageData, SteamGame, SteamDeal, SteamStats, SteamAccount, SteamLibraryItem, BindSteamRequest } from './types';

// ==================== Steam API ====================

export const steamApi = {
  // ==================== 账号绑定 ====================

  /** 绑定Steam账号 */
  bind: (data: BindSteamRequest) =>
    apiClient.post<ApiResponse<SteamAccount>>('/steam/bind', data),

  /** 获取Steam账号信息 */
  getProfile: () =>
    apiClient.get<ApiResponse<SteamAccount>>('/steam/profile'),

  /** 解绑Steam账号 */
  unbind: () =>
    apiClient.delete<ApiResponse<{ message: string }>>('/steam/unbind'),

  // ==================== 游戏库 ====================

  /** 获取游戏库列表 */
  listLibrary: (params?: { page?: number; page_size?: number; sort?: string }) =>
    apiClient.get<ApiResponse<PageData<SteamLibraryItem>>>('/steam/library', { params }),

  /** 获取最近游玩 */
  listRecent: (params?: { limit?: number }) =>
    apiClient.get<ApiResponse<{ list: SteamLibraryItem[]; total: number }>>('/steam/library/recent', { params }),

  /** 手动同步游戏库 */
  syncLibrary: () =>
    apiClient.post<ApiResponse<{ message: string }>>('/steam/sync'),

  // ==================== 游戏和促销 ====================

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
