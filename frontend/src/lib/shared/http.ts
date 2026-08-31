import type { KesalahanApi, ResponsApi } from '../../types/domain'

const API_URL = (import.meta.env.VITE_API_URL || '').replace(/\/$/, '')

export async function request<T = any>(path: string, options: RequestInit = {}): Promise<T> {
  const isFormData = options.body instanceof FormData
  const headers = {
    Accept: 'application/json',
    ...(!isFormData ? { 'Content-Type': 'application/json' } : {}),
    ...options.headers,
  }

  const response = await fetch(`${API_URL}${path}`, {
    ...options,
    headers,
  })

  const payload = await response.json().catch(() => null) as ResponsApi<T> | null
  if (!response.ok) {
    const error = new Error(payload?.error?.message || 'Tidak dapat terhubung ke server.') as KesalahanApi
    error.code = payload?.error?.code
    error.status = response.status
    throw error
  }

  return payload?.data as T
}

export async function requestBlob(path: string, options: RequestInit = {}): Promise<Blob> {
  const response = await fetch(`${API_URL}${path}`, options)
  if (!response.ok) {
    const payload = await response.json().catch(() => null) as ResponsApi<never> | null
    const error = new Error(payload?.error?.message || 'File tidak dapat dibaca dari server.') as KesalahanApi
    error.code = payload?.error?.code
    error.status = response.status
    throw error
  }
  return response.blob()
}
