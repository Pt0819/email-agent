import { useState, useEffect } from 'react';
import { Outlet, Link, useLocation, useNavigate } from 'react-router-dom';
import { Mail, Settings, LayoutDashboard, List, LogOut, User, Gamepad2 } from 'lucide-react';
import type { User as UserType } from '../../api/types';

export default function AppLayout() {
  const location = useLocation();
  const navigate = useNavigate();
  const [user, setUser] = useState<UserType | null>(null);

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

  const handleLogout = () => {
    localStorage.removeItem('token');
    localStorage.removeItem('user');
    navigate('/login');
  };

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Header */}
      <header className="bg-white border-b border-gray-200 sticky top-0 z-10">
        <div className="max-w-5xl mx-auto px-4 py-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              <Mail className="w-6 h-6 text-blue-600" />
              <h1 className="text-xl font-bold text-gray-900">邮件分类系统</h1>
            </div>

            <div className="flex items-center gap-4">
              {/* 导航 */}
              <nav className="flex items-center gap-1">
                <Link
                  to="/"
                  className={`p-2 rounded-lg transition-colors ${
                    isActive('/') && location.pathname === '/'
                      ? 'text-blue-600 bg-blue-50'
                      : 'text-gray-500 hover:text-gray-700 hover:bg-gray-100'
                  }`}
                  title="仪表盘"
                >
                  <LayoutDashboard className="w-5 h-5" />
                </Link>
                <Link
                  to="/emails"
                  className={`p-2 rounded-lg transition-colors ${
                    isActive('/emails')
                      ? 'text-blue-600 bg-blue-50'
                      : 'text-gray-500 hover:text-gray-700 hover:bg-gray-100'
                  }`}
                  title="邮件列表"
                >
                  <List className="w-5 h-5" />
                </Link>
                <Link
                  to="/steam/deals"
                  className={`p-2 rounded-lg transition-colors ${
                    isActive('/steam')
                      ? 'text-green-600 bg-green-50'
                      : 'text-gray-500 hover:text-gray-700 hover:bg-gray-100'
                  }`}
                  title="Steam促销"
                >
                  <Gamepad2 className="w-5 h-5" />
                </Link>
                <Link
                  to="/settings"
                  className={`p-2 rounded-lg transition-colors ${
                    isActive('/settings')
                      ? 'text-blue-600 bg-blue-50'
                      : 'text-gray-500 hover:text-gray-700 hover:bg-gray-100'
                  }`}
                  title="设置"
                >
                  <Settings className="w-5 h-5" />
                </Link>
              </nav>

              {/* 用户信息和登出 */}
              {user && (
                <div className="flex items-center gap-3 pl-4 border-l border-gray-200">
                  <div className="flex items-center gap-2">
                    <div className="w-8 h-8 bg-blue-100 rounded-full flex items-center justify-center">
                      <User className="w-4 h-4 text-blue-600" />
                    </div>
                    <div className="hidden sm:block">
                      <p className="text-sm font-medium text-gray-700">{user.username}</p>
                      <p className="text-xs text-gray-400">{user.email}</p>
                    </div>
                  </div>
                  <button
                    onClick={handleLogout}
                    className="p-2 text-gray-400 hover:text-red-600 hover:bg-red-50 rounded-lg transition-colors"
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
      <main className="max-w-5xl mx-auto px-4 py-6">
        <Outlet />
      </main>
    </div>
  );
}
