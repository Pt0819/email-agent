import { FILTERS } from '../../api/types';
import type { EmailCategory, EmailStatus, EmailAccount } from '../../api/types';
import { Search, Filter } from 'lucide-react';

interface FilterBarProps {
  selectedCategory: EmailCategory | 'all';
  selectedStatus: EmailStatus | 'all';
  selectedAccount: number | 'all';
  keyword: string;
  accounts: EmailAccount[];
  onCategoryChange: (category: EmailCategory | 'all') => void;
  onStatusChange: (status: EmailStatus | 'all') => void;
  onAccountChange: (accountId: number | 'all') => void;
  onKeywordChange: (keyword: string) => void;
  onSync?: () => void;
  syncStatus?: 'idle' | 'syncing' | 'error';
}

export default function FilterBar({
  selectedCategory,
  selectedStatus,
  selectedAccount,
  keyword,
  accounts,
  onCategoryChange,
  onStatusChange,
  onAccountChange,
  onKeywordChange,
  onSync,
  syncStatus,
}: FilterBarProps) {
  return (
    <div className="bg-white rounded-lg border border-gray-200 p-4 space-y-4">
      {/* 顶部：搜索和同步按钮 */}
      <div className="flex items-center gap-4">
        {/* 搜索框 */}
        <div className="flex-1 relative">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400" />
          <input
            type="text"
            placeholder="搜索邮件主题、发件人..."
            value={keyword}
            onChange={(e) => onKeywordChange(e.target.value)}
            className="w-full pl-10 pr-4 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-transparent"
          />
        </div>

        {/* 同步按钮 */}
        {onSync && (
          <button
            onClick={onSync}
            disabled={syncStatus === 'syncing'}
            className={`px-4 py-2 rounded-lg font-medium transition-colors ${
              syncStatus === 'syncing'
                ? 'bg-gray-100 text-gray-500 cursor-not-allowed'
                : 'bg-primary-600 text-white hover:bg-primary-700'
            }`}
          >
            {syncStatus === 'syncing' ? '同步中...' : '同步邮件'}
          </button>
        )}
      </div>

      {/* 底部：筛选器 */}
      <div className="flex items-center gap-6 flex-wrap">
        <div className="flex items-center gap-2">
          <Filter className="w-4 h-4 text-gray-500" />
          <span className="text-sm text-gray-700 font-medium">筛选:</span>
        </div>

        {/* 账户筛选 */}
        {accounts.length > 0 && (
          <div className="flex items-center gap-2">
            <span className="text-sm text-gray-600">账户:</span>
            <select
              value={selectedAccount}
              onChange={(e) => onAccountChange(e.target.value === 'all' ? 'all' : parseInt(e.target.value))}
              className="px-3 py-1.5 text-sm border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-primary-500"
            >
              <option value="all">全部账户</option>
              {accounts.map((acc) => (
                <option key={acc.id} value={acc.id}>
                  {acc.email || acc.account_email}
                </option>
              ))}
            </select>
          </div>
        )}

        {/* 分类筛选 */}
        <div className="flex items-center gap-2">
          <span className="text-sm text-gray-600">分类:</span>
          <select
            value={selectedCategory}
            onChange={(e) => onCategoryChange(e.target.value as EmailCategory | 'all')}
            className="px-3 py-1.5 text-sm border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-primary-500"
          >
            <option value="all">全部</option>
            {FILTERS.categories.map((cat) => (
              <option key={cat.value} value={cat.value}>
                {cat.label}
              </option>
            ))}
          </select>
        </div>

        {/* 状态筛选 */}
        <div className="flex items-center gap-2">
          <span className="text-sm text-gray-600">状态:</span>
          <select
            value={selectedStatus}
            onChange={(e) => onStatusChange(e.target.value as EmailStatus | 'all')}
            className="px-3 py-1.5 text-sm border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-primary-500"
          >
            <option value="all">全部</option>
            {FILTERS.statuses.map((status) => (
              <option key={status.value} value={status.value}>
                {status.label}
              </option>
            ))}
          </select>
        </div>
      </div>
    </div>
  );
}
