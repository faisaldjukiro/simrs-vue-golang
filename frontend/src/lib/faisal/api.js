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

export function aktivitasLogData(token, params = {}) {
  const query = new URLSearchParams(params).toString()
  const path = query ? `/api/aktivitas-log?${query}` : '/api/aktivitas-log'
  return request(path, { headers: { Authorization: `Bearer ${token}` } })
}

export function idrgData(token, params = {}) {
  const query = new URLSearchParams(params).toString()
  const path = query ? `/api/idrg?${query}` : '/api/idrg'

  return request(path, {
    headers: { Authorization: `Bearer ${token}` },
  })
}

export function monitoringDataKlaim(token, params = {}) {
  const query = new URLSearchParams(params).toString()
  return request(`/api/bpjs/monitoring/klaim?${query}`, {
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

export function kelolaMenuData(token) {
  return request('/api/kelola-menu', {
    headers: { Authorization: `Bearer ${token}` },
  })
}

export function tambahSidebarPasien(token, payload) {
  return request('/api/kelola-menu/sidebar', {
    method: 'POST',
    headers: { Authorization: `Bearer ${token}` },
    body: JSON.stringify(payload),
  })
}

export function ubahSidebarPasien(token, id, payload) {
  return request(`/api/kelola-menu/sidebar/${id}`, {
    method: 'PUT',
    headers: { Authorization: `Bearer ${token}` },
    body: JSON.stringify(payload),
  })
}

export function hapusSidebarPasien(token, id) {
  return request(`/api/kelola-menu/sidebar/${id}`, {
    method: 'DELETE',
    headers: { Authorization: `Bearer ${token}` },
  })
}

export function cpptData(token, noRawat, jenisRawat = 'ranap') {
  const query = new URLSearchParams({ no_rawat: noRawat, jenis_rawat: jenisRawat }).toString()
  return request(`/api/cppt?${query}`, {
    headers: { Authorization: `Bearer ${token}` },
  })
}

export function simpanCppt(token, payload) {
  return request('/api/cppt', {
    method: 'POST',
    headers: { Authorization: `Bearer ${token}` },
    body: JSON.stringify(payload),
  })
}

export function ubahCppt(token, payload) {
  return request('/api/cppt', {
    method: 'PUT',
    headers: { Authorization: `Bearer ${token}` },
    body: JSON.stringify(payload),
  })
}

export function hapusCppt(token, payload) {
  return request('/api/cppt', {
    method: 'DELETE',
    headers: { Authorization: `Bearer ${token}` },
    body: JSON.stringify(payload),
  })
}

export function cariPetugasCppt(token, kataKunci) {
  const query = new URLSearchParams({ q: kataKunci }).toString()
  return request(`/api/cppt/petugas?${query}`, {
    headers: { Authorization: `Bearer ${token}` },
  })
}

export function penangananDokterPetugasData(token, noRawat, jenisRawat = 'ranap') {
  const query = new URLSearchParams({ no_rawat: noRawat, jenis_rawat: jenisRawat }).toString()
  return request(`/api/penanganan-dokter-petugas?${query}`, { headers: { Authorization: `Bearer ${token}` } })
}

export function cariDokterPenanganan(token, kataKunci) {
  const query = new URLSearchParams({ q: kataKunci }).toString()
  return request(`/api/penanganan-dokter-petugas/dokter?${query}`, { headers: { Authorization: `Bearer ${token}` } })
}

export function cariPetugasPenanganan(token, kataKunci) {
  const query = new URLSearchParams({ q: kataKunci }).toString()
  return request(`/api/penanganan-dokter-petugas/petugas?${query}`, { headers: { Authorization: `Bearer ${token}` } })
}

export function cariTindakanPenanganan(token, noRawat, kataKunci, jenisRawat = 'ranap') {
  const query = new URLSearchParams({ no_rawat: noRawat, q: kataKunci, jenis_rawat: jenisRawat }).toString()
  return request(`/api/penanganan-dokter-petugas/tindakan?${query}`, { headers: { Authorization: `Bearer ${token}` } })
}

export function simpanPenangananDokterPetugas(token, payload) {
  return request('/api/penanganan-dokter-petugas', { method: 'POST', headers: { Authorization: `Bearer ${token}` }, body: JSON.stringify(payload) })
}

export function simpanBanyakPenangananDokterPetugas(token, payload) {
  return request('/api/penanganan-dokter-petugas/banyak', { method: 'POST', headers: { Authorization: `Bearer ${token}` }, body: JSON.stringify(payload) })
}

export function ubahPenangananDokterPetugas(token, payload) {
  return request('/api/penanganan-dokter-petugas', { method: 'PUT', headers: { Authorization: `Bearer ${token}` }, body: JSON.stringify(payload) })
}

export function hapusPenangananDokterPetugas(token, payload) {
  return request('/api/penanganan-dokter-petugas', { method: 'DELETE', headers: { Authorization: `Bearer ${token}` }, body: JSON.stringify(payload) })
}

export function permintaanRadiologiData(token, noRawat) {
  const query = new URLSearchParams({ no_rawat: noRawat }).toString()
  return request(`/api/permintaan-radiologi?${query}`, { headers: { Authorization: `Bearer ${token}` } })
}

export function cariDokterRadiologi(token, kataKunci) {
  const query = new URLSearchParams({ q: kataKunci }).toString()
  return request(`/api/permintaan-radiologi/dokter?${query}`, { headers: { Authorization: `Bearer ${token}` } })
}

export function cariTindakanRadiologi(token, noRawat, kataKunci) {
  const query = new URLSearchParams({ no_rawat: noRawat, q: kataKunci }).toString()
  return request(`/api/permintaan-radiologi/tindakan?${query}`, { headers: { Authorization: `Bearer ${token}` } })
}

export function simpanPermintaanRadiologi(token, payload) {
  return request('/api/permintaan-radiologi', { method: 'POST', headers: { Authorization: `Bearer ${token}` }, body: JSON.stringify(payload) })
}

export function ubahPermintaanRadiologi(token, nomor, payload) {
  return request(`/api/permintaan-radiologi/${encodeURIComponent(nomor)}`, { method: 'PUT', headers: { Authorization: `Bearer ${token}` }, body: JSON.stringify(payload) })
}

export function hapusPermintaanRadiologi(token, noRawat, nomor) {
  const query = new URLSearchParams({ no_rawat: noRawat }).toString()
  return request(`/api/permintaan-radiologi/${encodeURIComponent(nomor)}?${query}`, { method: 'DELETE', headers: { Authorization: `Bearer ${token}` } })
}

export function permintaanLaboratoriumData(token, noRawat) {
  const query = new URLSearchParams({ no_rawat: noRawat }).toString()
  return request(`/api/permintaan-laboratorium?${query}`, { headers: { Authorization: `Bearer ${token}` } })
}

export function cariDokterLaboratorium(token, kataKunci) {
  const query = new URLSearchParams({ q: kataKunci }).toString()
  return request(`/api/permintaan-laboratorium/dokter?${query}`, { headers: { Authorization: `Bearer ${token}` } })
}

export function cariTindakanLaboratorium(token, noRawat, kataKunci) {
  const query = new URLSearchParams({ no_rawat: noRawat, q: kataKunci }).toString()
  return request(`/api/permintaan-laboratorium/tindakan?${query}`, { headers: { Authorization: `Bearer ${token}` } })
}

export function detailTindakanLaboratorium(token, noRawat, kode) {
  const query = new URLSearchParams({ no_rawat: noRawat }).toString()
  return request(`/api/permintaan-laboratorium/tindakan/${encodeURIComponent(kode)}/detail?${query}`, { headers: { Authorization: `Bearer ${token}` } })
}

export function simpanPermintaanLaboratorium(token, payload) {
  return request('/api/permintaan-laboratorium', { method: 'POST', headers: { Authorization: `Bearer ${token}` }, body: JSON.stringify(payload) })
}

export function ubahPermintaanLaboratorium(token, nomor, payload) {
  return request(`/api/permintaan-laboratorium/${encodeURIComponent(nomor)}`, { method: 'PUT', headers: { Authorization: `Bearer ${token}` }, body: JSON.stringify(payload) })
}

export function hapusPermintaanLaboratorium(token, noRawat, nomor) {
  const query = new URLSearchParams({ no_rawat: noRawat }).toString()
  return request(`/api/permintaan-laboratorium/${encodeURIComponent(nomor)}?${query}`, { method: 'DELETE', headers: { Authorization: `Bearer ${token}` } })
}

export function riwayatPerawatanData(token, params = {}) {
  const query = new URLSearchParams(params).toString()
  return request(`/api/riwayat-perawatan?${query}`, { headers: { Authorization: `Bearer ${token}` } })
}

export function triaseIgdData(token, noRawat) {
  const query = new URLSearchParams({ no_rawat: noRawat }).toString()
  return request(`/api/triase-igd?${query}`, { headers: { Authorization: `Bearer ${token}` } })
}

export function simpanTriaseIgd(token, payload) {
  return request('/api/triase-igd', { method: 'POST', headers: { Authorization: `Bearer ${token}` }, body: JSON.stringify(payload) })
}

export function ubahTriaseIgd(token, payload) {
  return request('/api/triase-igd', { method: 'PUT', headers: { Authorization: `Bearer ${token}` }, body: JSON.stringify(payload) })
}

export function hapusTriaseIgd(token, noRawat, jenis) {
  const query = new URLSearchParams({ no_rawat: noRawat, jenis }).toString()
  return request(`/api/triase-igd?${query}`, { method: 'DELETE', headers: { Authorization: `Bearer ${token}` } })
}

export function awalKeperawatanIgdData(token, noRawat) {
  const query = new URLSearchParams({ no_rawat: noRawat }).toString()
  return request(`/api/awal-keperawatan-igd?${query}`, { headers: { Authorization: `Bearer ${token}` } })
}

export function simpanAwalKeperawatanIgd(token, payload) {
  return request('/api/awal-keperawatan-igd', { method: 'POST', headers: { Authorization: `Bearer ${token}` }, body: JSON.stringify(payload) })
}

export function ubahAwalKeperawatanIgd(token, payload) {
  return request('/api/awal-keperawatan-igd', { method: 'PUT', headers: { Authorization: `Bearer ${token}` }, body: JSON.stringify(payload) })
}

export function hapusAwalKeperawatanIgd(token, noRawat) {
  const query = new URLSearchParams({ no_rawat: noRawat }).toString()
  return request(`/api/awal-keperawatan-igd?${query}`, { method: 'DELETE', headers: { Authorization: `Bearer ${token}` } })
}

export function awalMedisUmumData(token, noRawat) {
  const query = new URLSearchParams({ no_rawat: noRawat }).toString()
  return request(`/api/awal-medis-umum?${query}`, { headers: { Authorization: `Bearer ${token}` } })
}

export function cariDokterAwalMedisUmum(token, kataKunci) {
  const query = new URLSearchParams({ q: kataKunci }).toString()
  return request(`/api/awal-medis-umum/dokter?${query}`, { headers: { Authorization: `Bearer ${token}` } })
}

export function simpanAwalMedisUmum(token, payload) {
  return request('/api/awal-medis-umum', { method: 'POST', headers: { Authorization: `Bearer ${token}` }, body: JSON.stringify(payload) })
}

export function ubahAwalMedisUmum(token, payload) {
  return request('/api/awal-medis-umum', { method: 'PUT', headers: { Authorization: `Bearer ${token}` }, body: JSON.stringify(payload) })
}

export function hapusAwalMedisUmum(token, noRawat) {
  const query = new URLSearchParams({ no_rawat: noRawat }).toString()
  return request(`/api/awal-medis-umum?${query}`, { method: 'DELETE', headers: { Authorization: `Bearer ${token}` } })
}

export function resumePasienRanapData(token, noRawat) {
  const query = new URLSearchParams({ no_rawat: noRawat }).toString()
  return request(`/api/resume-pasien-ranap?${query}`, { headers: { Authorization: `Bearer ${token}` } })
}

export function validasiCodingResumePasienRanap(token, params) {
  const query = new URLSearchParams(params).toString()
  return request(`/api/resume-pasien-ranap/validasi-coding?${query}`, { headers: { Authorization: `Bearer ${token}` } })
}

export function cariCodingResumePasienRanap(token, params) {
  const query = new URLSearchParams(params).toString()
  return request(`/api/resume-pasien-ranap/cari-coding?${query}`, { headers: { Authorization: `Bearer ${token}` } })
}

export function simpanResumePasienRanap(token, payload) {
  return request('/api/resume-pasien-ranap', { method: 'POST', headers: { Authorization: `Bearer ${token}` }, body: JSON.stringify(payload) })
}

export function ubahResumePasienRanap(token, payload) {
  return request('/api/resume-pasien-ranap', { method: 'PUT', headers: { Authorization: `Bearer ${token}` }, body: JSON.stringify(payload) })
}

export function hapusResumePasienRanap(token, noRawat) {
  const query = new URLSearchParams({ no_rawat: noRawat }).toString()
  return request(`/api/resume-pasien-ranap?${query}`, { method: 'DELETE', headers: { Authorization: `Bearer ${token}` } })
}

export function diagnosaPasienData(token, params) {
  const query = new URLSearchParams(params).toString()
  return request(`/api/diagnosa-pasien?${query}`, { headers: { Authorization: `Bearer ${token}` } })
}

export function cariCodingDiagnosaPasien(token, params) {
  const query = new URLSearchParams(params).toString()
  return request(`/api/diagnosa-pasien/cari-coding?${query}`, { headers: { Authorization: `Bearer ${token}` } })
}

export function simpanDiagnosaPasien(token, payload) {
  return request('/api/diagnosa-pasien', { method: 'PUT', headers: { Authorization: `Bearer ${token}` }, body: JSON.stringify(payload) })
}

export function hapusCodingDiagnosaPasien(token, { noRawat, status, jenis, kode }) {
  const query = new URLSearchParams({ no_rawat: noRawat, status }).toString()
  return request(`/api/diagnosa-pasien/${encodeURIComponent(jenis)}/${encodeURIComponent(kode)}?${query}`, {
    method: 'DELETE',
    headers: { Authorization: `Bearer ${token}` },
  })
}

export function awalMedisRanapData(token, noRawat) {
  const query = new URLSearchParams({ no_rawat: noRawat }).toString()
  return request(`/api/awal-medis-ranap?${query}`, { headers: { Authorization: `Bearer ${token}` } })
}

export function cariDokterAwalMedisRanap(token, kataKunci) {
  const query = new URLSearchParams({ q: kataKunci }).toString()
  return request(`/api/awal-medis-ranap/dokter?${query}`, { headers: { Authorization: `Bearer ${token}` } })
}

export function simpanAwalMedisRanap(token, payload) {
  return request('/api/awal-medis-ranap', { method: 'POST', headers: { Authorization: `Bearer ${token}` }, body: JSON.stringify(payload) })
}

export function ubahAwalMedisRanap(token, payload) {
  return request('/api/awal-medis-ranap', { method: 'PUT', headers: { Authorization: `Bearer ${token}` }, body: JSON.stringify(payload) })
}

export function hapusAwalMedisRanap(token, noRawat) {
  const query = new URLSearchParams({ no_rawat: noRawat }).toString()
  return request(`/api/awal-medis-ranap?${query}`, { method: 'DELETE', headers: { Authorization: `Bearer ${token}` } })
}
export function awalMedisIgdData(token, noRawat) {
  const query = new URLSearchParams({ no_rawat: noRawat }).toString()
  return request(`/api/awal-medis-igd?${query}`, { headers: { Authorization: `Bearer ${token}` } })
}

export function resepInfoPasien(token, noRawat) { return request(`/api/resep/info-pasien?${new URLSearchParams({ no_rawat: noRawat })}`, { headers: { Authorization: `Bearer ${token}` } }) }
export function resepDaftar(token, noRawat) { return request(`/api/resep?${new URLSearchParams({ no_rawat: noRawat })}`, { headers: { Authorization: `Bearer ${token}` } }) }
export function resepDaftarCopy(token, noRkmMedis) { return request(`/api/resep/copy?${new URLSearchParams({ no_rkm_medis: noRkmMedis })}`, { headers: { Authorization: `Bearer ${token}` } }) }
export function resepDetail(token, noResep) { return request(`/api/resep/detail/${encodeURIComponent(noResep)}`, { headers: { Authorization: `Bearer ${token}` } }) }
export function resepCariDokter(token, q) { return request(`/api/resep/cari-dokter?${new URLSearchParams({ q })}`, { headers: { Authorization: `Bearer ${token}` } }) }
export function resepCariDepo(token, q) { return request(`/api/resep/cari-depo?${new URLSearchParams({ q })}`, { headers: { Authorization: `Bearer ${token}` } }) }
export function resepCariObat(token, params) { return request(`/api/resep/cari-obat?${new URLSearchParams(params)}`, { headers: { Authorization: `Bearer ${token}` } }) }
export function resepMetodeRacik(token) { return request('/api/resep/metode-racik', { headers: { Authorization: `Bearer ${token}` } }) }
export function resepNomorAuto(token, tanggal) { return request(`/api/resep/nomor-auto?${new URLSearchParams({ tanggal })}`, { headers: { Authorization: `Bearer ${token}` } }) }
export function resepDepoDefault(token, noRawat, status) { return request(`/api/resep/depo-default?${new URLSearchParams({ no_rawat: noRawat, status })}`, { headers: { Authorization: `Bearer ${token}` } }) }
export function simpanResep(token, payload) { return request('/api/resep', { method: 'POST', headers: { Authorization: `Bearer ${token}` }, body: JSON.stringify(payload) }) }
export function hapusResep(token, noResep) { return request(`/api/resep/${encodeURIComponent(noResep)}`, { method: 'DELETE', headers: { Authorization: `Bearer ${token}` } }) }

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
