import type { Email } from '../../api/types';
import { CATEGORY_LABELS, CATEGORY_COLORS, PRIORITY_LABELS, PRIORITY_COLORS } from '../../api/types';
import { Mail, User, Clock, Paperclip, Eye } from 'lucide-react';

interface EmailCardProps {
  email: Email;
  onClassify?: (id: string | number) => void;
  onView?: (email: Email) => void;
}

export default function EmailCard({ email, onClassify, onView }: EmailCardProps) {
  const isUnread = email.status === 'unread';

  return (
    <div
      className={`p-4 bg-white rounded-lg border transition-all cursor-pointer ${
        isUnread
          ? 'border-primary-200 shadow-sm hover:shadow-md hover:border-primary-300'
          : 'border-gray-200 hover:border-gray-300'
      }`}
      onClick={() => onView?.(email)}
    >
      <div className="flex items-start justify-between gap-4">
        {/* 左侧内容 */}
        <div className="flex-1 min-w-0">
          {/* 标签行 */}
          <div className="flex items-center gap-2 mb-2 flex-wrap">
            <Mail className="w-4 h-4 text-gray-400 flex-shrink-0" />
            <span
              className={`px-2 py-0.5 text-xs rounded-full border ${
                CATEGORY_COLORS[email.category]
              }`}
            >
              {CATEGORY_LABELS[email.category]}
            </span>
            <span
              className={`text-xs font-medium ${PRIORITY_COLORS[email.priority]}`}
            >
              {PRIORITY_LABELS[email.priority]}
            </span>
            {isUnread && (
              <span className="flex items-center gap-1 text-xs text-primary-600">
                <Eye className="w-3 h-3" />
                未读
              </span>
            )}
            {email.has_attachment && (
              <span className="flex items-center gap-1 text-xs text-gray-500">
                <Paperclip className="w-3 h-3" />
                附件
              </span>
            )}
          </div>

          {/* 主题 */}
          <h3
            className={`text-lg mb-1 truncate ${
              isUnread ? 'font-semibold text-gray-900' : 'font-medium text-gray-700'
            }`}
          >
            {email.subject || '(无主题)'}
          </h3>

          {/* 发件人和时间 */}
          <div className="flex items-center gap-4 text-sm text-gray-500">
            <div className="flex items-center gap-1 truncate">
              <User className="w-4 h-4 flex-shrink-0" />
              <span className="truncate">
                {email.sender_name || email.sender_email}
              </span>
            </div>
            <div className="flex items-center gap-1 flex-shrink-0">
              <Clock className="w-4 h-4" />
              <span>{formatReceivedTime(email.received_at)}</span>
            </div>
          </div>

          {/* 预览内容 */}
          {email.content && (
            <p className="mt-2 text-sm text-gray-600 line-clamp-2">
              {email.content.replace(/<[^>]*>/g, '')}
            </p>
          )}
        </div>

        {/* 右侧操作按钮 */}
        <div className="flex flex-col gap-2 flex-shrink-0">
          {email.category === 'unclassified' && onClassify && (
            <button
              onClick={(e) => {
                e.stopPropagation();
                onClassify(email.id);
              }}
              className="px-3 py-1.5 text-sm border border-primary-300 text-primary-700 rounded-lg hover:bg-primary-50 transition-colors"
            >
              AI分类
            </button>
          )}
        </div>
      </div>
    </div>
  );
}

// 格式化接收时间
function formatReceivedTime(dateStr: string): string {
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
    hour: '2-digit',
    minute: '2-digit',
  });
}
