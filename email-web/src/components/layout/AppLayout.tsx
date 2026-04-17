import { Outlet, Link, useLocation } from 'react-router-dom';
import { Mail, Settings, LayoutDashboard, List } from 'lucide-react';

export default function AppLayout() {
  const location = useLocation();

  const isActive = (path: string) => {
    if (path === '/') {
      return location.pathname === '/';
    }
    return location.pathname.startsWith(path);
  };

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Header */}
      <header className="bg-white border-b border-gray-200 sticky top-0 z-10">
        <div className="max-w-5xl mx-auto px-4 py-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              <Mail className="w-6 h-6 text-primary-600" />
              <h1 className="text-xl font-bold text-gray-900">邮件分类系统</h1>
            </div>

            <nav className="flex items-center gap-1">
              <Link
                to="/"
                className={`p-2 rounded-lg transition-colors ${
                  isActive('/') && location.pathname === '/'
                    ? 'text-primary-600 bg-primary-50'
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
                    ? 'text-primary-600 bg-primary-50'
                    : 'text-gray-500 hover:text-gray-700 hover:bg-gray-100'
                }`}
                title="邮件列表"
              >
                <List className="w-5 h-5" />
              </Link>
              <Link
                to="/settings"
                className={`p-2 rounded-lg transition-colors ${
                  isActive('/settings')
                    ? 'text-primary-600 bg-primary-50'
                    : 'text-gray-500 hover:text-gray-700 hover:bg-gray-100'
                }`}
                title="设置"
              >
                <Settings className="w-5 h-5" />
              </Link>
            </nav>
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
