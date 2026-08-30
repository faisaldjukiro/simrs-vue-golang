import { computed, onMounted, reactive, ref } from "vue"
import * as LucideIcons from "@lucide/vue"
import { hapusSidebarPasien, kelolaMenuData, tambahSidebarPasien, ubahSidebarPasien } from "../../../lib/faisal/api"
import { useNotifikasi } from "../../../lib/shared/useNotifikasi"

export function useKelolaMenu(props, emit) {
  const { Edit3, LayoutPanelLeft, LoaderCircle, Plus, Search: SearchIcon, Trash2 } = LucideIcons
  
  const daftarIkon = Object.keys(LucideIcons)
    .filter((key) => key !== 'default' && key !== 'createLucideIcon' && /^[A-Z]/.test(key) && !key.endsWith('Icon'))
    .map((key) => ({
      label: key,
      value: key,
      component: LucideIcons[key]
    }))
  
  
  
  const notifikasi = useNotifikasi()
  
  const data = ref({ sidebar: [], pilihan_modul: [] })
  const loading = ref(true)
  const error = ref('')
  const pencarian = ref('')
  const modalForm = ref(false)
  const modalHapus = ref(false)
  const mode = ref('tambah')
  const sidebarDipilih = ref(null)
  const menyimpan = ref(false)
  const form = reactive({ kode: '', nama: '', ikon: 'LayoutDashboard', daftar_modul: [], urutan: 1, aktif: true })
  
  const daftarSidebar = computed(() => {
    const kata = pencarian.value.trim().toLowerCase()
    if (!kata) return data.value.sidebar
    return data.value.sidebar.filter((item) => [item.kode, item.nama, item.ikon, ...(item.daftar_modul || [])].join(' ').toLowerCase().includes(kata))
  })
  const judulModal = computed(() => mode.value === 'tambah' ? 'Tambah Sidebar Pasien' : 'Edit Sidebar Pasien')
  
  function resetForm() {
    Object.assign(form, {
      kode: '', nama: '', ikon: 'LayoutDashboard', daftar_modul: [],
      urutan: Math.max(0, ...data.value.sidebar.map((item) => Number(item.urutan) || 0)) + 1,
      aktif: true,
    })
  }
  
  function bukaTambah() {
    mode.value = 'tambah'
    sidebarDipilih.value = null
    resetForm()
    modalForm.value = true
  }
  
  function bukaEdit(item) {
    mode.value = 'edit'
    sidebarDipilih.value = item
    Object.assign(form, {
      kode: item.kode,
      nama: item.nama,
      ikon: item.ikon,
      daftar_modul: [...(item.daftar_modul || [])],
      urutan: Number(item.urutan) || 0,
      aktif: Boolean(item.aktif),
    })
    modalForm.value = true
  }
  
  function bukaHapus(item) {
    sidebarDipilih.value = item
    modalHapus.value = true
  }
  
  function toggleModul(modul) {
    form.daftar_modul = form.daftar_modul.includes(modul)
      ? form.daftar_modul.filter((item) => item !== modul)
      : [...form.daftar_modul, modul]
  }
  
  function buatKodeDariNama() {
    if (mode.value !== 'tambah' || form.kode) return
    form.kode = form.nama.toLowerCase().trim().replace(/[^a-z0-9]+/g, '_').replace(/^_+|_+$/g, '')
  }
  
  async function muatData() {
    loading.value = true
    error.value = ''
    try {
      data.value = await kelolaMenuData(props.token)
    } catch (err) {
      error.value = err.message || 'Konfigurasi sidebar tidak dapat dibaca.'
    } finally {
      loading.value = false
    }
  }
  
  async function simpan() {
    if (menyimpan.value) return
    if (!form.kode || !form.nama || !form.ikon || form.daftar_modul.length === 0) {
      notifikasi.peringatan('Kode, nama, ikon, dan minimal satu modul wajib diisi.')
      return
    }
    menyimpan.value = true
    try {
      const payload = { ...form, urutan: Number(form.urutan) || 0 }
      if (mode.value === 'tambah') await tambahSidebarPasien(props.token, payload)
      else await ubahSidebarPasien(props.token, sidebarDipilih.value.id, payload)
      notifikasi.sukses(mode.value === 'tambah' ? 'Sidebar berhasil ditambahkan.' : 'Sidebar berhasil diperbarui.')
      modalForm.value = false
      await muatData()
      emit('saved')
    } catch (err) {
      notifikasi.gagal(err.message || 'Sidebar tidak dapat disimpan.')
    } finally {
      menyimpan.value = false
    }
  }
  
  async function hapus() {
    if (!sidebarDipilih.value || menyimpan.value) return
    menyimpan.value = true
    try {
      await hapusSidebarPasien(props.token, sidebarDipilih.value.id)
      notifikasi.sukses('Sidebar berhasil dihapus.')
      modalHapus.value = false
      await muatData()
      emit('saved')
    } catch (err) {
      notifikasi.gagal(err.message || 'Sidebar tidak dapat dihapus.')
    } finally {
      menyimpan.value = false
    }
  }
  
  onMounted(muatData)
  return {
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
  }
}
