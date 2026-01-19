<script setup lang="ts">
/**
 * Settings View
 * Responsibility: System configuration and update management with tabbed interface.
 * Dependencies: Config API
 */
import { ref, onMounted, computed } from 'vue'
import { NCard, NForm, NFormItem, NSelect, NSwitch, NButton, NRadioGroup, NRadio, NInput, NInputNumber, NModal, useMessage, NDivider, NSlider } from 'naive-ui'
import { getConfig, updateConfig, checkUpdate, regenerateProxyKey, clearAllSessions, resetStats, type ConfigInfo, type UpdateInfo, type ModelSettingsConfig } from '../api/config'
import { Save, CheckCircle2, Eye, EyeOff, RefreshCw, Shield, Trash2, Cpu } from 'lucide-vue-next'
import { useGlobalStore } from '@/stores/global'

const globalStore = useGlobalStore()

const message = useMessage()
const loading = ref(false)
const checkingUpdate = ref(false)
const regeneratingKey = ref(false)
const activeTab = ref<'general' | 'security' | 'advanced' | 'model'>('general')
const deletingChats = ref(false)
const deletingStats = ref(false)
const showDeleteChatsModal = ref(false)
const showDeleteStatsModal = ref(false)

const config = ref<ConfigInfo>({
    server: { port: 8080, host: '0.0.0.0' },
    pool: { strategy: 'round_robin', cooldown_seconds: 3600, max_retries: 3 },
    logging: { level: 'info' },
    update: { enabled: true, check_interval: '24h' },
    security: { ip_whitelist_enabled: false, whitelist_ip: '', proxy_key: 'sk-mxln-proxy-local' },
    advanced: { request_timeout: 120 }
})

const updateInfo = ref<UpdateInfo | null>(null)
const showProxyKey = ref(false)

// Security form fields (separate from config for easier binding)
const ipWhitelistEnabled = ref(false)
const whitelistIP = ref('')
const proxyKey = ref('sk-mxln-proxy-local')

// Model settings form fields
const modelSettings = ref<ModelSettingsConfig>({
    system_prompt: '',
    temperature: null,
    max_output_tokens: null,
    top_p: null,
    top_k: null,
    thinking_level: null,
    media_resolution: null
})

const thinkingLevelOptions = [
    { label: 'Disabled', value: '' },
    { label: 'Low', value: 'LOW' },
    { label: 'Medium', value: 'MEDIUM' },
    { label: 'High', value: 'HIGH' }
]

const mediaResolutionOptions = [
    { label: 'Default', value: '' },
    { label: 'Low (64 tokens)', value: 'MEDIA_RESOLUTION_LOW' },
    { label: 'Medium (256 tokens)', value: 'MEDIA_RESOLUTION_MEDIUM' },
    { label: 'High (scaling)', value: 'MEDIA_RESOLUTION_HIGH' }
]

const strategyOptions = [
    { label: 'Round Robin (Sequential)', value: 'round_robin' },
    { label: 'Random Selection', value: 'random' },
    { label: 'Least Used First', value: 'least_used' },
    { label: 'Weighted Random', value: 'weighted' }
]

const logLevelOptions = [
    { label: 'Debug (Verbose)', value: 'debug' },
    { label: 'Info (Standard)', value: 'info' },
    { label: 'Warning', value: 'warn' },
    { label: 'Error (Critical only)', value: 'error' }
]

const updateSourceOptions = [
    { label: 'mxln Server (推荐)', value: 'mxln' },
    { label: 'GitHub', value: 'github' }
]

// 更新源选择
const updateSource = ref<'mxln' | 'github'>('mxln')

// Masked proxy key for display
const maskedProxyKey = computed(() => {
    if (showProxyKey.value) return proxyKey.value
    if (proxyKey.value.length <= 12) return '•'.repeat(proxyKey.value.length)
    return proxyKey.value.slice(0, 8) + '•'.repeat(8) + proxyKey.value.slice(-4)
})

async function loadConfig() {
    loading.value = true
    try {
        const res = await getConfig()
        if (res.success && res.data) {
            config.value = res.data
            // Sync security fields
            if (res.data.security) {
                ipWhitelistEnabled.value = res.data.security.ip_whitelist_enabled
                whitelistIP.value = res.data.security.whitelist_ip || ''
                proxyKey.value = res.data.security.proxy_key || 'sk-mxln-proxy-local'
            }
            // Sync update source
            if (res.data.update?.source) {
                updateSource.value = res.data.update.source
            }
            // Sync model settings
            if (res.data.model_settings) {
                modelSettings.value = {
                    system_prompt: res.data.model_settings.system_prompt || '',
                    temperature: res.data.model_settings.temperature ?? null,
                    max_output_tokens: res.data.model_settings.max_output_tokens ?? null,
                    top_p: res.data.model_settings.top_p ?? null,
                    top_k: res.data.model_settings.top_k ?? null,
                    thinking_level: res.data.model_settings.thinking_level ?? null,
                    media_resolution: res.data.model_settings.media_resolution ?? null
                }
            }
        }
    } catch (e: any) {
        console.error('Failed to load config:', e)
        const errorMsg = e.response?.data?.message || e.message || 'Unknown error'
        message.error(`Failed to load configuration: ${errorMsg}`)
    } finally {
        loading.value = false
    }
}

async function handleSave() {
    loading.value = true
    try {
        // Build the update request matching backend UpdateConfigRequest structure
        const configToSave: Record<string, unknown> = {}
        
        // Pool configuration
        configToSave.pool = {
            strategy: config.value.pool.strategy,
            cooldown_seconds: config.value.pool.cooldown_seconds,
            max_retries: config.value.pool.max_retries
        }
        
        // Logging configuration
        configToSave.logging = {
            level: config.value.logging.level
        }
        
        // Update configuration
        configToSave.update = {
            enabled: config.value.update.enabled,
            source: updateSource.value
        }
        
        // Security configuration
        configToSave.security = {
            ip_whitelist_enabled: ipWhitelistEnabled.value,
            whitelist_ip: whitelistIP.value,
            proxy_key: proxyKey.value
        }
        
        // Advanced configuration
        if (config.value.advanced) {
            configToSave.advanced = {
                request_timeout: config.value.advanced.request_timeout
            }
        }
        
        // Model settings configuration
        configToSave.model_settings = {
            system_prompt: modelSettings.value.system_prompt,
            temperature: modelSettings.value.temperature,
            max_output_tokens: modelSettings.value.max_output_tokens,
            top_p: modelSettings.value.top_p,
            top_k: modelSettings.value.top_k,
            thinking_level: modelSettings.value.thinking_level,
            media_resolution: modelSettings.value.media_resolution
        }
        
        const res = await updateConfig(configToSave as Partial<ConfigInfo>)
        if (res.success) {
            message.success('Configuration saved successfully')
        } else {
            message.error('Failed to save configuration')
        }
    } catch (e) {
        message.error('Network error during save')
    } finally {
        loading.value = false
    }
}

async function handleCheckUpdate() {
    checkingUpdate.value = true
    try {
        const res = await checkUpdate()
        if (res.success && res.data) {
            updateInfo.value = res.data
            if (!res.data.has_update) {
                message.info('You are using the latest version')
            }
        }
    } catch (e) {
        message.error('Failed to check for updates')
    } finally {
        checkingUpdate.value = false
    }
}

async function handleRegenerateKey() {
    regeneratingKey.value = true
    try {
        const res = await regenerateProxyKey()
        if (res.success && res.data) {
            proxyKey.value = res.data.proxy_key
            showProxyKey.value = true
            message.success('Proxy key regenerated successfully')
        }
    } catch (e) {
        message.error('Failed to regenerate proxy key')
    } finally {
        regeneratingKey.value = false
    }
}

async function handleDeleteChats() {
    deletingChats.value = true
    try {
        const res = await clearAllSessions()
        if (res.success) {
            message.success(`Deleted ${res.data?.deleted || 0} sessions successfully`)
            showDeleteChatsModal.value = false
        }
    } catch (e) {
        message.error('Failed to delete chat history')
    } finally {
        deletingChats.value = false
    }
}

async function handleResetStats() {
    deletingStats.value = true
    try {
        const res = await resetStats()
        if (res.success) {
            message.success(`Reset statistics for ${res.data?.keys_affected || 0} keys`)
            showDeleteStatsModal.value = false
        }
    } catch (e) {
        message.error('Failed to reset statistics')
    } finally {
        deletingStats.value = false
    }
}

onMounted(() => {
    loadConfig()
})
</script>

<template>
    <div :class="{ 'dark': globalStore.isDark }" class="min-h-screen bg-claude-bg dark:bg-claude-dark-bg text-claude-text dark:text-gray-200 p-8 font-sans transition-colors duration-200">
        <div class="max-w-4xl mx-auto space-y-6">
            <!-- Header with inline tabs -->
            <div class="flex items-end justify-between border-b border-claude-border dark:border-[#2A2A2E] pb-4">
                <div>
                    <h1 class="text-3xl font-light text-claude-text dark:text-white tracking-tight mb-1">Settings</h1>
                    <p class="text-claude-secondaryText dark:text-gray-500 text-sm">Configure system behavior and performance.</p>
                </div>
                <!-- Inline Tab Navigation -->
                <div class="flex gap-1 bg-gray-100 dark:bg-[#191919] rounded-lg p-1">
                    <button 
                        @click="activeTab = 'general'"
                        :class="[
                            'px-4 py-2 rounded-md text-sm font-medium transition-all duration-200',
                            activeTab === 'general' 
                                ? 'bg-white dark:bg-[#2A2A2E] text-claude-text dark:text-white shadow-sm' 
                                : 'text-claude-secondaryText dark:text-gray-500 hover:text-claude-text dark:hover:text-gray-300'
                        ]"
                    >General</button>
                    <button 
                        @click="activeTab = 'security'"
                        :class="[
                            'px-4 py-2 rounded-md text-sm font-medium transition-all duration-200 flex items-center gap-1.5',
                            activeTab === 'security' 
                                ? 'bg-white dark:bg-[#2A2A2E] text-claude-text dark:text-white shadow-sm' 
                                : 'text-claude-secondaryText dark:text-gray-500 hover:text-claude-text dark:hover:text-gray-300'
                        ]"
                    >
                        <Shield class="w-3.5 h-3.5" />
                        Security
                    </button>
                    <button 
                        @click="activeTab = 'advanced'"
                        :class="[
                            'px-4 py-2 rounded-md text-sm font-medium transition-all duration-200',
                            activeTab === 'advanced' 
                                ? 'bg-white dark:bg-[#2A2A2E] text-claude-text dark:text-white shadow-sm' 
                                : 'text-claude-secondaryText dark:text-gray-500 hover:text-claude-text dark:hover:text-gray-300'
                        ]"
                    >Advanced</button>
                    <button 
                        @click="activeTab = 'model'"
                        :class="[
                            'px-4 py-2 rounded-md text-sm font-medium transition-all duration-200 flex items-center gap-1.5',
                            activeTab === 'model' 
                                ? 'bg-white dark:bg-[#2A2A2E] text-claude-text dark:text-white shadow-sm' 
                                : 'text-claude-secondaryText dark:text-gray-500 hover:text-claude-text dark:hover:text-gray-300'
                        ]"
                    >
                        <Cpu class="w-3.5 h-3.5" />
                        Model
                    </button>
                </div>
            </div>

            <!-- Main Content -->
            <div class="space-y-6">
                <!-- General Tab -->
                <template v-if="activeTab === 'general'">
                    <!-- Strategy Section -->
                    <n-card class="!bg-white dark:!bg-[#212124] !border-claude-border dark:!border-[#2A2A2E] !text-claude-text dark:!text-gray-200 shadow-sm transition-colors duration-200" title="Key Management">
                        <n-form label-placement="top" class="mt-2">
                             <n-form-item label="Selection Strategy">
                                <n-select v-model:value="config.pool.strategy" :options="strategyOptions" />
                                <template #feedback>
                                    <span class="text-xs text-claude-secondaryText dark:text-gray-500">Algorithm used to select the next available API key.</span>
                                </template>
                            </n-form-item>
                        </n-form>
                        </n-card>

                         <!-- Logging Section -->
                        <n-card class="!bg-white dark:!bg-[#212124] !border-claude-border dark:!border-[#2A2A2E] !text-claude-text dark:!text-gray-200 shadow-sm transition-colors duration-200" title="Logging & Updates">
                             <n-form label-placement="top" class="mt-2">
                                <n-form-item label="Log Level">
                                    <n-select v-model:value="config.logging.level" :options="logLevelOptions" class="anthropic-select" />
                                </n-form-item>
                                <n-divider class="!my-4 !bg-claude-border dark:!bg-[#2A2A2E]" />
                                <div class="flex items-center justify-between">
                                    <div>
                                        <div class="text-sm font-medium text-claude-text dark:text-gray-200">Automatic Updates</div>
                                        <div class="text-xs text-claude-secondaryText dark:text-gray-500">Check for new versions on startup.</div>
                                    </div>
                                    <div class="flex gap-4 items-center">
                                        <n-button size="small" tertiary class="!text-gray-400 hover:!text-white" @click="handleCheckUpdate" :loading="checkingUpdate">Check Now</n-button>
                                        <n-switch v-model:value="config.update.enabled" :rail-style="({ checked }) => ({ backgroundColor: checked ? '#D97757' : '#4B5563' })" />
                                    </div>
                                </div>
                                
                                <!-- Update Source Selection -->
                                <div class="mt-4 pt-4 border-t border-claude-border dark:border-[#2A2A2E]">
                                    <div class="text-sm font-medium text-claude-text dark:text-gray-200 mb-2">Update Source</div>
                                    <n-radio-group v-model:value="updateSource" name="update-source" class="flex gap-4">
                                        <n-radio 
                                            v-for="opt in updateSourceOptions" 
                                            :key="opt.value" 
                                            :value="opt.value"
                                            class="!text-claude-text dark:!text-gray-300"
                                        >
                                            {{ opt.label }}
                                        </n-radio>
                                    </n-radio-group>
                                    <div class="text-xs text-claude-secondaryText dark:text-gray-500 mt-1">
                                        {{ updateSource === 'mxln' ? '使用 mxln 服务器检查更新 (中国用户推荐)' : '使用 GitHub Releases 检查更新' }}
                                    </div>
                                </div>
                            </n-form>

                            <div v-if="updateInfo && updateInfo.has_update" class="mt-6 p-4 bg-gray-50 dark:bg-[#2A2A2E] rounded border border-emerald-500/20">
                                <div class="flex items-start gap-3">
                                    <CheckCircle2 class="text-emerald-500 w-5 h-5 mt-0.5" />
                                    <div>
                                        <div class="text-emerald-500 font-medium">Update Available: v{{ updateInfo.latest_version }}</div>
                                        <div class="text-gray-400 text-xs mt-1">{{ updateInfo.changelog || 'Performance improvements and bug fixes.' }}</div>
                                        <a v-if="updateInfo.download_url" :href="updateInfo.download_url" target="_blank" class="text-emerald-400 text-xs mt-2 inline-block hover:underline">Download Update &rarr;</a>
                                    </div>
                                </div>
                            </div>
                        </n-card>
                    </template>

                    <!-- Security Tab -->
                    <template v-if="activeTab === 'security'">
                        <n-card class="!bg-white dark:!bg-[#212124] !border-claude-border dark:!border-[#2A2A2E] !text-claude-text dark:!text-gray-200 shadow-sm transition-colors duration-200" title="Access Control">
                            <n-form label-placement="top" class="mt-2">
                                <!-- IP Whitelist -->
                                <div class="flex items-center justify-between mb-4">
                                    <div>
                                        <div class="text-sm font-medium text-claude-text dark:text-gray-200">IP Whitelist</div>
                                        <div class="text-xs text-claude-secondaryText dark:text-gray-500">Only allow requests from specific IP address.</div>
                                    </div>
                                    <n-switch v-model:value="ipWhitelistEnabled" :rail-style="({ checked }) => ({ backgroundColor: checked ? '#D97757' : '#4B5563' })" />
                                </div>

                                <n-form-item label="Allowed IP Address" v-if="ipWhitelistEnabled">
                                    <n-input 
                                        v-model:value="whitelistIP" 
                                        placeholder="e.g., 192.168.1.100"
                                        class="!bg-gray-50 dark:!bg-[#191919]"
                                    />
                                    <template #feedback>
                                        <span class="text-xs text-claude-secondaryText dark:text-gray-500">Localhost (127.0.0.1) is always allowed to prevent lockout.</span>
                                    </template>
                                </n-form-item>

                                <n-divider class="!my-4 !bg-claude-border dark:!bg-[#2A2A2E]" />

                                <!-- Proxy API Key -->
                                <n-form-item label="Proxy API Key">
                                    <div class="flex gap-2 w-full">
                                        <div class="relative flex-1">
                                            <n-input 
                                                :value="maskedProxyKey" 
                                                readonly
                                                class="!bg-gray-50 dark:!bg-[#191919] !pr-10"
                                            />
                                            <button 
                                                @click="showProxyKey = !showProxyKey"
                                                class="absolute right-2 top-1/2 -translate-y-1/2 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300"
                                            >
                                                <Eye v-if="!showProxyKey" class="w-4 h-4" />
                                                <EyeOff v-else class="w-4 h-4" />
                                            </button>
                                        </div>
                                        <n-button 
                                            @click="handleRegenerateKey" 
                                            :loading="regeneratingKey"
                                            tertiary
                                            class="!text-[#D97757] hover:!bg-[#D97757]/10"
                                        >
                                            <template #icon>
                                                <RefreshCw class="w-4 h-4" />
                                            </template>
                                            Regenerate
                                        </n-button>
                                    </div>
                                    <template #feedback>
                                        <span class="text-xs text-claude-secondaryText dark:text-gray-500">Used to authenticate requests to this proxy. Share with authorized users only.</span>
                                    </template>
                                </n-form-item>
                            </n-form>
                        </n-card>
                    </template>

                    <!-- Advanced Tab -->
                    <template v-if="activeTab === 'advanced'">
                        <!-- Configuration Parameters -->
                        <n-card class="!bg-white dark:!bg-[#212124] !border-claude-border dark:!border-[#2A2A2E] !text-claude-text dark:!text-gray-200 shadow-sm transition-colors duration-200" title="Performance Tuning">
                            <n-form label-placement="top" class="mt-2">
                                <div class="grid grid-cols-3 gap-4">
                                    <div class="space-y-1">
                                        <div class="text-sm font-medium text-claude-text dark:text-gray-200">Cooldown Time</div>
                                        <n-input-number 
                                            v-model:value="config.pool.cooldown_seconds" 
                                            :min="60" 
                                            :max="86400"
                                            class="!w-full"
                                        >
                                            <template #suffix>
                                                <span class="text-xs text-gray-400">sec</span>
                                            </template>
                                        </n-input-number>
                                        <div class="text-xs text-claude-secondaryText dark:text-gray-500">Rate limit cooldown</div>
                                    </div>
                                    <div class="space-y-1">
                                        <div class="text-sm font-medium text-claude-text dark:text-gray-200">Max Retries</div>
                                        <n-input-number 
                                            v-model:value="config.pool.max_retries" 
                                            :min="0" 
                                            :max="10"
                                            class="!w-full"
                                        />
                                        <div class="text-xs text-claude-secondaryText dark:text-gray-500">Retry on failure</div>
                                    </div>
                                    <div class="space-y-1">
                                        <div class="text-sm font-medium text-claude-text dark:text-gray-200">Request Timeout</div>
                                        <n-input-number 
                                            v-model:value="config.advanced!.request_timeout" 
                                            :min="30" 
                                            :max="600"
                                            class="!w-full"
                                        >
                                            <template #suffix>
                                                <span class="text-xs text-gray-400">sec</span>
                                            </template>
                                        </n-input-number>
                                        <div class="text-xs text-claude-secondaryText dark:text-gray-500">API request timeout</div>
                                    </div>
                                </div>

                                <n-divider class="!my-4 !bg-claude-border dark:!bg-[#2A2A2E]" />

                                <div class="flex items-center justify-between">
                                    <div>
                                        <div class="text-sm font-medium text-claude-text dark:text-gray-200">Debug Mode</div>
                                        <div class="text-xs text-claude-secondaryText dark:text-gray-500">Enable verbose logging output.</div>
                                    </div>
                                    <n-switch 
                                        :value="config.logging.level === 'debug'"
                                        @update:value="(v: boolean) => config.logging.level = v ? 'debug' : 'info'"
                                        :rail-style="({ checked }) => ({ backgroundColor: checked ? '#D97757' : '#4B5563' })" 
                                    />
                                </div>
                            </n-form>
                        </n-card>

                        <!-- Data Management -->
                        <n-card class="!bg-white dark:!bg-[#212124] !border-claude-border dark:!border-[#2A2A2E] !text-claude-text dark:!text-gray-200 shadow-sm transition-colors duration-200" title="Data Management">
                            <n-form label-placement="top" class="mt-2">
                                <n-form-item label="Database Location">
                                    <n-input 
                                        value="./data/muxueTools.db" 
                                        readonly
                                        class="!bg-gray-50 dark:!bg-[#191919]"
                                    />
                                    <template #feedback>
                                        <span class="text-xs text-claude-secondaryText dark:text-gray-500">SQLite database file path (read-only)</span>
                                    </template>
                                </n-form-item>

                                <n-divider class="!my-4 !bg-claude-border dark:!bg-[#2A2A2E]" />

                                <div class="text-sm font-medium text-claude-text dark:text-gray-200 mb-3">Danger Zone</div>
                                <div class="space-y-3">
                                    <div class="flex items-center justify-between p-3 border border-red-500/20 rounded bg-red-500/5">
                                        <div>
                                            <div class="text-sm text-red-400">Delete Chat History</div>
                                            <div class="text-xs text-gray-500">Remove all chat sessions and messages.</div>
                                        </div>
                                        <n-button 
                                            @click="showDeleteChatsModal = true"
                                            size="small"
                                            type="error"
                                            ghost
                                        >
                                            <template #icon>
                                                <Trash2 class="w-4 h-4" />
                                            </template>
                                            Delete
                                        </n-button>
                                    </div>
                                    <div class="flex items-center justify-between p-3 border border-red-500/20 rounded bg-red-500/5">
                                        <div>
                                            <div class="text-sm text-red-400">Reset Statistics</div>
                                            <div class="text-xs text-gray-500">Clear all API key usage statistics.</div>
                                        </div>
                                        <n-button 
                                            @click="showDeleteStatsModal = true"
                                            size="small"
                                            type="error"
                                            ghost
                                        >
                                            <template #icon>
                                                <Trash2 class="w-4 h-4" />
                                            </template>
                                            Reset
                                        </n-button>
                                    </div>
                                </div>
                            </n-form>
                        </n-card>
                    </template>

                    <!-- Model Tab -->
                    <template v-if="activeTab === 'model'">
                        <!-- System Prompt -->
                        <n-card class="!bg-white dark:!bg-[#212124] !border-claude-border dark:!border-[#2A2A2E] !text-claude-text dark:!text-gray-200 shadow-sm transition-colors duration-200" title="System Prompt">
                            <n-form label-placement="top" class="mt-2">
                                <n-form-item label="Default System Prompt">
                                    <n-input 
                                        v-model:value="modelSettings.system_prompt" 
                                        type="textarea"
                                        :autosize="{ minRows: 3, maxRows: 8 }"
                                        placeholder="Enter a system prompt to be used for all requests..."
                                        class="!bg-gray-50 dark:!bg-[#191919]"
                                    />
                                    <template #feedback>
                                        <span class="text-xs text-claude-secondaryText dark:text-gray-500">This prompt will be prepended to all chat requests.</span>
                                    </template>
                                </n-form-item>
                            </n-form>
                        </n-card>

                        <!-- Generation Parameters -->
                        <n-card class="!bg-white dark:!bg-[#212124] !border-claude-border dark:!border-[#2A2A2E] !text-claude-text dark:!text-gray-200 shadow-sm transition-colors duration-200" title="Generation Parameters">
                            <n-form label-placement="top" class="mt-2">
                                <!-- Temperature -->
                                <div class="mb-6">
                                    <div class="text-sm font-medium text-claude-text dark:text-gray-200 mb-2">Temperature</div>
                                    <div class="flex gap-4 items-center">
                                        <n-slider 
                                            :value="modelSettings.temperature ?? 1" 
                                            @update:value="v => modelSettings.temperature = v"
                                            :min="0" 
                                            :max="2" 
                                            :step="0.1"
                                            :tooltip="false"
                                            class="flex-1"
                                        />
                                        <n-input-number 
                                            v-model:value="modelSettings.temperature" 
                                            :min="0" 
                                            :max="2" 
                                            :step="0.1"
                                            :precision="1"
                                            size="small"
                                            class="!w-24"
                                        />
                                    </div>
                                    <div class="text-xs text-claude-secondaryText dark:text-gray-500 mt-1">Controls randomness. Lower = more deterministic, Higher = more creative.</div>
                                </div>

                                <n-divider class="!my-4 !bg-claude-border dark:!bg-[#2A2A2E]" />

                                <div class="grid grid-cols-2 gap-6">
                                    <!-- Top-P -->
                                    <div>
                                        <div class="text-sm font-medium text-claude-text dark:text-gray-200 mb-2">Top-P</div>
                                        <n-input-number 
                                            v-model:value="modelSettings.top_p" 
                                            :min="0" 
                                            :max="1" 
                                            :step="0.05"
                                            :precision="2"
                                            placeholder="0.95"
                                            class="!w-full"
                                        />
                                        <div class="text-xs text-claude-secondaryText dark:text-gray-500 mt-1">Nucleus sampling threshold</div>
                                    </div>

                                    <!-- Top-K -->
                                    <div>
                                        <div class="text-sm font-medium text-claude-text dark:text-gray-200 mb-2">Top-K</div>
                                        <n-input-number 
                                            v-model:value="modelSettings.top_k" 
                                            :min="1" 
                                            :max="100" 
                                            :step="1"
                                            placeholder="40"
                                            class="!w-full"
                                        />
                                        <div class="text-xs text-claude-secondaryText dark:text-gray-500 mt-1">Top-K sampling</div>
                                    </div>
                                </div>

                                <n-divider class="!my-4 !bg-claude-border dark:!bg-[#2A2A2E]" />

                                <!-- Max Output Tokens -->
                                <n-form-item label="Max Output Tokens">
                                    <n-input-number 
                                        v-model:value="modelSettings.max_output_tokens" 
                                        :min="1" 
                                        :max="65536" 
                                        :step="100"
                                        placeholder="8192"
                                        class="!w-full"
                                    />
                                    <template #feedback>
                                        <span class="text-xs text-claude-secondaryText dark:text-gray-500">Maximum number of tokens to generate.</span>
                                    </template>
                                </n-form-item>
                            </n-form>
                        </n-card>

                        <!-- Advanced Model Features -->
                        <n-card class="!bg-white dark:!bg-[#212124] !border-claude-border dark:!border-[#2A2A2E] !text-claude-text dark:!text-gray-200 shadow-sm transition-colors duration-200" title="Advanced Features (Gemini 2.5+)">
                            <n-form label-placement="top" class="mt-2">
                                <div class="grid grid-cols-2 gap-6">
                                    <!-- Thinking Level -->
                                    <n-form-item label="Thinking Level">
                                        <n-select 
                                            v-model:value="modelSettings.thinking_level" 
                                            :options="thinkingLevelOptions" 
                                            placeholder="Select thinking level"
                                        />
                                        <template #feedback>
                                            <span class="text-xs text-claude-secondaryText dark:text-gray-500">Controls reasoning depth for supported models.</span>
                                        </template>
                                    </n-form-item>

                                    <!-- Media Resolution -->
                                    <n-form-item label="Media Resolution">
                                        <n-select 
                                            v-model:value="modelSettings.media_resolution" 
                                            :options="mediaResolutionOptions" 
                                            placeholder="Select resolution"
                                        />
                                        <template #feedback>
                                            <span class="text-xs text-claude-secondaryText dark:text-gray-500">Image/video processing resolution.</span>
                                        </template>
                                    </n-form-item>
                                </div>
                            </n-form>
                        </n-card>
                    </template>

                    <div class="flex justify-end pt-4">
                        <n-button 
                            @click="handleSave" 
                            :loading="loading"
                            class="!bg-[#D97757] !text-white !border-none hover:!bg-[#E6886A] !h-10 !px-8 rounded"
                        >
                            <template #icon>
                                <Save class="w-4 h-4 mr-2" />
                            </template>
                            Save Changes
                        </n-button>
                    </div>

            </div>
        </div>
    </div>

    <!-- Delete Chats Confirmation Modal -->
    <n-modal v-model:show="showDeleteChatsModal" preset="dialog" type="warning" title="Delete All Chat History">
        <template #default>
            <p class="text-sm text-gray-600 dark:text-gray-400">
                This action will permanently delete all chat sessions and messages. This cannot be undone.
            </p>
        </template>
        <template #action>
            <div class="flex gap-2 justify-end">
                <n-button @click="showDeleteChatsModal = false" size="small">Cancel</n-button>
                <n-button @click="handleDeleteChats" :loading="deletingChats" type="error" size="small">Delete All</n-button>
            </div>
        </template>
    </n-modal>

    <!-- Reset Stats Confirmation Modal -->
    <n-modal v-model:show="showDeleteStatsModal" preset="dialog" type="warning" title="Reset All Statistics">
        <template #default>
            <p class="text-sm text-gray-600 dark:text-gray-400">
                This action will reset all API key usage statistics (request counts, token usage, etc.). This cannot be undone.
            </p>
        </template>
        <template #action>
            <div class="flex gap-2 justify-end">
                <n-button @click="showDeleteStatsModal = false" size="small">Cancel</n-button>
                <n-button @click="handleResetStats" :loading="deletingStats" type="error" size="small">Reset All</n-button>
            </div>
        </template>
    </n-modal>
</template>

<style scoped>
/* Select Component Styling with dark mode support */
:deep(.n-base-selection) {
    --n-border: 1px solid #E1DFDD !important;
    --n-border-active: 1px solid #D97757 !important;
    --n-border-focus: 1px solid #D97757 !important;
}

:deep(.n-base-selection-label) {
    background-color: transparent !important;
}

/* Light mode */
.bg-claude-bg :deep(.n-base-selection) {
    background-color: #FFFFFF !important;
}
.bg-claude-bg :deep(.n-base-selection-input__content) {
    color: #1F1E1D !important;
}

/* Dark mode */
.dark :deep(.n-base-selection) {
    --n-border: 1px solid #2A2A2E !important;
    background-color: #191919 !important;
}
.dark :deep(.n-base-selection-input__content) {
    color: #E5E7EB !important;
}
.dark :deep(.n-base-selection-placeholder__inner) {
    color: #6B7280 !important;
}

/* InputNumber dark mode */
.dark :deep(.n-input-number) {
    background-color: #191919 !important;
}
.dark :deep(.n-input-number .n-input__input-el) {
    color: #E5E7EB !important;
}
.dark :deep(.n-input-number .n-input-wrapper) {
    background-color: #191919 !important;
}
</style>
