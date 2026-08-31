<script setup lang="ts">
// @ts-nocheck -- migrasi TypeScript bertahap; kontrak data modul lama belum sepenuhnya bertipe.
import { ChevronDown, ChevronUp, ClipboardPlus, FlaskConical, Plus, Save, Trash2 } from "@lucide/vue"
import FormInput from "../../../Components/Ui/FormInput.vue"
import InputPencarian from "../../../Components/Ui/InputPencarian.vue"
import Select from "../../../Components/Ui/Select.vue"
import { useInputResep } from "./useInputResep"

const props = defineProps({
  token: { type: String, required: true },
  patient: { type: Object, required: true },
  moduleName: { type: String, default: 'Rawat Inap' },
  copiedResep: { type: Object, default: null },
})

const {
  loading,
  saving,
  locked,
  recipes,
  methods,
  doctor,
  depoGudang,
  keywordObat,
  hasilObat,
  editing,
  formVisible,
  tab,
  form,
  items,
  racikan,
  racikDraft,
  racikMedicine,
  rupiah,
  tanggalPendek,
  jamPendek,
  labelStatusResep,
  noRacikDraft,
  total,
  ppn,
  tagihan,
  cariDokter,
  cariDepo,
  cariObat,
  reset,
  cariBanyakObat,
  obatTerpilih,
  togglePilihanObat,
  removeMedicine,
  hitungJumlahRacik,
  hitungKandunganDariP,
  hitungSemuaDetailRacik,
  addRacikMedicine,
  addRacik,
  editRecipe,
  save,
  remove,
} = useInputResep(props)
</script>

<template>
  <section class="recipe-page">
    <div v-if="locked" class="recipe-lock">Billing sudah diproses. Resep tidak dapat ditambah, diubah, atau dihapus.</div>

    <article class="recipe-card recipe-form-card">
      <header class="recipe-panel-header">
        <div class="recipe-panel-title">
          <ClipboardPlus :size="16" />
          <span>Peresepan Dokter</span>
        </div>
        <div class="recipe-section-tools">
          <button type="button" class="recipe-template-button" :disabled="locked">
            <ClipboardPlus :size="15" />
            Template Resep
          </button>
          <button class="recipe-toggle" type="button" :title="formVisible ? 'Sembunyikan Form Input' : 'Tampilkan Form Input'" @click="formVisible=!formVisible">
            <ChevronUp v-if="formVisible" :size="18"/>
            <ChevronDown v-else :size="18"/>
          </button>
        </div>
      </header>

      <form v-show="formVisible" class="recipe-form clinical-form" @submit.prevent="save">
        <div class="recipe-grid top-grid">
          <InputPencarian v-model="doctor" label="NIP Dokter" code-field="kd_dokter" name-field="nm_dokter" description-field="spesialis" placeholder="NIP" :search="cariDokter" required :disabled="locked"/>
          <FormInput v-model="doctor.nm_dokter" label="Nama Dokter" placeholder="Nama dokter" disabled/>
          <FormInput v-model="doctor.spesialis" label="Jabatan" placeholder="-" disabled/>
          <FormInput v-model="form.tgl_peresepan" label="Tanggal Peresepan" type="date" required :disabled="locked"/>
          <FormInput v-model="form.jam" label="Jam Peresepan" type="time" step="1" required :disabled="locked"/>
          <FormInput v-model="form.no_resep" label="No. Resep" placeholder="Auto" disabled/>
          <FormInput v-model="form.judul" label="Judul Resep" placeholder="Judul" :disabled="locked"/>
          <InputPencarian
            v-model="depoGudang"
            label="Depo / Gudang"
            code-field="kd_bangsal"
            name-field="nm_bangsal"
            placeholder="Cari depo/gudang..."
            :search="cariDepo"
            required
            :disabled="locked"
          />
          <FormInput v-model="form.status" label="Tarif" jenis="select" :options="[{label:'Rawat Jalan',value:'ralan'},{label:'Rawat Inap',value:'ranap'}]" :disabled="locked"/>
        </div>

        <div class="recipe-total-strip">
          <div><small>Total</small><strong>{{ rupiah(total) }}</strong></div>
          <div><small>PPN 11%</small><strong class="warn">{{ rupiah(ppn) }}</strong></div>
          <div><small>Tagihan</small><strong>{{ rupiah(tagihan) }}</strong></div>
        </div>

        <div class="recipe-tabs">
          <button type="button" :class="{active:tab==='resep'}" @click="tab='resep'"><ClipboardPlus :size="15" />Resep<b>{{ items.length }}</b></button>
          <button type="button" :class="{active:tab==='racikan'}" @click="tab='racikan'"><FlaskConical :size="15" />Racikan<b>{{ racikan.length }}</b></button>
        </div>

        <section v-show="tab === 'resep'" class="recipe-tab-panel">
          <div class="recipe-search recipe-bulk-search">
            <FormInput v-model="keywordObat" label="Cari Obat" placeholder="Cari kode/nama obat..." :disabled="locked" @keydown.enter.prevent="cariBanyakObat"/>
          </div>
          <div v-if="hasilObat.length" class="recipe-pick-list">
            <button v-for="obat in hasilObat" :key="obat.kode_brng" type="button" :class="{selected:obatTerpilih(obat)}" @click="togglePilihanObat(obat)">
              <span class="recipe-check" aria-hidden="true">{{ obatTerpilih(obat) ? '✓' : '' }}</span>
              <span><b>{{ obat.nama_brng }}</b><small>{{ obat.kode_brng }} - {{ obat.kode_sat }} <template v-if="obat.jenis">- {{ obat.jenis }}</template></small></span>
              <strong class="stock" :class="{danger:Number(obat.stok||0)<=0}">Stok {{ obat.stok ?? 0 }}</strong>
              <strong>{{ rupiah(obat.harga) }}</strong>
            </button>
          </div>
          <div class="recipe-items">
            <template v-if="items.length">
              <div class="recipe-row recipe-row-head">
                <span>K</span><span>Jumlah</span><span>Kode Barang</span><span>Nama Barang</span><span>Satuan</span><span>Harga(Rp)</span><span>Jenis Obat</span><span>Aturan Pakai</span><span>H.Beli</span><span>Stok</span><span>Fornas</span><span>#</span>
              </div>
              <div v-for="(item, i) in items" :key="item.kode_brng" class="recipe-row">
                <input v-model="item.pilih" type="checkbox" :disabled="locked"/>
                <input v-model.number="item.jml" class="form-input-element recipe-inline-input" type="number" min="1" :disabled="locked"/>
                <code>{{ item.kode_brng }}</code>
                <strong>{{ item.nama_brng }}</strong>
                <span>{{ item.kode_sat || '-' }}</span>
                <span class="money">{{ rupiah(item.harga) }}</span>
                <span>{{ item.jenis || '-' }}</span>
                <input v-model="item.aturan_pakai" class="form-input-element recipe-inline-input" placeholder="mis: 3×1" :disabled="locked"/>
                <span class="money">{{ rupiah(item.h_beli) }}</span>
                <span :class="{danger:Number(item.stok||0)<=0}">{{ item.stok ?? '-' }}</span>
                <span><i v-if="item.kategori_fornas" class="recipe-badge">{{ item.kategori_fornas }}</i><template v-else>-</template></span>
                <button type="button" class="icon-button danger" :disabled="locked" @click="removeMedicine(i)"><Trash2 :size="15"/></button>
              </div>
            </template>
            <div v-else class="recipe-empty-box">
              <ClipboardPlus :size="26" />
              <span>Belum ada obat. Cari obat menggunakan kolom pencarian di atas.</span>
            </div>
          </div>
        </section>

        <section v-show="tab === 'racikan'" class="recipe-tab-panel">
          <div class="racik-khanza-table racik-header-table">
            <div class="racik-header-row racik-header-head">
              <span>No</span>
              <span>Nama Racikan</span>
              <span>Metode Racik</span>
              <span>Jml.Racik</span>
              <span>Aturan Pakai</span>
              <span>Keterangan</span>
            </div>
            <div class="racik-header-row">
              <span>{{ noRacikDraft }}</span>
              <input v-model="racikDraft.nama_racik" class="form-input-element racik-cell-input" :disabled="locked" placeholder="Nama racikan"/>
              <Select
                v-model="racikDraft.kd_racik"
                class="racik-cell-select"
                :options="methods"
                option-label="nm_racik"
                option-value="kd_racik"
                placeholder="Pilih metode"
                append-to="body"
                overlay-class="recipe-select-overlay"
                filter
                :disabled="locked"
              />
              <input v-model.number="racikDraft.jml_dr" class="form-input-element racik-cell-input" type="number" min="1" :disabled="locked" @input="hitungSemuaDetailRacik"/>
              <input v-model="racikDraft.aturan_pakai" class="form-input-element racik-cell-input" :disabled="locked" placeholder="mis: 3×1"/>
              <input v-model="racikDraft.keterangan" class="form-input-element racik-cell-input" :disabled="locked" placeholder="Keterangan"/>
            </div>
          </div>

          <div class="recipe-search compact">
            <InputPencarian v-model="racikMedicine" label="Obat Racikan" code-field="kode_brng" name-field="nama_brng" right-field="harga" :right-formatter="rupiah" :search="cariObat" placeholder="Cari obat untuk racikan..." :disabled="locked"/>
            <button type="button" class="recipe-action-button" :disabled="locked||!racikMedicine.kode_brng" @click="addRacikMedicine"><Plus :size="15"/> Tambah</button>
          </div>

          <div v-if="racikDraft.detail.length" class="racik-khanza-table racik-detail-table">
            <div class="racik-detail-row racik-header-head">
              <span>No</span>
              <span>Kode Barang</span>
              <span>Nama Barang</span>
              <span>Satuan</span>
              <span>Harga(Rp)</span>
              <span>Jenis Obat</span>
              <span>Stok</span>
              <span>Kps</span>
              <span>P1</span>
              <span>/</span>
              <span>P2</span>
              <span>Kandungan</span>
              <span>Jml</span>
              <span>I.F.</span>
              <span>Komposisi</span>
              <span>#</span>
            </div>
            <div v-for="(d,i) in racikDraft.detail" :key="`${d.kode_brng}-${i}`" class="racik-detail-row">
              <span>{{ noRacikDraft }}</span>
              <code>{{ d.kode_brng }}</code>
              <strong>{{ d.nama_brng }}</strong>
              <span>{{ d.kode_sat || '-' }}</span>
              <span class="money">{{ rupiah(d.harga) }}</span>
              <span>{{ d.jenis || '-' }}</span>
              <span :class="{danger:Number(d.stok||0)<=0}">{{ d.stok ?? 0 }}</span>
              <span>{{ d.kapasitas || 0 }}</span>
              <input v-model.number="d.p1" class="form-input-element recipe-inline-input mini" type="number" step="0.001" :disabled="locked" @input="hitungKandunganDariP(d)"/>
              <span>/</span>
              <input v-model.number="d.p2" class="form-input-element recipe-inline-input mini" type="number" step="0.001" :disabled="locked" @input="hitungKandunganDariP(d)"/>
              <input v-model="d.kandungan" class="form-input-element recipe-inline-input" :disabled="locked" placeholder="Kandungan" @input="hitungJumlahRacik(d)"/>
              <input v-model.number="d.jml" class="form-input-element recipe-inline-input mini" type="number" step="0.1" :disabled="locked"/>
              <span>{{ d.nama_industri || 'lain-lain' }}</span>
              <span>{{ d.letak_barang || '-' }}</span>
              <button type="button" class="icon-button danger" :disabled="locked" @click="racikDraft.detail.splice(i,1)"><Trash2 :size="15"/></button>
            </div>
          </div>
          <div v-else class="recipe-empty-box"><FlaskConical :size="26" /><span>Cari obat racikan, lalu tambahkan ke tabel detail.</span></div>

          <div class="recipe-racik-actions">
            <small>{{ methods.length ? `${methods.length} metode racik tersedia` : 'Metode racik belum termuat dari SIMRS Khanza' }}</small>
            <button type="button" class="recipe-action-button" :disabled="locked" @click="addRacik"><Plus :size="15"/> Tambah Racikan</button>
          </div>
          <div v-if="racikan.length" class="racik-saved-list">
            <div v-for="(r,i) in racikan" :key="i" class="racik-saved">
              <b>{{ r.no_racik || i + 1 }}. {{ r.nama_racik }}</b>
              <small>{{ r.nm_racik || r.kd_racik }} - {{ r.detail?.length || 0 }} obat - {{ r.aturan_pakai || '-' }}</small>
              <button type="button" class="icon-button danger" :disabled="locked" @click="racikan.splice(i,1)"><Trash2 :size="15"/></button>
            </div>
          </div>
        </section>

        <footer class="recipe-actions">
          <button type="submit" class="recipe-action-button" :disabled="locked||saving"><Save :size="16"/> {{ saving ? 'Menyimpan...' : (editing ? 'Simpan Perubahan' : 'Simpan Resep') }}</button>
          <small>Aturan pakai wajib diisi untuk setiap obat.</small>
        </footer>
      </form>
    </article>

    <article class="recipe-card recipe-history-card">
      <header class="recipe-section-header">
        <div><span class="eyebrow">RIWAYAT RESEP</span><h3>Resep Pasien</h3><p>{{ recipes.length }} resep tersimpan</p></div>
        <button type="button" class="recipe-action-button secondary" :disabled="locked" @click="reset"><Plus :size="15"/> Resep Baru</button>
      </header>
      <div v-if="loading" class="empty-state">Memuat resep...</div>
      <div v-else-if="!recipes.length" class="empty-state">Belum ada resep untuk pasien ini.</div>
      <div v-else class="recipe-history">
        <div class="history-head">
          <span>No. Resep</span>
          <span>Tanggal / Dokter</span>
          <span>Isi Resep</span>
          <span>Status</span>
          <span>Total</span>
          <span>Aksi</span>
        </div>
        <div v-for="row in recipes" :key="row.no_resep" class="history-row">
          <div class="history-prescription">
            <b>{{ row.no_resep }}</b>
            <small>{{ row.judul || 'Tanpa judul resep' }}</small>
          </div>
          <div>
            <b>{{ tanggalPendek(row.tgl_peresepan) }} {{ jamPendek(row.jam) }}</b>
            <small>{{ row.nm_dokter || row.kd_dokter || '-' }}</small>
          </div>
          <div>
            <b>{{ row.jumlah_obat || 0 }} obat</b>
            <small>{{ row.jumlah_racik || 0 }} racikan</small>
          </div>
          <span class="recipe-status-pill">{{ labelStatusResep(row.status) }}</span>
          <strong class="history-total">{{ rupiah(row.total) }}</strong>
          <div class="history-actions">
            <button type="button" class="recipe-action-button secondary" :disabled="locked" @click="editRecipe(row)">Edit</button>
            <button type="button" class="icon-button danger" :disabled="locked" @click="remove(row)"><Trash2 :size="15"/></button>
          </div>
        </div>
      </div>
    </article>
  </section>
</template>

<style src="./input-resep.css" scoped></style>
