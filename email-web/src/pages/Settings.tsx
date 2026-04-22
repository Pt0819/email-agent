import { useState, useEffect } from 'react';
import { accountApi, syncApi } from '../api/client';
import type { EmailAccount, EmailProvider, SchedulerStatus } from '../api/types';
import {
  Plus, Trash2, TestTube, CheckCircle, XCircle, AlertCircle,
  Loader2, Play, Pause, Clock, Settings, RefreshCw, Mail,
} from 'lucide-react';

export default function SettingsPage() {
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

  useEffect(() => { fetchAccounts(); fetchSchedulerStatus(); }, []);

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

  // 获取账户显示邮箱（兼容字段名）
  const getEmail = (a: EmailAccount) => a.account_email || '';

  return (
    <div className="max-w-5xl mx-auto px-4 py-6">
      <div className="space-y-6 animate-fade-in">
        {/* 页面标题 */}
        <div className="flex items-center justify-between">
          <h2 className="text-xl font-bold text-gray-900">账户设置</h2>
          <button onClick={() => setShowForm(!showForm)} className="btn-primary shadow-glow">
            <Plus className="w-4 h-4" /> 添加账户
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
                  className="input" />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">邮箱服务商</label>
                <select value={formData.provider} onChange={(e) => setFormData({ ...formData, provider: e.target.value as EmailProvider })}
                  className="input">
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
                  className="input" />
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
    </div>
  );
}
