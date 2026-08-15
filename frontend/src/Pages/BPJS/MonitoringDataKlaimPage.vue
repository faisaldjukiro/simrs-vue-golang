<script setup>
import { computed, ref } from 'vue'
import { FileSpreadsheet, LoaderCircle, Search } from '@lucide/vue'
import Column from 'primevue/column'
import DataTable from '../../Components/Ui/DataTable.vue'
import DatePicker from '../../Components/Ui/DatePicker.vue'
import Select from '../../Components/Ui/Select.vue'
import { monitoringDataKlaim } from '../../lib/faisal/api'
import { useNotifikasi } from '../../lib/shared/useNotifikasi'

const props = defineProps({
  token: { type: String, required: true },
})

const notifikasi = useNotifikasi()
const hariIni = new Date()
const tanggalMulai = ref(new Date(hariIni.getFullYear(), hariIni.getMonth(), hariIni.getDate()))
const tanggalSelesai = ref(new Date(hariIni.getFullYear(), hariIni.getMonth(), hariIni.getDate()))
const jenisPelayanan = ref('semua')
const statusKlaim = ref('3')
const pencarian = ref('')
const sedangMemuat = ref(false)
const hasil = ref(null)
const pesanError = ref('')

const pilihanJenisPelayanan = [
  { label: 'Rawat Inap & Rawat Jalan', value: 'semua' },
  { label: 'Rawat Inap', value: '1' },
  { label: 'Rawat Jalan', value: '2' },
]

const pilihanStatusKlaim = [
  { label: 'Proses', value: '1' },
  { label: 'Pending', value: '2' },
  { label: 'Terbayar', value: '3' },
]

const daftarKlaim = computed(() => {
  const klaim = hasil.value?.response?.klaim
  return Array.isArray(klaim) ? klaim : []
})

const daftarTampil = computed(() => {
  const kataKunci = pencarian.value.trim().toLowerCase()
  if (!kataKunci) return daftarKlaim.value

  return daftarKlaim.value.filter((item) => [
    item?.noSEP,
    item?.noFPK,
    item?.peserta?.noMR,
    item?.peserta?.noKartu,
    item?.peserta?.nama,
    item?.poli,
    item?.jenisPelayanan,
    item?.Inacbg?.kode,
    item?.Inacbg?.nama,
    item?.status,
  ].some((nilai) => String(nilai || '').toLowerCase().includes(kataKunci)))
})

const totalPengajuan = computed(() => jumlahBiaya('byPengajuan'))
const totalDisetujui = computed(() => jumlahBiaya('bySetujui'))
const totalTarifRS = computed(() => jumlahBiaya('byTarifRS'))
const periode = computed(() => hasil.value?.periode || null)
const tanggalTanpaData = computed(() => hasil.value?.tanggal_tanpa_data || [])
const tanggalGagal = computed(() => hasil.value?.tanggal_gagal || [])

function jumlahBiaya(field) {
  return daftarKlaim.value.reduce((total, item) => total + angka(item?.biaya?.[field]), 0)
}

function angka(nilai) {
  const hasilAngka = Number(String(nilai ?? '0').replace(/[^0-9.-]/g, ''))
  return Number.isFinite(hasilAngka) ? hasilAngka : 0
}

function rupiah(nilai) {
  return new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    maximumFractionDigits: 0,
  }).format(angka(nilai))
}

function tanggalApi(nilai) {
  if (!(nilai instanceof Date) || Number.isNaN(nilai.getTime())) return ''
  const tahun = nilai.getFullYear()
  const bulan = String(nilai.getMonth() + 1).padStart(2, '0')
  const tanggal = String(nilai.getDate()).padStart(2, '0')
  return `${tahun}-${bulan}-${tanggal}`
}

function selisihHari(mulai, selesai) {
  const awal = Date.UTC(mulai.getFullYear(), mulai.getMonth(), mulai.getDate())
  const akhir = Date.UTC(selesai.getFullYear(), selesai.getMonth(), selesai.getDate())
  return Math.floor((akhir - awal) / 86400000) + 1
}

async function tampilkanData() {
  if (sedangMemuat.value) return
  if (!(tanggalMulai.value instanceof Date) || !(tanggalSelesai.value instanceof Date)) {
    notifikasi.peringatan('Tanggal mulai dan tanggal selesai wajib dipilih.')
    return
  }

  const jumlahHari = selisihHari(tanggalMulai.value, tanggalSelesai.value)
  if (jumlahHari < 1) {
    notifikasi.peringatan('Tanggal selesai tidak boleh sebelum tanggal mulai.')
    return
  }
  if (jumlahHari > 31) {
    notifikasi.peringatan('Rentang tanggal maksimal 31 hari.')
    return
  }

  sedangMemuat.value = true
  pesanError.value = ''
  try {
    hasil.value = await monitoringDataKlaim(props.token, {
      tanggal_mulai: tanggalApi(tanggalMulai.value),
      tanggal_selesai: tanggalApi(tanggalSelesai.value),
      jenis_pelayanan: jenisPelayanan.value,
      status_klaim: statusKlaim.value,
    })

    if (daftarKlaim.value.length === 0) {
      notifikasi.peringatan('Data klaim tidak ditemukan pada filter yang dipilih.', 'Data Kosong')
    } else if (tanggalGagal.value.length > 0) {
      notifikasi.peringatan(`${daftarKlaim.value.length} klaim ditemukan, tetapi ${tanggalGagal.value.length} tanggal gagal diproses BPJS.`, 'Selesai dengan Peringatan')
    } else {
      notifikasi.sukses(`${daftarKlaim.value.length} data klaim berhasil ditampilkan.`)
    }
  } catch (error) {
    hasil.value = null
    pesanError.value = error.message || 'Data klaim BPJS gagal dimuat.'
    notifikasi.gagal(pesanError.value)
  } finally {
    sedangMemuat.value = false
  }
}
</script>

<template>
  <section class="bpjs-claim-module">
    <header class="bpjs-claim-header">
      <div>
        <span>Monitoring BPJS VClaim</span>
        <h1>Data Klaim BPJS</h1>
        <p>Menampilkan data klaim berdasarkan tanggal pulang, jenis pelayanan, dan status klaim.</p>
      </div>
      <i><FileSpreadsheet :size="25" /></i>
    </header>

    <form class="bpjs-claim-filter" @submit.prevent="tampilkanData">
      <label>
        <span>Tanggal Mulai</span>
        <DatePicker v-model="tanggalMulai" :disabled="sedangMemuat" />
      </label>
      <label>
        <span>Sampai Dengan</span>
        <DatePicker v-model="tanggalSelesai" :disabled="sedangMemuat" />
      </label>
      <label>
        <span>Jenis Pelayanan</span>
        <Select v-model="jenisPelayanan" :options="pilihanJenisPelayanan" :disabled="sedangMemuat" />
      </label>
      <label>
        <span>Status Klaim</span>
        <Select v-model="statusKlaim" :options="pilihanStatusKlaim" :disabled="sedangMemuat" />
      </label>
      <button type="submit" :disabled="sedangMemuat">
        <LoaderCircle v-if="sedangMemuat" class="spin" :size="17" />
        <Search v-else :size="17" />
        {{ sedangMemuat ? 'Menarik Data...' : 'Tampilkan' }}
      </button>
    </form>

    <div v-if="pesanError" class="patient-error">{{ pesanError }}</div>

    <template v-if="hasil">
      <div class="bpjs-claim-summary">
        <article><span>Total Klaim</span><strong>{{ daftarKlaim.length }}</strong></article>
        <article><span>Total Tarif RS</span><strong>{{ rupiah(totalTarifRS) }}</strong></article>
        <article><span>Total Pengajuan</span><strong>{{ rupiah(totalPengajuan) }}</strong></article>
        <article><span>Total Disetujui</span><strong>{{ rupiah(totalDisetujui) }}</strong></article>
      </div>

      <div class="bpjs-claim-information">
        <span v-if="periode">
          Periode {{ periode.tanggal_mulai }} s.d. {{ periode.tanggal_selesai }} ·
          {{ periode.jumlah_tanggal_berhasil }} tanggal berhasil
        </span>
        <span v-if="tanggalTanpaData.length">{{ tanggalTanpaData.length }} tanggal tanpa data</span>
        <span v-if="tanggalGagal.length" class="warning">{{ tanggalGagal.length }} tanggal gagal diproses VClaim</span>
      </div>

      <label class="bpjs-claim-search">
        <Search :size="18" />
        <input v-model="pencarian" type="search" placeholder="Cari No. SEP, No. RM, pasien, poli, atau INA-CBG..." />
      </label>

      <DataTable
        class="bpjs-claim-table patient-list-table"
        :rows="daftarTampil"
        data-key="noSEP"
        empty-message="Data klaim tidak ditemukan."
        paginator
        :rows-per-page="10"
        :rows-per-page-options="[10, 25, 50]"
        :total-records="daftarTampil.length"
      >
        <Column header="No.">
          <template #body="{ index }"><strong class="table-number">{{ index + 1 }}</strong></template>
        </Column>
        <Column header="No. SEP / FPK">
          <template #body="{ data }">
            <div class="bpjs-claim-cell"><strong><code>{{ data.noSEP || '-' }}</code></strong><span>FPK {{ data.noFPK || 'belum ada' }}</span></div>
          </template>
        </Column>
        <Column header="No. RM">
          <template #body="{ data }"><strong><code>{{ data.peserta?.noMR || '-' }}</code></strong></template>
        </Column>
        <Column header="Nama Peserta">
          <template #body="{ data }">
            <div class="bpjs-claim-cell"><strong>{{ data.peserta?.nama || '-' }}</strong><span>No. Kartu {{ data.peserta?.noKartu || '-' }}</span></div>
          </template>
        </Column>
        <Column header="Jenis Rawat">
          <template #body="{ data }"><span class="bpjs-service-badge">{{ data.jenisPelayanan || '-' }}</span></template>
        </Column>
        <Column header="Poli / Kelas">
          <template #body="{ data }">
            <div class="bpjs-claim-cell"><strong>{{ data.poli || '-' }}</strong><span>Kelas rawat {{ data.kelasRawat || '-' }}</span></div>
          </template>
        </Column>
        <Column header="INA-CBG">
          <template #body="{ data }">
            <div class="bpjs-claim-cell"><strong>{{ data.Inacbg?.kode || '-' }}</strong><span>{{ data.Inacbg?.nama || '-' }}</span></div>
          </template>
        </Column>
        <Column header="Tanggal">
          <template #body="{ data }">
            <div class="bpjs-claim-cell"><strong>{{ data.tglSep || '-' }}</strong><span>Pulang {{ data.tglPulang || '-' }}</span></div>
          </template>
        </Column>
        <Column header="Status">
          <template #body="{ data }"><span class="bpjs-claim-status">{{ data.status || '-' }}</span></template>
        </Column>
        <Column header="Tarif RS">
          <template #body="{ data }"><strong>{{ rupiah(data.biaya?.byTarifRS) }}</strong></template>
        </Column>
        <Column header="Tarif Grouper">
          <template #body="{ data }"><strong>{{ rupiah(data.biaya?.byTarifGruper) }}</strong></template>
        </Column>
        <Column header="Pengajuan">
          <template #body="{ data }"><strong>{{ rupiah(data.biaya?.byPengajuan) }}</strong></template>
        </Column>
        <Column header="Disetujui">
          <template #body="{ data }"><strong>{{ rupiah(data.biaya?.bySetujui) }}</strong></template>
        </Column>
        <Column header="Top Up">
          <template #body="{ data }"><strong>{{ rupiah(data.biaya?.byTopup) }}</strong></template>
        </Column>
      </DataTable>
    </template>

    <div v-else-if="sedangMemuat" class="patient-loading-panel">
      <LoaderCircle class="spin" :size="28" />
      <strong>Sedang menarik data klaim BPJS...</strong>
      <span>Data setiap tanggal sedang dibaca dari VClaim.</span>
    </div>

    <div v-else class="bpjs-claim-empty">
      <FileSpreadsheet :size="34" />
      <strong>Pilih periode klaim</strong>
      <span>Atur tanggal dan filter di atas, lalu klik Tampilkan.</span>
    </div>
  </section>
</template>
