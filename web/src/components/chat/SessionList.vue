<script setup lang="ts">
/**
 * SessionList - 会话列表组件
 * 
 * 职责: 渲染会话列表，支持切换和删除操作
 */

import { ref } from 'vue'
import { NModal, NButton } from 'naive-ui'
import SessionItem from './SessionItem.vue'
import { useSessionStore } from '@/stores/sessionStore'

const sessionStore = useSessionStore()

/** 删除确认模态框状态 */
const showDeleteModal = ref(false)
const pendingDeleteId = ref<string | null>(null)

/**
 * 处理会话点击 - 切换会话
 */
function handleSessionClick(sessionId: string) {
    sessionStore.switchSession(sessionId)
}

/**
 * 处理删除会话 - 显示确认模态框
 */
function handleDeleteSession(sessionId: string) {
    pendingDeleteId.value = sessionId
    showDeleteModal.value = true
}

/**
 * 确认删除
 */
async function confirmDelete() {
    if (pendingDeleteId.value) {
        await sessionStore.deleteSession(pendingDeleteId.value)
    }
    showDeleteModal.value = false
    pendingDeleteId.value = null
}

/**
 * 取消删除
 */
function cancelDelete() {
    showDeleteModal.value = false
    pendingDeleteId.value = null
}
</script>

<template>
    <div class="flex-1 overflow-y-auto session-list-scroll">
        <!-- 加载状态 -->
        <div v-if="sessionStore.isLoading" class="px-3 py-4 text-center">
            <span class="text-sm text-claude-secondaryText dark:text-gray-500">加载中...</span>
        </div>

        <!-- 空状态 -->
        <div 
            v-else-if="sessionStore.sessions.length === 0" 
            class="px-3 py-4 text-center"
        >
            <span class="text-sm text-claude-secondaryText dark:text-gray-500">
                暂无会话
            </span>
        </div>

        <!-- 会话列表 -->
        <div v-else class="space-y-1 px-2">
            <SessionItem
                v-for="session in sessionStore.sessions"
                :key="session.id"
                :session="session"
                :active="session.id === sessionStore.currentSessionId"
                @click="handleSessionClick(session.id)"
                @delete="handleDeleteSession(session.id)"
            />
        </div>
    </div>

    <!-- 删除确认模态框 - Claude 风格 -->
    <n-modal 
        v-model:show="showDeleteModal"
        :mask-closable="true"
        preset="card"
        :bordered="false"
        size="small"
        class="delete-confirm-modal"
        style="width: 320px"
    >
        <template #header>
            <span class="text-base font-medium">Delete Chat</span>
        </template>
        <p class="text-sm text-claude-secondaryText dark:text-gray-400 mb-4">
            Are you sure you want to delete this chat? This action cannot be undone.
        </p>
        <div class="flex gap-2 justify-end">
            <n-button 
                @click="cancelDelete"
                size="small"
                quaternary
                class="!text-claude-secondaryText dark:!text-gray-400"
            >
                Cancel
            </n-button>
            <n-button 
                @click="confirmDelete"
                size="small"
                type="error"
                class="!bg-red-500 hover:!bg-red-600 !border-0"
            >
                Delete
            </n-button>
        </div>
    </n-modal>
</template>

<style scoped>
/* 隐藏滚动条 (参考 MessageList.vue) */
.session-list-scroll {
    scrollbar-width: none;
    -ms-overflow-style: none;
}

.session-list-scroll::-webkit-scrollbar {
    display: none;
}

/* 删除确认模态框样式覆盖 */
:deep(.delete-confirm-modal) {
    --n-color: var(--claude-bg, #FAF9F6);
    --n-border-radius: 12px;
}

.dark :deep(.delete-confirm-modal) {
    --n-color: #1a1a1a;
}

:deep(.n-card-header) {
    padding: 16px 20px 8px;
}

:deep(.n-card__content) {
    padding: 0 20px 16px;
}
</style>

