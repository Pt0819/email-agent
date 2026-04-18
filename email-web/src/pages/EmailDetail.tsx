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
      const response = await emailApi.getById(emailId);
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
      await fetchEmailDetail(email.id);
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
    <div className="max-w-4xl mx-auto space-y-6">
      {/* 顶部操作栏 */}
      <div className="flex items-center justify-between">
        <button
          onClick={() => navigate(-1)}
          className="flex items-center gap-2 text-gray-600 hover:text-gray-900 transition-colors"
        >
          <ArrowLeft className="w-5 h-5" />
          返回
        </button>

        <div className="flex items-center gap-2">
          {email.category === 'unclassified' && (
            <button
              onClick={handleClassify}
              disabled={classifying}
              className={`flex items-center gap-2 px-4 py-2 rounded-lg font-medium transition-colors ${
                classifying
                  ? 'bg-gray-100 text-gray-500 cursor-not-allowed'
                  : 'bg-primary-600 text-white hover:bg-primary-700'
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
              className="flex items-center gap-2 px-4 py-2 border border-gray-300 rounded-lg hover:bg-gray-50 transition-colors disabled:opacity-50"
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
              className="flex items-center gap-2 px-4 py-2 border border-gray-300 rounded-lg hover:bg-gray-50 transition-colors disabled:opacity-50"
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
        <div className="bg-amber-50 border border-amber-200 rounded-lg p-3 flex items-center gap-2">
          <Inbox className="w-4 h-4 text-amber-600" />
          <span className="text-sm text-amber-800">此邮件已归档</span>
        </div>
      )}

      {/* 邮件主体 */}
      <div className="bg-white rounded-lg border border-gray-200 overflow-hidden">
        {/* 邮件头部 */}
        <div className="p-6 border-b border-gray-200">
          {/* 标签行 */}
          <div className="flex items-center gap-2 mb-4 flex-wrap">
            <span
              className={`px-3 py-1 text-sm rounded-full border ${CATEGORY_COLORS[email.category]}`}
            >
              <Tag className="w-3 h-3 inline mr-1" />
              {CATEGORY_LABELS[email.category]}
            </span>
            <span
              className={`px-3 py-1 text-sm rounded-full border ${
                email.priority === 'critical'
                  ? 'bg-red-100 text-red-800 border-red-200'
                  : email.priority === 'high'
                  ? 'bg-orange-100 text-orange-800 border-orange-200'
                  : email.priority === 'medium'
                  ? 'bg-yellow-100 text-yellow-800 border-yellow-200'
                  : 'bg-gray-100 text-gray-800 border-gray-200'
              }`}
            >
              {PRIORITY_LABELS[email.priority]}
            </span>
            {email.has_attachment && (
              <span className="flex items-center gap-1 px-3 py-1 text-sm rounded-full border border-gray-200 text-gray-700">
                <Paperclip className="w-3 h-3" />
                有附件
              </span>
            )}
            {!isRead && (
              <span className="px-3 py-1 text-sm rounded-full border border-blue-200 bg-blue-50 text-blue-700">
                未读
              </span>
            )}
          </div>

          {/* 主题 */}
          <h1 className="text-2xl font-bold text-gray-900 mb-4">
            {email.subject || '(无主题)'}
          </h1>

          {/* 发件人信息 */}
          <div className="space-y-2 text-sm">
            <div className="flex items-center gap-3">
              <div className="w-10 h-10 rounded-full bg-gradient-to-br from-primary-500 to-primary-600 flex items-center justify-center text-white font-medium">
                {(email.sender_name || email.sender_email).charAt(0).toUpperCase()}
              </div>
              <div className="flex-1">
                <div className="font-medium text-gray-900">
                  {email.sender_name || email.sender_email}
                </div>
                <div className="text-gray-500">{email.sender_email}</div>
              </div>
            </div>

            <div className="flex items-center gap-6 text-gray-500 ml-13">
              <div className="flex items-center gap-1">
                <Clock className="w-4 h-4" />
                <span>{new Date(email.received_at).toLocaleString('zh-CN')}</span>
              </div>
              {email.message_id && (
                <div className="flex items-center gap-1">
                  <Mail className="w-4 h-4" />
                  <span className="text-xs font-mono">{email.message_id.slice(0, 30)}...</span>
                </div>
              )}
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
            <div className="text-gray-400 text-center py-8 flex flex-col items-center gap-2">
              <FileText className="w-8 h-8" />
              无正文内容
            </div>
          )}
        </div>

        {/* 附件区域 */}
        {email.has_attachment && (
          <div className="p-6 border-t border-gray-200 bg-gray-50">
            <div className="flex items-center gap-2 text-sm font-medium text-gray-700 mb-3">
              <Paperclip className="w-4 h-4" />
              附件
            </div>
            <div className="bg-white border border-gray-200 rounded-lg p-4">
              <div className="flex items-center gap-3 text-sm text-gray-600">
                <div className="w-10 h-10 bg-gray-100 rounded-lg flex items-center justify-center">
                  <FileText className="w-5 h-5 text-gray-400" />
                </div>
                <div>
                  <div className="text-gray-900">附件信息</div>
                  <div className="text-xs text-gray-500">附件功能正在完善中，暂不支持下载</div>
                </div>
              </div>
            </div>
          </div>
        )}
      </div>

      {/* 分类信息 */}
      {email.category !== 'unclassified' && (
        <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
          <div className="flex items-center gap-2 text-blue-900 font-medium mb-2">
            <Bot className="w-4 h-4" />
            AI分类结果
          </div>
          <div className="text-sm text-blue-800">
            该邮件已被分类为 <strong>{CATEGORY_LABELS[email.category]}</strong>
            ，优先级为 <strong>{PRIORITY_LABELS[email.priority]}</strong>
            {email.confidence > 0 && (
              <span className="ml-2 text-blue-600">
                (置信度: {Math.round(email.confidence * 100)}%)
              </span>
            )}
          </div>
          {email.reasoning && (
            <div className="mt-2 text-sm text-blue-700 bg-blue-100/50 rounded-lg p-3">
              <span className="font-medium">分析理由：</span>
              {email.reasoning}
            </div>
          )}
        </div>
      )}
    </div>
  );
}
