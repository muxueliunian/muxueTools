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
import { useI18n } from 'vue-i18n'

const globalStore = useGlobalStore()

const message = useMessage()
const { t } = useI18n()
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
    media_resolution: null,
    stream_output: true  // 默认启用流式输出
})

const thinkingLevelOptions = computed(() => [
    { label: t('settings.thinkingDisabled'), value: '' },
    { label: t('settings.thinkingLow'), value: 'LOW' },
    { label: t('settings.thinkingMedium'), value: 'MEDIUM' },
    { label: t('settings.thinkingHigh'), value: 'HIGH' }
])

const mediaResolutionOptions = computed(() => [
    { label: t('settings.mediaDefault'), value: '' },
    { label: t('settings.mediaLow'), value: 'MEDIA_RESOLUTION_LOW' },
    { label: t('settings.mediaMedium'), value: 'MEDIA_RESOLUTION_MEDIUM' },
    { label: t('settings.mediaHigh'), value: 'MEDIA_RESOLUTION_HIGH' }
])

const strategyOptions = computed(() => [
    { label: t('settings.roundRobin'), value: 'round_robin' },
    { label: t('settings.randomSelection'), value: 'random' },
    { label: t('settings.leastUsedFirst'), value: 'least_used' },
    { label: t('settings.weightedRandom'), value: 'weighted' }
])

const logLevelOptions = computed(() => [
    { label: t('settings.debugVerbose'), value: 'debug' },
    { label: t('settings.infoStandard'), value: 'info' },
    { label: t('settings.warning'), value: 'warn' },
    { label: t('settings.errorCritical'), value: 'error' }
])

const updateSourceOptions = computed(() => [
    { label: t('settings.mxlnServer'), value: 'mxln' },
    { label: t('settings.github'), value: 'github' }
])

// 更新源选择
const updateSource = ref<'mxln' | 'github'>('mxln')

// Server port configuration
const serverPort = ref<number>(8080)

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
                    media_resolution: res.data.model_settings.media_resolution ?? null,
                    stream_output: res.data.model_settings.stream_output ?? true
                }
            }
            // Sync server port (stored_port is the user configured value)
            if (res.data.server?.stored_port) {
                serverPort.value = res.data.server.stored_port
            } else if (res.data.server?.port) {
                serverPort.value = res.data.server.port
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
            media_resolution: modelSettings.value.media_resolution,
            stream_output: modelSettings.value.stream_output
        }
        
        // Server configuration (requires restart)
        configToSave.server = {
            port: serverPort.value
        }
        
        const res = await updateConfig(configToSave as Partial<ConfigInfo>)
        if (res.success) {
            message.success(t('settings.configSavedSuccess'))
            // Show restart hint if server port was included in save
            if (serverPort.value !== config.value.server.port) {
                message.warning(t('settings.serverPortDescription'), { duration: 5000 })
            }
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
            message.success(t('settings.deletedSessions', { count: res.data?.deleted || 0 }))
            showDeleteChatsModal.value = false
        }
    } catch (e) {
        message.error(t('settings.deleteChatDescription'))
    } finally {
        deletingChats.value = false
    }
}

async function handleResetStats() {
    deletingStats.value = true
    try {
        const res = await resetStats()
        if (res.success) {
            message.success(t('settings.resetKeysAffected', { count: res.data?.keys_affected || 0 }))
            showDeleteStatsModal.value = false
        }
    } catch (e) {
        message.error(t('settings.resetStatsDescription'))
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
                    <h1 class="text-3xl font-light text-claude-text dark:text-white tracking-tight mb-1">{{ $t('settings.title') }}</h1>
                    <p class="text-claude-secondaryText dark:text-gray-500 text-sm">{{ $t('settings.subtitle') }}</p>
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
                    >{{ $t('settings.general') }}</button>
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
                        {{ $t('settings.security') }}
                    </button>
                    <button 
                        @click="activeTab = 'advanced'"
                        :class="[
                            'px-4 py-2 rounded-md text-sm font-medium transition-all duration-200',
                            activeTab === 'advanced' 
                                ? 'bg-white dark:bg-[#2A2A2E] text-claude-text dark:text-white shadow-sm' 
                                : 'text-claude-secondaryText dark:text-gray-500 hover:text-claude-text dark:hover:text-gray-300'
                        ]"
                    >{{ $t('settings.advanced') }}</button>
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
                        {{ $t('settings.model') }}
                    </button>
                </div>
            </div>

            <!-- Main Content -->
            <div class="space-y-6">
                <!-- General Tab -->
                <template v-if="activeTab === 'general'">
                    <!-- Strategy Section -->
                    <n-card class="!bg-white dark:!bg-[#212124] !border-claude-border dark:!border-[#2A2A2E] !text-claude-text dark:!text-gray-200 shadow-sm transition-colors duration-200" :title="$t('settings.keyManagement')">
                        <n-form label-placement="top" class="mt-2">
                             <n-form-item :label="$t('settings.selectionStrategy')">
                                <n-select v-model:value="config.pool.strategy" :options="strategyOptions" />
                                <template #feedback>
                                    <span class="text-xs text-claude-secondaryText dark:text-gray-500">{{ $t('settings.strategyDescription') }}</span>
                                </template>
                            </n-form-item>
                        </n-form>
                        </n-card>

                         <!-- Logging Section -->
                        <n-card class="!bg-white dark:!bg-[#212124] !border-claude-border dark:!border-[#2A2A2E] !text-claude-text dark:!text-gray-200 shadow-sm transition-colors duration-200" :title="$t('settings.loggingAndUpdates')">
                             <n-form label-placement="top" class="mt-2">
                                <n-form-item :label="$t('settings.logLevel')">
                                    <n-select v-model:value="config.logging.level" :options="logLevelOptions" class="anthropic-select" />
                                </n-form-item>
                                <n-divider class="!my-4 !bg-claude-border dark:!bg-[#2A2A2E]" />
                                <div class="flex items-center justify-between">
                                    <div>
                                        <div class="text-sm font-medium text-claude-text dark:text-gray-200">{{ $t('settings.automaticUpdates') }}</div>
                                        <div class="text-xs text-claude-secondaryText dark:text-gray-500">{{ $t('settings.checkOnStartup') }}</div>
                                    </div>
                                    <div class="flex gap-4 items-center">
                                        <n-button size="small" tertiary class="!text-gray-400 hover:!text-white" @click="handleCheckUpdate" :loading="checkingUpdate">{{ $t('settings.checkNow') }}</n-button>
                                        <n-switch v-model:value="config.update.enabled" :rail-style="({ checked }) => ({ backgroundColor: checked ? '#D97757' : '#4B5563' })" />
                                    </div>
                                </div>
                                
                                <!-- Update Source Selection -->
                                <div class="mt-4 pt-4 border-t border-claude-border dark:border-[#2A2A2E]">
                                    <div class="text-sm font-medium text-claude-text dark:text-gray-200 mb-2">{{ $t('settings.updateSource') }}</div>
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
                                        {{ updateSource === 'mxln' ? $t('settings.mxlnDescription') : $t('settings.githubDescription') }}
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
                        <n-card class="!bg-white dark:!bg-[#212124] !border-claude-border dark:!border-[#2A2A2E] !text-claude-text dark:!text-gray-200 shadow-sm transition-colors duration-200" :title="$t('settings.accessControl')">
                            <n-form label-placement="top" class="mt-2">
                                <!-- IP Whitelist -->
                                <div class="flex items-center justify-between mb-4">
                                    <div>
                                        <div class="text-sm font-medium text-claude-text dark:text-gray-200">{{ $t('settings.ipWhitelist') }}</div>
                                        <div class="text-xs text-claude-secondaryText dark:text-gray-500">{{ $t('settings.ipWhitelistDescription') }}</div>
                                    </div>
                                    <n-switch v-model:value="ipWhitelistEnabled" :rail-style="({ checked }) => ({ backgroundColor: checked ? '#D97757' : '#4B5563' })" />
                                </div>

                                <n-form-item :label="$t('settings.allowedIpAddress')" v-if="ipWhitelistEnabled">
                                    <n-input 
                                        v-model:value="whitelistIP" 
                                        placeholder="e.g., 192.168.1.100"
                                        class="!bg-gray-50 dark:!bg-[#191919]"
                                    />
                                    <template #feedback>
                                        <span class="text-xs text-claude-secondaryText dark:text-gray-500">{{ $t('settings.localhostAlwaysAllowed') }}</span>
                                    </template>
                                </n-form-item>

                                <n-divider class="!my-4 !bg-claude-border dark:!bg-[#2A2A2E]" />

                                <!-- Proxy API Key -->
                                <n-form-item :label="$t('settings.proxyApiKey')">
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
                                            {{ $t('settings.regenerate') }}
                                        </n-button>
                                    </div>
                                    <template #feedback>
                                        <span class="text-xs text-claude-secondaryText dark:text-gray-500">{{ $t('settings.proxyKeyDescription') }}</span>
                                    </template>
                                </n-form-item>

                                <n-divider class="!my-4 !bg-claude-border dark:!bg-[#2A2A2E]" />

                                <!-- Server Port Configuration -->
                                <n-form-item :label="$t('settings.serverPort')">
                                <n-input-number 
                                        v-model:value="serverPort" 
                                        :min="1024" 
                                        :max="65535"
                                        placeholder="8080"
                                        class="w-48"
                                    />
                                    <template #feedback>
                                        <span class="text-xs text-claude-secondaryText dark:text-gray-500">{{ $t('settings.serverPortDescription') }}</span>
                                    </template>
                                </n-form-item>
                            </n-form>
                        </n-card>
                    </template>

                    <!-- Advanced Tab -->
                    <template v-if="activeTab === 'advanced'">
                        <!-- Configuration Parameters -->
                        <n-card class="!bg-white dark:!bg-[#212124] !border-claude-border dark:!border-[#2A2A2E] !text-claude-text dark:!text-gray-200 shadow-sm transition-colors duration-200" :title="$t('settings.performanceTuning')">
                            <n-form label-placement="top" class="mt-2">
                                <div class="grid grid-cols-3 gap-4">
                                    <div class="space-y-1">
                                        <div class="text-sm font-medium text-claude-text dark:text-gray-200">{{ $t('settings.cooldownTime') }}</div>
                                        <n-input-number 
                                            v-model:value="config.pool.cooldown_seconds" 
                                            :min="60" 
                                            :max="86400"
                                            class="!w-full"
                                        >
                                            <template #suffix>
                                                <span class="text-xs text-gray-400">{{ $t('settings.sec') }}</span>
                                            </template>
                                        </n-input-number>
                                        <div class="text-xs text-claude-secondaryText dark:text-gray-500">{{ $t('settings.cooldownDescription') }}</div>
                                    </div>
                                    <div class="space-y-1">
                                        <div class="text-sm font-medium text-claude-text dark:text-gray-200">{{ $t('settings.maxRetries') }}</div>
                                        <n-input-number 
                                            v-model:value="config.pool.max_retries" 
                                            :min="0" 
                                            :max="10"
                                            class="!w-full"
                                        />
                                        <div class="text-xs text-claude-secondaryText dark:text-gray-500">{{ $t('settings.retryOnFailure') }}</div>
                                    </div>
                                    <div class="space-y-1">
                                        <div class="text-sm font-medium text-claude-text dark:text-gray-200">{{ $t('settings.requestTimeout') }}</div>
                                        <n-input-number 
                                            v-model:value="config.advanced!.request_timeout" 
                                            :min="30" 
                                            :max="600"
                                            class="!w-full"
                                        >
                                            <template #suffix>
                                                <span class="text-xs text-gray-400">{{ $t('settings.sec') }}</span>
                                            </template>
                                        </n-input-number>
                                        <div class="text-xs text-claude-secondaryText dark:text-gray-500">{{ $t('settings.apiRequestTimeout') }}</div>
                                    </div>
                                </div>

                                <n-divider class="!my-4 !bg-claude-border dark:!bg-[#2A2A2E]" />

                                <div class="flex items-center justify-between">
                                    <div>
                                        <div class="text-sm font-medium text-claude-text dark:text-gray-200">{{ $t('settings.debugMode') }}</div>
                                        <div class="text-xs text-claude-secondaryText dark:text-gray-500">{{ $t('settings.verboseLogging') }}</div>
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
                        <n-card class="!bg-white dark:!bg-[#212124] !border-claude-border dark:!border-[#2A2A2E] !text-claude-text dark:!text-gray-200 shadow-sm transition-colors duration-200" :title="$t('settings.dataManagement')">
                            <n-form label-placement="top" class="mt-2">
                                <n-form-item :label="$t('settings.databaseLocation')">
                                    <n-input 
                                        value="./data/muxueTools.db" 
                                        readonly
                                        class="!bg-gray-50 dark:!bg-[#191919]"
                                    />
                                    <template #feedback>
                                        <span class="text-xs text-claude-secondaryText dark:text-gray-500">{{ $t('settings.sqlitePath') }}</span>
                                    </template>
                                </n-form-item>

                                <n-divider class="!my-4 !bg-claude-border dark:!bg-[#2A2A2E]" />

                                <div class="text-sm font-medium text-claude-text dark:text-gray-200 mb-3">{{ $t('settings.dangerZone') }}</div>
                                <div class="space-y-3">
                                    <div class="flex items-center justify-between p-3 border border-red-500/20 rounded bg-red-500/5">
                                        <div>
                                            <div class="text-sm text-red-400">{{ $t('settings.deleteChatHistory') }}</div>
                                            <div class="text-xs text-gray-500">{{ $t('settings.deleteChatDescription') }}</div>
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
                                            {{ $t('common.delete') }}
                                        </n-button>
                                    </div>
                                    <div class="flex items-center justify-between p-3 border border-red-500/20 rounded bg-red-500/5">
                                        <div>
                                            <div class="text-sm text-red-400">{{ $t('settings.resetStatistics') }}</div>
                                            <div class="text-xs text-gray-500">{{ $t('settings.resetStatsDescription') }}</div>
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
                                            {{ $t('settings.reset') }}
                                        </n-button>
                                    </div>
                                </div>
                            </n-form>
                        </n-card>
                    </template>

                    <!-- Model Tab -->
                    <template v-if="activeTab === 'model'">
                        <!-- System Prompt -->
                        <n-card class="!bg-white dark:!bg-[#212124] !border-claude-border dark:!border-[#2A2A2E] !text-claude-text dark:!text-gray-200 shadow-sm transition-colors duration-200" :title="$t('settings.systemPrompt')">
                            <n-form label-placement="top" class="mt-2">
                                <n-form-item :label="$t('settings.defaultSystemPrompt')">
                                    <n-input 
                                        v-model:value="modelSettings.system_prompt" 
                                        type="textarea"
                                        :autosize="{ minRows: 3, maxRows: 8 }"
                                        :placeholder="$t('settings.systemPromptPlaceholder')"
                                        class="!bg-gray-50 dark:!bg-[#191919]"
                                    />
                                    <template #feedback>
                                        <span class="text-xs text-claude-secondaryText dark:text-gray-500">{{ $t('settings.systemPromptDescription') }}</span>
                                    </template>
                                </n-form-item>
                            </n-form>
                        </n-card>

                        <!-- Generation Parameters -->
                        <n-card class="!bg-white dark:!bg-[#212124] !border-claude-border dark:!border-[#2A2A2E] !text-claude-text dark:!text-gray-200 shadow-sm transition-colors duration-200" :title="$t('settings.generationParameters')">
                            <n-form label-placement="top" class="mt-2">
                                <!-- Temperature -->
                                <div class="mb-6">
                                    <div class="text-sm font-medium text-claude-text dark:text-gray-200 mb-2">{{ $t('settings.temperature') }}</div>
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
                                    <div class="text-xs text-claude-secondaryText dark:text-gray-500 mt-1">{{ $t('settings.temperatureDescription') }}</div>
                                </div>

                                <n-divider class="!my-4 !bg-claude-border dark:!bg-[#2A2A2E]" />

                                <div class="grid grid-cols-2 gap-6">
                                    <!-- Top-P -->
                                    <div>
                                        <div class="text-sm font-medium text-claude-text dark:text-gray-200 mb-2">{{ $t('settings.topP') }}</div>
                                        <n-input-number 
                                            v-model:value="modelSettings.top_p" 
                                            :min="0" 
                                            :max="1" 
                                            :step="0.05"
                                            :precision="2"
                                            placeholder="0.95"
                                            class="!w-full"
                                        />
                                        <div class="text-xs text-claude-secondaryText dark:text-gray-500 mt-1">{{ $t('settings.topPDescription') }}</div>
                                    </div>

                                    <!-- Top-K -->
                                    <div>
                                        <div class="text-sm font-medium text-claude-text dark:text-gray-200 mb-2">{{ $t('settings.topK') }}</div>
                                        <n-input-number 
                                            v-model:value="modelSettings.top_k" 
                                            :min="1" 
                                            :max="100" 
                                            :step="1"
                                            placeholder="40"
                                            class="!w-full"
                                        />
                                        <div class="text-xs text-claude-secondaryText dark:text-gray-500 mt-1">{{ $t('settings.topKDescription') }}</div>
                                    </div>
                                </div>

                                <n-divider class="!my-4 !bg-claude-border dark:!bg-[#2A2A2E]" />

                                <!-- Max Output Tokens -->
                                <n-form-item :label="$t('settings.maxOutputTokens')">
                                    <n-input-number 
                                        v-model:value="modelSettings.max_output_tokens" 
                                        :min="1" 
                                        :max="65536" 
                                        :step="100"
                                        placeholder="8192"
                                        class="!w-full"
                                    />
                                    <template #feedback>
                                        <span class="text-xs text-claude-secondaryText dark:text-gray-500">{{ $t('settings.maxOutputTokensDescription') }}</span>
                                    </template>
                                </n-form-item>
                            </n-form>
                        </n-card>

                        <!-- Advanced Model Features -->
                        <n-card class="!bg-white dark:!bg-[#212124] !border-claude-border dark:!border-[#2A2A2E] !text-claude-text dark:!text-gray-200 shadow-sm transition-colors duration-200" :title="$t('settings.advancedFeatures')">
                            <n-form label-placement="top" class="mt-2">
                                <div class="grid grid-cols-2 gap-6">
                                    <!-- Thinking Level -->
                                    <n-form-item :label="$t('settings.thinkingLevel')">
                                        <n-select 
                                            v-model:value="modelSettings.thinking_level" 
                                            :options="thinkingLevelOptions" 
                                            :placeholder="$t('settings.selectThinkingLevel')"
                                        />
                                        <template #feedback>
                                            <span class="text-xs text-claude-secondaryText dark:text-gray-500">{{ $t('settings.thinkingLevelDescription') }}</span>
                                        </template>
                                    </n-form-item>

                                    <!-- Media Resolution -->
                                    <n-form-item :label="$t('settings.mediaResolution')">
                                        <n-select 
                                            v-model:value="modelSettings.media_resolution" 
                                            :options="mediaResolutionOptions" 
                                            :placeholder="$t('settings.selectResolution')"
                                        />
                                        <template #feedback>
                                            <span class="text-xs text-claude-secondaryText dark:text-gray-500">{{ $t('settings.mediaResolutionDescription') }}</span>
                                        </template>
                                    </n-form-item>
                                </div>

                                <n-divider class="!my-4 !bg-claude-border dark:!bg-[#2A2A2E]" />

                                <!-- Stream Output Toggle -->
                                <div class="flex items-center justify-between">
                                    <div>
                                        <div class="text-sm font-medium text-claude-text dark:text-gray-200">{{ $t('settings.streamOutput') }}</div>
                                        <div class="text-xs text-claude-secondaryText dark:text-gray-500">{{ $t('settings.streamOutputDescription') }}</div>
                                    </div>
                                    <n-switch 
                                        v-model:value="modelSettings.stream_output" 
                                        :rail-style="({ checked }) => ({ backgroundColor: checked ? '#D97757' : '#4B5563' })" 
                                    />
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
                            {{ $t('settings.saveChanges') }}
                        </n-button>
                    </div>

            </div>
        </div>
    </div>

    <!-- Delete Chats Confirmation Modal -->
    <n-modal v-model:show="showDeleteChatsModal" preset="dialog" type="warning" :title="$t('settings.deleteAllChatHistory')">
        <template #default>
            <p class="text-sm text-gray-600 dark:text-gray-400">
                {{ $t('settings.deleteChatsWarning') }}
            </p>
        </template>
        <template #action>
            <div class="flex gap-2 justify-end">
                <n-button @click="showDeleteChatsModal = false" size="small">{{ $t('common.cancel') }}</n-button>
                <n-button @click="handleDeleteChats" :loading="deletingChats" type="error" size="small">{{ $t('settings.deleteAll') }}</n-button>
            </div>
        </template>
    </n-modal>

    <!-- Reset Stats Confirmation Modal -->
    <n-modal v-model:show="showDeleteStatsModal" preset="dialog" type="warning" :title="$t('settings.resetAllStatistics')">
        <template #default>
            <p class="text-sm text-gray-600 dark:text-gray-400">
                {{ $t('settings.resetStatsWarning') }}
            </p>
        </template>
        <template #action>
            <div class="flex gap-2 justify-end">
                <n-button @click="showDeleteStatsModal = false" size="small">{{ $t('common.cancel') }}</n-button>
                <n-button @click="handleResetStats" :loading="deletingStats" type="error" size="small">{{ $t('settings.resetAll') }}</n-button>
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
