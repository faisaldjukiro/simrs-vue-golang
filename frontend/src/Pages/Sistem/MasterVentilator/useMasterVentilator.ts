import { computed, onMounted, reactive, ref } from "vue"
import { hapusMasterVentilator, masterVentilatorData, simpanMasterVentilator, ubahMasterVentilator } from "../../../lib/faisal/api"
import { useNotifikasi } from "../../../lib/shared/useNotifikasi"

export function useMasterVentilator(props) {
  const notifikasi = useNotifikasi()
  const records = ref([]), loading = ref(false), saving = ref(false), keyword = ref(''), editing = ref(''), formTerbuka = ref(false)
  const statuses = ['Tersedia', 'Digunakan', 'Pemeliharaan', 'Rusak', 'Nonaktif'].map((value) => ({ label: value, value }))
  const empty = () => ({ kode_ventilator: '', nama: '', merk: '', model: '', nomor_seri: '', ruangan: '', status: 'Tersedia', tanggal_maintenance_terakhir: '', tanggal_maintenance_berikutnya: '', keterangan: '' })
  const form = reactive(empty())
  const rows = computed(() => { const q = keyword.value.toLowerCase().trim(); return q ? records.value.filter((item) => Object.values(item).join(' ').toLowerCase().includes(q)) : records.value })
  const jumlahTersedia = computed(() => records.value.filter((item) => item.status === 'Tersedia').length)
  const jumlahDigunakan = computed(() => records.value.filter((item) => item.status === 'Digunakan').length)
  const jumlahPerhatian = computed(() => records.value.filter((item) => ['Pemeliharaan', 'Rusak', 'Nonaktif'].includes(item.status)).length)
  
  function statusClass(status) { if (status === 'Tersedia') return 'success'; if (status === 'Digunakan') return 'active'; if (['Rusak', 'Nonaktif'].includes(status)) return 'danger'; return 'pending' }
  function reset() { Object.assign(form, empty()); editing.value = ''; formTerbuka.value = false }
  function tambah() { Object.assign(form, empty()); editing.value = ''; formTerbuka.value = true }
  function edit(item) {
    editing.value = item.kode_ventilator
    Object.assign(form, empty(), item, { tanggal_maintenance_terakhir: item.tanggal_maintenance_terakhir || '', tanggal_maintenance_berikutnya: item.tanggal_maintenance_berikutnya || '' })
    formTerbuka.value = true
    window.scrollTo({ top: 0, behavior: 'smooth' })
  }
  async function load() { loading.value = true; try { records.value = await masterVentilatorData(props.token) || [] } catch (error) { notifikasi.gagal(error.message) } finally { loading.value = false } }
  async function save() {
    if (saving.value) return
    if (!form.nama.trim()) return notifikasi.peringatan('Nama ventilator wajib diisi.')
    saving.value = true
    try {
      const hasil = editing.value ? await ubahMasterVentilator(props.token, editing.value, form) : await simpanMasterVentilator(props.token, form)
      notifikasi.sukses(hasil.pesan); reset(); await load()
    } catch (error) { notifikasi.gagal(error.message) } finally { saving.value = false }
  }
  async function remove(item) {
    if (!window.confirm(`Hapus ventilator ${item.nama}?`)) return
    try { const hasil = await hapusMasterVentilator(props.token, item.kode_ventilator); notifikasi.sukses(hasil.pesan); await load() } catch (error) { notifikasi.gagal(error.message) }
  }
  onMounted(load)
  return {
    records,
    loading,
    saving,
    keyword,
    editing,
    formTerbuka,
    statuses,
    empty,
    form,
    rows,
    jumlahTersedia,
    jumlahDigunakan,
    jumlahPerhatian,
    statusClass,
    reset,
    tambah,
    edit,
    load,
    save,
    remove,
  }
}
