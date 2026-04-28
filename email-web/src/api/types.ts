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
  avatar_url?: string;
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
  | 'steam_promotion'
  | 'steam_wishlist'
  | 'steam_news'
  | 'steam_update'
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
  steam_promotion: 'Steam促销',
  steam_wishlist: 'Steam愿望单',
  steam_news: 'Steam资讯',
  steam_update: 'Steam更新',
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
  steam_promotion: 'bg-emerald-100 text-emerald-800 border-emerald-200',
  steam_wishlist: 'bg-teal-100 text-teal-800 border-teal-200',
  steam_news: 'bg-cyan-100 text-cyan-800 border-cyan-200',
  steam_update: 'bg-sky-100 text-sky-800 border-sky-200',
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
    { value: 'steam_promotion' as const, label: 'Steam促销' },
    { value: 'steam_wishlist' as const, label: 'Steam愿望单' },
    { value: 'steam_news' as const, label: 'Steam资讯' },
    { value: 'steam_update' as const, label: 'Steam更新' },
    { value: 'unclassified' as const, label: '未分类' },
  ],
  statuses: [
    { value: 'unread' as const, label: '未读' },
    { value: 'read' as const, label: '已读' },
    { value: 'archived' as const, label: '已归档' },
  ],
};

// ==================== Steam类型 ====================

export interface SteamGame {
  id: number;
  user_id: number;
  game_name: string;
  game_id: string;
  developer: string;
  publisher: string;
  genre: string;
  tags: string; // JSON字符串
  cover_url: string;
  store_url: string;
  playtime: number;
  is_owned: boolean;
  created_at: string;
  updated_at: string;
}

export interface SteamDeal {
  id: number;
  user_id: number;
  game_id: string;
  game_name: string;
  original_price: number;
  deal_price: number;
  discount: number;
  cover_url: string;
  store_url: string;
  start_date?: string;
  end_date?: string;
  is_active: boolean;
  email_id: number;
  created_at: string;
  updated_at: string;
}

export interface SteamStats {
  total_games: number;
  active_deals: number;
  is_bound: boolean;
  steam_nickname?: string;
  avatar_url?: string;
  last_sync?: string;
}

// Steam账号绑定
export interface SteamAccount {
  id: number;
  user_id: number;
  steam_id: string;
  steam_nickname: string;
  avatar_url: string;
  profile_url: string;
  real_name: string;
  location: string;
  last_sync_at?: string;
  is_active: boolean;
  created_at: string;
}

// Steam游戏库条目
export interface SteamLibraryItem {
  id: number;
  user_id: number;
  account_id: number;
  game_id: string;
  game_name: string;
  playtime: number; // 总游玩时长(分钟)
  playtime_2_weeks: number; // 最近两周(分钟)
  last_played_at?: string;
  icon_url: string;
  is_synced: boolean;
  created_at: string;
}

// 绑定Steam账号请求
export interface BindSteamRequest {
  steam_id: string;
}

// ==================== 偏好分析类型 ====================

export interface TagPreference {
  tag: string;
  weight: number;
  count?: number;
  source: string;
}

export interface GenreCount {
  genre: string;
  count: number;
}

export interface RecentActivitySummary {
  games_played_last_week: number;
  total_playtime_last_week: number;
  most_played_game: string;
  most_played_game_hours: number;
  new_games_added: number;
  genre_distribution: GenreCount[];
}

export interface UserGamingProfile {
  user_id: number;
  top_tags: TagPreference[];
  top_genres: TagPreference[];
  top_developers: TagPreference[];
  total_games: number;
  total_playtime: number;
  recent_activity: RecentActivitySummary | null;
  last_analyzed_at: string | null;
}

export interface PreferenceInsight {
  id: number;
  event_type: string;
  decision_type: string;
  trigger_desc: string;
  insight: string;
  reasoning: string;
  confidence: number;
  is_anomaly: boolean;
  anomaly_type: string;
  game_id: string;
  game_name: string;
  tags_changed: TagChange[];
  created_at: string;
}

export interface TagChange {
  tag: string;
  old: number;
  new: number;
  delta: number;
}

export interface InsightListResponse {
  list: PreferenceInsight[];
  total: number;
  page: number;
}

export interface PreferenceAnalysisResult {
  success: boolean;
  profile: UserGamingProfile | null;
  new_tags: TagPreference[];
  updated_tags: TagPreference[];
  insights: string[];
  error?: string;
}

// ==================== 推荐类型 ====================

export interface GameRecommendation {
  id: number;
  game_id: string;
  game_name: string;
  game_genre: string;
  game_tags: string[];
  cover_url: string;
  store_url: string;
  match_score: number;
  match_reasons: string[];
  has_deal: boolean;
  deal_price: number;
  deal_discount: number;
  deal_end_date?: string;
  source: string;
  status: string;
  created_at: string;
}

export interface RecStatsSummary {
  total_recommendations: number;
  clicked_count: number;
  purchase_count: number;
  ctr: number;
  purchase_rate: number;
}

export interface RecommendationListResponse {
  list: GameRecommendation[];
  total: number;
  page: number;
  page_size: number;
  stats?: RecStatsSummary;
}

export interface FeedbackRequest {
  action: 'like' | 'dislike' | 'click' | 'purchase' | 'ignore';
}

export interface GenerateRecommendationRequest {
  max_count?: number;
  deal_only?: boolean;
  min_score?: number;
  game_ids?: string[];
}
