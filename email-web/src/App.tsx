import { createBrowserRouter, RouterProvider, Navigate, Outlet } from 'react-router-dom';
import AppLayout from './components/layout/AppLayout';
import Dashboard from './pages/Dashboard';
import EmailList from './pages/EmailList';
import EmailDetail from './pages/EmailDetail';
import Settings from './pages/Settings';
import SteamDeals from './pages/SteamDeals';
import Login from './pages/Login';

// 路由守卫 - 检查是否已登录
function ProtectedRoute() {
  const token = localStorage.getItem('token');

  if (!token) {
    return <Navigate to="/login" replace />;
  }

  return <Outlet />;
}

const router = createBrowserRouter([
  {
    path: '/login',
    element: <Login />,
  },
  {
    element: <ProtectedRoute />,
    children: [
      {
        path: '/',
        element: <AppLayout />,
        children: [
          {
            index: true,
            element: <Dashboard />,
          },
          {
            path: 'emails',
            element: <EmailList />,
          },
          {
            path: 'emails/:id',
            element: <EmailDetail />,
          },
          {
            path: 'settings',
            element: <Settings />,
          },
          {
            path: 'steam/deals',
            element: <SteamDeals />,
          },
        ],
      },
    ],
  },
]);

export default function App() {
  return <RouterProvider router={router} />;
}
