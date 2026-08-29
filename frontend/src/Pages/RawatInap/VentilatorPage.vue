<script setup>
import { Activity, ChevronDown, ChevronUp, ClipboardCheck, Gauge, LoaderCircle, Save, Settings2, ShieldCheck, Wind } from '@lucide/vue'
import PrimeColumn from 'primevue/column'
import { computed, reactive, ref, watch } from 'vue'
import CariDokter from '../../Components/Ui/CariDokter.vue'
import CariPetugas from '../../Components/Ui/CariPetugas.vue'
import DataTable from '../../Components/Ui/DataTable.vue'
import FormInput from '../../Components/Ui/FormInput.vue'
import TableSearch from '../../Components/Ui/TableSearch.vue'
import { mulaiVentilator, simpanChecklistVAP, simpanMonitoringVentilator, simpanSettingVentilator, ventilatorPasienData } from '../../lib/faisal/api'
import { useNotifikasi } from '../../lib/shared/useNotifikasi'

const props = defineProps({ token: { type: String, required: true }, patient: { type: Object, required: true } })
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
</script>

<template>
  <section class="cppt-page ventilator-page">
    <article class="cppt-form-card">
      <header class="cppt-section-header">
        <div><span>Terapi Respirasi Rawat Inap</span><h3>{{ active ? 'Pencatatan Ventilator' : 'Mulai Pemakaian Ventilator' }}</h3><p>Setting, monitoring pasien, dan pencegahan ventilator-associated pneumonia.</p></div>
        <div class="cppt-section-tools"><Wind :size="24" /><button type="button" class="cppt-button toggle icon-only" :title="formVisible ? 'Sembunyikan Form Input' : 'Tampilkan Form Input'" @click="formVisible = !formVisible"><ChevronUp v-if="formVisible" :size="15" /><ChevronDown v-else :size="15" /></button></div>
      </header>
      <div v-if="data.billing_terkunci" class="billing-lock">Kunjungan sudah masuk billing. Data ventilator hanya dapat dilihat.</div>
      <div v-if="loading" class="cppt-state"><LoaderCircle class="spin" :size="25" /><strong>Menarik data ventilator...</strong></div>
      <div v-else-if="error" class="cppt-state error"><strong>{{ error }}</strong><button type="button" @click="load">Coba Lagi</button></div>

      <form v-else-if="!active" v-show="formVisible" class="cppt-form" @submit.prevent="saveUsage">
        <fieldset class="form-compact" :disabled="saving || data.billing_terkunci">
          <div class="ventilator-usage-grid">
            <FormInput v-model="usage.kode_ventilator" label="Perangkat Ventilator" jenis="select" :options="available" filter required />
            <FormInput v-model="usage.tanggal_mulai" label="Tanggal Mulai" type="datetime-local" required /><FormInput v-model="usage.ruangan" label="Ruangan" />
            <CariDokter v-model="doctor" :token="token" required /><CariPetugas v-model="officer" :token="token" label="Petugas Pemasangan" required />
            <FormInput v-model="usage.jenis_jalan_napas" label="Jalan Napas" jenis="select" :options="airwayOptions" required />
            <FormInput v-model="usage.ukuran_jalan_napas" label="Ukuran Jalan Napas" /><FormInput v-model="usage.kedalaman_jalan_napas" label="Kedalaman Jalan Napas (cm)" type="number" min="0" max="999.99" step="0.1" inputmode="decimal" />
            <FormInput v-model="usage.indikasi" class="wide" label="Indikasi" jenis="textarea" :rows="3" required /><FormInput v-model="usage.catatan" class="wide" label="Catatan" jenis="textarea" :rows="3" />
          </div>
          <footer class="cppt-form-actions"><button type="submit" class="cppt-button primary"><LoaderCircle v-if="saving" class="spin" :size="15" /><Save v-else :size="15" />{{ saving ? 'Menyimpan...' : 'Mulai Pemakaian' }}</button></footer>
        </fieldset>
      </form>

      <template v-else>
        <div class="active-device"><i><Activity :size="21" /></i><div><small>Sedang Digunakan</small><strong>{{ active.nama_ventilator }}</strong><span>{{ active.kode_ventilator }} · Sejak {{ formatDateTime(active.tanggal_mulai) }}</span></div><dl><div><dt>Ruangan</dt><dd>{{ active.ruangan || '-' }}</dd></div><div><dt>Jalan Napas</dt><dd>{{ active.jenis_jalan_napas || '-' }} {{ active.ukuran_jalan_napas || '' }}</dd></div><div><dt>Status</dt><dd>{{ active.status }}</dd></div></dl></div>
        <div v-show="formVisible" class="cppt-form">
          <nav class="vent-tabs">
            <button v-for="item in [{ key: 'setting', label: 'Setting', icon: Settings2 }, { key: 'monitoring', label: 'Monitoring', icon: Gauge }, { key: 'vap', label: 'Checklist VAP', icon: ShieldCheck }]" :key="item.key" type="button" :class="{ active: tab === item.key }" @click="tab = item.key"><component :is="item.icon" :size="15" />{{ item.label }}</button>
          </nav>
          <form @submit.prevent="saveClinical"><fieldset class="form-compact" :disabled="saving || data.billing_terkunci">
            <div v-if="tab === 'setting'" class="ventilator-clinical-grid">
              <FormInput v-model="setting.waktu_setting" label="Waktu Setting" type="datetime-local" required /><FormInput v-model="setting.mode" label="Mode Ventilator" jenis="select" :options="modeOptions" required />
              <FormInput v-model="setting.fio2" label="FiO2 (%)" type="number" min="0" max="100" /><FormInput v-model="setting.peep" label="PEEP (cmH2O)" type="number" min="0" step="0.1" />
              <FormInput v-model="setting.tidal_volume" label="Tidal Volume (mL)" type="number" min="0" /><FormInput v-model="setting.frekuensi_set" label="Frekuensi Set (/mnt)" type="number" min="0" />
              <FormInput v-model="setting.pressure_control" label="Pressure Control (cmH2O)" type="number" min="0" /><FormInput v-model="setting.pressure_support" label="Pressure Support (cmH2O)" type="number" min="0" />
              <FormInput v-model="setting.catatan" class="wide" label="Catatan Setting" jenis="textarea" :rows="2" />
            </div>
            <div v-else-if="tab === 'monitoring'" class="ventilator-clinical-grid">
              <FormInput v-model="monitoring.waktu_monitoring" label="Waktu Monitoring" type="datetime-local" required /><FormInput v-model="monitoring.tahap" label="Tahap" jenis="select" :options="stageOptions" />
              <FormInput v-model="monitoring.kesadaran" label="Kesadaran" jenis="select" :options="awarenessOptions" /><FormInput v-model="monitoring.tekanan_darah" label="Tekanan Darah (mmHg)" placeholder="120/80" />
              <FormInput v-model="monitoring.nadi" label="Nadi (/mnt)" type="number" min="0" /><FormInput v-model="monitoring.respirasi" label="Respirasi (/mnt)" type="number" min="0" />
              <FormInput v-model="monitoring.suhu" label="Suhu (°C)" type="number" min="0" step="0.1" /><FormInput v-model="monitoring.spo2" label="SpO2 (%)" type="number" min="0" max="100" />
              <FormInput v-model="monitoring.catatan" class="wide" label="Catatan Monitoring" jenis="textarea" :rows="2" />
            </div>
            <div v-else class="vap-form-grid">
              <FormInput v-model="vap.waktu_checklist" label="Waktu Checklist" type="datetime-local" required /><FormInput v-model="vap.tekanan_cuff" label="Tekanan Cuff (cmH2O)" type="number" min="0" step="0.1" />
              <div class="vap-checks wide"><label v-for="item in [['elevasi_kepala', 'Elevasi kepala 30-45°'], ['perawatan_mulut', 'Perawatan mulut'], ['suction', 'Suction sesuai indikasi'], ['evaluasi_sedasi', 'Evaluasi sedasi'], ['sat', 'Spontaneous Awakening Trial'], ['sbt', 'Spontaneous Breathing Trial'], ['pencegahan_dvt', 'Pencegahan DVT'], ['pencegahan_ulkus', 'Pencegahan ulkus stres']]" :key="item[0]"><input v-model="vap[item[0]]" type="checkbox" /><span>{{ item[1] }}</span></label></div>
              <FormInput v-model="vap.catatan" class="wide" label="Catatan Checklist" jenis="textarea" :rows="2" />
            </div>
            <footer class="cppt-form-actions"><button type="submit" class="cppt-button primary"><LoaderCircle v-if="saving" class="spin" :size="15" /><Save v-else :size="15" />{{ saving ? 'Menyimpan...' : `Simpan ${tab === 'vap' ? 'Checklist VAP' : tab === 'setting' ? 'Setting' : 'Monitoring'}` }}</button></footer>
          </fieldset></form>
        </div>
      </template>
    </article>

    <article class="cppt-history-card ventilator-history">
      <header class="cppt-section-header"><div><span>Riwayat Monitoring</span><h3>Catatan Ventilator</h3><p>Setting dan pemantauan dapat dicatat berkali-kali selama alat digunakan.</p></div><TableSearch v-if="active" v-model="keyword" placeholder="Cari riwayat ventilator..." :total="historyTab === 'setting' ? data.setting.length : historyTab === 'monitoring' ? data.monitoring.length : data.checklist_vap.length" :filtered="historyRows.length" /></header>
      <nav class="history-tabs"><button v-for="item in [{ key: 'setting', label: `Setting (${data.setting.length})` }, { key: 'monitoring', label: `Monitoring (${data.monitoring.length})` }, { key: 'vap', label: `Checklist VAP (${data.checklist_vap.length})` }]" :key="item.key" type="button" :class="{ active: historyTab === item.key }" @click="historyTab = item.key; keyword = ''">{{ item.label }}</button></nav>
      <DataTable v-if="historyTab === 'setting'" :rows="historyRows" data-key="id" empty-message="Belum ada riwayat setting ventilator.">
        <PrimeColumn header="WAKTU" style="min-width:145px"><template #body="{ data: row }"><div class="cppt-table-main"><strong>{{ formatDateTime(row.waktu_setting) }}</strong><span>{{ row.petugas || '-' }}</span></div></template></PrimeColumn><PrimeColumn field="mode" header="MODE" style="min-width:110px" />
        <PrimeColumn header="OKSIGENASI" style="min-width:160px"><template #body="{ data: row }"><div class="cppt-table-vitals"><span>FiO2 <b>{{ row.fio2 }}%</b></span><span>PEEP <b>{{ row.peep }}</b></span></div></template></PrimeColumn>
        <PrimeColumn header="VENTILASI" style="min-width:210px"><template #body="{ data: row }"><div class="cppt-table-vitals"><span>TV <b>{{ row.tidal_volume }} mL</b></span><span>Frekuensi <b>{{ row.frekuensi_set }}</b></span><span>PC <b>{{ row.pressure_control }}</b></span><span>PS <b>{{ row.pressure_support }}</b></span></div></template></PrimeColumn>
        <PrimeColumn header="CATATAN" style="min-width:260px"><template #body="{ data: row }"><p class="cppt-table-note">{{ row.catatan || '-' }}</p></template></PrimeColumn>
      </DataTable>
      <DataTable v-else-if="historyTab === 'monitoring'" :rows="historyRows" data-key="id" empty-message="Belum ada riwayat monitoring ventilator.">
        <PrimeColumn header="WAKTU" style="min-width:145px"><template #body="{ data: row }"><div class="cppt-table-main"><strong>{{ formatDateTime(row.waktu_monitoring) }}</strong><span>{{ row.petugas || '-' }}</span></div></template></PrimeColumn><PrimeColumn field="tahap" header="TAHAP" style="min-width:110px" />
        <PrimeColumn header="KESADARAN / TD" style="min-width:170px"><template #body="{ data: row }"><div class="cppt-table-main"><strong>{{ row.kesadaran || '-' }}</strong><span>{{ row.tekanan_darah || '-' }} mmHg</span></div></template></PrimeColumn>
        <PrimeColumn header="TANDA VITAL" style="min-width:260px"><template #body="{ data: row }"><div class="cppt-table-vitals"><span>Nadi <b>{{ row.nadi }}</b></span><span>RR <b>{{ row.respirasi }}</b></span><span>Suhu <b>{{ row.suhu }}°C</b></span><span>SpO2 <b>{{ row.spo2 }}%</b></span></div></template></PrimeColumn>
        <PrimeColumn header="CATATAN" style="min-width:280px"><template #body="{ data: row }"><p class="cppt-table-note">{{ row.catatan || '-' }}</p></template></PrimeColumn>
      </DataTable>
      <DataTable v-else :rows="historyRows" data-key="id" empty-message="Belum ada checklist pencegahan VAP.">
        <PrimeColumn header="WAKTU" style="min-width:145px"><template #body="{ data: row }"><div class="cppt-table-main"><strong>{{ formatDateTime(row.waktu_checklist) }}</strong><span>{{ row.petugas || '-' }}</span></div></template></PrimeColumn>
        <PrimeColumn header="BUNDLE DASAR" style="min-width:250px"><template #body="{ data: row }"><div class="cppt-table-vitals"><span>Elevasi <b>{{ yes(row.elevasi_kepala) }}</b></span><span>Oral care <b>{{ yes(row.perawatan_mulut) }}</b></span><span>Suction <b>{{ yes(row.suction) }}</b></span><span>Sedasi <b>{{ yes(row.evaluasi_sedasi) }}</b></span></div></template></PrimeColumn>
        <PrimeColumn header="WEANING / PROFILAKSIS" style="min-width:260px"><template #body="{ data: row }"><div class="cppt-table-vitals"><span>SAT <b>{{ yes(row.sat) }}</b></span><span>SBT <b>{{ yes(row.sbt) }}</b></span><span>DVT <b>{{ yes(row.pencegahan_dvt) }}</b></span><span>Ulkus <b>{{ yes(row.pencegahan_ulkus) }}</b></span></div></template></PrimeColumn>
        <PrimeColumn header="CUFF / CATATAN" style="min-width:300px"><template #body="{ data: row }"><div class="cppt-table-extra"><p><b>Tekanan cuff:</b> {{ row.tekanan_cuff || 0 }} cmH2O</p><p><b>Catatan:</b> {{ row.catatan || '-' }}</p></div></template></PrimeColumn>
      </DataTable>
    </article>

    <article class="cppt-history-card usage-history">
      <header class="cppt-section-header"><div><span>Riwayat Terapi</span><h3>Pemakaian Ventilator</h3><p>{{ data.pemakaian.length }} pemakaian ditemukan</p></div><ClipboardCheck :size="24" /></header>
      <DataTable :rows="data.pemakaian" data-key="id" empty-message="Belum ada pemakaian ventilator.">
        <PrimeColumn header="MULAI / SELESAI" style="min-width:190px"><template #body="{ data: row }"><div class="cppt-table-main"><strong>{{ formatDateTime(row.tanggal_mulai) }}</strong><span>{{ row.tanggal_selesai ? formatDateTime(row.tanggal_selesai) : 'Masih digunakan' }}</span></div></template></PrimeColumn>
        <PrimeColumn header="PERANGKAT" style="min-width:190px"><template #body="{ data: row }"><div class="cppt-table-main"><strong>{{ row.nama_ventilator }}</strong><span>{{ row.kode_ventilator }}</span></div></template></PrimeColumn>
        <PrimeColumn header="JALAN NAPAS" style="min-width:160px"><template #body="{ data: row }"><div class="cppt-table-main"><strong>{{ row.jenis_jalan_napas || '-' }}</strong><span>Ukuran {{ row.ukuran_jalan_napas || '-' }} · Kedalaman {{ row.kedalaman_jalan_napas || '-' }}</span></div></template></PrimeColumn>
        <PrimeColumn header="INDIKASI" style="min-width:280px"><template #body="{ data: row }"><p class="cppt-table-note">{{ row.indikasi || '-' }}</p></template></PrimeColumn><PrimeColumn field="status" header="STATUS" style="min-width:100px" />
      </DataTable>
    </article>
  </section>
</template>

<style scoped>
.ventilator-page{display:grid;gap:16px}.ventilator-usage-grid{display:grid;grid-template-columns:repeat(3,minmax(0,1fr));gap:14px}.ventilator-clinical-grid,.vap-form-grid{display:grid;grid-template-columns:repeat(4,minmax(0,1fr));gap:14px}.wide{grid-column:span 2}.active-device{display:grid;grid-template-columns:46px minmax(220px,1fr) minmax(360px,1.5fr);align-items:center;gap:14px;margin:16px 20px;padding:15px;border:1px solid var(--line);border-radius:13px;background:var(--surface)}.active-device>i{display:grid;width:42px;height:42px;place-items:center;border-radius:12px;color:#0d9488;background:rgba(20,184,166,.12)}.active-device>div{display:grid;gap:2px}.active-device small,.active-device dt{color:var(--muted);font-size:9px;font-weight:800;text-transform:uppercase}.active-device strong{font-size:14px}.active-device span{color:var(--muted);font-size:10px}.active-device dl{display:grid;grid-template-columns:repeat(3,minmax(0,1fr));gap:10px;margin:0}.active-device dl div{padding:8px 10px;border-left:1px solid var(--line)}.active-device dd{margin:3px 0 0;font-size:11px;font-weight:800}.vent-tabs,.history-tabs{display:flex;gap:7px;padding:14px 20px;border-bottom:1px solid var(--line)}.vent-tabs button,.history-tabs button{display:inline-flex;min-height:36px;align-items:center;gap:6px;padding:0 12px;border:1px solid var(--line);border-radius:9px;color:var(--muted);background:var(--surface);font-size:10px;font-weight:800;cursor:pointer}.vent-tabs button.active,.history-tabs button.active{border-color:#0d9488;color:#fff;background:#0d9488}.vap-checks{display:grid;grid-template-columns:repeat(4,minmax(0,1fr));gap:8px}.vap-checks label{display:flex;min-height:42px;align-items:center;gap:8px;padding:9px 11px;border:1px solid var(--line);border-radius:10px;color:var(--text);background:var(--surface);font-size:10px;font-weight:700}.vap-checks input{accent-color:#0d9488}.billing-lock{margin:14px 20px 0;padding:11px 13px;border:1px solid rgba(245,158,11,.35);border-radius:10px;color:#b45309;background:rgba(245,158,11,.09);font-size:10px;font-weight:750}:global(.theme-dark) .active-device,:global(.theme-dark) .vap-checks label,:global(.sirapi-dark) .active-device,:global(.sirapi-dark) .vap-checks label{border-color:#405269;background:#1c283a}:global(.theme-dark) .billing-lock,:global(.sirapi-dark) .billing-lock{color:#fbbf24}.ventilator-history :deep(.cppt-table-note),.usage-history :deep(.cppt-table-note){width:300px}.cppt-section-tools>svg,.usage-history .cppt-section-header>svg{color:#0d9488}@media(max-width:1100px){.ventilator-usage-grid{grid-template-columns:repeat(2,minmax(0,1fr))}.ventilator-clinical-grid,.vap-form-grid{grid-template-columns:repeat(2,minmax(0,1fr))}.active-device{grid-template-columns:46px 1fr}.active-device dl{grid-column:1/-1}.vap-checks{grid-template-columns:repeat(2,minmax(0,1fr))}}@media(max-width:680px){.ventilator-usage-grid,.ventilator-clinical-grid,.vap-form-grid{grid-template-columns:1fr}.wide{grid-column:auto}.active-device{grid-template-columns:1fr}.active-device dl{grid-template-columns:1fr}.active-device dl div{border-left:0;border-top:1px solid var(--line)}.vent-tabs,.history-tabs{overflow-x:auto}.vap-checks{grid-template-columns:1fr}}
</style>
