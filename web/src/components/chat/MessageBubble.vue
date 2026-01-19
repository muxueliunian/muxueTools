<script setup lang="ts">
/**
 * MessageBubble - 消息气泡组件
 * 
 * 职责: 渲染单条消息，支持 Markdown 和代码高亮
 */

import { computed, ref } from 'vue'
import { marked } from 'marked'
import hljs from 'highlight.js'
import 'highlight.js/styles/github-dark.css'
import type { ChatMessage } from '@/stores/chatStore'

interface Props {
    message: ChatMessage
}

const props = defineProps<Props>()

/** 复制状态追踪 (code block id -> isCopied) */
const copiedStates = ref<Record<string, boolean>>({})

/**
 * 配置 marked 渲染器
 * - 自定义代码块渲染，添加复制按钮
 */
const renderer = new marked.Renderer()

// 自定义代码块渲染
renderer.code = function({ text, lang }: { text: string; lang?: string }) {
    const language = lang && hljs.getLanguage(lang) ? lang : 'plaintext'
    const highlighted = hljs.highlight(text, { language }).value
    const codeId = `code-${Date.now()}-${Math.random().toString(36).substring(2, 7)}`
    
    return `
        <div class="code-block-wrapper group relative my-3">
            <div class="code-header flex items-center justify-between px-4 py-2 bg-[#1e1e1e] dark:bg-[#1a1a1a] rounded-t-lg border-b border-gray-700">
                <span class="text-xs text-gray-400">${language}</span>
                <button 
                    class="copy-btn text-gray-400 hover:text-white transition-colors p-1"
                    data-code-id="${codeId}"
                    onclick="window.__copyCode__('${codeId}')"
                >
                    <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                        <rect x="9" y="9" width="13" height="13" rx="2" ry="2"></rect>
                        <path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"></path>
                    </svg>
                </button>
            </div>
            <pre class="!m-0 !rounded-t-none"><code id="${codeId}" class="hljs language-${language}">${highlighted}</code></pre>
        </div>
    `
}

// 配置 marked
marked.setOptions({
    renderer,
    breaks: true,
    gfm: true,
})

/**
 * 渲染后的 HTML 内容
 */
const renderedContent = computed(() => {
    if (!props.message.content) return ''
    return marked.parse(props.message.content) as string
})

/**
 * 是否是用户消息
 */
const isUser = computed(() => props.message.role === 'user')

/**
 * 复制代码到剪贴板
 */
function copyCode(codeId: string): void {
    const codeElement = document.getElementById(codeId)
    if (codeElement) {
        const text = codeElement.textContent || ''
        navigator.clipboard.writeText(text).then(() => {
            copiedStates.value[codeId] = true
            setTimeout(() => {
                copiedStates.value[codeId] = false
            }, 2000)
        })
    }
}

// 暴露给全局的复制函数
if (typeof window !== 'undefined') {
    (window as unknown as { __copyCode__: (id: string) => void }).__copyCode__ = copyCode
}
</script>

<template>
    <div 
        class="message-bubble py-4"
        :class="isUser ? 'user-message' : 'assistant-message'"
    >
        <div 
            class="message-content max-w-none"
            :class="{
                'ml-auto text-right': isUser,
                'bg-[#F0EEEB] dark:bg-[#2A2A2E] rounded-2xl px-4 py-3': !isUser
            }"
        >
            <!-- 用户消息: 媒体 + 文本 -->
            <template v-if="isUser">
                <!-- 媒体预览 -->
                <div 
                    v-if="message.media && message.media.length > 0" 
                    class="flex flex-wrap gap-2 mb-2 justify-end"
                >
                    <template v-for="(media, idx) in message.media" :key="idx">
                        <!-- 图片：优先使用 dataUri（因为 previewUrl 可能被释放） -->
                        <img 
                            v-if="media.type === 'image'"
                            :src="media.dataUri || media.previewUrl"
                            :alt="media.name"
                            class="max-w-[200px] max-h-[200px] rounded-lg object-cover cursor-pointer border border-gray-200 dark:border-gray-600"
                        />
                        <!-- 视频：使用 dataUri 作为源 -->
                        <video 
                            v-else
                            :src="media.dataUri || media.previewUrl"
                            controls
                            class="max-w-[300px] max-h-[200px] rounded-lg border border-gray-200 dark:border-gray-600"
                        />
                    </template>
                </div>
                <!-- 文本内容 -->
                <p 
                    v-if="message.content"
                    class="text-claude-text dark:text-claude-dark-text whitespace-pre-wrap"
                >
                    {{ message.content }}
                </p>
            </template>
            
            <!-- 助手消息: Markdown 渲染 -->
            <template v-else>
                <div 
                    class="prose prose-sm dark:prose-invert max-w-none
                           prose-headings:font-serif prose-headings:font-medium
                           prose-p:text-claude-text dark:prose-p:text-claude-dark-text
                           prose-code:before:content-none prose-code:after:content-none
                           prose-code:bg-gray-200 dark:prose-code:bg-gray-700 
                           prose-code:px-1 prose-code:py-0.5 prose-code:rounded
                           prose-pre:bg-[#1e1e1e] prose-pre:rounded-lg"
                    v-html="renderedContent"
                />
            </template>
        </div>
    </div>
</template>

<style scoped>
.message-bubble {
    width: 100%;
}

.user-message .message-content {
    max-width: 85%;
    margin-left: auto;
}

.assistant-message .message-content {
    max-width: 100%;
}

/* 代码块样式 */
:deep(.code-block-wrapper) {
    border-radius: 8px;
    overflow: hidden;
}

:deep(.code-block-wrapper pre) {
    margin: 0;
    padding: 1rem;
    overflow-x: auto;
    background: #1e1e1e;
}

:deep(.code-block-wrapper code) {
    font-family: 'Fira Code', 'Monaco', 'Consolas', monospace;
    font-size: 0.875rem;
    line-height: 1.5;
}

:deep(.copy-btn) {
    opacity: 0;
    transition: opacity 0.2s;
}

:deep(.code-block-wrapper:hover .copy-btn) {
    opacity: 1;
}
</style>
