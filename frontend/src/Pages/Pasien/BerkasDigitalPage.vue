<script setup>
import { ChevronDown, ChevronUp, ExternalLink, Eye, FileUp, LoaderCircle, Save, Trash2, X } from "@lucide/vue"
import Column from "primevue/column"
import { ref } from "vue"
import DataTable from "../../Components/Ui/DataTable.vue"
import FormInput from "../../Components/Ui/FormInput.vue"
import TableSearch from "../../Components/Ui/TableSearch.vue"
import { useBerkasDigital } from "./BerkasDigital/useBerkasDigital.js"

const props = defineProps({
  token: { type: String, required: true },
  patient: { type: Object, required: true },
})

const {
  loading,
  saving,
  deleting,
  formVisible,
  fileInput,
  deleteTarget,
  previewTarget,
  kataKunciBerkas,
  error,
  berkas,
  billingLocked,
  form,
  opsiMaster,
  namaFile,
  rows,
  rowsTampil,
  previewUrl,
  previewNama,
  previewIsImage,
  previewIsPdf,
  resetForm,
  onFileChange,
  namaBerkas,
  openPreview,
  loadData,
  upload,
  confirmDelete,
} = useBerkasDigital(props)
</script>

<template>
  <section class="clinical-page berkas-digital-page">
    <article class="clinical-form-card berkas-digital-card">
      <header class="clinical-section-header">
        <div>
          <span>Berkas Digital Perawatan</span>
          <h3>Upload Berkas Digital</h3>
          <p>Data tersimpan ke tabel berkas_digital_perawatan berdasarkan nomor rawat pasien.</p>
        </div>
        <div class="clinical-section-tools">
          <button
            type="button"
            class="clinical-button toggle icon-only"
            :title="formVisible ? 'Sembunyikan Form Input' : 'Tampilkan Form Input'"
            @click="formVisible = !formVisible"
          >
            <ChevronUp v-if="formVisible" :size="15" />
            <ChevronDown v-else :size="15" />
          </button>
        </div>
      </header>

      <div v-if="billingLocked" class="berkas-digital-lock">
        Billing sudah terverifikasi atau kunjungan dibatalkan. Berkas digital hanya dapat dilihat.
      </div>

      <form v-show="formVisible" class="clinical-form berkas-digital-form" @submit.prevent="upload">
        <fieldset class="form-compact" :disabled="saving || billingLocked">
          <div class="berkas-digital-grid">
            <FormInput
              v-model="form.kode"
              label="Jenis Berkas"
              jenis="select"
              :options="opsiMaster"
              placeholder="Pilih jenis berkas"
              append-to="body"
              overlay-class="berkas-digital-select-overlay"
              filter
              required
            />

            <label class="form-input-field berkas-file-field">
              <span class="form-input-label">File Berkas <i aria-hidden="true">*</i></span>
              <div class="form-input-element berkas-file-picker">
                <input
                  ref="fileInput"
                  class="berkas-file-native"
                  type="file"
                  accept=".pdf,.jpg,.jpeg,.png"
                  required
                  @change="onFileChange"
                />
                <button type="button" class="berkas-file-button" @click="fileInput?.click()">
                  Choose File
                </button>
                <span class="berkas-file-name">{{ namaFile || 'No file chosen' }}</span>
              </div>
              <small class="form-input-message">
                {{ namaFile || 'Format PDF, JPG, JPEG, atau PNG. Maksimal 25 MB.' }}
              </small>
            </label>
          </div>

          <footer class="clinical-form-actions">
            <button type="button" class="clinical-button secondary" @click="resetForm">
              <X :size="15" /> Batal / Reset
            </button>
            <button type="submit" class="clinical-button primary">
              <LoaderCircle v-if="saving" class="spin" :size="15" />
              <Save v-else :size="15" />
              {{ saving ? 'Mengupload...' : 'Upload Berkas' }}
            </button>
          </footer>
        </fieldset>
      </form>
    </article>

    <article class="clinical-history-card berkas-digital-history">
      <header class="clinical-section-header">
        <div>
          <span>Riwayat Berkas</span>
          <h3>Berkas Digital Pasien</h3>
          <p>
            <template v-if="kataKunciBerkas">
              {{ rowsTampil.length }} dari {{ berkas.length }} berkas ditampilkan
            </template>
            <template v-else>
              {{ berkas.length }} berkas ditemukan
            </template>
          </p>
        </div>
        <TableSearch
          v-if="berkas.length > 0"
          v-model="kataKunciBerkas"
          placeholder="Cari berkas digital..."
          :total="berkas.length"
          :filtered="rowsTampil.length"
        />
      </header>

      <div v-if="loading" class="clinical-state">
        <LoaderCircle class="spin" :size="25" />
        <strong>Menarik data berkas digital...</strong>
      </div>
      <div v-else-if="error" class="clinical-state error">
        <strong>{{ error }}</strong>
        <button type="button" @click="loadData">Coba Lagi</button>
      </div>
      <div v-else-if="berkas.length === 0" class="clinical-state">
        <FileUp :size="28" />
        <strong>Belum ada berkas digital pada kunjungan ini.</strong>
      </div>
      <DataTable
        v-else
        :rows="rowsTampil"
        data-key="_key"
        empty-message="Berkas digital tidak ditemukan."
      >
        <Column header="JENIS BERKAS" style="min-width:220px">
          <template #body="{ data }">
            <div class="clinical-table-main">
              <strong>{{ namaBerkas(data) }}</strong>
              <span>{{ data.kode }}</span>
            </div>
          </template>
        </Column>
        <Column header="LOKASI FILE" style="min-width:360px">
          <template #body="{ data }">
            <div class="clinical-table-main">
              <strong>{{ data.lokasi_file }}</strong>
              <button v-if="data.url" type="button" class="berkas-inline-link" @click="openPreview(data)">
                Buka file
              </button>
              <span v-else>URL belum tersedia</span>
            </div>
          </template>
        </Column>
        <Column header="AKSI" frozen align-frozen="right" style="min-width:150px">
          <template #body="{ data }">
            <div class="clinical-table-actions">
              <button v-if="data.url" type="button" class="berkas-action-link" @click="openPreview(data)">
                <Eye :size="14" /> Buka
              </button>
              <button v-if="!billingLocked" type="button" class="danger" @click="deleteTarget = data">
                <Trash2 :size="14" /> Hapus
              </button>
              <span v-else>Terkunci</span>
            </div>
          </template>
        </Column>
      </DataTable>
    </article>

    <div v-if="deleteTarget" class="clinical-confirm-backdrop" @click.self="deleteTarget = null">
      <section class="clinical-confirm-dialog">
        <h3>Hapus Berkas Digital?</h3>
        <p>{{ namaBerkas(deleteTarget) }} akan dihapus dari tabel berkas digital perawatan. File fisik tidak ikut dihapus.</p>
        <div>
          <button type="button" class="clinical-button secondary" @click="deleteTarget = null">Batal</button>
          <button type="button" class="clinical-button danger" :disabled="deleting" @click="confirmDelete">
            <LoaderCircle v-if="deleting" class="spin" :size="15" />
            <Trash2 v-else :size="15" />
            {{ deleting ? 'Menghapus...' : 'Hapus' }}
          </button>
        </div>
      </section>
    </div>

    <div v-if="previewTarget" class="berkas-preview-backdrop" @click.self="previewTarget = null">
      <section class="berkas-preview-dialog" role="dialog" aria-modal="true" aria-labelledby="berkas-preview-title">
        <header class="berkas-preview-header">
          <div>
            <span>Preview Berkas</span>
            <h3 id="berkas-preview-title">{{ previewNama }}</h3>
            <p>{{ previewTarget.lokasi_file }}</p>
          </div>
          <button type="button" class="berkas-preview-close" title="Tutup preview" @click="previewTarget = null">
            <X :size="18" />
          </button>
        </header>

        <div class="berkas-preview-body">
          <img v-if="previewIsImage" :src="previewUrl" :alt="previewNama" />
          <iframe v-else-if="previewIsPdf" :src="previewUrl" title="Preview berkas digital"></iframe>
          <div v-else class="berkas-preview-empty">
            <FileUp :size="34" />
            <strong>Preview file ini belum didukung.</strong>
            <p>Silakan buka di tab baru untuk melihat file.</p>
          </div>
        </div>

        <footer class="berkas-preview-footer">
          <a :href="previewUrl" target="_blank" rel="noreferrer" class="berkas-preview-open">
            <ExternalLink :size="15" /> Buka di Tab Baru
          </a>
        </footer>
      </section>
    </div>
  </section>
</template>

<style src="./BerkasDigital/berkas-digital.css" scoped></style>
