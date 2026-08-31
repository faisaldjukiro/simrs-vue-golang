<script setup lang="ts">
import { AlertTriangle, Check, ClipboardCheck, LoaderCircle, Pencil, RotateCcw, Save, Trash2 } from "@lucide/vue"
import CariPetugas from "../../../Components/Ui/CariPetugas.vue"
import FormInput from "../../../Components/Ui/FormInput.vue"
import { useTriaseIgd } from "./useTriaseIgd"

const props = defineProps({
  token: { type: String, required: true },
  patient: { type: Object, required: true },
})

const {
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
} = useTriaseIgd(props)
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
                  <div :class="['triage-khanza-plan', activeType === 'primer' ? 'plan-primer' : 'plan-sekunder']">
                    <strong>Plan / Keputusan</strong>
                    <small>{{ planNote }}</small>
                    <div>
                      <label
                        v-for="option in planOptions"
                        :key="option.value"
                        :class="[{ selected: form.plan === option.value }, `plan-${option.value.toLowerCase().replaceAll(' ', '-')}`]"
                      >
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
