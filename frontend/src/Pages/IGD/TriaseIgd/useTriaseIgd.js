import { computed, reactive, ref, watch } from "vue"
import { hapusTriaseIgd, simpanTriaseIgd, triaseIgdData, ubahTriaseIgd } from "../../../lib/faisal/api"
import { useNotifikasi } from "../../../lib/shared/useNotifikasi"

export function useTriaseIgd(props) {
  const notifikasi = useNotifikasi()
  const loading = ref(false)
  const saving = ref(false)
  const deleting = ref(false)
  const activeType = ref('primer')
  const selectedOfficer = ref({ nip: '', nama: '', jabatan: '' })
  const deleteType = ref('')
  const data = reactive({ petugas: {}, macam_kasus: [], pemeriksaan: [], kriteria_skala: [], mendukung_hand_over: false, triase: null, billing_terkunci: false })
  const billingLocked = computed(() => Boolean(data.billing_terkunci))
  const form = reactive(emptyForm())
  
  const caraMasukOptions = toOptions(['Jalan', 'Brankar', 'Kursi Roda', 'Digendong'])
  const transportOptions = toOptions(['-', 'AGD', 'Sendiri', 'Swasta'])
  const reasonOptions = toOptions(['Datang Sendiri', 'Polisi', 'Rujukan', '-'])
  const handOverOptions = toOptions(['Bedah', 'Penyakit Dalam', 'OBGIN', 'Anak'])
  const specialNeedsOptions = toOptions(['-', 'UPPA', 'Airborne', 'Dekontaminan'])
  const scaleNames = { 1: 'Immediate / Segera', 2: 'Emergensi', 3: 'Urgensi', 4: 'Semi Urgensi', 5: 'Non Urgensi' }
  
  const editing = computed(() => Boolean(data.triase?.[activeType.value]))
  const scales = computed(() => activeType.value === 'primer' ? [1, 2] : [3, 4, 5])
  const planOptions = computed(() => toOptions(activeType.value === 'primer'
    ? ['Ruang Resusitasi', 'Ruang Kritis', 'Zona Kuning', 'Zona Hijau', 'Zona Hitam']
    : ['Zona Kuning', 'Zona Hijau']))
  const planNote = computed(() => activeType.value === 'primer'
    ? 'Ruang Resusitasi dan Ruang Kritis mengikuti pilihan Triase Primer SIMRS Khanza.'
    : 'Triase Sekunder di SIMRS Khanza hanya menerima Zona Kuning atau Zona Hijau.')
  const caseOptions = computed(() => data.macam_kasus.map((item) => ({ label: item.nama, value: item.kode })))
  const selectedCriteria = computed(() => data.kriteria_skala.filter((item) => item.skala === form.skala && form.kode_kriteria.includes(item.kode)))
  const criteriaGroups = computed(() => data.pemeriksaan.map((pemeriksaan) => ({
    ...pemeriksaan,
    items: data.kriteria_skala.filter((item) => item.skala === form.skala && item.kode_pemeriksaan === pemeriksaan.kode),
  })).filter((group) => group.items.length > 0))
  const existingRecords = computed(() => [
    data.triase?.primer ? { jenis: 'primer', label: 'Triase Primer', bagian: data.triase.primer } : null,
    data.triase?.sekunder ? { jenis: 'sekunder', label: 'Triase Sekunder', bagian: data.triase.sekunder } : null,
  ].filter(Boolean))
  
  function toOptions(items) { return items.map((value) => ({ label: value, value })) }
  function localDateTime() {
    const date = new Date(Date.now() - new Date().getTimezoneOffset() * 60000)
    return date.toISOString().slice(0, 19)
  }
  function forInput(value) { return String(value || '').replace(' ', 'T').slice(0, 19) }
  function forPayload(value) { return String(value || '').replace('T', ' ') }
  
  function emptyForm() {
    return {
      no_rawat: props.patient?.no_rawat || '', jenis: 'primer', tanggal_kunjungan: localDateTime(),
      cara_masuk: 'Jalan', alat_transportasi: '-', alasan_kedatangan: 'Datang Sendiri',
      keterangan_kedatangan: '', kode_kasus: '', tekanan_darah: '', nadi: '', pernapasan: '',
      suhu: '', saturasi_o2: '', nyeri: '', hand_over: 'Penyakit Dalam', isi_utama: '',
      kebutuhan_khusus: '-', catatan: '', plan: 'Ruang Resusitasi', tanggal_triase: localDateTime(),
      nip: '', skala: 1, kode_kriteria: [],
    }
  }
  
  function fillForm(jenis = activeType.value) {
    activeType.value = jenis
    const common = data.triase || {}
    const record = common?.[jenis]
    Object.assign(form, emptyForm(), {
      no_rawat: props.patient.no_rawat,
      jenis,
      tanggal_kunjungan: forInput(common.tanggal_kunjungan) || localDateTime(),
      cara_masuk: common.cara_masuk || 'Jalan',
      alat_transportasi: common.alat_transportasi || '-',
      alasan_kedatangan: common.alasan_kedatangan || 'Datang Sendiri',
      keterangan_kedatangan: common.keterangan_kedatangan || '',
      kode_kasus: common.kode_kasus || data.macam_kasus?.[0]?.kode || '',
      tekanan_darah: common.tekanan_darah || '', nadi: common.nadi || '', pernapasan: common.pernapasan || '',
      suhu: common.suhu || '', saturasi_o2: common.saturasi_o2 || '', nyeri: common.nyeri || '',
      hand_over: common.hand_over || 'Penyakit Dalam',
      isi_utama: record?.isi_utama || '', kebutuhan_khusus: record?.kebutuhan_khusus || '-',
      catatan: record?.catatan || '', plan: record?.plan || (jenis === 'primer' ? 'Ruang Resusitasi' : 'Zona Kuning'),
      tanggal_triase: forInput(record?.tanggal_triase) || localDateTime(), nip: record?.nip || data.petugas?.nip || '',
      skala: record?.skala || (jenis === 'primer' ? 1 : 3),
      kode_kriteria: (record?.kriteria_terpilih || []).map((item) => item.kode),
    })
    selectedOfficer.value = record
      ? { nip: record.nip, nama: record.nama_petugas, jabatan: record.jabatan_petugas }
      : { ...(data.petugas || {}) }
  }
  
  function chooseScale(value) {
    if (form.skala === value) return
    form.skala = value
    form.kode_kriteria = []
  }
  
  function toggleCriterion(code) {
    const index = form.kode_kriteria.indexOf(code)
    if (index >= 0) form.kode_kriteria.splice(index, 1)
    else form.kode_kriteria.push(code)
  }
  
  function payload() {
    return {
      ...form,
      jenis: activeType.value,
      tanggal_kunjungan: forPayload(form.tanggal_kunjungan),
      tanggal_triase: forPayload(form.tanggal_triase),
      nip: selectedOfficer.value?.nip || form.nip,
      kode_kriteria: [...form.kode_kriteria],
    }
  }
  
  async function loadData() {
    if (!props.token || !props.patient.no_rawat) return
    loading.value = true
    try {
      const response = await triaseIgdData(props.token, props.patient.no_rawat)
      Object.assign(data, response || {})
      const preferred = response?.triase?.primer ? 'primer' : response?.triase?.sekunder ? 'sekunder' : 'primer'
      fillForm(preferred)
    } catch (error) {
      notifikasi.gagal(error.message || 'Data triase IGD tidak dapat dibaca.')
    } finally { loading.value = false }
  }
  
  async function save() {
    if (billingLocked.value) { notifikasi.peringatan('Kunjungan sudah masuk billing. Triase IGD hanya dapat dilihat.'); return }
    if (!selectedOfficer.value?.nip) { notifikasi.peringatan('Pilih dokter atau petugas triase terlebih dahulu.'); return }
    if (form.kode_kriteria.length === 0) { notifikasi.peringatan('Pilih minimal satu kriteria skala triase.'); return }
    saving.value = true
    try {
      const response = editing.value ? await ubahTriaseIgd(props.token, payload()) : await simpanTriaseIgd(props.token, payload())
      notifikasi.sukses(response?.pesan || 'Triase IGD berhasil disimpan.')
      await loadData()
    } catch (error) { notifikasi.gagal(error.message || 'Triase IGD gagal disimpan.') }
    finally { saving.value = false }
  }
  
  async function confirmDelete() {
    if (!deleteType.value) return
    if (billingLocked.value) { deleteType.value = ''; notifikasi.peringatan('Kunjungan sudah masuk billing. Triase IGD tidak dapat dihapus.'); return }
    deleting.value = true
    try {
      const response = await hapusTriaseIgd(props.token, props.patient.no_rawat, deleteType.value)
      notifikasi.sukses(response?.pesan || 'Triase IGD berhasil dihapus.')
      deleteType.value = ''
      await loadData()
    } catch (error) { notifikasi.gagal(error.message || 'Triase IGD gagal dihapus.') }
    finally { deleting.value = false }
  }
  
  watch(() => props.patient.no_rawat, loadData, { immediate: true })
  return {
    loading,
    saving,
    deleting,
    activeType,
    selectedOfficer,
    deleteType,
    data,
    billingLocked,
    form,
    caraMasukOptions,
    transportOptions,
    reasonOptions,
    handOverOptions,
    specialNeedsOptions,
    scaleNames,
    scales,
    planOptions,
    planNote,
    caseOptions,
    selectedCriteria,
    criteriaGroups,
    existingRecords,
    fillForm,
    chooseScale,
    toggleCriterion,
    save,
    confirmDelete,
  }
}
