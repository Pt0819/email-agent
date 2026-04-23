import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { authApi } from '../api/authApi';
import { Mail, Lock, User as UserIcon, Eye, EyeOff, Sparkles } from 'lucide-react';

export default function Login() {
  const navigate = useNavigate();
  const [isLogin, setIsLogin] = useState(true);
  const [showPassword, setShowPassword] = useState(false);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const [form, setForm] = useState({
    username: '',
    email: '',
    password: '',
    confirmPassword: '',
  });

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setForm({ ...form, [e.target.name]: e.target.value });
    setError('');
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setLoading(true);

    try {
      if (isLogin) {
        const response = await authApi.login({
          email: form.email,
          password: form.password,
        }) as unknown as { code: number; message: string; data: { token: string; user: object } };

        if (response.code === 0 && response.data) {
          localStorage.setItem('token', response.data.token);
          localStorage.setItem('user', JSON.stringify(response.data.user));
          navigate('/');
        } else {
          setError(response.message || '登录失败');
        }
      } else {
        if (form.password !== form.confirmPassword) {
          setError('两次输入的密码不一致');
          setLoading(false);
          return;
        }

        const response = await authApi.register({
          username: form.username,
          email: form.email,
          password: form.password,
        }) as unknown as { code: number; message: string; data: { token: string; user: object } };

        if (response.code === 0 && response.data) {
          localStorage.setItem('token', response.data.token);
          localStorage.setItem('user', JSON.stringify(response.data.user));
          navigate('/');
        } else {
          setError(response.message || '注册失败');
        }
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : '请求失败');
    } finally {
      setLoading(false);
    }
  };

  const toggleMode = () => {
    setIsLogin(!isLogin);
    setError('');
    setForm({ username: '', email: '', password: '', confirmPassword: '' });
  };

  return (
    <div className="min-h-screen flex">
      {/* 左侧装饰面板 */}
      <div className="hidden lg:flex lg:w-1/2 bg-gradient-to-br from-primary-600 via-primary-500 to-blue-500 relative overflow-hidden">
        {/* 背景装饰 */}
        <div className="absolute inset-0">
          <div className="absolute top-20 left-20 w-72 h-72 bg-white/10 rounded-full blur-3xl" />
          <div className="absolute bottom-20 right-20 w-96 h-96 bg-blue-400/20 rounded-full blur-3xl" />
          <div className="absolute top-1/2 left-1/3 w-64 h-64 bg-primary-300/20 rounded-full blur-2xl" />
        </div>

        {/* 内容 */}
        <div className="relative z-10 flex flex-col justify-center px-16 text-white">
          <div className="flex items-center gap-3 mb-8">
            <div className="w-14 h-14 bg-white/20 backdrop-blur-sm rounded-2xl flex items-center justify-center shadow-lg">
              <Sparkles className="w-7 h-7" />
            </div>
            <div>
              <h1 className="text-2xl font-bold">Mail Agent</h1>
              <p className="text-white/70 text-sm">AI驱动的邮件智能助手</p>
            </div>
          </div>

          <h2 className="text-4xl font-bold leading-tight mb-6">
            让AI帮你管理<br />繁琐的邮件
          </h2>

          <p className="text-white/80 text-lg leading-relaxed mb-10">
            智能分类、自动摘要、优先级排序。<br />
            告别邮件混乱，专注于真正重要的事。
          </p>

          <div className="space-y-4">
            {[
              { label: '智能分类', desc: 'AI自动识别邮件类型' },
              { label: 'Steam资讯', desc: '游戏促销一手掌握' },
              { label: '每日摘要', desc: '关键信息不再遗漏' },
            ].map((item, i) => (
              <div key={i} className="flex items-center gap-4">
                <div className="w-10 h-10 bg-white/15 rounded-xl flex items-center justify-center text-sm font-bold backdrop-blur-sm">
                  {i + 1}
                </div>
                <div>
                  <p className="font-semibold">{item.label}</p>
                  <p className="text-white/60 text-sm">{item.desc}</p>
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>

      {/* 右侧登录面板 */}
      <div className="flex-1 flex items-center justify-center p-8 bg-gradient-to-b from-gray-50 to-white">
        <div className="w-full max-w-md animate-fade-in">
          {/* 移动端Logo */}
          <div className="lg:hidden text-center mb-8">
            <div className="inline-flex items-center justify-center w-14 h-14 bg-gradient-to-br from-primary-500 to-primary-600 rounded-2xl mb-4 shadow-lg">
              <Sparkles className="w-7 h-7 text-white" />
            </div>
            <h1 className="text-2xl font-bold text-gray-900">Mail Agent</h1>
            <p className="text-gray-500 mt-1">AI驱动的邮件智能助手</p>
          </div>

          {/* 表单卡片 */}
          <div className="card p-8 shadow-lg">
            {/* 标题 */}
            <div className="text-center mb-8">
              <h2 className="text-2xl font-bold text-gray-900">
                {isLogin ? '欢迎回来' : '创建账号'}
              </h2>
              <p className="text-gray-500 mt-2">
                {isLogin ? '登录您的账户继续使用' : '注册新账户开始体验'}
              </p>
            </div>

            {/* 切换标签 */}
            <div className="flex mb-8 bg-gray-100 rounded-xl p-1">
              <button
                type="button"
                onClick={() => setIsLogin(true)}
                className={`flex-1 py-2.5 text-sm font-medium rounded-lg transition-all duration-200 ${
                  isLogin
                    ? 'bg-white text-gray-900 shadow-sm'
                    : 'text-gray-500 hover:text-gray-700'
                }`}
              >
                登录
              </button>
              <button
                type="button"
                onClick={() => setIsLogin(false)}
                className={`flex-1 py-2.5 text-sm font-medium rounded-lg transition-all duration-200 ${
                  !isLogin
                    ? 'bg-white text-gray-900 shadow-sm'
                    : 'text-gray-500 hover:text-gray-700'
                }`}
              >
                注册
              </button>
            </div>

            {/* 错误提示 */}
            {error && (
              <div className="mb-6 p-4 bg-red-50 border border-red-100 text-red-600 text-sm rounded-xl flex items-center gap-3">
                <div className="w-8 h-8 bg-red-100 rounded-lg flex items-center justify-center flex-shrink-0">
                  <span className="text-lg font-bold">!</span>
                </div>
                <span>{error}</span>
              </div>
            )}

            {/* 表单 */}
            <form onSubmit={handleSubmit} className="space-y-5">
              {/* 用户名（仅注册） */}
              {!isLogin && (
                <div className="animate-slide-in">
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    用户名
                  </label>
                  <div className="relative">
                    <UserIcon className="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5 text-gray-400" />
                    <input
                      type="text"
                      name="username"
                      value={form.username}
                      onChange={handleChange}
                      placeholder="请输入用户名"
                      required={!isLogin}
                      minLength={2}
                      maxLength={50}
                      className="w-full pl-12 pr-3 py-3 border border-gray-200 rounded-lg text-gray-900 placeholder-gray-400 bg-white transition-all duration-200 focus:outline-none focus:ring-2 focus:ring-primary-500/20 focus:border-primary-500"
                    />
                  </div>
                </div>
              )}

              {/* 邮箱 */}
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  邮箱地址
                </label>
                <div className="relative">
                  <Mail className="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5 text-gray-400" />
                  <input
                    type="email"
                    name="email"
                    value={form.email}
                    onChange={handleChange}
                    placeholder="请输入邮箱"
                    required
                    className="w-full pl-12 pr-3 py-3 border border-gray-200 rounded-lg text-gray-900 placeholder-gray-400 bg-white transition-all duration-200 focus:outline-none focus:ring-2 focus:ring-primary-500/20 focus:border-primary-500"
                  />
                </div>
              </div>

              {/* 密码 */}
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  密码
                </label>
                <div className="relative">
                  <Lock className="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5 text-gray-400" />
                  <input
                    type={showPassword ? 'text' : 'password'}
                    name="password"
                    value={form.password}
                    onChange={handleChange}
                    placeholder={isLogin ? '请输入密码' : '请设置密码（至少6位）'}
                    required
                    minLength={6}
                    className="w-full pl-12 pr-10 py-3 border border-gray-200 rounded-lg text-gray-900 placeholder-gray-400 bg-white transition-all duration-200 focus:outline-none focus:ring-2 focus:ring-primary-500/20 focus:border-primary-500"
                  />
                  <button
                    type="button"
                    onClick={() => setShowPassword(!showPassword)}
                    className="absolute right-4 top-1/2 -translate-y-1/2 text-gray-400 hover:text-gray-600 transition-colors"
                  >
                    {showPassword ? <EyeOff className="w-5 h-5" /> : <Eye className="w-5 h-5" />}
                  </button>
                </div>
              </div>

              {/* 确认密码（仅注册） */}
              {!isLogin && (
                <div className="animate-slide-in">
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    确认密码
                  </label>
                  <div className="relative">
                    <Lock className="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5 text-gray-400" />
                    <input
                      type={showPassword ? 'text' : 'password'}
                      name="confirmPassword"
                      value={form.confirmPassword}
                      onChange={handleChange}
                      placeholder="请再次输入密码"
                      required={!isLogin}
                      className="w-full pl-12 pr-3 py-3 border border-gray-200 rounded-lg text-gray-900 placeholder-gray-400 bg-white transition-all duration-200 focus:outline-none focus:ring-2 focus:ring-primary-500/20 focus:border-primary-500"
                    />
                  </div>
                </div>
              )}

              {/* 提交按钮 */}
              <button
                type="submit"
                disabled={loading}
                className="w-full py-3.5 btn-primary shadow-glow text-lg"
              >
                {loading ? (
                  <div className="w-5 h-5 border-2 border-white/30 border-t-white rounded-full animate-spin" />
                ) : (
                  <>
                    {isLogin ? '登录' : '创建账号'}
                  </>
                )}
              </button>
            </form>

            {/* 切换提示 */}
            <p className="mt-8 text-center text-sm text-gray-500">
              {isLogin ? '还没有账号？' : '已有账号？'}
              <button
                type="button"
                onClick={toggleMode}
                className="ml-1.5 text-primary-600 hover:text-primary-700 font-semibold transition-colors"
              >
                {isLogin ? '立即注册' : '立即登录'}
              </button>
            </p>
          </div>

          {/* 底部版权 */}
          <p className="mt-8 text-center text-xs text-gray-400">
            © 2026 Mail Agent. All rights reserved.
          </p>
        </div>
      </div>
    </div>
  );
}
