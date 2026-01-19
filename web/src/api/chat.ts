/**
 * Chat API - SSE 流式调用封装
 * 
 * 职责: 封装 /v1/chat/completions 的流式调用，解析 SSE 格式
 */

import type { ChatCompletionMessage, ChatCompletionMessageMultimodal, ChatCompletionChunk } from './types'

/** API 基础路径 */
const API_BASE = '/v1'

/**
 * 流式调用 /v1/chat/completions
 * 使用 fetch + ReadableStream 处理 SSE
 * 
 * @param messages - 对话消息数组 (支持纯文本或多模态)
 * @param model - 模型名称
 * @param signal - 可选的 AbortSignal 用于取消请求
 * @yields 每个响应片段的文本内容
 * 
 * @example
 * ```typescript
 * for await (const chunk of streamChatCompletion(messages, 'gemini-pro')) {
 *   console.log(chunk) // 逐字输出
 * }
 * ```
 */
export async function* streamChatCompletion(
    messages: (ChatCompletionMessage | ChatCompletionMessageMultimodal)[],
    model: string,
    signal?: AbortSignal
): AsyncGenerator<string, void, unknown> {
    const response = await fetch(`${API_BASE}/chat/completions`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({
            model,
            messages,
            stream: true,
        }),
        signal,
    })

    if (!response.ok) {
        const errorText = await response.text()
        throw new Error(`API 请求失败: ${response.status} - ${errorText}`)
    }

    const reader = response.body?.getReader()
    if (!reader) {
        throw new Error('无法获取响应流')
    }

    const decoder = new TextDecoder()
    let buffer = ''

    try {
        while (true) {
            const { done, value } = await reader.read()
            if (done) break

            // 将 bytes 解码并追加到缓冲区
            buffer += decoder.decode(value, { stream: true })

            // 按行分割处理 SSE 事件
            const lines = buffer.split('\n')
            // 保留最后一个可能不完整的行
            buffer = lines.pop() || ''

            for (const line of lines) {
                const trimmedLine = line.trim()

                // 跳过空行
                if (!trimmedLine) continue

                // 检测结束信号
                if (trimmedLine === 'data: [DONE]') {
                    return
                }

                // 解析 "data: {...}" 格式
                if (trimmedLine.startsWith('data: ')) {
                    const jsonStr = trimmedLine.slice(6) // 移除 "data: " 前缀
                    try {
                        const chunk: ChatCompletionChunk = JSON.parse(jsonStr)
                        const content = chunk.choices[0]?.delta?.content
                        if (content) {
                            yield content
                        }
                    } catch (parseError) {
                        // JSON 解析失败，忽略此行（可能是不完整的数据）
                        console.warn('SSE 解析警告:', parseError)
                    }
                }
            }
        }
    } finally {
        reader.releaseLock()
    }
}

/**
 * 非流式调用 /v1/chat/completions
 * 用于获取完整响应
 * 
 * @param messages - 对话消息数组
 * @param model - 模型名称
 * @returns 完整的助手回复内容
 */
export async function chatCompletion(
    messages: ChatCompletionMessage[],
    model: string
): Promise<string> {
    const response = await fetch(`${API_BASE}/chat/completions`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({
            model,
            messages,
            stream: false,
        }),
    })

    if (!response.ok) {
        const errorText = await response.text()
        throw new Error(`API 请求失败: ${response.status} - ${errorText}`)
    }

    const data = await response.json()
    return data.choices[0]?.message?.content || ''
}

// ==================== Models API ====================

/**
 * 获取可用模型列表
 * 调用 /api/models 接口，使用 Key Pool 中的有效 Key 查询 Gemini API 获取真实模型列表
 * 
 * @returns 模型名称数组 (e.g., ['gemini-2.0-flash', 'gemini-1.5-pro'])
 * @throws 网络错误或解析错误
 * 
 * @example
 * ```typescript
 * const models = await fetchAvailableModels()
 * console.log(models) // ['gemini-2.0-flash', 'gemini-1.5-pro', ...]
 * ```
 */
export async function fetchAvailableModels(): Promise<string[]> {
    const response = await fetch('/api/models')
    if (!response.ok) {
        throw new Error(`获取模型列表失败: ${response.status}`)
    }
    const data = await response.json()
    return data.data || []
}
