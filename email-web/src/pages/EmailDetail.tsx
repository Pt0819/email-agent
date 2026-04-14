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
  Reply,
  Archive,
  Tag,
} from 'lucide-react';

export default function EmailDetail() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();

  const [email, setEmail] = useState<Email | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [classifying, setClassifying] = useState(false);

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
      // 重新获取邮件详情
      await fetchEmailDetail(email.id);
    } catch (err) {
      setError(err instanceof Error ? err.message : '分类失败');
    } finally {
      setClassifying(false);
    }
  };

  const handleMarkAsRead = async () => {
    // TODO: 实现标记已读功能
    console.log('Mark as read');
  };

  const handleReply = () => {
    // TODO: 实现回复功能
    console.log('Reply to email');
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

          <button
            onClick={handleMarkAsRead}
            className="flex items-center gap-2 px-4 py-2 border border-gray-300 rounded-lg hover:bg-gray-50 transition-colors"
          >
            <Check className="w-4 h-4" />
            标记已读
          </button>

          <button
            onClick={handleReply}
            className="flex items-center gap-2 px-4 py-2 border border-gray-300 rounded-lg hover:bg-gray-50 transition-colors"
          >
            <Reply className="w-4 h-4" />
            回复
          </button>

          <button
            onClick={() => {}}
            className="flex items-center gap-2 px-4 py-2 border border-gray-300 rounded-lg hover:bg-gray-50 transition-colors"
          >
            <Archive className="w-4 h-4" />
            归档
          </button>
        </div>
      </div>

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
                  <span className="text-xs font-mono">{email.message_id.slice(0, 16)}...</span>
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
            <div className="text-gray-400 text-center py-8">无正文内容</div>
          )}
        </div>

        {/* 附件区域 */}
        {email.has_attachment && (
          <div className="p-6 border-t border-gray-200 bg-gray-50">
            <div className="flex items-center gap-2 text-sm font-medium text-gray-700 mb-3">
              <Paperclip className="w-4 h-4" />
              附件
            </div>
            <div className="text-sm text-gray-500">
              附件功能待实现
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
          </div>
        </div>
      )}
    </div>
  );
}
