const API_URL = (import.meta.env.VITE_API_URL || '').replace(/\/$/, '')

export async function request(path, options = {}) {
  const response = await fetch(`${API_URL}${path}`, {
    ...options,
    headers: {
      Accept: 'application/json',
      'Content-Type': 'application/json',
      ...options.headers,
    },
  })

  const payload = await response.json().catch(() => null)
  if (!response.ok) {
    const error = new Error(payload?.error?.message || 'Tidak dapat terhubung ke server.')
    error.code = payload?.error?.code
    error.status = response.status
    throw error
  }

  return payload?.data
}
