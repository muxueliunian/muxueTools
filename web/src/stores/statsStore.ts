import { defineStore } from 'pinia';
import { ref } from 'vue';
import { statsApi } from '../api/stats';
import type { StatsTimeRange, TrendDataPoint, ModelUsageItem } from '../api/types';

export const useStatsStore = defineStore('stats', () => {
    // State
    const trendData = ref<TrendDataPoint[]>([]);
    const modelUsage = ref<ModelUsageItem[]>([]);
    const timeRange = ref<StatsTimeRange>('7d');
    const loading = ref(false);
    const error = ref<string | null>(null);

    // Actions
    async function fetchTrend(range?: StatsTimeRange) {
        if (range) {
            timeRange.value = range;
        }

        loading.value = true;
        error.value = null;

        try {
            const res = await statsApi.getTrend(timeRange.value);
            if (res.success) {
                trendData.value = res.data;
            } else {
                throw new Error('Failed to fetch trend data');
            }
        } catch (e) {
            error.value = e instanceof Error ? e.message : 'Unknown error';
            console.error('Fetch trend error:', e);
        } finally {
            loading.value = false;
        }
    }

    async function fetchModelUsage() {
        loading.value = true;
        error.value = null;

        try {
            const res = await statsApi.getModels();
            if (res.success) {
                modelUsage.value = res.data;
            } else {
                throw new Error('Failed to fetch model usage');
            }
        } catch (e) {
            error.value = e instanceof Error ? e.message : 'Unknown error';
            console.error('Fetch model usage error:', e);
        } finally {
            loading.value = false;
        }
    }

    async function fetchAll() {
        await Promise.all([
            fetchTrend(),
            fetchModelUsage()
        ]);
    }

    return {
        // State
        trendData,
        modelUsage,
        timeRange,
        loading,
        error,

        // Actions
        fetchTrend,
        fetchModelUsage,
        fetchAll
    };
});
