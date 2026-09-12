<script setup lang="ts">
import { ArrowRightToLine, ChevronDown, ChevronUp, LoaderCircle, Pencil, RefreshCw, Save, Trash2, X } from '@lucide/vue'
import Column from 'primevue/column'
import Dialog from 'primevue/dialog'
import DataTable from '../../../Components/Ui/DataTable.vue'
import FormInput from '../../../Components/Ui/FormInput.vue'
import InputPencarian from '../../../Components/Ui/InputPencarian.vue'
import TableSearch from '../../../Components/Ui/TableSearch.vue'
import type { PropsRujukanInternal } from '../../../types/rujukanInternal'
import { useRujukanInternal } from './useRujukanInternal'

const props = defineProps<PropsRujukanInternal>()
const {
  ranap, judul, loading, saving, error, errorSimpan, pesanKunci, formVisible,
  keyword, records, poli, dokter, form, terkunci, filteredRows,
  waktuSekarang, resetForm, muat, cariReferensi, simpan,
  editing, edit, konfirmasi, errorAksi, mintaKonfirmasi, jalankanAksi,
} = useRujukanInternal(props)
</script>

<template>
  <section class="rujukan-internal-page clinical-page">
    <article class="clinical-form-card">
      <header class="clinical-section-header">
        <div>
          <span>Pelayanan {{ moduleName }}</span>
          <h3>{{ judul }}</h3>
          <p>Pilih unit/poliklinik dan dokter yang dituju untuk kunjungan ini.</p>
        </div>
        <button
          type="button"
          class="clinical-button toggle"
          :aria-expanded="formVisible"
          @click="formVisible = !formVisible"
        >
          <ChevronUp v-if="formVisible" :size="16" />
          <ChevronDown v-else :size="16" />
          {{ formVisible ? 'Tutup Form' : 'Buka Form' }}
        </button>
      </header>

      <p class="rujukan-info">
        <ArrowRightToLine :size="18" aria-hidden="true" />
        Rujukan baru langsung disimpan di Khanza. Rujukan lokal lama dapat dikirim
        satu per satu melalui tombol Kirim ke Khanza, tanpa menimpa data yang sudah ada.
      </p>
      <p v-if="pesanKunci" class="billing-lock" role="status">{{ pesanKunci }}</p>

      <form v-show="formVisible" class="clinical-form form-compact rujukan-form" @submit.prevent="simpan">
        <fieldset :disabled="terkunci">
          <p v-if="editing" class="rujukan-edit-status" role="status">
            Mengedit rujukan {{ editing.sumber }}: {{ editing.nama_poli }} — {{ editing.nama_dokter }}.
            {{ editing.sumber === 'SIRAPI' ? 'Perubahan tetap lokal sampai dikirim.' : 'Perubahan disimpan langsung di Khanza.' }}
          </p>
          <div class="rujukan-fields" :key="`${patient.no_rawat}-${moduleName}`">
            <InputPencarian
              v-model="poli"
              label="Unit / Poliklinik Tujuan"
              placeholder="Cari kode atau nama poliklinik..."
              :search="kata => cariReferensi('poli', kata)"
              :disabled="terkunci"
              required
            />
            <InputPencarian
              v-model="dokter"
              label="Dokter Dituju"
              placeholder="Cari kode atau nama dokter..."
              :search="kata => cariReferensi('dokter', kata)"
              :disabled="terkunci"
              required
            />
            <template v-if="ranap">
              <FormInput v-model="form.tanggal" label="Tanggal Rujukan" type="date" :disabled="terkunci" required />
              <FormInput v-model="form.jam" label="Jam Rujukan (WITA)" type="time" step="1" :disabled="terkunci" required />
            </template>
          </div>
          <p class="rujukan-help">
            {{ ranap
              ? 'Satu rujukan per poli tujuan dalam kunjungan yang sama, mengikuti Khanza.'
              : 'Satu rujukan per dokter tujuan dalam kunjungan yang sama, mengikuti Khanza.' }}
          </p>
          <p v-if="errorSimpan" class="patient-error" role="alert">{{ errorSimpan }}</p>
          <footer class="clinical-form-actions">
            <button v-if="ranap" type="button" class="clinical-button secondary" :disabled="terkunci" @click="waktuSekarang">
              <RefreshCw :size="15" /> Waktu Sekarang
            </button>
            <button type="button" class="clinical-button secondary" :disabled="terkunci" @click="resetForm">
              <X :size="15" /> {{ editing ? 'Batal Edit' : 'Reset' }}
            </button>
            <button type="submit" class="clinical-button primary" :disabled="terkunci">
              <LoaderCircle v-if="saving" class="spin" :size="15" />
              <Save v-else :size="15" />
              {{ saving ? 'Menyimpan...' : editing ? 'Simpan Perubahan' : 'Simpan ke Khanza' }}
            </button>
          </footer>
        </fieldset>
      </form>
    </article>

    <article class="clinical-history-card">
      <header class="clinical-section-header">
        <div>
          <span>Kunjungan {{ patient.no_rawat }}</span>
          <h3>Riwayat Rujukan Internal</h3>
          <p v-if="!loading && !error">{{ filteredRows.length }} dari {{ records.length }} rujukan ditampilkan.</p>
        </div>
        <div class="clinical-section-tools">
          <TableSearch v-model="keyword" placeholder="Cari poli, dokter, atau sumber..." :total="records.length" :filtered="filteredRows.length" />
          <button type="button" class="clinical-button secondary" :disabled="loading || saving" @click="muat">
            <RefreshCw :size="15" /> Muat Ulang
          </button>
        </div>
      </header>
      <div v-if="loading" class="clinical-state" role="status">
        <LoaderCircle class="spin" :size="24" />
        <strong>Memuat riwayat rujukan...</strong>
      </div>
      <div v-else-if="error" class="clinical-state error" role="alert">
        <strong>{{ error }}</strong>
        <button type="button" class="clinical-button secondary" @click="muat">Coba Lagi</button>
      </div>
      <DataTable
        v-else
        class="rujukan-table"
        :rows="filteredRows"
        data-key="_key"
        empty-message="Tidak ada rujukan internal yang sesuai."
      >
        <Column header="No." style="width:65px;text-align:center">
          <template #body="{ index }">{{ index + 1 }}</template>
        </Column>
        <Column header="Unit / Poliklinik Tujuan" style="min-width:200px">
          <template #body="{ data }">
            <div class="clinical-table-main"><strong>{{ data.nama_poli || '-' }}</strong><span>{{ data.kd_poli }}</span></div>
          </template>
        </Column>
        <Column header="Dokter Dituju" style="min-width:220px">
          <template #body="{ data }">
            <div class="clinical-table-main"><strong>{{ data.nama_dokter || '-' }}</strong><span>{{ data.kd_dokter }}</span></div>
          </template>
        </Column>
        <Column v-if="ranap" header="Tanggal / Jam" style="min-width:165px">
          <template #body="{ data }">
            <div class="clinical-table-main"><strong>{{ data.tanggal || '-' }}</strong><span>{{ data.jam || '-' }}</span></div>
          </template>
        </Column>
        <Column header="Sumber" style="width:200px">
          <template #body="{ data }">
            <div class="clinical-table-main rujukan-source">
              <strong>{{ data.sumber }}</strong>
              <span v-if="data.konflik_khanza" class="rujukan-conflict">
                Kunci rujukan sudah ada di Khanza, tetapi isinya berbeda. Periksa kedua baris sebelum melanjutkan.
              </span>
              <span v-else>{{ data.sumber === 'SIRAPI' ? 'Lokal, belum dikirim ke Khanza' : 'Tersimpan di Khanza' }}</span>
            </div>
          </template>
        </Column>
        <Column header="Aksi" style="min-width:210px;width:210px">
          <template #body="{ data }">
            <div class="rujukan-actions">
              <button type="button" class="clinical-button secondary" :disabled="terkunci" @click="edit(data)">
                <Pencil :size="14" /> Edit
              </button>
              <button type="button" class="clinical-button secondary" :disabled="terkunci" @click="mintaKonfirmasi('hapus', data)">
                <Trash2 :size="14" /> Hapus
              </button>
              <button
                v-if="data.sumber === 'SIRAPI'"
                type="button"
                class="clinical-button primary rujukan-send"
                :disabled="terkunci || data.konflik_khanza"
                :title="data.konflik_khanza ? 'Data berbeda dengan Khanza. Periksa dan sesuaikan terlebih dahulu.' : 'Kirim rujukan lokal ini ke Khanza'"
                @click="mintaKonfirmasi('kirim', data)"
              >
                <ArrowRightToLine :size="14" /> Kirim ke Khanza
              </button>
            </div>
          </template>
        </Column>
      </DataTable>
    </article>

    <Dialog
      class="rujukan-confirm"
      :visible="!!konfirmasi"
      :header="konfirmasi?.aksi === 'hapus' ? 'Hapus rujukan?' : 'Kirim rujukan ke Khanza?'"
      modal
      append-to="self"
      :closable="!saving"
      :close-on-escape="!saving"
      :style="{ width: '480px', maxWidth: 'calc(100vw - 32px)' }"
      @update:visible="value => { if (!value && !saving) konfirmasi = null }"
    >
      <template v-if="konfirmasi">
        <div class="rujukan-confirm-detail">
          <strong>{{ konfirmasi.row.nama_poli }}</strong>
          <span>{{ konfirmasi.row.nama_dokter }}</span>
          <span>No. rawat: {{ konfirmasi.row.no_rawat }}</span>
          <span v-if="ranap">{{ konfirmasi.row.tanggal }} · {{ konfirmasi.row.jam }}</span>
          <span>Sumber: {{ konfirmasi.row.sumber }}</span>
        </div>
        <p v-if="konfirmasi.aksi === 'hapus'" class="rujukan-confirm-help">
          Hanya rujukan ini yang akan dihapus dari {{ konfirmasi.row.sumber }}.
          {{ konfirmasi.row.sumber === 'SIRAPI' ? 'Data di Khanza tidak dihapus.' : 'Penghapusan langsung berlaku di Khanza.' }}
          Tindakan ini tidak dapat dibatalkan.
        </p>
        <p v-else class="rujukan-confirm-help">
          Rujukan dikirim ke Khanza dan salinan lokal diarsipkan. Jika kuncinya sudah ada
          dengan isi berbeda, pengiriman ditolak. Data Khanza tidak ditimpa.
        </p>
        <p v-if="errorAksi" class="patient-error" role="alert">{{ errorAksi }}</p>
      </template>
      <template #footer>
        <button type="button" class="clinical-button secondary" :disabled="saving" autofocus @click="konfirmasi = null">
          Batal
        </button>
        <button type="button" class="clinical-button primary" :disabled="saving || terkunci" @click="jalankanAksi">
          <LoaderCircle v-if="saving" class="spin" :size="15" />
          {{ saving ? 'Memproses...' : konfirmasi?.aksi === 'hapus' ? 'Ya, Hapus Rujukan' : 'Ya, Kirim ke Khanza' }}
        </button>
      </template>
    </Dialog>
  </section>
</template>

<style src="./rujukan-internal.css" scoped></style>
