# Skill: React Web Project Setup

> 本Skill用于快速初始化React + TypeScript + Tailwind CSS项目

## 1. 项目结构

```
email-web/
├── src/
│   ├── api/
│   │   ├── config.ts        # API配置
│   │   ├── client.ts        # Axios客户端
│   │   └── types.ts         # TypeScript类型
│   │
│   ├── components/
│   │   ├── ui/              # 基础UI组件
│   │   │   ├── Button.tsx
│   │   │   ├── Card.tsx
│   │   │   └── Badge.tsx
│   │   └── layout/          # 布局组件
│   │       ├── Header.tsx
│   │       └── MainLayout.tsx
│   │
│   ├── pages/
│   │   ├── EmailList.tsx    # 邮件列表页
│   │   ├── EmailDetail.tsx  # 邮件详情页
│   │   ├── Settings.tsx     # 设置页面
│   │   └── Dashboard.tsx    # 首页仪表盘
│   │
│   ├── hooks/               # 自定义Hooks
│   │   ├── useEmails.ts
│   │   └── useAccounts.ts
│   │
│   ├── App.tsx              # 根组件
│   ├── main.tsx             # 入口文件
│   └── index.css            # 全局样式
│
├── public/
│   └── vite.svg
│
├── index.html
├── package.json
├── tsconfig.json
├── vite.config.ts
├── tailwind.config.js
└── postcss.config.js
```

## 2. 命名规范

### 文件命名
- 组件: 大驼峰: `EmailList.tsx`
- 工具函数: 小写: `api.ts`
- 类型文件: 小写: `types.ts`
- 测试文件: `EmailList.test.tsx`

### 组件命名
- 组件: 大驼峰: `export default function EmailList()`
- 类型: 大驼峰+Type后缀: `type EmailListProps`
- 接口: 大驼峰: `interface Email { ... }`

### 变量命名
- 常量: 大写下划线: `API_BASE_URL`
- 变量: 驼峰: `emailList`, `isLoading`
- 布尔值: is/has前缀: `isLoading`, `hasError`
- 事件处理: handle前缀: `handleClick`, `handleSubmit`

## 3. 开发规范

### 项目初始化
```bash
# 创建项目
npm create vite@latest . -- --template react-ts

# 安装依赖
npm install

# 安装Tailwind
npm install tailwindcss @tailwindcss/postcss

# 安装其他依赖
npm install @tanstack/react-query axios
npm install lucide-react clsx tailwind-merge
npm install class-variance-authority
```

### Tailwind配置
```javascript
// tailwind.config.js
export default {
  content: ["./index.html", "./src/**/*.{js,ts,jsx,tsx}"],
  theme: {
    extend: {
      colors: {
        primary: {
          500: '#3b82f6',
          600: '#2563eb',
        }
      }
    }
  }
}
```

### index.css
```css
@tailwind base;
@tailwind components;
@tailwind utilities;
```

## 4. 错误处理

### API错误处理
```typescript
// api/client.ts
import axios from 'axios';

const apiClient = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL,
  timeout: 30000,
});

// 响应拦截器
apiClient.interceptors.response.use(
  (response) => response.data,
  (error) => {
    const message = error.response?.data?.message || error.message;
    console.error('API Error:', message);
    return Promise.reject(new Error(message));
  }
);
```

### 组件错误处理
```typescript
// 使用try-catch + toast
import { toast } from 'sonner'; // 或其他toast库

const handleSubmit = async () => {
  try {
    await api.submit(data);
    toast.success('提交成功');
  } catch (error) {
    toast.error(error instanceof Error ? error.message : '操作失败');
  }
};
```

## 5. 类型定义

```typescript
// api/types.ts

// 邮件类型
export type EmailCategory = 'work_normal' | 'personal' | 'spam' | 'unclassified';
export type EmailPriority = 'critical' | 'high' | 'medium' | 'low';

export interface Email {
  id: string;
  subject: string;
  sender_email: string;
  category: EmailCategory;
  priority: EmailPriority;
  received_at: string;
}

// API响应
export interface ApiResponse<T> {
  code: number;
  message: string;
  data?: T;
}

// 分类标签映射
export const CATEGORY_LABELS: Record<EmailCategory, string> = {
  work_normal: '普通工作',
  personal: '个人邮件',
  spam: '垃圾邮件',
  unclassified: '未分类',
};
```

## 6. 快速开始命令

```bash
# 1. 安装依赖
cd email-web
npm install

# 2. 开发模式
npm run dev

# 3. 构建生产版本
npm run build

# 4. 预览生产版本
npm run preview

# 5. 类型检查
npm run typecheck  # 或 tsc --noEmit
```

## 7. 常用依赖

| 依赖 | 版本 | 用途 |
|------|------|------|
| react | 18.x | UI框架 |
| react-dom | 18.x | React DOM |
| typescript | 5.x | 类型支持 |
| vite | 5.x | 构建工具 |
| tailwindcss | 4.x | CSS框架 |
| axios | 1.x | HTTP客户端 |
| @tanstack/react-query | 5.x | 数据请求 |
| lucide-react | 最新 | 图标库 |
| clsx | 2.x | 类名合并 |

## 8. 环境变量

```bash
# .env.local
VITE_API_BASE_URL=http://localhost:8080/api/v1
VITE_AGENT_API_URL=http://localhost:8001
```

```typescript
// 使用环境变量
const API_URL = import.meta.env.VITE_API_BASE_URL;
```

## 9. 组件模板

```typescript
// 页面组件模板
import { useState, useEffect } from 'react';
import { api } from '../api/client';
import type { Data } from '../api/types';

export default function PageName() {
  const [data, setData] = useState<Data[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    fetchData();
  }, []);

  const fetchData = async () => {
    try {
      setLoading(true);
      const result = await api.getData();
      setData(result);
    } catch (err) {
      setError(err instanceof Error ? err.message : '获取数据失败');
    } finally {
      setLoading(false);
    }
  };

  if (loading) return <div>Loading...</div>;
  if (error) return <div>Error: {error}</div>;

  return (
    <div className="p-4">
      {/* 内容 */}
    </div>
  );
}
```

---

## 10. API客户端规范

### Axios封装模式
```typescript
// api/client.ts
const apiClient: AxiosInstance = axios.create({
  baseURL: API_BASE_URL,
  timeout: API_TIMEOUT,
  headers,
});

// 请求拦截器 - 自动添加Token
apiClient.interceptors.request.use((config) => {
  const token = localStorage.getItem('token');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

// 响应拦截器 - 自动解包
apiClient.interceptors.response.use(
  (response) => response.data,  // 直接返回data层
  (error: AxiosError<ApiResponse>) => {
    const message = error.response?.data?.message || error.message;
    return Promise.reject(new Error(message));
  }
);
```

### API模块化组织
```typescript
// api/client.ts - 按资源分组
export const emailApi = {
  list: (params?: EmailListParams) =>
    apiClient.get<ApiResponse<PageData<Email>>>('/emails', { params }),
  getById: (id: string) =>
    apiClient.get<ApiResponse<Email>>(`/emails/${id}`),
  classify: (id: string) =>
    apiClient.post<ApiResponse<ClassifyResponse>>(`/emails/${id}/classify`),
};

export const accountApi = {
  list: () => apiClient.get<ApiResponse<{ list: EmailAccount[] }>>('/accounts'),
  create: (data: CreateAccountRequest) =>
    apiClient.post<ApiResponse<EmailAccount>>('/accounts', data),
  delete: (id: number) => apiClient.delete(`/accounts/${id}`),
  test: (id: number) =>
    apiClient.post(`/accounts/${id}/test`),
};

export const syncApi = {
  trigger: (accountId?: number) =>
    apiClient.post('/sync', { account_id: accountId }),
  status: () => apiClient.get('/sync/status'),
};
```

### 类型定义规范
```typescript
// api/types.ts

// 1. 通用类型
export interface ApiResponse<T = unknown> {
  code: number;
  message: string;
  data?: T;
}

export interface PageData<T = unknown> {
  list: T[];
  total: number;
}

// 2. 枚举类型使用 union type
export type EmailCategory = 'work_urgent' | 'work_normal' | 'personal' | 'spam';
export type EmailPriority = 'critical' | 'high' | 'medium' | 'low';

// 3. 显示映射用 Record
export const CATEGORY_LABELS: Record<EmailCategory, string> = {
  work_urgent: '紧急工作',
  work_normal: '普通工作',
  ...
};

export const CATEGORY_COLORS: Record<EmailCategory, string> = {
  work_urgent: 'bg-red-100 text-red-800 border-red-200',
  work_normal: 'bg-blue-100 text-blue-800 border-blue-200',
  ...
};
```

## 11. 页面组件模式

### 列表页模式（带筛选/分页）
```typescript
export default function EmailList() {
  // 状态管理
  const [items, setItems] = useState<Email[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [total, setTotal] = useState(0);

  // 筛选/分页状态
  const [page, setPage] = useState(1);
  const [selectedCategory, setSelectedCategory] = useState<EmailCategory | 'all'>('all');

  // 数据获取（useCallback + useEffect模式）
  const fetchData = useCallback(async () => {
    try {
      setLoading(true);
      const response = await emailApi.list({ page, category, keyword });
      const pageData = response as unknown as { list: Email[]; total: number };
      setItems(pageData.list || []);
      setTotal(pageData.total || 0);
    } catch (err) {
      setError(err instanceof Error ? err.message : '获取数据失败');
    } finally {
      setLoading(false);
    }
  }, [page, selectedCategory]);

  useEffect(() => { fetchData(); }, [fetchData]);

  // 筛选变化重置分页
  const handleCategoryChange = (category: EmailCategory | 'all') => {
    setSelectedCategory(category);
    setPage(1);  // 重置到第一页
  };
}
```

### 表单页面模式（含提交/取消）
```typescript
export default function SettingsPage() {
  const [showForm, setShowForm] = useState(false);
  const [formData, setFormData] = useState({ email: '', provider: '126', credential: '' });
  const [submitting, setSubmitting] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      setSubmitting(true);
      await accountApi.create(formData);
      setFormData({ email: '', provider: '126', credential: '' });  // 重置
      setShowForm(false);
      fetchAccounts();  // 刷新列表
    } catch (err) {
      setError(err instanceof Error ? err.message : '操作失败');
    } finally {
      setSubmitting(false);
    }
  };
}
```

### Loading / Empty / Error 三态处理
```tsx
// Loading态
{loading && items.length === 0 && (
  <div className="flex items-center justify-center h-64">
    <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary-600" />
  </div>
)}

// 空状态
{items.length === 0 && !loading && (
  <div className="text-center py-12">
    <p className="text-gray-500">暂无数据</p>
  </div>
)}

// 错误提示
{error && (
  <div className="flex items-center p-4 bg-red-50 border border-red-200 rounded-lg text-red-600">
    <AlertCircle className="w-5 h-5 mr-2" />
    <span>{error}</span>
  </div>
)}
```

---

> 生成时间: 2026-04-08
> 更新: 2026-04-14 (添加API客户端规范和页面组件模式)
> 适用于: React前端开发