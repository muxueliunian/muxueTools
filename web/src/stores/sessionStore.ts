/**
 * Session Store - 会话状态管理
 * 
 * 职责: 管理会话列表、当前会话、消息持久化
 */

import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { Session } from '@/api/types'
import {
    getSessions,
    createSession,
    getSession,
    updateSession as apiUpdateSession,
    deleteSession as apiDeleteSession,
    addMessage
} from '@/api/sessions'
import { useChatStore } from './chatStore'

/** localStorage 键名 - 当前会话 ID */
const STORAGE_KEY_SESSION = 'mxln_current_session_id'

/** 标题最大长度 */
const MAX_TITLE_LENGTH = 20

export const useSessionStore = defineStore('session', () => {
    // ==================== State ====================

    /** 会话列表 */
    const sessions = ref<Session[]>([])

    /** 当前会话 ID */
    const currentSessionId = ref<string | null>(null)

    /** 加载状态 */
    const isLoading = ref(false)

    /** 是否已初始化 */
    const isInitialized = ref(false)

    // ==================== Computed ====================

    /**
     * 当前会话对象
     */
    const currentSession = computed(() => {
        if (!currentSessionId.value) return null
        return sessions.value.find(s => s.id === currentSessionId.value) || null
    })

    // ==================== Actions ====================

    /**
     * 初始化 - 加载会话列表并恢复上次会话
     */
    async function initialize(): Promise<void> {
        if (isInitialized.value) return

        await loadSessions()

        // 尝试恢复上次的会话
        const savedSessionId = localStorage.getItem(STORAGE_KEY_SESSION)
        if (savedSessionId && sessions.value.some(s => s.id === savedSessionId)) {
            await switchSession(savedSessionId)
        } else if (sessions.value.length > 0 && sessions.value[0]) {
            // 使用最新的会话
            await switchSession(sessions.value[0].id)
        } else {
            // 没有会话，创建新的
            await createNewSession()
        }

        isInitialized.value = true
    }

    /**
     * 从后端加载会话列表
     */
    async function loadSessions(): Promise<void> {
        isLoading.value = true
        try {
            const response = await getSessions(50, 0)
            // 按更新时间降序排序
            sessions.value = response.sessions.sort(
                (a, b) => new Date(b.updated_at).getTime() - new Date(a.updated_at).getTime()
            )
        } catch (err) {
            console.error('加载会话列表失败:', err)
            sessions.value = []
        } finally {
            isLoading.value = false
        }
    }

    /**
     * 创建新会话
     */
    async function createNewSession(): Promise<Session | null> {
        const chatStore = useChatStore()

        try {
            const response = await createSession({
                title: 'New Chat',
                model: chatStore.currentModel || 'gemini-2.0-flash'
            })

            if (response.success && response.data) {
                const newSession = response.data
                sessions.value.unshift(newSession)
                currentSessionId.value = newSession.id
                localStorage.setItem(STORAGE_KEY_SESSION, newSession.id)

                // 清空聊天消息
                chatStore.clearMessages()

                return newSession
            }
        } catch (err) {
            console.error('创建会话失败:', err)
        }
        return null
    }

    /**
     * 切换会话并加载历史消息
     * @param id - 会话 ID
     */
    async function switchSession(id: string): Promise<void> {
        if (currentSessionId.value === id) return

        isLoading.value = true
        const chatStore = useChatStore()

        try {
            const response = await getSession(id)

            if (response.success) {
                currentSessionId.value = id
                localStorage.setItem(STORAGE_KEY_SESSION, id)

                // 加载历史消息到 chatStore
                chatStore.loadFromSession(response.messages)
            }
        } catch (err) {
            console.error('切换会话失败:', err)
        } finally {
            isLoading.value = false
        }
    }

    /**
     * 删除会话
     * @param id - 会话 ID
     */
    async function deleteSessionById(id: string): Promise<boolean> {
        try {
            const response = await apiDeleteSession(id)

            if (response.success) {
                sessions.value = sessions.value.filter(s => s.id !== id)

                // 如果删除的是当前会话，切换到其他会话
                if (currentSessionId.value === id) {
                    const chatStore = useChatStore()
                    chatStore.clearMessages()

                    if (sessions.value.length > 0 && sessions.value[0]) {
                        await switchSession(sessions.value[0].id)
                    } else {
                        currentSessionId.value = null
                        localStorage.removeItem(STORAGE_KEY_SESSION)
                        // 创建新会话
                        await createNewSession()
                    }
                }
                return true
            }
        } catch (err) {
            console.error('删除会话失败:', err)
        }
        return false
    }
    /**
     * 保存消息到当前会话
     * @param role - 消息角色
     * @param content - 消息内容
     */
    async function saveMessage(
        role: 'user' | 'assistant',
        content: string
    ): Promise<void> {
        if (!currentSessionId.value) return

        try {
            await addMessage(currentSessionId.value, { role, content })

            // 更新会话的 updated_at 和 message_count (不重新排序，避免列表滚动)
            const session = sessions.value.find(s => s.id === currentSessionId.value)
            if (session) {
                session.updated_at = new Date().toISOString()
                session.message_count = (session.message_count || 0) + 1
            }
        } catch (err) {
            console.error('保存消息失败:', err)
        }
    }

    /**
     * 根据首条用户消息更新会话标题
     * @param firstMessage - 首条消息内容
     */
    async function updateSessionTitle(firstMessage: string): Promise<void> {
        if (!currentSessionId.value) return

        // 截取前20个字符作为标题
        const title = firstMessage.length > MAX_TITLE_LENGTH
            ? firstMessage.substring(0, MAX_TITLE_LENGTH) + '...'
            : firstMessage

        try {
            const response = await apiUpdateSession(currentSessionId.value, { title })

            if (response.success && response.data) {
                const session = sessions.value.find(s => s.id === currentSessionId.value)
                if (session) {
                    session.title = response.data.title
                }
            }
        } catch (err) {
            console.error('更新会话标题失败:', err)
        }
    }

    return {
        // State
        sessions,
        currentSessionId,
        isLoading,
        isInitialized,
        // Computed
        currentSession,
        // Actions
        initialize,
        loadSessions,
        createNewSession,
        switchSession,
        deleteSession: deleteSessionById,
        saveMessage,
        updateSessionTitle
    }
})
