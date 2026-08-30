import { computed, reactive, ref, watch } from "vue"
import { berkasDigitalData, hapusBerkasDigital, uploadBerkasDigital } from "../../../lib/faisal/api"
import { useNotifikasi } from "../../../lib/shared/useNotifikasi"

export function useBerkasDigital(props) {
  const notifikasi = useNotifikasi()
  const loading = ref(false)
  const saving = ref(false)
  const deleting = ref(false)
  const formVisible = ref(true)
  const fileInput = ref(null)
  const deleteTarget = ref(null)
  const previewTarget = ref(null)
  const kataKunciBerkas = ref('')
  const error = ref('')
  const master = ref([])
  const berkas = ref([])
  const billingLocked = ref(false)
  const form = reactive({ kode: '', file: null })
  
  const opsiMaster = computed(() => master.value.map((item) => ({ label: `${item.kode} - ${item.nama}`, value: item.kode })))
  const namaFile = computed(() => form.file?.name || '')
  const rows = computed(() => berkas.value.map((item) => ({
    ...item,
    _key: `${item.no_rawat}|${item.kode}|${item.lokasi_file}`,
  })))
  const rowsTampil = computed(() => {
    const keyword = kataKunciBerkas.value.trim().toLowerCase()
    if (!keyword) return rows.value
    return rows.value.filter((item) => [
      item.kode,
      item.nama,
      item.lokasi_file,
    ].some((value) => String(value || '').toLowerCase().includes(keyword)))
  })
  const previewUrl = computed(() => previewTarget.value?.url || '')
  const previewNama = computed(() => previewTarget.value ? namaBerkas(previewTarget.value) : '')
  const previewEkstensi = computed(() => String(previewTarget.value?.lokasi_file || previewTarget.value?.url || '').toLowerCase())
  const previewIsImage = computed(() => /\.(png|jpe?g|gif|webp)(\?|$)/i.test(previewEkstensi.value))
  const previewIsPdf = computed(() => /\.pdf(\?|$)/i.test(previewEkstensi.value))
  
  function resetForm() {
    form.kode = ''
    form.file = null
    if (fileInput.value) fileInput.value.value = ''
  }
  
  function onFileChange(event) {
    form.file = event.target.files?.[0] || null
  }
  
  function namaBerkas(item) {
    return item.nama || item.kode || 'Berkas Digital'
  }
  
  function openPreview(item) {
    if (!item?.url) {
      notifikasi.peringatan('URL file belum tersedia.')
      return
    }
    previewTarget.value = item
  }
  
  async function loadData() {
    if (!props.patient.no_rawat) return
    loading.value = true
    error.value = ''
    try {
      const data = await berkasDigitalData(props.token, props.patient.no_rawat)
      master.value = data?.master || []
      berkas.value = data?.berkas || []
      billingLocked.value = Boolean(data?.billing_terkunci)
    } catch (err) {
      error.value = err.message
      notifikasi.gagal(err.message)
    } finally {
      loading.value = false
    }
  }
  
  async function upload() {
    if (!form.kode || !form.file) {
      notifikasi.peringatan('Jenis berkas dan file wajib dipilih.')
      return
    }
    saving.value = true
    try {
      const payload = new FormData()
      payload.append('no_rawat', props.patient.no_rawat)
      payload.append('kode', form.kode)
      payload.append('file', form.file)
      const response = await uploadBerkasDigital(props.token, payload)
      notifikasi.sukses(response?.pesan || 'Berkas digital berhasil diupload.')
      resetForm()
      await loadData()
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
      const response = await hapusBerkasDigital(props.token, {
        no_rawat: deleteTarget.value.no_rawat,
        kode: deleteTarget.value.kode,
        lokasi_file: deleteTarget.value.lokasi_file,
      })
      notifikasi.sukses(response?.pesan || 'Berkas digital berhasil dihapus.')
      deleteTarget.value = null
      await loadData()
    } catch (err) {
      notifikasi.gagal(err.message)
    } finally {
      deleting.value = false
    }
  }
  
  watch(() => props.patient.no_rawat, () => {
    resetForm()
    kataKunciBerkas.value = ''
    loadData()
  }, { immediate: true })
  return {
    loading,
    saving,
    deleting,
    formVisible,
    fileInput,
    deleteTarget,
    previewTarget,
    kataKunciBerkas,
    error,
    berkas,
    billingLocked,
    form,
    opsiMaster,
    namaFile,
    rows,
    rowsTampil,
    previewUrl,
    previewNama,
    previewIsImage,
    previewIsPdf,
    resetForm,
    onFileChange,
    namaBerkas,
    openPreview,
    loadData,
    upload,
    confirmDelete,
  }
}
