import { useState, useEffect } from 'react';
import { accountApi } from '../api/client';
import type { EmailAccount, CreateAccountRequest, EmailProvider } from '../api/types';
import {
  Plus,
  Trash2,
  TestTube,
  CheckCircle,
  XCircle,
  AlertCircle,
  Loader2,
} from 'lucide-react';

export default function SettingsPage() {
  const [accounts, setAccounts] = useState<EmailAccount[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // 表单状态
  const [showForm, setShowForm] = useState(false);
  const [formData, setFormData] = useState({
    email: '',
    provider: '126' as EmailProvider,
    credential: '',
  });
  const [submitting, setSubmitting] = useState(false);

  // 测试连接状态
  const [testingAccountId, setTestingAccountId] = useState<number | null>(null);
  const [testResult, setTestResult] = useState<Record<number, { success: boolean; message: string }>>({});

  useEffect(() => {
    fetchAccounts();
  }, []);

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

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      setSubmitting(true);
      const data: CreateAccountRequest = {
        email: formData.email,
        provider: formData.provider,
        credential: formData.credential,
      };
      await accountApi.create(data);
      // 重置表单
      setFormData({ email: '', provider: '126', credential: '' });
      setShowForm(false);
      // 刷新列表
      fetchAccounts();
    } catch (err) {
      setError(err instanceof Error ? err.message : '添加账户失败');
    } finally {
      setSubmitting(false);
    }
  };

  const handleDelete = async (id: number) => {
    if (!confirm('确定要删除此账户吗？相关邮件数据也会被删除。')) {
      return;
    }
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
      const data = response as unknown as { status: string; message: string };
      setTestResult({
        ...testResult,
        [id]: { success: data.status === 'success', message: data.message },
      });
    } catch (err) {
      setTestResult({
        ...testResult,
        [id]: { success: false, message: err instanceof Error ? err.message : '测试连接失败' },
      });
    } finally {
      setTestingAccountId(null);
    }
  };

  return (
    <div className="space-y-6">
      {/* 页面标题 */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <div className="w-6 h-6" />
          <h2 className="text-xl font-semibold">账户设置</h2>
        </div>
        <button
          onClick={() => setShowForm(!showForm)}
          className="flex items-center gap-2 px-4 py-2 bg-primary-600 text-white rounded-lg hover:bg-primary-700 transition-colors"
        >
          <Plus className="w-4 h-4" />
          添加账户
        </button>
      </div>

      {/* 错误提示 */}
      {error && (
        <div className="flex items-center p-4 bg-red-50 border border-red-200 rounded-lg text-red-600">
          <AlertCircle className="w-5 h-5 mr-2 flex-shrink-0" />
          <span>{error}</span>
        </div>
      )}

      {/* 添加账户表单 */}
      {showForm && (
        <div className="bg-white rounded-lg border border-gray-200 p-6">
          <h3 className="text-lg font-medium mb-4">添加邮箱账户</h3>
          <form onSubmit={handleSubmit} className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                邮箱地址
              </label>
              <input
                type="email"
                required
                placeholder="yourname@126.com"
                value={formData.email}
                onChange={(e) => setFormData({ ...formData, email: e.target.value })}
                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-primary-500"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                邮箱服务商
              </label>
              <select
                value={formData.provider}
                onChange={(e) => setFormData({ ...formData, provider: e.target.value as EmailProvider })}
                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-primary-500"
              >
                <option value="126">网易126邮箱</option>
                <option value="gmail">Gmail</option>
                <option value="outlook">Outlook</option>
                <option value="imap">通用IMAP</option>
              </select>
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                授权码
              </label>
              <input
                type="password"
                required
                placeholder="请输入邮箱授权码（非登录密码）"
                value={formData.credential}
                onChange={(e) => setFormData({ ...formData, credential: e.target.value })}
                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-primary-500"
              />
              <p className="mt-1 text-xs text-gray-500">
                {formData.provider === '126'
                  ? '126邮箱授权码获取方式：设置 → POP3/SMTP/IMAP → 开启服务 → 生成授权码'
                  : '请输入邮箱服务商提供的授权码或应用密码'}
              </p>
            </div>

            <div className="flex gap-3">
              <button
                type="submit"
                disabled={submitting}
                className={`flex-1 px-4 py-2 rounded-lg font-medium transition-colors ${
                  submitting
                    ? 'bg-gray-100 text-gray-500 cursor-not-allowed'
                    : 'bg-primary-600 text-white hover:bg-primary-700'
                }`}
              >
                {submitting ? (
                  <span className="flex items-center justify-center gap-2">
                    <Loader2 className="w-4 h-4 animate-spin" />
                    添加中...
                  </span>
                ) : (
                  '添加账户'
                )}
              </button>
              <button
                type="button"
                onClick={() => setShowForm(false)}
                className="px-4 py-2 border border-gray-300 rounded-lg hover:bg-gray-50 transition-colors"
              >
                取消
              </button>
            </div>
          </form>
        </div>
      )}

      {/* 账户列表 */}
      <div className="bg-white rounded-lg border border-gray-200">
        <div className="p-4 border-b border-gray-200">
          <h3 className="font-medium">已添加的账户</h3>
        </div>

        {loading ? (
          <div className="flex items-center justify-center h-32">
            <Loader2 className="w-6 h-6 animate-spin text-primary-600" />
          </div>
        ) : accounts.length === 0 ? (
          <div className="text-center py-12 text-gray-500">
            暂无账户，请添加邮箱账户
          </div>
        ) : (
          <div className="divide-y divide-gray-200">
            {accounts.map((account) => {
              const result = testResult[account.id];
              return (
                <div key={account.id} className="p-4 hover:bg-gray-50 transition-colors">
                  <div className="flex items-center justify-between">
                    <div className="flex-1">
                      <div className="flex items-center gap-3">
                        <div className="w-10 h-10 rounded-full bg-gradient-to-br from-blue-500 to-blue-600 flex items-center justify-center text-white font-medium">
                          {account.account_email.charAt(0).toUpperCase()}
                        </div>
                        <div>
                          <div className="font-medium text-gray-900">
                            {account.account_email}
                          </div>
                          <div className="text-sm text-gray-500">
                            {account.provider.toUpperCase()} •
                            {account.last_sync_at
                              ? ` 最后同步: ${new Date(account.last_sync_at).toLocaleString()}`
                              : ' 尚未同步'}
                          </div>
                        </div>
                      </div>

                      {/* 测试结果 */}
                      {result && (
                        <div className={`mt-2 text-sm flex items-center gap-1 ${
                          result.success ? 'text-green-600' : 'text-red-600'
                        }`}>
                          {result.success ? (
                            <CheckCircle className="w-4 h-4" />
                          ) : (
                            <XCircle className="w-4 h-4" />
                          )}
                          {result.message}
                        </div>
                      )}
                    </div>

                    <div className="flex items-center gap-2">
                      <button
                        onClick={() => handleTest(account.id)}
                        disabled={testingAccountId === account.id}
                        className={`p-2 rounded-lg transition-colors ${
                          testingAccountId === account.id
                            ? 'bg-gray-100 text-gray-500 cursor-not-allowed'
                            : 'text-gray-600 hover:bg-gray-100 hover:text-gray-900'
                        }`}
                        title="测试连接"
                      >
                        {testingAccountId === account.id ? (
                          <Loader2 className="w-5 h-5 animate-spin" />
                        ) : (
                          <TestTube className="w-5 h-5" />
                        )}
                      </button>
                      <button
                        onClick={() => handleDelete(account.id)}
                        className="p-2 text-red-600 hover:bg-red-50 hover:text-red-700 rounded-lg transition-colors"
                        title="删除账户"
                      >
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
