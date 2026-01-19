<script setup lang="ts">
import { computed } from 'vue'
import { RouterView } from 'vue-router'
import { NConfigProvider, NMessageProvider, NDialogProvider, darkTheme } from 'naive-ui'
import { useGlobalStore } from '@/stores/global'
import { themeOverrides } from '@/theme'
import MainLayout from '@/layouts/MainLayout.vue'

const globalStore = useGlobalStore()

const theme = computed(() => globalStore.isDark ? darkTheme : null)
</script>

<template>
  <n-config-provider :theme="theme" :theme-overrides="themeOverrides" abstract>
    <n-message-provider>
      <n-dialog-provider>
        <div :class="{ dark: globalStore.isDark }" class="h-full">
            <MainLayout>
               <RouterView />
            </MainLayout>
        </div>
      </n-dialog-provider>
    </n-message-provider>
  </n-config-provider>
</template>

