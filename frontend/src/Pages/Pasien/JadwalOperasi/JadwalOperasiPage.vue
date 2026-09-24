<script setup lang="ts">
import { ChevronDown, ChevronUp, Pencil, Printer, RefreshCw, Save, Trash2, X } from '@lucide/vue'
import { watch } from 'vue'
import Column from 'primevue/column'
import Dialog from 'primevue/dialog'
import DataTable from '../../../Components/Ui/DataTable.vue'
import FormInput from '../../../Components/Ui/FormInput.vue'
import InputPencarian from '../../../Components/Ui/InputPencarian.vue'
import { statusOperasi, type JadwalOperasi, type PropsJadwalOperasi } from '../../../types/jadwalOperasi'
import { useJadwalOperasi } from './useJadwalOperasi'

const props = defineProps<PropsJadwalOperasi>()
const emit = defineEmits<{
  'pilih-jadwal': [jadwal: JadwalOperasi]
  'jadwal-dimuat': [jadwal: JadwalOperasi[]]
}>()
const {
  loading, saving, error, errorSimpan, pesanKunci, peringatan, terkunci,
  formVisible, form, filter, paket, dokter, ruang, filteredRows, editing, hapusTarget,
  reset, muat, cari, edit, simpan, hapus, cetak, records,
} = useJadwalOperasi(props)
watch(records, rows => emit('jadwal-dimuat', rows))
</script>

<template>
  <section class="jadwal-operasi clinical-page">
    <article class="clinical-form-card">
      <header class="clinical-section-header">
        <div>
          <span>Pelayanan Pasien</span>
          <h3>Jadwal Operasi</h3>
          <p>Paket tindakan, ruang, dan tim operasi untuk kunjungan ini.</p>
        </div>
        <button
          class="clinical-button toggle"
          :aria-expanded="formVisible"
          @click="formVisible = !formVisible"
        >
          <ChevronUp v-if="formVisible" :size="16" />
          <ChevronDown v-else :size="16" />
          {{ formVisible ? 'Tutup Form' : 'Buka Form' }}
        </button>
      </header>
      <p class="jadwal-info">
        Jadwal baru langsung disimpan ke Khanza.
        Edit dan hapus langsung berlaku pada sumber data baris yang dipilih.
      </p>
      <p v-if="pesanKunci" class="jadwal-info" role="status">{{ pesanKunci }}</p>
      <form
        v-show="formVisible"
        :key="String(patient.no_rawat)"
        class="jadwal-form"
        @submit.prevent="simpan"
      >
        <p v-if="editing" class="jadwal-edit" role="status">
          Mengedit jadwal {{ editing.sumber }}: {{ editing.nama_paket }} — {{ editing.tanggal }}.
        </p>
        <div class="jadwal-fields">
          <InputPencarian
            v-model="paket"
            label="Paket Operasi"
            :search="q => cari('paket', q)"
            :disabled="terkunci"
            required
          />
          <InputPencarian
            v-model="dokter"
            label="Operator / Dokter"
            :search="q => cari('dokter', q)"
            :disabled="terkunci"
            required
          />
          <InputPencarian
            v-model="ruang"
            label="Ruang Operasi / OK"
            :search="q => cari('ruang', q)"
            :disabled="terkunci"
            required
          />
          <FormInput
            v-model="form.tanggal"
            label="Tanggal Operasi"
            type="date"
            :disabled="terkunci"
            required
          />
          <FormInput
            v-model="form.jam_mulai"
            label="Jam Mulai (WITA)"
            type="time"
            step="1"
            :disabled="terkunci"
            required
          />
          <FormInput
            v-model="form.jam_selesai"
            label="Jam Selesai (WITA)"
            type="time"
            step="1"
            :disabled="terkunci"
            required
          />
          <FormInput
            v-model="form.status"
            label="Status Jadwal"
            jenis="select"
            :options="statusOperasi"
            :disabled="terkunci"
            required
          />
          <FormInput
            v-model="form.dokteranastesi"
            label="Dokter Anestesi"
            :maxlength="255"
            :disabled="terkunci"
          />
          <FormInput
            v-model="form.perawat"
            label="Perawat"
            :maxlength="255"
            :disabled="terkunci"
          />
        </div>
        <p class="jadwal-help">
          Pilihan paket mengikuti kelas dan penjamin sesuai pengaturan tarif Khanza.
          Jam 00:00:00–00:00:00 berarti waktu belum ditentukan dan belum diperiksa bentroknya.
        </p>
        <p v-if="errorSimpan" class="patient-error" role="alert">{{ errorSimpan }}</p>
        <footer class="jadwal-actions">
          <button type="button" class="clinical-button secondary" :disabled="saving" @click="reset">
            <X :size="16" /> {{ editing ? 'Batal Edit' : 'Reset' }}
          </button>
          <button type="submit" class="clinical-button primary" :disabled="terkunci">
            <Save :size="16" />
            {{ saving ? 'Menyimpan...' : editing ? 'Simpan Perubahan' : 'Simpan ke Khanza' }}
          </button>
        </footer>
      </form>
    </article>

    <article class="clinical-history-card">
      <header class="clinical-section-header">
        <div>
          <span>Kunjungan {{ patient.no_rawat }}</span>
          <h3>Riwayat Jadwal Operasi</h3>
          <p v-if="!loading && !error">{{ filteredRows.length }} jadwal ditampilkan.</p>
        </div>
        <div class="jadwal-actions">
          <button class="clinical-button secondary" :disabled="loading || saving" @click="muat">
            <RefreshCw :size="16" /> Muat Ulang
          </button>
          <button
            class="clinical-button secondary"
            :disabled="loading || !!error || !filteredRows.length"
            @click="cetak"
          >
            <Printer :size="16" /> Cetak
          </button>
        </div>
      </header>
      <div class="jadwal-history">
        <div class="jadwal-filter">
          <FormInput v-model="filter.q" label="Pencarian" type="search" placeholder="Paket, dokter, ruang, perawat..." />
          <FormInput
            v-model="filter.status"
            label="Status"
            jenis="select"
            :options="[{ label: 'Semua', value: '' }, ...statusOperasi]"
          />
          <FormInput v-model="filter.mulai" label="Dari Tanggal" type="date" />
          <FormInput v-model="filter.selesai" label="Sampai Tanggal" type="date" />
        </div>
        <p v-if="peringatan" class="jadwal-info" role="status">{{ peringatan }}</p>
        <p v-if="error" class="patient-error" role="alert">{{ error }}</p>
        <DataTable
          v-else
          class="visit-report-table"
          :rows="filteredRows"
          data-key="kunci"
          :loading="loading"
          paginator
          :rows-per-page="10"
          :rows-per-page-options="[10, 25, 50]"
          empty-message="Tidak ada jadwal operasi yang sesuai."
        >
          <Column header="Layanan Pendukung" style="min-width: 145px">
            <template #body="{ data: r }">
              <button
                type="button"
                class="clinical-button secondary"
                :disabled="loading || saving"
                @click="emit('pilih-jadwal', r)"
              >
                Pilih Jadwal
              </button>
            </template>
          </Column>
          <Column header="Tanggal / Waktu" style="min-width: 185px">
            <template #body="{ data: r }">
              <strong>{{ r.tanggal }}</strong>
              <small v-if="r.jam_mulai === '00:00:00' && r.jam_selesai === '00:00:00'">
                Waktu belum ditentukan
              </small>
              <small v-else>{{ r.jam_mulai }} – {{ r.jam_selesai }}</small>
            </template>
          </Column>
          <Column header="Paket Operasi" style="min-width: 220px">
            <template #body="{ data: r }">
              <strong>{{ r.nama_paket }}</strong>
              <small>{{ r.kode_paket }}</small>
            </template>
          </Column>
          <Column field="nama_dokter" header="Operator" style="min-width: 180px" />
          <Column field="nama_ruang" header="Ruang OK" style="min-width: 130px" />
          <Column field="status" header="Status" style="min-width: 130px" />
          <Column header="Anestesi / Perawat" style="min-width: 180px">
            <template #body="{ data: r }">
              <strong>{{ r.dokteranastesi || '-' }}</strong>
              <small>{{ r.perawat || '-' }}</small>
            </template>
          </Column>
          <Column field="sumber" header="Sumber" />
          <Column header="Aksi" style="min-width: 170px">
            <template #body="{ data: r }">
              <div v-if="r.bisa_ubah" class="clinical-table-actions">
                <button :disabled="terkunci" @click="edit(r)">
                  <Pencil :size="14" /> Edit
                </button>
                <button class="danger" :disabled="terkunci" @click="hapusTarget = r">
                  <Trash2 :size="14" /> Hapus
                </button>
              </div>
              <span v-else>Hanya baca</span>
            </template>
          </Column>
        </DataTable>
      </div>
    </article>
    <Dialog
      :visible="!!hapusTarget"
      header="Hapus jadwal operasi?"
      modal
      :closable="!saving"
      :close-on-escape="!saving"
      :style="{ width: '480px', maxWidth: '95vw' }"
      @update:visible="value => { if (!value && !saving) hapusTarget = null }"
    >
      <p>{{ hapusTarget?.nama_paket }} — {{ hapusTarget?.tanggal }}</p>
      <p>Jadwal ini akan dihapus dari {{ hapusTarget?.sumber }}.</p>
      <p v-if="hapusTarget?.sumber === 'Khanza'">Penghapusan langsung berlaku di Khanza dan tidak dapat dipulihkan melalui halaman ini.</p>
      <template #footer>
        <button class="clinical-button secondary" :disabled="saving" @click="hapusTarget = null">Batal</button>
        <button class="clinical-button danger" :disabled="terkunci" @click="hapus">
          {{ saving ? 'Memproses...' : 'Ya, Hapus' }}
        </button>
      </template>
    </Dialog>
  </section>
</template>

<style src="@/Components/Ui/report.css" scoped></style>
<style src="@/Pages/Pasien/JadwalOperasi/jadwal-operasi.css" scoped></style>
