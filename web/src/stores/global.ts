import { defineStore } from 'pinia';
import { ref } from 'vue';

export const useGlobalStore = defineStore('global', () => {
    const isDark = ref(true);
    const sidebarCollapsed = ref(false);

    function toggleTheme() {
        isDark.value = !isDark.value;
    }

    function toggleSidebar() {
        sidebarCollapsed.value = !sidebarCollapsed.value;
    }

    return { isDark, sidebarCollapsed, toggleTheme, toggleSidebar };
});
