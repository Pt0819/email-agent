import { useState, useEffect } from 'react';
import { emailApi } from '../api/client';
import { CATEGORY_LABELS, CATEGORY_COLORS, PRIORITY_LABELS, PRIORITY_COLORS } from '../api/types';
import type { Email } from '../api/types';
import { Mail, User, Clock, AlertCircle } from 'lucide-react';

export default function EmailList() {
  const [emails, setEmails] = useState<Email[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    fetchEmails();
  }, []);

  const fetchEmails = async () => {
    try {
      setLoading(true);
      const response = await emailApi.list({ page: 1, page_size: 20 });
      // 响应拦截器已解包，response就是data
      const pageData = response as unknown as { list: Email[]; total: number };
      setEmails(pageData.list || []);
      setError(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : '获取邮件列表失败');
    } finally {
      setLoading(false);
    }
  };

  const handleClassify = async (id: string) => {
    try {
      await emailApi.classify(id);
      fetchEmails(); // 刷新列表
    } catch (err) {
      setError(err instanceof Error ? err.message : '分类失败');
    }
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary-600"></div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="flex items-center justify-center h-64 text-red-600">
        <AlertCircle className="w-5 h-5 mr-2" />
        {error}
      </div>
    );
  }

  return (
    <div className="space-y-4">
      <div className="flex justify-between items-center">
        <h2 className="text-xl font-semibold">邮件列表</h2>
        <button
          onClick={fetchEmails}
          className="px-4 py-2 text-sm bg-primary-600 text-white rounded-lg hover:bg-primary-700"
        >
          刷新
        </button>
      </div>

      {emails.length === 0 ? (
        <div className="text-center py-12 text-gray-500">
          暂无邮件，请先添加邮箱账户
        </div>
      ) : (
        <div className="space-y-3">
          {emails.map((email) => (
            <div
              key={email.id}
              className="p-4 bg-white rounded-lg border border-gray-200 hover:border-primary-300 transition-colors"
            >
              <div className="flex items-start justify-between">
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-2 mb-2">
                    <Mail className="w-4 h-4 text-gray-400" />
                    <span className={`px-2 py-0.5 text-xs rounded-full border ${CATEGORY_COLORS[email.category]}`}>
                      {CATEGORY_LABELS[email.category]}
                    </span>
                    <span className={`text-xs font-medium ${PRIORITY_COLORS[email.priority]}`}>
                      {PRIORITY_LABELS[email.priority]}
                    </span>
                  </div>

                  <h3 className="text-lg font-medium text-gray-900 truncate mb-1">
                    {email.subject || '(无主题)'}
                  </h3>

                  <div className="flex items-center gap-4 text-sm text-gray-500">
                    <div className="flex items-center gap-1">
                      <User className="w-4 h-4" />
                      <span>{email.sender_name || email.sender_email}</span>
                    </div>
                    <div className="flex items-center gap-1">
                      <Clock className="w-4 h-4" />
                      <span>{new Date(email.received_at).toLocaleString()}</span>
                    </div>
                  </div>
                </div>

                <button
                  onClick={() => handleClassify(email.id)}
                  className="ml-4 px-3 py-1.5 text-sm border border-gray-300 rounded-lg hover:bg-gray-50"
                >
                  分类
                </button>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}