<script setup lang="ts">
import { ChevronDown, ChevronUp, Pencil, Printer, Save, Trash2 } from '@lucide/vue'
import Column from 'primevue/column'
import Dialog from 'primevue/dialog'
import FormInput from '../../../Components/Ui/FormInput.vue'
import InputPencarian from '../../../Components/Ui/InputPencarian.vue'
import DataTable from '../../../Components/Ui/DataTable.vue'
import { penilaianPemulangan, bantuanPemulangan } from '../../../types/perencanaanPemulangan'
import type { PropsPemulangan } from '../../../types/perencanaanPemulangan'
import { usePerencanaanPemulangan } from './usePerencanaanPemulangan'
const props = defineProps<PropsPemulangan>()
const {
  loading, saving, error, errorSimpan, terbuka, keyword, rows, editing,
  hapusTarget, petugas, bolehPilih, form, terkunci,
  reset, cariPetugas, edit, simpan, cetak,
} = usePerencanaanPemulangan(props)
const pilihanYa = [{ label: 'Tidak', value: 'Tidak' }, { label: 'Ya', value: 'Ya' }]
const pilihanBantuan = bantuanPemulangan.map(value => ({ label: value, value }))
</script>

<template>
  <section class="clinical-page pemulangan-page">
    <article class="clinical-form-card">
      <header class="clinical-section-header">
        <div>
          <span>Pelayanan Rawat Inap</span>
          <h3>{{ editing ? 'Edit Perencanaan Pemulangan' : 'Perencanaan Pemulangan Pasien' }}</h3>
          <p>Satu perencanaan per kunjungan. Data disimpan langsung ke SIMRS.</p>
        </div>
        <button type="button" class="clinical-button secondary" :aria-expanded="terbuka" @click="terbuka = !terbuka">
          <ChevronUp v-if="terbuka" :size="16" />
          <ChevronDown v-else :size="16" />
          {{ terbuka ? 'Tutup Form' : 'Buka Form' }}
        </button>
      </header>
      <form v-show="terbuka" class="clinical-form" @submit.prevent="simpan()">
        <fieldset class="form-compact" :disabled="terkunci">
          <div class="pemulangan-grid">
            <FormInput v-model="form.rencana_pulang" type="date" label="Rencana Pulang" required />
            <InputPencarian v-model="petugas" label="Petugas" :search="cariPetugas" :disabled="!bolehPilih || terkunci" required />
            <FormInput v-model="form.diagnosa_medis" label="Diagnosis Medis" :maxlength="50" required />
            <FormInput v-model="form.alasan_masuk" label="Alasan Masuk / Dirawat" :maxlength="150" required />
          </div>
          <section class="pemulangan-penilaian">
            <h4>Kebutuhan dan Persiapan Pemulangan</h4>
            <div v-for="(p, index) in penilaianPemulangan" :key="p.key" class="pemulangan-item">
              <FormInput v-model="form[p.key]" :label="(index + 1) + '. ' + p.label" jenis="select"
                :options="p.key === 'bantuan_diperlukan_dalam' ? pilihanBantuan : pilihanYa"
                :disabled="terkunci" />
              <FormInput v-model="form['keterangan_' + p.key]" label="Keterangan" :maxlength="100" />
            </div>
          </section>
          <FormInput v-model="form.nama_pasien_keluarga" label="Nama Pasien / Keluarga" :maxlength="50" required />
          <p v-if="errorSimpan" class="patient-error" role="alert">{{ errorSimpan }}</p>
          <footer class="clinical-form-actions">
            <button type="button" class="clinical-button secondary" :disabled="saving" @click="reset">Batal / Reset</button>
            <button type="submit" class="clinical-button primary" :disabled="terkunci">
              <Save :size="16" /> {{ saving ? 'Menyimpan...' : editing ? 'Simpan Perubahan' : 'Simpan Perencanaan' }}
            </button>
          </footer>
        </fieldset>
      </form>
    </article>
    <article class="clinical-history-card">
      <header class="clinical-section-header">
        <div>
          <span>Kunjungan {{ patient.no_rawat }}</span>
          <h3>Riwayat Perencanaan Pemulangan</h3>
        </div>
        <div class="clinical-section-tools">
          <FormInput v-model="keyword" label="Pencarian" placeholder="Cari diagnosis, tanggal, atau catatan..." />
        </div>
      </header>
      <p v-if="error" class="patient-error" role="alert">{{ error }}</p>
      <DataTable v-else :rows="rows" :loading="loading" data-key="data.no_rawat" empty-message="Belum ada perencanaan pemulangan.">
        <Column header="Rencana Pulang">
          <template #body="{ data: r }">{{ r.data.rencana_pulang }}</template>
        </Column>
        <Column header="Diagnosis / Alasan Masuk" style="min-width:230px">
          <template #body="{ data: r }">
            <div class="clinical-table-main">
              <strong>{{ r.data.diagnosa_medis }}</strong>
              <small>{{ r.data.alasan_masuk }}</small>
            </div>
          </template>
        </Column>
        <Column header="Pasien / Keluarga">
          <template #body="{ data: r }">{{ r.data.nama_pasien_keluarga }}</template>
        </Column>
        <Column field="nama_petugas" header="Petugas" />
        <Column header="Aksi" style="min-width:250px">
          <template #body="{ data: r }">
            <div class="clinical-table-actions">
              <button v-if="r.bisa_ubah" type="button" :disabled="terkunci" @click="edit(r)"><Pencil :size="14" /> Edit / Detail</button>
              <button type="button" :disabled="loading" @click="cetak(r)"><Printer :size="14" /> Cetak</button>
              <button v-if="r.bisa_ubah" type="button" class="danger" :disabled="terkunci" @click="hapusTarget = r; errorSimpan = ''"><Trash2 :size="14" /> Hapus</button>
            </div>
          </template>
        </Column>
      </DataTable>
    </article>
    <Dialog :visible="!!hapusTarget" modal header="Hapus Perencanaan Pemulangan?" :closable="!saving"
      :style="{ width: '480px', maxWidth: '95vw' }" @update:visible="!saving && (hapusTarget = null)">
      <p>Perencanaan kunjungan ini akan dihapus dari SIMRS. Tindakan ini tidak dapat dibatalkan.</p>
      <p v-if="errorSimpan" class="patient-error">{{ errorSimpan }}</p>
      <template #footer>
        <button class="clinical-button secondary" :disabled="saving" @click="hapusTarget = null">Batal</button>
        <button class="clinical-button danger" :disabled="saving" @click="simpan(true)">{{ saving ? 'Menghapus...' : 'Ya, Hapus' }}</button>
      </template>
    </Dialog>
  </section>
</template>
<style src="./perencanaan-pemulangan.css" scoped></style>
