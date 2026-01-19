/**
 * Session API - 会话管理接口封装
 * 
 * 提供会话的 CRUD 操作和消息持久化功能
 */

import apiClient from './client'
import type {
    Session,
    Message,
    SessionListResponse,
    SessionDetailResponse,
    CreateSessionRequest,
    AddMessageRequest,
    ApiResponse
} from './types'

/**
 * 获取会话列表
 * @param limit - 每页数量 (默认 20)
 * @param offset - 偏移量 (默认 0)
 */
export async function getSessions(limit = 20, offset = 0): Promise<SessionListResponse> {
    return apiClient.get('/api/sessions', {
        params: { limit, offset }
    })
}

/**
 * 创建新会话
 * @param data - 会话创建参数
 */
export async function createSession(data: CreateSessionRequest): Promise<ApiResponse<Session>> {
    return apiClient.post('/api/sessions', data)
}

/**
 * 获取会话详情 (含消息历史)
 * @param id - 会话 ID
 */
export async function getSession(id: string): Promise<SessionDetailResponse> {
    return apiClient.get(`/api/sessions/${id}`)
}

/**
 * 更新会话信息
 * @param id - 会话 ID
 * @param data - 更新数据 (title/model)
 */
export async function updateSession(
    id: string,
    data: { title?: string; model?: string }
): Promise<ApiResponse<Session>> {
    return apiClient.put(`/api/sessions/${id}`, data)
}

/**
 * 删除会话
 * @param id - 会话 ID
 */
export async function deleteSession(id: string): Promise<ApiResponse<{ message: string }>> {
    return apiClient.delete(`/api/sessions/${id}`)
}

/**
 * 向会话添加消息
 * @param sessionId - 会话 ID
 * @param data - 消息数据
 */
export async function addMessage(
    sessionId: string,
    data: AddMessageRequest
): Promise<ApiResponse<Message>> {
    return apiClient.post(`/api/sessions/${sessionId}/messages`, data)
}
