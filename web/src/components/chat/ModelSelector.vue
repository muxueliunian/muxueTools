<script setup lang="ts">
/**
 * 模型选择器组件
 * 
 * 职责: 显示可用模型下拉列表，允许用户切换当前对话模型
 * 依赖: ChatStore
 */

import { computed } from 'vue'
import { NSelect } from 'naive-ui'
import { useChatStore } from '@/stores/chatStore'

const chatStore = useChatStore()

/**
 * 转换模型列表为 NSelect 选项格式
 */
const modelOptions = computed(() => 
    chatStore.availableModels.map(id => ({
        label: formatModelLabel(id),
        value: id
    }))
)

/**
 * 格式化模型名称为更友好的显示格式
 * 例: gemini-2.0-flash -> Gemini 2.0 Flash
 * 
 * @param modelId - 模型 ID
 * @returns 格式化后的模型名称
 */
function formatModelLabel(modelId: string): string {
    return modelId
        .split('-')
        .map(part => {
            // 处理纯数字部分（如版本号）
            if (/^\d+(\.\d+)?$/.test(part)) {
                return part
            }
            // 首字母大写
            return part.charAt(0).toUpperCase() + part.slice(1)
        })
        .join(' ')
}

/**
 * 处理模型选择变更
 */
function handleModelChange(value: string): void {
    chatStore.setModel(value)
}
</script>

<template>
    <n-select
        :value="chatStore.currentModel"
        :options="modelOptions"
        :loading="chatStore.isLoadingModels"
        size="small"
        style="width: 260px"
        placeholder="选择模型"
        @update:value="handleModelChange"
    />
</template>
