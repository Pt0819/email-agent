import apiClient from './client';
import type { ApiResponse, GameRecommendation, RecommendationListResponse, RecStatsSummary, FeedbackRequest, GenerateRecommendationRequest } from './types';

export const recommendationApi = {
  /** 获取推荐列表 */
  list: (params?: {
    page?: number;
    page_size?: number;
    status?: string;
    deal_only?: boolean;
  }) =>
    apiClient.get<ApiResponse<RecommendationListResponse>>('/recommendations', { params }),

  /** 生成/刷新推荐 */
  generate: (data?: GenerateRecommendationRequest) =>
    apiClient.post<ApiResponse<RecommendationListResponse>>('/recommendations/generate', data || {}),

  /** 获取推荐详情 */
  getById: (id: number) =>
    apiClient.get<ApiResponse<GameRecommendation>>(`/recommendations/${id}`),

  /** 提交反馈 */
  submitFeedback: (id: number, data: FeedbackRequest) =>
    apiClient.post<ApiResponse<{ message: string }>>(`/recommendations/${id}/feedback`, data),

  /** 获取推荐统计 */
  getStats: () =>
    apiClient.get<ApiResponse<RecStatsSummary>>('/recommendations/stats'),
};
