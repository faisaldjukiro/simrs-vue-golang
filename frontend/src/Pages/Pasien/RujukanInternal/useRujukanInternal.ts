import { computed, nextTick, onBeforeUnmount, reactive, ref, watch } from 'vue'
import { cariReferensiRujukan, dataRujukanInternal, mutasiRujukanInternal, simpanRujukanInternal } from '../../../lib/faisal/rujukanInternal'
import { useNotifikasi } from '../../../lib/shared/useNotifikasi'
import type { PropsRujukanInternal, ReferensiRujukan, RujukanInternal } from '../../../types/rujukanInternal'

export function useRujukanInternal(props: PropsRujukanInternal) {
  const notifikasi = useNotifikasi()
  const ranap = computed(() => props.moduleName === 'Rawat Inap')
  const judul = computed(() => ranap.value ? 'Rujuk Internal Rawat Inap' : 'Rujuk Internal Poli')
  const modulValid = computed(() => ['Rawat Jalan', 'IGD/UGD', 'Rawat Inap'].includes(props.moduleName))
  const loading = ref(false)
  const saving = ref(false)
  const error = ref('')
  const errorSimpan = ref('')
  const pesanKunci = ref('')
  const bolehSimpan = ref(false)
  const formVisible = ref(true)
  const editing = ref<RujukanInternal | null>(null)
  const konfirmasi = ref<{ aksi: 'hapus' | 'kirim'; row: RujukanInternal } | null>(null)
  const errorAksi = ref('')
  const keyword = ref('')
  const records = ref<RujukanInternal[]>([])
  const poli = ref<ReferensiRujukan>({})
  const dokter = ref<ReferensiRujukan>({})
  const form = reactive({ tanggal: '', jam: '' })
  const terkunci = computed(() => loading.value || saving.value || !bolehSimpan.value || !modulValid.value)
  const filteredRows = computed(() => {
    const kata = keyword.value.trim().toLocaleLowerCase('id')
    return records.value.map(row => ({ ...row, _key: `${row.sumber}-${row.id || (ranap.value ? row.kd_poli : row.kd_dokter)}` })).filter(row =>
      [row.kd_poli, row.nama_poli, row.kd_dokter, row.nama_dokter, row.tanggal, row.jam, row.sumber]
        .some(value => value.toLocaleLowerCase('id').includes(kata)),
    )
  })
  let generasi = 0
  let requestID = 0
  let controller: AbortController | undefined

  function waktuSekarang() {
    const parts = new Intl.DateTimeFormat('sv-SE', {
      timeZone: 'Asia/Makassar', year: 'numeric', month: '2-digit', day: '2-digit',
      hour: '2-digit', minute: '2-digit', second: '2-digit', hourCycle: 'h23',
    }).formatToParts(new Date())
    const bagian = (type: Intl.DateTimeFormatPartTypes) => parts.find(part => part.type === type)?.value || ''
    form.tanggal = `${bagian('year')}-${bagian('month')}-${bagian('day')}`
    form.jam = `${bagian('hour')}:${bagian('minute')}:${bagian('second')}`
  }

  function resetForm() {
    editing.value = null
    poli.value = {}
    dokter.value = {}
    errorSimpan.value = ''
    waktuSekarang()
  }

  async function edit(row: RujukanInternal) {
    if (terkunci.value) return
    editing.value = { ...row }
    poli.value = { kode: row.kd_poli, nama: row.nama_poli }
    dokter.value = { kode: row.kd_dokter, nama: row.nama_dokter }
    form.tanggal = row.tanggal
    form.jam = row.jam
    errorSimpan.value = ''
    formVisible.value = true
    await nextTick()
    document.querySelector('.rujukan-form')?.scrollIntoView({ block: 'nearest' })
  }

  function mintaKonfirmasi(aksi: 'hapus' | 'kirim', row: RujukanInternal) {
    if (terkunci.value) return
    if (aksi === 'kirim' && (row.sumber !== 'SIRAPI' || row.konflik_khanza)) return
    errorAksi.value = ''
    konfirmasi.value = { aksi, row: { ...row } }
  }

  async function jalankanAksi() {
    if (!konfirmasi.value || terkunci.value) return
    const target = konfirmasi.value
    const konteks = generasi
    errorAksi.value = ''
    saving.value = true
    try {
      const response = await mutasiRujukanInternal(props.token, ranap.value, target.aksi, target.row)
      if (konteks !== generasi) return
      if (response.peringatan) notifikasi.peringatan(response.pesan)
      else notifikasi.sukses(response.pesan)
      konfirmasi.value = null
      resetForm()
      await muat()
    } catch (err) {
      if (konteks === generasi) errorAksi.value = err instanceof Error ? err.message : 'Rujukan gagal diproses.'
    } finally {
      if (konteks === generasi) saving.value = false
    }
  }

  async function muat() {
    controller?.abort()
    const request = ++requestID
    records.value = []
    bolehSimpan.value = false
    pesanKunci.value = ''
    error.value = ''
    if (!props.patient.no_rawat || !modulValid.value) {
      error.value = 'Pilih kunjungan Rawat Jalan, IGD, atau Rawat Inap terlebih dahulu.'
      loading.value = false
      return
    }
    controller = new AbortController()
    loading.value = true
    try {
      const response = await dataRujukanInternal(props.token, ranap.value, props.patient.no_rawat, controller.signal)
      if (request !== requestID) return
      records.value = response.daftar
      bolehSimpan.value = response.boleh_simpan
      pesanKunci.value = response.pesan
    } catch (err) {
      if (request !== requestID) return
      error.value = err instanceof Error ? err.message : 'Riwayat rujukan gagal dimuat.'
    } finally {
      if (request === requestID) loading.value = false
    }
  }

  async function cariReferensi(jenis: 'poli' | 'dokter', kata: string) {
    if (terkunci.value) return []
    const konteks = generasi
    try {
      const hasil = await cariReferensiRujukan(props.token, ranap.value, jenis, kata)
      return konteks === generasi ? hasil : []
    } catch (err) {
      if (konteks === generasi) notifikasi.gagal(err instanceof Error ? err.message : 'Referensi gagal dimuat.')
      return []
    }
  }

  async function simpan() {
    if (terkunci.value) return
    errorSimpan.value = ''
    if (!poli.value.kode || !dokter.value.kode) {
      errorSimpan.value = 'Pilih poli dan dokter tujuan dari hasil pencarian.'
      return
    }
    const konteks = generasi
    saving.value = true
    try {
      const input = {
        no_rawat: props.patient.no_rawat,
        kd_poli: poli.value.kode,
        kd_dokter: dokter.value.kode,
        tanggal: ranap.value ? form.tanggal : '',
        jam: ranap.value ? form.jam : '',
      }
      const response = editing.value
        ? await mutasiRujukanInternal(props.token, ranap.value, 'ubah', editing.value, input)
        : await simpanRujukanInternal(props.token, ranap.value, input)
      if (konteks !== generasi) return
      notifikasi.sukses(response.pesan)
      resetForm()
      await muat()
    } catch (err) {
      if (konteks === generasi) errorSimpan.value = err instanceof Error ? err.message : 'Rujukan gagal disimpan.'
    } finally {
      if (konteks === generasi) saving.value = false
    }
  }

  watch(() => [props.patient.no_rawat, props.moduleName, props.token], () => {
    generasi++
    saving.value = false
    keyword.value = ''
    formVisible.value = true
    konfirmasi.value = null
    errorAksi.value = ''
    resetForm()
    void muat()
  }, { immediate: true })

  onBeforeUnmount(() => {
    generasi++
    requestID++
    controller?.abort()
  })

  return {
    ranap, judul, loading, saving, error, errorSimpan, pesanKunci, formVisible,
    keyword, records, poli, dokter, form, terkunci, filteredRows,
    waktuSekarang, resetForm, muat, cariReferensi, simpan,
    editing, edit, konfirmasi, errorAksi, mintaKonfirmasi, jalankanAksi,
  }
}
