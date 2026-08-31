// @ts-nocheck -- migrasi TypeScript bertahap; kontrak data modul lama belum sepenuhnya bertipe.
import { computed, reactive, ref, watch } from "vue"
import { awalKeperawatanIgdData, hapusAwalKeperawatanIgd, simpanAwalKeperawatanIgd, ubahAwalKeperawatanIgd } from "../../../lib/faisal/api"
import { useNotifikasi } from "../../../lib/shared/useNotifikasi"

export function useAwalKeperawatanIgd(props) {
  const notifikasi = useNotifikasi()
  const loading = ref(false)
  const saving = ref(false)
  const deleting = ref(false)
  const confirmDelete = ref(false)
  const formOpen = ref(true)
  const problemSearch = ref('')
  const planSearch = ref('')
  const data = reactive({ petugas: {}, masalah_keperawatan: [], rencana_keperawatan: [], penilaian: null, billing_terkunci: false })
  const petugas = ref({ nip: '', nama: '', jabatan: '' })
  const form = reactive(emptyForm())
  
  const editing = computed(() => Boolean(data.penilaian))
  const billingLocked = computed(() => Boolean(data.billing_terkunci))
  const selectedProblems = computed(() => new Set(form.kode_masalah))
  const visibleProblems = computed(() => {
    const keyword = problemSearch.value.trim().toLocaleLowerCase('id')
    if (!keyword) return data.masalah_keperawatan
    return data.masalah_keperawatan.filter((item) => `${item.kode} ${item.nama}`.toLocaleLowerCase('id').includes(keyword))
  })
  const visiblePlans = computed(() => {
    if (selectedProblems.value.size === 0) return []
    const keyword = planSearch.value.trim().toLocaleLowerCase('id')
    return data.rencana_keperawatan.filter((item) => selectedProblems.value.has(item.kode_masalah) && (!keyword || `${item.kode} ${item.nama}`.toLocaleLowerCase('id').includes(keyword)))
  })
  const painScores = ['0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '10']
  
  const options = {
    informasi: list(['Autoanamnesis', 'Alloanamnesis']), kehamilan: list(['Tidak Hamil', 'Hamil']), tekanan: list(['TAK', 'Sakit Kepala', 'Muntah', 'Pusing', 'Bingung']),
    pupil: list(['Normal', 'Miosis', 'Isokor', 'Anisokor']), neuro: list(['TAK', 'Spasme Otot', 'Perubahan Sensorik', 'Perubahan Motorik', 'Perubahan Bentuk Ekstremitas', 'Penurunan Tingkat Kesadaran', 'Fraktur/Dislokasi', 'Luksasio', 'Kerusakan Jaringan/Luka']),
    integumen: list(['TAK', 'Luka Bakar', 'Luka Robek', 'Lecet', 'Luka Decubitus', 'Luka Gangren']), turgor: list(['Baik', 'Menurun']), edema: list(['Tidak Ada', 'Ekstremitas', 'Seluruh Tubuh', 'Asites', 'Palpebrae']), mukosa: list(['Lembab', 'Kering']),
    ada: list(['Tidak Ada', 'Ada']), intoksikasi: list(['Tidak Ada', 'Ada', 'Gigitan Binatang', 'Zat Kimia', 'Gas', 'Obat']), psikologis: list(['Tidak Ada Masalah', 'Marah', 'Takut', 'Depresi', 'Cepat Lelah', 'Cemas', 'Gelisah', 'Lain-lain']),
    yaTidak: list(['Tidak', 'Ya']), perilaku: list(['Perilaku Kekerasan', 'Gangguan Efek', 'Gangguan Memori', 'Halusinasi', 'Kecenderungan Percobaan Bunuh Diri', 'Lainnya']), hubungan: list(['Harmonis', 'Kurang Harmonis', 'Tidak Harmonis', 'Konflik Besar']),
    tinggal: list(['Sendiri', 'Orang Tua', 'Suami / Istri', 'Lainnya']), pendidikan: list(['-', 'TS', 'TK', 'SD', 'SMP', 'SMA', 'SLTA/SEDERAJAT', 'D1', 'D2', 'D3', 'D4', 'S1', 'S2', 'S3']), edukasi: list(['Pasien', 'Keluarga']),
    kemampuan: list(['Mandiri', 'Bantuan Minimal', 'Bantuan Sebagian', 'Ketergantungan Total']), aktifitas: list(['Tirah Baring', 'Duduk', 'Berjalan']), nyeri: list(['Tidak Ada Nyeri', 'Nyeri Akut', 'Nyeri Kronis']), provokes: list(['Proses Penyakit', 'Benturan', 'Lain-lain']),
    quality: list(['Seperti Tertusuk', 'Berdenyut', 'Teriris', 'Tertindih', 'Tertiban', 'Lain-lain']), skala: list(['0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '10']), hilang: list(['Istirahat', 'Medengar Musik', 'Minum Obat']),
    risiko: list(['Tidak beresiko (tidak ditemukan a dan b)', 'Resiko rendah (ditemukan a/b)', 'Resiko tinggi (ditemukan a dan b)']),
  }
  
  function list(items) { return items.map((value) => ({ label: value, value })) }
  function now() { const d = new Date(Date.now() - new Date().getTimezoneOffset() * 60000); return d.toISOString().slice(0, 19) }
  function inputDate(value) { return String(value || '').replace(' ', 'T').slice(0, 19) }
  function emptyForm() { return {
    no_rawat: props.patient?.no_rawat || '', tanggal: now(), informasi: 'Autoanamnesis', keluhan_utama: '', rpd: '', rpo: '', status_kehamilan: 'Tidak Hamil', gravida: '', para: '', abortus: '', hpht: '',
    tekanan: 'TAK', pupil: 'Normal', neurosensorik: 'TAK', integumen: 'TAK', turgor: 'Baik', edema: 'Tidak Ada', mukosa: 'Lembab', perdarahan: 'Tidak Ada', jumlah_perdarahan: '', warna_perdarahan: '', intoksikasi: 'Tidak Ada',
    bab: '', xbab: '', kbab: '', wbab: '', bak: '', xbak: '', wbak: '', lbak: '', psikologis: 'Tidak Ada Masalah', jiwa: 'Tidak', perilaku: 'Perilaku Kekerasan', dilaporkan: '', sebutkan: '', hubungan: 'Harmonis',
    tinggal_dengan: 'Sendiri', ket_tinggal: '', budaya: 'Tidak Ada', ket_budaya: '', pendidikan_pj: '-', ket_pendidikan_pj: '', edukasi: 'Pasien', ket_edukasi: '', kemampuan: 'Mandiri', aktifitas: 'Tirah Baring', alat_bantu: 'Tidak', ket_bantu: '',
    nyeri: 'Tidak Ada Nyeri', provokes: 'Proses Penyakit', ket_provokes: '', quality: 'Seperti Tertusuk', ket_quality: '', lokasi: '', menyebar: 'Tidak', skala_nyeri: '0', durasi: '', nyeri_hilang: 'Istirahat', ket_nyeri: '', pada_dokter: 'Tidak', ket_dokter: '',
    berjalan_a: 'Tidak', berjalan_b: 'Tidak', berjalan_c: 'Tidak', hasil: 'Tidak beresiko (tidak ditemukan a dan b)', lapor: 'Tidak', ket_lapor: '', rencana: '', nip: '', kode_masalah: [], kode_rencana: [],
  } }
  
  function fillForm() {
    const record = data.penilaian
    Object.assign(form, emptyForm(), record || {}, { no_rawat: props.patient.no_rawat, tanggal: inputDate(record?.tanggal) || now(), kode_masalah: [...(record?.kode_masalah || [])], kode_rencana: [...(record?.kode_rencana || [])] })
    petugas.value = record ? { nip: record.nip, nama: record.nama_petugas, jabatan: record.jabatan_petugas } : { ...(data.petugas || {}) }
  }
  
  async function loadData() {
    if (!props.token || !props.patient.no_rawat) return
    loading.value = true
    try { const response = await awalKeperawatanIgdData(props.token, props.patient.no_rawat); Object.assign(data, response || {}); fillForm(); formOpen.value = true }
    catch (error) { notifikasi.gagal(error.message || 'Penilaian awal keperawatan IGD tidak dapat dibaca.') }
    finally { loading.value = false }
  }
  
  function payload() { return { ...form, tanggal: form.tanggal.replace('T', ' '), nip: petugas.value?.nip || '', kode_masalah: [...form.kode_masalah], kode_rencana: [...form.kode_rencana] } }
  async function save() {
    if (billingLocked.value) { notifikasi.peringatan('Kunjungan sudah masuk billing. Penilaian hanya dapat dilihat.'); return }
    if (!petugas.value?.nip) { notifikasi.peringatan('Pilih petugas yang melakukan pengkajian.'); return }
    saving.value = true
    try { const response = editing.value ? await ubahAwalKeperawatanIgd(props.token, payload()) : await simpanAwalKeperawatanIgd(props.token, payload()); notifikasi.sukses(response?.pesan || 'Penilaian berhasil disimpan.'); await loadData() }
    catch (error) { notifikasi.gagal(error.message || 'Penilaian gagal disimpan.') }
    finally { saving.value = false }
  }
  async function remove() {
    if (billingLocked.value) { confirmDelete.value = false; notifikasi.peringatan('Kunjungan sudah masuk billing. Penilaian tidak dapat dihapus.'); return }
    deleting.value = true
    try { const response = await hapusAwalKeperawatanIgd(props.token, props.patient.no_rawat); confirmDelete.value = false; notifikasi.sukses(response?.pesan || 'Penilaian berhasil dihapus.'); await loadData() }
    catch (error) { notifikasi.gagal(error.message || 'Penilaian gagal dihapus.') }
    finally { deleting.value = false }
  }
  function toggle(array, value) { const index = array.indexOf(value); if (index >= 0) array.splice(index, 1); else array.push(value) }
  function toggleProblem(kode) {
    const index = form.kode_masalah.indexOf(kode)
    if (index >= 0) {
      form.kode_masalah.splice(index, 1)
      const related = new Set(data.rencana_keperawatan.filter((item) => item.kode_masalah === kode).map((item) => item.kode))
      form.kode_rencana = form.kode_rencana.filter((item) => !related.has(item))
      return
    }
    form.kode_masalah.push(kode)
  }
  watch(() => props.patient.no_rawat, loadData, { immediate: true })
  return {
    loading,
    saving,
    deleting,
    confirmDelete,
    formOpen,
    problemSearch,
    planSearch,
    petugas,
    form,
    editing,
    billingLocked,
    selectedProblems,
    visibleProblems,
    visiblePlans,
    painScores,
    options,
    fillForm,
    save,
    remove,
    toggle,
    toggleProblem,
  }
}
