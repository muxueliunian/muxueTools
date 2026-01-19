<script setup lang="ts">
/**
 * MessageList - 消息列表组件
 * 
 * 职责: 渲染消息列表，自动滚动到底部，显示生成中状态
 */

import { ref, watch, nextTick, onMounted } from 'vue'
import MessageBubble from './MessageBubble.vue'
import type { ChatMessage } from '@/stores/chatStore'

interface Props {
    messages: ChatMessage[]
    isGenerating: boolean
}

const props = defineProps<Props>()

/** 滚动容器引用 */
const scrollContainer = ref<HTMLElement | null>(null)

/**
 * 滚动到底部
 */
function scrollToBottom(): void {
    nextTick(() => {
        if (scrollContainer.value) {
            scrollContainer.value.scrollTop = scrollContainer.value.scrollHeight
        }
    })
}

// 监听消息变化，自动滚动到底部
watch(
    () => props.messages.length,
    () => {
        scrollToBottom()
    }
)

// 监听最后一条消息内容变化（流式响应时）
watch(
    () => props.messages[props.messages.length - 1]?.content,
    () => {
        scrollToBottom()
    }
)

onMounted(() => {
    scrollToBottom()
})
</script>

<template>
    <div 
        ref="scrollContainer"
        class="message-list h-full overflow-y-auto"
    >
        <!-- 空状态 -->
        <div 
            v-if="messages.length === 0"
            class="flex items-center justify-center h-full pb-20"
        >
            <div class="flex items-center gap-4 opacity-100">
                <div class="bg-white dark:bg-[#333] p-2 rounded-xl shadow-sm">
                     <img src="/logo.png" class="w-10 h-10" />
                </div>
                <h2 class="text-3xl font-serif font-medium text-claude-text dark:text-gray-200 tracking-tight">
                    Back at it, muxueliunian
                </h2>
            </div>
        </div>

        <!-- 消息列表 -->
        <div v-else class="space-y-2">
            <MessageBubble 
                v-for="message in messages" 
                :key="message.id"
                :message="message"
            />
            
            <!-- 生成中的打字机光标 -->
            <div 
                v-if="isGenerating && messages[messages.length - 1]?.role === 'assistant'"
                class="typing-indicator flex items-center gap-1 px-4 py-2"
            >
                <span class="w-2 h-2 bg-[#D97757] rounded-full animate-pulse" />
            </div>
        </div>
    </div>
</template>

<style scoped>
.message-list {
    scroll-behavior: smooth;
    /* 隐藏滚动条但保留滚动功能 */
    scrollbar-width: none; /* Firefox */
    -ms-overflow-style: none; /* IE/Edge */
}

/* Chrome/Safari/Opera */
.message-list::-webkit-scrollbar {
    display: none;
}

.typing-indicator span {
    animation: typing-blink 1s ease-in-out infinite;
}

@keyframes typing-blink {
    0%, 100% {
        opacity: 0.3;
    }
    50% {
        opacity: 1;
    }
}
</style>
