import { useState, useEffect, useRef } from 'react';
import { accountApi, syncApi } from '../api/client';
import { authApi } from '../api/authApi';
import type { EmailAccount, EmailProvider, SchedulerStatus, User, ApiResponse } from '../api/types';
import {
  Plus, Trash2, TestTube, CheckCircle, XCircle, AlertCircle,
  Loader2, Play, Pause, Clock, Settings, RefreshCw, Mail,
  Camera, Save, Key, ChevronDown,
} from 'lucide-react';

type TabType = 'sync' | 'profile';

export default function SettingsPage() {
  const [activeTab, setActiveTab] = useState<TabType>('sync');

  // ===== 同步设置状态 =====
  const [accounts, setAccounts] = useState<EmailAccount[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // 表单状态
  const [showForm, setShowForm] = useState(false);
  const [formData, setFormData] = useState({ email: '', provider: '126' as EmailProvider, credential: '' });
  const [submitting, setSubmitting] = useState(false);

  // 测试连接状态
  const [testingAccountId, setTestingAccountId] = useState<number | null>(null);
  const [testResult, setTestResult] = useState<Record<number, { success: boolean; message: string }>>({});

  // 调度器状态
  const [scheduler, setScheduler] = useState<SchedulerStatus | null>(null);
  const [schedulerLoading, setSchedulerLoading] = useState(false);
  const [intervalInput, setIntervalInput] = useState(5);

  // ===== 个人资料状态 =====
  const [user, setUser] = useState<User | null>(null);
  const [usernameInput, setUsernameInput] = useState('');
  const [profileLoading, setProfileLoading] = useState(false);
  const [profileSuccess, setProfileSuccess] = useState<string | null>(null);

  // 密码修改状态
  const [showPasswordForm, setShowPasswordForm] = useState(false);
  const [passwordData, setPasswordData] = useState({ old_password: '', new_password: '', confirm_password: '' });
  const [passwordLoading, setPasswordLoading] = useState(false);
  const [passwordSuccess, setPasswordSuccess] = useState<string | null>(null);
  const [passwordError, setPasswordError] = useState<string | null>(null);

  // 头像上传
  const fileInputRef = useRef<HTMLInputElement>(null);
  const [avatarLoading, setAvatarLoading] = useState(false);

  useEffect(() => {
    fetchAccounts();
    fetchSchedulerStatus();
    fetchUserProfile();
  }, []);

  // ===== 同步设置方法 =====
  const fetchAccounts = async () => {
    try {
      setLoading(true);
      const response = await accountApi.list();
      const data = response as unknown as { list: EmailAccount[] };
      setAccounts(data.list || []);
      setError(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : '获取账户列表失败');
    } finally {
      setLoading(false);
    }
  };

  const fetchSchedulerStatus = async () => {
    try {
      const response = await syncApi.schedulerStatus();
      const data = response as unknown as SchedulerStatus;
      setScheduler(data);
      if (data.interval) setIntervalInput(data.interval);
    } catch { /* 调度器状态获取失败不影响主流程 */ }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      setSubmitting(true);
      await accountApi.create({ email: formData.email, provider: formData.provider, credential: formData.credential });
      setFormData({ email: '', provider: '126', credential: '' });
      setShowForm(false);
      fetchAccounts();
    } catch (err) {
      setError(err instanceof Error ? err.message : '添加账户失败');
    } finally {
      setSubmitting(false);
    }
  };

  const handleDelete = async (id: number) => {
    if (!confirm('确定要删除此账户吗？相关邮件数据也会被删除。')) return;
    try {
      await accountApi.delete(id);
      fetchAccounts();
    } catch (err) {
      setError(err instanceof Error ? err.message : '删除账户失败');
    }
  };

  const handleTest = async (id: number) => {
    try {
      setTestingAccountId(id);
      const response = await accountApi.test(id);
      const data = response as unknown as { success: boolean; message: string };
      setTestResult({ ...testResult, [id]: { success: data.success, message: data.message } });
    } catch (err) {
      setTestResult({ ...testResult, [id]: { success: false, message: err instanceof Error ? err.message : '测试失败' } });
    } finally {
      setTestingAccountId(null);
    }
  };

  const handleToggleScheduler = async () => {
    setSchedulerLoading(true);
    try {
      if (scheduler?.running) {
        await syncApi.stopScheduler();
      } else {
        await syncApi.startScheduler();
      }
      await fetchSchedulerStatus();
    } catch (err) {
      setError(err instanceof Error ? err.message : '操作失败');
    } finally {
      setSchedulerLoading(false);
    }
  };

  const handleSetInterval = async () => {
    if (intervalInput < 1 || intervalInput > 1440) return;
    try {
      await syncApi.setInterval({ interval: intervalInput });
      await fetchSchedulerStatus();
    } catch (err) {
      setError(err instanceof Error ? err.message : '设置间隔失败');
    }
  };

  // ===== 个人资料方法 =====
  const fetchUserProfile = async () => {
    try {
      const response = await authApi.me();
      const apiResponse = response as unknown as ApiResponse<User>;
      if (apiResponse.data) {
        setUser(apiResponse.data);
        setUsernameInput(apiResponse.data.username);
      }
    } catch {
      // 从localStorage获取
      const userStr = localStorage.getItem('user');
      if (userStr) {
        const localUser = JSON.parse(userStr) as User;
        setUser(localUser);
        setUsernameInput(localUser.username);
      }
    }
  };

  const handleUpdateProfile = async () => {
    if (!usernameInput.trim()) return;
    setProfileLoading(true);
    setProfileSuccess(null);
    try {
      const response = await authApi.updateProfile({ username: usernameInput });
      const apiResponse = response as unknown as ApiResponse<User>;
      if (apiResponse.data) {
        setUser(apiResponse.data);
        localStorage.setItem('user', JSON.stringify(apiResponse.data));
      }
      setProfileSuccess('用户名更新成功');
      setTimeout(() => setProfileSuccess(null), 3000);
    } catch (err) {
      setError(err instanceof Error ? err.message : '更新失败');
    } finally {
      setProfileLoading(false);
    }
  };

  const handleChangePassword = async (e: React.FormEvent) => {
    e.preventDefault();
    setPasswordError(null);
    setPasswordSuccess(null);

    if (passwordData.new_password !== passwordData.confirm_password) {
      setPasswordError('两次输入的密码不一致');
      return;
    }

    if (passwordData.new_password.length < 6) {
      setPasswordError('密码长度至少6位');
      return;
    }

    setPasswordLoading(true);
    try {
      await authApi.changePassword(passwordData);
      setPasswordSuccess('密码修改成功');
      setPasswordData({ old_password: '', new_password: '', confirm_password: '' });
      setTimeout(() => setPasswordSuccess(null), 3000);
    } catch (err) {
      setPasswordError(err instanceof Error ? err.message : '修改失败');
    } finally {
      setPasswordLoading(false);
    }
  };

  const handleAvatarClick = () => {
    fileInputRef.current?.click();
  };

  const handleAvatarChange = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;

    // 验证文件类型
    const allowedTypes = ['image/jpeg', 'image/png', 'image/gif', 'image/webp'];
    if (!allowedTypes.includes(file.type)) {
      setError('仅支持 JPG、PNG、GIF、WebP 格式');
      return;
    }

    // 验证文件大小 (2MB)
    if (file.size > 2 * 1024 * 1024) {
      setError('图片大小不能超过 2MB');
      return;
    }

    setAvatarLoading(true);
    try {
      const response = await authApi.uploadAvatar(file);
      const apiResponse = response as unknown as ApiResponse<User>;
      if (apiResponse.data) {
        setUser(apiResponse.data);
        localStorage.setItem('user', JSON.stringify(apiResponse.data));
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : '上传失败');
    } finally {
      setAvatarLoading(false);
      // 清空input
      if (fileInputRef.current) {
        fileInputRef.current.value = '';
      }
    }
  };

  // 获取账户显示邮箱（兼容字段名）
  const getEmail = (a: EmailAccount) => a.account_email || '';

  // 获取用户名首字
  const getInitials = (name: string) => {
    return name.slice(0, 2).toUpperCase();
  };

  return (
    <div className="max-w-5xl mx-auto px-4 py-6">
      <div className="space-y-6 animate-fade-in">
        {/* 页面标题 */}
        <div className="flex items-center justify-between">
          <h2 className="text-xl font-bold text-gray-900">设置</h2>
        </div>

        {/* Tab 导航 */}
        <div className="flex gap-1 p-1 bg-gray-100 rounded-xl w-fit">
          <button
            onClick={() => setActiveTab('sync')}
            className={`px-5 py-2.5 rounded-lg font-medium text-sm transition-all duration-200 ${
              activeTab === 'sync'
                ? 'bg-white text-gray-900 shadow-sm'
                : 'text-gray-500 hover:text-gray-700'
            }`}
          >
            同步设置
          </button>
          <button
            onClick={() => setActiveTab('profile')}
            className={`px-5 py-2.5 rounded-lg font-medium text-sm transition-all duration-200 ${
              activeTab === 'profile'
                ? 'bg-white text-gray-900 shadow-sm'
                : 'text-gray-500 hover:text-gray-700'
            }`}
          >
            个人资料
          </button>
        </div>

        {/* 错误提示 */}
        {error && (
          <div className="flex items-center p-4 bg-red-50 border border-red-200 rounded-xl text-red-600">
            <AlertCircle className="w-5 h-5 mr-3 flex-shrink-0" />
            <span>{error}</span>
            <button onClick={() => setError(null)} className="ml-auto text-red-400 hover:text-red-600 font-bold">&times;</button>
          </div>
        )}

        {/* ===== 同步设置 Tab ===== */}
        {activeTab === 'sync' && (
          <div className="space-y-6">
            <div className="flex justify-end">
              <button onClick={() => setShowForm(!showForm)} className="btn-primary shadow-glow">
                <Plus className="w-4 h-4" /> 添加账户
              </button>
            </div>

            {/* 同步调度器面板 */}
            <div className="card">
              <div className="p-4 border-b border-gray-100 flex items-center gap-3">
                <div className="w-8 h-8 rounded-lg bg-primary-50 flex items-center justify-center">
                  <Settings className="w-4 h-4 text-primary-600" />
                </div>
                <h3 className="font-semibold text-gray-900">同步调度器</h3>
              </div>
              <div className="p-5 space-y-5">
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-3">
                    <div className={`w-3 h-3 rounded-full ${scheduler?.running ? 'bg-emerald-500 animate-pulse' : 'bg-gray-300'}`} />
                    <div>
                      <span className="font-medium text-gray-900">{scheduler?.running ? '运行中' : '已停止'}</span>
                      {scheduler?.running && scheduler.next_sync_time && (
                        <span className="text-sm text-gray-500 ml-3">
                          下次同步: {new Date(scheduler.next_sync_time).toLocaleTimeString()}
                        </span>
                      )}
                    </div>
                  </div>
                  <button onClick={handleToggleScheduler} disabled={schedulerLoading}
                    className={`flex items-center gap-2 px-4 py-2.5 rounded-xl font-medium transition-all duration-200 ${
                      scheduler?.running
                        ? 'bg-red-50 text-red-700 hover:bg-red-100 border border-red-200'
                        : 'btn-primary shadow-glow'
                    }`}>
                    {schedulerLoading ? <Loader2 className="w-4 h-4 animate-spin" /> : scheduler?.running ? <Pause className="w-4 h-4" /> : <Play className="w-4 h-4" />}
                    {scheduler?.running ? '停止' : '启动'}
                  </button>
                </div>

                {/* 同步间隔设置 */}
                <div className="flex items-center gap-3 p-4 bg-gray-50/80 rounded-xl">
                  <Clock className="w-5 h-5 text-gray-400" />
                  <span className="text-sm text-gray-600">同步间隔:</span>
                  <input type="number" min={1} max={1440} value={intervalInput}
                    onChange={(e) => setIntervalInput(Number(e.target.value))}
                    className="w-20 px-2 py-1.5 border border-gray-200 rounded-lg text-center text-sm bg-white focus:outline-none focus:ring-2 focus:ring-primary-500/20 focus:border-primary-500" />
                  <span className="text-sm text-gray-500">分钟</span>
                  <button onClick={handleSetInterval}
                    className="px-4 py-1.5 text-sm bg-white border border-gray-200 rounded-lg hover:bg-gray-50 hover:border-gray-300 transition-colors">
                    应用
                  </button>
                </div>

                {/* 统计信息 */}
                {scheduler && (
                  <div className="grid grid-cols-3 gap-4">
                    <div className="card p-4 text-center hover-lift">
                      <div className="text-2xl font-bold text-primary-600">{scheduler.sync_count}</div>
                      <div className="text-xs text-gray-500 mt-1">同步次数</div>
                    </div>
                    <div className="card p-4 text-center hover-lift">
                      <div className="text-2xl font-bold text-red-600">{scheduler.error_count}</div>
                      <div className="text-xs text-gray-500 mt-1">错误次数</div>
                    </div>
                    <div className="card p-4 text-center hover-lift">
                      <div className="text-2xl font-bold text-gray-600">{scheduler.interval}</div>
                      <div className="text-xs text-gray-500 mt-1">间隔(分钟)</div>
                    </div>
                  </div>
                )}
              </div>
            </div>

            {/* 添加账户表单 */}
            {showForm && (
              <div className="card p-6 animate-slide-in">
                <h3 className="text-lg font-semibold text-gray-900 mb-5">添加邮箱账户</h3>
                <form onSubmit={handleSubmit} className="space-y-5">
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-2">邮箱地址</label>
                    <input type="email" required placeholder="yourname@126.com" value={formData.email}
                      onChange={(e) => setFormData({ ...formData, email: e.target.value })}
                      className="w-full px-3 py-2.5 border border-gray-200 rounded-lg text-gray-900 placeholder-gray-400 bg-white transition-all duration-200 focus:outline-none focus:ring-2 focus:ring-primary-500/20 focus:border-primary-500" />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-2">邮箱服务商</label>
                    <select value={formData.provider} onChange={(e) => setFormData({ ...formData, provider: e.target.value as EmailProvider })}
                      className="w-full px-3 py-2.5 border border-gray-200 rounded-lg text-gray-900 bg-white transition-all duration-200 focus:outline-none focus:ring-2 focus:ring-primary-500/20 focus:border-primary-500">
                      <option value="126">网易126邮箱</option>
                      <option value="gmail">Gmail</option>
                      <option value="outlook">Outlook</option>
                      <option value="imap">通用IMAP</option>
                    </select>
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-2">授权码</label>
                    <input type="password" required placeholder="请输入邮箱授权码（非登录密码）" value={formData.credential}
                      onChange={(e) => setFormData({ ...formData, credential: e.target.value })}
                      className="w-full px-3 py-2.5 border border-gray-200 rounded-lg text-gray-900 placeholder-gray-400 bg-white transition-all duration-200 focus:outline-none focus:ring-2 focus:ring-primary-500/20 focus:border-primary-500" />
                    <p className="mt-2 text-xs text-gray-500">
                      {formData.provider === '126'
                        ? '126邮箱授权码获取方式：设置 → POP3/SMTP/IMAP → 开启服务 → 生成授权码'
                        : '请输入邮箱服务商提供的授权码或应用密码'}
                    </p>
                  </div>
                  <div className="flex gap-3 pt-2">
                    <button type="submit" disabled={submitting}
                      className={`flex-1 py-3 rounded-xl font-medium transition-all duration-200 ${
                        submitting ? 'bg-gray-100 text-gray-500 cursor-not-allowed' : 'btn-primary shadow-glow'
                      }`}>
                      {submitting ? <span className="flex items-center justify-center gap-2"><Loader2 className="w-4 h-4 animate-spin" />添加中...</span> : '添加账户'}
                    </button>
                    <button type="button" onClick={() => setShowForm(false)}
                      className="px-6 py-3 border border-gray-200 rounded-xl hover:bg-gray-50 transition-colors">取消</button>
                  </div>
                </form>
              </div>
            )}

            {/* 账户列表 */}
            <div className="card overflow-hidden">
              <div className="p-4 border-b border-gray-100 flex items-center justify-between">
                <h3 className="font-semibold text-gray-900">已添加的账户</h3>
                <button onClick={fetchAccounts} className="icon-btn">
                  <RefreshCw className="w-4 h-4" />
                </button>
              </div>
              {loading ? (
                <div className="flex items-center justify-center h-32">
                  <div className="spinner w-6 h-6" />
                </div>
              ) : accounts.length === 0 ? (
                <div className="empty-state py-12">
                  <div className="empty-state-icon">
                    <Mail className="w-8 h-8 text-gray-300" />
                  </div>
                  <p className="empty-state-title">暂无账户</p>
                  <p className="empty-state-desc">请添加邮箱账户开始使用</p>
                </div>
              ) : (
                <div className="divide-y divide-gray-50">
                  {accounts.map((account) => {
                    const result = testResult[account.id];
                    const email = getEmail(account);
                    return (
                      <div key={account.id} className="p-5 hover:bg-gray-50/50 transition-colors">
                        <div className="flex items-center justify-between">
                          <div className="flex items-center gap-4 flex-1">
                            <div className="w-12 h-12 rounded-xl bg-gradient-to-br from-primary-400 to-primary-500 flex items-center justify-center text-white font-semibold shadow-md">
                              {email.charAt(0).toUpperCase()}
                            </div>
                            <div>
                              <div className="font-medium text-gray-900 text-base">{email}</div>
                              <div className="text-sm text-gray-500 mt-0.5">
                                {account.provider.toUpperCase()} · {account.sync_enabled ? '已启用' : '已禁用'} · {account.last_sync_at ? `最后同步: ${new Date(account.last_sync_at).toLocaleString()}` : '尚未同步'}
                              </div>
                            </div>
                          </div>
                          <div className="flex items-center gap-2">
                            <button onClick={() => handleTest(account.id)} disabled={testingAccountId === account.id}
                              className={`icon-btn ${testingAccountId === account.id ? 'opacity-50 cursor-not-allowed' : 'hover:text-primary-600 hover:bg-primary-50'}`}
                              title="测试连接">
                              {testingAccountId === account.id ? <Loader2 className="w-5 h-5 animate-spin" /> : <TestTube className="w-5 h-5" />}
                            </button>
                            <button onClick={() => handleDelete(account.id)}
                              className="icon-btn hover:text-red-600 hover:bg-red-50" title="删除账户">
                              <Trash2 className="w-5 h-5" />
                            </button>
                          </div>
                        </div>
                        {result && (
                          <div className={`mt-3 text-sm flex items-center gap-2 p-3 rounded-lg ${
                            result.success ? 'bg-emerald-50 text-emerald-700' : 'bg-red-50 text-red-700'
                          }`}>
                            {result.success ? <CheckCircle className="w-4 h-4" /> : <XCircle className="w-4 h-4" />}
                            {result.message}
                          </div>
                        )}
                      </div>
                    );
                  })}
                </div>
              )}
            </div>
          </div>
        )}

        {/* ===== 个人资料 Tab ===== */}
        {activeTab === 'profile' && user && (
          <div className="space-y-6">
            {/* 头像区域 */}
            <div className="card p-6">
              <h3 className="font-semibold text-gray-900 mb-5">头像设置</h3>
              <div className="flex items-center gap-6">
                {/* 头像展示 */}
                <div className="relative group">
                  {user.avatar_url ? (
                    <img
                      src={user.avatar_url}
                      alt={user.username}
                      className="w-20 h-20 rounded-full object-cover shadow-md"
                    />
                  ) : (
                    <div className="w-20 h-20 rounded-full bg-gradient-to-br from-primary-400 to-primary-500 flex items-center justify-center shadow-md">
                      <span className="text-white text-xl font-medium">
                        {getInitials(user.username)}
                      </span>
                    </div>
                  )}
                  {/* 上传按钮 */}
                  <button
                    onClick={handleAvatarClick}
                    disabled={avatarLoading}
                    className="absolute inset-0 rounded-full bg-black/40 opacity-0 group-hover:opacity-100 transition-opacity flex items-center justify-center"
                  >
                    {avatarLoading ? (
                      <Loader2 className="w-6 h-6 text-white animate-spin" />
                    ) : (
                      <Camera className="w-6 h-6 text-white" />
                    )}
                  </button>
                  <input
                    ref={fileInputRef}
                    type="file"
                    accept="image/jpeg,image/png,image/gif,image/webp"
                    onChange={handleAvatarChange}
                    className="hidden"
                  />
                </div>
                <div>
                  <p className="text-sm text-gray-600">点击头像更换图片</p>
                  <p className="text-xs text-gray-400 mt-1">支持 JPG、PNG、GIF、WebP，最大 2MB</p>
                </div>
              </div>
            </div>

            {/* 用户资料 */}
            <div className="card p-6">
              <h3 className="font-semibold text-gray-900 mb-5">基本信息</h3>

              {profileSuccess && (
                <div className="mb-4 p-3 bg-emerald-50 text-emerald-700 rounded-lg flex items-center gap-2">
                  <CheckCircle className="w-4 h-4" />
                  {profileSuccess}
                </div>
              )}

              <div className="space-y-5">
                {/* 用户名 */}
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">用户名</label>
                  <div className="flex gap-3">
                    <input
                      type="text"
                      value={usernameInput}
                      onChange={(e) => setUsernameInput(e.target.value)}
                      className="flex-1 px-3 py-2.5 border border-gray-200 rounded-lg text-gray-900 bg-white focus:outline-none focus:ring-2 focus:ring-primary-500/20 focus:border-primary-500"
                    />
                    <button
                      onClick={handleUpdateProfile}
                      disabled={profileLoading || usernameInput === user.username}
                      className={`px-4 py-2.5 rounded-lg font-medium transition-all duration-200 flex items-center gap-2 ${
                        profileLoading || usernameInput === user.username
                          ? 'bg-gray-100 text-gray-400 cursor-not-allowed'
                          : 'btn-primary shadow-glow'
                      }`}
                    >
                      {profileLoading ? <Loader2 className="w-4 h-4 animate-spin" /> : <Save className="w-4 h-4" />}
                      保存
                    </button>
                  </div>
                </div>

                {/* 邮箱（只读） */}
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">邮箱</label>
                  <input
                    type="email"
                    value={user.email}
                    disabled
                    className="w-full px-3 py-2.5 border border-gray-200 rounded-lg text-gray-500 bg-gray-50 cursor-not-allowed"
                  />
                  <p className="text-xs text-gray-400 mt-1">邮箱地址不可修改</p>
                </div>

                {/* 注册时间（只读） */}
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">注册时间</label>
                  <input
                    type="text"
                    value={new Date(user.created_at).toLocaleString()}
                    disabled
                    className="w-full px-3 py-2.5 border border-gray-200 rounded-lg text-gray-500 bg-gray-50 cursor-not-allowed"
                  />
                </div>
              </div>
            </div>

            {/* 密码修改 */}
            <div className="card overflow-hidden">
              <button
                onClick={() => setShowPasswordForm(!showPasswordForm)}
                className="w-full p-4 flex items-center justify-between hover:bg-gray-50 transition-colors"
              >
                <div className="flex items-center gap-3">
                  <div className="w-8 h-8 rounded-lg bg-amber-50 flex items-center justify-center">
                    <Key className="w-4 h-4 text-amber-600" />
                  </div>
                  <span className="font-semibold text-gray-900">修改密码</span>
                </div>
                <ChevronDown className={`w-5 h-5 text-gray-400 transition-transform duration-200 ${showPasswordForm ? 'rotate-180' : ''}`} />
              </button>

              {showPasswordForm && (
                <form onSubmit={handleChangePassword} className="p-5 border-t border-gray-100 space-y-4">
                  {passwordSuccess && (
                    <div className="p-3 bg-emerald-50 text-emerald-700 rounded-lg flex items-center gap-2">
                      <CheckCircle className="w-4 h-4" />
                      {passwordSuccess}
                    </div>
                  )}

                  {passwordError && (
                    <div className="p-3 bg-red-50 text-red-700 rounded-lg flex items-center gap-2">
                      <AlertCircle className="w-4 h-4" />
                      {passwordError}
                    </div>
                  )}

                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-2">当前密码</label>
                    <input
                      type="password"
                      required
                      value={passwordData.old_password}
                      onChange={(e) => setPasswordData({ ...passwordData, old_password: e.target.value })}
                      className="w-full px-3 py-2.5 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-primary-500/20 focus:border-primary-500"
                    />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-2">新密码</label>
                    <input
                      type="password"
                      required
                      value={passwordData.new_password}
                      onChange={(e) => setPasswordData({ ...passwordData, new_password: e.target.value })}
                      className="w-full px-3 py-2.5 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-primary-500/20 focus:border-primary-500"
                      placeholder="至少6位"
                    />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-2">确认新密码</label>
                    <input
                      type="password"
                      required
                      value={passwordData.confirm_password}
                      onChange={(e) => setPasswordData({ ...passwordData, confirm_password: e.target.value })}
                      className="w-full px-3 py-2.5 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-primary-500/20 focus:border-primary-500"
                    />
                  </div>
                  <button
                    type="submit"
                    disabled={passwordLoading}
                    className={`w-full py-3 rounded-xl font-medium transition-all duration-200 ${
                      passwordLoading ? 'bg-gray-100 text-gray-500 cursor-not-allowed' : 'btn-primary shadow-glow'
                    }`}
                  >
                    {passwordLoading ? '修改中...' : '确认修改'}
                  </button>
                </form>
              )}
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
