import { useState, useEffect } from 'react';
import { RefreshCw, Gamepad2, ArrowUpDown } from 'lucide-react';
import { steamApi } from '../api/steamApi';
import DealCard from '../components/steam/DealCard';
import Pagination from '../components/ui/Pagination';
import type { SteamDeal, SteamStats } from '../api/types';

type SortOption = 'created_at' | 'discount' | 'price_asc' | 'price_desc' | 'end_date';

const SORT_OPTIONS: { value: SortOption; label: string }[] = [
  { value: 'created_at', label: '最新' },
  { value: 'discount', label: '折扣最高' },
  { value: 'price_asc', label: '价格从低到高' },
  { value: 'price_desc', label: '价格从高到低' },
  { value: 'end_date', label: '即将结束' },
];

export default function SteamDeals() {
  const [deals, setDeals] = useState<SteamDeal[]>([]);
  const [stats, setStats] = useState<SteamStats | null>(null);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [sortBy, setSortBy] = useState<SortOption>('created_at');
  const [activeOnly, setActiveOnly] = useState(true);
  const [loading, setLoading] = useState(false);

  const pageSize = 12;

  useEffect(() => {
    loadDeals();
  }, [page, sortBy, activeOnly]);

  useEffect(() => {
    loadStats();
  }, []);

  const loadDeals = async () => {
    setLoading(true);
    try {
      const res = await steamApi.listDeals({
        page,
        page_size: pageSize,
        sort: sortBy,
        active: activeOnly ? 'true' : 'false',
      });
      const data = res.data as any;
      setDeals(data?.list || []);
      setTotal(data?.total || 0);
    } catch (err) {
      console.error('加载促销列表失败:', err);
    } finally {
      setLoading(false);
    }
  };

  const loadStats = async () => {
    try {
      const res = await steamApi.getStats();
      setStats((res.data as any) || null);
    } catch {
      // ignore
    }
  };

  const totalPages = Math.ceil(total / pageSize);

  return (
    <div className="max-w-5xl mx-auto px-4 py-6 space-y-6 animate-fade-in">
      {/* 页面标题和统计 */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900 flex items-center gap-2">
            <Gamepad2 className="w-7 h-7 text-emerald-600" />
            Steam 促销
          </h1>
          <p className="mt-1 text-gray-500">
            自动从Steam邮件中提取的促销信息
          </p>
        </div>

        {stats && (
          <div className="flex gap-4">
            <div className="card px-5 py-3 text-center hover-lift">
              <p className="text-2xl font-bold text-emerald-600">{stats.active_deals}</p>
              <p className="text-xs text-gray-500 mt-0.5">活跃促销</p>
            </div>
            <div className="card px-5 py-3 text-center hover-lift">
              <p className="text-2xl font-bold text-blue-600">{stats.total_games}</p>
              <p className="text-xs text-gray-500 mt-0.5">已收录游戏</p>
            </div>
          </div>
        )}
      </div>

      {/* 筛选栏 */}
      <div className="card p-4">
        <div className="flex items-center justify-between gap-4 flex-wrap">
          {/* 排序选择 */}
          <div className="flex items-center gap-2">
            <ArrowUpDown className="w-4 h-4 text-gray-400" />
            <select
              value={sortBy}
              onChange={(e) => {
                setSortBy(e.target.value as SortOption);
                setPage(1);
              }}
              className="border border-gray-200 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-emerald-500/20 focus:border-emerald-500 bg-white"
            >
              {SORT_OPTIONS.map((opt) => (
                <option key={opt.value} value={opt.value}>{opt.label}</option>
              ))}
            </select>
          </div>

          <div className="flex items-center gap-3">
            {/* 活跃筛选 */}
            <label className="flex items-center gap-2 text-sm text-gray-600 cursor-pointer group">
              <input
                type="checkbox"
                checked={activeOnly}
                onChange={(e) => {
                  setActiveOnly(e.target.checked);
                  setPage(1);
                }}
                className="w-4 h-4 rounded border-gray-300 text-emerald-600 focus:ring-emerald-500 cursor-pointer"
              />
              <span className="group-hover:text-gray-900 transition-colors">仅显示促销中</span>
            </label>

            {/* 刷新按钮 */}
            <button
              onClick={loadDeals}
              disabled={loading}
              className="icon-btn hover:text-emerald-600 hover:bg-emerald-50"
              title="刷新"
            >
              <RefreshCw className={`w-4 h-4 ${loading ? 'animate-spin' : ''}`} />
            </button>
          </div>
        </div>
      </div>

      {/* 促销列表 */}
      {loading ? (
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {[1, 2, 3, 4].map((i) => (
            <div key={i} className="skeleton-card h-40" />
          ))}
        </div>
      ) : deals.length === 0 ? (
        <div className="empty-state py-16">
          <div className="empty-state-icon bg-emerald-50">
            <Gamepad2 className="w-10 h-10 text-emerald-300" />
          </div>
          <p className="empty-state-title">暂无促销信息</p>
          <p className="empty-state-desc">同步Steam邮件后将自动提取促销信息</p>
        </div>
      ) : (
        <>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            {deals.map((deal, index) => (
              <div key={deal.id} style={{ animationDelay: `${index * 50}ms` }} className="animate-fade-in">
                <DealCard deal={deal} />
              </div>
            ))}
          </div>

          {/* 分页 */}
          {totalPages > 1 && (
            <Pagination
              current={page}
              total={total}
              pageSize={pageSize}
              onPageChange={setPage}
            />
          )}
        </>
      )}
    </div>
  );
}
