import { useState, useEffect } from 'react';
import { RefreshCw, Gamepad2, Link, Unlink, Clock, TrendingUp, User } from 'lucide-react';
import { steamApi } from '../api/steamApi';
import type { SteamAccount, SteamLibraryItem, SteamStats } from '../api/types';

type SortOption = 'playtime' | 'name' | 'recent';

const SORT_OPTIONS: { value: SortOption; label: string }[] = [
  { value: 'playtime', label: '游玩时长' },
  { value: 'name', label: '游戏名称' },
  { value: 'recent', label: '最近游玩' },
];

// 格式化游玩时长
const formatPlaytime = (minutes: number): string => {
  if (minutes < 60) return `${minutes}分钟`;
  const hours = Math.floor(minutes / 60);
  if (hours < 1000) return `${hours}小时`;
  const days = Math.floor(hours / 24);
  if (days < 30) return `${days}天${hours % 24}小时`;
  const months = Math.floor(days / 30);
  return `${months}月${days % 30}天`;
};

export default function SteamLibrary() {
  const [account, setAccount] = useState<SteamAccount | null>(null);
  const [library, setLibrary] = useState<SteamLibraryItem[]>([]);
  const [recentGames, setRecentGames] = useState<SteamLibraryItem[]>([]);
  const [stats, setStats] = useState<SteamStats | null>(null);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [sortBy, setSortBy] = useState<SortOption>('playtime');
  const [loading, setLoading] = useState(false);
  const [syncing, setSyncing] = useState(false);
  const [bindInput, setBindInput] = useState('');
  const [binding, setBinding] = useState(false);

  const pageSize = 20;

  useEffect(() => {
    loadProfile();
    loadLibrary();
    loadRecent();
    loadStats();
  }, []);

  useEffect(() => {
    if (account) {
      loadLibrary();
    }
  }, [page, sortBy, account]);

  const loadProfile = async () => {
    try {
      const res = await steamApi.getProfile();
      setAccount((res.data as any) || null);
    } catch {
      setAccount(null);
    }
  };

  const loadLibrary = async () => {
    if (!account) return;
    setLoading(true);
    try {
      const res = await steamApi.listLibrary({
        page,
        page_size: pageSize,
        sort: sortBy,
      });
      const data = (res.data as any) || {};
      setLibrary(data.list || []);
      setTotal(data.total || 0);
    } catch (err) {
      console.error('加载游戏库失败:', err);
    } finally {
      setLoading(false);
    }
  };

  const loadRecent = async () => {
    if (!account) return;
    try {
      const res = await steamApi.listRecent({ limit: 6 });
      const data = (res.data as any) || {};
      setRecentGames(data.list || []);
    } catch {
      // ignore
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

  const handleBind = async () => {
    if (!bindInput.trim()) return;
    setBinding(true);
    try {
      await steamApi.bind({ steam_id: bindInput.trim() });
      setBindInput('');
      await loadProfile();
      await loadLibrary();
      await loadRecent();
      await loadStats();
    } catch (err: any) {
      alert('绑定失败: ' + (err.message || '未知错误'));
    } finally {
      setBinding(false);
    }
  };

  const handleUnbind = async () => {
    if (!confirm('确定要解绑Steam账号吗？')) return;
    try {
      await steamApi.unbind();
      setAccount(null);
      setLibrary([]);
      setRecentGames([]);
      await loadStats();
    } catch (err: any) {
      alert('解绑失败: ' + (err.message || '未知错误'));
    }
  };

  const handleSync = async () => {
    setSyncing(true);
    try {
      await steamApi.syncLibrary();
      await loadLibrary();
      await loadRecent();
      await loadStats();
    } catch (err: any) {
      alert('同步失败: ' + (err.message || '未知错误'));
    } finally {
      setSyncing(false);
    }
  };

  const totalPages = Math.ceil(total / pageSize);

  return (
    <div className="space-y-6">
      {/* 页面标题 */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900 flex items-center gap-2">
            <Gamepad2 className="w-7 h-7 text-blue-600" />
            Steam 游戏库
          </h1>
          <p className="mt-1 text-gray-500">
            {account ? '已同步您的Steam游戏库' : '绑定Steam账号，同步游戏库和游玩记录'}
          </p>
        </div>

        {account && (
          <div className="flex gap-4">
            <div className="bg-white rounded-lg border px-4 py-2 text-center">
              <p className="text-2xl font-bold text-blue-600">{stats?.total_games || 0}</p>
              <p className="text-xs text-gray-500">游戏库</p>
            </div>
            <div className="bg-white rounded-lg border px-4 py-2 text-center">
              <p className="text-2xl font-bold text-green-600">{recentGames.length}</p>
              <p className="text-xs text-gray-500">近两周游玩</p>
            </div>
          </div>
        )}
      </div>

      {/* Steam账号绑定卡片 */}
      {account ? (
        <div className="bg-white rounded-lg border p-6">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-4">
              {account.avatar_url ? (
                <img
                  src={account.avatar_url}
                  alt={account.steam_nickname}
                  className="w-16 h-16 rounded-full"
                />
              ) : (
                <div className="w-16 h-16 rounded-full bg-gray-200 flex items-center justify-center">
                  <User className="w-8 h-8 text-gray-400" />
                </div>
              )}
              <div>
                <h2 className="text-lg font-semibold text-gray-900">
                  {account.steam_nickname || 'Steam玩家'}
                </h2>
                <p className="text-sm text-gray-500">{account.steam_id}</p>
                {account.last_sync_at && (
                  <p className="text-xs text-gray-400 mt-1">
                    最后同步: {account.last_sync_at}
                  </p>
                )}
              </div>
            </div>
            <div className="flex items-center gap-2">
              <button
                onClick={handleSync}
                disabled={syncing}
                className="flex items-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50 transition-colors"
              >
                <RefreshCw className={`w-4 h-4 ${syncing ? 'animate-spin' : ''}`} />
                同步游戏库
              </button>
              <button
                onClick={handleUnbind}
                className="flex items-center gap-2 px-4 py-2 text-red-600 border border-red-200 rounded-lg hover:bg-red-50 transition-colors"
              >
                <Unlink className="w-4 h-4" />
                解绑
              </button>
            </div>
          </div>
        </div>
      ) : (
        <div className="bg-white rounded-lg border p-6">
          <div className="flex items-center gap-4 mb-4">
            <div className="w-12 h-12 rounded-full bg-blue-100 flex items-center justify-center">
              <Gamepad2 className="w-6 h-6 text-blue-600" />
            </div>
            <div>
              <h2 className="text-lg font-semibold text-gray-900">绑定Steam账号</h2>
              <p className="text-sm text-gray-500">输入您的Steam ID即可同步游戏库（使用Mock数据）</p>
            </div>
          </div>
          <div className="flex gap-3">
            <input
              type="text"
              value={bindInput}
              onChange={(e) => setBindInput(e.target.value)}
              placeholder="输入Steam ID（如 76561198012345678）"
              className="flex-1 border border-gray-300 rounded-lg px-4 py-2 focus:outline-none focus:ring-2 focus:ring-blue-500"
              onKeyDown={(e) => e.key === 'Enter' && handleBind()}
            />
            <button
              onClick={handleBind}
              disabled={binding || !bindInput.trim()}
              className="flex items-center gap-2 px-6 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50 transition-colors"
            >
              <Link className="w-4 h-4" />
              绑定
            </button>
          </div>
        </div>
      )}

      {/* 最近游玩 */}
      {account && recentGames.length > 0 && (
        <div className="bg-white rounded-lg border p-6">
          <h3 className="text-lg font-semibold text-gray-900 mb-4 flex items-center gap-2">
            <Clock className="w-5 h-5 text-green-600" />
            近两周游玩
          </h3>
          <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-6 gap-4">
            {recentGames.map((game) => (
              <div key={game.id} className="text-center">
                <div className="w-full aspect-square bg-gray-100 rounded-lg mb-2 flex items-center justify-center overflow-hidden">
                  {game.icon_url ? (
                    <img
                      src={`https://media.steampowered.com/steamcommunity/public/images/apps/${game.game_id}/${game.icon_url}.jpg`}
                      alt={game.game_name}
                      className="w-full h-full object-cover"
                      onError={(e) => {
                        (e.target as HTMLImageElement).style.display = 'none';
                      }}
                    />
                  ) : (
                    <Gamepad2 className="w-8 h-8 text-gray-300" />
                  )}
                </div>
                <p className="text-sm font-medium text-gray-900 truncate">{game.game_name}</p>
                <p className="text-xs text-green-600">{formatPlaytime(game.playtime_2_weeks)}</p>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* 游戏库列表 */}
      {account && (
        <>
          <div className="bg-white rounded-lg border p-4">
            <div className="flex items-center justify-between gap-4">
              <div className="flex items-center gap-2">
                <TrendingUp className="w-4 h-4 text-gray-400" />
                <select
                  value={sortBy}
                  onChange={(e) => {
                    setSortBy(e.target.value as SortOption);
                    setPage(1);
                  }}
                  className="border border-gray-300 rounded-lg px-3 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                >
                  {SORT_OPTIONS.map((opt) => (
                    <option key={opt.value} value={opt.value}>{opt.label}</option>
                  ))}
                </select>
              </div>
              <span className="text-sm text-gray-500">{total} 款游戏</span>
            </div>
          </div>

          {loading ? (
            <div className="flex justify-center py-12">
              <RefreshCw className="w-6 h-6 text-blue-500 animate-spin" />
            </div>
          ) : library.length === 0 ? (
            <div className="text-center py-12 bg-white rounded-lg border">
              <Gamepad2 className="w-12 h-12 text-gray-300 mx-auto mb-3" />
              <p className="text-gray-500">游戏库为空</p>
              <p className="text-gray-400 text-sm mt-1">点击"同步游戏库"按钮获取数据</p>
            </div>
          ) : (
            <>
              <div className="bg-white rounded-lg border overflow-hidden">
                <table className="w-full">
                  <thead className="bg-gray-50 border-b">
                    <tr>
                      <th className="px-4 py-3 text-left text-sm font-medium text-gray-600">游戏</th>
                      <th className="px-4 py-3 text-right text-sm font-medium text-gray-600">总游玩时长</th>
                      <th className="px-4 py-3 text-right text-sm font-medium text-gray-600">近两周</th>
                      <th className="px-4 py-3 text-right text-sm font-medium text-gray-600">最后游玩</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y">
                    {library.map((game) => (
                      <tr key={game.id} className="hover:bg-gray-50">
                        <td className="px-4 py-3">
                          <div className="flex items-center gap-3">
                            <div className="w-10 h-10 bg-gray-100 rounded flex-shrink-0 flex items-center justify-center overflow-hidden">
                              {game.icon_url ? (
                                <img
                                  src={`https://media.steampowered.com/steamcommunity/public/images/apps/${game.game_id}/${game.icon_url}.jpg`}
                                  alt={game.game_name}
                                  className="w-full h-full object-cover"
                                  onError={(e) => {
                                    (e.target as HTMLImageElement).style.display = 'none';
                                  }}
                                />
                              ) : (
                                <Gamepad2 className="w-5 h-5 text-gray-300" />
                              )}
                            </div>
                            <span className="font-medium text-gray-900">{game.game_name}</span>
                          </div>
                        </td>
                        <td className="px-4 py-3 text-right text-gray-700">{formatPlaytime(game.playtime)}</td>
                        <td className="px-4 py-3 text-right text-green-600">{formatPlaytime(game.playtime_2_weeks)}</td>
                        <td className="px-4 py-3 text-right text-gray-500 text-sm">
                          {game.last_played_at ? new Date(game.last_played_at).toLocaleDateString() : '-'}
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>

              {/* 分页 */}
              {totalPages > 1 && (
                <div className="flex justify-center gap-2 mt-4">
                  {Array.from({ length: Math.min(totalPages, 5) }, (_, i) => i + 1).map((p) => (
                    <button
                      key={p}
                      onClick={() => setPage(p)}
                      className={`w-10 h-10 rounded-lg border ${
                        p === page
                          ? 'bg-blue-600 text-white border-blue-600'
                          : 'hover:bg-gray-50'
                      }`}
                    >
                      {p}
                    </button>
                  ))}
                  {totalPages > 5 && <span className="px-2 py-2 text-gray-500">...</span>}
                </div>
              )}
            </>
          )}
        </>
      )}

      {/* 未绑定时显示示例数据 */}
      {!account && (
        <div className="bg-gray-50 rounded-lg border border-dashed p-8 text-center">
          <Gamepad2 className="w-12 h-12 text-gray-300 mx-auto mb-3" />
          <p className="text-gray-500">绑定Steam账号后即可查看您的游戏库</p>
        </div>
      )}
    </div>
  );
}
