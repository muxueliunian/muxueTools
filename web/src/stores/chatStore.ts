/**
 * Chat Store - 对话状态管理
 * 
 * 职责: 管理 Chat 页面的消息列表、生成状态、模型选择和错误处理
 */

import { defineStore } from 'pinia'
import { ref } from 'vue'
import { streamChatCompletion, fetchAvailableModels } from '@/api/chat'
import type { ChatCompletionMessage, ChatCompletionMessageMultimodal, ContentPart, Message } from '@/api/types'

/** localStorage 键名 */
const STORAGE_KEY_MODEL = 'mxln_preferred_model'

/** 文件大小限制 1GB */
const MAX_FILE_SIZE = 1024 * 1024 * 1024

/**
 * 媒体文件 (图片/视频)
 */
export interface MediaFile {
    /** 文件名 */
    name: string
    /** 媒体类型 */
    type: 'image' | 'video'
    /** MIME 类型 */
    mimeType: string
    /** Base64 Data URI (发送用) */
    dataUri: string
    /** Blob URL (预览用) */
    previewUrl: string
}

/**
 * UI 层使用的消息格式
 */
export interface ChatMessage {
    /** 唯一标识 */
    id: string
    /** 消息角色 */
    role: 'user' | 'assistant'
    /** 消息内容 */
    content: string
    /** 附带的媒体文件 (仅用户消息) */
    media?: MediaFile[]
    /** 创建时间 */
    createdAt: Date
}

/**
 * 生成唯一 ID
 */
function generateId(): string {
    return `msg-${Date.now()}-${Math.random().toString(36).substring(2, 9)}`
}

export const useChatStore = defineStore('chat', () => {
    // ==================== State ====================

    /** 消息列表 */
    const messages = ref<ChatMessage[]>([])

    /** 是否正在生成 */
    const isGenerating = ref(false)

    /** 错误信息 */
    const error = ref<string | null>(null)

    /** 用于取消请求的 AbortController */
    let abortController: AbortController | null = null

    /** 当前使用的模型 */
    const currentModel = ref('')

    /** 可用模型列表 */
    const availableModels = ref<string[]>([])

    /** 模型加载状态 */
    const isLoadingModels = ref(false)

    /** 待发送的媒体文件队列 */
    const pendingMedia = ref<MediaFile[]>([])

    // ==================== Model Actions ====================

    /**
     * 从后端获取真实可用模型列表
     * 后端会使用 Key Pool 中的有效 Key 调用 Gemini API
     */
    async function loadModels(): Promise<void> {
        if (isLoadingModels.value) return
        isLoadingModels.value = true
        try {
            // 调用 /api/models 接口获取真实模型列表
            const models = await fetchAvailableModels()
            availableModels.value = models

            // 如果当前模型为空或不在列表中，选择第一个可用模型
            const firstModel = availableModels.value[0]
            if (firstModel &&
                (!currentModel.value || !availableModels.value.includes(currentModel.value))) {
                currentModel.value = firstModel
                saveModelPreference()
            }
        } catch (err) {
            console.error('加载模型列表失败:', err)
            // 设置默认模型列表作为后备
            availableModels.value = [
                'gemini-2.0-flash',
                'gemini-1.5-pro',
                'gemini-1.5-flash',
            ]
            if (!currentModel.value) {
                currentModel.value = 'gemini-2.0-flash'
            }
        } finally {
            isLoadingModels.value = false
        }
    }

    /**
     * 设置当前模型并持久化到 localStorage
     * @param modelId - 模型 ID
     */
    function setModel(modelId: string): void {
        currentModel.value = modelId
        saveModelPreference()
    }

    /**
     * 保存模型偏好到 localStorage
     */
    function saveModelPreference(): void {
        localStorage.setItem(STORAGE_KEY_MODEL, currentModel.value)
    }

    /**
     * 从 localStorage 恢复模型偏好
     */
    function restoreModelPreference(): void {
        const saved = localStorage.getItem(STORAGE_KEY_MODEL)
        if (saved) {
            currentModel.value = saved
        }
    }

    // ==================== Media Actions ====================

    /**
     * 添加待发送的媒体文件
     * @param file - 文件对象
     * @returns 是否添加成功 (失败时返回错误信息)
     */
    async function addPendingMedia(file: File): Promise<{ success: boolean; error?: string }> {
        // 检查文件大小
        if (file.size > MAX_FILE_SIZE) {
            return { success: false, error: `文件过大：${file.name} (最大 200MB)` }
        }

        // 判断媒体类型
        const isImage = file.type.startsWith('image/')
        const isVideo = file.type.startsWith('video/')
        if (!isImage && !isVideo) {
            return { success: false, error: `不支持的文件类型：${file.type}` }
        }

        // 读取为 Base64
        const dataUri = await readFileAsDataUri(file)

        // 创建预览 URL
        const previewUrl = URL.createObjectURL(file)

        pendingMedia.value.push({
            name: file.name,
            type: isImage ? 'image' : 'video',
            mimeType: file.type,
            dataUri,
            previewUrl,
        })

        return { success: true }
    }

    /**
     * 移除待发送的媒体文件
     */
    function removePendingMedia(index: number): void {
        const media = pendingMedia.value[index]
        if (media) {
            URL.revokeObjectURL(media.previewUrl)
            pendingMedia.value.splice(index, 1)
        }
    }

    /**
     * 清空待发送的媒体文件
     */
    function clearPendingMedia(): void {
        pendingMedia.value.forEach(m => URL.revokeObjectURL(m.previewUrl))
        pendingMedia.value = []
    }

    /**
     * 读取文件为 Data URI
     */
    function readFileAsDataUri(file: File): Promise<string> {
        return new Promise((resolve, reject) => {
            const reader = new FileReader()
            reader.onload = () => resolve(reader.result as string)
            reader.onerror = () => reject(new Error('读取文件失败'))
            reader.readAsDataURL(file)
        })
    }

    // ==================== Chat Actions ====================

    /**
     * 发送消息并获取流式响应
     * 
     * @param content - 用户输入的消息内容
     * @param model - 可选的模型名称，默认使用 currentModel
     */
    /**
     * 发送消息并获取流式响应
     * 
     * @param content - 用户输入的消息内容
     * @param model - 可选的模型名称，默认使用 currentModel
     * @param sessionStore - 可选的 sessionStore 实例，用于持久化消息
     */
    async function sendMessage(
        content: string,
        model?: string,
        sessionStore?: {
            saveMessage: (role: 'user' | 'assistant', content: string) => Promise<void>
            updateSessionTitle: (message: string) => Promise<void>
        }
    ): Promise<void> {
        if (isGenerating.value || !content.trim()) return

        const targetModel = model || currentModel.value
        error.value = null
        const trimmedContent = content.trim()

        // 检查是否是首条消息 (用于更新标题)
        const isFirstMessage = messages.value.length === 0

        // 1. 捕获当前待发送的媒体文件
        const mediaToSend = [...pendingMedia.value]
        clearPendingMedia()

        // 2. 添加用户消息 (包含媒体)
        const userMessage: ChatMessage = {
            id: generateId(),
            role: 'user',
            content: trimmedContent,
            media: mediaToSend.length > 0 ? mediaToSend : undefined,
            createdAt: new Date(),
        }
        messages.value.push(userMessage)

        // 保存用户消息到后端
        if (sessionStore) {
            await sessionStore.saveMessage('user', trimmedContent)
            // 首次发送时更新会话标题
            if (isFirstMessage) {
                await sessionStore.updateSessionTitle(trimmedContent)
            }
        }

        // 2. 创建助手消息占位
        const assistantMessage: ChatMessage = {
            id: generateId(),
            role: 'assistant',
            content: '',
            createdAt: new Date(),
        }
        messages.value.push(assistantMessage)
        // 获取助手消息在数组中的索引，用于后续响应式更新
        const assistantIndex = messages.value.length - 1

        // 3. 开始生成
        isGenerating.value = true
        abortController = new AbortController()

        try {
            // 将 ChatMessage 转换为 API 格式 (支持多模态)
            const apiMessages = messages.value
                .slice(0, -1) // 排除空的助手消息
                .map(msg => {
                    // 如果消息带有媒体，构造多模态格式
                    if (msg.media && msg.media.length > 0) {
                        const contentParts: ContentPart[] = []
                        // 先添加媒体
                        for (const m of msg.media) {
                            contentParts.push({
                                type: 'image_url',
                                image_url: { url: m.dataUri }
                            })
                        }
                        // 再添加文本
                        if (msg.content) {
                            contentParts.push({ type: 'text', text: msg.content })
                        }
                        return {
                            role: msg.role,
                            content: contentParts
                        } as ChatCompletionMessageMultimodal
                    }
                    // 纯文本消息
                    return {
                        role: msg.role,
                        content: msg.content,
                    } as ChatCompletionMessage
                })

            // 流式获取响应
            for await (const chunk of streamChatCompletion(
                apiMessages,
                targetModel,
                abortController.signal
            )) {
                // 通过索引访问并更新消息，确保 Vue 响应式更新
                const currentMessage = messages.value[assistantIndex]
                if (currentMessage) {
                    currentMessage.content += chunk
                    // 强制触发响应式更新
                    messages.value = [...messages.value]
                }
            }

            // 生成完成后保存助手消息到后端
            const finalMessage = messages.value[assistantIndex]
            if (sessionStore && finalMessage && finalMessage.content) {
                await sessionStore.saveMessage('assistant', finalMessage.content)
            }
        } catch (err) {
            // 处理取消请求
            if (err instanceof Error && err.name === 'AbortError') {
                // 用户主动取消，不显示错误
                return
            }

            // 其他错误
            const errorMessage = err instanceof Error ? err.message : '未知错误'
            error.value = errorMessage

            // 如果助手消息为空，移除它
            const currentMessage = messages.value[assistantIndex]
            if (currentMessage && !currentMessage.content) {
                messages.value.splice(assistantIndex, 1)
            }
        } finally {
            isGenerating.value = false
            abortController = null
        }
    }

    /**
     * 停止当前生成
     */
    function stopGeneration(): void {
        if (abortController) {
            abortController.abort()
            abortController = null
        }
        isGenerating.value = false
    }

    /**
     * 清空所有消息
     */
    function clearMessages(): void {
        messages.value = []
        error.value = null
    }

    /**
     * 从会话历史加载消息
     * @param sessionMessages - 后端返回的消息列表
     */
    function loadFromSession(sessionMessages: Message[]): void {
        messages.value = sessionMessages
            .filter(msg => msg.role === 'user' || msg.role === 'assistant')
            .map(msg => ({
                id: msg.id,
                role: msg.role as 'user' | 'assistant',
                content: typeof msg.content === 'string' ? msg.content : '',
                createdAt: new Date(msg.created_at)
            }))
        error.value = null
    }

    return {
        // State
        messages,
        isGenerating,
        error,
        currentModel,
        availableModels,
        isLoadingModels,
        pendingMedia,
        // Actions
        sendMessage,
        stopGeneration,
        clearMessages,
        loadFromSession,
        setModel,
        loadModels,
        restoreModelPreference,
        addPendingMedia,
        removePendingMedia,
        clearPendingMedia,
    }
})
