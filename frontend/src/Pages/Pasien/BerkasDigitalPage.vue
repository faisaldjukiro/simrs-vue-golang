<script setup>
import { ChevronDown, ChevronUp, ExternalLink, Eye, FileUp, LoaderCircle, Save, Trash2, X } from '@lucide/vue'
import Column from 'primevue/column'
import { computed, reactive, ref, watch } from 'vue'
import DataTable from '../../Components/Ui/DataTable.vue'
import FormInput from '../../Components/Ui/FormInput.vue'
import TableSearch from '../../Components/Ui/TableSearch.vue'
import { berkasDigitalData, hapusBerkasDigital, uploadBerkasDigital } from '../../lib/faisal/api'
import { useNotifikasi } from '../../lib/shared/useNotifikasi'

const props = defineProps({
  token: { type: String, required: true },
  patient: { type: Object, required: true },
})

const notifikasi = useNotifikasi()
const loading = ref(false)
const saving = ref(false)
const deleting = ref(false)
const formVisible = ref(true)
const fileInput = ref(null)
const deleteTarget = ref(null)
const previewTarget = ref(null)
const kataKunciBerkas = ref('')
const error = ref('')
const master = ref([])
const berkas = ref([])
const billingLocked = ref(false)
const form = reactive({ kode: '', file: null })

const opsiMaster = computed(() => master.value.map((item) => ({ label: `${item.kode} - ${item.nama}`, value: item.kode })))
const namaFile = computed(() => form.file?.name || '')
const rows = computed(() => berkas.value.map((item) => ({
  ...item,
  _key: `${item.no_rawat}|${item.kode}|${item.lokasi_file}`,
})))
const rowsTampil = computed(() => {
  const keyword = kataKunciBerkas.value.trim().toLowerCase()
  if (!keyword) return rows.value
  return rows.value.filter((item) => [
    item.kode,
    item.nama,
    item.lokasi_file,
  ].some((value) => String(value || '').toLowerCase().includes(keyword)))
})
const previewUrl = computed(() => previewTarget.value?.url || '')
const previewNama = computed(() => previewTarget.value ? namaBerkas(previewTarget.value) : '')
const previewEkstensi = computed(() => String(previewTarget.value?.lokasi_file || previewTarget.value?.url || '').toLowerCase())
const previewIsImage = computed(() => /\.(png|jpe?g|gif|webp)(\?|$)/i.test(previewEkstensi.value))
const previewIsPdf = computed(() => /\.pdf(\?|$)/i.test(previewEkstensi.value))

function resetForm() {
  form.kode = ''
  form.file = null
  if (fileInput.value) fileInput.value.value = ''
}

function onFileChange(event) {
  form.file = event.target.files?.[0] || null
}

function namaBerkas(item) {
  return item.nama || item.kode || 'Berkas Digital'
}

function openPreview(item) {
  if (!item?.url) {
    notifikasi.peringatan('URL file belum tersedia.')
    return
  }
  previewTarget.value = item
}

async function loadData() {
  if (!props.patient.no_rawat) return
  loading.value = true
  error.value = ''
  try {
    const data = await berkasDigitalData(props.token, props.patient.no_rawat)
    master.value = data?.master || []
    berkas.value = data?.berkas || []
    billingLocked.value = Boolean(data?.billing_terkunci)
  } catch (err) {
    error.value = err.message
    notifikasi.gagal(err.message)
  } finally {
    loading.value = false
  }
}

async function upload() {
  if (!form.kode || !form.file) {
    notifikasi.peringatan('Jenis berkas dan file wajib dipilih.')
    return
  }
  saving.value = true
  try {
    const payload = new FormData()
    payload.append('no_rawat', props.patient.no_rawat)
    payload.append('kode', form.kode)
    payload.append('file', form.file)
    const response = await uploadBerkasDigital(props.token, payload)
    notifikasi.sukses(response?.pesan || 'Berkas digital berhasil diupload.')
    resetForm()
    await loadData()
  } catch (err) {
    notifikasi.gagal(err.message)
  } finally {
    saving.value = false
  }
}

async function confirmDelete() {
  if (!deleteTarget.value) return
  deleting.value = true
  try {
    const response = await hapusBerkasDigital(props.token, {
      no_rawat: deleteTarget.value.no_rawat,
      kode: deleteTarget.value.kode,
      lokasi_file: deleteTarget.value.lokasi_file,
    })
    notifikasi.sukses(response?.pesan || 'Berkas digital berhasil dihapus.')
    deleteTarget.value = null
    await loadData()
  } catch (err) {
    notifikasi.gagal(err.message)
  } finally {
    deleting.value = false
  }
}

watch(() => props.patient.no_rawat, () => {
  resetForm()
  kataKunciBerkas.value = ''
  loadData()
}, { immediate: true })
</script>

<template>
  <section class="cppt-page berkas-digital-page">
    <article class="cppt-form-card berkas-digital-card">
      <header class="cppt-section-header">
        <div>
          <span>Berkas Digital Perawatan</span>
          <h3>Upload Berkas Digital</h3>
          <p>Data tersimpan ke tabel berkas_digital_perawatan berdasarkan nomor rawat pasien.</p>
        </div>
        <div class="cppt-section-tools">
          <button
            type="button"
            class="cppt-button toggle icon-only"
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

      <form v-show="formVisible" class="cppt-form berkas-digital-form" @submit.prevent="upload">
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

          <footer class="cppt-form-actions">
            <button type="button" class="cppt-button secondary" @click="resetForm">
              <X :size="15" /> Batal / Reset
            </button>
            <button type="submit" class="cppt-button primary">
              <LoaderCircle v-if="saving" class="spin" :size="15" />
              <Save v-else :size="15" />
              {{ saving ? 'Mengupload...' : 'Upload Berkas' }}
            </button>
          </footer>
        </fieldset>
      </form>
    </article>

    <article class="cppt-history-card berkas-digital-history">
      <header class="cppt-section-header">
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

      <div v-if="loading" class="cppt-state">
        <LoaderCircle class="spin" :size="25" />
        <strong>Menarik data berkas digital...</strong>
      </div>
      <div v-else-if="error" class="cppt-state error">
        <strong>{{ error }}</strong>
        <button type="button" @click="loadData">Coba Lagi</button>
      </div>
      <div v-else-if="berkas.length === 0" class="cppt-state">
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
            <div class="cppt-table-main">
              <strong>{{ namaBerkas(data) }}</strong>
              <span>{{ data.kode }}</span>
            </div>
          </template>
        </Column>
        <Column header="LOKASI FILE" style="min-width:360px">
          <template #body="{ data }">
            <div class="cppt-table-main">
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
            <div class="cppt-table-actions">
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

    <div v-if="deleteTarget" class="cppt-confirm-backdrop" @click.self="deleteTarget = null">
      <section class="cppt-confirm-dialog">
        <h3>Hapus Berkas Digital?</h3>
        <p>{{ namaBerkas(deleteTarget) }} akan dihapus dari tabel berkas digital perawatan. File fisik tidak ikut dihapus.</p>
        <div>
          <button type="button" class="cppt-button secondary" @click="deleteTarget = null">Batal</button>
          <button type="button" class="cppt-button danger" :disabled="deleting" @click="confirmDelete">
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

<style scoped>
.berkas-digital-grid {
  display: grid;
  grid-template-columns: minmax(220px, 340px) minmax(260px, 1fr);
  gap: 12px;
}

.berkas-file-picker {
  min-height: 45px;
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 6px 10px;
}

.berkas-file-native {
  position: absolute;
  width: 1px;
  height: 1px;
  opacity: 0;
  pointer-events: none;
}

.berkas-file-button {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-height: 30px;
  border: 0;
  border-radius: 9px;
  padding: 0 13px;
  color: white;
  background: #0d9488;
  font-size: 12px;
  font-weight: 850;
  line-height: 1;
  white-space: nowrap;
  cursor: pointer;
}

.berkas-file-button:disabled {
  cursor: not-allowed;
  opacity: .65;
}

.berkas-file-name {
  min-width: 0;
  overflow: hidden;
  color: var(--text);
  font-size: 12px;
  font-weight: 700;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.berkas-digital-lock {
  margin: 14px 16px 0;
  border: 1px solid rgba(245, 158, 11, .28);
  border-radius: 12px;
  padding: 11px 13px;
  color: #92400e;
  background: rgba(245, 158, 11, .1);
  font-size: 12px;
  font-weight: 750;
}

.berkas-action-link {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  border: 1px solid var(--line);
  border-radius: 9px;
  padding: 7px 10px;
  color: var(--text);
  background: var(--surface-soft);
  font-size: 11px;
  font-weight: 800;
  text-decoration: none;
  cursor: pointer;
}

.berkas-action-link:hover {
  color: #0d9488;
  border-color: rgba(13, 148, 136, .4);
}

.berkas-inline-link {
  width: fit-content;
  border: 0;
  padding: 0;
  color: #0d9488;
  background: transparent;
  font: inherit;
  font-size: 12px;
  font-weight: 800;
  cursor: pointer;
}

.berkas-inline-link:hover {
  text-decoration: underline;
}

.berkas-preview-backdrop {
  position: fixed;
  inset: 0;
  z-index: 10020;
  display: grid;
  place-items: center;
  padding: 18px;
  background: rgba(2, 6, 23, .72);
  backdrop-filter: blur(8px);
}

.berkas-preview-dialog {
  width: min(980px, calc(100vw - 32px));
  max-height: calc(100vh - 36px);
  display: flex;
  flex-direction: column;
  overflow: hidden;
  border: 1px solid var(--line);
  border-radius: 18px;
  color: var(--text);
  background: var(--surface);
  box-shadow: 0 30px 90px rgba(2, 6, 23, .38);
}

.berkas-preview-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
  border-bottom: 1px solid var(--line);
  padding: 18px 20px;
}

.berkas-preview-header span {
  display: block;
  margin-bottom: 5px;
  color: #0d9488;
  font-size: 11px;
  font-weight: 850;
  letter-spacing: .08em;
  text-transform: uppercase;
}

.berkas-preview-header h3 {
  margin: 0;
  font-size: 20px;
  line-height: 1.2;
}

.berkas-preview-header p {
  margin: 6px 0 0;
  color: var(--muted);
  font-size: 12px;
  overflow-wrap: anywhere;
}

.berkas-preview-close {
  display: inline-grid;
  width: 38px;
  height: 38px;
  flex: 0 0 auto;
  place-items: center;
  border: 1px solid var(--line);
  border-radius: 12px;
  color: var(--text);
  background: var(--surface-soft);
  cursor: pointer;
}

.berkas-preview-body {
  min-height: 420px;
  display: grid;
  place-items: center;
  overflow: auto;
  padding: 14px;
  background: var(--surface-soft);
}

.berkas-preview-body iframe {
  width: 100%;
  height: min(70vh, 720px);
  border: 0;
  border-radius: 12px;
  background: white;
}

.berkas-preview-body img {
  max-width: 100%;
  max-height: min(70vh, 720px);
  border-radius: 12px;
  object-fit: contain;
  background: white;
}

.berkas-preview-empty {
  display: grid;
  gap: 8px;
  place-items: center;
  color: var(--muted);
  text-align: center;
}

.berkas-preview-empty strong {
  color: var(--text);
}

.berkas-preview-empty p {
  margin: 0;
}

.berkas-preview-footer {
  display: flex;
  justify-content: flex-end;
  border-top: 1px solid var(--line);
  padding: 12px 16px;
  background: var(--surface);
}

.berkas-preview-open {
  display: inline-flex;
  align-items: center;
  gap: 7px;
  border-radius: 10px;
  padding: 9px 12px;
  color: white;
  background: #0d9488;
  font-size: 12px;
  font-weight: 850;
  text-decoration: none;
}

.theme-dark .berkas-digital-lock,
.sirapi-dark .berkas-digital-lock {
  color: #fde68a;
  background: rgba(245, 158, 11, .12);
}

:global(.berkas-digital-select-overlay) {
  z-index: 9999 !important;
  border: 1px solid var(--line) !important;
  border-radius: 12px !important;
  color: var(--text) !important;
  background: var(--surface) !important;
  box-shadow: 0 24px 60px rgba(15, 23, 42, .28) !important;
  overflow: hidden;
}

:global(.berkas-digital-select-overlay .p-select-list) {
  background: var(--surface) !important;
}

:global(.berkas-digital-select-overlay .p-select-option) {
  color: var(--text) !important;
  background: transparent !important;
}

:global(.berkas-digital-select-overlay .p-select-option:not(.p-disabled):hover),
:global(.berkas-digital-select-overlay .p-select-option:not(.p-disabled).p-focus) {
  color: #0f766e !important;
  background: rgba(20, 184, 166, .12) !important;
}

:global(.berkas-digital-select-overlay .p-select-option.p-select-option-selected) {
  color: white !important;
  background: #0d9488 !important;
}

:global(.theme-dark .berkas-digital-select-overlay),
:global(.sirapi-dark .berkas-digital-select-overlay) {
  border-color: rgba(148, 163, 184, .22) !important;
  color: #f8fafc !important;
  background: #0f172a !important;
  box-shadow: 0 24px 70px rgba(0, 0, 0, .52) !important;
}

:global(.theme-dark .berkas-digital-select-overlay .p-select-list),
:global(.sirapi-dark .berkas-digital-select-overlay .p-select-list) {
  background: #0f172a !important;
}

:global(.theme-dark .berkas-digital-select-overlay .p-select-option),
:global(.sirapi-dark .berkas-digital-select-overlay .p-select-option) {
  color: #f8fafc !important;
}

:global(.theme-dark .berkas-digital-select-overlay .p-select-option:not(.p-disabled):hover),
:global(.theme-dark .berkas-digital-select-overlay .p-select-option:not(.p-disabled).p-focus),
:global(.sirapi-dark .berkas-digital-select-overlay .p-select-option:not(.p-disabled):hover),
:global(.sirapi-dark .berkas-digital-select-overlay .p-select-option:not(.p-disabled).p-focus) {
  color: #5eead4 !important;
  background: #1e293b !important;
}

:global(.theme-dark .berkas-digital-select-overlay .p-select-option.p-select-option-selected),
:global(.sirapi-dark .berkas-digital-select-overlay .p-select-option.p-select-option-selected) {
  color: #042f2e !important;
  background: #2dd4bf !important;
}

@media (max-width: 760px) {
  .berkas-digital-grid {
    grid-template-columns: 1fr;
  }

}
</style>
