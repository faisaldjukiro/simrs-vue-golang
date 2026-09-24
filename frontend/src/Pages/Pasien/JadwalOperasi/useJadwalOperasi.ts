import { computed, onBeforeUnmount, reactive, ref, watch } from 'vue'
import { request } from '../../../lib/shared/http'
import { useNotifikasi } from '../../../lib/shared/useNotifikasi'
import type { JadwalOperasi, PropsJadwalOperasi } from '../../../types/jadwalOperasi'

export function useJadwalOperasi(props: PropsJadwalOperasi) {
  const notifikasi = useNotifikasi()
  const loading = ref(false)
  const saving = ref(false)
  const error = ref('')
  const errorSimpan = ref('')
  const pesanKunci = ref('')
  const peringatan = ref('')
  const formVisible = ref(true)
  const records = ref<JadwalOperasi[]>([])
  const editing = ref<JadwalOperasi | null>(null)
  const hapusTarget = ref<JadwalOperasi | null>(null)
  const paket = ref<Record<string, string>>({})
  const dokter = ref<Record<string, string>>({})
  const ruang = ref<Record<string, string>>({})
  const form = reactive({
    tanggal: '', jam_mulai: '00:00:00', jam_selesai: '00:00:00',
    status: 'Permintaan', dokteranastesi: '', perawat: '',
  })
  const filter = reactive({ q: '', status: '', mulai: '', selesai: '' })
  const noRawat = computed(() => String(props.patient.no_rawat || ''))
  const terkunci = computed(() => saving.value || loading.value || !!pesanKunci.value || !!error.value || !noRawat.value)
  let generasi = 0
  let requestID = 0
  const filteredRows = computed(() => records.value
    .filter(r => (!filter.status || r.status === filter.status)
      && (!filter.mulai || r.tanggal >= filter.mulai)
      && (!filter.selesai || r.tanggal <= filter.selesai)
      && [r.nama_paket, r.kode_paket, r.nama_dokter, r.nama_ruang, r.dokteranastesi, r.perawat, r.sumber]
        .join(' ').toLowerCase().includes(filter.q.trim().toLowerCase()))
    .map((r, i) => ({ ...r, kunci: r.sumber + ':' + r.id + ':' + i })))

  function api<T>(path = '', options: RequestInit = {}) {
    return request<T>('/api/jadwal-operasi' + path, {
      ...options, headers: { Authorization: 'Bearer ' + props.token },
    })
  }
  function reset() {
    editing.value = null
    errorSimpan.value = ''
    paket.value = {}
    dokter.value = {}
    ruang.value = {}
    Object.assign(form, {
      tanggal: new Date(Date.now() + 8 * 3600000).toISOString().slice(0, 10),
      jam_mulai: '00:00:00', jam_selesai: '00:00:00',
      status: 'Permintaan', dokteranastesi: '', perawat: '',
    })
  }
  async function muat() {
    const konteks = generasi
    const urutan = ++requestID
    if (!noRawat.value) { error.value = 'Pilih kunjungan pasien terlebih dahulu.'; return }
    loading.value = true
    error.value = ''
    try {
      const hasil = await api<{ jadwal: JadwalOperasi[]; pesan_kunci: string; peringatan: string }>(
        '?no_rawat=' + encodeURIComponent(noRawat.value),
      )
      if (konteks !== generasi || urutan !== requestID) return
      records.value = hasil.jadwal
      pesanKunci.value = hasil.pesan_kunci
      peringatan.value = hasil.peringatan
    } catch (e) {
      if (konteks === generasi && urutan === requestID) error.value = pesan(e)
    } finally {
      if (konteks === generasi && urutan === requestID) loading.value = false
    }
  }
  async function cari(jenis: string, q: string) {
    return api<Record<string, string>[]>('/referensi?' + new URLSearchParams({ jenis, q, no_rawat: noRawat.value }))
  }
  function edit(r: JadwalOperasi) {
    if (terkunci.value || !r.bisa_ubah) return
    reset()
    editing.value = r
    for (const key of Object.keys(form) as (keyof typeof form)[]) form[key] = r[key]
    paket.value = { kode: r.kode_paket, nama: r.nama_paket }
    dokter.value = { kode: r.kd_dokter, nama: r.nama_dokter }
    ruang.value = { kode: r.kd_ruang_ok, nama: r.nama_ruang }
    formVisible.value = true
  }
  async function simpan() {
    if (terkunci.value) return
    errorSimpan.value = ''
    if (!paket.value.kode || !dokter.value.kode || !ruang.value.kode) {
      errorSimpan.value = 'Pilih paket operasi, operator, dan ruang OK.'
      return
    }
    const konteks = generasi
    saving.value = true
    try {
      const jam = (s: string) => s.length === 5 ? s + ':00' : s
      const hasil = await api<{ pesan: string }>(editing.value ? '/' + (editing.value.sumber === 'Khanza' ? 'khanza' : editing.value.id) : '', {
        method: editing.value ? 'PUT' : 'POST',
        body: JSON.stringify({
          asli: editing.value?.sumber === 'Khanza' ? editing.value : undefined,
          ...form, jam_mulai: jam(form.jam_mulai), jam_selesai: jam(form.jam_selesai),
          no_rawat: noRawat.value, versi: editing.value?.versi || 0,
          kode_paket: paket.value.kode, kd_dokter: dokter.value.kode, kd_ruang_ok: ruang.value.kode,
        }),
      })
      if (konteks !== generasi) return
      notifikasi.sukses(hasil.pesan)
      reset()
      await muat()
    } catch (e) {
      if (konteks === generasi) errorSimpan.value = pesan(e)
    } finally {
      if (konteks === generasi) saving.value = false
    }
  }
  async function hapus() {
    const r = hapusTarget.value
    if (!r || !r.bisa_ubah || terkunci.value) return
    const konteks = generasi
    saving.value = true
    try {
      const hasil = await api<{ pesan: string }>('/' + (r.sumber === 'Khanza' ? 'khanza' : r.id), {
        method: 'DELETE', body: JSON.stringify({ no_rawat: r.no_rawat, versi: r.versi, asli: r.sumber === 'Khanza' ? r : undefined }),
      })
      if (konteks !== generasi) return
      hapusTarget.value = null
      if (editing.value?.id === r.id) reset()
      notifikasi.sukses(hasil.pesan)
      await muat()
    } catch (e) {
      if (konteks === generasi) notifikasi.gagal(pesan(e))
    } finally {
      if (konteks === generasi) saving.value = false
    }
  }
  function cetak() {
    if (loading.value || error.value || !filteredRows.value.length) return
    const popup = window.open('', '_blank', 'width=1000,height=700')
    if (!popup) { notifikasi.peringatan('Izinkan popup untuk mencetak jadwal.'); return }
    popup.opener = null
    const doc = popup.document
    doc.title = 'Jadwal Operasi Pasien'
    const style = doc.createElement('style')
    style.textContent = 'body{font:12px Arial;color:#111;padding:20px}table{width:100%;border-collapse:collapse}td,th{padding:8px;border:1px solid #999;text-align:left}tr{break-inside:avoid}@page{size:landscape}'
    doc.head.append(style)
    const title = doc.createElement('h1')
    title.textContent = 'Jadwal Operasi Pasien'
    const patient = doc.createElement('p')
    patient.textContent = String(props.patient.nama_pasien || props.patient.nm_pasien || '') + ' · ' + noRawat.value + ' · Waktu WITA'
    const table = doc.createElement('table')
    const head = table.createTHead().insertRow()
    for (const label of ['Tanggal', 'Mulai', 'Selesai', 'Operasi', 'Operator', 'Ruang OK', 'Status', 'Anestesi', 'Perawat', 'Sumber']) {
      const th = doc.createElement('th'); th.textContent = label; head.append(th)
    }
    const tbody = table.createTBody()
    for (const r of filteredRows.value) {
      const row = tbody.insertRow()
      for (const value of [r.tanggal, r.jam_mulai, r.jam_selesai, r.nama_paket, r.nama_dokter, r.nama_ruang, r.status, r.dokteranastesi, r.perawat, r.sumber]) {
        row.insertCell().textContent = value || '-'
      }
    }
    doc.body.append(title, patient, table)
    popup.focus()
    popup.print()
  }
  function pesan(e: unknown) { return e instanceof Error ? e.message : 'Data belum dapat diproses.' }
  watch(() => [props.token, noRawat.value], () => {
    generasi++
    records.value = []
    saving.value = false
    loading.value = false
    pesanKunci.value = ''
    peringatan.value = ''
    hapusTarget.value = null
    Object.assign(filter, { q: '', status: '', mulai: '', selesai: '' })
    reset()
    void muat()
  }, { immediate: true })
  onBeforeUnmount(() => { generasi++ })
  return {
    loading, saving, error, errorSimpan, pesanKunci, peringatan, terkunci,
    formVisible, form, filter, paket, dokter, ruang, filteredRows, editing, hapusTarget,
    reset, muat, cari, edit, simpan, hapus, cetak, records,
  }
}
