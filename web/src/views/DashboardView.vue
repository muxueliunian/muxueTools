<script setup lang="ts">
/**
 * Dashboard View
 * Responsibility: Display proxy API usage guide and system status.
 * Shows API endpoint, quick start examples, and key statistics.
 */
import { ref, computed, onMounted } from 'vue'
import { NCard, NButton, NIcon, NAlert, NSpin, useMessage } from 'naive-ui'
import { Copy, CheckCircle, XCircle, Clock, Key, Activity, Zap } from 'lucide-vue-next'
import { useI18n } from 'vue-i18n'
import type { HealthStats } from '../api/types'
import { getConfig } from '../api/config'

const message = useMessage()
const { t } = useI18n()

// State
const health = ref<HealthStats | null>(null)
const loading = ref(true)
const error = ref<string | null>(null)

/** Dynamic Base URL computed from current origin */
const baseUrl = computed(() => `${window.location.origin}/v1`)

/** OpenAI-compatible API Key (from backend config) */
const apiKey = ref('sk-mxln-proxy-local')

/** Generate curl example with current base URL */
const curlExample = computed(() => `curl -X POST ${baseUrl.value}/chat/completions \\
  -H "Content-Type: application/json" \\
  -d '{
    "model": "gpt-4",
    "messages": [{"role": "user", "content": "Hello!"}]
  }'`)

/** Python example code */
const pythonExample = computed(() => `from openai import OpenAI

client = OpenAI(
    base_url="${baseUrl.value}",
    api_key="not-needed"  ${t('dashboard.pythonComment')}
)

response = client.chat.completions.create(
    model="gpt-4",
    messages=[{"role": "user", "content": "Hello!"}]
)
print(response.choices[0].message.content)`)

/**
 * Fetch health status from backend
 */
async function loadHealth() {
  loading.value = true
  error.value = null
  try {
    const res = await fetch('/health')
    if (!res.ok) throw new Error(`HTTP ${res.status}`)
    health.value = await res.json()
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Failed to load health status'
    console.error('Failed to load health:', e)
  } finally {
    loading.value = false
  }
}

/**
 * Copy text to clipboard with feedback
 */
async function copyToClipboard(text: string, label: string = 'Text') {
  try {
    await navigator.clipboard.writeText(text)
    message.success(`${label} copied to clipboard!`)
  } catch (e) {
    message.error('Failed to copy to clipboard')
  }
}

/** Format uptime from seconds to human readable */
function formatUptime(seconds: number): string {
  if (seconds < 60) return `${seconds}s`
  if (seconds < 3600) return `${Math.floor(seconds / 60)}m ${seconds % 60}s`
  const hours = Math.floor(seconds / 3600)
  const mins = Math.floor((seconds % 3600) / 60)
  return `${hours}h ${mins}m`
}

onMounted(async () => {
  // Load health status
  loadHealth()
  
  // Fetch actual proxy key from config
  try {
    const configRes = await getConfig()
    if (configRes.data?.security?.proxy_key) {
      apiKey.value = configRes.data.security.proxy_key
    }
  } catch (e) {
    console.error('Failed to load proxy key:', e)
  }
})
</script>

<template>
  <div class="min-h-screen bg-claude-bg dark:bg-claude-dark-bg text-claude-text dark:text-claude-dark-text p-8 font-sans transition-colors duration-200">
    <div class="max-w-5xl mx-auto space-y-8">
      
      <!-- Header -->
      <div class="mb-8">
        <h1 class="text-3xl font-light text-claude-text dark:text-white tracking-tight mb-2">{{ $t('dashboard.title') }}</h1>
        <p class="text-claude-secondaryText dark:text-gray-500 text-sm">{{ $t('dashboard.subtitle') }}</p>
      </div>

      <!-- Loading State -->
      <div v-if="loading" class="flex items-center justify-center py-16">
        <n-spin size="large" />
      </div>

      <!-- Error State -->
      <div v-else-if="error" class="bg-red-50 dark:bg-red-950/30 border border-red-200 dark:border-red-900 rounded-lg p-4 mb-6">
        <div class="flex items-center justify-between">
          <div class="text-red-700 dark:text-red-300">
            <strong>{{ $t('dashboard.connectionError') }}:</strong> {{ error }}
          </div>
          <n-button size="small" @click="loadHealth">{{ $t('common.retry') }}</n-button>
        </div>
      </div>

      <template v-else>
        <!-- API Endpoint Card -->
        <n-card 
          class="!bg-white dark:!bg-[#212124] !border-claude-border dark:!border-[#2A2A2E] shadow-sm"
          :header-style="{ borderBottom: 'none', paddingBottom: '0' }"
        >
          <template #header>
            <div class="flex items-center gap-2 text-claude-text dark:text-white">
              <n-icon :component="Zap" class="text-[#D97757]" />
              <span class="font-medium">{{ $t('dashboard.apiEndpoint') }}</span>
            </div>
          </template>
          
          <div class="space-y-4">
            <!-- Base URL -->
            <div class="flex items-center gap-3 bg-gray-50 dark:bg-[#191919] rounded-lg p-4">
              <div class="flex-1">
                <div class="text-xs text-claude-secondaryText dark:text-gray-500 mb-1">{{ $t('dashboard.baseUrl') }}</div>
                <code class="text-sm font-mono text-claude-text dark:text-white">{{ baseUrl }}</code>
              </div>
              <n-button 
                quaternary 
                circle 
                @click="copyToClipboard(baseUrl, 'Base URL')"
                class="!text-gray-400 hover:!text-[#D97757]"
              >
                <template #icon><n-icon :component="Copy" /></template>
              </n-button>
            </div>

            <!-- API Key -->
            <div class="flex items-center gap-3 bg-gray-50 dark:bg-[#191919] rounded-lg p-4">
              <div class="flex-1">
                <div class="text-xs text-claude-secondaryText dark:text-gray-500 mb-1">{{ $t('dashboard.apiKey') }}</div>
                <code class="text-sm font-mono text-claude-text dark:text-white">{{ apiKey }}</code>
              </div>
              <n-button 
                quaternary 
                circle 
                @click="copyToClipboard(apiKey, 'API Key')"
                class="!text-gray-400 hover:!text-[#D97757]"
              >
                <template #icon><n-icon :component="Copy" /></template>
              </n-button>
            </div>

            <!-- Status -->
            <div class="flex items-center gap-6">
              <div class="flex items-center gap-2">
                <div :class="[
                  'w-2 h-2 rounded-full',
                  health?.status === 'ok' 
                    ? 'bg-emerald-500 shadow-[0_0_8px_rgba(16,185,129,0.4)]' 
                    : 'bg-yellow-500'
                ]" />
                <span class="text-sm text-claude-secondaryText dark:text-gray-400">
                  {{ health?.status === 'ok' ? $t('dashboard.running') : $t('dashboard.degraded') }}
                </span>
              </div>
              <div class="text-sm text-claude-secondaryText dark:text-gray-500">
                {{ health?.keys.active }} / {{ health?.keys.total }} {{ $t('common.active') }}
              </div>
              <div class="text-sm text-claude-secondaryText dark:text-gray-500">
                {{ $t('dashboard.uptime') }}: {{ health?.uptime ? formatUptime(health.uptime) : '-' }}
              </div>
            </div>
          </div>
        </n-card>

        <!-- Quick Start Card -->
        <n-card 
          class="!bg-white dark:!bg-[#212124] !border-claude-border dark:!border-[#2A2A2E] shadow-sm"
          :header-style="{ borderBottom: 'none', paddingBottom: '0' }"
        >
          <template #header>
            <div class="flex items-center gap-2 text-claude-text dark:text-white">
              <n-icon :component="Activity" class="text-[#D97757]" />
              <span class="font-medium">{{ $t('dashboard.quickStart') }}</span>
            </div>
          </template>

          <div class="space-y-4">
            <!-- curl Example -->
            <div>
              <div class="flex items-center justify-between mb-2">
                <span class="text-xs font-medium text-claude-secondaryText dark:text-gray-500">cURL</span>
                <n-button 
                  text 
                  size="tiny" 
                  @click="copyToClipboard(curlExample, 'cURL command')"
                  class="!text-gray-400 hover:!text-[#D97757]"
                >
                  <template #icon><n-icon :component="Copy" size="14" /></template>
                  Copy
                </n-button>
              </div>
              <div class="bg-[#1e1e1e] rounded-lg p-4 overflow-x-auto">
                <pre class="text-sm font-mono text-gray-300 whitespace-pre">{{ curlExample }}</pre>
              </div>
            </div>

            <!-- Python Example -->
            <div>
              <div class="flex items-center justify-between mb-2">
                <span class="text-xs font-medium text-claude-secondaryText dark:text-gray-500">Python (OpenAI SDK)</span>
                <n-button 
                  text 
                  size="tiny" 
                  @click="copyToClipboard(pythonExample, 'Python code')"
                  class="!text-gray-400 hover:!text-[#D97757]"
                >
                  <template #icon><n-icon :component="Copy" size="14" /></template>
                  Copy
                </n-button>
              </div>
              <div class="bg-[#1e1e1e] rounded-lg p-4 overflow-x-auto">
                <pre class="text-sm font-mono text-gray-300 whitespace-pre">{{ pythonExample }}</pre>
              </div>
            </div>

            <!-- Tip -->
            <n-alert type="info" :show-icon="false" class="!bg-blue-50 dark:!bg-blue-950/30 !border-blue-200 dark:!border-blue-900">
              <div class="text-sm text-blue-700 dark:text-blue-300">
                {{ $t('dashboard.tip') }} {{ $t('dashboard.noApiKeyNeeded') }}
              </div>
            </n-alert>
          </div>
        </n-card>

        <!-- Stats Cards Row -->
        <div class="grid grid-cols-1 md:grid-cols-4 gap-4">
          <!-- Total Keys -->
          <div class="bg-white dark:bg-[#212124] border border-claude-border dark:border-[#2A2A2E] rounded-xl p-5 transition-colors">
            <div class="flex items-center gap-3 mb-3">
              <div class="w-10 h-10 rounded-lg bg-[#D97757]/10 flex items-center justify-center">
                <n-icon :component="Key" class="text-[#D97757]" size="20" />
              </div>
              <span class="text-sm text-claude-secondaryText dark:text-gray-500">{{ $t('dashboard.totalKeys') }}</span>
            </div>
            <div class="text-3xl font-light text-claude-text dark:text-white">
              {{ health?.keys.total ?? '-' }}
            </div>
          </div>

          <!-- Active Keys -->
          <div class="bg-white dark:bg-[#212124] border border-claude-border dark:border-[#2A2A2E] rounded-xl p-5 transition-colors">
            <div class="flex items-center gap-3 mb-3">
              <div class="w-10 h-10 rounded-lg bg-emerald-500/10 flex items-center justify-center">
                <n-icon :component="CheckCircle" class="text-emerald-500" size="20" />
              </div>
              <span class="text-sm text-claude-secondaryText dark:text-gray-500">{{ $t('dashboard.activeKeys') }}</span>
            </div>
            <div class="text-3xl font-light text-emerald-500">
              {{ health?.keys.active ?? '-' }}
            </div>
          </div>

          <!-- Rate Limited -->
          <div class="bg-white dark:bg-[#212124] border border-claude-border dark:border-[#2A2A2E] rounded-xl p-5 transition-colors">
            <div class="flex items-center gap-3 mb-3">
              <div class="w-10 h-10 rounded-lg bg-yellow-500/10 flex items-center justify-center">
                <n-icon :component="Clock" class="text-yellow-500" size="20" />
              </div>
              <span class="text-sm text-claude-secondaryText dark:text-gray-500">{{ $t('dashboard.rateLimited') }}</span>
            </div>
            <div class="text-3xl font-light text-yellow-500">
              {{ health?.keys.rate_limited ?? '-' }}
            </div>
          </div>

          <!-- Disabled -->
          <div class="bg-white dark:bg-[#212124] border border-claude-border dark:border-[#2A2A2E] rounded-xl p-5 transition-colors">
            <div class="flex items-center gap-3 mb-3">
              <div class="w-10 h-10 rounded-lg bg-red-500/10 flex items-center justify-center">
                <n-icon :component="XCircle" class="text-red-500" size="20" />
              </div>
              <span class="text-sm text-claude-secondaryText dark:text-gray-500">{{ $t('dashboard.disabledKeys') }}</span>
            </div>
            <div class="text-3xl font-light text-red-500">
              {{ health?.keys.disabled ?? '-' }}
            </div>
          </div>
        </div>

        <!-- Version Info -->
        <div class="text-center text-xs text-claude-secondaryText dark:text-gray-600">
          MuxueTools v{{ health?.version ?? '1.0.0' }}
        </div>
      </template>
    </div>
  </div>
</template>

<style scoped>
/* Custom scrollbar for code blocks */
pre::-webkit-scrollbar {
  height: 6px;
}
pre::-webkit-scrollbar-track {
  background: transparent;
}
pre::-webkit-scrollbar-thumb {
  background: #4a4a4a;
  border-radius: 3px;
}
</style>
