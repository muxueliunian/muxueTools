<script setup lang="ts">
/**
 * SessionItem - 单个会话项组件
 * 
 * 职责: 显示会话信息，支持点击切换和悬停删除
 */

import { computed } from 'vue'
import { Trash2 } from 'lucide-vue-next'
import { useI18n } from 'vue-i18n'
import type { Session } from '@/api/types'

interface Props {
    /** 会话数据 */
    session: Session
    /** 是否选中 */
    active?: boolean
}

interface Emits {
    (e: 'click'): void
    (e: 'delete'): void
}

const props = withDefaults(defineProps<Props>(), {
    active: false
})

const emit = defineEmits<Emits>()
const { t, locale } = useI18n()

/**
 * 格式化相对时间
 */
const relativeTime = computed(() => {
    const date = new Date(props.session.updated_at)
    const now = new Date()
    const diffMs = now.getTime() - date.getTime()
    const diffSecs = Math.floor(diffMs / 1000)
    const diffMins = Math.floor(diffSecs / 60)
    const diffHours = Math.floor(diffMins / 60)
    const diffDays = Math.floor(diffHours / 24)

    if (diffSecs < 60) return t('chat.justNow')
    if (diffMins < 60) return t('chat.minutesAgo', { n: diffMins })
    if (diffHours < 24) return t('chat.hoursAgo', { n: diffHours })
    if (diffDays < 7) return t('chat.daysAgo', { n: diffDays })
    
    // 根据当前语言格式化日期
    const localeMap: Record<string, string> = {
        'zh-CN': 'zh-CN',
        'en-US': 'en-US',
        'ja-JP': 'ja-JP'
    }
    return date.toLocaleDateString(localeMap[locale.value] || 'en-US', { month: 'short', day: 'numeric' })
})

/**
 * 截取后的标题
 */
const truncatedTitle = computed(() => {
    const title = props.session.title || t('chat.newChatTitle')
    if (title.length <= 20) return title
    return title.substring(0, 20) + '...'
})

/**
 * 处理点击事件
 */
function handleClick() {
    emit('click')
}

/**
 * 处理删除事件
 */
function handleDelete(e: MouseEvent) {
    e.stopPropagation()
    emit('delete')
}
</script>

<template>
    <div
        class="group relative flex flex-col px-3 py-2 rounded-md cursor-pointer transition-colors"
        :class="[
            active 
                ? 'bg-[#E1DFDD] dark:bg-[#303030]' 
                : 'hover:bg-claude-hover dark:hover:bg-claude-dark-hover'
        ]"
        @click="handleClick"
    >
        <!-- 左侧选中指示条 -->
        <div
            v-if="active"
            class="absolute left-0 top-1/2 -translate-y-1/2 w-[3px] h-4 bg-[#D97757] rounded-r"
        />

        <!-- 标题行 -->
        <div class="flex items-center justify-between gap-2">
            <span 
                class="text-sm font-medium truncate flex-1"
                :class="[
                    active 
                        ? 'text-claude-text dark:text-white' 
                        : 'text-claude-secondaryText dark:text-gray-400 group-hover:text-claude-text dark:group-hover:text-white'
                ]"
            >
                {{ truncatedTitle }}
            </span>

            <!-- 删除按钮 (悬停显示) -->
            <button
                class="opacity-0 group-hover:opacity-100 p-1 rounded transition-all hover:bg-red-100 dark:hover:bg-red-900/30"
                @click="handleDelete"
                :title="$t('chat.deleteSession')"
            >
                <Trash2 class="w-4 h-4 text-gray-400 hover:text-red-500 transition-colors" />
            </button>
        </div>

        <!-- 时间 -->
        <span class="text-xs text-claude-secondaryText dark:text-gray-500 mt-0.5">
            {{ relativeTime }}
        </span>
    </div>
</template>
