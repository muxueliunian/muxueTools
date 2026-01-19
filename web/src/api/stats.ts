import axios from 'axios';
import type { StatsTimeRange, TrendResponse, ModelUsageResponse } from './types';

// Use same base setup as other API clients
const API_BASE = '/api';

export const statsApi = {
    getTrend: async (range: StatsTimeRange = '7d') => {
        const res = await axios.get<TrendResponse>(`${API_BASE}/stats/trend`, {
            params: { range }
        });
        return res.data;
    },

    getModels: async () => {
        const res = await axios.get<ModelUsageResponse>(`${API_BASE}/stats/models`);
        return res.data;
    }
};
