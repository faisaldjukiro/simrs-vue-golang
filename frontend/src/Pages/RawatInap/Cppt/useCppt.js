import { computed, nextTick, reactive, ref, watch } from "vue"
import { cpptData, hapusCppt, simpanCppt, ubahCppt } from "../../../lib/faisal/api"
import { useNotifikasi } from "../../../lib/shared/useNotifikasi"

export function useCppt(props) {
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
  const defaultAwareness = ['Compos Mentis', 'Apatis', 'Somnolence', 'Sopor', 'Coma']
  const awarenessOptions = ref(toSelectOptions(defaultAwareness))
  const editingKey = ref(null)
  const deleteTarget = ref(null)
  const formVisible = ref(false)
  const kataKunciCatatan = ref('')
  const judulCatatan = computed(() => props.jenisRawat === 'ralan' ? 'Pemeriksaan & SOAP' : 'CPPT & SOAP')
  
  const form = reactive(emptyForm())
  const tableRecords = computed(() => records.value.map((item) => ({
    ...item,
    _key: keyFor(item),
  })))
  const tableRecordsTampil = computed(() => {
    const keyword = kataKunciCatatan.value.trim().toLowerCase()
    if (!keyword) return tableRecords.value
    return tableRecords.value.filter((item) => teksCatatan(item).includes(keyword))
  })
  
  function toSelectOptions(options = []) {
    return options.map((option) => {
      if (typeof option === 'object' && option !== null) return option
      return { label: String(option), value: String(option) }
    })
  }
  
  function currentDate() {
    const date = new Date()
    const offset = date.getTimezoneOffset() * 60000
    return new Date(date.getTime() - offset).toISOString().slice(0, 10)
  }
  
  function currentTime() {
    const date = new Date()
    return [date.getHours(), date.getMinutes(), date.getSeconds()]
      .map((value) => String(value).padStart(2, '0'))
      .join(':')
  }
  
  function emptyForm() {
    return {
      jenis_rawat: props.jenisRawat,
      no_rawat: props.patient?.no_rawat || '',
      tgl_perawatan: currentDate(),
      jam_rawat: currentTime(),
      suhu_tubuh: '',
      tensi: '',
      nadi: '',
      respirasi: '',
      tinggi: '',
      berat: '',
      spo2: '',
      gcs: '',
      kesadaran: 'Compos Mentis',
      subjek: '',
      objek: '',
      alergi: '',
      lingkar_perut: '',
      asesmen: '',
      plan: '',
      instruksi: '',
      evaluasi: '',
      nip: '',
    }
  }
  
  function resetForm() {
    Object.assign(form, emptyForm(), {
      no_rawat: props.patient.no_rawat,
      nip: petugas.value.nip || '',
    })
    editingKey.value = null
    selectedOfficer.value = { ...petugas.value }
  }
  
  function keyFor(item) {
    return `${item.no_rawat}|${item.tgl_perawatan}|${item.jam_rawat}`
  }
  
  function teksCatatan(item) {
    return [
      item.tgl_perawatan,
      item.jam_rawat,
      item.nama_petugas,
      item.jabatan,
      item.nip,
      item.kesadaran,
      item.suhu_tubuh,
      item.tensi,
      item.nadi,
      item.respirasi,
      item.spo2,
      item.gcs,
      item.tinggi,
      item.berat,
      item.alergi,
      item.lingkar_perut,
      item.subjek,
      item.objek,
      item.asesmen,
      item.plan,
      item.instruksi,
      item.evaluasi,
    ].join(' ').toLowerCase()
  }
  
  function recordKey(item) {
    return {
      jenis_rawat: item.jenis_rawat || props.jenisRawat,
      no_rawat: String(item?.no_rawat || props.patient?.no_rawat || '').trim(),
      tgl_perawatan: item.tgl_perawatan,
      jam_rawat: item.jam_rawat,
    }
  }
  
  async function loadRecords() {
    if (!props.token || !props.patient.no_rawat) return
    loading.value = true
    error.value = ''
    try {
      const data = await cpptData(props.token, props.patient.no_rawat, props.jenisRawat)
      records.value = data?.catatan || []
      petugas.value = data?.petugas || { nip: '', nama: '', jabatan: '' }
      canChooseOfficer.value = Boolean(data?.bisa_memilih_petugas)
      billingLocked.value = Boolean(data?.billing_terkunci)
      awarenessOptions.value = toSelectOptions(
        data?.pilihan_kesadaran?.length ? data.pilihan_kesadaran : defaultAwareness,
      )
      resetForm()
    } catch (err) {
      error.value = err.message
      notifikasi.gagal(err.message || 'Catatan CPPT/SOAP tidak dapat dibaca.')
    } finally {
      loading.value = false
    }
  }
  
  function editRecord(item) {
    if (billingLocked.value) {
      notifikasi.peringatan('Kunjungan sudah masuk billing. Catatan hanya dapat dilihat.')
      return
    }
    if (!item.bisa_diubah) {
      notifikasi.peringatan('Catatan ini hanya dapat diedit oleh petugas yang membuatnya.')
      return
    }
    editingKey.value = recordKey(item)
    formVisible.value = true
    selectedOfficer.value = {
      nip: item.nip || '',
      nama: item.nama_petugas || item.nip || '',
      jabatan: item.jabatan || '',
    }
    Object.assign(form, {
      ...emptyForm(),
      no_rawat: item.no_rawat,
      tgl_perawatan: item.tgl_perawatan,
      jam_rawat: item.jam_rawat,
      suhu_tubuh: item.suhu_tubuh,
      tensi: item.tensi,
      nadi: item.nadi,
      respirasi: item.respirasi,
      tinggi: item.tinggi,
      berat: item.berat,
      spo2: item.spo2,
      gcs: item.gcs,
      kesadaran: item.kesadaran,
      subjek: item.subjek,
      objek: item.objek,
      alergi: item.alergi,
      lingkar_perut: item.lingkar_perut,
      asesmen: item.asesmen,
      plan: item.plan,
      instruksi: item.instruksi,
      evaluasi: item.evaluasi,
      nip: item.nip,
      no_rawat: props.patient.no_rawat,
    })
    nextTick(() => {
      document.querySelector('.clinical-form-card')?.scrollIntoView({ behavior: 'smooth', block: 'start' })
    })
  }
  
  function hasClinicalContent() {
    return [
      form.suhu_tubuh, form.tensi, form.nadi, form.respirasi, form.tinggi,
      form.berat, form.spo2, form.gcs, form.subjek, form.objek, form.alergi,
      form.lingkar_perut, form.asesmen, form.plan, form.instruksi, form.evaluasi,
    ].some((value) => String(value || '').trim())
  }
  
  function payload() {
    return {
      kunci_lama: editingKey.value,
      jenis_rawat: props.jenisRawat,
      no_rawat: String(form.no_rawat || '').trim(),
      tgl_perawatan: String(form.tgl_perawatan || '').trim(),
      jam_rawat: String(form.jam_rawat || '').trim(),
      suhu_tubuh: String(form.suhu_tubuh || '').trim(),
      tensi: String(form.tensi || '').trim(),
      nadi: String(form.nadi || '').trim(),
      respirasi: String(form.respirasi || '').trim(),
      tinggi: String(form.tinggi || '').trim(),
      berat: String(form.berat || '').trim(),
      spo2: String(form.spo2 || '').trim(),
      gcs: String(form.gcs || '').trim(),
      kesadaran: String(form.kesadaran || '').trim(),
      subjek: String(form.subjek || '').trim(),
      objek: String(form.objek || '').trim(),
      alergi: String(form.alergi || '').trim(),
      lingkar_perut: String(form.lingkar_perut || '').trim(),
      asesmen: String(form.asesmen || '').trim(),
      plan: String(form.plan || '').trim(),
      instruksi: String(form.instruksi || '').trim(),
      evaluasi: String(form.evaluasi || '').trim(),
      nip: selectedOfficer.value?.nip || form.nip || '',
    }
  }
  
  async function saveRecord() {
    if (billingLocked.value) {
      notifikasi.peringatan('Kunjungan sudah masuk billing. Catatan tidak dapat disimpan.')
      return
    }
    if (!hasClinicalContent()) {
      notifikasi.peringatan('Isi minimal satu pemeriksaan atau catatan SOAP.')
      return
    }
    saving.value = true
    try {
      const response = editingKey.value
        ? await ubahCppt(props.token, payload())
        : await simpanCppt(props.token, payload())
      notifikasi.sukses(response?.pesan || (editingKey.value ? 'Catatan berhasil diperbarui.' : 'Catatan berhasil disimpan.'))
      await loadRecords()
      formVisible.value = false
    } catch (err) {
      notifikasi.gagal(err.message || 'Catatan CPPT/SOAP gagal disimpan.')
    } finally {
      saving.value = false
    }
  }
  
  async function confirmDelete() {
    if (!deleteTarget.value) return
    if (billingLocked.value) {
      deleteTarget.value = null
      notifikasi.peringatan('Kunjungan sudah masuk billing. Catatan tidak dapat dihapus.')
      return
    }
    deleting.value = true
    try {
      const response = await hapusCppt(props.token, recordKey(deleteTarget.value))
      notifikasi.sukses(response?.pesan || 'Catatan berhasil dihapus.')
      deleteTarget.value = null
      await loadRecords()
    } catch (err) {
      notifikasi.gagal(err.message || 'Catatan CPPT/SOAP gagal dihapus.')
    } finally {
      deleting.value = false
    }
  }
  
  function formatDate(value) {
    if (!value) return '-'
    const [year, month, day] = value.split('-')
    return `${day}/${month}/${year}`
  }
  
  watch([() => props.patient.no_rawat, () => props.jenisRawat], () => {
    formVisible.value = false
    kataKunciCatatan.value = ''
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
    awarenessOptions,
    editingKey,
    deleteTarget,
    formVisible,
    kataKunciCatatan,
    judulCatatan,
    form,
    tableRecordsTampil,
    resetForm,
    loadRecords,
    editRecord,
    saveRecord,
    confirmDelete,
    formatDate,
  }
}
