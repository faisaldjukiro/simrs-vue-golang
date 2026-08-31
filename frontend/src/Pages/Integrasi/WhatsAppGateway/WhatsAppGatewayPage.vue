<script setup lang="ts">
import Column from "primevue/column"
import Dialog from "primevue/dialog"
import { CheckCircle2, Link2, LoaderCircle, MessageCircle, Plus, QrCode, RefreshCw, Send, Smartphone } from "@lucide/vue"
import DataTable from "../../../Components/Ui/DataTable.vue"
import FormInput from "../../../Components/Ui/FormInput.vue"
import { useWhatsAppGateway } from "./useWhatsAppGateway"

const props = defineProps({ token: { type: String, required: true } })

const {
  loading,
  proses,
  error,
  konfigurasi,
  modalQR,
  modalPair,
  qrURL,
  perangkatAktif,
  kodePairing,
  formPerangkat,
  formPair,
  formPesan,
  perangkat,
  pesan,
  jumlahTerhubung,
  pilihanPerangkat,
  idPerangkat,
  warnaStatus,
  warnaStatusPerangkat,
  labelStatus,
  formatWaktu,
  muatData,
  buatPerangkat,
  refreshPerangkat,
  hubungkan,
  bukaQR,
  bukaPair,
  mintaKodePairing,
  kirimPesan,
} = useWhatsAppGateway(props)
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

<style src="./whats-app-gateway.css" scoped></style>
