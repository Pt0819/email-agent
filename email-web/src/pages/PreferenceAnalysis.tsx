import { useState, useEffect } from 'react';
import { RefreshCw, Brain, Tag, AlertTriangle, Clock, TrendingUp, BarChart3 } from 'lucide-react';
import { steamApi } from '../api/steamApi';
import type { UserGamingProfile, PreferenceInsight, TagPreference } from '../api/types';

// 偏好标签颜色映射
const TAG_COLORS = [
  'from-emerald-400 to-emerald-600',
  'from-blue-400 to-blue-600',
  'from-purple-400 to-purple-600',
  'from-pink-400 to-pink-600',
  'from-amber-400 to-amber-600',
  'from-cyan-400 to-cyan-600',
  'from-rose-400 to-rose-600',
  'from-indigo-400 to-indigo-600',
  'from-teal-400 to-teal-600',
  'from-orange-400 to-orange-600',
];

const getTagColor = (tag: string): string => {
  let hash = 0;
  for (let i = 0; i < tag.length; i++) {
    hash = tag.charCodeAt(i) + ((hash << 5) - hash);
  }
  return TAG_COLORS[Math.abs(hash) % TAG_COLORS.length];
};

// 格式化游玩时长
const formatPlaytime = (minutes: number): string => {
  if (minutes < 60) return `${minutes}分钟`;
  const hours = Math.floor(minutes / 60);
  if (hours < 24) return `${hours}小时`;
  const days = Math.floor(hours / 24);
  return `${days}天${hours % 24}小时`;
};

// 格式化日期
const formatDate = (dateStr: string): string => {
  if (!dateStr) return '-';
  const d = new Date(dateStr);
  return d.toLocaleDateString('zh-CN', {
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  });
};

// 事件类型中文映射
const EVENT_TYPE_LABELS: Record<string, string> = {
  steam_email_sync: 'Steam邮件同步',
  library_sync: '游戏库同步',
  playtime_update: '游玩时长更新',
  new_game_added: '新增游戏',
  user_feedback: '用户反馈',
  game_purchased: '游戏购买',
  game_wishlisted: '愿望单添加',
  periodic_check: '定时检查',
  manual_trigger: '手动触发',
};

// 决策类型中文映射
const DECISION_TYPE_LABELS: Record<string, string> = {
  no_action: '无需行动',
  profile_update: '画像更新',
  tag_weight_adjust: '标签调整',
  anomaly_detected: '异常检测',
  preference_drift: '偏好漂移',
  new_pattern: '新模式识别',
  push_notification: '推送通知',
  generate_recommendation: '生成推荐',
  request_confirm: '请求确认',
};

// 标签权重可视化
function TagBadge({ tag, weight, size = 'md' }: { tag: string; weight: number; size?: 'sm' | 'md' | 'lg' }) {
  const colorClass = getTagColor(tag);
  const sizeClasses = {
    sm: 'px-2 py-0.5 text-xs',
    md: 'px-3 py-1 text-sm',
    lg: 'px-4 py-2 text-base',
  };
  const fontSize = {
    sm: 'text-xs',
    md: 'text-sm',
    lg: 'text-base',
  };

  // 权重越高，标签越大/越醒目
  const opacity = Math.min(1, 0.4 + weight / 10);

  return (
    <span
      className={`inline-flex items-center rounded-full bg-gradient-to-r ${colorClass} text-white font-medium shadow-sm hover:shadow-md transition-shadow cursor-default`}
      style={{ opacity }}
      title={`权重: ${weight.toFixed(2)}`}
    >
      <span className={`${sizeClasses[size]} ${fontSize[size]}`}>{tag}</span>
    </span>
  );
}

// 洞察卡片
function InsightCard({ insight, index }: { insight: PreferenceInsight; index: number }) {
  const isAnomaly = insight.is_anomaly;
  const eventLabel = EVENT_TYPE_LABELS[insight.event_type] || insight.event_type;
  const decisionLabel = DECISION_TYPE_LABELS[insight.decision_type] || insight.decision_type;

  return (
    <div
      className={`card p-4 hover:shadow-md transition-shadow animate-fade-in`}
      style={{ animationDelay: `${index * 50}ms` }}
    >
      <div className="flex items-start gap-3">
        {/* 图标 */}
        <div className={`flex-shrink-0 w-9 h-9 rounded-lg flex items-center justify-center ${
          isAnomaly
            ? 'bg-red-100 text-red-600'
            : 'bg-emerald-100 text-emerald-600'
        }`}>
          {isAnomaly ? (
            <AlertTriangle className="w-5 h-5" />
          ) : (
            <Brain className="w-5 h-5" />
          )}
        </div>

        {/* 内容 */}
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2 flex-wrap mb-1">
            {/* 事件类型 */}
            <span className="text-xs px-2 py-0.5 rounded-full bg-gray-100 text-gray-600">
              {eventLabel}
            </span>
            {/* 决策类型 */}
            <span className={`text-xs px-2 py-0.5 rounded-full ${
              isAnomaly
                ? 'bg-red-100 text-red-700'
                : 'bg-blue-100 text-blue-700'
            }`}>
              {decisionLabel}
            </span>
            {/* 置信度 */}
            {insight.confidence > 0 && (
              <span className="text-xs text-gray-400">
                置信度: {(insight.confidence * 100).toFixed(0)}%
              </span>
            )}
          </div>

          {/* 洞察内容 */}
          <p className={`text-sm ${isAnomaly ? 'text-red-700 font-medium' : 'text-gray-700'}`}>
            {insight.insight || insight.reasoning || '无内容'}
          </p>

          {/* 游戏信息 */}
          {insight.game_name && (
            <p className="text-xs text-gray-500 mt-1">
              游戏: <span className="font-medium text-gray-700">{insight.game_name}</span>
            </p>
          )}

          {/* 标签变化 */}
          {insight.tags_changed && insight.tags_changed.length > 0 && (
            <div className="flex flex-wrap gap-1 mt-2">
              {insight.tags_changed.slice(0, 5).map((tc, i) => (
                <span key={i} className="text-xs px-1.5 py-0.5 rounded bg-gray-50 text-gray-600">
                  {tc.tag}: {tc.delta > 0 ? '+' : ''}{tc.delta.toFixed(2)}
                </span>
              ))}
            </div>
          )}

          {/* 时间 */}
          <p className="text-xs text-gray-400 mt-2 flex items-center gap-1">
            <Clock className="w-3 h-3" />
            {formatDate(insight.created_at)}
          </p>
        </div>
      </div>
    </div>
  );
}

export default function PreferenceAnalysis() {
  const [profile, setProfile] = useState<UserGamingProfile | null>(null);
  const [insights, setInsights] = useState<PreferenceInsight[]>([]);
  const [totalInsights, setTotalInsights] = useState(0);
  const [loading, setLoading] = useState(false);
  const [analyzing, setAnalyzing] = useState(false);
  const [activeTab, setActiveTab] = useState<'tags' | 'genres' | 'insights'>('tags');

  useEffect(() => {
    loadProfile();
    loadInsights();
  }, []);

  const loadProfile = async () => {
    setLoading(true);
    try {
      const res = await steamApi.getPreferenceProfile();
      setProfile((res.data as any) || null);
    } catch (err) {
      console.error('加载偏好画像失败:', err);
      setProfile(null);
    } finally {
      setLoading(false);
    }
  };

  const loadInsights = async () => {
    try {
      const res = await steamApi.getInsights({ page: 1, page_size: 20 });
      const data = (res.data as any) || {};
      setInsights(data.list || []);
      setTotalInsights(data.total || 0);
    } catch (err) {
      console.error('加载洞察记录失败:', err);
    }
  };

  const handleAnalyze = async () => {
    setAnalyzing(true);
    try {
      await steamApi.analyzePreferences();
      await loadProfile();
      await loadInsights();
    } catch (err: any) {
      alert('分析失败: ' + (err.message || '未知错误'));
    } finally {
      setAnalyzing(false);
    }
  };

  const displayTags = profile?.top_tags?.slice(0, 30) || [];
  const displayGenres = profile?.top_genres?.slice(0, 15) || [];

  return (
    <div className="max-w-5xl mx-auto px-4 py-6 space-y-6 animate-fade-in">
      {/* 页面标题 */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900 flex items-center gap-2">
            <Brain className="w-7 h-7 text-emerald-600" />
            偏好画像
          </h1>
          <p className="mt-1 text-sm text-gray-500">
            基于您的游戏库和游玩记录，智能分析游戏偏好
          </p>
        </div>
        <button
          onClick={handleAnalyze}
          disabled={analyzing}
          className="btn-primary shadow-glow-steam"
        >
          <RefreshCw className={`w-4 h-4 ${analyzing ? 'animate-spin' : ''}`} />
          {analyzing ? '分析中...' : '重新分析'}
        </button>
      </div>

      {/* 统计概览 */}
      {profile && (
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
          <div className="card px-5 py-4 text-center hover-lift">
            <p className="text-2xl font-bold text-emerald-600">{profile.total_games}</p>
            <p className="text-xs text-gray-500 mt-0.5">拥有游戏</p>
          </div>
          <div className="card px-5 py-4 text-center hover-lift">
            <p className="text-2xl font-bold text-blue-600">{formatPlaytime(profile.total_playtime)}</p>
            <p className="text-xs text-gray-500 mt-0.5">总游玩时长</p>
          </div>
          <div className="card px-5 py-4 text-center hover-lift">
            <p className="text-2xl font-bold text-purple-600">{displayTags.length}</p>
            <p className="text-xs text-gray-500 mt-0.5">偏好标签</p>
          </div>
          <div className="card px-5 py-4 text-center hover-lift">
            <p className="text-2xl font-bold text-amber-600">{totalInsights}</p>
            <p className="text-xs text-gray-500 mt-0.5">洞察记录</p>
          </div>
        </div>
      )}

      {/* 标签云 */}
      {loading ? (
        <div className="card p-6">
          <div className="skeleton h-6 w-32 mb-4 rounded" />
          <div className="flex flex-wrap gap-2">
            {Array.from({ length: 15 }).map((_, i) => (
              <div key={i} className="skeleton h-8 w-20 rounded-full" style={{ animationDelay: `${i * 50}ms` }} />
            ))}
          </div>
        </div>
      ) : displayTags.length > 0 ? (
        <div className="card p-6">
          <div className="flex items-center gap-2 mb-4">
            <Tag className="w-5 h-5 text-emerald-600" />
            <h2 className="text-lg font-semibold text-gray-900">偏好标签</h2>
            <span className="text-xs text-gray-400">（标签大小表示偏好强度）</span>
          </div>
          <div className="flex flex-wrap gap-2">
            {displayTags.map((tagPref, index) => (
              <div key={tagPref.tag} className="animate-fade-in" style={{ animationDelay: `${index * 30}ms` }}>
                <TagBadge tag={tagPref.tag} weight={tagPref.weight} size="md" />
              </div>
            ))}
          </div>
        </div>
      ) : null}

      {/* 近期活动 */}
      {profile?.recent_activity && (
        <div className="card p-6">
          <div className="flex items-center gap-2 mb-4">
            <TrendingUp className="w-5 h-5 text-emerald-600" />
            <h2 className="text-lg font-semibold text-gray-900">近期活动</h2>
          </div>
          <div className="grid grid-cols-2 md:grid-cols-3 gap-4">
            <div className="text-center p-3 bg-gray-50 rounded-lg">
              <p className="text-xl font-bold text-emerald-600">
                {profile.recent_activity.games_played_last_week}
              </p>
              <p className="text-xs text-gray-500">上周游玩游戏</p>
            </div>
            <div className="text-center p-3 bg-gray-50 rounded-lg">
              <p className="text-xl font-bold text-blue-600">
                {formatPlaytime(profile.recent_activity.total_playtime_last_week)}
              </p>
              <p className="text-xs text-gray-500">上周总时长</p>
            </div>
            <div className="text-center p-3 bg-gray-50 rounded-lg">
              <p className="text-xl font-bold text-purple-600">
                {profile.recent_activity.most_played_game || '-'}
              </p>
              <p className="text-xs text-gray-500">最常玩</p>
            </div>
          </div>
        </div>
      )}

      {/* 洞察日志 */}
      <div className="card">
        <div className="p-4 border-b border-gray-100">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              <AlertTriangle className="w-5 h-5 text-emerald-600" />
              <h2 className="text-lg font-semibold text-gray-900">Agent 洞察日志</h2>
            </div>
            <span className="text-sm text-gray-400">{totalInsights} 条记录</span>
          </div>
        </div>

        {insights.length === 0 ? (
          <div className="empty-state py-16">
            <div className="empty-state-icon">
              <Brain className="w-10 h-10 text-gray-300" />
            </div>
            <p className="empty-state-title">暂无洞察记录</p>
            <p className="empty-state-desc">
              点击「重新分析」让Agent分析您的游戏偏好
            </p>
          </div>
        ) : (
          <div className="divide-y divide-gray-50">
            {insights.map((insight, index) => (
              <InsightCard key={insight.id || index} insight={insight} index={index} />
            ))}
          </div>
        )}
      </div>

      {/* 无数据提示 */}
      {!loading && !profile && (
        <div className="card p-12 text-center">
          <div className="w-20 h-20 bg-gradient-to-br from-gray-100 to-gray-50 rounded-2xl flex items-center justify-center mx-auto mb-4 shadow-sm">
            <BarChart3 className="w-10 h-10 text-gray-300" />
          </div>
          <p className="text-gray-500 font-medium mb-2">暂无偏好数据</p>
          <p className="text-gray-400 text-sm mb-6">
            绑定Steam账号并同步游戏库后，即可查看您的游戏偏好画像
          </p>
          <button
            onClick={handleAnalyze}
            disabled={analyzing}
            className="btn-primary shadow-glow-steam"
          >
            <Brain className="w-4 h-4" />
            {analyzing ? '分析中...' : '开始分析'}
          </button>
        </div>
      )}
    </div>
  );
}
