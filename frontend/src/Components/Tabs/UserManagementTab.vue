<script setup>
import { computed, reactive, ref, watch } from 'vue'
import {
  Check,
  LoaderCircle,
  Plus,
  Search,
  ShieldCheck,
  UserRound,
  UsersRound,
} from '@lucide/vue'
import DataTable from '../Ui/DataTable.vue'
import Column from 'primevue/column'
import Dialog from 'primevue/dialog'
import { cariPegawaiUserManagement, tambahUserManagement, ubahAksesUserManagement } from '../../lib/faisal/api'
import { useNotifikasi } from '../../lib/shared/useNotifikasi'

const props = defineProps({
  data: { type: Object, default: () => ({ ringkasan: {}, pengguna: [], permission: [] }) },
  token: { type: String, default: '' },
  loading: Boolean,
  error: { type: String, default: '' },
})
const emit = defineEmits(['saved'])

const pencarian = ref('')
const modalUserTerbuka = ref(false)
const modeModal = ref('tambah')
const userDipilih = ref(null)
const sedangMenyimpan = ref(false)
const kataKunciPegawai = ref('')
const daftarPegawai = ref([])
const pegawaiDipilih = ref(null)
const sedangCariPegawai = ref(false)
const notifikasi = useNotifikasi()
const form = reactive({
  username: '',
  nama: '',
  email: '',
  aktif: true,
  all_access: false,
  permission: [],
})
let timerCariPegawai = 0
const pengguna = computed(() => props.data?.pengguna ?? [])
const permission = computed(() => props.data?.permission ?? [])
const ringkasan = computed(() => props.data?.ringkasan ?? {})
const modeTambah = computed(() => modeModal.value === 'tambah')
const judulModal = computed(() => modeTambah.value ? 'Tambah User Baru' : 'Atur Akses User')
const penggunaTampil = computed(() => {
  const kataKunci = pencarian.value.trim().toLowerCase()
  if (!kataKunci) return pengguna.value

  return pengguna.value.filter((user) => [
    user.username,
    user.nama,
    user.email,
    ...(user.permission ?? []),
  ].join(' ').toLowerCase().includes(kataKunci))
})

const labelPermission = computed(() => {
  const labels = {}
  permission.value.forEach((item) => {
    labels[item.kode] = item.nama
  })
  labels['*'] = 'Semua Akses'
  return labels
})

function namaPermission(kode) {
  return labelPermission.value[kode] || kode
}

function userAdmin(user) {
  return (user.permission ?? []).includes('*')
}

function labelJumlahPermission(user) {
  return userAdmin(user) ? 'Admin penuh' : `${user.permission?.length || 0} permission`
}

function bukaTambahUser() {
  resetForm()
  modeModal.value = 'tambah'
  userDipilih.value = null
  modalUserTerbuka.value = true
}

function bukaAturAkses(user) {
  resetForm()
  modeModal.value = 'akses'
  userDipilih.value = user
  form.aktif = Boolean(user.aktif)
  form.all_access = userAdmin(user)
  form.permission = [...(user.permission ?? [])]
  modalUserTerbuka.value = true
}

function resetForm() {
  form.username = ''
  form.nama = ''
  form.email = ''
  form.aktif = true
  form.all_access = false
  form.permission = []
  kataKunciPegawai.value = ''
  daftarPegawai.value = []
  pegawaiDipilih.value = null
}

function pilihPegawai(pegawai) {
  pegawaiDipilih.value = pegawai
  form.username = pegawai.nik
  form.nama = pegawai.nama
  form.email = `${pegawai.nik}@simrs.local`
  kataKunciPegawai.value = `${pegawai.nik} - ${pegawai.nama}`
  daftarPegawai.value = []
}

function togglePermission(kode) {
  if (kode === '*') {
    form.all_access = !form.all_access
    form.permission = form.all_access ? ['*'] : []
    return
  }
  if (form.all_access) return

  const tanpaSemuaAkses = form.permission.filter((item) => item !== '*')
  form.permission = tanpaSemuaAkses.includes(kode)
    ? tanpaSemuaAkses.filter((item) => item !== kode)
    : [...tanpaSemuaAkses, kode]
}

async function simpanUser() {
  if (sedangMenyimpan.value) return
  if (!props.token) {
    notifikasi.peringatan('Silakan login dulu.')
    return
  }
  if (!form.all_access && form.permission.length === 0) {
    notifikasi.peringatan('Pilih minimal satu akses untuk user.')
    return
  }
  if (modeTambah.value && !form.username) {
    notifikasi.peringatan('Pilih pegawai SIMRS dulu.')
    return
  }

  sedangMenyimpan.value = true
  try {
    if (modeTambah.value) {
      await tambahUserManagement(props.token, {
        username: form.username,
        nama: form.nama,
        email: form.email,
        aktif: form.aktif,
        all_access: form.all_access,
        permission: form.all_access ? ['*'] : form.permission,
      })
      notifikasi.sukses('User baru berhasil ditambahkan.')
    } else {
      await ubahAksesUserManagement(props.token, userDipilih.value.id, {
        aktif: form.aktif,
        all_access: form.all_access,
        permission: form.all_access ? ['*'] : form.permission,
      })
      notifikasi.sukses('Akses user berhasil diperbarui.')
    }
    modalUserTerbuka.value = false
    emit('saved')
  } catch (error) {
    notifikasi.gagal(error.message || 'Data user tidak dapat disimpan.')
  } finally {
    sedangMenyimpan.value = false
  }
}

watch(kataKunciPegawai, (nilai) => {
  window.clearTimeout(timerCariPegawai)
  const kataKunci = nilai.trim()
  if (!modeTambah.value || pegawaiDipilih.value || kataKunci.length < 2) {
    daftarPegawai.value = []
    return
  }

  timerCariPegawai = window.setTimeout(async () => {
    sedangCariPegawai.value = true
    try {
      daftarPegawai.value = await cariPegawaiUserManagement(props.token, kataKunci)
    } catch (error) {
      daftarPegawai.value = []
      notifikasi.gagal(error.message || 'Data pegawai SIMRS tidak dapat dibaca.')
    } finally {
      sedangCariPegawai.value = false
    }
  }, 350)
})
</script>

<template>
  <section class="user-management-module">
    <header class="user-management-header">
      <div>
        <span>Manajemen Akses</span>
        <h1>User Management</h1>
        <p>Kelola dan pantau user aplikasi SIRAVA.</p>
      </div>
    </header>

    <div v-if="error" class="patient-error">{{ error }}</div>

    <div v-else class="user-management-content">
      <div class="user-stats">
        <article>
          <UsersRound :size="22" />
          <span>Total User</span>
          <strong>{{ ringkasan.jumlah_pengguna ?? 0 }}</strong>
        </article>
        <article>
          <UserRound :size="22" />
          <span>User Aktif</span>
          <strong>{{ ringkasan.jumlah_aktif ?? 0 }}</strong>
        </article>
        <article>
          <ShieldCheck :size="22" />
          <span>Admin</span>
          <strong>{{ ringkasan.jumlah_admin ?? 0 }}</strong>
        </article>
      </div>

      <div class="user-management-toolbar">
        <label>
          <Search :size="17" />
          <input v-model="pencarian" type="search" placeholder="Cari username, nama, email, permission..." />
        </label>
        <button type="button" @click="bukaTambahUser">
          <Plus :size="17" />
          Tambah User
        </button>
      </div>

      <div v-if="loading" class="patient-loading-panel">
        <LoaderCircle class="spin" :size="28" />
        <strong>Sedang menarik data user...</strong>
        <span>Daftar user akan tampil setelah data berhasil dibaca.</span>
      </div>

      <DataTable
        v-else
        :rows="penggunaTampil"
        data-key="id"
        empty-message="User tidak ditemukan."
      >
        <Column header="User">
          <template #body="{ data: user }">
            <strong>{{ user.nama || user.username }}</strong>
            <span>{{ user.username }} - {{ user.email }}</span>
          </template>
        </Column>

        <Column header="Status">
          <template #body="{ data: user }">
            <span class="patient-status">{{ user.aktif ? 'Aktif' : 'Nonaktif' }}</span>
            <span>{{ userAdmin(user) ? 'Admin' : 'Petugas' }}</span>
          </template>
        </Column>

        <Column header="Permission">
          <template #body="{ data: user }">
            <strong>{{ labelJumlahPermission(user) }}</strong>
            <span>{{ (user.permission || []).map(namaPermission).join(', ') || '-' }}</span>
          </template>
        </Column>

        <Column header="Dibuat">
          <template #body="{ data: user }">
            <strong><code>{{ user.dibuat_pada || '-' }}</code></strong>
          </template>
        </Column>

        <Column header="Aksi">
          <template #body="{ data: user }">
            <button type="button" class="table-action-button" @click="bukaAturAkses(user)">
              Atur Akses
            </button>
          </template>
        </Column>
      </DataTable>
    </div>

    <Dialog
      v-model:visible="modalUserTerbuka"
      modal
      :header="judulModal"
      class="user-dialog"
      :style="{ width: 'min(760px, calc(100vw - 32px))' }"
    >
      <form class="user-form" @submit.prevent="simpanUser">
        <div v-if="modeTambah" class="employee-picker">
          <label>
            <span>Cari Pegawai SIMRS</span>
            <div>
              <Search :size="17" />
              <input
                v-model="kataKunciPegawai"
                type="search"
                placeholder="Ketik NIK, nama, atau jabatan pegawai..."
                required
                @input="pegawaiDipilih = null; form.username = ''; form.nama = ''; form.email = ''"
              />
            </div>
          </label>

          <p class="employee-hint">
            User SIRAVA mengikuti akun SIMRS lama. Password login tetap memakai password dari tabel user SIMRS.
          </p>

          <div v-if="pegawaiDipilih" class="selected-user-card">
            <UserRound :size="22" />
            <span>
              <strong>{{ pegawaiDipilih.nama }}</strong>
              <small>{{ pegawaiDipilih.nik }} - {{ pegawaiDipilih.jabatan || '-' }}</small>
            </span>
          </div>

          <div v-else class="employee-results">
            <span v-if="sedangCariPegawai">Mencari pegawai...</span>
            <span v-else-if="kataKunciPegawai.trim().length > 1 && daftarPegawai.length === 0">Pegawai tidak ditemukan.</span>
            <span v-else>Ketik minimal 2 huruf/NIK untuk mencari pegawai.</span>

            <button
              v-for="pegawai in daftarPegawai"
              :key="pegawai.nik"
              type="button"
              @click="pilihPegawai(pegawai)"
            >
              <UserRound :size="18" />
              <span>
                <strong>{{ pegawai.nama }}</strong>
                <small>{{ pegawai.nik }} - {{ pegawai.jabatan || '-' }}</small>
              </span>
            </button>
          </div>
        </div>

        <div v-else class="selected-user-card">
          <UserRound :size="22" />
          <span>
            <strong>{{ userDipilih?.nama || userDipilih?.username }}</strong>
            <small>{{ userDipilih?.username }} - {{ userDipilih?.email }}</small>
          </span>
        </div>

        <label class="user-active-check">
          <input v-model="form.aktif" type="checkbox" />
          <span>User aktif dan bisa login</span>
        </label>

        <label class="user-active-check admin-check">
          <input
            v-model="form.all_access"
            type="checkbox"
            @change="form.permission = form.all_access ? ['*'] : []"
          />
          <span>
            <strong>Jadikan Admin / Akses Penuh</strong>
            <small>Bisa membuka seluruh menu termasuk User Management.</small>
          </span>
        </label>

        <section class="permission-picker">
          <header>
            <strong>Pilih Akses</strong>
            <span>{{ form.all_access ? 'Semua akses' : `${form.permission.length} permission dipilih` }}</span>
          </header>

          <p v-if="form.all_access" class="employee-hint">
            Admin sudah mendapat semua akses. Matikan pilihan admin jika ingin memilih menu satu-satu.
          </p>

          <div class="permission-list">
            <button
              v-for="item in permission"
              :key="item.kode"
              type="button"
              :class="{ selected: form.permission.includes(item.kode), disabled: form.all_access && item.kode !== '*' }"
              @click="togglePermission(item.kode)"
            >
              <i><Check v-if="form.permission.includes(item.kode)" :size="14" /></i>
              <span>
                <strong>{{ item.kode === '*' ? 'Semua Akses' : item.nama }}</strong>
                <small>{{ item.grup }} - {{ item.kode }}</small>
              </span>
            </button>
          </div>
        </section>

        <footer class="user-form-actions">
          <button type="button" class="secondary" :disabled="sedangMenyimpan" @click="modalUserTerbuka = false">Batal</button>
          <button type="submit" :disabled="sedangMenyimpan || (modeTambah && !form.username)">
            <LoaderCircle v-if="sedangMenyimpan" class="spin" :size="17" />
            <Plus v-else :size="17" />
            {{ sedangMenyimpan ? 'Menyimpan...' : modeTambah ? 'Simpan User' : 'Simpan Akses' }}
          </button>
        </footer>
      </form>
    </Dialog>
  </section>
</template>
