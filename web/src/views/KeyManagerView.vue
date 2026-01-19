<script setup lang="ts">
/**
 * Key Management View
 * Responsibility: List, create, revoke, and test API keys.
 * Features: Multi-step wizard for adding keys with validation.
 * Dependencies: KeyStore
 */
import { h, ref, computed, onMounted, watch } from 'vue'
import { NDataTable, NButton, NTag, NInput, NSpace, NModal, NFormItem, NCard, useMessage, useDialog, NIcon, NTooltip, NSelect, NSteps, NStep, NAlert } from 'naive-ui'
import { useKeyStore } from '../stores/keyStore'
import { useGlobalStore } from '../stores/global'
import { Search, Plus, Trash2, Copy, Play, ArrowLeft, ArrowRight, Check, Upload } from 'lucide-vue-next'
import type { KeyInfo, KeyImportItem } from '../api/types'
import { format } from 'date-fns'
import { validateKey, type ValidateKeyResult } from '../api/keys'

const store = useKeyStore()
const globalStore = useGlobalStore()
const message = useMessage()
const dialog = useDialog()

// State
const showAddModal = ref(false)
const showImportModal = ref(false)
const searchText = ref('')
const importText = ref('')
const importing = ref(false)

const filteredKeys = computed(() => {
  if (!searchText.value) return store.keys
  const search = searchText.value.toLowerCase()
  return store.keys.filter(key => 
    key.name?.toLowerCase().includes(search) ||
    key.key?.toLowerCase().includes(search) ||
    key.tags?.some(tag => tag.toLowerCase().includes(search))
  )
})

// Wizard State
const currentStep = ref(1)
const validating = ref(false)
const creating = ref(false)
const validateResult = ref<ValidateKeyResult | null>(null)

const newKeyForm = ref({
  key: '',
  name: '',
  tags: '',
  provider: 'google_aistudio',
  defaultModel: ''
})

/** Provider options for selection */
const providerOptions = [
  { label: 'Google AI Studio', value: 'google_aistudio' },
  { label: 'Gemini API', value: 'gemini_api' }
]

/** Model options computed from validation result */
const modelOptions = computed(() => 
  validateResult.value?.models?.map(m => ({ label: m, value: m })) || []
)

/** Reset wizard state when modal closes */
watch(showAddModal, (show) => {
  if (!show) {
    resetWizard()
  }
})

/**
 * Reset wizard to initial state
 */
function resetWizard() {
  currentStep.value = 1
  validating.value = false
  creating.value = false
  validateResult.value = null
  newKeyForm.value = {
    key: '',
    name: '',
    tags: '',
    provider: 'google_aistudio',
    defaultModel: ''
  }
}

/**
 * Move to previous step
 */
function prevStep() {
  if (currentStep.value > 1) {
    currentStep.value--
  }
}

/**
 * Move to next step
 */
function nextStep() {
  if (currentStep.value < 4) {
    currentStep.value++
  }
}

// Columns
const columns = [
  {
    title: 'STATUS',
    key: 'status',
    width: 100,
    render(row: KeyInfo) {
      return h(
        'div',
        { class: 'flex items-center gap-2' },
        [
          h('div', {
            class: [
              'w-2 h-2 rounded-full',
              row.enabled ? 'bg-emerald-500 shadow-[0_0_8px_rgba(16,185,129,0.4)]' : 'bg-red-500'
            ]
          }),
          h('span', { class: 'text-xs text-claude-secondaryText dark:text-gray-400 font-medium' }, row.enabled ? 'Active' : 'Disabled')
        ]
      )
    }
  },
  {
    title: 'NAME',
    key: 'name',
    render(row: KeyInfo) {
      return h('span', { class: 'font-medium text-claude-text dark:text-gray-200' }, row.name || 'Untitled Key')
    }
  },
  {
    title: 'KEY',
    key: 'key',
    render(row: KeyInfo) {
      return h(
        'div',
        { class: 'font-mono text-xs text-claude-secondaryText dark:text-gray-400 select-all bg-gray-100 dark:bg-[#2A2A2E] px-2 py-1 rounded inline-block' },
        row.key
      )
    }
  },
  {
    title: 'TAGS',
    key: 'tags',
    render(row: KeyInfo) {
      if (!row.tags || row.tags.length === 0) return '-'
      return h(
        NSpace,
        { size: 4 },
        () => row.tags.map(tag => h(
          NTag,
          { size: 'small', bordered: false, class: '!bg-gray-100 dark:!bg-[#2A2A2E] !text-claude-secondaryText dark:!text-gray-400 !text-xs' },
          () => tag
        ))
      )
    }
  },
  {
    title: 'USAGE (24H)',
    key: 'usage',
    render(row: KeyInfo) {
      return h('div', { class: 'flex flex-col' }, [
        h('span', { class: 'text-xs text-gray-300' }, `${row.stats?.request_count || 0} reqs`),
        h('span', { class: 'text-[10px] text-gray-500' }, row.stats?.last_used_at ? format(new Date(row.stats.last_used_at), 'MM/dd HH:mm') : 'Never used')
      ])
    }
  },
  {
    title: 'ACTIONS',
    key: 'actions',
    width: 150,
    render(row: KeyInfo) {
      return h(NSpace, { size: 8 }, () => [
        h(NButton, {
          size: 'tiny',
          quaternary: true,
          circle: true,
          class: 'text-gray-400 hover:text-white',
          onClick: () => handleTestKey(row.id)
        }, { icon: () => h(NIcon, null, { default: () => h(Play, { size: 14 }) }) }),
        h(NButton, {
          size: 'tiny',
          quaternary: true,
          circle: true,
          class: 'text-gray-400 hover:text-white',
          onClick: () => handleCopy(row.key)
        }, { icon: () => h(NIcon, null, { default: () => h(Copy, { size: 14 }) }) }),
        h(NButton, {
          size: 'tiny',
          quaternary: true,
          circle: true,
          class: 'text-gray-400 hover:text-red-400',
          onClick: () => handleDelete(row.id)
        }, { icon: () => h(NIcon, null, { default: () => h(Trash2, { size: 14 }) }) })
      ])
    }
  }
]

// Handlers
const handleCopy = async (text: string) => {
  await navigator.clipboard.writeText(text)
  message.success('Copied to clipboard')
}

const handleTestKey = async (id: string) => {
  message.loading('Testing key connection...')
  const result = await store.testKeyConnection(id)
  if (result && result.valid) {
    message.success(`Connection successful (${result.latency_ms}ms)`)
  } else {
    message.error('Connection failed or key invalid')
  }
}

const handleDelete = async (id: string) => {
  dialog.warning({
    title: 'Revoke Key',
    content: 'Are you sure you want to revoke this API key? This action cannot be undone.',
    positiveText: 'Revoke',
    negativeText: 'Cancel',
    onPositiveClick: async () => {
      const success = await store.removeKey(id)
      if (success) message.success('Key revoked successfully')
      else message.error('Failed to revoke key')
    }
  })
}

/**
 * Validate the API key and fetch available models
 */
async function handleValidate() {
  if (!newKeyForm.value.key) {
    message.warning('Please enter an API Key')
    return
  }
  
  validating.value = true
  try {
    const res = await validateKey({ 
      key: newKeyForm.value.key, 
      provider: newKeyForm.value.provider 
    })
    
    if (res.success && res.data) {
      validateResult.value = res.data
      if (res.data.valid) {
        message.success(`Found ${res.data.models?.length || 0} models (${res.data.latency_ms}ms)`)
        // Auto-select first model as default
        if (res.data.models && res.data.models.length > 0) {
          newKeyForm.value.defaultModel = res.data.models[0] ?? ''
        }
        currentStep.value = 2
      } else {
        message.error(res.data.error || 'Invalid API Key')
      }
    } else {
      message.error('Validation request failed')
    }
  } catch (e) {
    message.error('Validation failed: Network error')
  } finally {
    validating.value = false
  }
}

/**
 * Create the API key with all collected information
 */
async function handleAddKey() {
  creating.value = true
  try {
    const success = await store.createKey({
      key: newKeyForm.value.key,
      name: newKeyForm.value.name || undefined,
      provider: newKeyForm.value.provider,
      default_model: newKeyForm.value.defaultModel || undefined,
      tags: newKeyForm.value.tags ? newKeyForm.value.tags.split(',').map(t => t.trim()).filter(Boolean) : undefined
    })
    
    if (success) {
      message.success('Key created successfully')
      showAddModal.value = false
    } else {
      message.error('Failed to create key')
    }
  } finally {
    creating.value = false
  }
}

/**
 * Handle batch import of keys
 */
async function handleImport() {
  if (!importText.value.trim()) {
    message.warning('Please enter keys to import')
    return
  }

  importing.value = true
  try {
    let keysToImport: KeyImportItem[] = []
    const text = importText.value.trim()
    
    // Try parsing as JSON
    if (text.startsWith('[') || text.startsWith('{')) {
      try {
        const parsed = JSON.parse(text)
        if (Array.isArray(parsed)) {
          keysToImport = parsed
        } else if (parsed.keys && Array.isArray(parsed.keys)) {
          keysToImport = parsed.keys
        }
      } catch (e) {
        // Fallback to text parsing if JSON fails
      }
    }

    // Fallback: Parse as newline separated text
    if (keysToImport.length === 0) {
      keysToImport = text.split('\n')
        .map(line => line.trim())
        .filter(line => line.length > 0)
        .map(line => ({ key: line }))
    }

    if (keysToImport.length === 0) {
      message.error('No valid keys found to import')
      return
    }

    const result = await store.importBatchKeys(keysToImport)
    if (result) {
      message.success(`Imported ${result.imported} keys (${result.skipped} skipped)`)
      if (result.errors && result.errors.length > 0) {
        // Show partial errors if any (simplified)
        message.warning(`Some keys failed: ${result.errors.length} errors`)
      }
      showImportModal.value = false
      importText.value = ''
    } else {
      message.error('Import failed')
    }
  } finally {
    importing.value = false
  }
}

// Lifecycle
onMounted(() => {
  store.fetchKeys()
})
</script>

<template>
  <div class="min-h-screen bg-claude-bg dark:bg-claude-dark-bg text-claude-text dark:text-claude-dark-text p-8 font-sans transition-colors duration-200">
    <div class="max-w-7xl mx-auto space-y-8">
      
      <!-- Header -->
      <div class="flex items-center justify-between">
        <div>
          <h1 class="text-3xl font-light text-claude-text dark:text-white tracking-tight mb-2">API Keys</h1>
          <p class="text-claude-secondaryText dark:text-gray-500 text-sm">Manage authentication keys for your AI models.</p>
        </div>
        <div class="flex items-center gap-3">
          <n-input 
            v-model:value="searchText" 
            placeholder="Search keys..." 
            class="!bg-white dark:!bg-[#212124] !border-claude-border dark:!border-[#2A2A2E] !text-claude-text dark:!text-gray-300 w-64"
            round size="medium"
          >
            <template #prefix><n-icon :component="Search" class="text-gray-500" /></template>
          </n-input>
          <n-button 
            class="!bg-[#D97757] !text-white !border-none hover:!bg-[#E6886A] transition-colors"
            @click="showAddModal = true"
            icon-placement="left"
          >
            <template #icon><n-icon :component="Plus" /></template>
            Create Key
          </n-button>
          <n-tooltip trigger="hover">
            <template #trigger>
              <n-button 
                strong secondary circle 
                class="!bg-white dark:!bg-[#212124] !text-gray-500 dark:!text-gray-400 !border-claude-border dark:!border-[#2A2A2E] mr-2"
                @click="showImportModal = true"
              >
                <template #icon><n-icon :component="Upload" /></template>
              </n-button>
            </template>
            Import Keys
          </n-tooltip>
        </div>
      </div>

      <!-- Main Content -->
      <div class="bg-white dark:bg-[#212124] border border-claude-border dark:border-[#2A2A2E] rounded-xl overflow-hidden shadow-sm transition-colors duration-200">
        <n-data-table
          :columns="columns"
          :data="filteredKeys"
          :loading="store.loading"
          :bordered="false"
          :single-line="false"
          class="anthropic-table"
          :row-class-name="'anthropic-row'"
        />
      </div>
      
    </div>

    <!-- Add Key Modal (Wizard) -->
    <n-modal v-model:show="showAddModal" :mask-closable="false">
      <div :class="{ 'dark': globalStore.isDark }">
        <n-card
          class="!bg-white dark:!bg-[#212124] !text-claude-text dark:!text-gray-200 !border-claude-border dark:!border-[#2A2A2E] w-[680px] shadow-2xl"
          title="Add New API Key"
          :header-style="globalStore.isDark ? { color: 'white', borderBottom: '1px solid #2A2A2E' } : { color: '#1F1E1D', borderBottom: '1px solid #E1DFDD' }"
          size="huge"
          aria-modal="true"
        >
          <!-- Steps Indicator -->
          <n-steps :current="currentStep" size="small" class="mb-6">
            <n-step title="Provider & Key" />
            <n-step title="Select Model" />
            <n-step title="Details" />
            <n-step title="Confirm" />
          </n-steps>

          <!-- Step 1: Provider & Key Input -->
          <div v-show="currentStep === 1" class="space-y-4">
            <n-form-item label="Provider" label-placement="top">
              <n-select 
                v-model:value="newKeyForm.provider" 
                :options="providerOptions" 
                placeholder="Select Provider"
                class="!bg-gray-50 dark:!bg-[#191919]"
              />
            </n-form-item>
            <n-form-item label="API Key" label-placement="top">
              <n-input 
                v-model:value="newKeyForm.key" 
                type="password" 
                show-password-on="click"
                placeholder="Enter your API Key (e.g., AIzaSy...)" 
                class="!bg-gray-50 dark:!bg-[#191919] !border-gray-200 dark:!border-[#2A2A2E] !text-gray-900 dark:!text-white"
              />
            </n-form-item>
            <n-button 
              type="primary" 
              :loading="validating"
              :disabled="!newKeyForm.key"
              @click="handleValidate"
              class="!bg-[#D97757] !text-white !border-none hover:!bg-[#E6886A] w-full"
            >
              <template #icon v-if="!validating"><n-icon :component="Check" /></template>
              {{ validating ? 'Validating...' : 'Validate & Fetch Models' }}
            </n-button>
          </div>

          <!-- Step 2: Model Selection -->
          <div v-show="currentStep === 2" class="space-y-4">
            <n-alert type="success" class="mb-4">
              Key validated successfully! Latency: {{ validateResult?.latency_ms }}ms
            </n-alert>
            <n-form-item label="Default Model" label-placement="top">
              <n-select 
                v-model:value="newKeyForm.defaultModel" 
                :options="modelOptions" 
                placeholder="Select a default model"
                filterable
                class="!bg-gray-50 dark:!bg-[#191919]"
              />
            </n-form-item>
            <p class="text-xs text-gray-500 dark:text-gray-400">
              Found {{ modelOptions.length }} available models. You can skip model selection if preferred.
            </p>
          </div>

          <!-- Step 3: Key Details -->
          <div v-show="currentStep === 3" class="space-y-4">
            <n-form-item label="Key Name (Optional)" label-placement="top">
              <n-input 
                v-model:value="newKeyForm.name" 
                placeholder="e.g., Production Key, Dev Team Key"
                class="!bg-gray-50 dark:!bg-[#191919] !border-gray-200 dark:!border-[#2A2A2E] !text-gray-900 dark:!text-white"
              />
            </n-form-item>
            <n-form-item label="Tags (Optional)" label-placement="top">
              <n-input 
                v-model:value="newKeyForm.tags" 
                placeholder="production, high-priority (comma separated)"
                class="!bg-gray-50 dark:!bg-[#191919] !border-gray-200 dark:!border-[#2A2A2E] !text-gray-900 dark:!text-white"
              />
            </n-form-item>
          </div>

          <!-- Step 4: Confirmation -->
          <div v-show="currentStep === 4" class="space-y-4">
            <div class="bg-gray-50 dark:bg-[#191919] rounded-lg p-4 space-y-3">
              <div class="flex justify-between">
                <span class="text-gray-500 dark:text-gray-400">Provider</span>
                <span class="font-medium text-gray-900 dark:text-white">{{ providerOptions.find(p => p.value === newKeyForm.provider)?.label }}</span>
              </div>
              <div class="flex justify-between">
                <span class="text-gray-500 dark:text-gray-400">API Key</span>
                <span class="font-mono text-xs text-gray-900 dark:text-white">{{ newKeyForm.key.slice(0, 10) }}...{{ newKeyForm.key.slice(-4) }}</span>
              </div>
              <div class="flex justify-between">
                <span class="text-gray-500 dark:text-gray-400">Default Model</span>
                <span class="font-medium text-gray-900 dark:text-white">{{ newKeyForm.defaultModel || 'Not selected' }}</span>
              </div>
              <div class="flex justify-between">
                <span class="text-gray-500 dark:text-gray-400">Name</span>
                <span class="font-medium text-gray-900 dark:text-white">{{ newKeyForm.name || 'Untitled' }}</span>
              </div>
              <div class="flex justify-between">
                <span class="text-gray-500 dark:text-gray-400">Tags</span>
                <span class="font-medium text-gray-900 dark:text-white">{{ newKeyForm.tags || 'None' }}</span>
              </div>
            </div>
          </div>

          <template #footer>
            <div class="flex justify-between pt-4 border-t border-gray-200 dark:border-[#2A2A2E]">
              <div>
                <n-button 
                  v-if="currentStep > 1" 
                  @click="prevStep" 
                  class="!text-gray-500 dark:!text-gray-400"
                  text
                >
                  <template #icon><n-icon :component="ArrowLeft" /></template>
                  Back
                </n-button>
              </div>
              <div class="flex gap-3">
                <n-button @click="showAddModal = false" class="!text-gray-500 dark:!text-gray-400 hover:!text-gray-900 dark:hover:!text-white" text>Cancel</n-button>
                
                <!-- Next button (Steps 2-3) -->
                <n-button 
                  v-if="currentStep >= 2 && currentStep < 4" 
                  @click="nextStep"
                  class="!bg-[#D97757] !text-white !border-none hover:!bg-[#E6886A]"
                >
                  Next
                  <template #icon><n-icon :component="ArrowRight" /></template>
                </n-button>
                
                <!-- Create button (Step 4) -->
                <n-button 
                  v-if="currentStep === 4" 
                  @click="handleAddKey"
                  :loading="creating"
                  class="!bg-[#D97757] !text-white !border-none hover:!bg-[#E6886A]"
                >
                  <template #icon v-if="!creating"><n-icon :component="Check" /></template>
                  Create Key
                </n-button>
              </div>
            </div>
          </template>
        </n-card>
      </div>
    </n-modal>

    <!-- Import Keys Modal -->
    <n-modal v-model:show="showImportModal">
      <div :class="{ 'dark': globalStore.isDark }">
        <n-card
          class="!bg-white dark:!bg-[#212124] !text-claude-text dark:!text-gray-200 !border-claude-border dark:!border-[#2A2A2E] w-[600px] shadow-2xl"
          title="Import Keys"
          :header-style="globalStore.isDark ? { color: 'white', borderBottom: '1px solid #2A2A2E' } : { color: '#1F1E1D', borderBottom: '1px solid #E1DFDD' }"
          size="huge"
          aria-modal="true"
        >
          <div class="space-y-4">
            <n-alert type="info" :bordered="false" class="mb-4">
              Paste keys line by line, or provide a JSON array.
            </n-alert>
            <n-form-item label="Keys" label-placement="top">
              <n-input
                v-model:value="importText"
                type="textarea"
                placeholder="AIzaSyABC123...&#10;AIzaSyDEF456...&#10;&#10;OR JSON:&#10;[&#10;  { &quot;key&quot;: &quot;AIzaSyABC123...&quot; },&#10;  { &quot;key&quot;: &quot;AIzaSyDEF456...&quot;, &quot;name&quot;: &quot;Production Key&quot; },&#10;  { &quot;key&quot;: &quot;AIzaSyGHI789...&quot;, &quot;name&quot;: &quot;Dev Key&quot;, &quot;tags&quot;: [&quot;dev&quot;] }&#10;]"
                :rows="10"
                class="!bg-gray-50 dark:!bg-[#191919] !border-gray-200 dark:!border-[#2A2A2E] !text-gray-900 dark:!text-white font-mono text-xs"
              />
            </n-form-item>
          </div>

          <template #footer>
            <div class="flex justify-end gap-3 pt-4 border-t border-gray-200 dark:border-[#2A2A2E]">
              <n-button @click="showImportModal = false" class="!text-gray-500 dark:!text-gray-400 hover:!text-gray-900 dark:hover:!text-white" text>Cancel</n-button>
              <n-button 
                @click="handleImport"
                :loading="importing"
                class="!bg-[#D97757] !text-white !border-none hover:!bg-[#E6886A]"
              >
                <template #icon v-if="!importing"><n-icon :component="Upload" /></template>
                Import Keys
              </n-button>
            </div>
          </template>
        </n-card>
      </div>
    </n-modal>
  </div>
</template>

<style scoped>
/* Anthropic Global Overrides for this view */
:deep(.n-data-table) {
    background-color: transparent !important;
}

:deep(.n-data-table-th) {
    background-color: transparent !important;
    border-bottom: 1px solid #E1DFDD !important; /* claude-border */
    color: #6F6F78 !important; /* claude-secondaryText */
    font-size: 12px !important;
    font-weight: 600 !important;
    letter-spacing: 0.05em !important;
    text-transform: uppercase !important;
    padding: 12px 16px !important;
}

:deep(.dark .n-data-table-th),
:host-context(.dark) :deep(.n-data-table-th) {
    border-bottom: 1px solid #2A2A2E !important;
    color: #6B7280 !important;
}

:deep(.n-data-table-td) {
    background-color: transparent !important;
    border-bottom: 1px solid #E1DFDD !important;
    color: #1F1E1D !important; /* claude-text */
    padding: 12px 16px !important;
}

:deep(.dark .n-data-table-td),
:host-context(.dark) :deep(.n-data-table-td) {
    border-bottom: 1px solid #2A2A2E !important;
    color: #E5E7EB !important;
}

:deep(.n-data-table-tr:hover .n-data-table-td) {
    background-color: #F0EEEB !important; /* claude-sidebar/hover */
}

:deep(.dark .n-data-table-tr:hover .n-data-table-td),
:host-context(.dark) :deep(.n-data-table-tr:hover .n-data-table-td)  {
    background-color: #2A2A2E !important;
}

:deep(.n-data-table-empty) {
    background-color: transparent !important;
    color: #6B7280 !important;
}

/* Steps styling */
:deep(.n-steps) {
    --n-indicator-color-process: #D97757 !important;
    --n-splitor-color-process: #D97757 !important;
}
</style>
