<script setup>
import { ArrowLeft, CheckCircle2, ClipboardList, Database, Edit3, FileSpreadsheet, FileText, Layers, LoaderCircle, PanelLeftClose, PanelLeftOpen, RefreshCw, Search, Send } from "@lucide/vue"
import Column from "primevue/column"
import AutoComplete from "primevue/autocomplete"
import DataTable from "../../Components/Ui/DataTable.vue"
import DatePicker from "../../Components/Ui/DatePicker.vue"
import Select from "../../Components/Ui/Select.vue"
import { useIdrg } from "./Idrg/useIdrg.js"

const props = defineProps({
  token: { type: String, default: '' },
  user: { type: Object, default: null },
})

const {
  loadingDaftar,
  loadingDetail,
  loadingCodingIdrg,
  loadingCodingInacbg,
  loadingProses,
  loadingImportInacbg,
  loadingNewClaimOtomatis,
  pasienTerpilih,
  menuSidebarAktif,
  tabAktif,
  sidebarTertutup,
  diagnosa,
  prosedur,
  diagnosaInacbg,
  prosedurInacbg,
  pilihanSpecialCmg,
  daftarPasien,
  paginasi,
  filter,
  formKlaim,
  menuSidebar,
  opsiJenisRawat,
  opsiKelasRawat,
  opsiKelasPelayanan,
  opsiCaraMasuk,
  opsiCaraPulang,
  opsiCob,
  opsiKodeTarif,
  formTarifRS,
  tanggalFilter,
  daftarTarifRS,
  responseIdrgTerakhir,
  responseInacbgTerakhir,
  opsiTopupIdrg,
  idrgSudahFinal,
  bolehBukaIdrg,
  bolehBukaInacbg,
  idrgBisaEditUlang,
  inacbgSudahFinal,
  klaimSudahTerkirim,
  klaimSudahFinal,
  inacbgBisaEditUlang,
  idrgGroupingFinal,
  inacbgGroupingFinal,
  hasilGroupingIdrgTabel,
  hasilGroupingInacbgTabel,
  opsiSpecialCmgPerKategori,
  adaOpsiSpecialCmg,
  inacbgGroupingGagal,
  terapkanFilter,
  gantiHalaman,
  toggleSidebar,
  pilihPasien,
  pilihMenuSidebar,
  muatCodingPasien,
  hapusDiagnosa,
  hapusProsedur,
  searchHasilDiagnosa,
  activeSubstitusiDiagnosa,
  substitusiKeyword,
  searchHasilProsedur,
  activeSubstitusiProsedur,
  searchHasilDiagnosaInacbg,
  searchHasilProsedurInacbg,
  substitusiDiagnosaInacbg,
  substitusiProsedurInacbg,
  activeSubstitusiDiagnosaInacbg,
  activeSubstitusiProsedurInacbg,
  muatCodingIdrg,
  onSearchProsedur,
  applySubstitusiProsedur,
  onSearchDiagnosa,
  applySubstitusi,
  setPrimaryDiagnosa,
  bukaEditorDiagnosa,
  bukaEditorProsedur,
  pilihTabDetail,
  muatCodingInacbg,
  importIdrgKeInacbg,
  onSearchDiagnosaInacbg,
  onSearchProsedurInacbg,
  terapkanSubstitusiDiagnosaInacbg,
  terapkanSubstitusiProsedurInacbg,
  hapusDiagnosaInacbg,
  hapusProsedurInacbg,
  setPrimaryDiagnosaInacbg,
  bukaEditorDiagnosaInacbg,
  bukaEditorProsedurInacbg,
  jalankanProses,
  parseTanggal,
  formatTanggal,
  parseWaktu,
  formatWaktu,
  labelJenisKelamin,
  statusKlaimLabel,
  statusKlaimHeader,
  kelasStatusKlaim,
  aksiKlaimDaftar,
  kelasAksiKlaimDaftar,
  labelTanggalPulang,
  labelJenisRawat,
  labelDiagnosa,
  rupiah,
  ubahTarifRS,
  tanggalIndo,
  tipeKlaim,
  umurPasien,
  losHari,
  losJam,
  totalTarifRS,
} = useIdrg(props)
</script>

<template>
  <section class="idrg-module" :class="{ 'sidebar-closed': sidebarTertutup }">
    <aside class="idrg-sidebar">
      <div class="idrg-brand">
        <i><FileSpreadsheet :size="23" /></i>
        <span v-if="!sidebarTertutup">
          <small>Bridging</small>
          <strong>E-Klaim</strong>
        </span>
        <button type="button" class="idrg-sidebar-toggle" @click="toggleSidebar" :title="sidebarTertutup ? 'Buka Sidebar' : 'Sembunyikan Sidebar'">
          <PanelLeftOpen v-if="sidebarTertutup" :size="18" />
          <PanelLeftClose v-else :size="18" />
        </button>
      </div>

      <nav>
        <button
          v-for="menu in menuSidebar"
          :key="menu.id"
          type="button"
          :class="{ active: menuSidebarAktif === menu.id }"
          @click="pilihMenuSidebar(menu.id)"
        >
          <component :is="menu.icon" :size="18" />
          <span>{{ menu.label }}</span>
          <b v-if="menu.badge">{{ paginasi.total ?? 0 }}</b>
        </button>
      </nav>

      <div class="idrg-sidebar-note">
        <strong>IDRG / INA-CBG</strong>
        <span>Alur dibuat mengikuti SIMRS lama, dipindahkan bertahap ke SIRAPI.</span>
      </div>
    </aside>

    <main class="idrg-main">
      <template v-if="menuSidebarAktif === 'pengajuan'">
        <form v-if="!pasienTerpilih" class="idrg-filter" @submit.prevent="terapkanFilter">
          <label>
            <span>Tanggal</span>
            <DatePicker v-model="tanggalFilter" placeholder="Tanggal pasien" :disabled="loadingDaftar" />
          </label>
          <label>
            <span>Jenis Rawat</span>
            <Select v-model="filter.jenis_rawat" :options="opsiJenisRawat" :disabled="loadingDaftar" />
          </label>
          <label class="idrg-search">
            <span>Pencarian</span>
            <i>
              <Search :size="17" />
              <input v-model="filter.cari" type="search" placeholder="Cari no rawat, RM, pasien, SEP..." :disabled="loadingDaftar" />
            </i>
          </label>
          <button type="submit" :disabled="loadingDaftar">
            <LoaderCircle v-if="loadingDaftar" class="spin" :size="17" />
            <Search v-else :size="17" />
            Cari
          </button>
        </form>

        <div v-if="!pasienTerpilih" class="idrg-card idrg-list">
          <DataTable
            class="idrg-claim-table"
            :rows="daftarPasien"
            data-key="no_rawat"
            :loading="loadingDaftar"
            loading-message="Menarik pasien E-Klaim..."
            empty-message="Data pasien klaim tidak ditemukan."
            paginator
            lazy
            :first="(paginasi.halaman - 1) * paginasi.batas"
            :rows-per-page="paginasi.batas"
            :total-records="paginasi.total"
            @page="gantiHalaman"
          >
            <Column header="No SEP / Rawat">
              <template #body="{ data }">
                <strong class="idrg-sep">{{ data.no_sep || 'SEP belum ada' }}</strong>
                <span>{{ data.no_rawat }}</span>
              </template>
            </Column>
            <Column header="No RM">
              <template #body="{ data }">
                <strong class="idrg-rm">{{ data.no_rkm_medis }}</strong>
              </template>
            </Column>
            <Column header="Nama Pasien">
              <template #body="{ data }">
                <strong>{{ data.nm_pasien }}</strong>
                <span>{{ labelJenisRawat(data) }}</span>
              </template>
            </Column>
            <Column header="Dokter">
              <template #body="{ data }">
                <strong>{{ data.nm_dokter || '-' }}</strong>
                <span>{{ data.penanggung || '-' }}</span>
              </template>
            </Column>
            <Column header="Ruang/Poli">
              <template #body="{ data }">
                <strong>{{ data.kamar || '-' }}</strong>
              </template>
            </Column>
            <Column header="Tanggal">
              <template #body="{ data }">
                <strong>{{ data.tgl_masuk }}</strong>
                <span v-if="labelTanggalPulang(data)">Pulang {{ labelTanggalPulang(data) }}</span>
              </template>
            </Column>
            <Column header="Diagnosa">
              <template #body="{ data }">
                <strong>{{ labelDiagnosa(data) }}</strong>
              </template>
            </Column>
            <Column header="Aksi">
              <template #body="{ data }">
                <div class="idrg-action-cell">
                  <button class="table-action-button" :class="kelasAksiKlaimDaftar(data)" type="button" @click="pilihPasien(data)">
                    <Database :size="13" />
                    {{ aksiKlaimDaftar(data) }}
                  </button>
                </div>
              </template>
            </Column>
          </DataTable>
        </div>

        <div v-else class="idrg-detail-page">
          <header class="idrg-detail-top">
            <div class="idrg-detail-title">
              <button class="idrg-icon-back" type="button" @click="pasienTerpilih = null">
                <ArrowLeft :size="16" />
              </button>
              <div>
                <h2>
                  {{ pasienTerpilih.nm_pasien }}
                  <span class="idrg-status-pill" :class="kelasStatusKlaim(pasienTerpilih.status_klaim)">
                    {{ statusKlaimHeader(pasienTerpilih.status_klaim) }}
                  </span>
                </h2>
                <p>RM {{ pasienTerpilih.no_rkm_medis }} · {{ labelJenisKelamin(pasienTerpilih.jk) }} · SEP: <b>{{ pasienTerpilih.no_sep || '-' }}</b></p>
              </div>
            </div>
            <div class="idrg-detail-tools">
              <button type="button" :disabled="loadingDetail" @click="muatCodingPasien">
                <LoaderCircle v-if="loadingDetail" class="spin" :size="15" />
                <RefreshCw v-else :size="15" />
                Cek Status
              </button>
              <button type="button" :disabled="!pasienTerpilih.no_sep || !klaimSudahFinal || Boolean(loadingProses)" :title="klaimSudahFinal ? '' : 'Final Klaim harus berhasil terlebih dahulu'" @click="jalankanProses('cetak_klaim')">
                <LoaderCircle v-if="loadingProses === 'cetak_klaim'" class="spin" :size="15" />
                <FileText v-else :size="15" />
                Cetak
              </button>
            </div>
          </header>

          <nav class="idrg-tabs idrg-detail-tabs">
            <button :class="{ active: tabAktif === 'data-pasien' }" type="button" @click="pilihTabDetail('data-pasien')">Data Pasien</button>
            <button :class="{ active: tabAktif === 'idrg' }" type="button" :disabled="!bolehBukaIdrg" :title="bolehBukaIdrg ? '' : 'Set Data Klaim harus berhasil terlebih dahulu'" @click="pilihTabDetail('idrg')">IDRG</button>
            <button :class="{ active: tabAktif === 'inacbg' }" type="button" :disabled="!bolehBukaInacbg" :title="bolehBukaInacbg ? '' : 'Final IDRG harus berhasil terlebih dahulu'" @click="pilihTabDetail('inacbg')">INA-CBG</button>
          </nav>

          <section v-if="tabAktif === 'data-pasien'" class="idrg-detail-content">
            <article class="idrg-claim-sheet">
              <div class="claim-summary-grid">
                <span><small>Tanggal Masuk</small><strong>{{ tanggalIndo(formKlaim.tgl_masuk) }}</strong></span>
                <span><small>Tanggal Pulang</small><strong>{{ tanggalIndo(formKlaim.tgl_pulang) }}</strong></span>
                <span><small>Jaminan</small><strong>{{ formKlaim.jaminan }}</strong></span>
                <span><small>No. SEP</small><strong>{{ pasienTerpilih.no_sep || '-' }}</strong></span>
                <span><small>Tipe</small><strong>{{ tipeKlaim() }}</strong></span>
                <span><small>CBG</small><strong>-</strong></span>
                <span><small>Status</small><strong>{{ statusKlaimLabel(pasienTerpilih.status_klaim) }}</strong></span>
                <span><small>Petugas</small><strong>{{ props.user?.name || '-' }}</strong></span>
              </div>

              <div class="claim-sheet-top">
                <label>
                  <small>Jaminan / Cara Bayar</small>
                  <select v-model="formKlaim.jaminan">
                    <option value="JKN">JKN</option>
                    <option value="UMUM">UMUM</option>
                  </select>
                </label>
                <label>
                  <small>No. Peserta</small>
                  <input :value="pasienTerpilih.no_kartu || '-'" type="text" disabled />
                </label>
                <label>
                  <small>No. SEP</small>
                  <input :value="pasienTerpilih.no_sep || '-'" type="text" disabled />
                </label>
                <label>
                  <small>COB</small>
                  <select v-model="formKlaim.cob_cd">
                    <option v-for="item in opsiCob" :key="item.value" :value="item.value">{{ item.label }}</option>
                  </select>
                </label>
              </div>

              <div class="claim-sheet-row">
                <b>Jenis Rawat</b>
                <div class="claim-value">
                  <label class="claim-radio"><input v-model="formKlaim.jenis_rawat" type="radio" value="2" /> Jalan</label>
                  <label class="claim-radio"><input v-model="formKlaim.jenis_rawat" type="radio" value="1" /> Inap</label>
                  <template v-if="formKlaim.jenis_rawat === '2'">
                    <label class="claim-radio claim-help"><span>?</span><input :checked="Number(formKlaim.tarif_poli_eks) > 0" type="checkbox" @change="formKlaim.tarif_poli_eks = $event.target.checked ? '1' : '0'" /> Kelas Eksekutif</label>
                  </template>
                  <template v-else>
                    <label class="claim-radio claim-help"><span>?</span><input v-model="formKlaim.upgrade_class_ind" type="checkbox" /> Naik/Turun Kelas</label>
                    <label class="claim-radio claim-help"><span>?</span><input v-model="formKlaim.icu_indikator" type="checkbox" /> Ada Rawat Intensif</label>
                  </template>
                </div>
                <b>Kelas Hak</b>
                <div class="claim-value">
                  <template v-if="formKlaim.jenis_rawat === '1'">
                    <label v-for="item in opsiKelasRawat" :key="item.value" class="claim-radio">
                      <input v-model="formKlaim.kelas_rawat" type="radio" :value="item.value" />
                      {{ item.label }}
                    </label>
                  </template>
                  <span v-else>-</span>
                </div>
              </div>

              <div v-if="formKlaim.jenis_rawat === '1' && formKlaim.upgrade_class_ind" class="claim-sheet-row claim-upgrade-row">
                <b><span class="claim-help-icon">?</span> Kelas Pelayanan</b>
                <div class="claim-value claim-upgrade-class">
                  <label v-for="item in opsiKelasPelayanan" :key="item.value" class="claim-radio">
                    <input v-model="formKlaim.upgrade_class_class" type="radio" :value="item.value" />
                    {{ item.label }}
                  </label>
                </div>
                <b><span class="claim-help-icon">?</span> Lama <span>(hari)</span></b>
                <div class="claim-value">
                  <input v-model="formKlaim.upgrade_class_los" type="number" min="0" />
                </div>
              </div>

              <div v-if="formKlaim.jenis_rawat === '1' && formKlaim.icu_indikator" class="claim-sheet-row claim-intensive-row">
                <b><span class="claim-help-icon">?</span> Ventilator</b>
                <div class="claim-value claim-intensive-main">
                  <label class="claim-radio"><input v-model="formKlaim.ventilator_use_ind" type="checkbox" /> Ya</label>
                  <template v-if="formKlaim.ventilator_use_ind">
                    <span>Intubasi : <DatePicker v-model="formKlaim.ventilator_start_dttm" :show-icon="false" show-time hour-format="24" placeholder="Waktu Intubasi" /></span>
                    <span>Ekstubasi : <DatePicker v-model="formKlaim.ventilator_stop_dttm" :show-icon="false" show-time hour-format="24" placeholder="Waktu Ekstubasi" /></span>
                  </template>
                </div>
                <b><span class="claim-help-icon">?</span> Rawat Intensif <span>(hari)</span></b>
                <div class="claim-value">
                  <input v-model="formKlaim.icu_los" type="number" min="0" />
                </div>
              </div>

              <div class="claim-sheet-row">
                <b>Tanggal Rawat</b>
                <div class="claim-value claim-date-line">
                  <span>Masuk : 
                    <DatePicker class="w-date" :model-value="parseTanggal(formKlaim.tgl_masuk)" @update:model-value="val => formKlaim.tgl_masuk = formatTanggal(val)" :show-icon="false" placeholder="Tgl Masuk" />
                    <DatePicker class="w-time" :model-value="parseWaktu(formKlaim.jam_masuk)" @update:model-value="val => formKlaim.jam_masuk = formatWaktu(val)" time-only hour-format="24" :show-icon="false" placeholder="Jam" />
                  </span>
                  <span>Pulang : 
                    <DatePicker class="w-date" :model-value="parseTanggal(formKlaim.tgl_pulang)" @update:model-value="val => formKlaim.tgl_pulang = formatTanggal(val)" :show-icon="false" placeholder="Tgl Pulang" />
                    <DatePicker class="w-time" :model-value="parseWaktu(formKlaim.jam_pulang)" @update:model-value="val => formKlaim.jam_pulang = formatWaktu(val)" time-only hour-format="24" :show-icon="false" placeholder="Jam" />
                  </span>
                </div>
                <b>Umur</b>
                <div class="claim-value">{{ umurPasien() }}</div>
              </div>

              <div class="claim-sheet-row single">
                <b>Cara Masuk</b>
                <div class="claim-value claim-select-compact">
                  <select v-model="formKlaim.cara_masuk">
                    <option v-for="item in opsiCaraMasuk" :key="item.value" :value="item.value">{{ item.label }}</option>
                  </select>
                </div>
              </div>

              <div class="claim-sheet-row">
                <b>LOS</b>
                <div class="claim-value claim-los-line">
                  <span>{{ losHari() }}</span>
                  <strong>{{ losJam() }}</strong>
                </div>
                <b>Berat Lahir <span>(gram)</span></b>
                <div class="claim-value"><input v-model="formKlaim.birth_weight" type="number" min="0" /></div>
              </div>

              <div class="claim-sheet-row">
                <b>ADL Score</b>
                <div class="claim-value claim-date-line">
                  <span>Sub Acute : <input v-model="formKlaim.adl_sub_acute" type="number" min="0" /></span>
                  <span>Chronic : <input v-model="formKlaim.adl_chronic" type="number" min="0" /></span>
                </div>
                <b>Cara Pulang</b>
                <div class="claim-value">
                  <select v-model="formKlaim.discharge_status">
                    <option v-for="item in opsiCaraPulang" :key="item.value" :value="item.value">{{ item.label }}</option>
                  </select>
                </div>
              </div>

              <div class="claim-sheet-row tall">
                <b>DPJP</b>
                <div class="claim-value"><textarea v-model="formKlaim.nama_dokter" rows="5"></textarea></div>
                <b>Jenis Tarif</b>
                <div class="claim-value">
                  <select v-model="formKlaim.kode_tarif">
                    <option v-for="item in opsiKodeTarif" :key="item.value" :value="item.value">{{ item.label }}</option>
                  </select>
                </div>
              </div>

              <div class="claim-sheet-row single">
                <b>Pasien TB</b>
                <div class="claim-value claim-tb-line">
                  <label class="claim-radio"><input v-model="formKlaim.pasien_tb" type="checkbox" /> Ya</label>
                  <input v-if="formKlaim.pasien_tb" v-model="formKlaim.nomor_register_sitb" type="text" placeholder="No register SITB" />
                  <button v-if="formKlaim.pasien_tb" type="button" class="table-action-button" :disabled="loadingProses === 'validasi_sitb'" @click="jalankanProses('validasi_sitb')">
                    Validasi SITB
                  </button>
                </div>
              </div>

              <div class="claim-sheet-row">
                <b>Data Apgar</b>
                <div class="claim-value">
                  <label class="claim-radio"><input v-model="formKlaim.tampilkan_apgar" type="checkbox" /> Tambahkan Apgar</label>
                </div>
                <b>Data Persalinan</b>
                <div class="claim-value">
                  <label class="claim-radio"><input v-model="formKlaim.tampilkan_persalinan" type="checkbox" /> Tambahkan Persalinan</label>
                </div>
              </div>

            </article>

            <article class="idrg-claim-sheet idrg-tarif-sheet">
              <header class="claim-tarif-total">
                <span class="claim-help-icon">?</span>
                <em>Tarif Rumah Sakit :</em>
                <strong>Rp {{ rupiah(totalTarifRS()) }}</strong>
              </header>
              <div class="claim-tarif-grid">
                <label v-for="item in daftarTarifRS" :key="item.key" class="claim-tarif-item">
                  <span class="claim-help-icon">?</span>
                  <small>{{ item.label }}</small>
                  <input :value="rupiah(formTarifRS[item.key])" type="text" inputmode="numeric" @input="ubahTarifRS(item.key, $event.target.value)" />
                </label>
              </div>
              <label class="claim-tarif-confirm">
                <input v-model="formKlaim.tarif_dikonfirmasi" type="checkbox" />
                <span>Menyatakan benar bahwa data tarif yang tersebut di atas adalah benar sesuai dengan kondisi yang sesungguhnya.</span>
              </label>
              <div class="idrg-claim-actions">
                <span v-if="loadingNewClaimOtomatis" class="idrg-auto-claim-status">
                  <LoaderCircle class="spin" :size="14" />
                  New Claim otomatis sedang dikirim...
                </span>
                <button type="button" :disabled="loadingNewClaimOtomatis || loadingProses === 'atur_data_klaim' || !formKlaim.tarif_dikonfirmasi" @click="jalankanProses('atur_data_klaim')">
                  <LoaderCircle v-if="loadingProses === 'atur_data_klaim'" class="spin" :size="16" />
                  <ClipboardList v-else :size="16" />
                  Set Data Klaim
                </button>
              </div>
            </article>
          </section>

          <section v-else-if="tabAktif === 'idrg'" class="idrg-panel-card idrg-detail-panel" :class="{ 'idrg-final-locked': idrgSudahFinal }">
            <header class="coding-work-header">
              <div>
                <span>IDRG</span>
                <h3>Coding dan Grouping IDRG</h3>
              </div>
              <div class="coding-work-actions">
                <button type="button" :disabled="loadingCodingIdrg || idrgSudahFinal" @click="muatCodingIdrg('diagnosa')">
                  <LoaderCircle v-if="loadingCodingIdrg === 'diagnosa'" class="spin" :size="15" />
                  <RefreshCw v-else :size="15" />
                  Muat Diagnosa
                </button>
                <button type="button" :disabled="loadingCodingIdrg || idrgSudahFinal" @click="muatCodingIdrg('prosedur')">
                  <LoaderCircle v-if="loadingCodingIdrg === 'prosedur'" class="spin" :size="15" />
                  <RefreshCw v-else :size="15" />
                  Muat Prosedur
                </button>
              </div>
            </header>

            <div class="idrg-coding-grid idrg-coding-grid-table">
              <article>
                <div class="coding-header eklaim-header">
                  <h4>Diagnosa <span>(ICD-10)</span>:</h4>
                </div>
                <div class="idrg-table-container">
                  <div v-if="diagnosa.length === 0" class="idrg-table-empty">Belum ada diagnosa. Klik "Muat Diagnosa".</div>
                  <table v-else class="idrg-coding-table">
                    <thead>
                      <tr>
                        <th>Kode</th>
                        <th>Nama Diagnosa</th>
                        <th>Status</th>
                      </tr>
                    </thead>
                    <tbody>
                      <template v-for="(item, index) in diagnosa" :key="`idrg-diagnosa-${index}`">
                      <tr>
                        <td class="col-kode">
                          <button type="button" class="coding-code-button" :disabled="idrgSudahFinal" @click="bukaEditorDiagnosa(index)">{{ item.kd_diag || '-' }}</button>
                        </td>
                        <td class="col-nama">
                          <button type="button" class="idrg-btn-edit-nama" :disabled="idrgSudahFinal" @click="bukaEditorDiagnosa(index)">
                            {{ item.nm_diag || 'Klik untuk cari diagnosa' }}
                          </button>
                        </td>
                        <td class="col-status">
                          <span class="status-badge" :class="{ 'is-primer': index === 0 }">{{ index === 0 ? 'Primer' : 'Sekunder' }}</span>
                        </td>
                      </tr>
                      <tr v-if="!idrgSudahFinal && activeSubstitusiDiagnosa === index" class="coding-edit-row">
                        <td colspan="3">
                          <div class="coding-inline-editor">
                            <AutoComplete
                              v-model="substitusiKeyword"
                              :suggestions="searchHasilDiagnosa"
                              option-label="label"
                              @complete="onSearchDiagnosa"
                              @item-select="applySubstitusi(index, $event)"
                              placeholder="Substitusi"
                              class="eklaim-substitusi"
                            >
                              <template #option="slotProps">
                                <span v-if="slotProps.option"><b>{{ slotProps.option.kode }}</b> - {{ slotProps.option.deskripsi }}</span>
                              </template>
                            </AutoComplete>
                            <button type="button" class="danger" @click="hapusDiagnosa(index)">Delete</button>
                            <button type="button" :disabled="index === 0" @click="setPrimaryDiagnosa(index)">Set Primary</button>
                          </div>
                        </td>
                      </tr>
                      </template>
                    </tbody>
                  </table>
                </div>
              </article>
              <article>
                <div class="coding-header eklaim-header">
                  <h4>Prosedur <span>(ICD-9CM)</span>:</h4>
                </div>
                <div class="idrg-table-container">
                  <div v-if="prosedur.length === 0" class="idrg-table-empty">Belum ada prosedur. Klik "Muat Prosedur".</div>
                  <table v-else class="idrg-coding-table">
                    <thead>
                      <tr>
                        <th>Kode</th>
                        <th>Nama Prosedur</th>
                        <th>Status</th>
                      </tr>
                    </thead>
                    <tbody>
                      <template v-for="(item, index) in prosedur" :key="`idrg-prosedur-${index}`">
                      <tr>
                        <td class="col-kode">
                          <button type="button" class="coding-code-button" :disabled="idrgSudahFinal" @click="bukaEditorProsedur(index)">{{ item.kd_prosedur || '-' }}</button>
                        </td>
                        <td class="col-nama">
                          <button type="button" class="idrg-btn-edit-nama" :disabled="idrgSudahFinal" @click="bukaEditorProsedur(index)">
                            {{ item.nm_prosedur || 'Klik untuk cari prosedur' }}
                          </button>
                        </td>
                        <td class="col-status">
                          <span class="status-badge">{{ index === 0 ? 'Primer' : 'Sekunder' }}</span>
                        </td>
                      </tr>
                      <tr v-if="!idrgSudahFinal && activeSubstitusiProsedur === index" class="coding-edit-row">
                        <td colspan="3">
                          <div class="coding-inline-editor">
                            <AutoComplete
                              v-model="substitusiKeyword"
                              :suggestions="searchHasilProsedur"
                              option-label="label"
                              @complete="onSearchProsedur"
                              @item-select="applySubstitusiProsedur(index, $event)"
                              placeholder="Substitusi"
                              class="eklaim-substitusi"
                            >
                              <template #option="slotProps">
                                <span v-if="slotProps.option"><b>{{ slotProps.option.kode }}</b> - {{ slotProps.option.deskripsi }}</span>
                              </template>
                            </AutoComplete>
                            <button type="button" class="danger" @click="hapusProsedur(index)">Delete</button>
                          </div>
                        </td>
                      </tr>
                      </template>
                    </tbody>
                  </table>
                </div>
              </article>
            </div>

            <section class="coding-result-card" :class="{ final: idrgGroupingFinal }">
              <header>
                <h4>Hasil Grouping IDRG{{ idrgGroupingFinal ? ' - Final' : '' }}</h4>
                <div class="coding-result-actions">
                  <button v-if="idrgBisaEditUlang" type="button" class="edit-action" :disabled="!pasienTerpilih.no_sep || loadingProses === 'reedit_idrg'" @click="jalankanProses('reedit_idrg')">
                    <LoaderCircle v-if="loadingProses === 'reedit_idrg'" class="spin" :size="15" />
                    <Edit3 v-else :size="15" />
                    Edit Ulang IDRG
                  </button>
                  <template v-else>
                    <button type="button" class="grouping-action" :disabled="!pasienTerpilih.no_sep || loadingProses === 'grouping_idrg'" @click="jalankanProses('grouping_idrg')">
                      <LoaderCircle v-if="loadingProses === 'grouping_idrg'" class="spin" :size="15" />
                      <Layers v-else :size="15" />
                      Grouping
                    </button>
                    <button type="button" class="final-action" :disabled="!pasienTerpilih.no_sep || loadingProses === 'final_idrg' || !responseIdrgTerakhir" @click="jalankanProses('final_idrg')">
                      <LoaderCircle v-if="loadingProses === 'final_idrg'" class="spin" :size="15" />
                      <CheckCircle2 v-else :size="15" />
                      Final IDRG
                    </button>
                  </template>
                </div>
              </header>

              <div v-if="responseIdrgTerakhir" class="idrg-grouping-sheet" :class="{ final: idrgGroupingFinal }">
              <h4>Hasil Grouping iDRG{{ idrgGroupingFinal ? ' - Final' : '' }}</h4>
              <div class="idrg-grouping-table-wrap">
                <table>
                  <tbody>
                    <tr>
                      <th>Info</th>
                      <td colspan="3">{{ hasilGroupingIdrgTabel.info }}</td>
                    </tr>
                    <tr>
                      <th>Jenis Rawat</th>
                      <td colspan="3">{{ hasilGroupingIdrgTabel.jenisRawat }}</td>
                    </tr>
                    <tr>
                      <th>MDC</th>
                      <td>{{ hasilGroupingIdrgTabel.mdcDeskripsi }}</td>
                      <td class="text-right" colspan="2">{{ hasilGroupingIdrgTabel.mdcNomor }}</td>
                    </tr>
                    <tr>
                      <th>DRG</th>
                      <td>{{ hasilGroupingIdrgTabel.drgDeskripsi }}</td>
                      <td class="text-right">{{ hasilGroupingIdrgTabel.drgKode }}</td>
                      <td class="weight-cell"><span>DRG Cost Weight:</span><strong>** {{ hasilGroupingIdrgTabel.drgCostWeight }}</strong></td>
                    </tr>
                    <tr>
                      <th>KRIS</th>
                      <td>{{ hasilGroupingIdrgTabel.kris }}</td>
                      <td></td>
                      <td class="weight-cell"><span>KRIS Cost Weight:</span><strong>** {{ hasilGroupingIdrgTabel.krisCostWeight }}</strong></td>
                    </tr>
                    <tr>
                      <th>NBR</th>
                      <td class="blue-value">** {{ hasilGroupingIdrgTabel.nbr }}</td>
                      <td></td>
                      <td class="weight-cell"><span>Total Cost Weight:</span><strong>** {{ hasilGroupingIdrgTabel.totalCostWeight }}</strong></td>
                    </tr>
                    <tr>
                      <th>Total Klaim</th>
                      <td></td>
                      <td class="text-right">{{ hasilGroupingIdrgTabel.totalKlaim === '-' ? '' : 'Rp' }}</td>
                      <td class="blue-value strong-value">{{ hasilGroupingIdrgTabel.totalKlaim === '-' ? '-' : hasilGroupingIdrgTabel.totalKlaim.replace(/^Rp\s*/i, '') }}</td>
                    </tr>
                    <tr>
                      <th>Status</th>
                      <td colspan="3">{{ hasilGroupingIdrgTabel.status }}</td>
                    </tr>
                  </tbody>
                </table>
              </div>

              <p class="idrg-grouping-note">**) Catatan: Nilai belum final, sewaktu-waktu bisa berubah</p>

              <ul v-if="opsiTopupIdrg.length" class="idrg-grouping-topup">
                <li v-for="item in opsiTopupIdrg" :key="item.code">
                  <b>{{ item.code }}</b>
                  <span>{{ item.description }}</span>
                  <small>Rp {{ rupiah(item.tariff || 0) }}</small>
                </li>
              </ul>

              </div>
              <div v-else class="coding-result-empty">Hasil grouping IDRG belum tersedia.</div>
            </section>

          </section>

          <section v-else class="idrg-panel-card idrg-detail-panel inacbg-panel">
            <section class="inacbg-work-card">
              <header class="inacbg-work-header">
                <div>
                  <span>INA-CBG</span>
                  <h3>Coding dan Grouping INA-CBG</h3>
                </div>
                <div class="inacbg-header-actions">
                  <button type="button" class="purple" :disabled="!pasienTerpilih.no_sep || loadingImportInacbg || inacbgSudahFinal" @click="importIdrgKeInacbg">
                    <LoaderCircle v-if="loadingImportInacbg" class="spin" :size="15" />
                    <FileSpreadsheet v-else :size="15" />
                    Import IDRG ke INA-CBG
                  </button>
                  <button type="button" class="coding-load-action" :disabled="Boolean(loadingCodingInacbg) || inacbgSudahFinal" @click="muatCodingInacbg('diagnosa')">
                    <LoaderCircle v-if="loadingCodingInacbg === 'diagnosa'" class="spin" :size="15" />
                    <RefreshCw v-else :size="15" />
                    Muat Diagnosa
                  </button>
                  <button type="button" class="coding-load-action" :disabled="Boolean(loadingCodingInacbg) || inacbgSudahFinal" @click="muatCodingInacbg('prosedur')">
                    <LoaderCircle v-if="loadingCodingInacbg === 'prosedur'" class="spin" :size="15" />
                    <RefreshCw v-else :size="15" />
                    Muat Prosedur
                  </button>
                </div>
              </header>

              <div class="inacbg-code-form">
                <article>
                  <h4>Diagnosa <span>(ICD-10)</span></h4>
                  <div v-if="diagnosaInacbg.length === 0" class="inacbg-empty">Klik Import IDRG ke INA-CBG untuk menampilkan coding.</div>
                  <div v-else class="coding-list-table">
                    <div class="coding-list-head"><span>Kode</span><span>Nama Diagnosa</span><span>Status</span></div>
                    <template v-for="(item, index) in diagnosaInacbg" :key="`inacbg-diagnosa-${index}`">
                      <button type="button" class="coding-list-row" :class="{ invalid: item.validcode === '0' }" :disabled="inacbgSudahFinal" :title="inacbgSudahFinal ? 'Coding sudah final' : 'Klik untuk edit diagnosa'" @click="bukaEditorDiagnosaInacbg(index)">
                        <b>{{ item.code || '-' }}</b>
                        <span>{{ item.display || '-' }} <em v-if="item.validcode === '0'">{{ item.metadata?.message || 'Kode tidak berlaku' }}</em></span>
                        <small class="coding-status-edit" :class="{ primary: index === 0 }">
                          {{ index === 0 ? 'Primer' : 'Sekunder' }}
                          <span v-if="!inacbgSudahFinal"><Edit3 :size="10" /> Edit</span>
                        </small>
                      </button>
                      <div v-if="!inacbgSudahFinal && activeSubstitusiDiagnosaInacbg === index" class="coding-inline-editor">
                        <AutoComplete
                          v-model="substitusiDiagnosaInacbg"
                          :suggestions="searchHasilDiagnosaInacbg"
                          option-label="label"
                          placeholder="Substitusi"
                          @complete="onSearchDiagnosaInacbg"
                          @item-select="terapkanSubstitusiDiagnosaInacbg(index, $event)"
                        />
                        <button type="button" class="danger" @click="hapusDiagnosaInacbg(index)">Delete</button>
                        <button type="button" :disabled="index === 0" @click="setPrimaryDiagnosaInacbg(index)">Set Primary</button>
                      </div>
                    </template>
                  </div>
                </article>

                <article>
                  <h4>Prosedur <span>(ICD-9-CM)</span></h4>
                  <div v-if="prosedurInacbg.length === 0" class="inacbg-empty">Klik Import IDRG ke INA-CBG untuk menampilkan coding.</div>
                  <div v-else class="coding-list-table">
                    <div class="coding-list-head"><span>Kode</span><span>Nama Prosedur</span><span>Status</span></div>
                    <template v-for="(item, index) in prosedurInacbg" :key="`inacbg-prosedur-${index}`">
                      <button type="button" class="coding-list-row" :class="{ invalid: item.validcode === '0' }" :disabled="inacbgSudahFinal" :title="inacbgSudahFinal ? 'Coding sudah final' : 'Klik untuk edit prosedur'" @click="bukaEditorProsedurInacbg(index)">
                        <b>{{ item.code || '-' }}</b>
                        <span>{{ item.display || '-' }} <em v-if="item.validcode === '0'">{{ item.metadata?.message || 'Kode tidak berlaku' }}</em></span>
                        <small class="coding-status-edit">
                          {{ index === 0 ? 'Primer' : 'Sekunder' }}
                          <span v-if="!inacbgSudahFinal"><Edit3 :size="10" /> Edit</span>
                        </small>
                      </button>
                      <div v-if="!inacbgSudahFinal && activeSubstitusiProsedurInacbg === index" class="coding-inline-editor">
                        <AutoComplete
                          v-model="substitusiProsedurInacbg"
                          :suggestions="searchHasilProsedurInacbg"
                          option-label="label"
                          placeholder="Substitusi"
                          @complete="onSearchProsedurInacbg"
                          @item-select="terapkanSubstitusiProsedurInacbg(index, $event)"
                        />
                        <button type="button" class="danger" @click="hapusProsedurInacbg(index)">Delete</button>
                      </div>
                    </template>
                  </div>
                </article>
              </div>
            </section>

            <section class="inacbg-result-card" :class="{ final: inacbgGroupingFinal }">
              <header>
                <h4>Hasil Grouping INA-CBG{{ inacbgGroupingFinal ? ' - Final' : '' }}</h4>
                <div class="inacbg-process-actions">
                  <template v-if="klaimSudahFinal">
                    <button type="button" class="secondary" :disabled="!pasienTerpilih.no_sep || Boolean(loadingProses)" @click="jalankanProses('reedit_klaim')">
                      <LoaderCircle v-if="loadingProses === 'reedit_klaim'" class="spin" :size="15" />
                      <Edit3 v-else :size="15" />
                      Edit Ulang Klaim
                    </button>
                    <button type="button" class="teal" :disabled="!pasienTerpilih.no_sep || Boolean(loadingProses)" @click="jalankanProses('send_claim_individual')">
                      <LoaderCircle v-if="loadingProses === 'send_claim_individual'" class="spin" :size="15" />
                      <Send v-else :size="15" />
                      {{ klaimSudahTerkirim ? 'Kirim Ulang Klaim' : 'Kirim Klaim' }}
                    </button>
                  </template>
                  <template v-else-if="inacbgBisaEditUlang">
                    <button type="button" class="secondary" :disabled="!pasienTerpilih.no_sep || Boolean(loadingProses)" @click="jalankanProses('reedit_inacbg')">
                      <LoaderCircle v-if="loadingProses === 'reedit_inacbg'" class="spin" :size="15" />
                      <Edit3 v-else :size="15" />
                      Edit Ulang INA-CBG
                    </button>
                    <button type="button" class="teal" :disabled="!pasienTerpilih.no_sep || Boolean(loadingProses)" @click="jalankanProses('final_klaim')">
                      <LoaderCircle v-if="loadingProses === 'final_klaim'" class="spin" :size="15" />
                      <CheckCircle2 v-else :size="15" />
                      Final Klaim
                    </button>
                  </template>
                  <template v-else>
                    <button type="button" class="purple" :disabled="!pasienTerpilih.no_sep || loadingProses === 'grouping_inacbg'" @click="jalankanProses('grouping_inacbg')">
                      <LoaderCircle v-if="loadingProses === 'grouping_inacbg'" class="spin" :size="15" />
                      <Layers v-else :size="15" />
                      Grouping
                    </button>
                    <button type="button" class="teal" :disabled="!pasienTerpilih.no_sep || loadingProses === 'final_inacbg' || !responseInacbgTerakhir || inacbgGroupingGagal" :title="inacbgGroupingGagal ? 'Edit coding dan Grouping ulang terlebih dahulu' : ''" @click="jalankanProses('final_inacbg')">
                      <LoaderCircle v-if="loadingProses === 'final_inacbg'" class="spin" :size="15" />
                      <CheckCircle2 v-else :size="15" />
                      Final INA-CBG
                    </button>
                  </template>
                </div>
              </header>

              <div v-if="responseInacbgTerakhir" class="inacbg-grouping-sheet" :class="{ final: inacbgGroupingFinal }">
                <div class="inacbg-grouping-table-wrap">
                  <table>
                    <tbody>
                      <tr>
                        <th>Info</th>
                        <td colspan="4">{{ hasilGroupingInacbgTabel.info }}</td>
                      </tr>
                      <tr class="main-row">
                        <th>Group</th>
                        <td>{{ hasilGroupingInacbgTabel.deskripsi }}</td>
                        <td class="text-right">{{ hasilGroupingInacbgTabel.kode }}</td>
                        <td class="text-right">Rp</td>
                        <td class="text-right blue-value">{{ rupiah(hasilGroupingInacbgTabel.tarifUtama) }}</td>
                      </tr>
                      <tr>
                        <th>Sub Acute</th>
                        <td>-</td>
                        <td class="text-right">-</td>
                        <td class="text-right">Rp</td>
                        <td class="text-right">{{ rupiah(hasilGroupingInacbgTabel.subAcute) }}</td>
                      </tr>
                      <tr>
                        <th>Chronic</th>
                        <td>-</td>
                        <td class="text-right">-</td>
                        <td class="text-right">Rp</td>
                        <td class="text-right">{{ rupiah(hasilGroupingInacbgTabel.chronic) }}</td>
                      </tr>
                      <tr v-for="item in hasilGroupingInacbgTabel.specialRows" :key="item.type">
                        <th>{{ item.type }}</th>
                        <td>
                          <Select
                            v-if="!inacbgGroupingFinal && opsiSpecialCmgPerKategori[item.type]?.length > 1"
                            v-model="pilihanSpecialCmg[item.type]"
                            class="inacbg-special-select"
                            :options="opsiSpecialCmgPerKategori[item.type]"
                            option-label="label"
                            option-value="value"
                            placeholder="None"
                            append-to="body"
                            overlay-class="inacbg-special-cmg-overlay"
                            filter
                          />
                          <span v-else>{{ item.description }}</span>
                        </td>
                        <td class="text-right">{{ item.code }}</td>
                        <td class="text-right">Rp</td>
                        <td class="text-right">{{ rupiah(item.tariff) }}</td>
                      </tr>
                      <tr class="total-row">
                        <th colspan="3">Total Klaim</th>
                        <td class="text-right">Rp</td>
                        <td class="text-right">{{ rupiah(hasilGroupingInacbgTabel.total) }}</td>
                      </tr>
                      <tr>
                        <th>Status</th>
                        <td colspan="4">{{ hasilGroupingInacbgTabel.status }}</td>
                      </tr>
                    </tbody>
                  </table>
                </div>
                <p v-if="inacbgGroupingGagal" class="inacbg-grouping-warning">
                  Grouper tidak menemukan kelompok INA-CBG untuk kombinasi coding ini. Edit diagnosis/prosedur, kemudian lakukan Grouping ulang.
                </p>
                <p v-else-if="adaOpsiSpecialCmg && !inacbgGroupingFinal" class="inacbg-option-note">
                  Pilih Special CMG yang sesuai, lalu klik Grouping kembali agar pilihan dikirim ke INA-CBG Stage 2 dan tarif dihitung ulang.
                </p>
              </div>
              <div v-else class="inacbg-empty-result">Hasil grouping INA-CBG belum ada.</div>
            </section>
          </section>

        </div>
      </template>

      <section v-else class="idrg-card idrg-coming-soon">
        <component :is="menuSidebar.find((item) => item.id === menuSidebarAktif)?.icon || FileText" :size="34" />
        <span>{{ menuSidebar.find((item) => item.id === menuSidebarAktif)?.label }}</span>
        <h2>Tahap berikutnya</h2>
        <p>Bagian ini akan disambungkan setelah alur Pengajuan Klaim IDRG stabil dulu, pak.</p>
      </section>
    </main>
  </section>
</template>
