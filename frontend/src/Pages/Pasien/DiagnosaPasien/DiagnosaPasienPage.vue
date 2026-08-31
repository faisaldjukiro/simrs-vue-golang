<script setup lang="ts">
import { ArrowDown, ArrowUp, FileCheck2, LoaderCircle, Save, ShieldAlert, Trash2 } from "@lucide/vue"
import InputPencarian from "../../../Components/Ui/InputPencarian.vue"
import { useDiagnosaPasien } from "./useDiagnosaPasien"

const props = defineProps({
  token: { type: String, required: true },
  patient: { type: Object, required: true },
  moduleName: { type: String, required: true },
})

const {
  loading,
  saving,
  deletingKey,
  billingLocked,
  diagnosa,
  prosedur,
  diagnosaBaru,
  prosedurBaru,
  jumlahCoding,
  jumlahBaru,
  cariCoding,
  tambahCoding,
  hapusCoding,
  pindahCoding,
  simpan,
} = useDiagnosaPasien(props)
</script>

<template>
  <section class="patient-coding-page">
    <article class="patient-coding-card">
      <header class="patient-coding-header">
        <div>
          <span>CODING PASIEN</span>
          <h3>Diagnosa & Prosedur</h3>
          <p>Data disimpan langsung ke diagnosa_pasien dan prosedur_pasien SIMRS Khanza.</p>
        </div>
        <strong><FileCheck2 :size="16" /> {{ jumlahCoding }} coding</strong>
      </header>

      <div v-if="billingLocked" class="patient-coding-lock">
        <ShieldAlert :size="17" />
        <span><strong>Billing sudah terverifikasi.</strong> Diagnosa dan prosedur hanya dapat dilihat.</span>
      </div>

      <div v-if="loading" class="patient-coding-state">
        <LoaderCircle class="spin" :size="24" />
        <span>Memuat diagnosa dan prosedur pasien...</span>
      </div>

      <div v-else class="patient-coding-content">
        <section class="patient-coding-section">
          <header>
            <div><span>ICD-10</span><h4>Diagnosa</h4></div>
            <small>{{ diagnosa.length }} diagnosa</small>
          </header>

          <div class="patient-coding-list">
            <article v-for="(item, index) in diagnosa" :key="`diagnosa-${item.kode}`">
              <code>{{ item.kode }}</code>
              <div><strong>{{ item.nama }}</strong><span>{{ index === 0 ? 'Diagnosa utama' : `Diagnosa sekunder ${index}` }}</span></div>
              <i v-if="!item.tersimpan">Belum disimpan</i>
              <i v-else>{{ item.status_penyakit || 'Baru' }}<template v-if="item.im === '1'"> · IM</template></i>
              <div class="patient-coding-actions">
                <button type="button" title="Naikkan urutan pilihan baru" :disabled="billingLocked || item.tersimpan || index === 0 || diagnosa[index - 1]?.tersimpan" @click="pindahCoding('diagnosa', index, -1)"><ArrowUp :size="14" /></button>
                <button type="button" title="Turunkan urutan pilihan baru" :disabled="billingLocked || item.tersimpan || index === diagnosa.length - 1" @click="pindahCoding('diagnosa', index, 1)"><ArrowDown :size="14" /></button>
                <button type="button" class="danger" title="Hapus diagnosa" :disabled="billingLocked || deletingKey === `diagnosa-${item.kode}`" @click="hapusCoding('diagnosa', index)"><LoaderCircle v-if="deletingKey === `diagnosa-${item.kode}`" class="spin" :size="14" /><Trash2 v-else :size="14" /></button>
              </div>
            </article>
            <p v-if="diagnosa.length === 0">Belum ada diagnosa pasien.</p>
          </div>

          <InputPencarian
            v-model="diagnosaBaru"
            label="Tambah Diagnosa (ICD-10)"
            placeholder="Ketik minimal 3 karakter kode atau nama diagnosa..."
            :search="(kataKunci) => cariCoding('diagnosa', kataKunci)"
            right-field="penanda"
            :disabled="billingLocked || saving"
            @update:model-value="tambahCoding('diagnosa', $event)"
          />
        </section>

        <section class="patient-coding-section">
          <header>
            <div><span>ICD-9-CM</span><h4>Prosedur / Tindakan</h4></div>
            <small>{{ prosedur.length }} prosedur</small>
          </header>

          <div class="patient-coding-list">
            <article v-for="(item, index) in prosedur" :key="`prosedur-${item.kode}`" class="procedure">
              <code>{{ item.kode }}</code>
              <div><strong>{{ item.nama }}</strong><span>{{ index === 0 ? 'Prosedur utama' : `Prosedur sekunder ${index}` }}</span></div>
              <label class="patient-coding-quantity">
                <span>Jml</span>
                <input v-model.number="item.jumlah" type="number" min="1" max="999" :disabled="billingLocked || item.tersimpan" />
              </label>
              <i v-if="!item.tersimpan">Belum disimpan</i>
              <i v-else-if="item.im === '1'">IM</i>
              <div class="patient-coding-actions">
                <button type="button" title="Naikkan urutan pilihan baru" :disabled="billingLocked || item.tersimpan || index === 0 || prosedur[index - 1]?.tersimpan" @click="pindahCoding('prosedur', index, -1)"><ArrowUp :size="14" /></button>
                <button type="button" title="Turunkan urutan pilihan baru" :disabled="billingLocked || item.tersimpan || index === prosedur.length - 1" @click="pindahCoding('prosedur', index, 1)"><ArrowDown :size="14" /></button>
                <button type="button" class="danger" title="Hapus prosedur" :disabled="billingLocked || deletingKey === `prosedur-${item.kode}`" @click="hapusCoding('prosedur', index)"><LoaderCircle v-if="deletingKey === `prosedur-${item.kode}`" class="spin" :size="14" /><Trash2 v-else :size="14" /></button>
              </div>
            </article>
            <p v-if="prosedur.length === 0">Belum ada prosedur pasien.</p>
          </div>

          <InputPencarian
            v-model="prosedurBaru"
            label="Tambah Prosedur (ICD-9-CM)"
            placeholder="Ketik minimal 3 karakter kode atau nama prosedur..."
            :search="(kataKunci) => cariCoding('prosedur', kataKunci)"
            right-field="penanda"
            :disabled="billingLocked || saving"
            @update:model-value="tambahCoding('prosedur', $event)"
          />
        </section>

        <footer v-if="!billingLocked" class="patient-coding-footer">
          <button type="button" :disabled="saving" @click="simpan">
            <LoaderCircle v-if="saving" class="spin" :size="15" />
            <Save v-else :size="15" />
            {{ saving ? 'Menyimpan...' : `Simpan ${jumlahBaru} Pilihan Baru` }}
          </button>
        </footer>
      </div>
    </article>
  </section>
</template>

<style src="./diagnosa-pasien.css"></style>
