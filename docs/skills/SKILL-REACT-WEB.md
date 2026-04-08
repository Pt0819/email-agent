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

> 生成时间: 2026-04-08
> 适用于: React前端开发