import apiClient from './client'
import type { KeyInfo, ApiResponse, ListResponse, KeyImportItem } from './types'

/**
 * Validation result returned from /api/keys/validate
 */
export interface ValidateKeyResult {
    /** Whether the key is valid */
    valid: boolean;
    /** API latency in milliseconds */
    latency_ms: number;
    /** List of available models (empty if invalid) */
    models: string[];
    /** Error message (present if invalid) */
    error?: string;
}

/**
 * Validate an API key and fetch available models
 * @param data - Key and optional provider
 */
export const validateKey = async (data: { key: string; provider?: string }) =>
    (await apiClient.post<ApiResponse<ValidateKeyResult>>('/api/keys/validate', data)) as unknown as ApiResponse<ValidateKeyResult>

/**
 * Get list of all API keys
 */
export const getKeys = async () => (await apiClient.get<ListResponse<KeyInfo>>('/api/keys')) as unknown as ListResponse<KeyInfo>

/**
 * Create a new API key
 * @param data - Key creation payload with optional provider and default_model
 */
export const addKey = async (data: { key: string; name?: string; tags?: string[]; provider?: string; default_model?: string }) =>
    (await apiClient.post<ApiResponse<KeyInfo>>('/api/keys', data)) as unknown as ApiResponse<KeyInfo>

export const deleteKey = async (id: string) =>
    (await apiClient.delete<ApiResponse<void>>(`/api/keys/${id}`)) as unknown as ApiResponse<void>

export const updateKey = async (id: string, data: { name?: string; tags?: string[]; enabled?: boolean }) =>
    (await apiClient.put<ApiResponse<KeyInfo>>(`/api/keys/${id}`, data)) as unknown as ApiResponse<KeyInfo>

export const testKey = async (id: string) =>
    (await apiClient.post<ApiResponse<{ valid: boolean; latency_ms: number }>>(`/api/keys/${id}/test`)) as unknown as ApiResponse<{ valid: boolean; latency_ms: number }>

// ... (existing code)

export const importKeys = async (data: { keys: KeyImportItem[] }) =>
    (await apiClient.post<ApiResponse<{ imported: number; skipped: number; errors: string[] }>>('/api/keys/import', data)) as unknown as ApiResponse<{ imported: number; skipped: number; errors: string[] }>

export const exportKeys = async () =>
    (await apiClient.get('/api/keys/export', { responseType: 'blob' })) as unknown as Blob
