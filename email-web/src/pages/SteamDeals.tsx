import { useState, useEffect } from 'react';
import { RefreshCw, Gamepad2, ArrowUpDown } from 'lucide-react';
import { steamApi } from '../../api/steamApi';
import DealCard from '../../components/steam/DealCard';
import Pagination from '../../components/ui/Pagination';
import type { SteamDeal, SteamStats } from '../../api/types';

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
    <div className="space-y-6">
      {/* 页面标题和统计 */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900 flex items-center gap-2">
            <Gamepad2 className="w-7 h-7 text-green-600" />
            Steam 促销
          </h1>
          <p className="mt-1 text-gray-500">
            自动从Steam邮件中提取的促销信息
          </p>
        </div>

        {stats && (
          <div className="flex gap-4">
            <div className="bg-white rounded-lg border px-4 py-2 text-center">
              <p className="text-2xl font-bold text-green-600">{stats.active_deals}</p>
              <p className="text-xs text-gray-500">活跃促销</p>
            </div>
            <div className="bg-white rounded-lg border px-4 py-2 text-center">
              <p className="text-2xl font-bold text-blue-600">{stats.total_games}</p>
              <p className="text-xs text-gray-500">已收录游戏</p>
            </div>
          </div>
        )}
      </div>

      {/* 筛选栏 */}
      <div className="bg-white rounded-lg border border-gray-200 p-4">
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
              className="border border-gray-300 rounded-lg px-3 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-green-500"
            >
              {SORT_OPTIONS.map((opt) => (
                <option key={opt.value} value={opt.value}>{opt.label}</option>
              ))}
            </select>
          </div>

          <div className="flex items-center gap-3">
            {/* 活跃筛选 */}
            <label className="flex items-center gap-2 text-sm text-gray-600 cursor-pointer">
              <input
                type="checkbox"
                checked={activeOnly}
                onChange={(e) => {
                  setActiveOnly(e.target.checked);
                  setPage(1);
                }}
                className="rounded border-gray-300 text-green-600 focus:ring-green-500"
              />
              仅显示促销中
            </label>

            {/* 刷新按钮 */}
            <button
              onClick={loadDeals}
              disabled={loading}
              className="p-2 text-gray-500 hover:text-green-600 hover:bg-green-50 rounded-lg transition-colors disabled:opacity-50"
              title="刷新"
            >
              <RefreshCw className={`w-4 h-4 ${loading ? 'animate-spin' : ''}`} />
            </button>
          </div>
        </div>
      </div>

      {/* 促销列表 */}
      {loading ? (
        <div className="flex justify-center py-12">
          <RefreshCw className="w-6 h-6 text-green-500 animate-spin" />
        </div>
      ) : deals.length === 0 ? (
        <div className="text-center py-12 bg-white rounded-lg border">
          <Gamepad2 className="w-12 h-12 text-gray-300 mx-auto mb-3" />
          <p className="text-gray-500">暂无促销信息</p>
          <p className="text-gray-400 text-sm mt-1">同步Steam邮件后将自动提取促销信息</p>
        </div>
      ) : (
        <>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            {deals.map((deal) => (
              <DealCard key={deal.id} deal={deal} />
            ))}
          </div>

          {/* 分页 */}
          {totalPages > 1 && (
            <Pagination
              currentPage={page}
              totalPages={totalPages}
              onPageChange={setPage}
            />
          )}
        </>
      )}
    </div>
  );
}
