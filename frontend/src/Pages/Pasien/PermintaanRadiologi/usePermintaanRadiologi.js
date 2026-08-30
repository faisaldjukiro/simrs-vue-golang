import { computed, nextTick, reactive, ref, watch } from "vue"
import { hapusPermintaanRadiologi, permintaanRadiologiData, simpanPermintaanRadiologi, ubahPermintaanRadiologi } from "../../../lib/faisal/api"
import { useNotifikasi } from "../../../lib/shared/useNotifikasi"

export function usePermintaanRadiologi(props) {
  const notifikasi = useNotifikasi()
  const loading = ref(false)
  const saving = ref(false)
  const deleting = ref(false)
  const error = ref('')
  const formVisible = ref(true)
  const requests = ref([])
  const doctor = ref({})
  const defaultDoctor = ref({})
  const treatments = ref([])
  const billingLocked = ref(false)
  const scope = ref({})
  const deleteTarget = ref(null)
  const editingNumber = ref('')
  const kataKunciPermintaan = ref('')
  const form = reactive(emptyForm())
  const tableRows = computed(() => requests.value.map((item) => ({ ...item, _key: item.nomor })))
  const tableRowsTampil = computed(() => {
    const keyword = kataKunciPermintaan.value.trim().toLowerCase()
    if (!keyword) return tableRows.value
    return tableRows.value.filter((item) => teksPermintaan(item).includes(keyword))
  })
  
  function today() {
    const date = new Date()
    return new Date(date.getTime() - date.getTimezoneOffset() * 60000).toISOString().slice(0, 10)
  }
  function now() { return new Date().toTimeString().slice(0, 8) }
  function emptyForm() {
    return {
      no_rawat: props.patient?.no_rawat || '',
      tanggal: today(),
      jam: now(),
      informasi_tambahan: '',
      diagnosis_klinis: '',
    }
  }
  function rupiah(value) {
    return new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR', maximumFractionDigits: 0 }).format(Number(value || 0))
  }
  function formatDate(value) {
    if (!value) return '-'
    const [year, month, day] = value.split('-')
    return `${day}/${month}/${year}`
  }
  function teksPermintaan(item) {
    const pemeriksaan = (item.pemeriksaan || []).flatMap((pemeriksaan) => [
      pemeriksaan.kode,
      pemeriksaan.nama,
      pemeriksaan.kelas,
      pemeriksaan.total,
    ])
    return [
      item.nomor,
      item.tanggal,
      item.jam,
      item.kode_dokter,
      item.nama_dokter,
      item.informasi_tambahan,
      item.diagnosis_klinis,
      item.status_pemeriksaan,
      item.status_bayar,
      item.total,
      ...pemeriksaan,
    ].join(' ').toLowerCase()
  }
  function waktuStatus(item) {
    if (item.status_pemeriksaan === 'Selesai') {
      return `${formatDate(item.tanggal_hasil)} · ${item.jam_hasil || '-'}`
    }
    if (item.status_pemeriksaan === 'Sedang Dikerjakan') {
      return `${formatDate(item.tanggal_diterima)} · ${item.jam_diterima || '-'}`
    }
    return 'Belum diterima petugas'
  }
  function kelasStatusPemeriksaan(status) {
    if (status === 'Selesai') return 'completed'
    if (status === 'Sedang Dikerjakan') return 'processing'
    return 'waiting'
  }
  function kelasStatusBayar(status) {
    if (status === 'Sudah Bayar') return 'paid'
    if (status === 'Sebagian Dibayar') return 'partial'
    return 'unpaid'
  }
  function resetForm() {
    Object.assign(form, emptyForm())
    doctor.value = { ...defaultDoctor.value }
    treatments.value = []
    editingNumber.value = ''
  }
  
  function editRequest(item) {
    if (billingLocked.value || !item.dapat_diubah) return
    editingNumber.value = item.nomor
    Object.assign(form, {
      no_rawat: item.no_rawat,
      tanggal: item.tanggal,
      jam: item.jam,
      informasi_tambahan: item.informasi_tambahan,
      diagnosis_klinis: item.diagnosis_klinis,
    })
    doctor.value = { kode: item.kode_dokter, nama: item.nama_dokter }
    treatments.value = (item.pemeriksaan || []).map((tindakan) => ({
      kode: tindakan.kode,
      nama: tindakan.nama,
      kode_cara_bayar: tindakan.kode_cara_bayar,
      kelas: tindakan.kelas,
      total: tindakan.total,
    }))
    formVisible.value = true
    nextTick(() => document.querySelector('.radiology-form-card')?.scrollIntoView({ behavior: 'smooth', block: 'start' }))
  }
  
  async function loadData() {
    if (!props.patient.no_rawat) return
    loading.value = true
    error.value = ''
    try {
      const data = await permintaanRadiologiData(props.token, props.patient.no_rawat)
      requests.value = data?.permintaan || []
      billingLocked.value = Boolean(data?.billing_terkunci)
      defaultDoctor.value = data?.dokter_perujuk || {}
      scope.value = {
        status: data?.status_rawat || '',
        kodeCaraBayar: data?.kode_cara_bayar || '',
        kelas: data?.kelas_pasien || '',
        filterCaraBayar: Boolean(data?.filter_cara_bayar),
        filterKelas: Boolean(data?.filter_kelas),
      }
      if (!doctor.value?.kode) doctor.value = { ...defaultDoctor.value }
    } catch (err) {
      error.value = err.message
      notifikasi.gagal(err.message)
    } finally {
      loading.value = false
    }
  }
  
  async function saveRequest() {
    if (!doctor.value?.kode || treatments.value.length === 0) {
      notifikasi.peringatan('Dokter perujuk dan minimal satu tindakan radiologi wajib dipilih.')
      return
    }
    saving.value = true
    try {
      const payload = {
        ...form,
        kode_dokter: doctor.value.kode,
        kode_tindakan: treatments.value.map((item) => item.kode),
      }
      const response = editingNumber.value
        ? await ubahPermintaanRadiologi(props.token, editingNumber.value, payload)
        : await simpanPermintaanRadiologi(props.token, payload)
      notifikasi.sukses(`${response?.pesan || 'Permintaan radiologi berhasil disimpan.'} Nomor: ${response?.nomor || '-'}`)
      resetForm()
      await loadData()
      formVisible.value = false
    } catch (err) {
      notifikasi.gagal(err.message)
    } finally {
      saving.value = false
    }
  }
  
  async function confirmDelete() {
    if (!deleteTarget.value) return
    deleting.value = true
    try {
      const response = await hapusPermintaanRadiologi(props.token, props.patient.no_rawat, deleteTarget.value.nomor)
      notifikasi.sukses(response?.pesan || 'Permintaan radiologi berhasil dihapus.')
      deleteTarget.value = null
      await loadData()
    } catch (err) {
      notifikasi.gagal(err.message)
    } finally {
      deleting.value = false
    }
  }
  
  watch(() => props.patient.no_rawat, () => {
    kataKunciPermintaan.value = ''
    resetForm()
    loadData()
  }, { immediate: true })
  return {
    loading,
    saving,
    deleting,
    error,
    formVisible,
    requests,
    doctor,
    treatments,
    billingLocked,
    scope,
    deleteTarget,
    editingNumber,
    kataKunciPermintaan,
    form,
    tableRowsTampil,
    rupiah,
    formatDate,
    waktuStatus,
    kelasStatusPemeriksaan,
    kelasStatusBayar,
    resetForm,
    editRequest,
    loadData,
    saveRequest,
    confirmDelete,
  }
}
