export function awalMedisIgdData(token, noRawat) {
  const query = new URLSearchParams({ no_rawat: noRawat }).toString()
  return request(`/api/awal-medis-igd?${query}`, { headers: { Authorization: `Bearer ${token}` } })
}

export function cariDokterAwalMedisIgd(token, kataKunci) {
  const query = new URLSearchParams({ q: kataKunci }).toString()
  return request(`/api/awal-medis-igd/dokter?${query}`, { headers: { Authorization: `Bearer ${token}` } })
}

export function simpanAwalMedisIgd(token, payload) {
  return request('/api/awal-medis-igd', { method: 'POST', headers: { Authorization: `Bearer ${token}` }, body: JSON.stringify(payload) })
}

export function ubahAwalMedisIgd(token, payload) {
  return request('/api/awal-medis-igd', { method: 'PUT', headers: { Authorization: `Bearer ${token}` }, body: JSON.stringify(payload) })
}

export function hapusAwalMedisIgd(token, noRawat) {
  const query = new URLSearchParams({ no_rawat: noRawat }).toString()
  return request(`/api/awal-medis-igd?${query}`, { method: 'DELETE', headers: { Authorization: `Bearer ${token}` } })
}
