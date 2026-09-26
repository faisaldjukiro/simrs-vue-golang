import { computed, onBeforeUnmount, reactive, ref, watch } from 'vue'
import { request } from '../../../lib/shared/http'
import { useNotifikasi } from '../../../lib/shared/useNotifikasi'
import { bidangHais } from '../../../types/dataHais'
import type { CatatanHais, HasilHais, PropsHais } from '../../../types/dataHais'

export function useDataHais(props: PropsHais) {
  const notifikasi = useNotifikasi()
  const loading = ref(false)
  const saving = ref(false)
  const error = ref('')
  const errorSimpan = ref('')
  const keyword = ref('')
  const mulai = ref('')
  const selesai = ref('')
  const formVisible = ref(true)
  const records = ref<CatatanHais[]>([])
  const editing = ref<CatatanHais | null>(null)
  const detail = ref<CatatanHais | null>(null)
  const hapusTarget = ref<CatatanHais | null>(null)
  const kamar = ref('')
  const form = reactive<Record<string, string>>({})
  let generasi = 0
  let urutan = 0
  const errorFilter = computed(() => mulai.value && selesai.value && mulai.value > selesai.value
    ? 'Tanggal mulai tidak boleh melewati tanggal selesai.' : '')
  const rows = computed(() => records.value.filter(r =>
    !errorFilter.value &&
    (!mulai.value || r.data.tanggal >= mulai.value) &&
    (!selesai.value || r.data.tanggal <= selesai.value) &&
    [r.sumber, ...Object.values(r.data)].join(' ').toLocaleLowerCase()
      .includes(keyword.value.trim().toLocaleLowerCase()),
  ).map(r => ({ ...r, kunci: r.sumber + ' ' + r.data.tanggal })))
  const terkunci = computed(() => loading.value || saving.value || !!error.value)

  function api<T>(path = '', options: RequestInit = {}) {
    return request<T>('/api/data-hais' + path, {
      ...options,
      headers: { Authorization: 'Bearer ' + props.token },
    })
  }

  function waktuSekarang() {
    const sekarang = new Date(Date.now() + 8 * 3600000).toISOString()
    form.tanggal = sekarang.slice(0, 10)
  }

  function reset() {
    editing.value = null
    errorSimpan.value = ''

    Object.keys(form).forEach(k => delete form[k])
    for (const b of bidangHais) form[b.key] = b.angka ? '0' : ''
    form.DEKU = 'TIDAK'
    form.kd_kamar = kamar.value
    waktuSekarang()
  }

  async function muat() {
    const id = ++urutan
    const konteks = generasi
    loading.value = true
    error.value = ''
    try {
      const hasil = await api<HasilHais>('?' + new URLSearchParams({ no_rawat: props.patient.no_rawat }))
      if (konteks !== generasi || id !== urutan) return
      records.value = hasil.catatan
      kamar.value = hasil.kamar
      if (!editing.value) form.kd_kamar = hasil.kamar
    } catch (e) {
      if (konteks === generasi && id === urutan) error.value = e instanceof Error ? e.message : 'Gagal memuat HAIs.'
    } finally {
      if (konteks === generasi && id === urutan) loading.value = false
    }
  }

  function cetak() {
    if (loading.value || error.value || !rows.value.length) return
    const popup = window.open('', '_blank', 'width=1100,height=750')
    if (!popup) {
      notifikasi.peringatan('Izinkan jendela cetak pada browser terlebih dahulu.')
      return
    }
    const doc = popup.document
    doc.title = 'Data HAIs'
    doc.documentElement.lang = 'id'
    const style = doc.createElement('style')
    style.textContent = '@page { size: A4 landscape; margin: 12mm; } body { font: 12px Arial, sans-serif; color: #111; } h1 { font-size: 20px; } table { width: 100%; border-collapse: collapse; } th, td { border: 1px solid #aaa; padding: 8px; text-align: left; overflow-wrap: anywhere; } th { background: #eee; } thead { display: table-header-group; } tr { break-inside: avoid; }'
    doc.head.append(style)
    const title = doc.createElement('h1')
    title.textContent = 'Data HAIs'
    const identity = doc.createElement('p')
    identity.textContent = `${props.patient.nm_pasien || '-'} | RM: ${props.patient.no_rkm_medis || '-'} | No. Rawat: ${props.patient.no_rawat}`
    const period = doc.createElement('p')
    period.textContent = `Periode: ${mulai.value || 'Awal kunjungan'} s.d. ${selesai.value || 'Terakhir'} | Pencarian: ${keyword.value || '-'} | ${rows.value.length} catatan | Waktu WITA`
    const table = doc.createElement('table')
    const head = table.createTHead().insertRow()
    for (const label of ['Tanggal', ...bidangHais.map(b => b.label), 'Dekubitus', 'Kamar', 'Sumber']) {
      const th = doc.createElement('th')
      th.textContent = label
      head.append(th)
    }
    const body = table.createTBody()
    for (const r of rows.value) {
      const tr = body.insertRow()
      for (const value of [r.data.tanggal, ...bidangHais.map(b => r.data[b.key]), r.data.DEKU, r.data.kd_kamar, r.sumber]) {
        tr.insertCell().textContent = value || '-'
      }
    }
    doc.body.append(title, identity, period, table)
    popup.focus()
    popup.requestAnimationFrame(() => popup.print())
  }

  function edit(row: CatatanHais) {
    if (terkunci.value || !row.bisa_ubah) return
    reset()
    editing.value = { ...row, data: { ...row.data } }
    formVisible.value = true
    Object.assign(form, row.data)
  }

  async function mutasi(hapus = false) {
    if (terkunci.value || (hapus && !hapusTarget.value)) return
    errorSimpan.value = ''
    if (!hapus && (!form.kd_kamar || !form.tanggal)) {
      errorSimpan.value = 'Tanggal dan kamar rawat inap wajib tersedia.'
      return
    }
    const konteks = generasi
    saving.value = true
    try {
      const hasil = await api<{ pesan: string }>('', {
        method: hapus ? 'DELETE' : editing.value ? 'PUT' : 'POST',
        body: JSON.stringify({
          no_rawat: props.patient.no_rawat,
          data: hapus ? undefined : { ...form },
          asli: hapus ? hapusTarget.value?.data : editing.value?.data,
        }),
      })
      if (konteks !== generasi) return
      hapusTarget.value = null
      reset()
      notifikasi.sukses(hasil.pesan)
      await muat()
    } catch (e) {
      if (konteks === generasi) errorSimpan.value = e instanceof Error ? e.message : 'Hais gagal disimpan.'
    } finally {
      if (konteks === generasi) saving.value = false
    }
  }

  watch(() => [props.token, props.patient.no_rawat], () => {
    generasi++
    saving.value = false
    records.value = []
    kamar.value = ''
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
    hapusTarget, kamar, form, terkunci,
    mulai, selesai, errorFilter, cetak,
    waktuSekarang, reset, muat, edit, mutasi,
  }
}
