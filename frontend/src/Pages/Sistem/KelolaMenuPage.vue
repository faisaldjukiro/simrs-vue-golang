<script setup>
import Column from "primevue/column"
import Dialog from "primevue/dialog"
import Select from "primevue/select"
import DataTable from "../../Components/Ui/DataTable.vue"
import { useKelolaMenu } from "./KelolaMenu/useKelolaMenu.js"

const props = defineProps({ token: { type: String, required: true } })
const emit = defineEmits(['saved'])

const {
  Edit3,
  LayoutPanelLeft,
  LoaderCircle,
  Plus,
  SearchIcon,
  Trash2,
  daftarIkon,
  data,
  loading,
  error,
  pencarian,
  modalForm,
  modalHapus,
  sidebarDipilih,
  menyimpan,
  form,
  daftarSidebar,
  judulModal,
  bukaTambah,
  bukaEdit,
  bukaHapus,
  toggleModul,
  buatKodeDariNama,
  simpan,
  hapus,
} = useKelolaMenu(props, emit)
</script>

<template>
  <section class="menu-management-module">
    <header class="user-management-header">
      <div>
        <span>Konfigurasi Aplikasi</span>
        <h1>Kelola Menu</h1>
        <p>Atur sidebar yang tampil pada ruang kerja pasien. Akses mengikuti modul besar: IGD/UGD, Rawat Jalan, dan Rawat Inap.</p>
      </div>
      <button type="button" @click="bukaTambah"><Plus :size="17" /> Tambah Sidebar</button>
    </header>

    <div v-if="error" class="patient-error">{{ error }}</div>
    <template v-else>
      <div class="menu-management-summary">
        <LayoutPanelLeft :size="22" />
        <span><strong>{{ data.sidebar.length }}</strong> sidebar terdaftar</span>
        <small>Perubahan tersimpan di database lokal SIRAPI.</small>
      </div>

      <div class="user-management-toolbar">
        <label><SearchIcon :size="17" /><input v-model="pencarian" type="search" placeholder="Cari kode, nama, ikon, atau modul..." /></label>
      </div>

      <DataTable
        :rows="daftarSidebar"
        data-key="id"
        :loading="loading"
        loading-message="Memuat konfigurasi sidebar..."
        empty-message="Sidebar tidak ditemukan."
        paginator
        :rows-per-page="10"
        :rows-per-page-options="[10, 25, 50]"
      >
        <Column header="Urutan">
          <template #body="{ data: item }"><strong>{{ item.urutan }}</strong></template>
        </Column>
        <Column header="Sidebar">
          <template #body="{ data: item }"><strong>{{ item.nama }}</strong><span>{{ item.kode }} · {{ item.ikon }}</span></template>
        </Column>
        <Column header="Modul">
          <template #body="{ data: item }"><div class="sidebar-module-list"><span v-for="modul in item.daftar_modul" :key="modul">{{ modul }}</span></div></template>
        </Column>
        <Column header="Status">
          <template #body="{ data: item }"><span class="patient-status" :class="{ inactive: !item.aktif }">{{ item.aktif ? 'Aktif' : 'Nonaktif' }}</span></template>
        </Column>
        <Column header="Aksi">
          <template #body="{ data: item }">
            <div class="sidebar-actions">
              <button type="button" title="Edit sidebar" @click="bukaEdit(item)"><Edit3 :size="15" /> Edit</button>
              <button type="button" class="danger" title="Hapus sidebar" @click="bukaHapus(item)"><Trash2 :size="15" /> Hapus</button>
            </div>
          </template>
        </Column>
      </DataTable>
    </template>

    <Dialog v-model:visible="modalForm" modal :header="judulModal" class="user-dialog sidebar-dialog" :style="{ width: 'min(760px, calc(100vw - 32px))' }">
      <form class="user-form" @submit.prevent="simpan">
        <div class="user-form-grid">
          <label><span>Nama Sidebar</span><input v-model="form.nama" maxlength="80" required @blur="buatKodeDariNama" /></label>
          <label><span>Kode Sidebar</span><input v-model="form.kode" maxlength="80" placeholder="contoh: cppt_soap" required /><small>Kode harus sama dengan halaman yang didaftarkan oleh programmer frontend.</small></label>
          <label class="sidebar-icon-field">
            <span>Ikon Sidebar</span>
            <Select v-model="form.ikon" :options="daftarIkon" option-label="label" option-value="value" append-to="body" fluid placeholder="Pilih Ikon" filter :virtualScrollerOptions="{ itemSize: 34 }">
              <template #value="slotProps">
                <div v-if="slotProps.value" style="display: flex; align-items: center; gap: 8px;">
                  <component :is="daftarIkon.find(i => i.value === slotProps.value)?.component" :size="16" />
                  <span>{{ slotProps.value }}</span>
                </div>
                <span v-else>Pilih Ikon</span>
              </template>
              <template #option="slotProps">
                <div style="display: flex; align-items: center; gap: 8px;">
                  <component :is="slotProps.option.component" :size="16" />
                  <span>{{ slotProps.option.label }}</span>
                </div>
              </template>
            </Select>
          </label>
          <label><span>Urutan</span><input v-model.number="form.urutan" type="number" min="0" max="65535" required /></label>
        </div>

        <section class="sidebar-option-section">
          <header><strong>Tampil Pada Modul</strong><span>{{ form.daftar_modul.length }} dipilih</span></header>
          <div class="sidebar-module-picker">
            <button v-for="modul in data.pilihan_modul" :key="modul" type="button" :class="{ selected: form.daftar_modul.includes(modul) }" @click="toggleModul(modul)">
              <i>{{ form.daftar_modul.includes(modul) ? '✓' : '' }}</i>{{ modul }}
            </button>
          </div>
        </section>

        <label class="user-active-check"><input v-model="form.aktif" type="checkbox" /><span>Sidebar aktif dan ditampilkan sesuai modul yang dipilih</span></label>

        <footer class="user-form-actions">
          <button type="button" class="secondary" :disabled="menyimpan" @click="modalForm = false">Batal</button>
          <button type="submit" :disabled="menyimpan"><LoaderCircle v-if="menyimpan" class="spin" :size="17" /><Plus v-else :size="17" />{{ menyimpan ? 'Menyimpan...' : 'Simpan Sidebar' }}</button>
        </footer>
      </form>
    </Dialog>

    <Dialog v-model:visible="modalHapus" modal header="Hapus Sidebar" class="user-dialog" :style="{ width: 'min(430px, calc(100vw - 32px))' }">
      <div class="sidebar-delete-confirm">
        <p>Sidebar <strong>{{ sidebarDipilih?.nama }}</strong> akan dihapus dari konfigurasi aplikasi.</p>
        <small>Data medis pasien tidak ikut dihapus.</small>
        <footer class="user-form-actions">
          <button type="button" class="secondary" :disabled="menyimpan" @click="modalHapus = false">Batal</button>
          <button type="button" class="danger" :disabled="menyimpan" @click="hapus"><Trash2 :size="16" />{{ menyimpan ? 'Menghapus...' : 'Hapus Sidebar' }}</button>
        </footer>
      </div>
    </Dialog>
  </section>
</template>
