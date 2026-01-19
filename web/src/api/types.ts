export interface KeyStats {
    request_count: number;
    success_count: number;
    error_count: number;
    prompt_tokens: number;
    completion_tokens: number;
    last_used_at: string | null;
}

/**
 * API Key Information
 * Represents a single API Key managed by the system.
 */
export interface KeyInfo {
    /** Unique UUID for the key */
    id: string;
    /** Masked key string (e.g., "AIzaSy...xyz") */
    key: string;
    /** User-friendly name for the key */
    name: string;
    status: 'active' | 'rate_limited' | 'disabled';
    enabled: boolean;
    tags: string[];
    stats: KeyStats;
    cooldown_until: string | null;
    created_at: string;
    updated_at: string;
    /** Provider identifier (e.g., 'google_aistudio') */
    provider: string;
    /** Default model name for this key */
    default_model?: string;
}

export interface KeyImportItem {
    key: string;
    name?: string;
    tags?: string[];
}

export interface Session {
    id: string;
    title: string;
    model: string;
    message_count: number;
    total_tokens: number;
    created_at: string;
    updated_at: string;
}

export interface Message {
    id: string;
    session_id: string;
    role: 'user' | 'assistant' | 'system';
    content: string | any[]; // any[] for multimodal content
    prompt_tokens: number;
    completion_tokens: number;
    created_at: string;
}

export interface ApiResponse<T> {
    success: boolean;
    data?: T;
    error?: {
        code: number;
        message: string;
        type: string;
        param?: string;
    };
    message?: string;
}

export interface ListResponse<T> {
    success: boolean;
    data?: T[];
    sessions?: T[]; // Special case for sessions endpoint
    total?: number;
    message?: string;
}

export interface HealthStats {
    status: string;
    version: string;
    uptime: number;
    keys: {
        total: number;
        active: number;
        rate_limited: number;
        disabled: number;
    };
}

// ==================== Chat Types ====================

/**
 * 多模态内容部分 (OpenAI image_url 格式)
 */
export interface ContentPart {
    type: 'text' | 'image_url'
    text?: string
    image_url?: { url: string }
}

/**
 * OpenAI 格式的 Chat 消息 (纯文本)
 */
export interface ChatCompletionMessage {
    role: 'user' | 'assistant' | 'system';
    content: string;
}

/**
 * OpenAI 格式的 Chat 消息 (多模态)
 */
export interface ChatCompletionMessageMultimodal {
    role: 'user' | 'assistant' | 'system';
    content: string | ContentPart[];
}

/**
 * Chat Completion 请求参数
 */
export interface ChatCompletionRequest {
    model: string;
    messages: ChatCompletionMessage[];
    temperature?: number;
    max_tokens?: number;
    stream?: boolean;
    top_p?: number;
    stop?: string | string[];
}

/**
 * SSE 流式响应 chunk 的 delta 部分
 */
export interface ChatCompletionDelta {
    role?: string;
    content?: string;
}

/**
 * SSE 流式响应 chunk 的 choice 部分
 */
export interface ChatCompletionChunkChoice {
    index: number;
    delta: ChatCompletionDelta;
    finish_reason: string | null;
}

/**
 * SSE 流式响应 chunk
 */
export interface ChatCompletionChunk {
    id: string;
    object: 'chat.completion.chunk';
    created: number;
    model: string;
    choices: ChatCompletionChunkChoice[];
}

// ==================== Session API Types ====================

/**
 * 会话列表响应
 */
export interface SessionListResponse {
    success: boolean;
    sessions: Session[];
    total: number;
}

/**
 * 会话详情响应 (含消息)
 */
export interface SessionDetailResponse {
    success: boolean;
    session: Session;
    messages: Message[];
}

/**
 * 创建会话请求
 */
export interface CreateSessionRequest {
    title?: string;
    model?: string;
}

/**
 * 添加消息请求
 */
export interface AddMessageRequest {
    role: 'user' | 'assistant' | 'system';
    content: string;
    prompt_tokens?: number;
    completion_tokens?: number;
}

// ==================== Statistics Types ====================

export type StatsTimeRange = '24h' | '7d' | '30d';

export interface TrendDataPoint {
    timestamp: string;
    requests: number;
    tokens: number;
    errors: number;
}

export interface TrendResponse {
    success: boolean;
    data: TrendDataPoint[];
    time_range: StatsTimeRange;
}

export interface ModelUsageItem {
    model: string;
    request_count: number;
    token_usage: number;
    percentage: number;
}

export interface ModelUsageResponse {
    success: boolean;
    data: ModelUsageItem[];
}

