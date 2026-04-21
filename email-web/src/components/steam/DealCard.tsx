import { useState } from 'react';
import { Tag, Clock, ExternalLink } from 'lucide-react';
import type { SteamDeal } from '../../api/types';

interface DealCardProps {
  deal: SteamDeal;
}

export default function DealCard({ deal }: DealCardProps) {
  const [expanded, setExpanded] = useState(false);

  // 计算剩余时间
  const getTimeRemaining = () => {
    if (!deal.end_date) return null;
    const end = new Date(deal.end_date);
    const now = new Date();
    const diff = end.getTime() - now.getTime();
    if (diff <= 0) return '已结束';
    const days = Math.floor(diff / (1000 * 60 * 60 * 24));
    const hours = Math.floor((diff % (1000 * 60 * 60 * 24)) / (1000 * 60 * 60));
    if (days > 0) return `${days}天${hours}小时`;
    return `${hours}小时`;
  };

  const timeRemaining = getTimeRemaining();
  const isExpiringSoon = timeRemaining && timeRemaining !== '已结束' && !timeRemaining.includes('天');

  return (
    <div className="bg-white rounded-lg border border-gray-200 overflow-hidden hover:shadow-md transition-shadow">
      <div className="flex">
        {/* 游戏封面 */}
        <div className="w-36 h-48 bg-gray-100 flex-shrink-0 relative">
          {deal.cover_url ? (
            <img
              src={deal.cover_url}
              alt={deal.game_name}
              className="w-full h-full object-cover"
              onError={(e) => {
                (e.target as HTMLImageElement).style.display = 'none';
              }}
            />
          ) : (
            <div className="w-full h-full flex items-center justify-center text-gray-400">
              <svg className="w-12 h-12" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M15 5v2m0 4v2m0 4v2M5 5a2 2 0 00-2 2v3a2 2 0 110 4v3a2 2 0 002 2h14a2 2 0 002-2v-3a2 2 0 110-4V7a2 2 0 00-2-2H5z" />
              </svg>
            </div>
          )}
          {/* 折扣标签 */}
          {deal.discount > 0 && (
            <div className="absolute top-2 left-2 bg-green-600 text-white px-2 py-1 rounded text-sm font-bold">
              -{deal.discount}%
            </div>
          )}
        </div>

        {/* 游戏信息 */}
        <div className="flex-1 p-4 flex flex-col justify-between">
          <div>
            <h3 className="font-semibold text-gray-900 text-lg leading-tight mb-1">
              {deal.game_name}
            </h3>

            {/* 价格信息 */}
            <div className="flex items-center gap-3 mt-2">
              {deal.discount > 0 && (
                <span className="text-gray-400 line-through text-sm">
                  ¥{deal.original_price.toFixed(2)}
                </span>
              )}
              <span className="text-green-600 font-bold text-xl">
                ¥{deal.deal_price.toFixed(2)}
              </span>
            </div>

            {/* 标签 */}
            <button
              onClick={() => setExpanded(!expanded)}
              className="mt-2 text-sm text-gray-500 hover:text-gray-700 flex items-center gap-1"
            >
              <Tag className="w-3 h-3" />
              {expanded ? '收起详情' : '查看详情'}
            </button>
          </div>

          {/* 底部信息 */}
          <div className="flex items-center justify-between mt-3">
            {/* 倒计时 */}
            {timeRemaining && (
              <div className={`flex items-center gap-1 text-sm ${
                isExpiringSoon ? 'text-red-600' : 'text-gray-500'
              }`}>
                <Clock className="w-3.5 h-3.5" />
                <span>{timeRemaining}</span>
              </div>
            )}

            {/* 状态标签 */}
            <div className="flex items-center gap-2">
              {deal.is_active ? (
                <span className="px-2 py-0.5 text-xs font-medium bg-green-100 text-green-800 rounded-full">
                  促销中
                </span>
              ) : (
                <span className="px-2 py-0.5 text-xs font-medium bg-gray-100 text-gray-600 rounded-full">
                  已结束
                </span>
              )}
            </div>
          </div>
        </div>
      </div>

      {/* 展开详情 */}
      {expanded && (
        <div className="px-4 pb-4 border-t border-gray-100 pt-3">
          <div className="grid grid-cols-2 gap-2 text-sm">
            <div>
              <span className="text-gray-500">游戏ID:</span>
              <span className="ml-2 text-gray-700">{deal.game_id || '未知'}</span>
            </div>
            <div>
              <span className="text-gray-500">折扣率:</span>
              <span className="ml-2 text-gray-700">{deal.discount}%</span>
            </div>
            <div>
              <span className="text-gray-500">原价:</span>
              <span className="ml-2 text-gray-700">¥{deal.original_price.toFixed(2)}</span>
            </div>
            <div>
              <span className="text-gray-500">现价:</span>
              <span className="ml-2 text-green-600 font-medium">¥{deal.deal_price.toFixed(2)}</span>
            </div>
          </div>

          {/* 商店链接 */}
          {deal.store_url && (
            <a
              href={deal.store_url}
              target="_blank"
              rel="noopener noreferrer"
              className="mt-3 inline-flex items-center gap-1 text-sm text-blue-600 hover:text-blue-800"
            >
              <ExternalLink className="w-3.5 h-3.5" />
              在Steam商店中查看
            </a>
          )}
        </div>
      )}
    </div>
  );
}
