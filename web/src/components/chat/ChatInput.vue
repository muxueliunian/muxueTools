<script setup lang="ts">
/**
 * ChatInput - 输入框组件
 * 
 * 职责: Claude.ai 风格的输入框，支持多行自适应、快捷键和媒体上传
 */

import { ref, computed, nextTick } from 'vue'
import { NInput, NButton, NIcon } from 'naive-ui'
import { ArrowUp, StopCircle, ImageOutline, CloseCircle } from '@vicons/ionicons5'
import { useChatStore } from '@/stores/chatStore'
import { useMessage } from 'naive-ui'

interface Props {
    /** 是否正在生成 */
    loading: boolean
    /** 占位符文本 */
    placeholder?: string
}

interface Emits {
    (e: 'send', content: string): void
    (e: 'stop'): void
}

const props = withDefaults(defineProps<Props>(), {
    placeholder: '有什么可以帮到你的？',
})

const emit = defineEmits<Emits>()
const chatStore = useChatStore()
const message = useMessage()

/** 输入内容 */
const inputValue = ref('')

/** 输入框引用 */
const inputRef = ref<InstanceType<typeof NInput> | null>(null)

/** 文件选择器引用 */
const fileInputRef = ref<HTMLInputElement | null>(null)

/** 是否正在拖拽 */
const isDragging = ref(false)

/**
 * 是否可以发送 (有文本内容或有待发送媒体)
 */
const canSend = computed(() => 
    (inputValue.value.trim().length > 0 || chatStore.pendingMedia.length > 0) && !props.loading
)

/**
 * 处理发送
 */
function handleSend(): void {
    if (!canSend.value) return
    
    const content = inputValue.value.trim()
    inputValue.value = ''
    emit('send', content)
    
    // 聚焦输入框
    nextTick(() => {
        inputRef.value?.focus()
    })
}

/**
 * 处理停止
 */
function handleStop(): void {
    emit('stop')
}

/**
 * 处理按键事件
 * Enter 发送，Shift+Enter 换行
 */
function handleKeydown(event: KeyboardEvent): void {
    if (event.key === 'Enter' && !event.shiftKey) {
        event.preventDefault()
        handleSend()
    }
}

/**
 * 触发文件选择
 */
function triggerFileSelect(): void {
    fileInputRef.value?.click()
}

/**
 * 处理文件选择
 */
async function handleFileSelect(event: Event): Promise<void> {
    const target = event.target as HTMLInputElement
    const files = target.files
    if (files) {
        await processFiles(Array.from(files))
    }
    // 重置 input 以允许再次选择同一文件
    target.value = ''
}

/**
 * 处理拖拽进入
 */
function handleDragOver(event: DragEvent): void {
    event.preventDefault()
    isDragging.value = true
}

/**
 * 处理拖拽离开
 */
function handleDragLeave(): void {
    isDragging.value = false
}

/**
 * 处理拖放
 */
async function handleDrop(event: DragEvent): Promise<void> {
    event.preventDefault()
    isDragging.value = false
    
    const files = event.dataTransfer?.files
    if (files) {
        await processFiles(Array.from(files))
    }
}

/**
 * 处理多个文件
 */
async function processFiles(files: File[]): Promise<void> {
    for (const file of files) {
        const result = await chatStore.addPendingMedia(file)
        if (!result.success && result.error) {
            message.error(result.error)
        }
    }
}

/**
 * 处理粘贴事件 (支持 Ctrl+V 粘贴图片)
 */
async function handlePaste(event: ClipboardEvent): Promise<void> {
    const items = event.clipboardData?.items
    if (!items) return

    const files: File[] = []
    for (const item of items) {
        // 只处理图片和视频类型
        if (item.type.startsWith('image/') || item.type.startsWith('video/')) {
            const file = item.getAsFile()
            if (file) {
                files.push(file)
            }
        }
    }

    if (files.length > 0) {
        // 阻止默认粘贴行为（避免图片被粘贴为文本）
        event.preventDefault()
        await processFiles(files)
    }
}

/**
 * 移除待发送媒体
 */
function removeMedia(index: number): void {
    chatStore.removePendingMedia(index)
}
</script>

<template>
    <div 
        class="chat-input-wrapper"
        @dragover="handleDragOver"
        @dragleave="handleDragLeave"
        @drop="handleDrop"
    >
        <div 
            class="chat-input-container bg-white dark:bg-[#303030] border border-gray-200 dark:border-gray-600 rounded-2xl shadow-sm transition-colors"
            :class="{ 'border-dashed border-2 border-orange-400': isDragging }"
        >
            <!-- 媒体预览区 -->
            <div 
                v-if="chatStore.pendingMedia.length > 0" 
                class="flex flex-wrap gap-2 px-4 pt-3"
            >
                <div 
                    v-for="(media, index) in chatStore.pendingMedia" 
                    :key="index"
                    class="relative group"
                >
                    <!-- 图片预览 -->
                    <img 
                        v-if="media.type === 'image'"
                        :src="media.previewUrl"
                        :alt="media.name"
                        class="w-20 h-20 object-cover rounded-lg border border-gray-200 dark:border-gray-600"
                    />
                    <!-- 视频预览 -->
                    <video 
                        v-else
                        :src="media.previewUrl"
                        class="w-20 h-20 object-cover rounded-lg border border-gray-200 dark:border-gray-600"
                        muted
                    />
                    <!-- 文件类型标签 -->
                    <span 
                        v-if="media.type === 'video'"
                        class="absolute bottom-1 left-1 text-xs bg-black/60 text-white px-1 rounded"
                    >
                        视频
                    </span>
                    <!-- 删除按钮 -->
                    <button 
                        class="absolute -top-2 -right-2 opacity-0 group-hover:opacity-100 transition-opacity bg-white dark:bg-gray-700 rounded-full shadow"
                        @click="removeMedia(index)"
                    >
                        <n-icon :size="20" color="#999">
                            <CloseCircle />
                        </n-icon>
                    </button>
                </div>
            </div>
            
            <div class="px-4 pt-3 pb-2">
                <n-input
                    ref="inputRef"
                    v-model:value="inputValue"
                    type="textarea"
                    :placeholder="placeholder"
                    :autosize="{ minRows: 1, maxRows: 8 }"
                    :bordered="false"
                    class="!bg-transparent text-base"
                    @keydown="handleKeydown"
                    @paste="handlePaste"
                />
            </div>
            
            <div class="flex justify-between items-center px-4 pb-3">
                <!-- 左侧：媒体上传按钮 -->
                <div class="flex gap-2">
                    <n-button 
                        quaternary 
                        circle 
                        size="small"
                        @click="triggerFileSelect"
                    >
                        <template #icon>
                            <n-icon :size="18" color="#888">
                                <ImageOutline />
                            </n-icon>
                        </template>
                    </n-button>
                    <input 
                        ref="fileInputRef"
                        type="file" 
                        accept="image/*,video/*" 
                        multiple
                        hidden 
                        @change="handleFileSelect"
                    />
                </div>
                
                <!-- 发送/停止按钮 -->
                <div>
                    <!-- 生成中显示停止按钮 -->
                    <n-button 
                        v-if="loading"
                        circle 
                        secondary
                        type="error"
                        @click="handleStop"
                    >
                        <template #icon>
                            <n-icon :size="18">
                                <StopCircle />
                            </n-icon>
                        </template>
                    </n-button>
                    
                    <!-- 发送按钮 -->
                    <n-button 
                        v-else
                        circle 
                        :disabled="!canSend"
                        :style="{
                            backgroundColor: canSend ? '#D97757' : '#E5E5E5',
                            borderColor: canSend ? '#D97757' : '#E5E5E5',
                        }"
                        @click="handleSend"
                    >
                        <template #icon>
                            <n-icon 
                                :size="18" 
                                :color="canSend ? 'white' : '#999'"
                            >
                                <ArrowUp />
                            </n-icon>
                        </template>
                    </n-button>
                </div>
            </div>
        </div>
    </div>
</template>

<style scoped>
.chat-input-wrapper {
    max-width: 100%;
}

.chat-input-container {
    transition: box-shadow 0.2s, border-color 0.2s;
}

.chat-input-container:focus-within {
    border-color: #D97757;
    box-shadow: 0 0 0 2px rgba(217, 119, 87, 0.1);
}

:deep(.n-input) {
    --n-border: none !important;
    --n-border-hover: none !important;
    --n-border-focus: none !important;
    --n-box-shadow-focus: none !important;
}

:deep(.n-input__textarea-el) {
    resize: none !important;
}
</style>
