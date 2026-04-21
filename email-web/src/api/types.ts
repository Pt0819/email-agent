// ==================== 通用类型 ====================

export interface ApiResponse<T = unknown> {
  code: number;
  message: string;
  data?: T;
}

export interface PageData<T = unknown> {
  list: T[];
  total: number;
  page: number;
  page_size: number;
}

// ==================== 认证类型 ====================

export interface User {
  id: number;
  user_id: string;
  username: string;
  email: string;
  created_at: string;
}

export interface LoginRequest {
  email: string;
  password: string;
}

export interface RegisterRequest {
  username: string;
  email: string;
  password: string;
}

export interface AuthResponse {
  token: string;
  expires_at: string;
  user: User;
}

// ==================== 邮件类型 ====================

export type EmailCategory =
  | 'work_urgent'
  | 'work_normal'
  | 'personal'
  | 'subscription'
  | 'notification'
  | 'promotion'
  | 'spam'
  | 'unclassified';

export type EmailPriority = 'critical' | 'high' | 'medium' | 'low';

export type EmailStatus = 'unread' | 'read' | 'processed' | 'archived';

export interface Email {
  id: number;
  message_id: string;
  user_id: number;
  account_id: number;
  sender_name: string;
  sender_email: string;
  subject: string;
  content: string;
  content_html?: string;
  content_type?: string;
  category: EmailCategory;
  priority: EmailPriority;
  confidence: number;
  reasoning?: string;
  status: EmailStatus;
  is_processed?: boolean;
  has_attachment: boolean;
  received_at: string;
  processed_at?: string;
  created_at: string;
  updated_at?: string;
}

export interface EmailListParams {
  page?: number;
  page_size?: number;
  account_id?: number;
  category?: EmailCategory | 'all';
  status?: EmailStatus | 'all';
  keyword?: string;
}

// ==================== 账户类型 ====================

export type EmailProvider = '126' | 'gmail' | 'outlook' | 'imap';

export interface EmailAccount {
  id: number;
  user_id: number;
  provider: EmailProvider;
  account_email: string; // 后端返回的字段名
  last_sync_at?: string;
  sync_enabled: boolean;
  created_at: string;
  updated_at?: string;
}

export interface CreateAccountRequest {
  email: string;
  provider: EmailProvider;
  credential: string; // 授权码
}

// ==================== 分类类型 ====================

export interface ClassificationResult {
  category: EmailCategory;
  priority: EmailPriority;
  confidence: number;
  reasoning: string;
}

export interface ClassifyResponse {
  email_id: string;
  category: EmailCategory;
  priority: EmailPriority;
  confidence: number;
  reasoning: string;
}

// ==================== 同步类型 ====================

export interface SyncResult {
  account_id: number;
  account_email: string;
  success: boolean;
  message: string;
  total_count: number;
  synced_count: number;
  error_count: number;
  classified_count: number;
  synced_at: string;
}

export interface SyncResponse {
  status: string;
  all_success: boolean;
  results: SyncResult[];
}

export interface SyncStatusData {
  accounts: EmailAccount[];
  last_sync?: string;
}

// ==================== 调度器类型 ====================

export interface SchedulerStatus {
  running: boolean;
  interval: number; // 分钟
  last_sync_time?: string;
  sync_count: number;
  error_count: number;
  next_sync_time?: string;
}

export interface SetIntervalRequest {
  interval: number; // 分钟，1-1440
}

// ==================== 每日摘要类型 ====================

export interface ImportantEmail {
  email_id: string;
  subject: string;
  sender: string;
  category: string;
  priority: string;
  summary: string;
}

export interface ActionItemSummary {
  task: string;
  priority: string;
}

export interface DailySummary {
  date: string;
  total_emails: number;
  by_category: Record<string, number>;
  category_labels: Record<string, string>;
  important_emails: ImportantEmail[];
  action_items: ActionItemSummary[];
  summary_text: string;
}

// ==================== 分类映射 ====================

export const CATEGORY_LABELS: Record<EmailCategory, string> = {
  work_urgent: '紧急工作',
  work_normal: '普通工作',
  personal: '个人邮件',
  subscription: '订阅邮件',
  notification: '系统通知',
  promotion: '营销推广',
  spam: '垃圾邮件',
  unclassified: '未分类',
};

export const CATEGORY_COLORS: Record<EmailCategory, string> = {
  work_urgent: 'bg-red-100 text-red-800 border-red-200',
  work_normal: 'bg-blue-100 text-blue-800 border-blue-200',
  personal: 'bg-green-100 text-green-800 border-green-200',
  subscription: 'bg-purple-100 text-purple-800 border-purple-200',
  notification: 'bg-gray-100 text-gray-800 border-gray-200',
  promotion: 'bg-yellow-100 text-yellow-800 border-yellow-200',
  spam: 'bg-pink-100 text-pink-800 border-pink-200',
  unclassified: 'bg-slate-100 text-slate-800 border-slate-200',
};

export const PRIORITY_LABELS: Record<EmailPriority, string> = {
  critical: '紧急',
  high: '高',
  medium: '中',
  low: '低',
};

export const PRIORITY_COLORS: Record<EmailPriority, string> = {
  critical: 'text-red-600',
  high: 'text-orange-600',
  medium: 'text-yellow-600',
  low: 'text-gray-600',
};

// ==================== 筛选器选项 ====================

export const FILTERS = {
  categories: [
    { value: 'work_urgent' as const, label: '紧急工作' },
    { value: 'work_normal' as const, label: '普通工作' },
    { value: 'personal' as const, label: '个人邮件' },
    { value: 'subscription' as const, label: '订阅邮件' },
    { value: 'notification' as const, label: '系统通知' },
    { value: 'promotion' as const, label: '营销推广' },
    { value: 'spam' as const, label: '垃圾邮件' },
    { value: 'unclassified' as const, label: '未分类' },
  ],
  statuses: [
    { value: 'unread' as const, label: '未读' },
    { value: 'read' as const, label: '已读' },
    { value: 'archived' as const, label: '已归档' },
  ],
};
