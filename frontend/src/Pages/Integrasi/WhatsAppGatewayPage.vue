<script setup>
import { computed, onBeforeUnmount, onMounted, reactive, ref } from 'vue'
import Column from 'primevue/column'
import Dialog from 'primevue/dialog'
import { CheckCircle2, Link2, LoaderCircle, MessageCircle, Plus, QrCode, RefreshCw, Send, Smartphone } from '@lucide/vue'
import DataTable from '../../Components/Ui/DataTable.vue'
import FormInput from '../../Components/Ui/FormInput.vue'
import {
  buatPerangkatWhatsApp,
  daftarPerangkatWhatsApp,
  daftarPesanWhatsApp,
  detailPerangkatWhatsApp,
  hubungkanPerangkatWhatsApp,
  kirimPesanWhatsApp,
  kodePasanganWhatsApp,
  qrPerangkatWhatsApp,
  statusWhatsAppGateway,
} from '../../lib/faisal/api'
import { useNotifikasi } from '../../lib/shared/useNotifikasi'

const props = defineProps({ token: { type: String, required: true } })
const notifikasi = useNotifikasi()

const loading = ref(true)
const proses = ref('')
const error = ref('')
const konfigurasi = ref({ dikonfigurasi: false })
const responsPerangkat = ref([])
const responsPesan = ref([])
const modalQR = ref(false)
const modalPair = ref(false)
const qrURL = ref('')
const perangkatAktif = ref(null)
const kodePairing = ref('')
const formPerangkat = reactive({ nama: '' })
const formPair = reactive({ nomor: '' })
const formPesan = reactive({ device_id: '', tujuan: '', nama_penerima: '', external_id: '', idempotency_key: '', pesan: '' })

function ambilArray(data, candidates = []) {
  if (Array.isArray(data)) return data
  for (const key of candidates) if (Array.isArray(data?.[key])) return data[key]
  if (Array.isArray(data?.items)) return data.items
  return []
}

const perangkat = computed(() => ambilArray(responsPerangkat.value, ['devices', 'perangkat']))
const pesan = computed(() => ambilArray(responsPesan.value, ['messages', 'pesan']))
const jumlahTerhubung = computed(() => perangkat.value.filter((item) => ['connected', 'online'].includes(String(item?.status || item?.connection_status).toLowerCase())).length)
const pilihanPerangkat = computed(() => perangkat.value.map((item) => ({
  label: `${item.name || item.nama || 'Perangkat'}${item.phone ? ` - ${item.phone}` : ''}`,
  value: item.id || item.device_id,
})).filter((item) => item.value))

function idPerangkat(item) { return item?.id || item?.device_id || '' }
function warnaStatus(status) { return ['connected', 'sent', 'delivered', 'read'].includes(String(status).toLowerCase()) ? 'success' : ['failed', 'disconnected'].includes(String(status).toLowerCase()) ? 'danger' : 'pending' }
function warnaStatusPerangkat(item) { return warnaStatus(item?.status || item?.connection_status) }
function labelStatus(status) {
  return ({ connected: 'Terhubung', online: 'Terhubung', disconnected: 'Terputus', queued: 'Antrean', processing: 'Diproses', sent: 'Terkirim', delivered: 'Diterima', read: 'Dibaca', failed: 'Gagal' })[String(status || '').toLowerCase()] || String(status || '-').replaceAll('_', ' ')
}
function formatWaktu(value) {
  if (!value) return '-'
  const tanggal = new Date(value)
  return Number.isNaN(tanggal.getTime()) ? value : tanggal.toLocaleString('id-ID', { dateStyle: 'medium', timeStyle: 'short' })
}
function idempotencyBaru() { return `sirapi-${Date.now()}-${Math.random().toString(36).slice(2, 8)}` }
function tunggu(ms) { return new Promise((resolve) => window.setTimeout(resolve, ms)) }
function qrBelumSiap(err) { return /qr.+not ready|connect the device|poll this endpoint/i.test(String(err?.message || '')) }

async function muatData() {
  loading.value = true
  error.value = ''
  try {
    konfigurasi.value = await statusWhatsAppGateway(props.token)
    if (!konfigurasi.value.dikonfigurasi) {
      error.value = 'URL_WHATSAPP dan KEY_WHATSAPP belum diisi pada env backend.'
      return
    }
    const [dataPerangkat, dataPesan] = await Promise.all([
      daftarPerangkatWhatsApp(props.token),
      daftarPesanWhatsApp(props.token).catch(() => []),
    ])
    responsPerangkat.value = dataPerangkat
    responsPesan.value = dataPesan
    if (!formPesan.device_id && pilihanPerangkat.value.length) formPesan.device_id = pilihanPerangkat.value[0].value
  } catch (err) {
    error.value = err.message || 'WhatsApp Gateway tidak dapat dibaca.'
  } finally {
    loading.value = false
  }
}

async function buatPerangkat() {
  if (proses.value) return
  if (!formPerangkat.nama.trim()) {
    notifikasi.peringatan('Nama perangkat WhatsApp wajib diisi.')
    return
  }
  proses.value = 'buat'
  try {
    await buatPerangkatWhatsApp(props.token, formPerangkat.nama.trim())
    formPerangkat.nama = ''
    notifikasi.sukses('Perangkat WhatsApp berhasil dibuat.')
    await muatData()
  } catch (err) { notifikasi.gagal(err.message) } finally { proses.value = '' }
}

async function refreshPerangkat(item) {
  proses.value = `status-${idPerangkat(item)}`
  try {
    await detailPerangkatWhatsApp(props.token, idPerangkat(item))
    await muatData()
  } catch (err) { notifikasi.gagal(err.message) } finally { proses.value = '' }
}

async function hubungkan(item) {
  proses.value = `connect-${idPerangkat(item)}`
  try {
    const hasil = await hubungkanPerangkatWhatsApp(props.token, idPerangkat(item))
    notifikasi.info(hasil?.already_connected ? 'Perangkat sudah terhubung.' : 'Perangkat sedang dihubungkan. SIRAPI menunggu QR dari gateway...')
    await muatData()
    if (!hasil?.already_connected) await bukaQR(item)
  } catch (err) { notifikasi.gagal(err.message) } finally { proses.value = '' }
}

async function bukaQR(item) {
  perangkatAktif.value = item
  proses.value = `qr-${idPerangkat(item)}`
  try {
    let errorTerakhir = null
    for (let percobaan = 1; percobaan <= 20; percobaan += 1) {
      try {
        const gambar = await qrPerangkatWhatsApp(props.token, idPerangkat(item))
        if (qrURL.value) URL.revokeObjectURL(qrURL.value)
        qrURL.value = URL.createObjectURL(gambar)
        modalQR.value = true
        return
      } catch (err) {
        errorTerakhir = err
        if (!qrBelumSiap(err)) throw err
        if (percobaan < 20) await tunggu(1500)
      }
    }
    throw new Error(qrBelumSiap(errorTerakhir)
      ? 'QR belum tersedia setelah 30 detik. Tekan Hubungkan lalu coba buka QR kembali.'
      : errorTerakhir?.message || 'QR perangkat tidak dapat dibaca.')
  } catch (err) { notifikasi.peringatan(err.message || 'QR belum siap. Hubungkan perangkat lalu coba lagi.') } finally { proses.value = '' }
}

function bukaPair(item) {
  perangkatAktif.value = item
  formPair.nomor = ''
  kodePairing.value = ''
  modalPair.value = true
}

async function mintaKodePairing() {
  if (!formPair.nomor.trim() || !perangkatAktif.value || proses.value) return
  proses.value = 'pair'
  try {
    const hasil = await kodePasanganWhatsApp(props.token, idPerangkat(perangkatAktif.value), formPair.nomor)
    kodePairing.value = hasil?.code || hasil?.pairing_code || hasil?.pair_code || ''
    notifikasi.sukses('Kode pasangan berhasil dibuat.')
  } catch (err) { notifikasi.gagal(err.message) } finally { proses.value = '' }
}

async function kirimPesan() {
  if (proses.value) return
  if (!formPesan.device_id || !formPesan.tujuan.trim() || !formPesan.pesan.trim()) {
    notifikasi.peringatan('Perangkat, nomor tujuan, dan isi pesan wajib diisi.')
    return
  }
  proses.value = 'kirim'
  try {
    if (!formPesan.idempotency_key) formPesan.idempotency_key = idempotencyBaru()
    await kirimPesanWhatsApp(props.token, { ...formPesan })
    notifikasi.sukses('Pesan diterima antrean WhatsApp Gateway. Status 202 belum berarti pesan sudah terkirim.')
    Object.assign(formPesan, { tujuan: '', nama_penerima: '', external_id: '', idempotency_key: idempotencyBaru(), pesan: '' })
    responsPesan.value = await daftarPesanWhatsApp(props.token).catch(() => responsPesan.value)
  } catch (err) { notifikasi.gagal(err.message) } finally { proses.value = '' }
}

onMounted(() => { formPesan.idempotency_key = idempotencyBaru(); muatData() })
onBeforeUnmount(() => { if (qrURL.value) URL.revokeObjectURL(qrURL.value) })
</script>

<template>
  <section class="whatsapp-page">
    <header class="module-header">
      <div class="module-identity">
        <div class="module-icon"><MessageCircle :size="25" /></div>
        <div>
          <span class="eyebrow">INTEGRASI PESAN</span>
          <h1>WhatsApp Gateway</h1>
          <p>Kelola perangkat pengirim dan pesan WhatsApp SIRAPI.</p>
        </div>
      </div>
      <div class="header-actions">
        <span class="configuration-status" :class="loading ? 'checking' : konfigurasi.dikonfigurasi ? 'ready' : 'not-ready'">
          <i /> {{ loading ? 'Memeriksa gateway...' : konfigurasi.dikonfigurasi ? 'Gateway siap' : 'Belum dikonfigurasi' }}
        </span>
        <button type="button" class="secondary-button" :disabled="loading" @click="muatData">
          <RefreshCw :size="16" :class="{ spin: loading }" /> {{ loading ? 'Memuat...' : 'Muat Data' }}
        </button>
      </div>
    </header>

    <div v-if="error" class="whatsapp-alert">{{ error }}</div>
    <template v-else>
      <div class="summary-grid">
        <article><span>Total Perangkat</span><strong>{{ perangkat.length }}</strong><Smartphone :size="21" /></article>
        <article><span>Perangkat Terhubung</span><strong>{{ jumlahTerhubung }}</strong><Link2 :size="21" /></article>
        <article><span>Riwayat Pesan</span><strong>{{ pesan.length }}</strong><MessageCircle :size="21" /></article>
      </div>

      <div class="content-grid">
        <section class="module-card device-section">
          <div class="card-header">
            <div><span class="eyebrow">PERANGKAT PENGIRIM</span><h2>Perangkat WhatsApp</h2><p>Hubungkan nomor WhatsApp menggunakan QR atau kode pasangan.</p></div>
          </div>
          <form class="device-create form-area" @submit.prevent="buatPerangkat">
            <FormInput v-model="formPerangkat.nama" label="Nama Perangkat" placeholder="Contoh: Informasi Rumah Sakit" required />
            <button type="submit" class="primary-button" :disabled="Boolean(proses)"><Plus :size="16" /> Tambah Device</button>
          </form>
          <div v-if="loading" class="empty-state"><LoaderCircle :size="27" class="spin" /><strong>Memuat perangkat...</strong></div>
          <div v-else-if="!perangkat.length" class="empty-state"><Smartphone :size="31" /><strong>Belum ada perangkat</strong><span>Tambahkan perangkat untuk mulai menghubungkan WhatsApp.</span></div>
          <div v-else class="device-list">
            <article v-for="item in perangkat" :key="idPerangkat(item)" class="device-item">
              <div class="device-symbol"><Smartphone :size="20" /></div>
              <div class="device-info">
                <strong>{{ item.name || item.nama || 'Perangkat WhatsApp' }}</strong>
                <span>{{ item.phone || 'Nomor belum terhubung' }}</span>
                <code>Device ID: {{ idPerangkat(item) || '-' }}</code>
              </div>
              <em class="status-pill" :class="warnaStatusPerangkat(item)">{{ labelStatus(item.status || item.connection_status) }}</em>
              <div class="device-actions">
                <button type="button" title="Periksa status" :disabled="Boolean(proses)" @click="refreshPerangkat(item)"><RefreshCw :size="15" /> Status</button>
                <button type="button" title="Hubungkan perangkat" :disabled="Boolean(proses)" @click="hubungkan(item)"><Link2 :size="15" /> Hubungkan</button>
                <button type="button" title="Tampilkan QR" :disabled="Boolean(proses)" @click="bukaQR(item)"><QrCode :size="15" /> QR</button>
                <button type="button" title="Gunakan kode pasangan" :disabled="Boolean(proses)" @click="bukaPair(item)"><Smartphone :size="15" /> Kode</button>
              </div>
            </article>
          </div>
        </section>

        <section class="module-card message-section">
          <div class="card-header">
            <div><span class="eyebrow">PESAN TUNGGAL</span><h2>Kirim Pesan</h2><p>Nomor tujuan otomatis disesuaikan ke format Indonesia 62.</p></div>
          </div>
          <form class="message-form form-area" @submit.prevent="kirimPesan">
            <FormInput v-model="formPesan.device_id" class="span-all" label="Perangkat Pengirim" jenis="select" :options="pilihanPerangkat" placeholder="Pilih perangkat" required />
            <FormInput v-model="formPesan.tujuan" label="Nomor Tujuan" placeholder="6281234567890" required />
            <FormInput v-model="formPesan.nama_penerima" label="Nama Penerima" placeholder="Opsional" />
            <FormInput v-model="formPesan.external_id" class="span-all" label="Referensi" placeholder="Contoh: RM-12345 atau nomor tagihan" />
            <FormInput v-model="formPesan.pesan" class="span-all" label="Isi Pesan" jenis="textarea" :rows="5" placeholder="Tuliskan pesan WhatsApp..." required />
            <div class="send-note span-all"><span><CheckCircle2 :size="15" /> API key tersimpan aman di backend.</span><small>Status antrean belum berarti pesan sudah diterima.</small></div>
            <div class="form-actions span-all"><button type="submit" class="primary-button" :disabled="Boolean(proses) || !pilihanPerangkat.length"><Send :size="16" /> {{ proses === 'kirim' ? 'Mengantrekan...' : 'Kirim Pesan' }}</button></div>
          </form>
        </section>
      </div>

      <section class="module-card history-card">
        <div class="card-header history-header">
          <div><span class="eyebrow">RIWAYAT GATEWAY</span><h2>Pesan Terbaru</h2><p>Lihat tujuan, isi pesan, perangkat pengirim, dan status pengiriman.</p></div>
          <span class="record-count">{{ pesan.length }} pesan</span>
        </div>
        <DataTable :rows="pesan" data-key="id" empty-message="Belum ada riwayat pesan." paginator :rows-per-page="10" :rows-per-page-options="[10, 25]">
          <Column header="Penerima"><template #body="{ data }"><div class="table-primary"><strong>{{ data.recipient_name || data.to || data.recipient || '-' }}</strong><span>{{ data.recipient_name ? (data.to || data.recipient || '-') : (data.external_id || '') }}</span></div></template></Column>
          <Column header="Pesan"><template #body="{ data }"><span class="message-preview">{{ data.text?.body || data.body || data.message || data.type || '-' }}</span></template></Column>
          <Column header="Perangkat"><template #body="{ data }"><span>{{ data.device_name || data.device_id || '-' }}</span></template></Column>
          <Column header="Status"><template #body="{ data }"><em class="status-pill" :class="warnaStatus(data.status)">{{ labelStatus(data.status) }}</em></template></Column>
          <Column header="Waktu"><template #body="{ data }"><span>{{ formatWaktu(data.created_at || data.updated_at) }}</span></template></Column>
        </DataTable>
      </section>
    </template>

    <Dialog v-model:visible="modalQR" modal header="Hubungkan dengan QR" :style="{ width: 'min(92vw, 430px)' }">
      <div class="qr-dialog"><img v-if="qrURL" :src="qrURL" alt="QR WhatsApp" /><strong>{{ perangkatAktif?.name || perangkatAktif?.nama || 'Perangkat WhatsApp' }}</strong><p>Scan melalui menu Perangkat Tertaut pada aplikasi WhatsApp. QR memiliki waktu berlaku terbatas.</p></div>
    </Dialog>
    <Dialog v-model:visible="modalPair" modal header="Hubungkan dengan Pair Code" :style="{ width: 'min(92vw, 520px)' }">
      <div class="pair-dialog"><FormInput v-model="formPair.nomor" label="Nomor WhatsApp Pengirim" placeholder="6281234567890" hint="Gunakan format 62 tanpa tanda + dan tanpa awalan 0." required /><div class="dialog-actions"><button type="button" class="primary-button" :disabled="!formPair.nomor.trim() || proses === 'pair'" @click="mintaKodePairing"><Smartphone :size="16" /> {{ proses === 'pair' ? 'Memproses...' : 'Minta Kode' }}</button></div><div v-if="kodePairing" class="pair-code"><CheckCircle2 :size="20" /><span>Kode pasangan</span><strong>{{ kodePairing }}</strong><small>Masukkan melalui menu Perangkat Tertaut di WhatsApp.</small></div></div>
    </Dialog>
  </section>
</template>

<style scoped>
.whatsapp-page{position:relative;z-index:1;display:grid;width:calc(100% - 40px);box-sizing:border-box;gap:16px;margin:20px;color:var(--text)}
.module-header,.module-card{border:1px solid var(--line);border-radius:16px;background:var(--surface);overflow:hidden}
.module-header{display:flex;align-items:center;justify-content:space-between;gap:20px;padding:20px 22px}
.module-identity,.header-actions{display:flex;align-items:center;gap:14px}.module-icon,.device-symbol{display:grid;place-items:center;flex:0 0 auto;background:rgba(13,148,136,.12);color:#0d9488}.module-icon{width:50px;height:50px;border-radius:14px}.device-symbol{width:40px;height:40px;border-radius:11px}
.eyebrow{font-size:10px;font-weight:700;letter-spacing:.12em;color:#0d9488}.module-header h1,.card-header h2{margin:3px 0;color:var(--text)}.module-header h1{font-size:24px}.card-header h2{font-size:18px}.module-header p,.card-header p{margin:0;color:var(--muted);font-size:13px;line-height:1.5}
.configuration-status{display:inline-flex;align-items:center;gap:7px;white-space:nowrap;color:var(--muted);font-size:12px}.configuration-status i{width:8px;height:8px;border-radius:50%;background:#ef4444}.configuration-status.ready i{background:#10b981}.configuration-status.checking i{background:#f59e0b}
.summary-grid{display:grid;grid-template-columns:repeat(3,minmax(0,1fr));gap:12px}.summary-grid article{position:relative;display:grid;gap:5px;padding:16px 18px;border:1px solid var(--line);border-radius:14px;background:var(--surface)}.summary-grid span{color:var(--muted);font-size:12px}.summary-grid strong{font-size:23px}.summary-grid svg{position:absolute;right:17px;top:50%;transform:translateY(-50%);color:#0d9488}
.content-grid{display:grid;grid-template-columns:minmax(0,1.05fr) minmax(390px,.95fr);gap:16px;align-items:start}.card-header{display:flex;align-items:center;justify-content:space-between;gap:16px;padding:17px 20px;border-bottom:1px solid var(--line)}.form-area{padding:20px;background:var(--surface-soft)}
.device-create{display:grid;grid-template-columns:minmax(220px,1fr) auto;align-items:end;gap:10px}.device-list{display:grid;gap:9px;padding:0 20px 20px;background:var(--surface-soft)}.device-item{display:grid;grid-template-columns:auto minmax(130px,1fr) auto;align-items:center;gap:11px;padding:13px;border:1px solid var(--line);border-radius:12px;background:var(--surface)}.device-info{display:grid;gap:3px;min-width:0}.device-info strong,.device-info span,.device-info code{overflow:hidden;text-overflow:ellipsis;white-space:nowrap}.device-info strong{font-size:14px}.device-info span{color:var(--muted);font-size:12px}.device-info code{margin-top:2px;color:#0d9488;font-size:10px;font-weight:600}.device-actions{grid-column:1/-1;display:flex;flex-wrap:wrap;gap:7px;padding-top:10px;border-top:1px solid var(--line)}
.message-form{display:grid;grid-template-columns:repeat(2,minmax(0,1fr));gap:13px}.span-all{grid-column:1/-1}.send-note{display:flex;align-items:center;justify-content:space-between;gap:12px;color:var(--muted);font-size:11px}.send-note span{display:flex;align-items:center;gap:6px;color:#0d9488}.form-actions,.dialog-actions{display:flex;justify-content:flex-end}
.primary-button,.secondary-button,.device-actions button{display:inline-flex;align-items:center;justify-content:center;gap:7px;min-height:40px;border:1px solid var(--line);border-radius:10px;padding:0 14px;font:inherit;font-size:13px;font-weight:700;cursor:pointer}.primary-button{border-color:#0d9488;background:#0d9488;color:#fff}.secondary-button,.device-actions button{background:var(--surface-soft);color:var(--text)}.device-actions button{min-height:34px;padding:0 10px;font-size:11px}.device-actions button:hover,.secondary-button:hover{border-color:#0d9488;color:#0d9488}button:disabled{opacity:.5;cursor:not-allowed}
.status-pill{display:inline-flex;width:max-content;border-radius:999px;padding:5px 9px;font-size:10px;font-style:normal;font-weight:700;text-transform:capitalize}.success{background:rgba(13,148,136,.13);color:#0d9488}.pending{background:rgba(245,158,11,.14);color:#d97706}.danger{background:rgba(239,68,68,.13);color:#dc2626}
.empty-state{display:flex;min-height:146px;flex-direction:column;align-items:center;justify-content:center;gap:7px;color:var(--muted);padding:20px;background:var(--surface-soft);text-align:center}.empty-state strong{color:var(--text)}.empty-state span{font-size:12px}.whatsapp-alert{padding:15px 17px;border:1px solid rgba(239,68,68,.3);border-radius:12px;background:rgba(239,68,68,.08);color:#dc2626}
.history-header{padding-bottom:15px}.record-count{border-radius:999px;padding:6px 10px;background:var(--surface-soft);color:var(--muted);font-size:11px}.history-card :deep(.data-table-wrap){margin:14px}.history-card :deep(td>span),.history-card :deep(td>strong){display:block}.table-primary{display:grid;gap:3px}.table-primary span{color:var(--muted);font-size:11px}.message-preview{display:block;max-width:480px;white-space:normal;line-height:1.45}
.qr-dialog{display:grid;place-items:center;gap:10px;text-align:center}.qr-dialog img{width:270px;max-width:100%;padding:10px;border:1px solid var(--line);border-radius:14px;background:#fff}.qr-dialog p{margin:0;color:var(--muted);font-size:12px;line-height:1.5}.pair-dialog{display:grid;gap:15px}.pair-code{display:grid;place-items:center;gap:5px;padding:19px;border:1px solid var(--line);border-radius:13px;background:var(--surface-soft)}.pair-code strong{font-size:30px;letter-spacing:.14em;color:#0d9488}.pair-code small{color:var(--muted);text-align:center}.spin{animation:spin 1s linear infinite}@keyframes spin{to{transform:rotate(360deg)}}
@media(max-width:1050px){.content-grid{grid-template-columns:1fr}.device-list{grid-template-columns:repeat(2,minmax(0,1fr))}}
@media(max-width:760px){.module-header{align-items:flex-start}.header-actions{align-items:flex-end;flex-direction:column}.device-list{grid-template-columns:1fr}.summary-grid{grid-template-columns:1fr}.message-form{grid-template-columns:1fr}.span-all{grid-column:span 1}}
@media(max-width:560px){.whatsapp-page{width:calc(100% - 20px);margin:10px}.module-header{display:grid;padding:16px}.header-actions{align-items:stretch}.configuration-status{justify-content:center}.module-header .secondary-button{width:100%}.card-header,.form-area,.device-list{padding-left:14px;padding-right:14px}.device-create{grid-template-columns:1fr}.device-create .primary-button{width:100%}.send-note{align-items:flex-start;flex-direction:column}.device-actions button{flex:1}.history-header{align-items:flex-start}.record-count{white-space:nowrap}}
</style>
