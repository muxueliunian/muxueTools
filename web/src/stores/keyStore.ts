import { defineStore } from 'pinia'
import { ref } from 'vue'
import { getKeys, addKey, deleteKey, updateKey, testKey, importKeys } from '../api/keys'
import type { KeyInfo, KeyImportItem } from '../api/types'


/**
 * Key Management Store
 * Manages the state of API keys including fetching, creating, and deleting.
 */
export const useKeyStore = defineStore('keys', () => {
    const keys = ref<KeyInfo[]>([])
    const loading = ref(false)
    const error = ref<string | null>(null)

    /**
     * Fetch all keys from backend
     * @param force - Force refresh even if data exists
     */
    async function fetchKeys(force = false) {
        if (!force && keys.value.length > 0) return

        loading.value = true
        error.value = null
        try {
            const res = await getKeys()
            if (res.success && res.data) {
                keys.value = res.data
            } else {
                error.value = res.message || 'Failed to fetch keys'
            }
        } catch (e) {
            if (e instanceof Error) {
                error.value = e.message
            } else {
                error.value = 'Network error'
            }
        } finally {
            loading.value = false
        }
    }

    /**
     * Create a new API key
     * @param data - Key creation payload with optional provider and default_model
     */
    async function createKey(data: { key: string; name?: string; tags?: string[]; provider?: string; default_model?: string }) {
        try {
            const res = await addKey(data)
            if (res.success && res.data) {
                keys.value.unshift(res.data)
                return true
            }
            return false
        } catch (e) {
            return false
        }
    }

    /**
     * Revoke an API key
     */
    async function removeKey(id: string) {
        try {
            const res = await deleteKey(id)
            if (res.success) {
                keys.value = keys.value.filter(k => k.id !== id)
                return true
            }
            return false
        } catch (e) {
            return false
        }
    }

    async function updateKeyInfo(id: string, data: { name?: string; tags?: string[]; enabled?: boolean }) {
        try {
            const res = await updateKey(id, data)
            if (res.success && res.data) {
                const index = keys.value.findIndex(k => k.id === id)
                if (index !== -1) {
                    keys.value[index] = res.data
                }
                return true
            }
            return false
        } catch (e) {
            return false
        }
    }

    /**
     * Test connection for a specific key
     */
    async function testKeyConnection(id: string) {
        try {
            const res = await testKey(id)
            return res.success ? res.data : null
        } catch (e) {
            return null
        }
    }

    async function importBatchKeys(keys: KeyImportItem[]) {
        try {
            const res = await importKeys({ keys })
            if (res.success && res.data) {
                // Refresh list to show new keys
                await fetchKeys(true)
                return res.data
            }
            return null
        } catch (e) {
            return null
        }
    }

    return {
        keys,
        loading,
        error,
        fetchKeys,
        createKey,
        removeKey,
        updateKeyInfo,
        testKeyConnection,
        importBatchKeys
    }
})
