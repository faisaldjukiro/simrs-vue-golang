// @ts-nocheck -- migrasi TypeScript bertahap; kontrak data modul lama belum sepenuhnya bertipe.
import { computed, nextTick, reactive, ref, watch } from "vue"
import { hapusImplementasiKeperawatan, implementasiKeperawatanData, simpanImplementasiKeperawatan, ubahImplementasiKeperawatan } from "../../../lib/faisal/api"
import { useNotifikasi } from "../../../lib/shared/useNotifikasi"

export function useImplementasiKeperawatan(props) {
  const notifikasi = useNotifikasi()
  const loading = ref(false)
  const saving = ref(false)
  const deleting = ref(false)
  const error = ref('')
  const records = ref([])
  const petugas = ref({ nip: '', nama: '', jabatan: '' })
  const selectedOfficer = ref({ nip: '', nama: '', jabatan: '' })
  const canChooseOfficer = ref(false)
  const billingLocked = ref(false)
  const editingKey = ref(null)
  const deleteTarget = ref(null)
  const formVisible = ref(false)
  const keyword = ref('')
  const form = reactive(emptyForm())
  
  const tableRows = computed(() => records.value.map((item) => ({ ...item, _key: keyFor(item) })))
  const filteredRows = computed(() => {
    const query = keyword.value.trim().toLowerCase()
    if (!query) return tableRows.value
    return tableRows.value.filter((item) => [
      item.tanggal, item.jam, item.uraian, item.nip, item.nama_petugas, item.jabatan,
    ].join(' ').toLowerCase().includes(query))
  })
  
  function currentDate() {
    const now = new Date()
    const offset = now.getTimezoneOffset() * 60000
    return new Date(now.getTime() - offset).toISOString().slice(0, 10)
  }
  
  function currentTime() {
    return [new Date().getHours(), new Date().getMinutes(), new Date().getSeconds()]
      .map((value) => String(value).padStart(2, '0')).join(':')
  }
  
  function emptyForm() {
    return {
      no_rawat: props.patient?.no_rawat || '',
      tanggal: currentDate(),
      jam: currentTime(),
      uraian: '',
      nip: '',
    }
  }
  
  function resetForm() {
    Object.assign(form, emptyForm(), { no_rawat: props.patient?.no_rawat || '', nip: petugas.value.nip || '' })
    selectedOfficer.value = { ...petugas.value }
    editingKey.value = null
  }
  
  function keyFor(item) {
    return `${item.no_rawat}|${item.tanggal}|${item.jam}`
  }
  
  function recordKey(item) {
    return { no_rawat: item.no_rawat, tanggal: item.tanggal, jam: item.jam }
  }
  
  function formatDate(value) {
    if (!value) return '-'
    const [year, month, day] = value.split('-')
    return `${day}/${month}/${year}`
  }
  
  async function loadRecords() {
    if (!props.token || !props.patient?.no_rawat) return
    loading.value = true
    error.value = ''
    try {
      const data = await implementasiKeperawatanData(props.token, props.patient.no_rawat)
      records.value = data?.catatan || []
      petugas.value = data?.petugas || { nip: '', nama: '', jabatan: '' }
      canChooseOfficer.value = Boolean(data?.bisa_memilih_petugas)
      billingLocked.value = Boolean(data?.billing_terkunci)
      resetForm()
    } catch (err) {
      error.value = err.message || 'Data implementasi keperawatan tidak dapat dibaca.'
      notifikasi.gagal(error.value)
    } finally {
      loading.value = false
    }
  }
  
  function editRecord(item) {
    if (billingLocked.value || !item.bisa_diubah) {
      notifikasi.peringatan(billingLocked.value
        ? 'Kunjungan sudah masuk billing. Catatan hanya dapat dilihat.'
        : 'Catatan ini hanya dapat diedit oleh petugas yang membuatnya.')
      return
    }
    editingKey.value = recordKey(item)
    Object.assign(form, {
      no_rawat: item.no_rawat,
      tanggal: item.tanggal,
      jam: item.jam,
      uraian: item.uraian,
      nip: item.nip,
    })
    selectedOfficer.value = { nip: item.nip, nama: item.nama_petugas || item.nip, jabatan: item.jabatan || '' }
    formVisible.value = true
    nextTick(() => document.querySelector('.implementasi-page .clinical-form-card')?.scrollIntoView({ behavior: 'smooth', block: 'start' }))
  }
  
  function payload() {
    return {
      kunci_lama: editingKey.value,
      no_rawat: String(props.patient?.no_rawat || '').trim(),
      tanggal: String(form.tanggal || '').trim(),
      jam: String(form.jam || '').trim(),
      uraian: String(form.uraian || '').trim(),
      nip: selectedOfficer.value?.nip || form.nip || '',
    }
  }
  
  async function saveRecord() {
    if (billingLocked.value) return notifikasi.peringatan('Kunjungan sudah masuk billing. Catatan tidak dapat disimpan.')
    if (!String(form.uraian || '').trim()) {
      notifikasi.peringatan('Uraian implementasi keperawatan wajib diisi.')
      return nextTick(() => document.querySelector('.implementasi-uraian textarea')?.focus())
    }
    saving.value = true
    try {
      const response = editingKey.value
        ? await ubahImplementasiKeperawatan(props.token, payload())
        : await simpanImplementasiKeperawatan(props.token, payload())
      notifikasi.sukses(response?.pesan || 'Implementasi keperawatan berhasil disimpan.')
      await loadRecords()
      formVisible.value = false
    } catch (err) {
      notifikasi.gagal(err.message || 'Implementasi keperawatan gagal disimpan.')
    } finally {
      saving.value = false
    }
  }
  
  async function confirmDelete() {
    if (!deleteTarget.value) return
    deleting.value = true
    try {
      const response = await hapusImplementasiKeperawatan(props.token, recordKey(deleteTarget.value))
      notifikasi.sukses(response?.pesan || 'Catatan berhasil dihapus.')
      deleteTarget.value = null
      await loadRecords()
    } catch (err) {
      notifikasi.gagal(err.message || 'Catatan gagal dihapus.')
    } finally {
      deleting.value = false
    }
  }
  
  watch(() => props.patient?.no_rawat, () => {
    formVisible.value = false
    keyword.value = ''
    loadRecords()
  }, { immediate: true })
  return {
    loading,
    saving,
    deleting,
    error,
    records,
    petugas,
    selectedOfficer,
    canChooseOfficer,
    billingLocked,
    editingKey,
    deleteTarget,
    formVisible,
    keyword,
    form,
    filteredRows,
    resetForm,
    formatDate,
    loadRecords,
    editRecord,
    saveRecord,
    confirmDelete,
  }
}
