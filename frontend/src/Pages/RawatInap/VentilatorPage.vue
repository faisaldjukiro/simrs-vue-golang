<script setup>
import { Activity, ChevronDown, ChevronUp, ClipboardCheck, Gauge, LoaderCircle, Save, Settings2, ShieldCheck, Wind } from "@lucide/vue"
import PrimeColumn from "primevue/column"
import CariDokter from "../../Components/Ui/CariDokter.vue"
import CariPetugas from "../../Components/Ui/CariPetugas.vue"
import DataTable from "../../Components/Ui/DataTable.vue"
import FormInput from "../../Components/Ui/FormInput.vue"
import TableSearch from "../../Components/Ui/TableSearch.vue"
import { useVentilator } from "./Ventilator/useVentilator.js"

const props = defineProps({ token: { type: String, required: true }, patient: { type: Object, required: true } })

const {
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
} = useVentilator(props)
</script>

<template>
  <section class="clinical-page ventilator-page">
    <article class="clinical-form-card">
      <header class="clinical-section-header">
        <div><span>Terapi Respirasi Rawat Inap</span><h3>{{ active ? 'Pencatatan Ventilator' : 'Mulai Pemakaian Ventilator' }}</h3><p>Setting, monitoring pasien, dan pencegahan ventilator-associated pneumonia.</p></div>
        <div class="clinical-section-tools"><Wind :size="24" /><button type="button" class="clinical-button toggle icon-only" :title="formVisible ? 'Sembunyikan Form Input' : 'Tampilkan Form Input'" @click="formVisible = !formVisible"><ChevronUp v-if="formVisible" :size="15" /><ChevronDown v-else :size="15" /></button></div>
      </header>
      <div v-if="data.billing_terkunci" class="billing-lock">Kunjungan sudah masuk billing. Data ventilator hanya dapat dilihat.</div>
      <div v-if="loading" class="clinical-state"><LoaderCircle class="spin" :size="25" /><strong>Menarik data ventilator...</strong></div>
      <div v-else-if="error" class="clinical-state error"><strong>{{ error }}</strong><button type="button" @click="load">Coba Lagi</button></div>

      <form v-else-if="!active" v-show="formVisible" class="clinical-form" @submit.prevent="saveUsage">
        <fieldset class="form-compact" :disabled="saving || data.billing_terkunci">
          <div class="ventilator-usage-grid">
            <FormInput v-model="usage.kode_ventilator" label="Perangkat Ventilator" jenis="select" :options="available" filter required />
            <FormInput v-model="usage.tanggal_mulai" label="Tanggal Mulai" type="datetime-local" required /><FormInput v-model="usage.ruangan" label="Ruangan" />
            <CariDokter v-model="doctor" :token="token" required /><CariPetugas v-model="officer" :token="token" label="Petugas Pemasangan" required />
            <FormInput v-model="usage.jenis_jalan_napas" label="Jalan Napas" jenis="select" :options="airwayOptions" required />
            <FormInput v-model="usage.ukuran_jalan_napas" label="Ukuran Jalan Napas" /><FormInput v-model="usage.kedalaman_jalan_napas" label="Kedalaman Jalan Napas (cm)" type="number" min="0" max="999.99" step="0.1" inputmode="decimal" />
            <FormInput v-model="usage.indikasi" class="wide" label="Indikasi" jenis="textarea" :rows="3" required /><FormInput v-model="usage.catatan" class="wide" label="Catatan" jenis="textarea" :rows="3" />
          </div>
          <footer class="clinical-form-actions"><button type="submit" class="clinical-button primary"><LoaderCircle v-if="saving" class="spin" :size="15" /><Save v-else :size="15" />{{ saving ? 'Menyimpan...' : 'Mulai Pemakaian' }}</button></footer>
        </fieldset>
      </form>

      <template v-else>
        <div class="active-device"><i><Activity :size="21" /></i><div><small>Sedang Digunakan</small><strong>{{ active.nama_ventilator }}</strong><span>{{ active.kode_ventilator }} · Sejak {{ formatDateTime(active.tanggal_mulai) }}</span></div><dl><div><dt>Ruangan</dt><dd>{{ active.ruangan || '-' }}</dd></div><div><dt>Jalan Napas</dt><dd>{{ active.jenis_jalan_napas || '-' }} {{ active.ukuran_jalan_napas || '' }}</dd></div><div><dt>Status</dt><dd>{{ active.status }}</dd></div></dl></div>
        <div v-show="formVisible" class="clinical-form">
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
            <footer class="clinical-form-actions"><button type="submit" class="clinical-button primary"><LoaderCircle v-if="saving" class="spin" :size="15" /><Save v-else :size="15" />{{ saving ? 'Menyimpan...' : `Simpan ${tab === 'vap' ? 'Checklist VAP' : tab === 'setting' ? 'Setting' : 'Monitoring'}` }}</button></footer>
          </fieldset></form>
        </div>
      </template>
    </article>

    <article class="clinical-history-card ventilator-history">
      <header class="clinical-section-header"><div><span>Riwayat Monitoring</span><h3>Catatan Ventilator</h3><p>Setting dan pemantauan dapat dicatat berkali-kali selama alat digunakan.</p></div><TableSearch v-if="active" v-model="keyword" placeholder="Cari riwayat ventilator..." :total="historyTab === 'setting' ? data.setting.length : historyTab === 'monitoring' ? data.monitoring.length : data.checklist_vap.length" :filtered="historyRows.length" /></header>
      <nav class="history-tabs"><button v-for="item in [{ key: 'setting', label: `Setting (${data.setting.length})` }, { key: 'monitoring', label: `Monitoring (${data.monitoring.length})` }, { key: 'vap', label: `Checklist VAP (${data.checklist_vap.length})` }]" :key="item.key" type="button" :class="{ active: historyTab === item.key }" @click="historyTab = item.key; keyword = ''">{{ item.label }}</button></nav>
      <DataTable v-if="historyTab === 'setting'" :rows="historyRows" data-key="id" empty-message="Belum ada riwayat setting ventilator.">
        <PrimeColumn header="WAKTU" style="min-width:145px"><template #body="{ data: row }"><div class="clinical-table-main"><strong>{{ formatDateTime(row.waktu_setting) }}</strong><span>{{ row.petugas || '-' }}</span></div></template></PrimeColumn><PrimeColumn field="mode" header="MODE" style="min-width:110px" />
        <PrimeColumn header="OKSIGENASI" style="min-width:160px"><template #body="{ data: row }"><div class="clinical-table-vitals"><span>FiO2 <b>{{ row.fio2 }}%</b></span><span>PEEP <b>{{ row.peep }}</b></span></div></template></PrimeColumn>
        <PrimeColumn header="VENTILASI" style="min-width:210px"><template #body="{ data: row }"><div class="clinical-table-vitals"><span>TV <b>{{ row.tidal_volume }} mL</b></span><span>Frekuensi <b>{{ row.frekuensi_set }}</b></span><span>PC <b>{{ row.pressure_control }}</b></span><span>PS <b>{{ row.pressure_support }}</b></span></div></template></PrimeColumn>
        <PrimeColumn header="CATATAN" style="min-width:260px"><template #body="{ data: row }"><p class="clinical-table-note">{{ row.catatan || '-' }}</p></template></PrimeColumn>
      </DataTable>
      <DataTable v-else-if="historyTab === 'monitoring'" :rows="historyRows" data-key="id" empty-message="Belum ada riwayat monitoring ventilator.">
        <PrimeColumn header="WAKTU" style="min-width:145px"><template #body="{ data: row }"><div class="clinical-table-main"><strong>{{ formatDateTime(row.waktu_monitoring) }}</strong><span>{{ row.petugas || '-' }}</span></div></template></PrimeColumn><PrimeColumn field="tahap" header="TAHAP" style="min-width:110px" />
        <PrimeColumn header="KESADARAN / TD" style="min-width:170px"><template #body="{ data: row }"><div class="clinical-table-main"><strong>{{ row.kesadaran || '-' }}</strong><span>{{ row.tekanan_darah || '-' }} mmHg</span></div></template></PrimeColumn>
        <PrimeColumn header="TANDA VITAL" style="min-width:260px"><template #body="{ data: row }"><div class="clinical-table-vitals"><span>Nadi <b>{{ row.nadi }}</b></span><span>RR <b>{{ row.respirasi }}</b></span><span>Suhu <b>{{ row.suhu }}°C</b></span><span>SpO2 <b>{{ row.spo2 }}%</b></span></div></template></PrimeColumn>
        <PrimeColumn header="CATATAN" style="min-width:280px"><template #body="{ data: row }"><p class="clinical-table-note">{{ row.catatan || '-' }}</p></template></PrimeColumn>
      </DataTable>
      <DataTable v-else :rows="historyRows" data-key="id" empty-message="Belum ada checklist pencegahan VAP.">
        <PrimeColumn header="WAKTU" style="min-width:145px"><template #body="{ data: row }"><div class="clinical-table-main"><strong>{{ formatDateTime(row.waktu_checklist) }}</strong><span>{{ row.petugas || '-' }}</span></div></template></PrimeColumn>
        <PrimeColumn header="BUNDLE DASAR" style="min-width:250px"><template #body="{ data: row }"><div class="clinical-table-vitals"><span>Elevasi <b>{{ yes(row.elevasi_kepala) }}</b></span><span>Oral care <b>{{ yes(row.perawatan_mulut) }}</b></span><span>Suction <b>{{ yes(row.suction) }}</b></span><span>Sedasi <b>{{ yes(row.evaluasi_sedasi) }}</b></span></div></template></PrimeColumn>
        <PrimeColumn header="WEANING / PROFILAKSIS" style="min-width:260px"><template #body="{ data: row }"><div class="clinical-table-vitals"><span>SAT <b>{{ yes(row.sat) }}</b></span><span>SBT <b>{{ yes(row.sbt) }}</b></span><span>DVT <b>{{ yes(row.pencegahan_dvt) }}</b></span><span>Ulkus <b>{{ yes(row.pencegahan_ulkus) }}</b></span></div></template></PrimeColumn>
        <PrimeColumn header="CUFF / CATATAN" style="min-width:300px"><template #body="{ data: row }"><div class="clinical-table-extra"><p><b>Tekanan cuff:</b> {{ row.tekanan_cuff || 0 }} cmH2O</p><p><b>Catatan:</b> {{ row.catatan || '-' }}</p></div></template></PrimeColumn>
      </DataTable>
    </article>

    <article class="clinical-history-card usage-history">
      <header class="clinical-section-header"><div><span>Riwayat Terapi</span><h3>Pemakaian Ventilator</h3><p>{{ data.pemakaian.length }} pemakaian ditemukan</p></div><ClipboardCheck :size="24" /></header>
      <DataTable :rows="data.pemakaian" data-key="id" empty-message="Belum ada pemakaian ventilator.">
        <PrimeColumn header="MULAI / SELESAI" style="min-width:190px"><template #body="{ data: row }"><div class="clinical-table-main"><strong>{{ formatDateTime(row.tanggal_mulai) }}</strong><span>{{ row.tanggal_selesai ? formatDateTime(row.tanggal_selesai) : 'Masih digunakan' }}</span></div></template></PrimeColumn>
        <PrimeColumn header="PERANGKAT" style="min-width:190px"><template #body="{ data: row }"><div class="clinical-table-main"><strong>{{ row.nama_ventilator }}</strong><span>{{ row.kode_ventilator }}</span></div></template></PrimeColumn>
        <PrimeColumn header="JALAN NAPAS" style="min-width:160px"><template #body="{ data: row }"><div class="clinical-table-main"><strong>{{ row.jenis_jalan_napas || '-' }}</strong><span>Ukuran {{ row.ukuran_jalan_napas || '-' }} · Kedalaman {{ row.kedalaman_jalan_napas || '-' }}</span></div></template></PrimeColumn>
        <PrimeColumn header="INDIKASI" style="min-width:280px"><template #body="{ data: row }"><p class="clinical-table-note">{{ row.indikasi || '-' }}</p></template></PrimeColumn><PrimeColumn field="status" header="STATUS" style="min-width:100px" />
      </DataTable>
    </article>
  </section>
</template>

<style src="./Ventilator/ventilator.css" scoped></style>
