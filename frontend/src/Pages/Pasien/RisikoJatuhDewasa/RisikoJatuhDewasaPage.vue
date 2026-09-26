<script setup lang="ts">
import { ChevronDown, ChevronUp, Eye, LoaderCircle, Pencil, RefreshCw, Save, Trash2, X } from '@lucide/vue'
import Column from 'primevue/column'
import Dialog from 'primevue/dialog'
import DataTable from '../../../Components/Ui/DataTable.vue'
import FormInput from '../../../Components/Ui/FormInput.vue'
import InputPencarian from '../../../Components/Ui/InputPencarian.vue'
import TableSearch from '../../../Components/Ui/TableSearch.vue'
import type { PropsRisikoJatuh } from '../../../types/risikoJatuhDewasa'
import { useRisikoJatuhDewasa } from './useRisikoJatuhDewasa'

const props = defineProps<PropsRisikoJatuh>()
const {
  loading, saving, error, errorSimpan, keyword, skala, rows, records, formVisible, editing, detail,
  hapusTarget, petugas, form, nilai, total, risiko, terkunci,
  waktuSekarang, reset, muat, cariPetugas, edit, mutasi,
} = useRisikoJatuhDewasa(props)
</script>

<template>
  <section class="risiko-jatuh-page clinical-page">
    <article class="clinical-form-card">
      <header class="clinical-section-header">
        <div>
          <span>Rekam Medis · Rawat Inap</span>
          <h3>{{ editing ? 'Edit Penilaian' : 'Lanjutan Risiko Jatuh Dewasa' }}</h3>
          <p>Penilaian Morse mengikuti SIMRS. Penyimpanan langsung ke SIMRS.</p>
        </div>
        <div class="clinical-section-tools">
          <button v-if="editing" type="button" class="clinical-button secondary" :disabled="saving" @click="reset">
            <X :size="15" /> Batal Edit
          </button>
          <button
            type="button"
            class="clinical-button toggle icon-only"
            :title="formVisible ? 'Sembunyikan Form Input' : 'Tampilkan Form Input'"
            :aria-label="formVisible ? 'Sembunyikan Form Input' : 'Tampilkan Form Input'"
            :aria-expanded="formVisible"
            aria-controls="risiko-jatuh-form"
            @click="formVisible = !formVisible"
          >
            <ChevronUp v-if="formVisible" :size="15" />
            <ChevronDown v-else :size="15" />
          </button>
        </div>
      </header>
      <form v-show="formVisible" id="risiko-jatuh-form" class="clinical-form" @submit.prevent="mutasi()">
        <fieldset class="form-compact risiko-fields" :disabled="saving">
          <p v-if="editing" role="status">Mengedit penilaian {{ editing.data.tanggal }}.</p>
          <div class="risiko-grid">
            <FormInput
              v-model="form.tanggal"
              label="Tanggal / Jam (WITA)"
              type="datetime-local"
              step="1"
              required
              :disabled="terkunci"
            />
            <InputPencarian
              v-model="petugas"
              label="Petugas Penilai"
              :search="cariPetugas"
              required
              :disabled="terkunci"
            />
          </div>
          <div class="risiko-grid">
            <div v-for="(item, i) in skala" :key="item.label" class="risiko-pertanyaan">
              <FormInput
                v-model="form['penilaian_jatuhmorse_skala' + (i + 1)]"
                :label="`${i + 1}. ${item.label}`"
                jenis="select"
                :options="item.pilihan.map((label, n) => ({ label: `${label} (${item.nilai[n]})`, value: label }))"
                required
                :disabled="terkunci"
              />
              <small>Skor: {{ nilai[i] ?? 'Belum dipilih' }}</small>
            </div>
          </div>
          <section class="risiko-ringkasan" aria-label="Ringkasan penilaian Morse" aria-live="polite" aria-atomic="true">
            <div class="risiko-total">
              <div class="risiko-angka">{{ total ?? '—' }}</div>
              <div class="risiko-keterangan">
                <span class="risiko-label">Total Skor</span>
                <strong>Morse</strong>
                <small>{{ total === null ? 'Lengkapi enam komponen penilaian' : 'Dihitung dari enam komponen penilaian' }}</small>
              </div>
            </div>
            <div class="risiko-status">
              <span class="risiko-label">Tingkat Risiko</span>
              <strong
                class="risiko-badge"
                :class="{
                  rendah: total !== null && total < 25,
                  sedang: total !== null && total >= 25 && total < 45,
                  tinggi: total !== null && total >= 45,
                }"
              >
                <span class="risiko-titik" aria-hidden="true"></span>
                {{ total === null ? 'Belum lengkap' : `Risiko ${risiko.toLowerCase()}` }}
              </strong>
              <small>{{ total === null ? 'Kategori muncul setelah semua pilihan diisi.' : 'Berdasarkan total skor saat ini.' }}</small>
            </div>
            <div class="risiko-panduan">
              <span class="risiko-label">Batas Kategori</span>
              <dl>
                <div :class="{ aktif: total !== null && total < 25 }">
                  <dt>Rendah</dt>
                  <dd>&lt; 25</dd>
                </div>
                <div :class="{ aktif: total !== null && total >= 25 && total < 45 }">
                  <dt>Sedang</dt>
                  <dd>25–44</dd>
                </div>
                <div :class="{ aktif: total !== null && total >= 45 }">
                  <dt>Tinggi</dt>
                  <dd>≥ 45</dd>
                </div>
              </dl>
            </div>
          </section>
          <div class="clinical-soap-grid">
            <FormInput v-model="form.hasil_skrining" label="Hasil Skrining" jenis="textarea" :maxlength="200" required :disabled="terkunci" />
            <FormInput v-model="form.saran" label="Saran" jenis="textarea" :maxlength="200" required :disabled="terkunci" />
          </div>
          <p v-if="errorSimpan" class="patient-error" role="alert">{{ errorSimpan }}</p>
          <footer class="clinical-form-actions">
            <button type="button" class="clinical-button secondary" :disabled="terkunci" @click="waktuSekarang">
              <RefreshCw :size="16" /> Waktu Sekarang
            </button>
            <button type="button" class="clinical-button secondary" :disabled="saving" @click="reset">
              <X :size="16" /> Batal / Reset
            </button>
            <button type="submit" class="clinical-button primary" :disabled="terkunci">
              <LoaderCircle v-if="saving" class="spin" :size="15" />
              <Save v-else :size="15" />
              {{ saving ? 'Menyimpan...' : editing ? 'Simpan Perubahan' : 'Simpan Penilaian' }}
            </button>
          </footer>
        </fieldset>
      </form>
    </article>

    <article class="clinical-history-card">
      <header class="clinical-section-header">
        <div>
          <span>Kunjungan {{ patient.no_rawat }}</span>
          <h3>Riwayat Penilaian</h3>
          <p>{{ rows.length }} dari {{ records.length }} penilaian ditampilkan.</p>
        </div>
        <div class="clinical-section-tools">
          <TableSearch
            v-model="keyword"
            placeholder="Cari tanggal, petugas, atau penilaian..."
            :total="records.length"
            :filtered="rows.length"
            aria-label="Cari riwayat penilaian"
          />
        <button class="clinical-button secondary" :disabled="loading || saving" @click="muat">
          <RefreshCw :size="16" /> Muat Ulang
        </button>
        </div>
      </header>
      <div v-if="loading" class="clinical-state" role="status">
        <LoaderCircle class="spin" :size="25" />
        <strong>Menarik riwayat penilaian...</strong>
      </div>
      <div v-else-if="error" class="clinical-state error" role="alert">
        <strong>{{ error }}</strong>
        <button type="button" @click="muat">Coba Lagi</button>
      </div>
      <DataTable
        v-else
        :rows="rows"
        data-key="kunci"
        :loading="loading"
        paginator
        :rows-per-page="10"
        :rows-per-page-options="[10, 25, 50]"
        empty-message="Belum ada penilaian pada kunjungan ini."
      >
        <Column header="Tanggal / Jam" style="min-width:150px">
          <template #body="{ data: r }">
            <div class="clinical-table-main">
              <strong>{{ r.data.tanggal.slice(0, 10) }}</strong>
              <span>{{ r.data.tanggal.slice(11) }} WITA</span>
            </div>
          </template>
        </Column>
        <Column header="Petugas" style="min-width:180px">
          <template #body="{ data: r }">
            <div class="clinical-table-main">
              <strong>{{ r.nama_petugas || r.data.nip }}</strong>
              <span>{{ r.data.nip }}</span>
            </div>
          </template>
        </Column>
        <Column field="data.penilaian_jatuhmorse_totalnilai" header="Skor" header-class="right-align-header" body-class="right-align-body" />
        <Column field="risiko" header="Risiko" />
        <Column header="Hasil Skrining" style="min-width:200px">
          <template #body="{ data: r }"><p class="clinical-table-note">{{ r.data.hasil_skrining }}</p></template>
        </Column>
        <Column header="Saran" style="min-width:200px">
          <template #body="{ data: r }"><p class="clinical-table-note">{{ r.data.saran }}</p></template>
        </Column>
        <Column header="Aksi" style="min-width:150px">
          <template #body="{ data: r }">
            <div class="clinical-table-actions">
              <button type="button" @click="detail = r"><Eye :size="14" /> Detail</button>
              <button v-if="r.bisa_ubah" type="button" :disabled="terkunci" @click="edit(r)"><Pencil :size="14" /> Edit</button>
              <button v-if="r.bisa_ubah" type="button" class="danger" :disabled="terkunci" @click="hapusTarget = r; errorSimpan = ''"><Trash2 :size="14" /> Hapus</button>
            </div>
          </template>
        </Column>
      </DataTable>
    </article>

    <Dialog :visible="!!detail" modal header="Detail Penilaian Morse" :style="{ width: '720px', maxWidth: '95vw' }" @update:visible="detail = null">
      <div v-if="detail" class="risiko-detail">
        <p>{{ detail.data.tanggal }} · {{ detail.nama_petugas }}</p>
        <dl>
          <template v-for="(item, i) in skala" :key="item.label">
            <dt>{{ item.label }}</dt>
            <dd>{{ detail.data['penilaian_jatuhmorse_skala' + (i + 1)] }} — Skor {{ detail.data['penilaian_jatuhmorse_nilai' + (i + 1)] }}</dd>
          </template>
          <dt>Total / Risiko</dt><dd>{{ detail.data.penilaian_jatuhmorse_totalnilai }} / {{ detail.risiko }}</dd>
          <dt>Hasil Skrining</dt><dd>{{ detail.data.hasil_skrining }}</dd>
          <dt>Saran</dt><dd>{{ detail.data.saran }}</dd>
        </dl>
      </div>
    </Dialog>
    <Dialog :visible="!!hapusTarget" modal header="Hapus Penilaian?" :closable="!saving" :style="{ width: '480px', maxWidth: '95vw' }" @update:visible="!saving && (hapusTarget = null)">
      <p>Penilaian tanggal {{ hapusTarget?.data.tanggal }} akan dihapus langsung dari SIMRS. Tindakan ini tidak dapat dibatalkan.</p>
      <p v-if="errorSimpan" role="alert">{{ errorSimpan }}</p>
      <template #footer>
        <button class="clinical-button secondary" :disabled="saving" @click="hapusTarget = null">Batal</button>
        <button class="clinical-button danger" :disabled="saving" @click="mutasi(true)">{{ saving ? 'Menghapus...' : 'Ya, Hapus' }}</button>
      </template>
    </Dialog>
  </section>
</template>

<style src="@/Pages/Pasien/RisikoJatuhDewasa/risiko-jatuh-dewasa.css" scoped></style>
