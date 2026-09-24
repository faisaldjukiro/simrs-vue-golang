import { computed, nextTick, reactive, ref, watch } from 'vue'
import {
  detailTindakanLaboratorium,
  hapusPermintaanLaboratorium,
  permintaanLaboratoriumData,
  simpanPermintaanLaboratorium,
  ubahPermintaanLaboratorium,
} from '../../../lib/faisal/api'
import { useNotifikasi } from '../../../lib/shared/useNotifikasi'
import { kategoriLaboratorium } from '../../../types/permintaanLaboratorium'
import type {
  DetailLaboratorium,
  KategoriLaboratorium,
  PemeriksaanLaboratorium,
  PermintaanLaboratorium,
  PropsLaboratorium,
  SpesimenLaboratorium,
} from '../../../types/permintaanLaboratorium'
import { cetakPermintaanLaboratorium } from './cetakPermintaanLaboratorium'

export function usePermintaanLaboratorium(props: PropsLaboratorium) {
  const notifikasi = useNotifikasi()
  const kategori = ref<KategoriLaboratorium>('PK')
  const loading = ref(false)
  const saving = ref(false)
  const editing = ref(false)
  const detailLoading = ref(false)
  const deleting = ref(false)
  const error = ref('')
  const formVisible = ref(true)
  const confirmVisible = ref(false)
  const requests = ref<PermintaanLaboratorium[]>([])
  const doctor = ref<{ kode?: string; nama?: string }>({})
  const defaultDoctor = ref<{ kode?: string; nama?: string }>({})
  const treatments = ref<PemeriksaanLaboratorium[]>([])
  const billingLocked = ref(true)
  const scope = ref({ status: '', kodeCaraBayar: '', kelas: '' })
  const deleteTarget = ref<PermintaanLaboratorium | null>(null)
  const editingNumber = ref('')
  const kataKunciPermintaan = ref('')
  const form = reactive(emptyForm())
  let versiKonteks = 0
  let versiMuat = 0
  const busy = computed(() => loading.value || saving.value || deleting.value || editing.value)
  const namaKategori = computed(() => kategoriLaboratorium.find((item) => item.value === kategori.value)?.label || kategori.value)
  const tableRowsTampil = computed(() => {
    const keyword = kataKunciPermintaan.value.trim().toLowerCase()
    return requests.value
      .filter((item) => !keyword || teksPermintaan(item).includes(keyword))
      .map((item) => ({ ...item, _key: `${item.kategori}:${item.nomor}` }))
  })

  function pesanError(err: unknown) {
    return err instanceof Error ? err.message : 'Permintaan laboratorium tidak dapat diproses.'
  }

  function waktuSekarang() {
    const parts = new Intl.DateTimeFormat('sv-SE', {
      timeZone: 'Asia/Makassar',
      year: 'numeric', month: '2-digit', day: '2-digit',
      hour: '2-digit', minute: '2-digit', second: '2-digit', hourCycle: 'h23',
    }).formatToParts(new Date())
    const nilai = (type: Intl.DateTimeFormatPartTypes) => parts.find((part) => part.type === type)?.value || ''
    return {
      tanggal: `${nilai('year')}-${nilai('month')}-${nilai('day')}`,
      jam: `${nilai('hour')}:${nilai('minute')}:${nilai('second')}`,
    }
  }

  function emptyForm() {
    const waktu = waktuSekarang()
    const spesimen: SpesimenLaboratorium = {
      pengambilan_bahan: waktu.tanggal,
      diperoleh_dengan: '', lokasi_jaringan: '', diawetkan_dengan: '',
      pernah_dilakukan_di: '', tanggal_pa_sebelumnya: '',
      nomor_pa_sebelumnya: '', diagnosa_pa_sebelumnya: '',
    }
    return {
      no_rawat: props.patient.no_rawat || '',
      ...waktu,
      informasi_tambahan: '',
      diagnosis_klinis: '',
      spesimen,
    }
  }

  function rupiah(value: number) {
    return new Intl.NumberFormat('id-ID', {
      style: 'currency', currency: 'IDR', maximumFractionDigits: 0,
    }).format(Number(value || 0))
  }

  function formatDate(value: string) {
    if (!value) return '-'
    const [year, month, day] = value.split('-')
    return `${day}/${month}/${year}`
  }

  function teksPermintaan(item: PermintaanLaboratorium) {
    return [
      item.nomor, item.kategori, item.tanggal, item.jam,
      item.kode_dokter, item.nama_dokter, item.informasi_tambahan,
      item.diagnosis_klinis, item.status_pemeriksaan, item.status_bayar,
      ...Object.values(item.spesimen || {}),
      ...(item.pemeriksaan || []).flatMap((pemeriksaan) => [
        pemeriksaan.kode, pemeriksaan.nama, pemeriksaan.kelas,
        ...(pemeriksaan.detail || []).map((detail) => detail.nama),
      ]),
    ].join(' ').toLowerCase()
  }

  function waktuStatus(item: PermintaanLaboratorium) {
    if (item.status_pemeriksaan === 'Selesai') return `${formatDate(item.tanggal_hasil)} · ${item.jam_hasil || '-'}`
    if (item.status_pemeriksaan === 'Sampel Diterima') return `${formatDate(item.tanggal_sampel)} · ${item.jam_sampel || '-'}`
    return 'Belum diterima petugas'
  }

  function kelasStatusPemeriksaan(status: string) {
    if (status === 'Selesai') return 'completed'
    if (status === 'Sampel Diterima') return 'processing'
    return 'waiting'
  }

  function kelasStatusBayar(status: string) {
    if (status === 'Sudah Bayar') return 'paid'
    if (status === 'Sebagian Dibayar') return 'partial'
    return 'unpaid'
  }

  function resetForm() {
    Object.assign(form, emptyForm())
    doctor.value = { ...defaultDoctor.value }
    treatments.value = []
    editingNumber.value = ''
    confirmVisible.value = false
  }

  function gantiKategori(nilai: KategoriLaboratorium) {
    if (nilai === kategori.value || busy.value || detailLoading.value) return
    if ((treatments.value.length || form.informasi_tambahan || form.diagnosis_klinis || editingNumber.value)
      && !window.confirm('Ganti kategori dan kosongkan form yang belum disimpan?')) return
    kategori.value = nilai
  }

  async function editRequest(item: PermintaanLaboratorium) {
    if (busy.value || billingLocked.value || !item.dapat_diubah) return
    const versi = versiKonteks
    editing.value = true
    try {
      const pilihan = await Promise.all(item.pemeriksaan.map(async (pemeriksaan) => {
        const tersimpan = new Set(pemeriksaan.detail.map((detail) => Number(detail.id)))
        const detail: DetailLaboratorium[] = await detailTindakanLaboratorium(
          props.token, item.no_rawat, pemeriksaan.kode, item.kategori,
        )
        return {
          ...pemeriksaan,
          detail_opsi: detail.map((opsi) => ({ ...opsi, dipilih: tersimpan.has(Number(opsi.id)) })),
        }
      }))
      if (versi !== versiKonteks) return
      resetForm()
      editingNumber.value = item.nomor
      Object.assign(form, {
        no_rawat: item.no_rawat, tanggal: item.tanggal, jam: item.jam,
        informasi_tambahan: item.informasi_tambahan,
        diagnosis_klinis: item.diagnosis_klinis,
        spesimen: { ...emptyForm().spesimen, ...item.spesimen },
      })
      doctor.value = { kode: item.kode_dokter, nama: item.nama_dokter }
      treatments.value = pilihan
      formVisible.value = true
      await nextTick()
      document.querySelector('.laboratory-form-card')?.scrollIntoView({ behavior: 'smooth', block: 'start' })
    } catch (err) {
      if (versi === versiKonteks) notifikasi.gagal(pesanError(err))
    } finally {
      if (versi === versiKonteks) editing.value = false
    }
  }

  async function loadData() {
    const versi = ++versiMuat
    const konteks = versiKonteks
    if (!props.patient.no_rawat) return
    loading.value = true
    billingLocked.value = true
    error.value = ''
    try {
      const data = await permintaanLaboratoriumData(props.token, props.patient.no_rawat, kategori.value)
      if (versi !== versiMuat || konteks !== versiKonteks) return
      requests.value = data?.permintaan || []
      billingLocked.value = Boolean(data?.billing_terkunci)
      const jadwal = props.jadwalOperasi
      defaultDoctor.value = jadwal && jadwal.no_rawat === props.patient.no_rawat
        ? { kode: jadwal.kd_dokter, nama: jadwal.nama_dokter }
        : data?.dokter_perujuk || {}
      scope.value = { status: data?.status_rawat || '', kodeCaraBayar: data?.kode_cara_bayar || '', kelas: data?.kelas_pasien || '' }
      if (!doctor.value.kode) doctor.value = { ...defaultDoctor.value }
    } catch (err) {
      if (versi !== versiMuat || konteks !== versiKonteks) return
      requests.value = []
      error.value = pesanError(err)
      notifikasi.gagal(error.value)
    } finally {
      if (versi === versiMuat && konteks === versiKonteks) loading.value = false
    }
  }

  function saveRequest() {
    if (busy.value || billingLocked.value || detailLoading.value) return
    if (!doctor.value.kode || !treatments.value.length) {
      notifikasi.peringatan('Dokter perujuk dan minimal satu pemeriksaan wajib dipilih.')
      return
    }
    confirmVisible.value = true
  }

  async function confirmSave() {
    if (busy.value || billingLocked.value || !confirmVisible.value) return
    const versi = versiKonteks
    const payload = {
      ...form,
      spesimen: { ...form.spesimen },
      kategori: kategori.value,
      kode_dokter: doctor.value.kode,
      pemeriksaan: treatments.value.map((item) => ({
        kode: item.kode,
        id_detail: (item.detail_opsi || []).filter((detail) => detail.dipilih).map((detail) => detail.id),
      })),
    }
    saving.value = true
    try {
      const response = editingNumber.value
        ? await ubahPermintaanLaboratorium(props.token, editingNumber.value, payload)
        : await simpanPermintaanLaboratorium(props.token, payload)
      if (versi !== versiKonteks) return
      notifikasi.sukses(`${response?.pesan || 'Permintaan berhasil disimpan.'} Nomor: ${response?.nomor || '-'}`)
      resetForm()
      await loadData()
      if (versi === versiKonteks) formVisible.value = false
    } catch (err) {
      if (versi === versiKonteks) notifikasi.gagal(pesanError(err))
    } finally {
      if (versi === versiKonteks) {
        saving.value = false
        confirmVisible.value = false
      }
    }
  }

  async function confirmDelete() {
    if (!deleteTarget.value || busy.value || billingLocked.value) return
    const versi = versiKonteks
    const target = deleteTarget.value
    deleting.value = true
    try {
      const response = await hapusPermintaanLaboratorium(props.token, target.no_rawat, target.nomor, target.kategori)
      if (versi !== versiKonteks) return
      notifikasi.sukses(response?.pesan || 'Permintaan berhasil dihapus.')
      deleteTarget.value = null
      await loadData()
    } catch (err) {
      if (versi === versiKonteks) notifikasi.gagal(pesanError(err))
    } finally {
      if (versi === versiKonteks) deleting.value = false
    }
  }

  function cetak(item: PermintaanLaboratorium) {
    if (item.no_rawat !== props.patient.no_rawat) return
    cetakPermintaanLaboratorium(item, props.patient)
  }

  watch([() => props.patient.no_rawat, kategori], () => {
    versiKonteks++
    kataKunciPermintaan.value = ''
    requests.value = []
    defaultDoctor.value = {}
    deleteTarget.value = null
    saving.value = false
    deleting.value = false
    editing.value = false
    detailLoading.value = false
    loading.value = false
    billingLocked.value = true
    resetForm()
    void loadData()
  }, { immediate: true })

  return {
    kategori, kategoriLaboratorium, namaKategori, gantiKategori,
    loading, saving, deleting, busy, detailLoading, error, formVisible, confirmVisible,
    requests, doctor, treatments, billingLocked, scope, deleteTarget,
    editingNumber, kataKunciPermintaan, form, tableRowsTampil,
    rupiah, formatDate, waktuStatus, kelasStatusPemeriksaan, kelasStatusBayar,
    resetForm, editRequest, loadData, saveRequest, confirmSave, confirmDelete, cetak,
    waktuSekarang,
  }
}
