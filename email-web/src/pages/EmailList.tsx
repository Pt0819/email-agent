import { useState, useEffect, useCallback } from 'react';
import { useNavigate } from 'react-router-dom';
import { emailApi, syncApi } from '../api/client';
import type { Email, EmailCategory, EmailStatus } from '../api/types';
import { AlertCircle, RefreshCw } from 'lucide-react';
import EmailCard from '../components/email/EmailCard';
import FilterBar from '../components/email/FilterBar';
import Pagination from '../components/ui/Pagination';

interface EmailListParams {
  page?: number;
  page_size?: number;
  category?: EmailCategory | 'all';
  status?: EmailStatus | 'all';
  keyword?: string;
}

export default function EmailList() {
  const navigate = useNavigate();

  // 状态管理
  const [emails, setEmails] = useState<Email[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [total, setTotal] = useState(0);

  // 分页和筛选状态
  const [page, setPage] = useState(1);
  const pageSize = 20;
  const [selectedCategory, setSelectedCategory] = useState<EmailCategory | 'all'>('all');
  const [selectedStatus, setSelectedStatus] = useState<EmailStatus | 'all'>('all');
  const [keyword, setKeyword] = useState('');
  const [syncStatus, setSyncStatus] = useState<'idle' | 'syncing' | 'error'>('idle');

  // 获取邮件列表
  const fetchEmails = useCallback(async () => {
    try {
      setLoading(true);
      const params: EmailListParams = {
        page,
        page_size: pageSize,
        keyword: keyword || undefined,
      };

      if (selectedCategory !== 'all') {
        params.category = selectedCategory as EmailCategory;
      }
      if (selectedStatus !== 'all') {
        params.status = selectedStatus as EmailStatus;
      }

      const response = await emailApi.list(params);
      const pageData = response as unknown as { list: Email[]; total: number };
      setEmails(pageData.list || []);
      setTotal(pageData.total || 0);
      setError(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : '获取邮件列表失败');
    } finally {
      setLoading(false);
    }
  }, [page, pageSize, selectedCategory, selectedStatus, keyword]);

  // 初始加载
  useEffect(() => {
    fetchEmails();
  }, [fetchEmails]);

  // 处理分类
  const handleClassify = async (id: string) => {
    try {
      await emailApi.classify(id);
      fetchEmails(); // 刷新列表
    } catch (err) {
      setError(err instanceof Error ? err.message : '分类失败');
    }
  };

  // 处理查看详情
  const handleView = (email: Email) => {
    navigate(`/emails/${email.id}`);
  };

  // 处理同步
  const handleSync = async () => {
    try {
      setSyncStatus('syncing');
      await syncApi.trigger();
      // 同步后刷新列表
      setTimeout(() => {
        fetchEmails();
        setSyncStatus('idle');
      }, 2000);
    } catch (err) {
      setSyncStatus('error');
      setError(err instanceof Error ? err.message : '同步失败');
      setTimeout(() => setSyncStatus('idle'), 3000);
    }
  };

  // 处理分类变化
  const handleCategoryChange = (category: EmailCategory | 'all') => {
    setSelectedCategory(category);
    setPage(1);
  };

  // 处理状态变化
  const handleStatusChange = (status: EmailStatus | 'all') => {
    setSelectedStatus(status);
    setPage(1);
  };

  // 处理关键词搜索
  const handleKeywordChange = (kw: string) => {
    setKeyword(kw);
    setPage(1);
  };

  // 加载状态
  if (loading && emails.length === 0) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary-600"></div>
      </div>
    );
  }

  return (
    <div className="space-y-4">
      {/* 筛选栏 */}
      <FilterBar
        selectedCategory={selectedCategory}
        selectedStatus={selectedStatus}
        keyword={keyword}
        onCategoryChange={handleCategoryChange}
        onStatusChange={handleStatusChange}
        onKeywordChange={handleKeywordChange}
        onSync={handleSync}
        syncStatus={syncStatus}
      />

      {/* 错误提示 */}
      {error && (
        <div className="flex items-center justify-center p-4 bg-red-50 border border-red-200 rounded-lg text-red-600">
          <AlertCircle className="w-5 h-5 mr-2" />
          {error}
        </div>
      )}

      {/* 统计信息 */}
      <div className="flex items-center justify-between text-sm text-gray-600">
        <span>共 {total} 封邮件</span>
        {loading && (
          <span className="flex items-center gap-2">
            <RefreshCw className="w-4 h-4 animate-spin" />
            加载中...
          </span>
        )}
      </div>

      {/* 邮件列表 */}
      {emails.length === 0 ? (
        <div className="text-center py-12 bg-white rounded-lg border border-gray-200">
          <div className="text-gray-400 mb-2">
            <svg className="w-16 h-16 mx-auto" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
            </svg>
          </div>
          <p className="text-gray-500 mb-1">暂无邮件</p>
          <p className="text-sm text-gray-400">
            {keyword || selectedCategory !== 'all' || selectedStatus !== 'all'
              ? '尝试调整筛选条件'
              : '请先添加邮箱账户并同步邮件'}
          </p>
        </div>
      ) : (
        <>
          <div className="space-y-3">
            {emails.map((email) => (
              <EmailCard
                key={email.id}
                email={email}
                onClassify={handleClassify}
                onView={handleView}
              />
            ))}
          </div>

          {/* 分页 */}
          <div className="pt-4">
            <Pagination
              current={page}
              total={total}
              pageSize={pageSize}
              onPageChange={setPage}
            />
          </div>
        </>
      )}
    </div>
  );
}
