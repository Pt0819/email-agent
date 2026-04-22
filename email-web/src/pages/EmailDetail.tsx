import { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { emailApi } from '../api/client';
import { CATEGORY_LABELS, CATEGORY_COLORS, PRIORITY_LABELS } from '../api/types';
import type { Email } from '../api/types';
import {
  ArrowLeft,
  Clock,
  Mail,
  Paperclip,
  AlertCircle,
  Bot,
  Check,
  Archive,
  Tag,
  FileText,
  Inbox,
} from 'lucide-react';

export default function EmailDetail() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();

  const [email, setEmail] = useState<Email | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [classifying, setClassifying] = useState(false);
  const [actionLoading, setActionLoading] = useState<string | null>(null);

  useEffect(() => {
    if (id) {
      fetchEmailDetail(id);
    }
  }, [id]);

  const fetchEmailDetail = async (emailId: string) => {
    try {
      setLoading(true);
      const response = await emailApi.getById(Number(emailId));
      setEmail(response as unknown as Email);
      setError(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : '获取邮件详情失败');
    } finally {
      setLoading(false);
    }
  };

  const handleClassify = async () => {
    if (!email) return;
    try {
      setClassifying(true);
      await emailApi.classify(email.id);
      await fetchEmailDetail(String(email.id));
    } catch (err) {
      setError(err instanceof Error ? err.message : '分类失败');
    } finally {
      setClassifying(false);
    }
  };

  const handleMarkAsRead = async () => {
    if (!email || email.status === 'read') return;
    try {
      setActionLoading('read');
      await emailApi.updateStatus(email.id, 'read');
      setEmail({ ...email, status: 'read' });
    } catch (err) {
      setError(err instanceof Error ? err.message : '标记已读失败');
    } finally {
      setActionLoading(null);
    }
  };

  const handleArchive = async () => {
    if (!email || email.status === 'archived') return;
    try {
      setActionLoading('archive');
      await emailApi.updateStatus(email.id, 'archived');
      setEmail({ ...email, status: 'archived' });
    } catch (err) {
      setError(err instanceof Error ? err.message : '归档失败');
    } finally {
      setActionLoading(null);
    }
  };

  // 加载状态
  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary-600"></div>
      </div>
    );
  }

  // 错误状态
  if (error || !email) {
    return (
      <div className="flex items-center justify-center h-64 text-red-600">
        <AlertCircle className="w-5 h-5 mr-2" />
        {error || '邮件不存在'}
      </div>
    );
  }

  const isArchived = email.status === 'archived';
  const isRead = email.status === 'read';

  return (
    <div className="max-w-5xl mx-auto px-4 py-6 animate-fade-in">
      <div className="max-w-4xl mx-auto space-y-6">
      {/* 顶部操作栏 */}
      <div className="flex items-center justify-between">
        <button
          onClick={() => navigate(-1)}
          className="flex items-center gap-2 text-gray-600 hover:text-gray-900 transition-colors group"
        >
          <ArrowLeft className="w-5 h-5 group-hover:-translate-x-0.5 transition-transform" />
          返回
        </button>

        <div className="flex items-center gap-2">
          {email.category === 'unclassified' && (
            <button
              onClick={handleClassify}
              disabled={classifying}
              className={`flex items-center gap-2 px-4 py-2.5 rounded-xl font-medium transition-all duration-200 ${
                classifying
                  ? 'bg-gray-100 text-gray-500 cursor-not-allowed'
                  : 'btn-primary shadow-glow'
              }`}
            >
              <Bot className="w-4 h-4" />
              {classifying ? '分类中...' : 'AI分类'}
            </button>
          )}

          {!isRead && (
            <button
              onClick={handleMarkAsRead}
              disabled={actionLoading === 'read'}
              className="btn-secondary"
            >
              {actionLoading === 'read' ? (
                <div className="w-4 h-4 border-2 border-gray-300 border-t-primary-600 rounded-full animate-spin"></div>
              ) : (
                <Check className="w-4 h-4" />
              )}
              标记已读
            </button>
          )}

          {!isArchived && (
            <button
              onClick={handleArchive}
              disabled={actionLoading === 'archive'}
              className="btn-secondary"
            >
              {actionLoading === 'archive' ? (
                <div className="w-4 h-4 border-2 border-gray-300 border-t-primary-600 rounded-full animate-spin"></div>
              ) : (
                <Archive className="w-4 h-4" />
              )}
              归档
            </button>
          )}
        </div>
      </div>

      {/* 归档提示 */}
      {isArchived && (
        <div className="bg-amber-50 border border-amber-200 rounded-xl p-4 flex items-center gap-3">
          <Inbox className="w-5 h-5 text-amber-600" />
          <span className="text-sm text-amber-800 font-medium">此邮件已归档</span>
        </div>
      )}

      {/* 邮件主体 */}
      <div className="card overflow-hidden">
        {/* 邮件头部 */}
        <div className="p-6 border-b border-gray-100">
          {/* 标签行 */}
          <div className="flex items-center gap-2 mb-5 flex-wrap">
            <span
              className={`tag ${CATEGORY_COLORS[email.category]}`}
            >
              <Tag className="w-3 h-3 inline mr-1" />
              {CATEGORY_LABELS[email.category]}
            </span>
            <span
              className={`tag ${
                email.priority === 'critical'
                  ? 'bg-red-50 text-red-700 border border-red-100'
                  : email.priority === 'high'
                  ? 'bg-orange-50 text-orange-700 border border-orange-100'
                  : email.priority === 'medium'
                  ? 'bg-yellow-50 text-yellow-700 border border-yellow-100'
                  : 'bg-gray-50 text-gray-700 border border-gray-100'
              }`}
            >
              {PRIORITY_LABELS[email.priority]}
            </span>
            {email.has_attachment && (
              <span className="tag border border-gray-200 text-gray-700">
                <Paperclip className="w-3 h-3 mr-1" />
                有附件
              </span>
            )}
            {!isRead && (
              <span className="tag bg-primary-50 text-primary-700 border border-primary-100">
                未读
              </span>
            )}
          </div>

          {/* 主题 */}
          <h1 className="text-2xl font-bold text-gray-900 mb-5 leading-snug">
            {email.subject || '(无主题)'}
          </h1>

          {/* 发件人信息 */}
          <div className="flex items-center gap-4">
            <div className="w-12 h-12 rounded-xl bg-gradient-to-br from-primary-400 to-primary-500 flex items-center justify-center text-white font-semibold shadow-md">
              {(email.sender_name || email.sender_email).charAt(0).toUpperCase()}
            </div>
            <div className="flex-1">
              <div className="font-semibold text-gray-900">
                {email.sender_name || email.sender_email}
              </div>
              <div className="text-sm text-gray-500 mt-0.5">{email.sender_email}</div>
              <div className="flex items-center gap-4 text-sm text-gray-400 mt-1">
                <div className="flex items-center gap-1.5">
                  <Clock className="w-3.5 h-3.5" />
                  <span>{new Date(email.received_at).toLocaleString('zh-CN')}</span>
                </div>
                {email.message_id && (
                  <div className="flex items-center gap-1.5">
                    <Mail className="w-3.5 h-3.5" />
                    <span className="text-xs font-mono">{email.message_id.slice(0, 30)}...</span>
                  </div>
                )}
              </div>
            </div>
          </div>
        </div>

        {/* 邮件正文 */}
        <div className="p-6">
          {email.content_html ? (
            <div
              className="prose max-w-none"
              dangerouslySetInnerHTML={{ __html: email.content_html }}
            />
          ) : email.content ? (
            <div className="whitespace-pre-wrap text-gray-700 leading-relaxed">
              {email.content}
            </div>
          ) : (
            <div className="empty-state py-8">
              <div className="empty-state-icon">
                <FileText className="w-8 h-8 text-gray-300" />
              </div>
              <p className="text-gray-400">无正文内容</p>
            </div>
          )}
        </div>

        {/* 附件区域 */}
        {email.has_attachment && (
          <div className="p-6 border-t border-gray-100 bg-gray-50/50">
            <div className="flex items-center gap-2 text-sm font-semibold text-gray-700 mb-3">
              <Paperclip className="w-4 h-4" />
              附件
            </div>
            <div className="card p-4">
              <div className="flex items-center gap-3 text-sm text-gray-600">
                <div className="w-10 h-10 bg-gray-100 rounded-lg flex items-center justify-center">
                  <FileText className="w-5 h-5 text-gray-400" />
                </div>
                <div>
                  <div className="text-gray-900 font-medium">附件信息</div>
                  <div className="text-xs text-gray-500">附件功能正在完善中，暂不支持下载</div>
                </div>
              </div>
            </div>
          </div>
        )}
      </div>

      {/* 分类信息 */}
      {email.category !== 'unclassified' && (
        <div className="bg-gradient-to-r from-primary-50 to-blue-50 border border-primary-100 rounded-xl p-5">
          <div className="flex items-center gap-2 text-primary-700 font-semibold mb-3">
            <div className="w-8 h-8 rounded-lg bg-primary-100 flex items-center justify-center">
              <Bot className="w-4 h-4 text-primary-600" />
            </div>
            AI分类结果
          </div>
          <div className="text-sm text-primary-800 ml-10">
            该邮件已被分类为 <strong>{CATEGORY_LABELS[email.category]}</strong>
            ，优先级为 <strong>{PRIORITY_LABELS[email.priority]}</strong>
            {email.confidence > 0 && (
              <span className="ml-2 text-primary-600">
                (置信度: {Math.round(email.confidence * 100)}%)
              </span>
            )}
          </div>
          {email.reasoning && (
            <div className="mt-3 ml-10 text-sm text-primary-700 bg-white/60 rounded-lg p-3">
              <span className="font-semibold">分析理由：</span>
              {email.reasoning}
            </div>
          )}
        </div>
      )}
    </div>
    </div>
  );
}
