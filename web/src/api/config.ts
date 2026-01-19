import apiClient from './client'
import type { ApiResponse } from './types'

export interface ConfigInfo {
    server: {
        port: number;
        host: string;
    };
    pool: {
        strategy: 'round_robin' | 'random' | 'least_used' | 'weighted';
        cooldown_seconds: number;
        max_retries: number;
    };
    logging: {
        level: 'debug' | 'info' | 'warn' | 'error';
    };
    update: {
        enabled: boolean;
        check_interval: string;
        source?: 'mxln' | 'github';
    };
    security?: {
        ip_whitelist_enabled: boolean;
        whitelist_ip: string;
        proxy_key: string;
    };
    advanced?: {
        request_timeout: number;  // seconds
    };
}

export interface UpdateInfo {
    has_update: boolean;
    current_version: string;
    latest_version: string;
    download_url?: string;
    changelog?: string;
}

export interface RegenerateKeyResponse {
    proxy_key: string;
}

/**
 * Get current system configuration
 */
export const getConfig = async () => (await apiClient.get<ApiResponse<ConfigInfo>>('/api/config')) as unknown as ApiResponse<ConfigInfo>

/**
 * Update system configuration
 * @param data - Partial config to update
 */
export const updateConfig = async (data: Partial<ConfigInfo>) =>
    (await apiClient.put<ApiResponse<ConfigInfo>>('/api/config', data)) as unknown as ApiResponse<ConfigInfo>

/**
 * Check for available updates
 */
export const checkUpdate = async () => (await apiClient.get<ApiResponse<UpdateInfo>>('/api/update/check')) as unknown as ApiResponse<UpdateInfo>

/**
 * Regenerate proxy API key
 */
export const regenerateProxyKey = async () =>
    (await apiClient.post<ApiResponse<RegenerateKeyResponse>>('/api/config/regenerate-proxy-key')) as unknown as ApiResponse<RegenerateKeyResponse>

/**
 * Clear all chat sessions and messages
 */
export const clearAllSessions = async () =>
    (await apiClient.delete<ApiResponse<{ deleted: number }>>('/api/sessions')) as unknown as ApiResponse<{ deleted: number }>

/**
 * Reset all key statistics
 */
export const resetStats = async () =>
    (await apiClient.delete<ApiResponse<{ keys_affected: number }>>('/api/stats/reset')) as unknown as ApiResponse<{ keys_affected: number }>

