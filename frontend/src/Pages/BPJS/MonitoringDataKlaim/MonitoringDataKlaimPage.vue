<script setup lang="ts">
import { FileSpreadsheet, LoaderCircle, Search } from "@lucide/vue"
import Column from "primevue/column"
import DataTable from "../../../Components/Ui/DataTable.vue"
import DatePicker from "../../../Components/Ui/DatePicker.vue"
import Select from "../../../Components/Ui/Select.vue"
import { useMonitoringDataKlaim } from "./useMonitoringDataKlaim"

const props = defineProps({
  token: { type: String, required: true },
})

const {
  tanggalMulai,
  tanggalSelesai,
  jenisPelayanan,
  statusKlaim,
  pencarian,
  sedangMemuat,
  hasil,
  pesanError,
  pilihanJenisPelayanan,
  pilihanStatusKlaim,
  daftarKlaim,
  daftarTampil,
  totalPengajuan,
  totalDisetujui,
  totalTarifRS,
  periode,
  tanggalTanpaData,
  tanggalGagal,
  rupiah,
  tampilkanData,
} = useMonitoringDataKlaim(props)
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
