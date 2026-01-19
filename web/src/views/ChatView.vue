<script setup lang="ts">
/**
 * ChatView - Chat 对话页面
 * 
 * 职责: 组装 Chat 组件，处理消息发送和错误显示
 * 依赖: useChatStore, useSessionStore
 */

import { watch } from 'vue'
import { useMessage } from 'naive-ui'
import { useChatStore } from '@/stores/chatStore'
import { useSessionStore } from '@/stores/sessionStore'
import MessageList from '@/components/chat/MessageList.vue'
import ChatInput from '@/components/chat/ChatInput.vue'
import ModelSelector from '@/components/chat/ModelSelector.vue'

const chatStore = useChatStore()
const sessionStore = useSessionStore()
const message = useMessage()

/**
 * 处理发送消息 - 传递 sessionStore 以持久化消息
 */
async function handleSend(content: string): Promise<void> {
    await chatStore.sendMessage(content, undefined, sessionStore)
}

/**
 * 处理停止生成
 */
function handleStop(): void {
    chatStore.stopGeneration()
    message.info('已停止生成')
}

// 监听错误并显示 Toast
watch(
    () => chatStore.error,
    (error) => {
        if (error) {
            message.error(error)
        }
    }
)
</script>

<template>
    <div class="chat-view flex flex-col h-full bg-claude-bg dark:bg-claude-dark-bg transition-colors">
        <!-- 消息区域 (可滚动) -->
        <div class="flex-1 overflow-hidden">
            <div class="h-full max-w-3xl mx-auto px-4 py-4">
                <MessageList 
                    :messages="chatStore.messages" 
                    :is-generating="chatStore.isGenerating"
                />
            </div>
        </div>
        
        <!-- 输入区域 (固定底部) -->
        <div class="flex-shrink-0 border-t border-gray-200 dark:border-gray-700 bg-claude-bg dark:bg-claude-dark-bg transition-colors">
            <div class="max-w-3xl mx-auto px-4 py-4">
                <!-- 模型选择器和提示 -->
                <div class="flex items-center gap-4 mb-3">
                    <ModelSelector />
                    <span class="text-xs text-claude-secondaryText dark:text-gray-500">
                        按 Enter 发送，Shift + Enter 换行
                    </span>
                </div>
                
                <!-- 输入框 -->
                <ChatInput 
                    :loading="chatStore.isGenerating"
                    @send="handleSend" 
                    @stop="handleStop"
                />
            </div>
        </div>
    </div>
</template>

<style scoped>
.chat-view {
    /* 确保填满父容器 */
    min-height: 0;
}
</style>
