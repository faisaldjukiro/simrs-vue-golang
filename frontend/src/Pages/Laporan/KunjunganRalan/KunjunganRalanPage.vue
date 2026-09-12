<script setup lang="ts">
import {
  BarChart3,
  Download,
  FileBarChart,
  Filter,
  Printer,
  X,
} from "@lucide/vue";
import Column from "primevue/column";
import DataTable from "../../../Components/Ui/DataTable.vue";
import FormInput from "../../../Components/Ui/FormInput.vue";
import InputPencarian from "../../../Components/Ui/InputPencarian.vue";
import { useKunjunganRalan } from "./useKunjunganRalan";
const props = defineProps({ token: { type: String, required: true } });
const {
  loading,
  error,
  data,
  ringkasan,
  filter,
  pilihan,
  berulang,
  firstRow,
  grafikVisible,
  trenHarian,
  kategoriTerbanyak,
  penyakitTerbanyak,
  dokterTerbanyak,
  persenGender,
  persenBaru,
  muat,
  gantiJenis,
  gantiHalaman,
  cariReferensi,
  excel,
  cetak,
} = useKunjunganRalan(props);
</script>
<template>
  <section class="visit-report-module">
    <header class="visit-report-header">
      <div>
        <span>Laporan Pelayanan</span>
        <h1>Kunjungan Rawat Jalan</h1>
      </div>
      <FileBarChart :size="30" />
    </header>
    <nav class="visit-report-tabs">
      <button v-for="tab in [
        { v: 'detail', l: 'Seluruh Kunjungan' },
        { v: 'rekap', l: 'Kunjungan Non Batal' },
        { v: 'berulang', l: 'Kunjungan Berulang' },
      ]" :key="tab.v" :class="{ active: filter.jenis === tab.v }" @click="gantiJenis(tab.v)">
        {{ tab.l }}
      </button>
    </nav>
    <form class="visit-report-filter" @submit.prevent="muat">
      <FormInput v-model="filter.tanggal_mulai" label="Tanggal Mulai" type="date" required />
      <FormInput v-model="filter.tanggal_selesai" label="Sampai Dengan" type="date" required />
      <template v-if="!berulang">
        <FormInput v-model="filter.status" label="Status Kunjungan" jenis="select" :options="[
          { label: 'Semua', value: '' },
          { label: 'Baru', value: 'Baru' },
          { label: 'Lama', value: 'Lama' },
        ]" />
        <InputPencarian v-model="pilihan.poli" label="Poliklinik" placeholder="Cari kode atau nama poli..."
          :search="(q) => cariReferensi('poli', q)" />
        <InputPencarian v-model="pilihan.dokter" label="Dokter" placeholder="Cari kode atau nama dokter..."
          :search="(q) => cariReferensi('dokter', q)" />
        <InputPencarian v-model="pilihan.penjamin" label="Cara Bayar" placeholder="Cari cara bayar..."
          :search="(q) => cariReferensi('penjamin', q)" />
        <InputPencarian v-model="pilihan.kabupaten" label="Kabupaten" placeholder="Cari kabupaten..."
          :search="(q) => cariReferensi('kabupaten', q)" />
        <InputPencarian v-model="pilihan.kecamatan" label="Kecamatan" placeholder="Cari kecamatan..."
          :search="(q) => cariReferensi('kecamatan', q)" />
        <InputPencarian v-model="pilihan.kelurahan" label="Kelurahan" placeholder="Cari kelurahan..."
          :search="(q) => cariReferensi('kelurahan', q)" />
      </template>
      <FormInput v-model="filter.q" class="visit-search" label="Pencarian"
        placeholder="No. rawat, RM, pasien, alamat, diagnosis..." />
      <button type="submit" :disabled="loading">
        <Filter :size="16" />{{ loading ? "Memuat..." : "Tampilkan" }}
      </button>
    </form>
    <div class="visit-report-summary">
      <article>
        <span>Total Data</span><strong>{{ ringkasan.total }}</strong>
      </article>
      <template v-if="!berulang">
        <article>
          <span>Pasien Lama</span><strong>{{ ringkasan.lama }}</strong>
        </article>
        <article>
          <span>Pasien Baru</span><strong>{{ ringkasan.baru }}</strong>
        </article>
      </template><template v-else>
        <article>
          <span>Berulang</span><strong>{{ ringkasan.berulang }}</strong>
        </article>
        <article>
          <span>Tidak Berulang</span><strong>{{ ringkasan.tidak_berulang }}</strong>
        </article>
      </template>
      <article>
        <span>Laki-laki</span><strong>{{ ringkasan.laki_laki }}</strong>
      </article>
      <article>
        <span>Perempuan</span><strong>{{ ringkasan.perempuan }}</strong>
      </article>
    </div>
    <div class="visit-report-tools">
      <span>{{ data.length }} baris ditampilkan</span><button class="chart-button" :class="{ active: grafikVisible }"
        :aria-expanded="grafikVisible"
        aria-controls="grafik-kunjungan-ralan"
        @click="grafikVisible = !grafikVisible">
        <X v-if="grafikVisible" :size="15" />
        <BarChart3 v-else :size="15" />{{
          grafikVisible ? "Tutup Grafik" : "Lihat Grafik"
        }}
      </button><button @click="excel">
        <Download :size="15" /> Excel
      </button><button @click="cetak">
        <Printer :size="15" /> Cetak
      </button>
    </div>
    <section v-if="grafikVisible" id="grafik-kunjungan-ralan" class="visit-analytics">
      <header>
        <div>
          <span>Visualisasi Data</span>
          <h2>Grafik Kunjungan Rawat Jalan</h2>
          <p>Grafik mengikuti periode dan filter laporan yang sedang aktif.</p>
        </div>
        <BarChart3 :size="25" />
      </header>
      <div class="visit-chart-grid">
        <article v-if="!berulang" class="visit-chart-card trend-chart">
          <h3>Tren Kunjungan per Hari</h3>
          <p>Jumlah registrasi rawat jalan berdasarkan tanggal.</p>
          <div v-if="trenHarian.length" class="vertical-chart">
            <div v-for="item in trenHarian" :key="item.label">
              <span><i :style="{ height: `${Math.max(item.persen, 3)}%` }"></i><b>{{ item.nilai }}</b></span><small>{{
                item.label.slice(5) }}</small>
            </div>
          </div>
          <em v-else>Belum ada data untuk divisualisasikan.</em>
        </article>
        <article class="visit-chart-card rank-chart">
          <h3>
            {{ berulang ? "Status Kunjungan" : "Kunjungan per Poliklinik" }}
          </h3>
          <p v-if="berulang">
            Perbandingan kunjungan berulang dan tidak berulang.
          </p>
          <p v-else>
            Seluruh poliklinik pada hasil laporan, diurutkan berdasarkan jumlah
            kunjungan.
          </p>
          <div v-if="kategoriTerbanyak.length" class="horizontal-chart">
            <div v-for="item in kategoriTerbanyak" :key="item.label">
              <label><span>{{ item.label }}</span><b>{{ item.nilai }}</b></label><i><u
                  :style="{ width: `${item.persen}%` }"></u></i>
            </div>
          </div>
          <em v-else>Belum ada data untuk divisualisasikan.</em>
        </article>
        <article v-if="!berulang" class="visit-chart-card rank-chart">
          <h3>Kunjungan berdasarkan Penyakit</h3>
          <p>Seluruh diagnosis pada hasil laporan.</p>
          <div v-if="penyakitTerbanyak.length" class="horizontal-chart">
            <div v-for="item in penyakitTerbanyak" :key="item.label">
              <label><span :title="item.label">{{ item.label }}</span><b>{{ item.nilai }}</b></label><i><u
                  :style="{ width: `${item.persen}%` }"></u></i>
            </div>
          </div>
          <em v-else>Belum ada diagnosis yang tercatat.</em>
        </article>
        <article v-if="!berulang" class="visit-chart-card rank-chart">
          <h3>Kunjungan berdasarkan Dokter</h3>
          <p>Jumlah kunjungan yang dilayani masing-masing dokter.</p>
          <div v-if="dokterTerbanyak.length" class="horizontal-chart">
            <div v-for="item in dokterTerbanyak" :key="item.label">
              <label><span :title="item.label">{{ item.label }}</span><b>{{ item.nilai }}</b></label><i><u
                  :style="{ width: `${item.persen}%` }"></u></i>
            </div>
          </div>
          <em v-else>Belum ada data dokter.</em>
        </article>
        <article class="visit-chart-card donut-card">
          <h3>Jenis Kelamin</h3>
          <p>Komposisi pasien pada hasil laporan.</p>
          <div class="donut-content">
            <i class="donut gender" :style="{ '--value': `${persenGender}%` }"><b>{{ persenGender
                }}%</b><small>Laki-laki</small></i>
            <ul>
              <li>
                <i></i>Laki-laki <b>{{ ringkasan.laki_laki }}</b>
              </li>
              <li>
                <i></i>Perempuan <b>{{ ringkasan.perempuan }}</b>
              </li>
            </ul>
          </div>
        </article>
        <article v-if="!berulang" class="visit-chart-card donut-card">
          <h3>Status Kunjungan</h3>
          <p>Perbandingan pasien baru dan lama.</p>
          <div class="donut-content">
            <i class="donut status" :style="{ '--value': `${persenBaru}%` }"><b>{{ persenBaru }}%</b><small>Pasien
                baru</small></i>
            <ul>
              <li>
                <i></i>Baru <b>{{ ringkasan.baru }}</b>
              </li>
              <li>
                <i></i>Lama <b>{{ ringkasan.lama }}</b>
              </li>
            </ul>
          </div>
        </article>
      </div>
    </section>
    <div v-if="error" class="patient-error">{{ error }}</div>
    <DataTable v-else class="visit-report-table" :rows="data" :data-key="berulang ? 'no_rm' : 'no_rawat'"
      :loading="loading" paginator :rows-per-page="25" :rows-per-page-options="[25, 50, 100]"
      empty-message="Data kunjungan tidak ditemukan." @page="gantiHalaman">
      <Column header="No." style="width: 60px; min-width: 60px; text-align: center"><template
          #body="{ index }"><strong>{{ firstRow + index + 1 }}</strong></template></Column>
      <template v-if="!berulang">
        <Column header="Tanggal / No. Rawat" style="min-width: 165px"><template #body="{ data: r }"><strong>{{ r.tanggal
              }} {{ r.jam }}</strong><small>{{ r.no_rawat }}</small></template>
        </Column>
        <Column header="RM / Pasien" style="min-width: 220px"><template #body="{ data: r }"><strong>{{ r.nama_pasien
              }}</strong><small>{{ r.no_rm }} · {{ r.status_daftar }}</small></template>
        </Column>
        <Column header="Umur" style="min-width: 75px"><template #body="{ data: r }"><strong>{{ r.umur || "-"
              }}</strong></template></Column>
        <Column header="Jenis Kelamin" style="min-width: 130px"><template #body="{ data: r }"><strong>{{
          r.jenis_kelamin === "L"
            ? "Laki-laki"
            : r.jenis_kelamin === "P"
              ? "Perempuan"
              : "-"
        }}</strong></template></Column>
        <Column field="alamat" header="Alamat" style="min-width: 300px" />
        <Column header="Diagnosis" style="min-width: 320px"><template #body="{ data: r }"><strong>{{ r.kode_diagnosa ||
              "-" }}</strong><small>{{ r.diagnosa || "-" }}</small></template>
        </Column>
        <Column header="Pelayanan" style="min-width: 240px"><template #body="{ data: r }"><strong>{{ r.poli
              }}</strong><small>{{ r.dokter }}</small></template></Column>
        <Column header="Penjamin / SEP" style="min-width: 210px"><template #body="{ data: r }"><strong>{{ r.penjamin
              }}</strong><small>{{ r.no_sep || "-" }}</small></template></Column>
      </template>
      <template v-else>
        <Column field="no_rm" header="No. RM" />
        <Column field="nama_pasien" header="Nama Pasien" style="min-width: 220px" />
        <Column field="tanggal_lahir" header="Tanggal Lahir" />
        <Column field="alamat" header="Alamat" style="min-width: 300px" />
        <Column header="Jenis Kelamin"><template #body="{ data: r }">{{
          r.jenis_kelamin === "L"
            ? "Laki-laki"
            : r.jenis_kelamin === "P"
              ? "Perempuan"
              : "-"
        }}</template>
        </Column>
        <Column field="kode_diagnosa" header="Kode Diagnosis" style="min-width: 220px" />
        <Column field="status_kunjungan" header="Status" />
        <Column field="jumlah_kunjungan" header="Jumlah" />
      </template>
    </DataTable>
  </section>
</template>
<style src="../../../Components/Ui/report.css" scoped></style>
