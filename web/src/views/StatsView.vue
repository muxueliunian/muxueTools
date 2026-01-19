<script setup lang="ts">
import { computed, onMounted } from 'vue';
import { useStatsStore } from '../stores/statsStore';
import { NGrid, NGridItem, NSelect, NIcon, NSpin, NEmpty } from 'naive-ui';
import { Activity, PieChart, BarChart3 } from 'lucide-vue-next';
import VChart from 'vue-echarts';
import { use } from 'echarts/core';
import { CanvasRenderer } from 'echarts/renderers';
import { LineChart, PieChart as EPieChart } from 'echarts/charts';
import {
    TitleComponent,
    TooltipComponent,
    LegendComponent,
    GridComponent,
    DataZoomComponent
} from 'echarts/components';

// Register ECharts components
use([
    CanvasRenderer,
    LineChart,
    EPieChart,
    TitleComponent,
    TooltipComponent,
    LegendComponent,
    GridComponent,
    DataZoomComponent
]);

// Store
const statsStore = useStatsStore();

// Time range options
const rangeOptions = [
    { label: 'Last 24 Hours', value: '24h' },
    { label: 'Last 7 Days', value: '7d' },
    { label: 'Last 30 Days', value: '30d' }
];

// Computed Summary Data (from trend data)
const summary = computed(() => {
    const data = statsStore.trendData;
    const totalRequests = data.reduce((sum, item) => sum + item.requests, 0);
    const totalTokens = data.reduce((sum, item) => sum + item.tokens, 0);
    const totalErrors = data.reduce((sum, item) => sum + item.errors, 0);
    const errorRate = totalRequests > 0 ? (totalErrors / totalRequests * 100).toFixed(2) : '0.00';
    
    return {
        requests: totalRequests,
        tokens: totalTokens,
        errorRate: errorRate + '%'
    };
});

// Chart Options: Trend (Line)
const trendOption = computed(() => {
    const isDark = document.documentElement.classList.contains('dark');
    const textColor = isDark ? '#ccc' : '#333';
    const gridColor = isDark ? '#333' : '#eee';

    return {
        backgroundColor: 'transparent',
        tooltip: {
            trigger: 'axis',
            backgroundColor: isDark ? '#333' : '#fff',
            borderColor: isDark ? '#555' : '#ccc',
            textStyle: { color: textColor }
        },
        legend: {
            data: ['Requests', 'Errors'],
            textStyle: { color: textColor },
            bottom: 0
        },
        grid: {
            left: '3%',
            right: '4%',
            bottom: '10%',
            containLabel: true
        },
        xAxis: {
            type: 'category',
            boundaryGap: false,
            data: statsStore.trendData.map(item => {
                const date = new Date(item.timestamp);
                // Format depends on range
                if (statsStore.timeRange === '24h') {
                    return date.getHours() + ':00';
                }
                const month = (date.getMonth() + 1).toString().padStart(2, '0');
                const day = date.getDate().toString().padStart(2, '0');
                return `${month}-${day}`;
            }),
            axisLine: { lineStyle: { color: gridColor } },
            axisLabel: { color: textColor }
        },
        yAxis: {
            type: 'value',
            splitLine: { lineStyle: { color: gridColor } },
            axisLabel: { color: textColor }
        },
        series: [
            {
                name: 'Requests',
                type: 'line',
                smooth: true,
                showSymbol: false,
                data: statsStore.trendData.map(item => item.requests),
                itemStyle: { color: '#D97757' },
                areaStyle: {
                    color: {
                        type: 'linear',
                        x: 0, y: 0, x2: 0, y2: 1,
                        colorStops: [
                            { offset: 0, color: 'rgba(217, 119, 87, 0.5)' },
                            { offset: 1, color: 'rgba(217, 119, 87, 0)' }
                        ]
                    }
                }
            },
            {
                name: 'Errors',
                type: 'line',
                smooth: true,
                showSymbol: false,
                data: statsStore.trendData.map(item => item.errors),
                itemStyle: { color: '#ef4444' },
                lineStyle: { type: 'dashed' }
            }
        ]
    };
});

// Chart Options: Model Usage (Pie)
const modelOption = computed(() => {
    const isDark = document.documentElement.classList.contains('dark');
    const textColor = isDark ? '#ccc' : '#333';

    return {
        backgroundColor: 'transparent',
        tooltip: {
            trigger: 'item',
            backgroundColor: isDark ? '#333' : '#fff',
            borderColor: isDark ? '#555' : '#ccc',
            textStyle: { color: textColor },
            formatter: '{b}: {c} ({d}%)'
        },
        legend: {
            top: '5%',
            left: 'center',
            textStyle: { color: textColor }
        },
        series: [
            {
                name: 'Model Usage',
                type: 'pie',
                radius: ['40%', '70%'],
                center: ['50%', '60%'],
                roseType: 'radius',
                itemStyle: {
                    borderRadius: 8
                },
                label: {
                    show: false
                },
                emphasis: {
                    label: {
                        show: true
                    }
                },
                data: statsStore.modelUsage.map(item => ({
                    value: item.request_count,
                    name: item.model
                }))
            }
        ]
    };
});

function handleRangeChange(val: string) {
    statsStore.fetchTrend(val as any);
}

onMounted(() => {
    statsStore.fetchAll();
});
</script>

<template>
    <div class="min-h-screen bg-claude-bg dark:bg-claude-dark-bg text-claude-text dark:text-claude-dark-text p-8 transition-colors duration-200">
        <div class="max-w-6xl mx-auto space-y-8">
            
            <!-- Header -->
            <div class="flex flex-col md:flex-row md:items-center justify-between gap-4">
                <div>
                    <h1 class="text-3xl font-light tracking-tight mb-2">Statistics</h1>
                    <p class="text-claude-secondaryText dark:text-gray-500 text-sm">Monitor API usage, trends, and model distribution.</p>
                </div>
                
                <!-- Time Range Selector -->
                <div class="w-48">
                    <n-select 
                        v-model:value="statsStore.timeRange" 
                        :options="rangeOptions" 
                        @update:value="handleRangeChange"
                        size="medium"
                    />
                </div>
            </div>

            <!-- Loading State -->
            <div v-if="statsStore.loading && !statsStore.trendData.length" class="flex justify-center py-20">
                <n-spin size="large" />
            </div>

            <template v-else>
                <!-- Summary Cards -->
                <n-grid x-gap="16" y-gap="16" :cols="3" responsive="screen" item-responsive>
                    <n-grid-item span="3 s:3 m:1">
                        <div class="bg-white dark:bg-[#212124] border border-claude-border dark:border-[#2A2A2E] rounded-xl p-6 h-full">
                            <div class="flex items-center gap-3 mb-2">
                                <n-icon :component="Activity" class="text-[#D97757]" size="20" />
                                <span class="text-sm text-claude-secondaryText dark:text-gray-500">Total Requests</span>
                            </div>
                            <div class="text-3xl font-light">{{ summary.requests.toLocaleString() }}</div>
                        </div>
                    </n-grid-item>
                    <n-grid-item span="3 s:3 m:1">
                        <div class="bg-white dark:bg-[#212124] border border-claude-border dark:border-[#2A2A2E] rounded-xl p-6 h-full">
                            <div class="flex items-center gap-3 mb-2">
                                <n-icon :component="BarChart3" class="text-emerald-500" size="20" />
                                <span class="text-sm text-claude-secondaryText dark:text-gray-500">Total Tokens</span>
                            </div>
                            <div class="text-3xl font-light">{{ summary.tokens.toLocaleString() }}</div>
                        </div>
                    </n-grid-item>
                    <n-grid-item span="3 s:3 m:1">
                        <div class="bg-white dark:bg-[#212124] border border-claude-border dark:border-[#2A2A2E] rounded-xl p-6 h-full">
                            <div class="flex items-center gap-3 mb-2">
                                <n-icon :component="PieChart" class="text-red-500" size="20" />
                                <span class="text-sm text-claude-secondaryText dark:text-gray-500">Error Rate</span>
                            </div>
                            <div class="text-3xl font-light">{{ summary.errorRate }}</div>
                        </div>
                    </n-grid-item>
                </n-grid>

                <!-- Charts Section -->
                <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
                    <!-- Trend Chart -->
                    <div class="lg:col-span-2 bg-white dark:bg-[#212124] border border-claude-border dark:border-[#2A2A2E] rounded-xl p-6">
                         <h3 class="text-lg font-medium mb-6">Request Trend</h3>
                         <div class="h-[350px]">
                             <v-chart class="chart" :option="trendOption" :autoresize="true" />
                         </div>
                    </div>

                    <!-- Model Distribution -->
                    <div class="bg-white dark:bg-[#212124] border border-claude-border dark:border-[#2A2A2E] rounded-xl p-6">
                        <h3 class="text-lg font-medium mb-6">Model Distribution</h3>
                        <div class="h-[350px]">
                             <v-chart v-if="statsStore.modelUsage.length" class="chart" :option="modelOption" :autoresize="true" />
                             <div v-else class="h-full flex items-center justify-center">
                                 <n-empty description="No model usage data" />
                             </div>
                        </div>
                    </div>
                </div>
            </template>
        </div>
    </div>
</template>

<style scoped>
.chart {
    height: 100%;
    width: 100%;
}
</style>
