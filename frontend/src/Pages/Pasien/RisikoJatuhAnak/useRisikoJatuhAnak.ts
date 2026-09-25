import { computed, onBeforeUnmount, reactive, ref, watch } from 'vue'
import { request } from '../../../lib/shared/http'
import { useNotifikasi } from '../../../lib/shared/useNotifikasi'
import type { CatatanRisikoJatuh, HasilRisikoJatuh, PropsRisikoJatuh, SkalaHumptyDumpty } from '../../../types/risikoJatuhAnak'

export function useRisikoJatuhAnak(props: PropsRisikoJatuh) {
  const notifikasi = useNotifikasi()
  const loading = ref(false)
  const saving = ref(false)
  const error = ref('')
  const errorSimpan = ref('')
  const keyword = ref('')
  const formVisible = ref(true)
  const skala = ref<SkalaHumptyDumpty[]>([])
  const records = ref<CatatanRisikoJatuh[]>([])
  const editing = ref<CatatanRisikoJatuh | null>(null)
  const detail = ref<CatatanRisikoJatuh | null>(null)
  const hapusTarget = ref<CatatanRisikoJatuh | null>(null)
  const petugas = ref<Record<string, string>>({})
  const form = reactive<Record<string, string>>({})
  let generasi = 0
  let urutan = 0
  const nilai = computed(() => skala.value.map((s, i) => {
    const index = s.pilihan.indexOf(form['penilaian_humptydumpty_skala' + (i + 1)])
    return index < 0 ? null : s.nilai[index]
  }))
  const total = computed(() => nilai.value.length === 7 && nilai.value.every(n => n != null)
    ? nilai.value.reduce<number>((jumlah, n) => jumlah + (n ?? 0), 0)
    : null)
  const risiko = computed(() => total.value === null ? 'Lengkapi tujuh penilaian'
    : total.value < 12 ? 'Rendah' : 'Tinggi')
  const rows = computed(() => records.value.filter(r =>
    [r.nama_petugas, r.risiko, ...Object.values(r.data)].join(' ').toLocaleLowerCase()
      .includes(keyword.value.trim().toLocaleLowerCase()),
  ).map(r => ({ ...r, kunci: r.data.tanggal })))
  const terkunci = computed(() => loading.value || saving.value || !!error.value || skala.value.length !== 7)

  function api<T>(path = '', options: RequestInit = {}) {
    return request<T>('/api/risiko-jatuh-anak' + path, {
      ...options,
      headers: { Authorization: 'Bearer ' + props.token },
    })
  }

  function waktuSekarang() {
    form.tanggal = new Date(Date.now() + 8 * 3600000).toISOString().slice(0, 19)
  }

  function reset() {
    editing.value = null
    errorSimpan.value = ''
    petugas.value = {}
    Object.keys(form).forEach(k => delete form[k])
    for (let i = 1; i <= 7; i++) form['penilaian_humptydumpty_skala' + i] = ''
    form.hasil_skrining = ''
    form.saran = ''
    waktuSekarang()
  }

  async function muat() {
    const id = ++urutan
    const konteks = generasi
    loading.value = true
    error.value = ''
    try {
      const hasil = await api<HasilRisikoJatuh>('?' + new URLSearchParams({ no_rawat: props.patient.no_rawat }))
      if (konteks !== generasi || id !== urutan) return
      skala.value = hasil.skala
      records.value = hasil.catatan
    } catch (e) {
      if (konteks === generasi && id === urutan) error.value = e instanceof Error ? e.message : 'Gagal memuat penilaian.'
    } finally {
      if (konteks === generasi && id === urutan) loading.value = false
    }
  }

  function cariPetugas(q: string) {
    return request<Record<string, string>[]>('/api/checklist-pre-operasi/referensi?' + new URLSearchParams({ jenis: 'petugas', q }), {
      headers: { Authorization: 'Bearer ' + props.token },
    })
  }

  function edit(row: CatatanRisikoJatuh) {
    if (terkunci.value || !row.bisa_ubah) return
    reset()
    editing.value = { ...row, data: { ...row.data } }
    formVisible.value = true
    Object.assign(form, row.data)
    form.tanggal = row.data.tanggal.replace(' ', 'T')
    petugas.value = { kode: row.data.nip, nama: row.nama_petugas }
  }

  async function mutasi(hapus = false) {
    if (terkunci.value || (hapus && !hapusTarget.value)) return
    errorSimpan.value = ''
    if (!hapus && (total.value === null || !petugas.value.kode || !form.tanggal || !form.hasil_skrining.trim() || !form.saran.trim())) {
      errorSimpan.value = 'Lengkapi tanggal, petugas, tujuh penilaian, hasil skrining, dan saran.'
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
      if (konteks === generasi) errorSimpan.value = e instanceof Error ? e.message : 'Penilaian gagal disimpan.'
    } finally {
      if (konteks === generasi) saving.value = false
    }
  }

  watch(() => [props.token, props.patient.no_rawat], () => {
    generasi++
    saving.value = false
    records.value = []
    skala.value = []
    detail.value = null
    hapusTarget.value = null
    keyword.value = ''
    formVisible.value = true
    reset()
    void muat()
  }, { immediate: true })
  onBeforeUnmount(() => { generasi++ })

  return {
    loading, saving, error, errorSimpan, keyword, skala, rows, records, formVisible, editing, detail,
    hapusTarget, petugas, form, nilai, total, risiko, terkunci,
    waktuSekarang, reset, muat, cariPetugas, edit, mutasi,
  }
}
