import { computed, reactive, ref, watch } from "vue"
import { awalMedisUmumData, hapusAwalMedisUmum, simpanAwalMedisUmum, ubahAwalMedisUmum } from "../../../lib/faisal/api"
import { useNotifikasi } from "../../../lib/shared/useNotifikasi"

export function useAwalMedisUmum(props) {
  const notifikasi = useNotifikasi()
  const loading = ref(false)
  const saving = ref(false)
  const deleting = ref(false)
  const formOpen = ref(true)
  const confirmDelete = ref(false)
  const data = reactive({ tersedia: false, billing_terkunci: false, penilaian: null, dokter: null, pilihan: {} })
  const dokter = ref({})
  const form = reactive(formKosong())
  
  const editing = computed(() => Boolean(data.tersedia))
  const billingLocked = computed(() => Boolean(data.billing_terkunci))
  const pilihan = computed(() => ({
    anamnesis: list(data.pilihan?.anamnesis || ['Autoanamnesis', 'Alloanamnesis']),
    keadaan: list(data.pilihan?.keadaan || ['Sehat', 'Sakit Ringan', 'Sakit Sedang', 'Sakit Berat']),
    kesadaran: list(data.pilihan?.kesadaran || ['Compos Mentis', 'Apatis', 'Somnolen', 'Sopor', 'Koma']),
    pemeriksaan: list(data.pilihan?.pemeriksaan || ['Normal', 'Abnormal', 'Tidak Diperiksa']),
  }))
  
  function list(items) { return items.map((value) => ({ label: value, value })) }
  function sekarang() { const d = new Date(Date.now() - new Date().getTimezoneOffset() * 60000); return d.toISOString().slice(0, 19) }
  function tanggalInput(value) { return String(value || '').replace(' ', 'T').slice(0, 19) }
  function formKosong() {
    return {
      no_rawat: props.patient?.no_rawat || '', tanggal: sekarang(), kode_dokter: '', anamnesis: 'Autoanamnesis', hubungan: '',
      keluhan_utama: '', rps: '', rpd: '', rpk: '', rpo: '', alergi: '', keadaan: 'Sehat', gcs: '', kesadaran: 'Compos Mentis',
      td: '', nadi: '', rr: '', suhu: '', spo2: '', bb: '', tb: '', kepala: 'Normal', gigi: 'Normal', tht: 'Normal',
      thoraks: 'Normal', abdomen: 'Normal', genital: 'Normal', ekstremitas: 'Normal', kulit: 'Normal', ket_fisik: '', ket_lokalis: '',
      penunjang: '', diagnosis: '', tata: '', konsulrujuk: '',
    }
  }
  
  function isiForm() {
    const record = data.penilaian || {}
    Object.assign(form, formKosong(), record, {
      no_rawat: props.patient.no_rawat,
      tanggal: tanggalInput(record.tanggal) || sekarang(),
    })
    dokter.value = data.dokter ? { ...data.dokter } : {}
  }
  
  async function loadData() {
    if (!props.token || !props.patient?.no_rawat) return
    loading.value = true
    try {
      const response = await awalMedisUmumData(props.token, props.patient.no_rawat)
      Object.assign(data, response || {})
      isiForm()
      formOpen.value = true
    } catch (error) {
      notifikasi.gagal(error.message || 'Penilaian awal medis umum tidak dapat dibaca.')
    } finally { loading.value = false }
  }
  
  function payload() {
    return { ...form, tanggal: form.tanggal.replace('T', ' '), kode_dokter: dokter.value?.kode || '' }
  }
  
  async function save() {
    if (billingLocked.value) { notifikasi.peringatan('Kunjungan sudah masuk billing. Penilaian hanya dapat dilihat.'); return }
    if (!dokter.value?.kode) { notifikasi.peringatan('Pilih dokter yang melakukan penilaian.'); return }
    saving.value = true
    try {
      const response = editing.value ? await ubahAwalMedisUmum(props.token, payload()) : await simpanAwalMedisUmum(props.token, payload())
      notifikasi.sukses(response?.pesan || 'Penilaian awal medis umum berhasil disimpan.')
      await loadData()
    } catch (error) {
      notifikasi.gagal(error.message || 'Penilaian awal medis umum gagal disimpan.')
    } finally { saving.value = false }
  }
  
  async function remove() {
    deleting.value = true
    try {
      const response = await hapusAwalMedisUmum(props.token, props.patient.no_rawat)
      confirmDelete.value = false
      notifikasi.sukses(response?.pesan || 'Penilaian awal medis umum berhasil dihapus.')
      await loadData()
    } catch (error) {
      notifikasi.gagal(error.message || 'Penilaian awal medis umum gagal dihapus.')
    } finally { deleting.value = false }
  }
  
  watch(() => props.patient.no_rawat, loadData, { immediate: true })
  return {
    loading,
    saving,
    deleting,
    formOpen,
    confirmDelete,
    dokter,
    form,
    editing,
    billingLocked,
    pilihan,
    isiForm,
    save,
    remove,
  }
}
