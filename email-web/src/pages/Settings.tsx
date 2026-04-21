import { useState, useEffect } from 'react';
import { accountApi, syncApi } from '../api/client';
import type { EmailAccount, CreateAccountRequest, EmailProvider, SchedulerStatus } from '../api/types';
import {
  Plus, Trash2, TestTube, CheckCircle, XCircle, AlertCircle,
  Loader2, Play, Pause, Clock, Settings, RefreshCw,
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
  const getEmail = (a: EmailAccount) => a.account_email || (a as Record<string, unknown>).email as string || '';

  return (
    <div className="space-y-6">
      {/* 页面标题 */}
      <div className="flex items-center justify-between">
        <h2 className="text-xl font-semibold">账户设置</h2>
        <button onClick={() => setShowForm(!showForm)} className="flex items-center gap-2 px-4 py-2 bg-primary-600 text-white rounded-lg hover:bg-primary-700 transition-colors">
          <Plus className="w-4 h-4" /> 添加账户
        </button>
      </div>

      {/* 错误提示 */}
      {error && (
        <div className="flex items-center p-4 bg-red-50 border border-red-200 rounded-lg text-red-600">
          <AlertCircle className="w-5 h-5 mr-2 flex-shrink-0" />
          <span>{error}</span>
          <button onClick={() => setError(null)} className="ml-auto text-red-400 hover:text-red-600">&times;</button>
        </div>
      )}

      {/* 同步调度器面板 */}
      <div className="bg-white rounded-lg border border-gray-200">
        <div className="p-4 border-b border-gray-200 flex items-center gap-2">
          <Settings className="w-5 h-5 text-gray-500" />
          <h3 className="font-medium">同步调度器</h3>
        </div>
        <div className="p-4 space-y-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <div className={`w-3 h-3 rounded-full ${scheduler?.running ? 'bg-green-500 animate-pulse' : 'bg-gray-300'}`} />
              <div>
                <span className="font-medium">{scheduler?.running ? '运行中' : '已停止'}</span>
                {scheduler?.running && scheduler.next_sync_time && (
                  <span className="text-sm text-gray-500 ml-2">
                    下次同步: {new Date(scheduler.next_sync_time).toLocaleTimeString()}
                  </span>
                )}
              </div>
            </div>
            <button onClick={handleToggleScheduler} disabled={schedulerLoading}
              className={`flex items-center gap-2 px-4 py-2 rounded-lg font-medium transition-colors ${
                scheduler?.running
                  ? 'bg-red-50 text-red-700 hover:bg-red-100 border border-red-200'
                  : 'bg-green-50 text-green-700 hover:bg-green-100 border border-green-200'
              }`}>
              {schedulerLoading ? <Loader2 className="w-4 h-4 animate-spin" /> : scheduler?.running ? <Pause className="w-4 h-4" /> : <Play className="w-4 h-4" />}
              {scheduler?.running ? '停止' : '启动'}
            </button>
          </div>

          {/* 同步间隔设置 */}
          <div className="flex items-center gap-3">
            <Clock className="w-4 h-4 text-gray-400" />
            <span className="text-sm text-gray-600">同步间隔:</span>
            <input type="number" min={1} max={1440} value={intervalInput}
              onChange={(e) => setIntervalInput(Number(e.target.value))}
              className="w-20 px-2 py-1 border border-gray-300 rounded text-center text-sm" />
            <span className="text-sm text-gray-500">分钟</span>
            <button onClick={handleSetInterval}
              className="px-3 py-1 text-sm bg-gray-100 hover:bg-gray-200 rounded transition-colors">
              应用
            </button>
          </div>

          {/* 统计信息 */}
          {scheduler && (
            <div className="grid grid-cols-3 gap-4 text-center">
              <div className="p-3 bg-gray-50 rounded-lg">
                <div className="text-lg font-semibold text-gray-900">{scheduler.sync_count}</div>
                <div className="text-xs text-gray-500">同步次数</div>
              </div>
              <div className="p-3 bg-gray-50 rounded-lg">
                <div className="text-lg font-semibold text-gray-900">{scheduler.error_count}</div>
                <div className="text-xs text-gray-500">错误次数</div>
              </div>
              <div className="p-3 bg-gray-50 rounded-lg">
                <div className="text-lg font-semibold text-gray-900">{scheduler.interval}</div>
                <div className="text-xs text-gray-500">间隔(分钟)</div>
              </div>
            </div>
          )}
        </div>
      </div>

      {/* 添加账户表单 */}
      {showForm && (
        <div className="bg-white rounded-lg border border-gray-200 p-6">
          <h3 className="text-lg font-medium mb-4">添加邮箱账户</h3>
          <form onSubmit={handleSubmit} className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">邮箱地址</label>
              <input type="email" required placeholder="yourname@126.com" value={formData.email}
                onChange={(e) => setFormData({ ...formData, email: e.target.value })}
                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-primary-500" />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">邮箱服务商</label>
              <select value={formData.provider} onChange={(e) => setFormData({ ...formData, provider: e.target.value as EmailProvider })}
                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-primary-500">
                <option value="126">网易126邮箱</option>
                <option value="gmail">Gmail</option>
                <option value="outlook">Outlook</option>
                <option value="imap">通用IMAP</option>
              </select>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">授权码</label>
              <input type="password" required placeholder="请输入邮箱授权码（非登录密码）" value={formData.credential}
                onChange={(e) => setFormData({ ...formData, credential: e.target.value })}
                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-primary-500" />
              <p className="mt-1 text-xs text-gray-500">
                {formData.provider === '126'
                  ? '126邮箱授权码获取方式：设置 → POP3/SMTP/IMAP → 开启服务 → 生成授权码'
                  : '请输入邮箱服务商提供的授权码或应用密码'}
              </p>
            </div>
            <div className="flex gap-3">
              <button type="submit" disabled={submitting}
                className={`flex-1 px-4 py-2 rounded-lg font-medium transition-colors ${submitting ? 'bg-gray-100 text-gray-500 cursor-not-allowed' : 'bg-primary-600 text-white hover:bg-primary-700'}`}>
                {submitting ? <span className="flex items-center justify-center gap-2"><Loader2 className="w-4 h-4 animate-spin" />添加中...</span> : '添加账户'}
              </button>
              <button type="button" onClick={() => setShowForm(false)}
                className="px-4 py-2 border border-gray-300 rounded-lg hover:bg-gray-50 transition-colors">取消</button>
            </div>
          </form>
        </div>
      )}

      {/* 账户列表 */}
      <div className="bg-white rounded-lg border border-gray-200">
        <div className="p-4 border-b border-gray-200 flex items-center justify-between">
          <h3 className="font-medium">已添加的账户</h3>
          <button onClick={fetchAccounts} className="p-1.5 text-gray-400 hover:text-gray-600 hover:bg-gray-100 rounded transition-colors">
            <RefreshCw className="w-4 h-4" />
          </button>
        </div>
        {loading ? (
          <div className="flex items-center justify-center h-32"><Loader2 className="w-6 h-6 animate-spin text-primary-600" /></div>
        ) : accounts.length === 0 ? (
          <div className="text-center py-12 text-gray-500">暂无账户，请添加邮箱账户</div>
        ) : (
          <div className="divide-y divide-gray-200">
            {accounts.map((account) => {
              const result = testResult[account.id];
              const email = getEmail(account);
              return (
                <div key={account.id} className="p-4 hover:bg-gray-50 transition-colors">
                  <div className="flex items-center justify-between">
                    <div className="flex-1">
                      <div className="flex items-center gap-3">
                        <div className="w-10 h-10 rounded-full bg-gradient-to-br from-blue-500 to-blue-600 flex items-center justify-center text-white font-medium">
                          {email.charAt(0).toUpperCase()}
                        </div>
                        <div>
                          <div className="font-medium text-gray-900">{email}</div>
                          <div className="text-sm text-gray-500">
                            {account.provider.toUpperCase()} •
                            {account.sync_enabled ? ' 已启用' : ' 已禁用'} •
                            {account.last_sync_at ? ` 最后同步: ${new Date(account.last_sync_at).toLocaleString()}` : ' 尚未同步'}
                          </div>
                        </div>
                      </div>
                      {result && (
                        <div className={`mt-2 text-sm flex items-center gap-1 ${result.success ? 'text-green-600' : 'text-red-600'}`}>
                          {result.success ? <CheckCircle className="w-4 h-4" /> : <XCircle className="w-4 h-4" />}
                          {result.message}
                        </div>
                      )}
                    </div>
                    <div className="flex items-center gap-2">
                      <button onClick={() => handleTest(account.id)} disabled={testingAccountId === account.id}
                        className={`p-2 rounded-lg transition-colors ${testingAccountId === account.id ? 'bg-gray-100 text-gray-500 cursor-not-allowed' : 'text-gray-600 hover:bg-gray-100 hover:text-gray-900'}`}
                        title="测试连接">
                        {testingAccountId === account.id ? <Loader2 className="w-5 h-5 animate-spin" /> : <TestTube className="w-5 h-5" />}
                      </button>
                      <button onClick={() => handleDelete(account.id)}
                        className="p-2 text-red-600 hover:bg-red-50 hover:text-red-700 rounded-lg transition-colors" title="删除账户">
                        <Trash2 className="w-5 h-5" />
                      </button>
                    </div>
                  </div>
                </div>
              );
            })}
          </div>
        )}
      </div>
    </div>
  );
}
