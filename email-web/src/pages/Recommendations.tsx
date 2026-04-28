import { useState, useEffect } from 'react';
import { RefreshCw, Star, ThumbsUp, ThumbsDown, ExternalLink, Clock, TrendingUp, Filter, Loader2 } from 'lucide-react';
import { recommendationApi } from '../api/recommendationApi';
import type { GameRecommendation, RecStatsSummary } from '../api/types';

// 匹配度颜色
function getScoreColor(score: number): string {
  if (score >= 80) return 'text-emerald-600';
  if (score >= 60) return 'text-blue-600';
  if (score >= 40) return 'text-yellow-600';
  return 'text-gray-600';
}

// 匹配度背景色
function getScoreBgColor(score: number): string {
  if (score >= 80) return 'bg-emerald-100';
  if (score >= 60) return 'bg-blue-100';
  if (score >= 40) return 'bg-yellow-100';
  return 'bg-gray-100';
}

// 格式化游玩时长
function formatPlaytime(minutes: number): string {
  if (minutes < 60) return `${minutes}分钟`;
  const hours = Math.floor(minutes / 60);
  const mins = minutes % 60;
  if (hours < 24) return `${hours}小时${mins > 0 ? mins + '分钟' : ''}`;
  const days = Math.floor(hours / 24);
  return `${days}天${hours % 24 > 0 ? (hours % 24) + '小时' : ''}`;
}

export default function Recommendations() {
  const [recommendations, setRecommendations] = useState<GameRecommendation[]>([]);
  const [stats, setStats] = useState<RecStatsSummary | null>(null);
  const [loading, setLoading] = useState(true);
  const [generating, setGenerating] = useState(false);
  const [filter, setFilter] = useState<'all' | 'deals'>('all');
  const [expandedId, setExpandedId] = useState<number | null>(null);
  const [feedbackLoading, setFeedbackLoading] = useState<number | null>(null);

  const fetchRecommendations = async () => {
    setLoading(true);
    try {
      const res = await recommendationApi.list({
        page: 1,
        page_size: 50,
        status: 'all',
        deal_only: filter === 'deals',
      });
      if (res.data?.data) {
        setRecommendations(res.data.data.list);
        setStats(res.data.data.stats || null);
      }
    } catch (err) {
      console.error('获取推荐失败:', err);
    } finally {
      setLoading(false);
    }
  };

  const generateRecommendations = async () => {
    setGenerating(true);
    try {
      const res = await recommendationApi.generate({
        max_count: 20,
        deal_only: filter === 'deals',
        min_score: 40,
      });
      if (res.data?.data) {
        setRecommendations(res.data.data.list);
        setStats(res.data.data.stats || null);
      }
    } catch (err) {
      console.error('生成推荐失败:', err);
    } finally {
      setGenerating(false);
    }
  };

  const submitFeedback = async (id: number, action: 'like' | 'dislike' | 'click') => {
    setFeedbackLoading(id);
    try {
      await recommendationApi.submitFeedback(id, { action });
      // 乐观更新UI
      if (action === 'like' || action === 'dislike') {
        setRecommendations(prev =>
          prev.map(rec => rec.id === id ? { ...rec, status: action === 'like' ? 'clicked' : rec.status } : rec)
        );
      }
    } catch (err) {
      console.error('提交反馈失败:', err);
    } finally {
      setFeedbackLoading(null);
    }
  };

  useEffect(() => {
    fetchRecommendations();
  }, [filter]);

  return (
    <div className="p-6 max-w-7xl mx-auto">
      {/* 页面头部 */}
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">个性化推荐</h1>
          <p className="text-sm text-gray-500 mt-1">基于你的游戏偏好和Steam促销推荐</p>
        </div>
        <button
          onClick={generateRecommendations}
          disabled={generating}
          className="flex items-center gap-2 px-4 py-2 bg-gradient-to-r from-blue-600 to-purple-600 text-white rounded-lg hover:from-blue-700 hover:to-purple-700 transition-all disabled:opacity-50"
        >
          {generating ? (
            <Loader2 className="w-4 h-4 animate-spin" />
          ) : (
            <RefreshCw className="w-4 h-4" />
          )}
          {generating ? '生成中...' : '刷新推荐'}
        </button>
      </div>

      {/* 统计卡片 */}
      {stats && (
        <div className="grid grid-cols-4 gap-4 mb-6">
          <div className="bg-white rounded-lg shadow-sm p-4 border border-gray-100">
            <div className="flex items-center gap-3">
              <div className="p-2 bg-blue-100 rounded-lg">
                <TrendingUp className="w-5 h-5 text-blue-600" />
              </div>
              <div>
                <p className="text-2xl font-bold text-gray-900">{stats.total_recommendations}</p>
                <p className="text-xs text-gray-500">推荐总数</p>
              </div>
            </div>
          </div>
          <div className="bg-white rounded-lg shadow-sm p-4 border border-gray-100">
            <div className="flex items-center gap-3">
              <div className="p-2 bg-emerald-100 rounded-lg">
                <ThumbsUp className="w-5 h-5 text-emerald-600" />
              </div>
              <div>
                <p className="text-2xl font-bold text-gray-900">{stats.clicked_count}</p>
                <p className="text-xs text-gray-500">点击次数</p>
              </div>
            </div>
          </div>
          <div className="bg-white rounded-lg shadow-sm p-4 border border-gray-100">
            <div className="flex items-center gap-3">
              <div className="p-2 bg-purple-100 rounded-lg">
                <Star className="w-5 h-5 text-purple-600" />
              </div>
              <div>
                <p className="text-2xl font-bold text-gray-900">{stats.purchase_count}</p>
                <p className="text-xs text-gray-500">购买数</p>
              </div>
            </div>
          </div>
          <div className="bg-white rounded-lg shadow-sm p-4 border border-gray-100">
            <div className="flex items-center gap-3">
              <div className="p-2 bg-amber-100 rounded-lg">
                <TrendingUp className="w-5 h-5 text-amber-600" />
              </div>
              <div>
                <p className="text-2xl font-bold text-gray-900">{stats.ctr.toFixed(1)}%</p>
                <p className="text-xs text-gray-500">点击率</p>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* 筛选器 */}
      <div className="flex items-center gap-4 mb-6">
        <div className="flex items-center gap-2 bg-gray-100 p-1 rounded-lg">
          <button
            onClick={() => setFilter('all')}
            className={`px-4 py-2 rounded-md text-sm font-medium transition-all ${
              filter === 'all' ? 'bg-white text-gray-900 shadow-sm' : 'text-gray-600 hover:text-gray-900'
            }`}
          >
            全部推荐
          </button>
          <button
            onClick={() => setFilter('deals')}
            className={`px-4 py-2 rounded-md text-sm font-medium transition-all ${
              filter === 'deals' ? 'bg-white text-gray-900 shadow-sm' : 'text-gray-600 hover:text-gray-900'
            }`}
          >
            <span className="flex items-center gap-1">
              <span>促销专享</span>
              {filter === 'deals' && <span className="w-2 h-2 bg-red-500 rounded-full"></span>}
            </span>
          </button>
        </div>
        <div className="text-sm text-gray-500">
          共 {recommendations.length} 个推荐
        </div>
      </div>

      {/* 推荐列表 */}
      {loading ? (
        <div className="flex items-center justify-center h-64">
          <Loader2 className="w-8 h-8 animate-spin text-blue-600" />
        </div>
      ) : recommendations.length === 0 ? (
        <div className="bg-white rounded-lg shadow-sm p-12 text-center">
          <div className="w-16 h-16 bg-gray-100 rounded-full flex items-center justify-center mx-auto mb-4">
            <TrendingUp className="w-8 h-8 text-gray-400" />
          </div>
          <h3 className="text-lg font-medium text-gray-900 mb-2">暂无推荐</h3>
          <p className="text-gray-500 mb-4">绑定Steam账号并同步游戏库后，将为你生成个性化推荐</p>
          <button
            onClick={generateRecommendations}
            className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
          >
            立即生成推荐
          </button>
        </div>
      ) : (
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
          {recommendations.map((rec) => (
            <div
              key={rec.id}
              className="bg-white rounded-xl shadow-sm border border-gray-100 overflow-hidden hover:shadow-md transition-shadow"
            >
              <div className="p-4 flex gap-4">
                {/* 游戏封面 */}
                <div className="relative w-24 h-24 flex-shrink-0 bg-gray-100 rounded-lg overflow-hidden">
                  {rec.cover_url ? (
                    <img
                      src={rec.cover_url}
                      alt={rec.game_name}
                      className="w-full h-full object-cover"
                      onError={(e) => {
                        (e.target as HTMLImageElement).style.display = 'none';
                      }}
                    />
                  ) : (
                    <div className="w-full h-full flex items-center justify-center bg-gradient-to-br from-gray-200 to-gray-300">
                      <span className="text-2xl">🎮</span>
                    </div>
                  )}
                  {/* 匹配度徽章 */}
                  <div className={`absolute top-1 right-1 px-1.5 py-0.5 rounded text-xs font-bold ${getScoreBgColor(rec.match_score)} ${getScoreColor(rec.match_score)}`}>
                    {rec.match_score.toFixed(0)}%
                  </div>
                </div>

                {/* 游戏信息 */}
                <div className="flex-1 min-w-0">
                  <div className="flex items-start justify-between gap-2">
                    <h3 className="font-semibold text-gray-900 truncate">{rec.game_name}</h3>
                    <a
                      href={rec.store_url || `https://store.steampowered.com/app/${rec.game_id}`}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="p-1 text-gray-400 hover:text-blue-600 transition-colors flex-shrink-0"
                    >
                      <ExternalLink className="w-4 h-4" />
                    </a>
                  </div>

                  {/* 游戏标签 */}
                  <div className="flex flex-wrap gap-1 mt-1">
                    {rec.game_genre && (
                      <span className="px-2 py-0.5 bg-gray-100 text-gray-600 text-xs rounded-full">
                        {rec.game_genre}
                      </span>
                    )}
                    {rec.game_tags?.slice(0, 2).map((tag, idx) => (
                      <span key={idx} className="px-2 py-0.5 bg-blue-50 text-blue-600 text-xs rounded-full">
                        {tag}
                      </span>
                    ))}
                  </div>

                  {/* 促销信息 */}
                  {rec.has_deal && (
                    <div className="flex items-center gap-2 mt-2">
                      <span className="px-2 py-0.5 bg-red-100 text-red-600 text-xs font-medium rounded">
                        -{rec.deal_discount}%
                      </span>
                      <span className="text-lg font-bold text-red-600">
                        ¥{rec.deal_price.toFixed(2)}
                      </span>
                      {rec.deal_end_date && (
                        <span className="flex items-center gap-1 text-xs text-gray-500">
                          <Clock className="w-3 h-3" />
                          {rec.deal_end_date}截止
                        </span>
                      )}
                    </div>
                  )}

                  {/* 匹配理由 */}
                  <div className={`mt-2 transition-all ${expandedId === rec.id ? 'block' : 'hidden'}`}>
                    <div className="bg-blue-50 rounded-lg p-2 text-sm text-blue-800">
                      <p className="font-medium mb-1">推荐理由:</p>
                      <ul className="list-disc list-inside space-y-1">
                        {rec.match_reasons?.map((reason, idx) => (
                          <li key={idx}>{reason}</li>
                        ))}
                      </ul>
                    </div>
                  </div>

                  {/* 操作按钮 */}
                  <div className="flex items-center gap-2 mt-3">
                    <button
                      onClick={() => setExpandedId(expandedId === rec.id ? null : rec.id)}
                      className="px-3 py-1 text-xs text-blue-600 hover:bg-blue-50 rounded-lg transition-colors"
                    >
                      {expandedId === rec.id ? '收起理由' : '查看理由'}
                    </button>
                    <div className="flex-1" />
                    <button
                      onClick={() => submitFeedback(rec.id, 'like')}
                      disabled={feedbackLoading === rec.id}
                      className="p-1.5 text-gray-400 hover:text-emerald-600 hover:bg-emerald-50 rounded-lg transition-colors disabled:opacity-50"
                      title="喜欢"
                    >
                      <ThumbsUp className="w-4 h-4" />
                    </button>
                    <button
                      onClick={() => submitFeedback(rec.id, 'dislike')}
                      disabled={feedbackLoading === rec.id}
                      className="p-1.5 text-gray-400 hover:text-red-600 hover:bg-red-50 rounded-lg transition-colors disabled:opacity-50"
                      title="不喜欢"
                    >
                      <ThumbsDown className="w-4 h-4" />
                    </button>
                    <button
                      onClick={() => submitFeedback(rec.id, 'click')}
                      disabled={feedbackLoading === rec.id}
                      className="px-3 py-1 text-xs bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors disabled:opacity-50"
                    >
                      查看详情
                    </button>
                  </div>
                </div>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
