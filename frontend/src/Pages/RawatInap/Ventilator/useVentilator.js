import { computed, reactive, ref, watch } from "vue"
import { mulaiVentilator, simpanChecklistVAP, simpanMonitoringVentilator, simpanSettingVentilator, ventilatorPasienData } from "../../../lib/faisal/api"
import { useNotifikasi } from "../../../lib/shared/useNotifikasi"

export function useVentilator(props) {
  const notifikasi = useNotifikasi()
  const data = ref(emptyData())
  const loading = ref(false)
  const saving = ref(false)
  const error = ref('')
  const formVisible = ref(true)
  const tab = ref('setting')
  const historyTab = ref('setting')
  const keyword = ref('')
  const doctor = ref({})
  const officer = ref({})
  const usage = reactive(emptyUsage())
  const setting = reactive(emptySetting())
  const monitoring = reactive(emptyMonitoring())
  const vap = reactive(emptyVAP())
  
  const active = computed(() => (data.value.pemakaian || []).find(item => ['Aktif', 'Weaning'].includes(item.status)))
  const available = computed(() => (data.value.master || [])
    .filter(item => item.status === 'Tersedia' || item.kode_ventilator === active.value?.kode_ventilator)
    .map(item => ({ label: `${item.nama} - ${item.kode_ventilator}${item.ruangan ? ` - ${item.ruangan}` : ''}`, value: item.kode_ventilator })))
  const historyRows = computed(() => {
    const source = historyTab.value === 'setting' ? data.value.setting : historyTab.value === 'monitoring' ? data.value.monitoring : data.value.checklist_vap
    const query = keyword.value.trim().toLowerCase()
    return query ? (source || []).filter(item => Object.values(item).join(' ').toLowerCase().includes(query)) : (source || [])
  })
  
  const option = value => ({ label: value, value })
  const modeOptions = ['VCV', 'PCV', 'SIMV-VC', 'SIMV-PC', 'PSV', 'CPAP', 'BiPAP', 'APRV', 'Lain-lain'].map(option)
  const airwayOptions = ['ETT', 'Trakeostomi', 'NIV Mask', 'Lain-lain'].map(option)
  const awarenessOptions = ['Compos Mentis', 'Apatis', 'Somnolence', 'Sopor', 'Koma'].map(option)
  const stageOptions = ['Monitoring', 'SBT', 'Weaning', 'Ekstubasi'].map(option)
  
  function emptyData() { return { master: [], pemakaian: [], setting: [], monitoring: [], checklist_vap: [], billing_terkunci: false } }
  function localDateTime() { const value = new Date(Date.now() - new Date().getTimezoneOffset() * 60000); return value.toISOString().slice(0, 16) }
  function emptyUsage() { return { no_rawat: props.patient?.no_rawat || '', kode_ventilator: '', tanggal_mulai: localDateTime(), ruangan: props.patient?.kamar || props.patient?.poliklinik || '', indikasi: '', jenis_jalan_napas: 'ETT', ukuran_jalan_napas: '', kedalaman_jalan_napas: '', dokter_penanggung_jawab: '', petugas_pemasangan: '', catatan: '' } }
  function emptySetting(id = 0) { return { id_pemakaian: id, waktu_setting: localDateTime(), mode: '', fio2: '', peep: '', tidal_volume: '', frekuensi_set: '', pressure_control: '', pressure_support: '', petugas: '', catatan: '' } }
  function emptyMonitoring(id = 0) { return { id_pemakaian: id, waktu_monitoring: localDateTime(), kesadaran: '', tekanan_darah: '', nadi: '', respirasi: '', suhu: '', spo2: '', tahap: 'Monitoring', petugas: '', catatan: '' } }
  function emptyVAP(id = 0) { return { id_pemakaian: id, waktu_checklist: localDateTime(), elevasi_kepala: false, perawatan_mulut: false, suction: false, evaluasi_sedasi: false, sat: false, sbt: false, pencegahan_dvt: false, pencegahan_ulkus: false, tekanan_cuff: '', petugas: '', catatan: '' } }
  function number(value) { return value === '' || value === null ? 0 : Number(value) }
  function formatDateTime(value) { if (!value) return '-'; const [date, time = ''] = String(value).replace('T', ' ').split(' '); const [year, month, day] = date.split('-'); return `${day}/${month}/${year} ${time.slice(0, 5)}` }
  function yes(value) { return value ? 'Ya' : 'Tidak' }
  function resetClinicalForms(id) { Object.assign(setting, emptySetting(id)); Object.assign(monitoring, emptyMonitoring(id)); Object.assign(vap, emptyVAP(id)) }
  
  async function load() {
    if (!props.patient?.no_rawat) return
    loading.value = true; error.value = ''
    try {
      data.value = await ventilatorPasienData(props.token, props.patient.no_rawat)
      Object.assign(usage, emptyUsage(), { no_rawat: props.patient.no_rawat, ruangan: props.patient.kamar || props.patient.poliklinik || '' })
      resetClinicalForms(active.value?.id || data.value.pemakaian?.[0]?.id || 0)
    } catch (err) { error.value = err.message || 'Data ventilator tidak dapat dibaca.'; notifikasi.gagal(error.value) }
    finally { loading.value = false }
  }
  
  async function saveUsage() {
    if (!usage.kode_ventilator || !usage.tanggal_mulai || !usage.indikasi.trim() || !doctor.value.kode || !(officer.value.nip || officer.value.kode)) { notifikasi.peringatan('Ventilator, tanggal mulai, indikasi, dokter, dan petugas wajib diisi.'); return }
  
    const kedalaman = String(usage.kedalaman_jalan_napas ?? '').trim().replace(',', '.')
    if (kedalaman !== '' && (!Number.isFinite(Number(kedalaman)) || Number(kedalaman) < 0 || Number(kedalaman) > 999.99)) {
      notifikasi.peringatan('Kedalaman jalan napas harus berupa angka antara 0 sampai 999,99.')
      return
    }
    saving.value = true
    try {
      usage.dokter_penanggung_jawab = doctor.value.kode
      usage.petugas_pemasangan = officer.value.nip || officer.value.kode
      const response = await mulaiVentilator(props.token, { ...usage, kedalaman_jalan_napas: kedalaman })
      notifikasi.sukses(response?.pesan || 'Pemakaian ventilator berhasil dimulai.')
      await load()
    } catch (err) { notifikasi.gagal(err.message) }
    finally { saving.value = false }
  }
  
  async function saveClinical() {
    if (!active.value?.id) return
    if (tab.value === 'setting' && !setting.mode) { notifikasi.peringatan('Mode ventilator wajib dipilih.'); return }
    saving.value = true
    try {
      let response
      if (tab.value === 'setting') response = await simpanSettingVentilator(props.token, { ...setting, fio2: number(setting.fio2), peep: number(setting.peep), tidal_volume: number(setting.tidal_volume), frekuensi_set: number(setting.frekuensi_set), pressure_control: number(setting.pressure_control), pressure_support: number(setting.pressure_support) })
      else if (tab.value === 'monitoring') response = await simpanMonitoringVentilator(props.token, { ...monitoring, nadi: number(monitoring.nadi), respirasi: number(monitoring.respirasi), suhu: number(monitoring.suhu), spo2: number(monitoring.spo2) })
      else response = await simpanChecklistVAP(props.token, { ...vap, tekanan_cuff: number(vap.tekanan_cuff) })
      notifikasi.sukses(response?.pesan || 'Catatan ventilator berhasil disimpan.')
      historyTab.value = tab.value
      await load()
    } catch (err) { notifikasi.gagal(err.message) }
    finally { saving.value = false }
  }
  
  watch(() => props.patient?.no_rawat, () => { doctor.value = {}; officer.value = {}; keyword.value = ''; load() }, { immediate: true })
  return {
    data,
    loading,
    saving,
    error,
    formVisible,
    tab,
    historyTab,
    keyword,
    doctor,
    officer,
    usage,
    setting,
    monitoring,
    vap,
    active,
    available,
    historyRows,
    modeOptions,
    airwayOptions,
    awarenessOptions,
    stageOptions,
    number,
    formatDateTime,
    yes,
    load,
    saveUsage,
    saveClinical,
  }
}
