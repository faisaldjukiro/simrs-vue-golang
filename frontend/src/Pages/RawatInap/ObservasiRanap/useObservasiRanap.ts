import { computed, onBeforeUnmount, reactive, ref, watch } from 'vue'
import { request } from '../../../lib/shared/http'
import { useNotifikasi } from '../../../lib/shared/useNotifikasi'
import { bidangObservasi } from '../../../types/observasiRanap'
import type { CatatanObservasi, HasilObservasi, PropsObservasi } from '../../../types/observasiRanap'

export function useObservasiRanap(props: PropsObservasi) {
  const notifikasi = useNotifikasi()
  const loading = ref(false)
  const saving = ref(false)
  const error = ref('')
  const errorSimpan = ref('')
  const keyword = ref('')
  const mulai = ref('')
  const selesai = ref('')
  const formVisible = ref(true)
  const records = ref<CatatanObservasi[]>([])
  const editing = ref<CatatanObservasi | null>(null)
  const detail = ref<CatatanObservasi | null>(null)
  const hapusTarget = ref<CatatanObservasi | null>(null)
  const petugas = ref<Record<string, string>>({})
  const form = reactive<Record<string, string>>({})
  let generasi = 0
  let urutan = 0
  const errorFilter = computed(() => mulai.value && selesai.value && mulai.value > selesai.value
    ? 'Tanggal mulai tidak boleh melewati tanggal selesai.' : '')
  const rows = computed(() => records.value.filter(r =>
    !errorFilter.value &&
    (!mulai.value || r.data.tgl_perawatan >= mulai.value) &&
    (!selesai.value || r.data.tgl_perawatan <= selesai.value) &&
    [r.nama_petugas, ...Object.values(r.data)].join(' ').toLocaleLowerCase()
      .includes(keyword.value.trim().toLocaleLowerCase()),
  ).map(r => ({ ...r, kunci: r.data.tgl_perawatan + ' ' + r.data.jam_rawat })))
  const terkunci = computed(() => loading.value || saving.value || !!error.value)

  function api<T>(path = '', options: RequestInit = {}) {
    return request<T>('/api/observasi-ranap' + path, {
      ...options,
      headers: { Authorization: 'Bearer ' + props.token },
    })
  }

  function waktuSekarang() {
    const sekarang = new Date(Date.now() + 8 * 3600000).toISOString()
    form.tgl_perawatan = sekarang.slice(0, 10)
    form.jam_rawat = sekarang.slice(11, 19)
  }

  function reset() {
    editing.value = null
    errorSimpan.value = ''
    petugas.value = {}
    Object.keys(form).forEach(k => delete form[k])
    for (const key of ['gcs', 'td', 'hr', 'rr', 'suhu', 'spo2']) form[key] = ''
    waktuSekarang()
  }

  async function muat() {
    const id = ++urutan
    const konteks = generasi
    loading.value = true
    error.value = ''
    try {
      const hasil = await api<HasilObservasi>('?' + new URLSearchParams({ no_rawat: props.patient.no_rawat }))
      if (konteks !== generasi || id !== urutan) return
      records.value = hasil.catatan
    } catch (e) {
      if (konteks === generasi && id === urutan) error.value = e instanceof Error ? e.message : 'Gagal memuat observasi.'
    } finally {
      if (konteks === generasi && id === urutan) loading.value = false
    }
  }

  function cariPetugas(q: string) {
    return request<Record<string, string>[]>('/api/checklist-pre-operasi/referensi?' + new URLSearchParams({ jenis: 'petugas', q }), {
      headers: { Authorization: 'Bearer ' + props.token },
    })
  }

  function cetak() {
    if (loading.value || error.value || !rows.value.length) return
    const popup = window.open('', '_blank', 'width=1100,height=750')
    if (!popup) {
      notifikasi.peringatan('Izinkan jendela cetak pada browser terlebih dahulu.')
      return
    }
    const doc = popup.document
    doc.title = 'Catatan Observasi Rawat Inap'
    doc.documentElement.lang = 'id'
    const style = doc.createElement('style')
    style.textContent = '@page { size: A4 landscape; margin: 12mm; } body { font: 12px Arial, sans-serif; color: #111; } h1 { font-size: 20px; } table { width: 100%; border-collapse: collapse; } th, td { border: 1px solid #aaa; padding: 8px; text-align: left; overflow-wrap: anywhere; } th { background: #eee; } thead { display: table-header-group; } tr { break-inside: avoid; }'
    doc.head.append(style)
    const title = doc.createElement('h1')
    title.textContent = 'Catatan Observasi Rawat Inap'
    const identity = doc.createElement('p')
    identity.textContent = `${props.patient.nm_pasien || '-'} | RM: ${props.patient.no_rkm_medis || '-'} | No. Rawat: ${props.patient.no_rawat}`
    const period = doc.createElement('p')
    period.textContent = `Periode: ${mulai.value || 'Awal kunjungan'} s.d. ${selesai.value || 'Terakhir'} | Pencarian: ${keyword.value || '-'} | ${rows.value.length} catatan | Waktu WITA`
    const table = doc.createElement('table')
    const head = table.createTHead().insertRow()
    for (const label of ['Tanggal', 'Jam', ...bidangObservasi.map(b => b.label), 'Petugas']) {
      const th = doc.createElement('th')
      th.textContent = label
      head.append(th)
    }
    const body = table.createTBody()
    for (const r of rows.value) {
      const tr = body.insertRow()
      for (const value of [r.data.tgl_perawatan, r.data.jam_rawat, ...bidangObservasi.map(b => r.data[b.key]), `${r.nama_petugas || '-'} (${r.data.nip})`]) {
        tr.insertCell().textContent = value || '-'
      }
    }
    doc.body.append(title, identity, period, table)
    popup.focus()
    popup.requestAnimationFrame(() => popup.print())
  }

  function edit(row: CatatanObservasi) {
    if (terkunci.value || !row.bisa_ubah) return
    reset()
    editing.value = { ...row, data: { ...row.data } }
    formVisible.value = true
    Object.assign(form, row.data)
    petugas.value = { kode: row.data.nip, nama: row.nama_petugas }
  }

  async function mutasi(hapus = false) {
    if (terkunci.value || (hapus && !hapusTarget.value)) return
    errorSimpan.value = ''
    if (!hapus && (!petugas.value.kode || !form.tgl_perawatan || !form.jam_rawat)) {
      errorSimpan.value = 'Lengkapi tanggal, jam, dan petugas.'
      return
    }
    const konteks = generasi
    saving.value = true
    try {
      const hasil = await api<{ pesan: string }>('', {
        method: hapus ? 'DELETE' : editing.value ? 'PUT' : 'POST',
        body: JSON.stringify({
          no_rawat: props.patient.no_rawat,
          data: hapus ? undefined : { ...form, nip: petugas.value.kode },
          asli: hapus ? hapusTarget.value?.data : editing.value?.data,
        }),
      })
      if (konteks !== generasi) return
      hapusTarget.value = null
      reset()
      notifikasi.sukses(hasil.pesan)
      await muat()
    } catch (e) {
      if (konteks === generasi) errorSimpan.value = e instanceof Error ? e.message : 'Observasi gagal disimpan.'
    } finally {
      if (konteks === generasi) saving.value = false
    }
  }

  watch(() => [props.token, props.patient.no_rawat], () => {
    generasi++
    saving.value = false
    records.value = []
    detail.value = null
    hapusTarget.value = null
    keyword.value = ''
    mulai.value = ''
    selesai.value = ''
    formVisible.value = true
    reset()
    void muat()
  }, { immediate: true })
  onBeforeUnmount(() => { generasi++ })

  return {
    loading, saving, error, errorSimpan, keyword, rows, records, formVisible, editing, detail,
    hapusTarget, petugas, form, terkunci,
    mulai, selesai, errorFilter, cetak,
    waktuSekarang, reset, muat, cariPetugas, edit, mutasi,
  }
}
