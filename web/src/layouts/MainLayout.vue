<script setup lang="ts">
/**
 * MainLayout - 主布局组件
 * 
 * 职责: 提供侧边栏导航、会话列表和内容区布局
 */

import { h, type Component, onMounted, computed } from 'vue'
import { NIcon, NMenu, NDropdown } from 'naive-ui'
import { RouterLink, useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { 
    LayoutDashboard, 
    KeyRound, 
    BarChart3, 
    Settings, 
    MessageSquare,
    Plus,
    Sun,
    Moon,
    Languages
} from 'lucide-vue-next'
import { useGlobalStore } from '@/stores/global'
import { useSessionStore } from '@/stores/sessionStore'
import { useChatStore } from '@/stores/chatStore'
import { supportedLocales, setLocale, getLocale, type LocaleCode } from '@/i18n'
import SessionList from '@/components/chat/SessionList.vue'

const globalStore = useGlobalStore()
const sessionStore = useSessionStore()
const chatStore = useChatStore()

const route = useRoute()
const { t } = useI18n()

/**
 * 初始化应用
 */
onMounted(async () => {
    // 恢复模型偏好
    chatStore.restoreModelPreference()
    // 加载模型列表
    await chatStore.loadModels()
    // 加载流式输出设置
    await chatStore.loadStreamSetting()
    // 初始化会话
    await sessionStore.initialize()
})

/**
 * 创建新会话
 */
async function handleNewChat() {
    await sessionStore.createNewSession()
}

function renderIcon (icon: Component) {
  return () => h(NIcon, null, { default: () => h(icon, { size: 18, strokeWidth: 2 }) })
}

// 语言切换器选项
const languageOptions = supportedLocales.map(locale => ({
  label: locale.name,
  key: locale.code
}))

// 当前语言显示名称
const currentLanguageName = computed(() => {
  const current = getLocale()
  return supportedLocales.find(l => l.code === current)?.name || 'English'
})

// 处理语言切换
function handleLanguageSelect(key: string) {
  setLocale(key as LocaleCode)
}

const menuOptions = [
  {
    label: () => h(RouterLink, { to: '/' }, { default: () => t('sidebar.chat') }),
    key: 'chat',
    icon: renderIcon(MessageSquare)
  },
  {
    label: () => h(RouterLink, { to: '/dashboard' }, { default: () => t('sidebar.dashboard') }),
    key: 'dashboard',
    icon: renderIcon(LayoutDashboard)
  },
  {
    label: () => h(RouterLink, { to: '/keys' }, { default: () => t('sidebar.keys') }),
    key: 'keys',
    icon: renderIcon(KeyRound)
  },
  {
    label: () => h(RouterLink, { to: '/stats' }, { default: () => t('sidebar.stats') }),
    key: 'stats',
    icon: renderIcon(BarChart3)
  },
    {
    label: () => h(RouterLink, { to: '/settings' }, { default: () => t('sidebar.settings') }),
    key: 'settings',
    icon: renderIcon(Settings)
  }
]
</script>

<template>
  <div :class="{ 'dark': globalStore.isDark }" class="h-screen w-full bg-claude-bg dark:bg-claude-dark-bg text-claude-text dark:text-claude-dark-text font-sans flex overflow-hidden transition-colors duration-200">
    <!-- Sidebar -->
    <div class="w-60 flex-shrink-0 bg-claude-sidebar dark:bg-claude-dark-sidebar border-r border-claude-border dark:border-claude-dark-border flex flex-col transition-colors duration-200">
        <div class="p-4">
            
            <button 
                @click="handleNewChat"
                class="w-full flex items-center gap-2 bg-[#D97757] hover:bg-[#E6886A] text-white px-4 py-2 rounded transition-colors mb-4 text-sm font-medium"
            >
                <Plus class="w-4 h-4" />
                <span>{{ $t('sidebar.newChat') }}</span>
            </button>
        </div>
        
        <!-- 会话列表 -->
        <SessionList />
        
        <!-- 导航菜单 -->
        <div class="border-t border-claude-border dark:border-claude-dark-border px-3 py-2">
              <n-menu 
                :options="menuOptions" 
                :value="String(route.name)"
                class="anthropic-menu"
            />
        </div>
        
        <div class="p-4 border-t border-claude-border dark:border-claude-dark-border space-y-1">
            <!-- 语言切换器 -->
            <n-dropdown 
                :options="languageOptions" 
                @select="handleLanguageSelect"
                trigger="click"
                placement="top-start"
            >
                <button 
                    class="w-full flex items-center gap-3 px-2 py-2 rounded hover:bg-claude-hover dark:hover:bg-claude-dark-hover cursor-pointer transition-colors text-claude-secondaryText dark:text-claude-dark-secondaryText hover:text-claude-text dark:hover:text-white"
                >
                    <Languages class="w-5 h-5" />
                    <span class="text-sm font-medium">{{ currentLanguageName }}</span>
                </button>
            </n-dropdown>
            
            <!-- 主题切换 -->
            <button 
                @click="globalStore.toggleTheme"
                class="w-full flex items-center gap-3 px-2 py-2 rounded hover:bg-claude-hover dark:hover:bg-claude-dark-hover cursor-pointer transition-colors text-claude-secondaryText dark:text-claude-dark-secondaryText hover:text-claude-text dark:hover:text-white"
            >
                <Sun v-if="!globalStore.isDark" class="w-5 h-5" />
                <Moon v-else class="w-5 h-5" />
                <span class="text-sm font-medium">{{ globalStore.isDark ? $t('sidebar.darkMode') : $t('sidebar.lightMode') }}</span>
            </button>
        </div>
    </div>

    <!-- Main Content -->
    <div class="flex-1 flex flex-col h-full overflow-hidden relative bg-claude-bg dark:bg-claude-dark-bg transition-colors duration-200">
        <main class="flex-1 overflow-y-auto relative scrollbar-hide">
             <!-- Top gradient fade for scroll -->
            <div class="max-w-7xl mx-auto w-full h-full">
                <slot></slot>
            </div>
        </main>
    </div>
  </div>
</template>

<style scoped>
/* Anthropic Menu Overrides */
:deep(.n-menu-item-content) {
    border-radius: 6px !important;
    margin-bottom: 2px;
    padding-left: 12px !important;
}

:deep(.n-menu-item-content-header) {
    font-size: 14px;
    font-weight: 500;
    color: #6F6F78 !important; /* claude-secondaryText */
    transition: color 0.2s;
}

/* Dark mode header */
.dark :deep(.n-menu-item-content-header) {
     color: #9CA3AF !important; /* claude-dark-secondaryText */
}
/* Note: Since we can't easily use Tailwind classes inside deep selectors without apply, and NMenu doesn't fully support scoped styles, we relies on global overrides or explicit colors.
   Ideally we should use CSS vars. Let's try CSS vars approach later if this is messy. 
   But for now, simply overriding the dark/light colors. 
   Actually, the best way is to use the .dark class on the body/parent to style these.
*/

/* Icon Color */
:deep(.n-menu-item-content .n-icon) {
    color: #6F6F78 !important;
    transition: color 0.2s;
}

.dark :deep(.n-menu-item-content .n-icon) {
    color: #6B7280 !important;
}

/* Hover State */
:deep(.n-menu-item-content:hover:not(.n-menu-item-content--selected)) {
    background-color: #E6E4E1 !important; /* claude-hover */
}

/* Dark Hover */
.dark :deep(.n-menu-item-content:hover:not(.n-menu-item-content--selected)) {
    background-color: #212124 !important; /* claude-dark-hover */
}

:deep(.n-menu-item-content:hover .n-menu-item-content-header) {
    color: #1F1E1D !important; /* claude-text */
}
:deep(.n-menu-item-content:hover .n-icon) {
    color: #1F1E1D !important;
}

.dark :deep(.n-menu-item-content:hover .n-menu-item-content-header) {
     color: #E5E7EB !important;
}
.dark :deep(.n-menu-item-content:hover .n-icon) {
     color: #E5E7EB !important;
}

/* Selected State */
:deep(.n-menu-item-content--selected) {
    background-color: #E1DFDD !important; /* claude-border/lighter selection for light mode */
}
:deep(.n-menu-item-content--selected::before) {
    display: none !important;
}

.dark :deep(.n-menu-item-content--selected) {
    background-color: rgb(48,48,48) !important;
}

:deep(.n-menu-item-content--selected .n-menu-item-content-header) {
    color: #1F1E1D !important;
}
:deep(.n-menu-item-content--selected .n-icon) {
    color: #1F1E1D !important;
}

.dark :deep(.n-menu-item-content--selected .n-menu-item-content-header) {
    color: #FFFFFF !important;
}
.dark :deep(.n-menu-item-content--selected .n-icon) {
    color: #FFFFFF !important;
}

/* Menu Indent */
:deep(.n-menu .n-menu-item) {
    margin-top: 0;
}
</style>

