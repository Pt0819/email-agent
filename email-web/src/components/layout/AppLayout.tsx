import { useState, useEffect } from 'react';
import { Outlet, Link, useLocation, useNavigate } from 'react-router-dom';
import { Settings, LayoutDashboard, List, LogOut, Gamepad2, ChevronDown, Library, TrendingUp } from 'lucide-react';
import type { User as UserType } from '../../api/types';

// 预设的渐变配色方案（用于随机头像）
const AVATAR_GRADIENTS = [
  'from-pink-400 to-pink-600',
  'from-purple-400 to-purple-600',
  'from-indigo-400 to-indigo-600',
  'from-blue-400 to-blue-600',
  'from-cyan-400 to-cyan-600',
  'from-teal-400 to-teal-600',
  'from-emerald-400 to-emerald-600',
  'from-green-400 to-green-600',
  'from-lime-400 to-lime-600',
  'from-amber-400 to-amber-600',
  'from-orange-400 to-orange-600',
  'from-red-400 to-red-600',
];

// 根据用户名生成固定颜色
function getAvatarGradient(username: string): string {
  let hash = 0;
  for (let i = 0; i < username.length; i++) {
    hash = username.charCodeAt(i) + ((hash << 5) - hash);
  }
  return AVATAR_GRADIENTS[Math.abs(hash) % AVATAR_GRADIENTS.length];
}

// 获取用户名第一个字符（处理中英文）
function getUsernameFirstChar(username: string): string {
  if (!username) return '?';
  return username.charAt(0);
}

export default function AppLayout() {
  const location = useLocation();
  const navigate = useNavigate();
  const [user, setUser] = useState<UserType | null>(null);
  const [showSteamMenu, setShowSteamMenu] = useState(false);

  useEffect(() => {
    const userStr = localStorage.getItem('user');
    if (userStr) {
      try {
        setUser(JSON.parse(userStr));
      } catch {
        setUser(null);
      }
    }
  }, []);

  const isActive = (path: string) => {
    if (path === '/') {
      return location.pathname === '/';
    }
    return location.pathname.startsWith(path);
  };

  const isSteamActive = () => {
    return location.pathname.startsWith('/steam');
  };

  const handleLogout = () => {
    localStorage.removeItem('token');
    localStorage.removeItem('user');
    navigate('/login');
  };

  return (
    <div className="min-h-screen bg-gray-50/50">
      {/* Header */}
      <header className="glass border-b border-gray-100 sticky top-0 z-50">
        <div className="max-w-5xl mx-auto px-6">
          <div className="flex items-center justify-between h-16">
            {/* Logo */}
            <Link to="/" className="flex items-center gap-3 group">
              <img
                src="/logo.svg"
                alt="Mail Agent"
                className="w-9 h-9 object-contain group-hover:scale-105 transition-transform duration-200"
              />
              <span className="text-lg font-bold text-gray-900 group-hover:text-primary-600 transition-colors">Mail Agent</span>
            </Link>

            <div className="flex items-center gap-2">
              {/* 导航 */}
              <nav className="flex items-center gap-1 mr-2">
                <Link
                  to="/"
                  className={`relative p-2.5 rounded-xl transition-all duration-200 ${
                    isActive('/') && location.pathname === '/'
                      ? 'text-primary-600 bg-primary-50 shadow-sm'
                      : 'text-gray-500 hover:text-gray-700 hover:bg-gray-100'
                  }`}
                  title="仪表盘"
                >
                  <LayoutDashboard className="w-5 h-5" />
                  {isActive('/') && location.pathname === '/' && (
                    <span className="absolute bottom-0 left-1/2 -translate-x-1/2 w-5 h-0.5 bg-primary-600 rounded-full" />
                  )}
                </Link>
                <Link
                  to="/emails"
                  className={`relative p-2.5 rounded-xl transition-all duration-200 ${
                    isActive('/emails')
                      ? 'text-primary-600 bg-primary-50 shadow-sm'
                      : 'text-gray-500 hover:text-gray-700 hover:bg-gray-100'
                  }`}
                  title="邮件列表"
                >
                  <List className="w-5 h-5" />
                  {isActive('/emails') && (
                    <span className="absolute bottom-0 left-1/2 -translate-x-1/2 w-5 h-0.5 bg-primary-600 rounded-full" />
                  )}
                </Link>

                {/* Steam下拉菜单 */}
                <div className="relative">
                  <button
                    onClick={() => setShowSteamMenu(!showSteamMenu)}
                    className={`relative p-2.5 rounded-xl transition-all duration-200 flex items-center gap-1 ${
                      isSteamActive()
                        ? 'text-emerald-600 bg-emerald-50 shadow-sm'
                        : 'text-gray-500 hover:text-gray-700 hover:bg-gray-100'
                    }`}
                  >
                    <Gamepad2 className="w-5 h-5" />
                    <ChevronDown className={`w-3 h-3 transition-transform duration-200 ${showSteamMenu ? 'rotate-180' : ''}`} />
                    {isSteamActive() && (
                      <span className="absolute bottom-0 left-1/2 -translate-x-1/2 w-5 h-0.5 bg-emerald-600 rounded-full" />
                    )}
                  </button>

                  {showSteamMenu && (
                    <>
                      <div className="fixed inset-0 z-10" onClick={() => setShowSteamMenu(false)} />
                      <div className="absolute right-0 mt-2 w-48 bg-white rounded-xl shadow-lg border border-gray-100 py-1.5 z-20 animate-fade-in">
                        <Link
                          to="/steam/library"
                          onClick={() => setShowSteamMenu(false)}
                          className={`flex items-center gap-3 px-4 py-2.5 text-sm transition-colors ${
                            location.pathname === '/steam/library'
                              ? 'text-emerald-600 bg-emerald-50'
                              : 'text-gray-600 hover:bg-gray-50'
                          }`}
                        >
                          <Library className="w-4 h-4" />
                          游戏库
                        </Link>
                        <Link
                          to="/steam/deals"
                          onClick={() => setShowSteamMenu(false)}
                          className={`flex items-center gap-3 px-4 py-2.5 text-sm transition-colors ${
                            location.pathname === '/steam/deals'
                              ? 'text-emerald-600 bg-emerald-50'
                              : 'text-gray-600 hover:bg-gray-50'
                          }`}
                        >
                          <Gamepad2 className="w-4 h-4" />
                          促销信息
                        </Link>
                        <Link
                          to="/steam/profile"
                          onClick={() => setShowSteamMenu(false)}
                          className={`flex items-center gap-3 px-4 py-2.5 text-sm transition-colors ${
                            location.pathname === '/steam/profile'
                              ? 'text-emerald-600 bg-emerald-50'
                              : 'text-gray-600 hover:bg-gray-50'
                          }`}
                        >
                          <TrendingUp className="w-4 h-4" />
                          偏好画像
                        </Link>
                      </div>
                    </>
                  )}
                </div>

                <Link
                  to="/settings"
                  className={`relative p-2.5 rounded-xl transition-all duration-200 ${
                    isActive('/settings')
                      ? 'text-primary-600 bg-primary-50 shadow-sm'
                      : 'text-gray-500 hover:text-gray-700 hover:bg-gray-100'
                  }`}
                  title="设置"
                >
                  <Settings className="w-5 h-5" />
                  {isActive('/settings') && (
                    <span className="absolute bottom-0 left-1/2 -translate-x-1/2 w-5 h-0.5 bg-primary-600 rounded-full" />
                  )}
                </Link>
              </nav>

              {/* 用户信息和登出 */}
              {user && (
                <div className="flex items-center gap-3 pl-3 border-l border-gray-200">
                  <div className="flex items-center gap-2.5">
                    {/* 用户头像 - 圆形 */}
                    {user.avatar_url ? (
                      <img
                        src={user.avatar_url}
                        alt={user.username}
                        className="w-9 h-9 rounded-full object-cover shadow-sm"
                      />
                    ) : (
                      <div className={`w-9 h-9 rounded-full bg-gradient-to-br ${getAvatarGradient(user.username)} flex items-center justify-center shadow-sm`}>
                        <span className="text-white text-sm font-medium">
                          {getUsernameFirstChar(user.username)}
                        </span>
                      </div>
                    )}
                    <div className="hidden sm:block">
                      <p className="text-sm font-medium text-gray-800">{user.username}</p>
                      <p className="text-xs text-gray-400">{user.email}</p>
                    </div>
                  </div>
                  <button
                    onClick={handleLogout}
                    className="p-2 text-gray-400 hover:text-red-500 hover:bg-red-50 rounded-xl transition-all duration-200"
                    title="退出登录"
                  >
                    <LogOut className="w-5 h-5" />
                  </button>
                </div>
              )}
            </div>
          </div>
        </div>
      </header>

      {/* Main Content */}
      <main className="w-full">
        <Outlet />
      </main>
    </div>
  );
}
