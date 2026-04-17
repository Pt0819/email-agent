import { useState, useEffect, useCallback } from 'react';
import { Link } from 'react-router-dom';
import { emailApi, syncApi } from '../api/client';
import type { Email, EmailCategory } from '../api/types';
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
      unclassified: 0,
    },
  });
  const [recentEmails, setRecentEmails] = useState<Email[]>([]);
  const [urgentEmails, setUrgentEmails] = useState<Email[]>([]);
  const [loading, setLoading] = useState(true);
  const [syncing, setSyncing] = useState(false);

  // 获取统计数据
  const fetchStats = useCallback(async () => {
    try {
      const today = new Date();
      today.setHours(0, 0, 0, 0);

      // 获取所有邮件统计
      const response = await emailApi.list({ page: 1, page_size: 1 });
      const pageData = response as unknown as { list: Email[]; total: number };
      const total = pageData.total || 0;

      // 获取今日邮件
      const todayResponse = await emailApi.list({ page: 1, page_size: 100 });
      const allEmails = (todayResponse as unknown as { list: Email[] }).list || [];

      // 统计今日邮件
      let todayCount = 0;
      const byCategory: Record<EmailCategory, number> = {
        work_urgent: 0,
        work_normal: 0,
        personal: 0,
        subscription: 0,
        notification: 0,
        promotion: 0,
        spam: 0,
        unclassified: 0,
      };

      allEmails.forEach((email) => {
        const emailDate = new Date(email.received_at);
        if (emailDate >= today) {
          todayCount++;
        }
        byCategory[email.category]++;
      });

      // 未处理邮件
      const unprocessed = allEmails.filter(
        (e) => e.category === 'unclassified'
      ).length;

      setStats({ total, today: todayCount, unprocessed, byCategory });

      // 最近邮件（按时间排序取前5封）
      const sorted = [...allEmails].sort(
        (a, b) =>
          new Date(b.received_at).getTime() - new Date(a.received_at).getTime()
      );
      setRecentEmails(sorted.slice(0, 5));

      // 紧急邮件
      const urgent = allEmails.filter(
        (e) => e.category === 'work_urgent' || e.priority === 'critical'
      );
      setUrgentEmails(urgent.slice(0, 3));
    } catch (err) {
      console.error('获取统计数据失败:', err);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchStats();
  }, [fetchStats]);

  // 处理同步
  const handleSync = async () => {
    try {
      setSyncing(true);
      await syncApi.trigger();
      setTimeout(() => {
        fetchStats();
        setSyncing(false);
      }, 3000);
    } catch (err) {
      console.error('同步失败:', err);
      setSyncing(false);
    }
  };

  // 统计卡片数据
  const statCards = [
    {
      title: '总邮件数',
      value: stats.total,
      icon: <Mail className="w-6 h-6" />,
      color: 'bg-blue-500',
      bgColor: 'bg-blue-50',
      textColor: 'text-blue-600',
    },
    {
      title: '今日新增',
      value: stats.today,
      icon: <TrendingUp className="w-6 h-6" />,
      color: 'bg-green-500',
      bgColor: 'bg-green-50',
      textColor: 'text-green-600',
    },
    {
      title: '待分类',
      value: stats.unprocessed,
      icon: <Inbox className="w-6 h-6" />,
      color: 'bg-yellow-500',
      bgColor: 'bg-yellow-50',
      textColor: 'text-yellow-600',
    },
    {
      title: '紧急工作',
      value: stats.byCategory.work_urgent,
      icon: <AlertTriangle className="w-6 h-6" />,
      color: 'bg-red-500',
      bgColor: 'bg-red-50',
      textColor: 'text-red-600',
    },
  ];

  // 分类统计（排除未分类和垃圾邮件）
  const categoryStats = Object.entries(stats.byCategory)
    .filter(([key]) => key !== 'unclassified' && key !== 'spam')
    .sort(([, a], [, b]) => b - a)
    .slice(0, 5);

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary-600"></div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
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
          className={`flex items-center gap-2 px-4 py-2 rounded-lg font-medium transition-colors ${
            syncing
              ? 'bg-gray-100 text-gray-500 cursor-not-allowed'
              : 'bg-primary-600 text-white hover:bg-primary-700'
          }`}
        >
          <RefreshCw className={`w-4 h-4 ${syncing ? 'animate-spin' : ''}`} />
          {syncing ? '同步中...' : '同步邮件'}
        </button>
      </div>

      {/* 统计卡片 */}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
        {statCards.map((card, index) => (
          <div
            key={index}
            className="bg-white rounded-lg border border-gray-200 p-4"
          >
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-gray-500">{card.title}</p>
                <p className="text-3xl font-bold text-gray-900 mt-1">
                  {card.value}
                </p>
              </div>
              <div className={`${card.bgColor} p-3 rounded-lg`}>
                <div className={card.textColor}>{card.icon}</div>
              </div>
            </div>
          </div>
        ))}
      </div>

      {/* 分类统计 */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        {/* 分类分布 */}
        <div className="bg-white rounded-lg border border-gray-200 p-6">
          <h2 className="text-lg font-semibold text-gray-900 mb-4">
            邮件分类分布
          </h2>
          <div className="space-y-3">
            {categoryStats.map(([category, count]) => {
              const percentage = stats.total > 0 ? (count / stats.total) * 100 : 0;
              return (
                <div key={category}>
                  <div className="flex items-center justify-between text-sm mb-1">
                    <span className="text-gray-700">
                      {CATEGORY_LABELS[category as EmailCategory]}
                    </span>
                    <span className="text-gray-500">
                      {count} 封 ({percentage.toFixed(1)}%)
                    </span>
                  </div>
                  <div className="h-2 bg-gray-100 rounded-full overflow-hidden">
                    <div
                      className={`h-full rounded-full ${CATEGORY_COLORS[category as EmailCategory]}`}
                      style={{ width: `${percentage}%` }}
                    ></div>
                  </div>
                </div>
              );
            })}
          </div>
        </div>

        {/* 紧急邮件 */}
        <div className="bg-white rounded-lg border border-gray-200 p-6">
          <div className="flex items-center justify-between mb-4">
            <h2 className="text-lg font-semibold text-gray-900 flex items-center gap-2">
              <AlertTriangle className="w-5 h-5 text-red-500" />
              紧急邮件
            </h2>
            {urgentEmails.length > 0 && (
              <Link
                to="/emails?category=work_urgent"
                className="text-sm text-primary-600 hover:text-primary-700 flex items-center gap-1"
              >
                查看全部 <ArrowRight className="w-4 h-4" />
              </Link>
            )}
          </div>

          {urgentEmails.length === 0 ? (
            <div className="text-center py-8 text-gray-400">
              <Bot className="w-12 h-12 mx-auto mb-2 opacity-50" />
              <p>暂无紧急邮件</p>
            </div>
          ) : (
            <div className="space-y-3">
              {urgentEmails.map((email) => (
                <Link
                  key={email.id}
                  to={`/emails/${email.id}`}
                  className="block p-3 rounded-lg border border-red-100 bg-red-50 hover:bg-red-100 transition-colors"
                >
                  <div className="flex items-start justify-between">
                    <div className="flex-1 min-w-0">
                      <p className="text-sm font-medium text-gray-900 truncate">
                        {email.subject || '(无主题)'}
                      </p>
                      <p className="text-xs text-gray-500 mt-1">
                        {email.sender_name || email.sender_email}
                      </p>
                    </div>
                    <span className="text-xs text-red-600 flex-shrink-0 ml-2">
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
      <div className="bg-white rounded-lg border border-gray-200 p-6">
        <div className="flex items-center justify-between mb-4">
          <h2 className="text-lg font-semibold text-gray-900 flex items-center gap-2">
            <Clock className="w-5 h-5 text-gray-500" />
            最近邮件
          </h2>
          <Link
            to="/emails"
            className="text-sm text-primary-600 hover:text-primary-700 flex items-center gap-1"
          >
            查看全部 <ArrowRight className="w-4 h-4" />
          </Link>
        </div>

        {recentEmails.length === 0 ? (
          <div className="text-center py-12 text-gray-400">
            <Mail className="w-12 h-12 mx-auto mb-2 opacity-50" />
            <p>暂无邮件</p>
            <p className="text-sm mt-1">点击上方同步按钮拉取邮件</p>
          </div>
        ) : (
          <div className="divide-y divide-gray-100">
            {recentEmails.map((email) => (
              <Link
                key={email.id}
                to={`/emails/${email.id}`}
                className="flex items-center justify-between py-3 hover:bg-gray-50 transition-colors -mx-2 px-2 rounded-lg"
              >
                <div className="flex items-center gap-3 flex-1 min-w-0">
                  <div className="w-10 h-10 rounded-full bg-gradient-to-br from-primary-400 to-primary-600 flex items-center justify-center text-white text-sm font-medium flex-shrink-0">
                    {(email.sender_name || email.sender_email).charAt(0).toUpperCase()}
                  </div>
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2">
                      <p className="text-sm font-medium text-gray-900 truncate">
                        {email.sender_name || email.sender_email}
                      </p>
                      <span
                        className={`px-2 py-0.5 text-xs rounded-full border ${CATEGORY_COLORS[email.category]}`}
                      >
                        {CATEGORY_LABELS[email.category]}
                      </span>
                    </div>
                    <p className="text-sm text-gray-500 truncate mt-0.5">
                      {email.subject || '(无主题)'}
                    </p>
                  </div>
                </div>
                <span className="text-xs text-gray-400 flex-shrink-0 ml-2">
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

// 格式化时间
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

  return date.toLocaleDateString('zh-CN', {
    month: '2-digit',
    day: '2-digit',
  });
}
