<script setup lang="ts">
import { ChevronDown, ChevronUp, LoaderCircle, RotateCcw, Save, Search, Trash2 } from "@lucide/vue"
import Dialog from "primevue/dialog"
import CariPetugas from "../../../Components/Ui/CariPetugas.vue"
import FormInput from "../../../Components/Ui/FormInput.vue"
import { useAwalKeperawatanIgd } from "./useAwalKeperawatanIgd"

const props = defineProps({ token: { type: String, required: true }, patient: { type: Object, required: true } })

const {
  loading,
  saving,
  deleting,
  confirmDelete,
  formOpen,
  problemSearch,
  planSearch,
  petugas,
  form,
  editing,
  billingLocked,
  selectedProblems,
  visibleProblems,
  visiblePlans,
  painScores,
  options,
  fillForm,
  save,
  remove,
  toggle,
  toggleProblem,
} = useAwalKeperawatanIgd(props)
</script>

<template>
  <section class="initial-nursing-page">
    <div v-if="loading" class="initial-nursing-state"><LoaderCircle class="spin" :size="28"/><strong>Menarik penilaian awal keperawatan IGD...</strong></div>
    <template v-else>
      <form class="initial-nursing-form clinical-form-card clinical-form" @submit.prevent="save">
        <header class="clinical-section-header">
          <div><span>PENGKAJIAN KEPERAWATAN</span><h3>{{ editing ? 'Edit' : 'Input' }} Awal Keperawatan IGD</h3><p>Pengkajian menyeluruh pasien saat menerima pelayanan IGD.</p></div>
          <div class="form-header-actions">
            <button v-if="editing && !billingLocked" type="button" class="clinical-button danger" :disabled="deleting" @click="confirmDelete = true"><Trash2 :size="15"/>Hapus</button>
            <button type="button" class="collapse clinical-button toggle icon-only" :title="formOpen ? 'Sembunyikan form input' : 'Tampilkan form input'" @click="formOpen = !formOpen"><ChevronUp v-if="formOpen" :size="18"/><ChevronDown v-else :size="18"/></button>
          </div>
        </header>

        <div v-if="billingLocked" class="handling-lock">Billing sudah terverifikasi atau kunjungan dibatalkan. Form Awal Keperawatan IGD hanya dapat dilihat.</div>
        <fieldset v-show="formOpen" class="form-compact" :disabled="saving || billingLocked">
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

          <footer class="clinical-form-actions"><button type="button" class="clinical-button secondary" @click="fillForm"><RotateCcw :size="15"/>Batal / Reset</button><button type="submit" class="clinical-button primary"><LoaderCircle v-if="saving" class="spin" :size="15"/><Save v-else :size="15"/>{{ saving ? 'Menyimpan...' : editing ? 'Simpan Perubahan' : 'Simpan Penilaian' }}</button></footer>
        </fieldset>
      </form>

      <Dialog v-model:visible="confirmDelete" modal header="Hapus Penilaian" class="initial-nursing-dialog" :style="{ width: 'min(430px, 92vw)' }">
        <p>Penilaian awal keperawatan IGD pasien ini akan dihapus dari SIMRS Khanza.</p>
        <template #footer><button type="button" class="secondary" @click="confirmDelete = false">Batal</button><button type="button" class="danger" :disabled="deleting" @click="remove"><LoaderCircle v-if="deleting" class="spin" :size="14"/><Trash2 v-else :size="14"/>{{ deleting ? 'Menghapus...' : 'Hapus' }}</button></template>
      </Dialog>
    </template>
  </section>
</template>

<style src="./awal-keperawatan-igd.css" scoped></style>
