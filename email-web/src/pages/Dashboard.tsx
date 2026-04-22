import { useState, useEffect, useCallback } from 'react';
import { Link } from 'react-router-dom';
import { emailApi, syncApi, summaryApi } from '../api/client';
import type { Email, EmailCategory, DailySummary } from '../api/types';
import { CATEGORY_LABELS, CATEGORY_COLORS } from '../api/types';
import {
  Mail,
  RefreshCw,
  AlertTriangle,
  Clock,
  TrendingUp,
  Inbox,
  ArrowRight,
  Bot,
  CheckCircle,
  ListTodo,
  Sparkles,
} from 'lucide-react';

interface Stats {
  total: number;
  today: number;
  unprocessed: number;
  byCategory: Record<EmailCategory, number>;
}

export default function Dashboard() {
  const [stats, setStats] = useState<Stats>({
    total: 0,
    today: 0,
    unprocessed: 0,
    byCategory: {
      work_urgent: 0,
      work_normal: 0,
      personal: 0,
      subscription: 0,
      notification: 0,
      promotion: 0,
      spam: 0,
      steam_promotion: 0,
      steam_wishlist: 0,
      steam_news: 0,
      steam_update: 0,
      unclassified: 0,
    },
  });
  const [recentEmails, setRecentEmails] = useState<Email[]>([]);
  const [urgentEmails, setUrgentEmails] = useState<Email[]>([]);
  const [loading, setLoading] = useState(true);
  const [syncing, setSyncing] = useState(false);
  const [dailySummary, setDailySummary] = useState<DailySummary | null>(null);

  const fetchStats = useCallback(async () => {
    try {
      const today = new Date();
      today.setHours(0, 0, 0, 0);

      const response = await emailApi.list({ page: 1, page_size: 1 });
      const pageData = response as unknown as { list: Email[]; total: number };
      const total = pageData.total || 0;

      const todayResponse = await emailApi.list({ page: 1, page_size: 100 });
      const allEmails = (todayResponse as unknown as { list: Email[] }).list || [];

      let todayCount = 0;
      const byCategory: Record<EmailCategory, number> = {
        work_urgent: 0, work_normal: 0, personal: 0, subscription: 0,
        notification: 0, promotion: 0, spam: 0, steam_promotion: 0,
        steam_wishlist: 0, steam_news: 0, steam_update: 0, unclassified: 0,
      };

      allEmails.forEach((email) => {
        const emailDate = new Date(email.received_at);
        if (emailDate >= today) todayCount++;
        byCategory[email.category]++;
      });

      const unprocessed = allEmails.filter((e) => e.category === 'unclassified').length;
      setStats({ total, today: todayCount, unprocessed, byCategory });

      const sorted = [...allEmails].sort(
        (a, b) => new Date(b.received_at).getTime() - new Date(a.received_at).getTime()
      );
      setRecentEmails(sorted.slice(0, 5));

      const urgent = allEmails.filter((e) => e.category === 'work_urgent' || e.priority === 'critical');
      setUrgentEmails(urgent.slice(0, 3));
    } catch (err) {
      console.error('获取统计数据失败:', err);
    } finally {
      setLoading(false);
    }
  }, []);

  const fetchDailySummary = useCallback(async () => {
    try {
      const response = await summaryApi.daily();
      setDailySummary(response as unknown as DailySummary);
    } catch (err) {
      console.error('获取每日摘要失败:', err);
    }
  }, []);

  useEffect(() => {
    fetchStats();
    fetchDailySummary();
  }, [fetchStats, fetchDailySummary]);

  const handleSync = async () => {
    try {
      setSyncing(true);
      await syncApi.trigger();
      setTimeout(() => {
        fetchStats();
        fetchDailySummary();
        setSyncing(false);
      }, 3000);
    } catch (err) {
      console.error('同步失败:', err);
      setSyncing(false);
    }
  };

  const statCards = [
    { title: '总邮件数', value: stats.total, icon: Mail, gradient: 'from-blue-500 to-blue-600', bg: 'bg-blue-50', text: 'text-blue-600' },
    { title: '今日新增', value: dailySummary?.total_emails ?? stats.today, icon: TrendingUp, gradient: 'from-emerald-500 to-emerald-600', bg: 'bg-emerald-50', text: 'text-emerald-600' },
    { title: '待分类', value: stats.unprocessed, icon: Inbox, gradient: 'from-amber-500 to-amber-600', bg: 'bg-amber-50', text: 'text-amber-600' },
    { title: '紧急工作', value: stats.byCategory.work_urgent, icon: AlertTriangle, gradient: 'from-red-500 to-red-600', bg: 'bg-red-50', text: 'text-red-600' },
  ];

  const categoryStats = Object.entries(stats.byCategory)
    .filter(([key]) => key !== 'unclassified' && key !== 'spam')
    .sort(([, a], [, b]) => b - a)
    .slice(0, 5);

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="w-8 h-8 spinner"></div>
      </div>
    );
  }

  return (
    <div className="page-container space-y-8 animate-fade-in">
      {/* 页面标题 */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">仪表盘</h1>
          <p className="text-sm text-gray-500 mt-1">
            {new Date().toLocaleDateString('zh-CN', {
              year: 'numeric',
              month: 'long',
              day: 'numeric',
              weekday: 'long',
            })}
          </p>
        </div>
        <button
          onClick={handleSync}
          disabled={syncing}
          className={`btn-primary shadow-glow ${syncing ? 'opacity-70' : ''}`}
        >
          <RefreshCw className={`w-4 h-4 ${syncing ? 'animate-spin' : ''}`} />
          {syncing ? '同步中...' : '同步邮件'}
        </button>
      </div>

      {/* 统计卡片 */}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-5">
        {statCards.map((card, index) => {
          const Icon = card.icon;
          return (
            <div
              key={index}
              className="card p-5 hover-lift glow-on-hover"
              style={{ animationDelay: `${index * 80}ms` }}
            >
              <div className="flex items-start justify-between">
                <div>
                  <p className="text-sm text-gray-500 font-medium">{card.title}</p>
                  <p className="text-3xl font-bold text-gray-900 mt-2 tracking-tight">
                    {card.value}
                  </p>
                </div>
                <div className={`w-11 h-11 rounded-xl ${card.bg} flex items-center justify-center shadow-sm`}>
                  <Icon className={`w-5 h-5 ${card.text}`} />
                </div>
              </div>
            </div>
          );
        })}
      </div>

      {/* AI每日摘要 */}
      {dailySummary && (
        <div className="bg-gradient-to-r from-primary-50 via-blue-50 to-indigo-50 rounded-2xl border border-primary-100 p-7 shadow-md hover-lift transition-all duration-300">
          <div className="flex items-start gap-5">
            <div className="w-12 h-12 bg-white rounded-xl flex items-center justify-center flex-shrink-0 shadow-sm ring-2 ring-primary-100">
              <Sparkles className="w-6 h-6 text-primary-600" />
            </div>
            <div className="flex-1 min-w-0">
              <h2 className="text-lg font-bold text-gray-900 mb-3">AI 每日摘要</h2>
              <p className="text-gray-700 leading-relaxed">
                {dailySummary.summary_text}
              </p>
              {dailySummary.important_emails.length > 0 && (
                <div className="mt-5">
                  <h3 className="text-sm font-semibold text-gray-700 mb-3 flex items-center gap-1.5">
                    <AlertTriangle className="w-4 h-4 text-amber-500" />
                    重要邮件
                  </h3>
                  <div className="space-y-2.5">
                    {dailySummary.important_emails.slice(0, 3).map((email) => (
                      <Link
                        key={email.email_id}
                        to={`/emails/${email.email_id}`}
                        className="block p-4 bg-white rounded-xl border border-gray-100 hover:border-primary-200 hover:shadow-md transition-all duration-200"
                      >
                        <div className="flex items-start justify-between gap-3">
                          <div className="flex-1 min-w-0">
                            <p className="text-sm font-medium text-gray-900 truncate">{email.subject}</p>
                            <p className="text-xs text-gray-500 mt-1">{email.sender}</p>
                          </div>
                          <span className={`px-2.5 py-1 text-xs font-medium rounded-lg ${
                            email.priority === 'critical'
                              ? 'bg-red-50 text-red-700 border border-red-100'
                              : 'bg-orange-50 text-orange-700 border border-orange-100'
                          }`}>
                            {email.priority === 'critical' ? '紧急' : '高优先'}
                          </span>
                        </div>
                        {email.summary && (
                          <p className="text-xs text-gray-500 mt-2 line-clamp-1">{email.summary}</p>
                        )}
                      </Link>
                    ))}
                  </div>
                </div>
              )}
              {dailySummary.action_items.length > 0 && (
                <div className="mt-5">
                  <h3 className="text-sm font-semibold text-gray-700 mb-3 flex items-center gap-1.5">
                    <ListTodo className="w-4 h-4 text-blue-500" />
                    待办事项
                  </h3>
                  <div className="space-y-2">
                    {dailySummary.action_items.slice(0, 5).map((item, idx) => (
                      <div key={idx} className="flex items-center gap-2.5 text-sm text-gray-600">
                        <div className="w-5 h-5 rounded-md border-2 border-gray-200 flex items-center justify-center flex-shrink-0 bg-white">
                          <CheckCircle className="w-3 h-3 text-gray-300" />
                        </div>
                        <span>{item.task}</span>
                      </div>
                    ))}
                  </div>
                </div>
              )}
            </div>
          </div>
        </div>
      )}

      {/* 分类统计 + 紧急邮件 */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        {/* 分类分布 */}
        <div className="card p-6">
          <h2 className="text-lg font-bold text-gray-900 mb-5">邮件分类分布</h2>
          <div className="space-y-4">
            {categoryStats.map(([category, count]) => {
              const percentage = stats.total > 0 ? (count / stats.total) * 100 : 0;
              return (
                <div key={category}>
                  <div className="flex items-center justify-between text-sm mb-2">
                    <span className="font-medium text-gray-700">
                      {CATEGORY_LABELS[category as EmailCategory]}
                    </span>
                    <span className="text-gray-400 tabular-nums">
                      {count} <span className="text-gray-300">·</span> {percentage.toFixed(0)}%
                    </span>
                  </div>
                  <div className="progress-bar">
                    <div
                      className={`progress-bar-fill ${CATEGORY_COLORS[category as EmailCategory]?.split(' ')[0] || 'bg-gray-300'}`}
                      style={{ width: `${Math.max(percentage, 2)}%` }}
                    />
                  </div>
                </div>
              );
            })}
          </div>
        </div>

        {/* 紧急邮件 */}
        <div className="card p-6">
          <div className="flex items-center justify-between mb-5">
            <h2 className="text-lg font-bold text-gray-900 flex items-center gap-2">
              <AlertTriangle className="w-5 h-5 text-red-500" />
              紧急邮件
            </h2>
            {urgentEmails.length > 0 && (
              <Link
                to="/emails?category=work_urgent"
                className="text-sm text-primary-600 hover:text-primary-700 font-medium flex items-center gap-1 transition-colors"
              >
                查看全部 <ArrowRight className="w-4 h-4" />
              </Link>
            )}
          </div>

          {urgentEmails.length === 0 ? (
            <div className="empty-state">
              <div className="empty-state-icon">
                <Bot className="w-10 h-10 text-gray-300" />
              </div>
              <p className="empty-state-title">暂无紧急邮件</p>
              <p className="empty-state-desc">一切正常，可以安心处理其他事务</p>
            </div>
          ) : (
            <div className="space-y-3">
              {urgentEmails.map((email) => (
                <Link
                  key={email.id}
                  to={`/emails/${email.id}`}
                  className="block p-4 rounded-xl border border-red-100 bg-red-50/50 hover:bg-red-50 hover:shadow-md hover-lift transition-all duration-200"
                >
                  <div className="flex items-start justify-between">
                    <div className="flex-1 min-w-0">
                      <p className="text-sm font-medium text-gray-900 truncate">
                        {email.subject || '(无主题)'}
                      </p>
                      <p className="text-xs text-gray-500 mt-1.5">
                        {email.sender_name || email.sender_email}
                      </p>
                    </div>
                    <span className="text-xs text-red-500 flex-shrink-0 ml-3 font-medium">
                      {formatTime(email.received_at)}
                    </span>
                  </div>
                </Link>
              ))}
            </div>
          )}
        </div>
      </div>

      {/* 最近邮件 */}
      <div className="card p-6">
        <div className="flex items-center justify-between mb-5">
          <h2 className="text-lg font-bold text-gray-900 flex items-center gap-2">
            <Clock className="w-5 h-5 text-gray-400" />
            最近邮件
          </h2>
          <Link
            to="/emails"
            className="text-sm text-primary-600 hover:text-primary-700 font-medium flex items-center gap-1 transition-colors"
          >
            查看全部 <ArrowRight className="w-4 h-4" />
          </Link>
        </div>

        {recentEmails.length === 0 ? (
          <div className="empty-state">
            <div className="empty-state-icon">
              <Mail className="w-10 h-10 text-gray-300" />
            </div>
            <p className="empty-state-title">暂无邮件</p>
            <p className="empty-state-desc">点击上方同步按钮拉取邮件</p>
          </div>
        ) : (
          <div className="divide-y divide-gray-50">
            {recentEmails.map((email) => (
              <Link
                key={email.id}
                to={`/emails/${email.id}`}
                className="flex items-center justify-between py-4 hover:bg-gray-50/50 transition-colors -mx-3 px-3 rounded-xl group"
              >
                <div className="flex items-center gap-4 flex-1 min-w-0">
                  <div className="w-10 h-10 rounded-xl bg-gradient-to-br from-primary-400 to-primary-500 flex items-center justify-center text-white text-sm font-bold flex-shrink-0 shadow-sm">
                    {(email.sender_name || email.sender_email).charAt(0).toUpperCase()}
                  </div>
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2.5">
                      <p className="text-sm font-medium text-gray-900 truncate">
                        {email.sender_name || email.sender_email}
                      </p>
                      <span
                        className={`px-2 py-0.5 text-xs font-medium rounded-md ${CATEGORY_COLORS[email.category]}`}
                      >
                        {CATEGORY_LABELS[email.category]}
                      </span>
                    </div>
                    <p className="text-sm text-gray-500 truncate mt-0.5">
                      {email.subject || '(无主题)'}
                    </p>
                  </div>
                </div>
                <span className="text-xs text-gray-400 flex-shrink-0 ml-3 group-hover:text-gray-600 transition-colors">
                  {formatTime(email.received_at)}
                </span>
              </Link>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}

function formatTime(dateStr: string): string {
  const date = new Date(dateStr);
  const now = new Date();
  const diffMs = now.getTime() - date.getTime();
  const diffMins = Math.floor(diffMs / 60000);
  const diffHours = Math.floor(diffMs / 3600000);
  const diffDays = Math.floor(diffMs / 86400000);

  if (diffMins < 1) return '刚刚';
  if (diffMins < 60) return `${diffMins}分钟前`;
  if (diffHours < 24) return `${diffHours}小时前`;
  if (diffDays < 7) return `${diffDays}天前`;

  return date.toLocaleDateString('zh-CN', { month: '2-digit', day: '2-digit' });
}
