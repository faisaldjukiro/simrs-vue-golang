<script setup>
import { ChevronDown, ChevronUp, LoaderCircle, RotateCcw, Save, Search, Trash2 } from '@lucide/vue'
import { computed, reactive, ref, watch } from 'vue'
import Dialog from 'primevue/dialog'
import CariPetugas from '../../Components/Ui/CariPetugas.vue'
import FormInput from '../../Components/Ui/FormInput.vue'
import { awalKeperawatanIgdData, hapusAwalKeperawatanIgd, simpanAwalKeperawatanIgd, ubahAwalKeperawatanIgd } from '../../lib/faisal/api'
import { useNotifikasi } from '../../lib/shared/useNotifikasi'

const props = defineProps({ token: { type: String, required: true }, patient: { type: Object, required: true } })
const notifikasi = useNotifikasi()
const loading = ref(false)
const saving = ref(false)
const deleting = ref(false)
const confirmDelete = ref(false)
const formOpen = ref(true)
const problemSearch = ref('')
const planSearch = ref('')
const data = reactive({ petugas: {}, masalah_keperawatan: [], rencana_keperawatan: [], penilaian: null })
const petugas = ref({ nip: '', nama: '', jabatan: '' })
const form = reactive(emptyForm())

const editing = computed(() => Boolean(data.penilaian))
const selectedProblems = computed(() => new Set(form.kode_masalah))
const visibleProblems = computed(() => {
  const keyword = problemSearch.value.trim().toLocaleLowerCase('id')
  if (!keyword) return data.masalah_keperawatan
  return data.masalah_keperawatan.filter((item) => `${item.kode} ${item.nama}`.toLocaleLowerCase('id').includes(keyword))
})
const visiblePlans = computed(() => {
  if (selectedProblems.value.size === 0) return []
  const keyword = planSearch.value.trim().toLocaleLowerCase('id')
  return data.rencana_keperawatan.filter((item) => selectedProblems.value.has(item.kode_masalah) && (!keyword || `${item.kode} ${item.nama}`.toLocaleLowerCase('id').includes(keyword)))
})
const painScores = ['0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '10']

const options = {
  informasi: list(['Autoanamnesis', 'Alloanamnesis']), kehamilan: list(['Tidak Hamil', 'Hamil']), tekanan: list(['TAK', 'Sakit Kepala', 'Muntah', 'Pusing', 'Bingung']),
  pupil: list(['Normal', 'Miosis', 'Isokor', 'Anisokor']), neuro: list(['TAK', 'Spasme Otot', 'Perubahan Sensorik', 'Perubahan Motorik', 'Perubahan Bentuk Ekstremitas', 'Penurunan Tingkat Kesadaran', 'Fraktur/Dislokasi', 'Luksasio', 'Kerusakan Jaringan/Luka']),
  integumen: list(['TAK', 'Luka Bakar', 'Luka Robek', 'Lecet', 'Luka Decubitus', 'Luka Gangren']), turgor: list(['Baik', 'Menurun']), edema: list(['Tidak Ada', 'Ekstremitas', 'Seluruh Tubuh', 'Asites', 'Palpebrae']), mukosa: list(['Lembab', 'Kering']),
  ada: list(['Tidak Ada', 'Ada']), intoksikasi: list(['Tidak Ada', 'Ada', 'Gigitan Binatang', 'Zat Kimia', 'Gas', 'Obat']), psikologis: list(['Tidak Ada Masalah', 'Marah', 'Takut', 'Depresi', 'Cepat Lelah', 'Cemas', 'Gelisah', 'Lain-lain']),
  yaTidak: list(['Tidak', 'Ya']), perilaku: list(['Perilaku Kekerasan', 'Gangguan Efek', 'Gangguan Memori', 'Halusinasi', 'Kecenderungan Percobaan Bunuh Diri', 'Lainnya']), hubungan: list(['Harmonis', 'Kurang Harmonis', 'Tidak Harmonis', 'Konflik Besar']),
  tinggal: list(['Sendiri', 'Orang Tua', 'Suami / Istri', 'Lainnya']), pendidikan: list(['-', 'TS', 'TK', 'SD', 'SMP', 'SMA', 'SLTA/SEDERAJAT', 'D1', 'D2', 'D3', 'D4', 'S1', 'S2', 'S3']), edukasi: list(['Pasien', 'Keluarga']),
  kemampuan: list(['Mandiri', 'Bantuan Minimal', 'Bantuan Sebagian', 'Ketergantungan Total']), aktifitas: list(['Tirah Baring', 'Duduk', 'Berjalan']), nyeri: list(['Tidak Ada Nyeri', 'Nyeri Akut', 'Nyeri Kronis']), provokes: list(['Proses Penyakit', 'Benturan', 'Lain-lain']),
  quality: list(['Seperti Tertusuk', 'Berdenyut', 'Teriris', 'Tertindih', 'Tertiban', 'Lain-lain']), skala: list(['0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '10']), hilang: list(['Istirahat', 'Medengar Musik', 'Minum Obat']),
  risiko: list(['Tidak beresiko (tidak ditemukan a dan b)', 'Resiko rendah (ditemukan a/b)', 'Resiko tinggi (ditemukan a dan b)']),
}

function list(items) { return items.map((value) => ({ label: value, value })) }
function now() { const d = new Date(Date.now() - new Date().getTimezoneOffset() * 60000); return d.toISOString().slice(0, 19) }
function inputDate(value) { return String(value || '').replace(' ', 'T').slice(0, 19) }
function emptyForm() { return {
  no_rawat: props.patient?.no_rawat || '', tanggal: now(), informasi: 'Autoanamnesis', keluhan_utama: '', rpd: '', rpo: '', status_kehamilan: 'Tidak Hamil', gravida: '', para: '', abortus: '', hpht: '',
  tekanan: 'TAK', pupil: 'Normal', neurosensorik: 'TAK', integumen: 'TAK', turgor: 'Baik', edema: 'Tidak Ada', mukosa: 'Lembab', perdarahan: 'Tidak Ada', jumlah_perdarahan: '', warna_perdarahan: '', intoksikasi: 'Tidak Ada',
  bab: '', xbab: '', kbab: '', wbab: '', bak: '', xbak: '', wbak: '', lbak: '', psikologis: 'Tidak Ada Masalah', jiwa: 'Tidak', perilaku: 'Perilaku Kekerasan', dilaporkan: '', sebutkan: '', hubungan: 'Harmonis',
  tinggal_dengan: 'Sendiri', ket_tinggal: '', budaya: 'Tidak Ada', ket_budaya: '', pendidikan_pj: '-', ket_pendidikan_pj: '', edukasi: 'Pasien', ket_edukasi: '', kemampuan: 'Mandiri', aktifitas: 'Tirah Baring', alat_bantu: 'Tidak', ket_bantu: '',
  nyeri: 'Tidak Ada Nyeri', provokes: 'Proses Penyakit', ket_provokes: '', quality: 'Seperti Tertusuk', ket_quality: '', lokasi: '', menyebar: 'Tidak', skala_nyeri: '0', durasi: '', nyeri_hilang: 'Istirahat', ket_nyeri: '', pada_dokter: 'Tidak', ket_dokter: '',
  berjalan_a: 'Tidak', berjalan_b: 'Tidak', berjalan_c: 'Tidak', hasil: 'Tidak beresiko (tidak ditemukan a dan b)', lapor: 'Tidak', ket_lapor: '', rencana: '', nip: '', kode_masalah: [], kode_rencana: [],
} }

function fillForm() {
  const record = data.penilaian
  Object.assign(form, emptyForm(), record || {}, { no_rawat: props.patient.no_rawat, tanggal: inputDate(record?.tanggal) || now(), kode_masalah: [...(record?.kode_masalah || [])], kode_rencana: [...(record?.kode_rencana || [])] })
  petugas.value = record ? { nip: record.nip, nama: record.nama_petugas, jabatan: record.jabatan_petugas } : { ...(data.petugas || {}) }
}

async function loadData() {
  if (!props.token || !props.patient.no_rawat) return
  loading.value = true
  try { const response = await awalKeperawatanIgdData(props.token, props.patient.no_rawat); Object.assign(data, response || {}); fillForm(); formOpen.value = true }
  catch (error) { notifikasi.gagal(error.message || 'Penilaian awal keperawatan IGD tidak dapat dibaca.') }
  finally { loading.value = false }
}

function payload() { return { ...form, tanggal: form.tanggal.replace('T', ' '), nip: petugas.value?.nip || '', kode_masalah: [...form.kode_masalah], kode_rencana: [...form.kode_rencana] } }
async function save() {
  if (!petugas.value?.nip) { notifikasi.peringatan('Pilih petugas yang melakukan pengkajian.'); return }
  saving.value = true
  try { const response = editing.value ? await ubahAwalKeperawatanIgd(props.token, payload()) : await simpanAwalKeperawatanIgd(props.token, payload()); notifikasi.sukses(response?.pesan || 'Penilaian berhasil disimpan.'); await loadData() }
  catch (error) { notifikasi.gagal(error.message || 'Penilaian gagal disimpan.') }
  finally { saving.value = false }
}
async function remove() {
  deleting.value = true
  try { const response = await hapusAwalKeperawatanIgd(props.token, props.patient.no_rawat); confirmDelete.value = false; notifikasi.sukses(response?.pesan || 'Penilaian berhasil dihapus.'); await loadData() }
  catch (error) { notifikasi.gagal(error.message || 'Penilaian gagal dihapus.') }
  finally { deleting.value = false }
}
function toggle(array, value) { const index = array.indexOf(value); if (index >= 0) array.splice(index, 1); else array.push(value) }
function toggleProblem(kode) {
  const index = form.kode_masalah.indexOf(kode)
  if (index >= 0) {
    form.kode_masalah.splice(index, 1)
    const related = new Set(data.rencana_keperawatan.filter((item) => item.kode_masalah === kode).map((item) => item.kode))
    form.kode_rencana = form.kode_rencana.filter((item) => !related.has(item))
    return
  }
  form.kode_masalah.push(kode)
}
watch(() => props.patient.no_rawat, loadData, { immediate: true })
</script>

<template>
  <section class="initial-nursing-page">
    <div v-if="loading" class="initial-nursing-state"><LoaderCircle class="spin" :size="28"/><strong>Menarik penilaian awal keperawatan IGD...</strong></div>
    <template v-else>
      <form class="initial-nursing-form cppt-form-card cppt-form" @submit.prevent="save">
        <header class="cppt-section-header">
          <div><span>PENGKAJIAN KEPERAWATAN</span><h3>{{ editing ? 'Edit' : 'Input' }} Awal Keperawatan IGD</h3><p>Pengkajian menyeluruh pasien saat menerima pelayanan IGD.</p></div>
          <div class="form-header-actions">
            <button v-if="editing" type="button" class="cppt-button danger" :disabled="deleting" @click="confirmDelete = true"><Trash2 :size="15"/>Hapus</button>
            <button type="button" class="collapse cppt-button toggle icon-only" :title="formOpen ? 'Sembunyikan form input' : 'Tampilkan form input'" @click="formOpen = !formOpen"><ChevronUp v-if="formOpen" :size="18"/><ChevronDown v-else :size="18"/></button>
          </div>
        </header>

        <fieldset v-show="formOpen" class="form-compact" :disabled="saving">
          <div class="nursing-sheet">
            <section class="assessment-strip assessment-meta">
              <div class="nursing-grid intro-grid">
              <FormInput v-model="form.tanggal" label="Tanggal" type="datetime-local" step="1" required/>
              <FormInput v-model="form.informasi" label="Informasi Didapat Dari" jenis="select" :options="options.informasi" required/>
              <CariPetugas v-model="petugas" class="staff-field" :token="token" label="Petugas" required/>
              </div>
            </section>

            <section class="assessment-strip">
              <h4><span>I.</span> Riwayat Kesehatan Pasien</h4>
              <div class="nursing-grid history-grid">
              <FormInput v-model="form.keluhan_utama" class="wide" label="Riwayat Penyakit Sekarang" jenis="textarea" :rows="3" required/>
              <FormInput v-model="form.rpd" label="Riwayat Penyakit Dahulu" jenis="textarea" :rows="3" required/>
              <FormInput v-model="form.rpo" label="Riwayat Penggunaan Obat" jenis="textarea" :rows="3" required/>
              <div class="pregnancy-fields">
                <FormInput v-model="form.status_kehamilan" label="Status Kehamilan" jenis="select" :options="options.kehamilan"/>
                <FormInput v-model="form.hpht" label="HPHT"/><FormInput v-model="form.para" label="Para"/><FormInput v-model="form.abortus" label="Abortus"/><FormInput v-model="form.gravida" label="Gravida"/>
              </div>
              </div>
            </section>

            <section class="assessment-strip">
              <h4><span>II.</span> Pemeriksaan Fisik</h4>
              <div class="nursing-grid physical-grid">
              <FormInput v-model="form.tekanan" label="Tekanan Intrakranial" jenis="select" :options="options.tekanan"/><FormInput v-model="form.pupil" label="Pupil" jenis="select" :options="options.pupil"/>
              <FormInput v-model="form.neurosensorik" label="Neurosensorik / Muskuloskeletal" jenis="select" :options="options.neuro"/><FormInput v-model="form.integumen" label="Integumen" jenis="select" :options="options.integumen"/>
              <FormInput v-model="form.turgor" label="Turgor Kulit" jenis="select" :options="options.turgor"/><FormInput v-model="form.edema" label="Edema" jenis="select" :options="options.edema"/><FormInput v-model="form.mukosa" label="Mukosa Mulut" jenis="select" :options="options.mukosa"/>
              <FormInput v-model="form.perdarahan" label="Perdarahan" jenis="select" :options="options.ada"/><FormInput v-model="form.jumlah_perdarahan" label="Jumlah (cc)"/><FormInput v-model="form.warna_perdarahan" label="Warna"/>
              <FormInput v-model="form.intoksikasi" label="Intoksikasi" jenis="select" :options="options.intoksikasi"/>
              </div>
              <h5 class="assessment-subtitle">Eliminasi</h5>
              <div class="nursing-grid elimination-grid"><FormInput v-model="form.bab" label="BAB · Frekuensi"/><FormInput v-model="form.xbab" label="BAB · x/"/><FormInput v-model="form.kbab" label="BAB · Konsistensi"/><FormInput v-model="form.wbab" label="BAB · Warna"/><FormInput v-model="form.bak" label="BAK · Frekuensi"/><FormInput v-model="form.xbak" label="BAK · x/"/><FormInput v-model="form.wbak" label="BAK · Warna"/><FormInput v-model="form.lbak" label="BAK · Lain-lain"/></div>
            </section>

            <section class="assessment-strip">
              <h4><span>III.</span> Riwayat Psikologis · Sosial · Ekonomi · Budaya · Spiritual</h4>
              <div class="nursing-grid psycho-grid"><FormInput v-model="form.psikologis" label="a. Kondisi Psikologis" jenis="select" :options="options.psikologis"/><FormInput v-model="form.jiwa" label="b. Gangguan Jiwa Di Masa Lalu" jenis="select" :options="options.yaTidak"/><FormInput v-model="form.perilaku" label="d. Adakah Perilaku" jenis="select" :options="options.perilaku"/><FormInput v-model="form.dilaporkan" label="Dilaporkan Ke"/><FormInput v-model="form.sebutkan" label="Sebutkan"/><FormInput v-model="form.hubungan" label="e. Hubungan Pasien Dengan Anggota Keluarga" jenis="select" :options="options.hubungan"/><FormInput v-model="form.tinggal_dengan" label="h. Tinggal Dengan" jenis="select" :options="options.tinggal"/><FormInput v-model="form.ket_tinggal" label="Keterangan Tinggal Dengan"/><FormInput v-model="form.budaya" label="l. Kepercayaan / Budaya / Nilai-nilai Khusus Yang Perlu Diperhatikan" jenis="select" :options="options.ada"/><FormInput v-model="form.ket_budaya" label="Keterangan Kepercayaan / Budaya"/><FormInput v-model="form.pendidikan_pj" label="m. Pendidikan P.J." jenis="select" :options="options.pendidikan"/><FormInput v-model="form.ket_pendidikan_pj" label="Keterangan Pendidikan P.J."/><FormInput v-model="form.edukasi" label="n. Edukasi Diberikan Kepada" jenis="select" :options="options.edukasi"/><FormInput v-model="form.ket_edukasi" label="Keterangan Edukasi Diberikan Kepada"/></div>
            </section>

            <section class="assessment-strip">
              <h4><span>IV.</span> Pengkajian Fungsi</h4>
              <div class="nursing-grid function-grid"><FormInput v-model="form.kemampuan" label="a. Kemampuan Aktifitas Sehari-hari" jenis="select" :options="options.kemampuan"/><FormInput v-model="form.aktifitas" label="b. Aktifitas" jenis="select" :options="options.aktifitas"/><FormInput v-model="form.alat_bantu" label="d. Alat Bantu" jenis="select" :options="options.yaTidak"/><FormInput v-model="form.ket_bantu" label="Keterangan Alat Bantu"/></div>
            </section>

            <section class="assessment-strip">
              <h4><span>V.</span> Skala Nyeri</h4>
              <div class="pain-assessment-layout">
                <figure class="pain-scale-guide"><img src="/img/skala-nyeri-khanza.png" alt="Panduan Wong-Baker Faces Pain Rating Scale dari SIMRS Khanza"><figcaption>Pilih angka skala nyeri pasien</figcaption><div><button v-for="score in painScores" :key="score" type="button" :class="{ active: form.skala_nyeri === score }" :aria-pressed="form.skala_nyeri === score" @click="form.skala_nyeri = score">{{ score }}</button></div></figure>
                <div class="nursing-grid pain-fields"><FormInput v-model="form.nyeri" label="Jenis Nyeri" jenis="select" :options="options.nyeri"/><FormInput v-model="form.provokes" label="Penyebab" jenis="select" :options="options.provokes"/><FormInput v-model="form.ket_provokes" label="Keterangan Penyebab"/><FormInput v-model="form.quality" label="Kualitas" jenis="select" :options="options.quality"/><FormInput v-model="form.ket_quality" label="Keterangan Kualitas"/><FormInput v-model="form.lokasi" label="Lokasi"/><FormInput v-model="form.menyebar" label="Menyebar" jenis="select" :options="options.yaTidak"/><FormInput v-model="form.skala_nyeri" label="Severity · Skala Nyeri" jenis="select" :options="options.skala"/><FormInput v-model="form.durasi" label="Waktu / Durasi (Menit)"/><FormInput v-model="form.nyeri_hilang" label="Nyeri Hilang Bila" jenis="select" :options="options.hilang"/><FormInput v-model="form.ket_nyeri" label="Keterangan Nyeri Hilang"/><FormInput v-model="form.pada_dokter" label="Diberitahukan Pada Dokter?" jenis="select" :options="options.yaTidak"/><FormInput v-model="form.ket_dokter" label="Jam"/></div>
              </div>
            </section>

            <section class="assessment-strip">
              <h4><span>VI.</span> Penilaian Resiko Jatuh (Get Up and Go)</h4>
              <div class="nursing-grid risk-grid"><FormInput v-model="form.berjalan_a" label="1. Tidak Seimbang / Sempoyongan / Limbung" jenis="select" :options="options.yaTidak"/><FormInput v-model="form.berjalan_c" label="2. Jalan dengan Menggunakan Alat Bantu" jenis="select" :options="options.yaTidak"/><FormInput v-model="form.berjalan_b" label="b. Menopang Saat Akan Duduk" jenis="select" :options="options.yaTidak"/><FormInput v-model="form.hasil" label="Hasil" jenis="select" :options="options.risiko"/><FormInput v-model="form.lapor" label="Dilaporkan Kepada Dokter?" jenis="select" :options="options.yaTidak"/><FormInput v-model="form.ket_lapor" label="Jam Dilaporkan"/></div>
            </section>

            <section class="assessment-strip nursing-plan">
              <h4>Masalah & Rencana Keperawatan</h4>
              <div class="plan-columns">
                <div class="plan-master">
                  <div class="plan-tools">
                    <h5>Masalah Keperawatan</h5>
                    <div class="plan-search"><Search :size="15"/><input v-model="problemSearch" type="search" placeholder="Cari masalah keperawatan..."></div>
                  </div>
                  <p v-if="visibleProblems.length === 0" class="empty-master">Masalah keperawatan tidak ditemukan.</p>
                  <label v-for="item in visibleProblems" :key="item.kode"><input type="checkbox" :checked="form.kode_masalah.includes(item.kode)" @change="toggleProblem(item.kode)"><span><b>{{ item.kode }}</b>{{ item.nama }}</span></label>
                </div>
                <div class="plan-master">
                  <div class="plan-tools">
                    <h5>Rencana Keperawatan</h5>
                    <div class="plan-search"><Search :size="15"/><input v-model="planSearch" type="search" placeholder="Cari rencana keperawatan..." :disabled="selectedProblems.size === 0"></div>
                  </div>
                  <p v-if="selectedProblems.size === 0" class="empty-master">Pilih masalah keperawatan terlebih dahulu.</p>
                  <p v-else-if="visiblePlans.length === 0" class="empty-master">Rencana keperawatan tidak ditemukan.</p>
                  <label v-for="item in visiblePlans" :key="item.kode"><input type="checkbox" :checked="form.kode_rencana.includes(item.kode)" @change="toggle(form.kode_rencana,item.kode)"><span><b>{{ item.kode }}</b>{{ item.nama }}</span></label>
                </div>
              </div>
              <FormInput v-model="form.rencana" label="Catatan Rencana Keperawatan" jenis="textarea" :rows="3"/>
            </section>
          </div>

          <footer class="cppt-form-actions"><button type="button" class="cppt-button secondary" @click="fillForm"><RotateCcw :size="15"/>Batal / Reset</button><button type="submit" class="cppt-button primary"><LoaderCircle v-if="saving" class="spin" :size="15"/><Save v-else :size="15"/>{{ saving ? 'Menyimpan...' : editing ? 'Simpan Perubahan' : 'Simpan Penilaian' }}</button></footer>
        </fieldset>
      </form>

      <Dialog v-model:visible="confirmDelete" modal header="Hapus Penilaian" class="initial-nursing-dialog" :style="{ width: 'min(430px, 92vw)' }">
        <p>Penilaian awal keperawatan IGD pasien ini akan dihapus dari SIMRS Khanza.</p>
        <template #footer><button type="button" class="secondary" @click="confirmDelete = false">Batal</button><button type="button" class="danger" :disabled="deleting" @click="remove"><LoaderCircle v-if="deleting" class="spin" :size="14"/><Trash2 v-else :size="14"/>{{ deleting ? 'Menghapus...' : 'Hapus' }}</button></template>
      </Dialog>
    </template>
  </section>
</template>

<style scoped>
.initial-nursing-page{display:grid;gap:16px}.initial-nursing-state,.initial-nursing-empty{min-height:260px;display:grid;place-content:center;justify-items:center;gap:10px;color:var(--text-muted)}.initial-nursing-form,.initial-nursing-result{border:1px solid var(--border);border-radius:16px;background:var(--surface);overflow:visible}.initial-nursing-form>header,.initial-nursing-result>header{display:flex;align-items:center;justify-content:space-between;gap:16px;padding:18px 20px;border-bottom:1px solid var(--border)}header span{color:var(--primary);font-size:.69rem;font-weight:700;letter-spacing:.12em}header h3{margin:4px 0;font-size:1.12rem}header p{margin:0;color:var(--text-muted);font-size:.8rem}.collapse,.primary,.secondary,.danger{display:inline-flex;align-items:center;justify-content:center;gap:7px;border:1px solid var(--border);border-radius:10px;min-height:38px;padding:0 14px;font-weight:650;cursor:pointer}.collapse,.primary{background:var(--primary);color:#fff;border-color:transparent}.secondary{background:var(--surface);color:var(--text)}.danger{background:color-mix(in srgb,#ef4444 10%,var(--surface));color:#dc2626;border-color:color-mix(in srgb,#ef4444 30%,var(--border))}.initial-nursing-form fieldset{border:0;margin:0;padding:14px;background:color-mix(in srgb,var(--primary) 3%,var(--surface));display:grid;gap:12px}.nursing-section{padding:14px;border:1px solid var(--border);border-radius:12px;background:var(--surface)}.nursing-section h4{margin:0 0 13px;font-size:.92rem;border-bottom:1px solid var(--border);padding-bottom:10px}.nursing-grid{display:grid;grid-template-columns:repeat(4,minmax(0,1fr));gap:10px}.intro-grid{grid-template-columns:240px 220px minmax(320px,1fr)}.intro-grid .wide{grid-column:1/-1}.intro-grid>:nth-last-child(-n+2){grid-column:span 1}.compact-grid{grid-template-columns:repeat(4,minmax(130px,1fr))}.compact-grid .wide{grid-column:span 2}.two-columns{display:grid;grid-template-columns:1fr 1fr;gap:18px}.two-columns>div+div{border-left:1px solid var(--border);padding-left:18px}.plan-columns{display:grid;grid-template-columns:1fr 1fr;gap:16px;margin-bottom:12px}.plan-columns>div{border:1px solid var(--border);border-radius:10px;overflow:hidden;max-height:310px;overflow-y:auto}.plan-columns h5{position:sticky;top:0;z-index:1;margin:0;padding:10px 12px;background:var(--surface-soft);border-bottom:1px solid var(--border)}.plan-columns label{display:flex;gap:9px;padding:9px 12px;border-bottom:1px solid var(--border);cursor:pointer;font-size:.82rem}.plan-columns label:hover{background:color-mix(in srgb,var(--primary) 6%,var(--surface))}.plan-columns input{accent-color:var(--primary)}.plan-columns span{display:flex;gap:8px}.plan-columns b{color:var(--primary);font-family:monospace}.empty-master{padding:14px;color:var(--text-muted)}fieldset>footer{display:flex;justify-content:flex-end;gap:9px;padding:5px 0}.initial-nursing-result>header>div:last-child{display:flex;gap:8px}.result-grid{display:grid;grid-template-columns:repeat(3,1fr);gap:1px;background:var(--border)}.result-grid article{padding:14px;background:var(--surface)}.result-grid .wide{grid-column:span 2}.result-grid span{display:block;color:var(--text-muted);font-size:.72rem;margin-bottom:5px}.result-grid strong{white-space:pre-wrap;font-size:.86rem}.result-grid p{margin:5px 0 0;color:var(--text-muted);font-size:.76rem}@media(max-width:1100px){.intro-grid,.nursing-grid{grid-template-columns:repeat(2,minmax(0,1fr))}.staff-field{grid-column:1/-1}.two-columns{grid-template-columns:1fr}.two-columns>div+div{border-left:0;border-top:1px solid var(--border);padding:16px 0 0}.plan-columns{grid-template-columns:1fr}}@media(max-width:680px){.intro-grid,.nursing-grid,.result-grid{grid-template-columns:1fr}.intro-grid>*,.compact-grid .wide,.result-grid .wide{grid-column:auto}.initial-nursing-form>header,.initial-nursing-result>header{align-items:flex-start}.initial-nursing-result>header{flex-direction:column}.plan-columns span{display:block}}
.plan-master>.plan-tools{position:sticky;top:0;z-index:4;background:var(--surface);box-shadow:0 1px 0 var(--border)}
.plan-master>.plan-tools h5{position:static;top:auto}
.plan-search{position:static;display:flex;align-items:center;gap:8px;margin:0;padding:8px 10px;background:var(--surface);border-bottom:1px solid var(--border);color:var(--text-muted)}
.plan-search input{width:100%;min-height:34px;padding:0 10px;border:1px solid var(--border);border-radius:8px;background:var(--surface-soft);color:var(--text);font:inherit;outline:none}
.plan-search input:focus{border-color:var(--primary);box-shadow:0 0 0 2px color-mix(in srgb,var(--primary) 14%,transparent)}
.plan-search input:disabled{cursor:not-allowed;opacity:.55}
.initial-nursing-page{--primary:#0d9488;--border:var(--line);--text-muted:var(--muted)}
.pain-scale-guide{display:grid;justify-items:center;gap:8px;margin:0 0 12px;padding:12px;border:1px solid var(--border);border-radius:12px;background:#fff}.pain-scale-guide img{display:block;width:min(468px,100%);height:auto}.pain-scale-guide figcaption{color:#475569;font-size:.74rem}.pain-scale-guide>div{display:grid;width:100%;grid-template-columns:repeat(11,minmax(28px,1fr));gap:4px}.pain-scale-guide button{min-height:30px;border:1px solid #cbd5e1;border-radius:7px;color:#334155;background:#f8fafc;font-size:.74rem;font-weight:650;cursor:pointer}.pain-scale-guide button:hover{border-color:#0d9488}.pain-scale-guide button.active{border-color:#0d9488;color:#fff;background:#0d9488;box-shadow:0 0 0 2px rgba(13,148,136,.14)}
.initial-nursing-form.cppt-form fieldset{padding:20px}.initial-nursing-form .nursing-section{border-color:#d5dee8;background:#fff;box-shadow:0 1px 4px rgba(15,23,42,.04)}.initial-nursing-form .nursing-section h4{color:#172033}.initial-nursing-form .cppt-form-actions{margin-top:6px;padding-top:17px}
:global(.theme-light .initial-nursing-form.cppt-form fieldset){background:#f6f8fb}
:global(.theme-dark .initial-nursing-form.cppt-form fieldset),:global(.sirava-dark .initial-nursing-form.cppt-form fieldset){background:linear-gradient(135deg,#111c2c 0%,#0e1b29 100%)}
:global(.theme-dark .initial-nursing-form .nursing-section),:global(.sirava-dark .initial-nursing-form .nursing-section){border-color:#304258;background:#142132;box-shadow:none}
:global(.theme-dark .initial-nursing-form .nursing-section h4),:global(.sirava-dark .initial-nursing-form .nursing-section h4){color:#f8fafc}
:global(.initial-nursing-dialog){--surface:#fff;--border:#dbe4ee;--text:#0f172a;color:var(--text);background:var(--surface)}
:global(body.sirava-dark .initial-nursing-dialog){--surface:#0f172a;--border:rgba(148,163,184,.18);--text:#f8fafc;color:#f8fafc;background:#0f172a}

/* Lembar pengkajian disusun berurutan seperti form Khanza. */
.nursing-sheet{overflow:hidden;border:1px solid var(--border);border-radius:12px;background:var(--surface)}
.assessment-strip{padding:13px 15px;border-bottom:1px solid var(--border)}
.assessment-strip:last-child{border-bottom:0}
.assessment-strip>h4{display:flex;align-items:center;gap:7px;margin:0 0 12px;color:var(--text);font-size:.79rem;font-weight:700;letter-spacing:.025em;text-transform:uppercase}
.assessment-strip>h4 span{min-width:22px;color:var(--primary);font-size:inherit;letter-spacing:0}
.assessment-meta{background:var(--surface-soft)}
.assessment-meta .intro-grid{grid-template-columns:220px 220px minmax(320px,1fr)}
.history-grid{grid-template-columns:1fr 1fr}
.history-grid>.wide,.history-grid>.pregnancy-fields{grid-column:1/-1}
.pregnancy-fields{display:grid;grid-template-columns:1.35fr repeat(4,minmax(100px,.65fr));gap:10px;padding-top:2px}
.physical-grid{grid-template-columns:repeat(4,minmax(145px,1fr))}
.assessment-subtitle{margin:13px 0 9px;padding-top:11px;border-top:1px dashed var(--border);color:var(--text-muted);font-size:.7rem;font-weight:700;letter-spacing:.08em;text-transform:uppercase}
.elimination-grid{grid-template-columns:repeat(4,minmax(130px,1fr))}
.psycho-grid{grid-template-columns:repeat(4,minmax(150px,1fr))}
.function-grid{grid-template-columns:repeat(4,minmax(150px,1fr))}
.pain-assessment-layout{display:grid;grid-template-columns:minmax(350px,.78fr) minmax(540px,1.22fr);gap:15px;align-items:start}
.pain-assessment-layout .pain-scale-guide{margin:0}
.pain-fields{grid-template-columns:repeat(3,minmax(145px,1fr))}
.risk-grid{grid-template-columns:repeat(3,minmax(180px,1fr))}
.assessment-strip.nursing-plan{border:0;border-radius:0;background:transparent;box-shadow:none}
.assessment-strip.nursing-plan>h4{padding-bottom:10px;border-bottom:1px solid var(--border)}
.initial-nursing-form .cppt-form-actions{margin-top:0;padding:14px 0 0}
.form-header-actions{display:flex;align-items:center;gap:8px}

:global(.theme-light .initial-nursing-form .nursing-sheet){background:#fff;border-color:#d8e1eb}
:global(.theme-light .initial-nursing-form .assessment-meta){background:#f5f8fb}
:global(.theme-light .initial-nursing-form .assessment-strip:nth-child(even)){background:#fbfcfd}
:global(.theme-dark .initial-nursing-form .nursing-sheet),:global(.sirava-dark .initial-nursing-form .nursing-sheet){background:#142132;border-color:#304258}
:global(.theme-dark .initial-nursing-form .assessment-strip),:global(.sirava-dark .initial-nursing-form .assessment-strip){border-color:#304258;background:#142132}
:global(.theme-dark .initial-nursing-form .assessment-meta),:global(.sirava-dark .initial-nursing-form .assessment-meta){background:#111d2c}
:global(.theme-dark .initial-nursing-form .assessment-strip>h4),:global(.sirava-dark .initial-nursing-form .assessment-strip>h4){color:#f8fafc}

@media(max-width:1200px){.assessment-meta .intro-grid{grid-template-columns:210px 210px 1fr}.physical-grid,.psycho-grid{grid-template-columns:repeat(3,minmax(140px,1fr))}.pain-assessment-layout{grid-template-columns:1fr}.pain-fields{grid-template-columns:repeat(3,minmax(140px,1fr))}}
@media(max-width:850px){.assessment-meta .intro-grid,.history-grid,.pregnancy-fields,.physical-grid,.elimination-grid,.psycho-grid,.function-grid,.pain-fields,.risk-grid{grid-template-columns:repeat(2,minmax(0,1fr))}.assessment-meta .staff-field,.history-grid>.wide,.history-grid>.pregnancy-fields{grid-column:1/-1}}
@media(max-width:560px){.assessment-meta .intro-grid,.history-grid,.pregnancy-fields,.physical-grid,.elimination-grid,.psycho-grid,.function-grid,.pain-fields,.risk-grid{grid-template-columns:1fr}.assessment-meta .staff-field,.history-grid>.wide,.history-grid>.pregnancy-fields{grid-column:auto}.assessment-strip{padding:12px}.pain-scale-guide>div{grid-template-columns:repeat(6,minmax(30px,1fr))}.form-header-actions{align-items:flex-end;flex-direction:column-reverse}}
</style>
