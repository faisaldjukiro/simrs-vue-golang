<script setup>
import { AlertTriangle, Check, ClipboardCheck, LoaderCircle, Pencil, RotateCcw, Save, Trash2 } from '@lucide/vue'
import { computed, reactive, ref, watch } from 'vue'
import CariPetugas from '../../Components/Ui/CariPetugas.vue'
import FormInput from '../../Components/Ui/FormInput.vue'
import { hapusTriaseIgd, simpanTriaseIgd, triaseIgdData, ubahTriaseIgd } from '../../lib/faisal/api'
import { useNotifikasi } from '../../lib/shared/useNotifikasi'

const props = defineProps({
  token: { type: String, required: true },
  patient: { type: Object, required: true },
})

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
</script>

<template>
  <section class="triage-page">
    <article class="triage-khanza-form">
      <div v-if="loading" class="triage-state"><LoaderCircle class="spin" :size="26"/><strong>Menarik data triase IGD...</strong></div>

      <div v-if="!loading && billingLocked" class="handling-lock">Billing sudah terverifikasi atau kunjungan dibatalkan. Form Triase IGD hanya dapat dilihat.</div>
      <form v-if="!loading" @submit.prevent="save">
        <fieldset :disabled="saving || billingLocked">
          <section class="triage-khanza-arrival">
            <header>
              <strong>Data Triase IGD</strong>
              <div>
                <button type="button" class="secondary" @click="fillForm(activeType)"><RotateCcw :size="14"/> Batal</button>
                <button type="submit" class="primary"><LoaderCircle v-if="saving" class="spin" :size="14"/><Save v-else :size="14"/>{{ saving ? 'Menyimpan...' : 'Simpan' }}</button>
              </div>
            </header>
            <div class="triage-khanza-arrival-grid">
              <FormInput v-model="form.tanggal_kunjungan" label="Tgl. Kunjungan" type="datetime-local" step="1" required/>
              <FormInput v-model="form.cara_masuk" label="Cara Masuk" jenis="select" :options="caraMasukOptions" required/>
              <FormInput v-model="form.alat_transportasi" label="Transportasi" jenis="select" :options="transportOptions" required/>
              <FormInput v-model="form.alasan_kedatangan" label="Alasan Kedatangan" jenis="select" :options="reasonOptions" required/>
              <FormInput v-if="data.mendukung_hand_over" v-model="form.hand_over" label="Hand Over Tim Jaga" jenis="select" :options="handOverOptions" required/>
              <FormInput v-model="form.kode_kasus" class="case-field" label="Macam Kasus" jenis="select" :options="caseOptions" required/>
              <FormInput v-model="form.keterangan_kedatangan" class="arrival-note" label="Keterangan Kedatangan" maxlength="100" required/>
            </div>
          </section>

          <section class="triage-khanza-clinical">
            <nav class="triage-khanza-tabs" aria-label="Jenis triase">
              <button type="button" :class="{ active: activeType === 'primer' }" @click="fillForm('primer')">Triase Primer <Check v-if="data.triase?.primer" :size="12"/></button>
              <button type="button" :class="{ active: activeType === 'sekunder' }" @click="fillForm('sekunder')">Triase Sekunder <Check v-if="data.triase?.sekunder" :size="12"/></button>
            </nav>

            <div class="triage-khanza-body">
              <div class="triage-khanza-left">
                <section class="triage-khanza-box triage-khanza-assessment">
                  <FormInput v-model="form.isi_utama" class="main-note" :label="activeType === 'primer' ? 'Keluhan Utama' : 'Anamnesa Singkat'" jenis="textarea" :rows="4" maxlength="400" required/>
                  <FormInput v-if="activeType === 'primer'" v-model="form.kebutuhan_khusus" class="full" label="Kebutuhan Khusus" jenis="select" :options="specialNeedsOptions"/>
                  <div class="triage-khanza-vitals">
                    <FormInput v-model="form.suhu" label="Suhu (°C)" maxlength="5" required/>
                    <FormInput v-model="form.nyeri" label="Nyeri" maxlength="5" required/>
                    <FormInput v-model="form.tekanan_darah" label="Tensi" maxlength="8" placeholder="120/80" required/>
                    <FormInput v-model="form.nadi" label="Nadi (/menit)" maxlength="3" required/>
                    <FormInput v-model="form.saturasi_o2" label="Saturasi O₂ (%)" maxlength="3" required/>
                    <FormInput v-model="form.pernapasan" label="Respirasi (/menit)" maxlength="3" required/>
                  </div>
                </section>

                <section class="triage-khanza-box triage-khanza-decision">
                  <FormInput v-model="form.catatan" label="Catatan" maxlength="100" required/>
                  <div class="triage-khanza-plan">
                    <strong>Plan / Keputusan</strong>
                    <div>
                      <label v-for="option in planOptions" :key="option.value" :class="{ selected: form.plan === option.value }">
                        <input v-model="form.plan" type="radio" :value="option.value"/>
                        <span>{{ option.label }}</span>
                      </label>
                    </div>
                  </div>
                  <FormInput v-model="form.tanggal_triase" label="Tgl. Triase" type="datetime-local" step="1" required/>
                  <CariPetugas v-model="selectedOfficer" :token="token" label="Dokter / Petugas IGD" required/>
                </section>
              </div>

              <section :class="['triage-khanza-examination', `scale-${form.skala}`]">
                <header>
                  <strong>Pemeriksaan &amp; Skala Triase</strong>
                  <div class="triage-khanza-scales">
                    <button v-for="scale in scales" :key="scale" type="button" :class="[`scale-${scale}`, { active: form.skala === scale }]" @click="chooseScale(scale)">Skala {{ scale }}: {{ data.kriteria_skala.filter((item) => item.skala === scale).length }}</button>
                  </div>
                </header>
                <h4>Skala {{ form.skala }} - {{ scaleNames[form.skala] }}</h4>
                <div class="triage-khanza-criteria">
                  <section v-for="group in criteriaGroups" :key="group.kode">
                    <h5>{{ group.nama }}</h5>
                    <label v-for="item in group.items" :key="`${group.kode}-${item.kode}`" :class="{ selected: form.kode_kriteria.includes(item.kode) }">
                      <input type="checkbox" :checked="form.kode_kriteria.includes(item.kode)" @change="toggleCriterion(item.kode)"/>
                      <span>{{ item.pengkajian }}</span>
                    </label>
                  </section>
                  <p v-if="criteriaGroups.length === 0">Belum ada kriteria Skala {{ form.skala }} pada master SIMRS Khanza.</p>
                </div>
                <footer><b>{{ selectedCriteria.length }}</b> kriteria dipilih</footer>
              </section>
            </div>
          </section>
        </fieldset>
      </form>
    </article>

    <article class="triage-history-card">
      <header class="triage-section-header"><div><span>Riwayat Triase Pasien</span><h3>Hasil Triase IGD</h3><p>{{ existingRecords.length }} pengkajian tersimpan.</p></div></header>
      <div v-if="!loading && existingRecords.length === 0" class="triage-state"><ClipboardCheck :size="28"/><strong>Belum ada data triase untuk kunjungan ini.</strong></div>
      <dl v-if="data.triase" class="triage-common-summary">
        <div><dt>Kunjungan</dt><dd>{{ data.triase.tanggal_kunjungan }}</dd></div>
        <div><dt>Kasus</dt><dd>{{ data.triase.nama_kasus || data.triase.kode_kasus }}</dd></div>
        <div><dt>Kedatangan</dt><dd>{{ data.triase.cara_masuk }} · {{ data.triase.alat_transportasi }}</dd></div>
        <div><dt>Tanda Vital</dt><dd>TD {{ data.triase.tekanan_darah }} · N {{ data.triase.nadi }} · RR {{ data.triase.pernapasan }} · T {{ data.triase.suhu }}°C · SpO₂ {{ data.triase.saturasi_o2 }}% · Nyeri {{ data.triase.nyeri }}</dd></div>
      </dl>
      <div class="triage-records">
        <section v-for="record in existingRecords" :key="record.jenis" :class="['triage-record', `scale-${record.bagian.skala}`]">
          <header><span><b>{{ record.label }} · Skala {{ record.bagian.skala }}</b><small>{{ scaleNames[record.bagian.skala] }} · {{ record.bagian.plan }}</small></span><div v-if="!billingLocked"><button type="button" title="Edit triase" @click="fillForm(record.jenis)"><Pencil :size="14"/> Edit</button><button type="button" class="danger" title="Hapus triase" @click="deleteType = record.jenis"><Trash2 :size="14"/> Hapus</button></div><small v-else>Terkunci billing</small></header>
          <dl><div><dt>{{ record.jenis === 'primer' ? 'Keluhan Utama' : 'Anamnesa Singkat' }}</dt><dd>{{ record.bagian.isi_utama }}</dd></div><div><dt>Catatan</dt><dd>{{ record.bagian.catatan }}</dd></div><div><dt>Petugas Triase</dt><dd>{{ record.bagian.nama_petugas || record.bagian.nip }}<small>{{ record.bagian.jabatan_petugas }}</small></dd></div><div><dt>Tanggal Triase</dt><dd>{{ record.bagian.tanggal_triase }}</dd></div></dl>
          <div class="triage-record-criteria"><span v-for="item in record.bagian.kriteria_terpilih" :key="item.kode"><b>{{ item.kode }}</b>{{ item.pengkajian }}</span></div>
        </section>
      </div>
    </article>

    <div v-if="deleteType" class="triage-confirm-backdrop" @click.self="deleteType = ''"><section role="dialog" aria-modal="true"><AlertTriangle :size="26"/><h3>Hapus Triase {{ deleteType === 'primer' ? 'Primer' : 'Sekunder' }}?</h3><p>Data pengkajian dan seluruh kriteria skala yang dipilih akan dihapus dari SIMRS Khanza.</p><div><button type="button" :disabled="deleting" @click="deleteType = ''">Batal</button><button type="button" class="danger" :disabled="deleting" @click="confirmDelete"><LoaderCircle v-if="deleting" class="spin" :size="14"/><Trash2 v-else :size="14"/>{{ deleting ? 'Menghapus...' : 'Hapus' }}</button></div></section></div>
  </section>
</template>
