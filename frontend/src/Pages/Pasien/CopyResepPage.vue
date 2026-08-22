<script setup>
import { ClipboardCopy, Copy, LoaderCircle, Pill, RefreshCcw, Save } from '@lucide/vue'
import { computed, onMounted, ref, watch } from 'vue'
import {
  resepDaftarCopy,
  resepDetail,
  resepInfoPasien,
} from '../../lib/faisal/api'
import { useNotifikasi } from '../../lib/shared/useNotifikasi'

const props = defineProps({
  token: { type: String, required: true },
  patient: { type: Object, required: true },
  moduleName: { type: String, default: 'Rawat Inap' },
})
const emit = defineEmits(['copy-resep'])

const notifikasi = useNotifikasi()
const loading = ref(false)
const loadingDetail = ref(false)
const locked = ref(false)
const recipes = ref([])
const selected = ref(null)
const detail = ref(null)

const noRkmMedis = computed(() => props.patient?.no_rekam_medis || props.patient?.no_rkm_medis || props.patient?.no_rm || '')
const jumlahIsi = computed(() => (detail.value?.obat?.length || 0) + (detail.value?.racikan?.length || 0))

function rupiah(value) {
  return new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR', maximumFractionDigits: 0 }).format(Number(value || 0))
}
function tanggalPendek(value) {
  if (!value) return '-'
  const clean = String(value).slice(0, 10)
  const [y, m, d] = clean.split('-')
  return y && m && d ? `${d}/${m}/${y}` : value
}
function jamPendek(value) {
  return value ? String(value).slice(0, 5) : '-'
}
function labelStatus(value) {
  const status = String(value || '').toLowerCase()
  if (status === 'ranap') return 'Rawat Inap'
  if (status === 'ralan') return 'Rawat Jalan'
  return value || '-'
}
function angka(value, fallback = 0) {
  const n = Number(value)
  return Number.isFinite(n) ? n : fallback
}
function normalisasiObat(item) {
  return {
    ...item,
    jml: angka(item.jml ?? item.jumlah ?? item.jumlah_obat, 1),
    harga: angka(item.harga, 0),
    aturan_pakai: item.aturan_pakai || '',
  }
}
function normalisasiRacikan(item) {
  return {
    ...item,
    no_racik: '',
    jml_dr: angka(item.jml_dr, 1),
    aturan_pakai: item.aturan_pakai || '',
    keterangan: item.keterangan || '',
    detail: (item.detail || []).map((obat) => ({
      ...obat,
      jml: angka(obat.jml ?? obat.jumlah, 1),
      p1: angka(obat.p1, 1),
      p2: angka(obat.p2, 1),
      kandungan: obat.kandungan || '',
    })),
  }
}

async function load() {
  if (!props.patient?.no_rawat || !noRkmMedis.value) return
  loading.value = true
  selected.value = null
  detail.value = null
  try {
    const [info, daftar] = await Promise.all([
      resepInfoPasien(props.token, props.patient.no_rawat),
      resepDaftarCopy(props.token, noRkmMedis.value),
    ])
    locked.value = Boolean(info?.billing_terkunci)
    recipes.value = Array.isArray(daftar) ? daftar : []
  } catch (error) {
    notifikasi.gagal(error.message)
  } finally {
    loading.value = false
  }
}

async function pilihResep(row) {
  selected.value = row
  detail.value = null
  loadingDetail.value = true
  try {
    detail.value = await resepDetail(props.token, row.no_resep)
  } catch (error) {
    notifikasi.gagal(error.message)
  } finally {
    loadingDetail.value = false
  }
}

function salinResep() {
  if (locked.value) {
    notifikasi.peringatan('Billing sudah diproses. Resep tidak dapat dicopy.')
    return
  }
  if (!detail.value?.header) {
    notifikasi.peringatan('Pilih resep yang ingin disalin dulu.')
    return
  }
  const obat = (detail.value.obat || []).map(normalisasiObat)
  const racikan = (detail.value.racikan || []).map(normalisasiRacikan)
  if (!obat.length && !racikan.length) {
    notifikasi.peringatan('Resep lama tidak memiliki obat atau racikan.')
    return
  }
  emit('copy-resep', {
    sumber: detail.value.header,
    obat,
    racikan,
  })
}

watch(() => props.patient?.no_rawat, load)
onMounted(load)
</script>

<template>
  <section class="copy-recipe-page">
    <div v-if="locked" class="copy-recipe-lock">
      Billing sudah diproses. Copy resep ke kunjungan ini tidak bisa dilakukan.
    </div>

    <article class="copy-recipe-card">
      <header class="copy-recipe-header">
        <div>
          <span class="eyebrow">COPY RESEP</span>
          <h3>Riwayat Resep Berdasarkan No. CM</h3>
          <p>Menampilkan seluruh resep pasien {{ noRkmMedis || '-' }} untuk disalin ke kunjungan saat ini.</p>
        </div>
        <button type="button" class="copy-recipe-button secondary" :disabled="loading" @click="load">
          <RefreshCcw :size="16" :class="{ spin: loading }" />
          Muat Ulang
        </button>
      </header>

      <div class="copy-recipe-layout">
        <section class="copy-recipe-list">
          <div v-if="loading" class="copy-recipe-empty">
            <LoaderCircle class="spin" :size="24" />
            <span>Memuat riwayat resep pasien...</span>
          </div>
          <div v-else-if="!recipes.length" class="copy-recipe-empty">
            <ClipboardCopy :size="28" />
            <span>Belum ada resep lama untuk No. CM ini.</span>
          </div>
          <button
            v-for="row in recipes"
            v-else
            :key="row.no_resep"
            type="button"
            class="copy-recipe-item"
            :class="{ active: selected?.no_resep === row.no_resep }"
            @click="pilihResep(row)"
          >
            <span>
              <strong>{{ row.no_resep }}</strong>
              <small>{{ tanggalPendek(row.tgl_peresepan) }} {{ jamPendek(row.jam) }} - {{ row.nm_dokter || '-' }}</small>
              <small>No. Rawat {{ row.no_rawat || '-' }}</small>
            </span>
            <span>
              <b>{{ rupiah(row.total) }}</b>
              <i>{{ row.jumlah_obat || 0 }} obat / {{ row.jumlah_racik || 0 }} racikan</i>
              <em>{{ labelStatus(row.status) }}</em>
            </span>
          </button>
        </section>

        <section class="copy-recipe-preview">
          <div v-if="loadingDetail" class="copy-recipe-empty">
            <LoaderCircle class="spin" :size="24" />
            <span>Membaca isi resep...</span>
          </div>

          <template v-else-if="detail?.header">
            <div class="copy-recipe-preview-head">
              <div>
                <span class="eyebrow">RESEP DIPILIH</span>
                <h3>{{ detail.header.no_resep }}</h3>
                <p>{{ tanggalPendek(detail.header.tgl_peresepan) }} {{ jamPendek(detail.header.jam) }} - {{ detail.header.nm_dokter || '-' }}</p>
              </div>
              <button type="button" class="copy-recipe-button" :disabled="locked || jumlahIsi === 0" @click="salinResep">
                <Copy :size="16" />
                Copy ke Input Resep
              </button>
            </div>

            <div class="copy-recipe-table">
              <div class="copy-recipe-table-head">
                <span>Jenis</span>
                <span>Kode</span>
                <span>Nama</span>
                <span>Jumlah</span>
                <span>Aturan Pakai</span>
              </div>
              <div v-for="item in detail.obat" :key="`obat-${item.kode_brng}`" class="copy-recipe-table-row">
                <span><Pill :size="14" /> Obat</span>
                <code>{{ item.kode_brng }}</code>
                <strong>{{ item.nama_brng }}</strong>
                <b>{{ item.jml }}</b>
                <small>{{ item.aturan_pakai || '-' }}</small>
              </div>
              <template v-for="racik in detail.racikan" :key="`racik-${racik.no_racik}`">
                <div class="copy-recipe-table-row racik">
                  <span>Racikan</span>
                  <code>{{ racik.kd_racik }}</code>
                  <strong>{{ racik.nama_racik }}</strong>
                  <b>{{ racik.jml_dr }}</b>
                  <small>{{ racik.aturan_pakai || '-' }}</small>
                </div>
                <div v-for="obat in racik.detail" :key="`${racik.no_racik}-${obat.kode_brng}`" class="copy-recipe-table-row child">
                  <span>Isi racik</span>
                  <code>{{ obat.kode_brng }}</code>
                  <strong>{{ obat.nama_brng }}</strong>
                  <b>{{ obat.jml }}</b>
                  <small>{{ obat.kandungan || '-' }}</small>
                </div>
              </template>
            </div>

            <footer class="copy-recipe-footer">
              <Save :size="15" />
              Setelah tombol copy diklik, data akan dibawa ke Input Resep. Resep lama tidak diubah.
            </footer>
          </template>

          <div v-else class="copy-recipe-empty">
            <ClipboardCopy :size="30" />
            <span>Pilih salah satu resep lama untuk melihat isinya.</span>
          </div>
        </section>
      </div>
    </article>
  </section>
</template>

<style scoped>
.copy-recipe-page{display:grid;gap:14px;color:var(--text)}
.copy-recipe-lock{padding:12px 16px;border:1px solid #e6a23c;color:#f59e0b;background:rgba(230,162,60,.12);border-radius:10px;font-weight:700}
.copy-recipe-card{border:1px solid var(--line);border-radius:14px;background:var(--surface);overflow:hidden}
.copy-recipe-header,.copy-recipe-preview-head{display:flex;align-items:center;justify-content:space-between;gap:16px;padding:16px 18px;border-bottom:1px solid var(--line)}
.copy-recipe-header{background:color-mix(in srgb,#0d9488 7%,var(--surface))}
.copy-recipe-header h3,.copy-recipe-preview-head h3{margin:4px 0 0;font-size:1.15rem}
.copy-recipe-header p,.copy-recipe-preview-head p{margin:4px 0 0;color:var(--muted);font-size:.86rem}
.copy-recipe-button{display:inline-flex;align-items:center;justify-content:center;gap:8px;min-height:38px;border:0;border-radius:10px;background:#0d9488;color:#fff;font-weight:800;cursor:pointer;padding:0 14px;white-space:nowrap}
.copy-recipe-button.secondary{border:1px solid var(--line);background:var(--surface-soft);color:var(--text)}
.copy-recipe-button:disabled{opacity:.55;cursor:not-allowed}
.copy-recipe-layout{display:grid;grid-template-columns:minmax(310px,.85fr) minmax(0,1.3fr);gap:14px;padding:16px;background:color-mix(in srgb,#0d9488 3%,var(--surface-soft))}
.copy-recipe-list,.copy-recipe-preview{min-height:420px;border:1px solid var(--line);border-radius:12px;background:var(--surface);overflow:hidden}
.copy-recipe-list{display:grid;align-content:start;max-height:680px;overflow:auto}
.copy-recipe-item{display:grid;grid-template-columns:minmax(0,1fr) auto;gap:12px;border:0;border-bottom:1px solid var(--line);background:transparent;color:var(--text);padding:13px 14px;text-align:left;cursor:pointer}
.copy-recipe-item:hover,.copy-recipe-item.active{background:rgba(13,148,136,.10)}
.copy-recipe-item span{display:grid;gap:4px;min-width:0}
.copy-recipe-item strong,.copy-recipe-item b{overflow:hidden;text-overflow:ellipsis;white-space:nowrap}
.copy-recipe-item strong{color:var(--text)}
.copy-recipe-item small,.copy-recipe-item i{color:var(--muted);font-size:.78rem;font-style:normal;overflow:hidden;text-overflow:ellipsis;white-space:nowrap}
.copy-recipe-item b{color:#0d9488;text-align:right}
.copy-recipe-item em{justify-self:end;width:max-content;border-radius:999px;background:rgba(20,184,166,.14);color:#0f766e;font-size:.72rem;font-style:normal;font-weight:800;padding:4px 8px}
.copy-recipe-empty{min-height:260px;display:grid;place-items:center;align-content:center;gap:10px;color:var(--muted);text-align:center;padding:24px}
.copy-recipe-table{margin:16px;border:1px solid var(--line);border-radius:12px;overflow:auto;background:var(--surface)}
.copy-recipe-table-head,.copy-recipe-table-row{display:grid;grid-template-columns:110px 110px minmax(240px,1fr) 90px minmax(160px,.8fr);gap:12px;align-items:center;min-width:780px}
.copy-recipe-table-head{position:sticky;top:0;z-index:1;padding:11px 14px;background:color-mix(in srgb,#0d9488 14%,var(--surface-soft));color:var(--muted);font-size:.7rem;font-weight:900;text-transform:uppercase;letter-spacing:.08em}
.copy-recipe-table-row{padding:12px 14px;border-top:1px solid var(--line)}
.copy-recipe-table-row:nth-child(odd){background:color-mix(in srgb,#0d9488 3%,var(--surface))}
.copy-recipe-table-row span{display:inline-flex;align-items:center;gap:7px;color:var(--muted);font-weight:800}
.copy-recipe-table-row code{color:#2563eb}
.copy-recipe-table-row strong{min-width:0;overflow:hidden;text-overflow:ellipsis;white-space:nowrap}
.copy-recipe-table-row small{color:var(--muted)}
.copy-recipe-table-row.racik{background:rgba(139,92,246,.10)}
.copy-recipe-table-row.child{padding-left:28px}
.copy-recipe-footer{display:flex;align-items:center;gap:8px;margin:0 16px 16px;padding:11px 13px;border:1px dashed var(--line);border-radius:10px;color:var(--muted);background:var(--surface-soft);font-size:.86rem}
:global(.theme-dark) .copy-recipe-list,:global(.theme-dark) .copy-recipe-preview,:global(.theme-dark) .copy-recipe-table,:global(.sirapi-dark) .copy-recipe-list,:global(.sirapi-dark) .copy-recipe-preview,:global(.sirapi-dark) .copy-recipe-table{border-color:#355064;background:#07111d;box-shadow:0 6px 18px rgba(0,0,0,.28)}
:global(.theme-dark) .copy-recipe-table-head,:global(.sirapi-dark) .copy-recipe-table-head{border-color:rgba(255,255,255,.11);color:#f8fafc;background:#0d555a}
:global(.theme-dark) .copy-recipe-item,:global(.theme-dark) .copy-recipe-table-row,:global(.sirapi-dark) .copy-recipe-item,:global(.sirapi-dark) .copy-recipe-table-row{border-color:#2b4153;color:#f8fafc;background:#111c2c}
:global(.theme-dark) .copy-recipe-item:nth-child(even),:global(.theme-dark) .copy-recipe-table-row:nth-child(odd),:global(.sirapi-dark) .copy-recipe-item:nth-child(even),:global(.sirapi-dark) .copy-recipe-table-row:nth-child(odd){background:#102d38}
:global(.theme-dark) .copy-recipe-item:hover,:global(.theme-dark) .copy-recipe-item.active,:global(.theme-dark) .copy-recipe-table-row:hover,:global(.sirapi-dark) .copy-recipe-item:hover,:global(.sirapi-dark) .copy-recipe-item.active,:global(.sirapi-dark) .copy-recipe-table-row:hover{background:#16404a}
:global(.theme-dark) .copy-recipe-item small,:global(.theme-dark) .copy-recipe-item i,:global(.theme-dark) .copy-recipe-table-row small,:global(.theme-dark) .copy-recipe-table-row span,:global(.sirapi-dark) .copy-recipe-item small,:global(.sirapi-dark) .copy-recipe-item i,:global(.sirapi-dark) .copy-recipe-table-row small,:global(.sirapi-dark) .copy-recipe-table-row span{color:#a9bfd3}
:global(.theme-dark) .copy-recipe-item b,:global(.theme-dark) .copy-recipe-table-row code,:global(.sirapi-dark) .copy-recipe-item b,:global(.sirapi-dark) .copy-recipe-table-row code{color:#5eead4}
:global(.theme-dark) .copy-recipe-item em,:global(.sirapi-dark) .copy-recipe-item em{color:#5eead4;background:rgba(20,184,166,.16)}
:global(.theme-dark) .copy-recipe-table-row.racik,:global(.sirapi-dark) .copy-recipe-table-row.racik{background:rgba(45,33,70,.55)}
.spin{animation:spin .9s linear infinite}
@keyframes spin{to{transform:rotate(360deg)}}
@media(max-width:1100px){.copy-recipe-layout{grid-template-columns:1fr}.copy-recipe-list,.copy-recipe-preview{min-height:auto}}
@media(max-width:650px){.copy-recipe-header,.copy-recipe-preview-head,.copy-recipe-item{grid-template-columns:1fr;align-items:flex-start}.copy-recipe-button{width:100%}}
</style>
