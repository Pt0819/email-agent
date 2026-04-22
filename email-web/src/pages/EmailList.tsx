import { useState, useEffect, useCallback } from 'react';
import { useNavigate } from 'react-router-dom';
import { emailApi, syncApi, accountApi } from '../api/client';
import type { Email, EmailCategory, EmailStatus, EmailAccount } from '../api/types';
import { AlertCircle, RefreshCw, Mail } from 'lucide-react';
import EmailCard from '../components/email/EmailCard';
import FilterBar from '../components/email/FilterBar';
import Pagination from '../components/ui/Pagination';

export default function EmailList() {
  const navigate = useNavigate();

  // 状态管理
  const [emails, setEmails] = useState<Email[]>([]);
  const [accounts, setAccounts] = useState<EmailAccount[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [total, setTotal] = useState(0);

  // 分页和筛选状态
  const [page, setPage] = useState(1);
  const pageSize = 20;
  const [selectedCategory, setSelectedCategory] = useState<EmailCategory | 'all'>('all');
  const [selectedStatus, setSelectedStatus] = useState<EmailStatus | 'all'>('all');
  const [selectedAccount, setSelectedAccount] = useState<number | 'all'>('all');
  const [keyword, setKeyword] = useState('');
  const [syncStatus, setSyncStatus] = useState<'idle' | 'syncing' | 'error'>('idle');

  // 获取账户列表
  const fetchAccounts = useCallback(async () => {
    try {
      const response = await accountApi.list();
      const data = response as unknown as { list: EmailAccount[] };
      setAccounts(data.list || []);
    } catch (err) {
      console.error('获取账户列表失败:', err);
    }
  }, []);

  // 获取邮件列表
  const fetchEmails = useCallback(async () => {
    try {
      setLoading(true);
      const params: {
        page: number;
        page_size: number;
        account_id?: number;
        category?: EmailCategory;
        status?: EmailStatus;
        keyword?: string;
      } = {
        page,
        page_size: pageSize,
        keyword: keyword || undefined,
      };

      if (selectedAccount !== 'all') {
        params.account_id = selectedAccount;
      }
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
  }, [page, pageSize, selectedCategory, selectedStatus, selectedAccount, keyword]);

  // 初始加载
  useEffect(() => {
    fetchAccounts();
  }, [fetchAccounts]);

  useEffect(() => {
    fetchEmails();
  }, [fetchEmails]);

  // 处理分类
  const handleClassify = async (id: string | number) => {
    try {
      await emailApi.classify(Number(id));
      fetchEmails();
    } catch (err) {
      setError(err instanceof Error ? err.message : '分类失败');
    }
  };

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

  // 处理账户变化
  const handleAccountChange = (accountId: number | 'all') => {
    setSelectedAccount(accountId);
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
    <div className="max-w-5xl mx-auto px-4 py-6 space-y-5 animate-fade-in">
      {/* 筛选栏 */}
      <FilterBar
        selectedCategory={selectedCategory}
        selectedStatus={selectedStatus}
        selectedAccount={selectedAccount}
        keyword={keyword}
        accounts={accounts}
        onCategoryChange={handleCategoryChange}
        onStatusChange={handleStatusChange}
        onAccountChange={handleAccountChange}
        onKeywordChange={handleKeywordChange}
        onSync={handleSync}
        syncStatus={syncStatus}
      />

      {/* 错误提示 */}
      {error && (
        <div className="flex items-center justify-center p-4 bg-red-50 border border-red-200 rounded-xl text-red-600">
          <AlertCircle className="w-5 h-5 mr-2" />
          {error}
        </div>
      )}

      {/* 统计信息 */}
      <div className="flex items-center justify-between text-sm text-gray-600">
        <span>共 <span className="font-semibold text-gray-900">{total}</span> 封邮件</span>
        {loading && (
          <span className="flex items-center gap-2 text-primary-600">
            <RefreshCw className="w-4 h-4 animate-spin" />
            加载中...
          </span>
        )}
      </div>

      {/* 邮件列表 */}
      {emails.length === 0 ? (
        <div className="card p-12">
          <div className="empty-state">
            <div className="empty-state-icon">
              <Mail className="w-10 h-10 text-gray-300" />
            </div>
            <p className="empty-state-title">暂无邮件</p>
            <p className="empty-state-desc">
              {keyword || selectedCategory !== 'all' || selectedStatus !== 'all' || selectedAccount !== 'all'
                ? '尝试调整筛选条件'
                : '请先添加邮箱账户并同步邮件'}
            </p>
          </div>
        </div>
      ) : (
        <>
          <div className="space-y-3">
            {emails.map((email, index) => (
              <div key={email.id} style={{ animationDelay: `${index * 30}ms` }} className="animate-fade-in">
                <EmailCard
                  email={email}
                  onClassify={handleClassify}
                  onView={handleView}
                />
              </div>
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
