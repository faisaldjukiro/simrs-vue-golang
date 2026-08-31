// @ts-nocheck -- migrasi TypeScript bertahap; kontrak data modul lama belum sepenuhnya bertipe.
import { computed, nextTick, reactive, ref, watch } from "vue"
import { hapusPenangananDokterPetugas, penangananDokterPetugasData, simpanBanyakPenangananDokterPetugas, ubahPenangananDokterPetugas } from "../../../lib/faisal/api"
import { useNotifikasi } from "../../../lib/shared/useNotifikasi"

export function usePenangananDokterPetugas(props) {
  const notifikasi = useNotifikasi()
  const loading = ref(false)
  const saving = ref(false)
  const deleting = ref(false)
  const error = ref('')
  const records = ref([])
  const billingLocked = ref(false)
  const formVisible = ref(true)
  const editingKey = ref(null)
  const deleteTarget = ref(null)
  const kataKunciTindakan = ref('')
  const doctor = ref({})
  const officer = ref({})
  const treatments = ref([])
  const form = reactive(emptyForm())
  const labelJenisRawat = computed(() => props.jenisRawat === 'ralan' ? props.namaModul : 'Rawat Inap')
  const tableRecords = computed(() => records.value.map((item) => ({
    ...item,
    _key: [item.no_rawat, item.kode_tindakan, item.kode_dokter, item.kode_petugas, item.tanggal, item.jam].join('|'),
  })))
  const tableRecordsTampil = computed(() => {
    const keyword = kataKunciTindakan.value.trim().toLowerCase()
    if (!keyword) return tableRecords.value
    return tableRecords.value.filter((item) => teksTindakan(item).includes(keyword))
  })
  
  function today() { const d = new Date(); return new Date(d.getTime() - d.getTimezoneOffset() * 60000).toISOString().slice(0, 10) }
  function now() { return new Date().toTimeString().slice(0, 8) }
  function emptyForm() { return { no_rawat: props.patient?.no_rawat || '', tanggal: today(), jam: now() } }
  function keyFor(item) { return { jenis_rawat: item.jenis_rawat || props.jenisRawat, no_rawat: item.no_rawat, kode_tindakan: item.kode_tindakan, kode_dokter: item.kode_dokter, kode_petugas: item.kode_petugas, tanggal: item.tanggal, jam: item.jam } }
  function rupiah(value) { return new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR', maximumFractionDigits: 0 }).format(Number(value || 0)) }
  function formatDate(value) { if (!value) return '-'; const [y, m, d] = value.split('-'); return `${d}/${m}/${y}` }
  function teksTindakan(item) {
    return [
      item.tanggal,
      item.jam,
      item.kode_tindakan,
      item.nama_tindakan,
      item.kode_dokter,
      item.nama_dokter,
      item.kode_petugas,
      item.nama_petugas,
      item.kelas,
      item.total,
    ].join(' ').toLowerCase()
  }
  
  function resetForm(defaults = true) {
    Object.assign(form, emptyForm())
    editingKey.value = null
    treatments.value = []
    if (!defaults) { doctor.value = {}; officer.value = {} }
  }
  
  async function loadRecords() {
    if (!props.patient.no_rawat) return
    loading.value = true; error.value = ''
    try {
      const data = await penangananDokterPetugasData(props.token, props.patient.no_rawat, props.jenisRawat)
      records.value = data?.catatan || []
      billingLocked.value = Boolean(data?.billing_terkunci)
      if (!editingKey.value) {
        doctor.value = data?.dokter_dpjp || {}
        officer.value = data?.petugas_login || {}
      }
    } catch (err) { error.value = err.message; notifikasi.gagal(err.message) }
    finally { loading.value = false }
  }
  
  function editRecord(item) {
    if (billingLocked.value) return
    editingKey.value = keyFor(item)
    Object.assign(form, { no_rawat: item.no_rawat, tanggal: item.tanggal, jam: item.jam })
    doctor.value = { kode: item.kode_dokter, nama: item.nama_dokter }
    officer.value = { kode: item.kode_petugas, nama: item.nama_petugas }
    treatments.value = [{ kode: item.kode_tindakan, nama: item.nama_tindakan, kelas: item.kelas, material: item.material, bhp: item.bhp, tarif_dokter: item.tarif_dokter, tarif_petugas: item.tarif_petugas, kso: item.kso, manajemen: item.manajemen, total: item.total }]
    formVisible.value = true
    nextTick(() => document.querySelector('.handling-form-card')?.scrollIntoView({ behavior: 'smooth', block: 'start' }))
  }
  
  function catatan(tindakan) { return { jenis_rawat: props.jenisRawat, no_rawat: form.no_rawat, kode_tindakan: tindakan.kode || '', kode_dokter: doctor.value.kode || '', kode_petugas: officer.value.kode || '', tanggal: form.tanggal, jam: form.jam } }
  function payload() { return { kunci_lama: editingKey.value, ...catatan(treatments.value[0] || {}) } }
  async function saveRecord() {
    if (!doctor.value.kode || !officer.value.kode || treatments.value.length === 0) { notifikasi.peringatan('Dokter, petugas, dan minimal satu tindakan wajib dipilih.'); return }
    saving.value = true
    try {
      const response = editingKey.value
        ? await ubahPenangananDokterPetugas(props.token, payload())
        : await simpanBanyakPenangananDokterPetugas(props.token, { catatan: treatments.value.map(catatan) })
      notifikasi.sukses(response?.pesan || 'Penanganan berhasil disimpan.')
      resetForm(false); await loadRecords(); formVisible.value = false
    } catch (err) { notifikasi.gagal(err.message) }
    finally { saving.value = false }
  }
  async function confirmDelete() {
    deleting.value = true
    try { const response = await hapusPenangananDokterPetugas(props.token, keyFor(deleteTarget.value)); notifikasi.sukses(response?.pesan || 'Data berhasil dihapus.'); deleteTarget.value = null; await loadRecords() }
    catch (err) { notifikasi.gagal(err.message) }
    finally { deleting.value = false }
  }
  
  watch([() => props.patient.no_rawat, () => props.jenisRawat], () => { kataKunciTindakan.value = ''; resetForm(false); loadRecords() }, { immediate: true })
  return {
    loading,
    saving,
    deleting,
    error,
    records,
    billingLocked,
    formVisible,
    editingKey,
    deleteTarget,
    kataKunciTindakan,
    doctor,
    officer,
    treatments,
    form,
    labelJenisRawat,
    tableRecordsTampil,
    rupiah,
    formatDate,
    resetForm,
    loadRecords,
    editRecord,
    saveRecord,
    confirmDelete,
  }
}
