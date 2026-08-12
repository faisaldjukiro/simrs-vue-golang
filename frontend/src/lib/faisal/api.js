import { request } from '../shared/http'

export function login(credentials) {
  return request('/api/auth/login', {
    method: 'POST',
    body: JSON.stringify(credentials),
  })
}

export function currentUser(token) {
  return request('/api/auth/me', {
    headers: { Authorization: `Bearer ${token}` },
  })
}

export function logout(token) {
  return request('/api/auth/logout', {
    method: 'POST',
    headers: { Authorization: `Bearer ${token}` },
  })
}

export function dashboardData(token, params = {}) {
  const query = new URLSearchParams(params).toString()
  const path = query ? `/api/beranda?${query}` : '/api/beranda'

  return request(path, {
    headers: { Authorization: `Bearer ${token}` },
  })
}

export function userManagementData(token) {
  return request('/api/user-management', {
    headers: { Authorization: `Bearer ${token}` },
  })
}

export function idrgData(token, params = {}) {
  const query = new URLSearchParams(params).toString()
  const path = query ? `/api/idrg?${query}` : '/api/idrg'

  return request(path, {
    headers: { Authorization: `Bearer ${token}` },
  })
}

export function idrgDiagnosa(token, noRawat) {
  const query = new URLSearchParams({ no_rawat: noRawat }).toString()
  return request(`/api/idrg/diagnosa?${query}`, {
    headers: { Authorization: `Bearer ${token}` },
  })
}

export function idrgProsedur(token, noRawat) {
  const query = new URLSearchParams({ no_rawat: noRawat }).toString()
  return request(`/api/idrg/prosedur?${query}`, {
    headers: { Authorization: `Bearer ${token}` },
  })
}

export function prosesIdrg(token, payload) {
  return request('/api/idrg/proses', {
    method: 'POST',
    headers: { Authorization: `Bearer ${token}` },
    body: JSON.stringify(payload),
  })
}

export function cariPegawaiUserManagement(token, kataKunci) {
  const query = new URLSearchParams({ q: kataKunci }).toString()
  return request(`/api/user-management/pegawai?${query}`, {
    headers: { Authorization: `Bearer ${token}` },
  })
}

export function tambahUserManagement(token, payload) {
  return request('/api/user-management', {
    method: 'POST',
    headers: { Authorization: `Bearer ${token}` },
    body: JSON.stringify(payload),
  })
}

export function ubahAksesUserManagement(token, id, payload) {
  return request(`/api/user-management/${id}/akses`, {
    method: 'PUT',
    headers: { Authorization: `Bearer ${token}` },
    body: JSON.stringify(payload),
  })
}
