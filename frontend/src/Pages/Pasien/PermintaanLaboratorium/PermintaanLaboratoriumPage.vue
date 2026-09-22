<script setup lang="ts">
import { ChevronDown, ChevronUp, LoaderCircle, Pencil, Printer, Save, Trash2, X } from '@lucide/vue'
import Column from 'primevue/column'
import Dialog from 'primevue/dialog'
import CariDokter from '../../../Components/Ui/CariDokter.vue'
import CariTindakanLaboratorium from '../../../Components/Ui/CariTindakanLaboratorium.vue'
import DataTable from '../../../Components/Ui/DataTable.vue'
import FormInput from '../../../Components/Ui/FormInput.vue'
import TableSearch from '../../../Components/Ui/TableSearch.vue'
import type { PropsLaboratorium } from '../../../types/permintaanLaboratorium'
import { usePermintaanLaboratorium } from './usePermintaanLaboratorium'

const props = defineProps<PropsLaboratorium>()
const {
  kategori, kategoriLaboratorium, namaKategori, gantiKategori,
  loading, saving, deleting, busy, detailLoading, error, formVisible, confirmVisible,
  requests, doctor, treatments, billingLocked, scope, deleteTarget,
  editingNumber, kataKunciPermintaan, form, tableRowsTampil,
  rupiah, formatDate, waktuStatus, kelasStatusPemeriksaan, kelasStatusBayar,
  resetForm, editRequest, loadData, saveRequest, confirmSave, confirmDelete, cetak,
  waktuSekarang,
} = usePermintaanLaboratorium(props)
</script>

<template>
  <section class="clinical-page radiology-request-page laboratory-request-page">
    <nav class="clinical-section-tools laboratory-category-tabs" aria-label="Kategori laboratorium">
      <button
        v-for="item in kategoriLaboratorium"
        :key="item.value"
        type="button"
        class="clinical-button"
        :class="kategori === item.value ? 'primary' : 'secondary'"
        :aria-pressed="kategori === item.value"
        :disabled="busy || detailLoading"
        @click="gantiKategori(item.value)"
      >
        {{ item.label }}
      </button>
    </nav>

    <article class="clinical-form-card radiology-form-card laboratory-form-card">
      <header class="clinical-section-header">
        <div>
          <span>Penunjang Medis · {{ namaKategori }}</span>
          <h3>{{ editingNumber ? 'Edit Permintaan Laboratorium' : 'Input Permintaan Laboratorium' }}</h3>
          <p>Satu nomor permintaan untuk satu kategori. Simpan kategori lain secara terpisah.</p>
        </div>
        <div class="clinical-section-tools">
          <button
            v-if="editingNumber"
            type="button"
            class="clinical-button secondary"
            :disabled="busy"
            @click="resetForm"
          >
            <X :size="15" /> Batal Edit
          </button>
          <button
            type="button"
            class="clinical-button toggle icon-only"
            :title="formVisible ? 'Sembunyikan Form Input' : 'Tampilkan Form Input'"
            :aria-expanded="formVisible"
            @click="formVisible = !formVisible"
          >
            <ChevronUp v-if="formVisible" :size="15" />
            <ChevronDown v-else :size="15" />
          </button>
        </div>
      </header>

      <div v-if="billingLocked && !loading && !error" class="handling-lock">
        Billing terkunci atau kunjungan dibatalkan. Data permintaan tidak dapat diubah.
      </div>

      <form v-show="formVisible" class="clinical-form radiology-form" @submit.prevent="saveRequest">
        <fieldset class="form-compact" :disabled="busy || billingLocked">
          <div class="radiology-form-main">
            <FormInput v-model="form.tanggal" label="Tanggal Permintaan" type="date" required />
            <FormInput v-model="form.jam" label="Jam Permintaan (WITA)" type="time" step="1" required />
            <CariDokter
              v-model="doctor"
              class="radiology-referrer"
              :token="token"
              sumber="laboratorium"
              :disabled="busy || billingLocked"
              required
            />
          </div>
          <div class="radiology-notes-grid">
            <FormInput
              v-model="form.informasi_tambahan"
              label="Informasi Tambahan"
              jenis="textarea"
              :rows="2"
              maxlength="60"
              required
            />
            <FormInput
              v-model="form.diagnosis_klinis"
              label="Diagnosis Klinis"
              jenis="textarea"
              :rows="2"
              maxlength="80"
              required
            />
          </div>
          <div v-if="kategori === 'PA'" class="radiology-notes-grid">
            <FormInput v-model="form.spesimen.pengambilan_bahan" label="Tanggal Pengambilan Bahan" type="date" required />
            <FormInput v-model="form.spesimen.diperoleh_dengan" label="Bahan Diperoleh Dengan" maxlength="40" />
            <FormInput v-model="form.spesimen.lokasi_jaringan" label="Lokasi Pengambilan Jaringan" maxlength="40" />
            <FormInput v-model="form.spesimen.diawetkan_dengan" label="Diawetkan Dengan" maxlength="40" />
            <FormInput v-model="form.spesimen.pernah_dilakukan_di" label="PA Sebelumnya Dilakukan Di" maxlength="100" />
            <FormInput
              v-model="form.spesimen.tanggal_pa_sebelumnya"
              label="Tanggal PA Sebelumnya"
              type="date"
              :required="Boolean(form.spesimen.pernah_dilakukan_di.trim())"
            />
            <FormInput v-model="form.spesimen.nomor_pa_sebelumnya" label="Nomor PA Sebelumnya" maxlength="20" />
            <FormInput v-model="form.spesimen.diagnosa_pa_sebelumnya" label="Diagnosis PA Sebelumnya" maxlength="100" />
          </div>
          <CariTindakanLaboratorium
            :key="`${patient.no_rawat}:${kategori}`"
            v-model="treatments"
            :token="token"
            :no-rawat="patient.no_rawat || ''"
            :kategori="kategori"
            :disabled="busy || billingLocked"
            @loading="detailLoading = $event"
            required
          />
          <div class="radiology-filter-note">
            <span>{{ scope.status === 'ranap' ? 'Rawat Inap' : 'Rawat Jalan' }}</span>
            <span>Penjamin pasien: {{ scope.kodeCaraBayar || '-' }}</span>
            <span>Kelas pasien: {{ scope.kelas || '-' }}</span>
            <span>Pilihan mengikuti pengaturan tarif Khanza.</span>
          </div>
          <footer class="clinical-form-actions">
            <button type="button" class="clinical-button secondary" @click="Object.assign(form, waktuSekarang())">
              Waktu Sekarang
            </button>
            <button type="button" class="clinical-button secondary" @click="resetForm">
              <X :size="15" /> Batal / Reset
            </button>
            <button type="submit" class="clinical-button primary" :disabled="detailLoading">
              <LoaderCircle v-if="saving" class="spin" :size="15" />
              <Save v-else :size="15" />
              {{ saving ? 'Menyimpan...' : editingNumber ? 'Simpan Perubahan' : 'Simpan Permintaan' }}
            </button>
          </footer>
        </fieldset>
      </form>
    </article>

    <article class="clinical-history-card radiology-history-card">
      <header class="clinical-section-header">
        <div>
          <span>Riwayat Pasien · {{ namaKategori }}</span>
          <h3>Permintaan Laboratorium</h3>
          <p>{{ tableRowsTampil.length }} dari {{ requests.length }} permintaan ditampilkan</p>
        </div>
        <TableSearch
          v-if="requests.length"
          v-model="kataKunciPermintaan"
          placeholder="Cari permintaan, dokter, atau pemeriksaan..."
          :total="requests.length"
          :filtered="tableRowsTampil.length"
        />
      </header>
      <div v-if="loading" class="clinical-state">
        <LoaderCircle class="spin" :size="25" />
        <strong>Menarik permintaan laboratorium...</strong>
      </div>
      <div v-else-if="error" class="clinical-state error">
        <strong>{{ error }}</strong>
        <button type="button" class="clinical-button secondary" @click="loadData">Coba Lagi</button>
      </div>
      <div v-else-if="!requests.length" class="clinical-state">
        <strong>Belum ada permintaan {{ namaKategori }}.</strong>
      </div>
      <DataTable v-else :rows="tableRowsTampil" data-key="_key" empty-message="Permintaan laboratorium tidak ditemukan.">
        <Column header="NO. PERMINTAAN" style="min-width: 175px">
          <template #body="{ data }">
            <div class="clinical-table-main">
              <strong>{{ data.nomor }}</strong>
              <span>{{ formatDate(data.tanggal) }} · {{ data.jam }}</span>
            </div>
          </template>
        </Column>
        <Column header="DOKTER PERUJUK" style="min-width: 230px">
          <template #body="{ data }">
            <div class="clinical-table-main">
              <strong>{{ data.nama_dokter }}</strong>
              <span>{{ data.kode_dokter }}</span>
            </div>
          </template>
        </Column>
        <Column header="INFORMASI / DIAGNOSIS" style="min-width: 300px">
          <template #body="{ data }">
            <div class="radiology-clinical">
              <p><b>Informasi:</b> {{ data.informasi_tambahan }}</p>
              <p><b>Diagnosis:</b> {{ data.diagnosis_klinis }}</p>
              <template v-if="data.kategori === 'PA'">
                <p><b>Bahan:</b> {{ data.spesimen.diperoleh_dengan || '-' }}</p>
                <p><b>Lokasi:</b> {{ data.spesimen.lokasi_jaringan || '-' }}</p>
                <p><b>Pengambilan:</b> {{ formatDate(data.spesimen.pengambilan_bahan) }}</p>
              </template>
            </div>
          </template>
        </Column>
        <Column header="PEMERIKSAAN LABORATORIUM" style="min-width: 360px">
          <template #body="{ data }">
            <div class="radiology-exam-list">
              <article v-for="item in data.pemeriksaan" :key="item.kode">
                <span>
                  <strong>{{ item.nama }}</strong>
                  <small>{{ item.kode }} · {{ item.kelas }}</small>
                  <small v-if="item.detail?.length" class="laboratory-detail-summary">
                    {{ item.detail.map((detail) => detail.nama).join(', ') }}
                  </small>
                </span>
                <b>{{ rupiah(item.total) }}</b>
              </article>
            </div>
          </template>
        </Column>
        <Column header="TOTAL" style="min-width: 130px">
          <template #body="{ data }"><strong>{{ rupiah(data.total) }}</strong></template>
        </Column>
        <Column header="STATUS PEMERIKSAAN" style="min-width: 180px">
          <template #body="{ data }">
            <div class="radiology-status-cell">
              <span class="radiology-status-badge" :class="kelasStatusPemeriksaan(data.status_pemeriksaan)">
                {{ data.status_pemeriksaan }}
              </span>
              <small>{{ waktuStatus(data) }}</small>
            </div>
          </template>
        </Column>
        <Column header="STATUS BAYAR" style="min-width: 145px">
          <template #body="{ data }">
            <span class="radiology-status-badge" :class="kelasStatusBayar(data.status_bayar)">
              {{ data.status_bayar }}
            </span>
          </template>
        </Column>
        <Column header="AKSI" frozen align-frozen="right" style="min-width: 200px">
          <template #body="{ data }">
            <div class="clinical-table-actions">
              <button type="button" @click="cetak(data)"><Printer :size="14" /> Cetak</button>
              <button
                v-if="!billingLocked && data.dapat_diubah"
                type="button"
                :disabled="busy"
                @click="editRequest(data)"
              >
                <Pencil :size="14" /> Edit
              </button>
              <button
                v-if="!billingLocked && data.dapat_dihapus"
                type="button"
                class="danger"
                :disabled="busy"
                @click="deleteTarget = data"
              >
                <Trash2 :size="14" /> Hapus
              </button>
            </div>
          </template>
        </Column>
      </DataTable>
    </article>

    <Dialog
      v-model:visible="confirmVisible"
      modal
      header="Simpan Permintaan Laboratorium?"
      :closable="!saving"
      :close-on-escape="!saving"
      :style="{ width: '30rem', maxWidth: 'calc(100vw - 2rem)' }"
    >
      <p>Pastikan dokter, diagnosis, dan {{ treatments.length }} pemeriksaan {{ namaKategori }} sudah benar.</p>
      <template #footer>
        <button type="button" class="clinical-button secondary" :disabled="saving" @click="confirmVisible = false">
          Tidak, Periksa Lagi
        </button>
        <button type="button" class="clinical-button primary" :disabled="saving" @click="confirmSave">
          <LoaderCircle v-if="saving" class="spin" :size="15" />
          <Save v-else :size="15" /> Ya, Simpan
        </button>
      </template>
    </Dialog>
    <Dialog
      :visible="Boolean(deleteTarget)"
      modal
      header="Hapus Permintaan Laboratorium?"
      :closable="!deleting"
      :close-on-escape="!deleting"
      :style="{ width: '30rem', maxWidth: 'calc(100vw - 2rem)' }"
      @update:visible="(visible) => { if (!visible && !deleting) deleteTarget = null }"
    >
      <p>Permintaan {{ deleteTarget?.nomor }} beserta pemeriksaannya akan dihapus.</p>
      <template #footer>
        <button type="button" class="clinical-button secondary" :disabled="deleting" @click="deleteTarget = null">Batal</button>
        <button type="button" class="clinical-button danger" :disabled="deleting" @click="confirmDelete">
          <LoaderCircle v-if="deleting" class="spin" :size="15" />
          <Trash2 v-else :size="15" /> Hapus
        </button>
      </template>
    </Dialog>
  </section>
</template>

<style src="./permintaan-laboratorium.css" scoped></style>
