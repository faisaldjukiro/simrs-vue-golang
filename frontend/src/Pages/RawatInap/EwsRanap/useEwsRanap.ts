// @ts-nocheck -- migrasi TypeScript bertahap; kontrak data modul lama belum sepenuhnya bertipe.
import { computed, nextTick, reactive, ref, watch } from "vue"
import { ewsRanapData, hapusEwsRanap, simpanEwsRanap, ubahEwsRanap } from "../../../lib/faisal/api"
import { useNotifikasi } from "../../../lib/shared/useNotifikasi"

export function useEwsRanap(props) {
  const notifikasi = useNotifikasi()
  const loading = ref(false)
  const saving = ref(false)
  const deleting = ref(false)
  const error = ref('')
  const records = ref([])
  const petugas = ref({ nip: '', nama: '', jabatan: '' })
  const selectedOfficer = ref({ nip: '', nama: '', jabatan: '' })
  const canChooseOfficer = ref(false)
  const billingLocked = ref(false)
  const editingKey = ref(null)
  const deleteTarget = ref(null)
  const formVisible = ref(true)
  const kataKunciRiwayat = ref('')
  
  const pilihanAlat = ref(toSelectOptions(['Ya', 'Tidak']))
  const pilihanKesadaran = ref(toSelectOptions(['A', 'P-V-U']))
  const pilihanSkalaNyeri = ref(toSelectOptions(['0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '10']))
  const form = reactive(emptyForm())
  const validationErrors = reactive({})
  
  const tableRecords = computed(() => records.value.map((item) => ({ ...item, _key: keyFor(item) })))
  const tableRecordsTampil = computed(() => {
    const kata = kataKunciRiwayat.value.trim().toLowerCase()
    if (!kata) return tableRecords.value
    return tableRecords.value.filter((item) => teksRiwayat(item).includes(kata))
  })
  
  const totalMasuk = computed(() => toNumber(form.masuk1) + toNumber(form.masuk2))
  const totalKeluar = computed(() => (
    toNumber(form.keluar1) + toNumber(form.keluar2) + toNumber(form.keluar3) + toNumber(form.keluar4) + toNumber(form.keluar5)
  ))
  const balanceCairan = computed(() => totalMasuk.value - totalKeluar.value)
  
  watch(
    () => [
      form.pernafasan, form.saturasi, form.alat, form.suhu, form.denyut, form.tekanan,
      form.kesadaran, form.masuk1, form.masuk2, form.keluar1, form.keluar2, form.keluar3,
      form.keluar4, form.keluar5,
    ],
    hitungOtomatis,
    { immediate: true },
  )
  
  watch(() => selectedOfficer.value, (value) => {
    form.nip = value?.nip || value?.kode || ''
    if (form.nip) delete validationErrors.nip
  }, { deep: true })
  
  watch(
    () => daftarFieldWajib().map((field) => field.nilai()),
    () => {
      for (const field of daftarFieldWajib()) {
        if (!fieldKosong(field.nilai())) delete validationErrors[field.nama]
      }
    },
  )
  
  watch(() => props.patient?.no_rawat, () => loadRecords(), { immediate: true })
  
  function toSelectOptions(options = []) {
    return options.map((option) => ({ label: String(option), value: String(option) }))
  }
  
  function currentDate() {
    const date = new Date()
    const offset = date.getTimezoneOffset() * 60000
    return new Date(date.getTime() - offset).toISOString().slice(0, 10)
  }
  
  function currentTime() {
    const date = new Date()
    return [date.getHours(), date.getMinutes(), date.getSeconds()]
      .map((value) => String(value).padStart(2, '0'))
      .join(':')
  }
  
  function emptyForm() {
    return {
      no_rawat: props.patient?.no_rawat || '',
      tanggal: currentDate(),
      jam: currentTime(),
      nip: '',
      pernafasan: '',
      score_pernafasan: '',
      saturasi: '',
      score_saturasi: '',
      alat: 'Tidak',
      score_alat: '0',
      suhu: '',
      score_suhu: '',
      denyut: '',
      score_denyut: '',
      tekanan: '',
      diastol: '',
      score_tekanan: '',
      kesadaran: 'A',
      score_kesadaran: '0',
      total_score: '0',
      klasifikasi: 'Sangat Rendah',
      respon: 'Dilakukan monitoring',
      tindakan: 'Melanjutkan monitoring',
      frekuensi: 'Minimal 12 Jam',
      skala_nyeri: '0',
      bb: '',
      tb: '',
      lk: '',
      lp: '',
      masuk1: '0',
      masuk2: '0',
      jumlahmasuk: '0',
      keluar1: '0',
      keluar2: '0',
      keluar3: '0',
      keluar4: '0',
      keluar5: '0',
      jumlahkeluar: '0',
      bc: '0',
    }
  }
  
  function resetForm() {
    Object.assign(form, emptyForm(), {
      no_rawat: props.patient.no_rawat,
      nip: petugas.value.nip || '',
    })
    editingKey.value = null
    selectedOfficer.value = { ...petugas.value }
    hapusSemuaErrorValidasi()
    hitungOtomatis()
  }
  
  function keyFor(item) {
    return `${item.no_rawat}|${item.tanggal}|${item.jam}`
  }
  
  function recordKey(item) {
    return {
      no_rawat: nilaiPayload(item.no_rawat || props.patient.no_rawat),
      tanggal: nilaiPayload(item.tanggal),
      jam: nilaiPayload(item.jam),
    }
  }
  
  function teksRiwayat(item) {
    return [
      item.tanggal, item.jam, item.nip, item.nama_petugas, item.jabatan, item.total_score,
      item.klasifikasi, item.respon, item.tindakan, item.frekuensi, item.kesadaran,
      item.pernafasan, item.saturasi, item.suhu, item.denyut, item.tekanan, item.diastol,
    ].join(' ').toLowerCase()
  }
  
  async function loadRecords() {
    if (!props.token || !props.patient?.no_rawat) return
    loading.value = true
    error.value = ''
    try {
      const data = await ewsRanapData(props.token, props.patient.no_rawat)
      records.value = data?.catatan || []
      petugas.value = data?.petugas || { nip: '', nama: '', jabatan: '' }
      canChooseOfficer.value = Boolean(data?.bisa_memilih_petugas)
      billingLocked.value = Boolean(data?.billing_terkunci)
      pilihanAlat.value = toSelectOptions(data?.pilihan_alat?.length ? data.pilihan_alat : ['Ya', 'Tidak'])
      pilihanKesadaran.value = toSelectOptions(data?.pilihan_kesadaran?.length ? data.pilihan_kesadaran : ['A', 'P-V-U'])
      pilihanSkalaNyeri.value = toSelectOptions(data?.pilihan_skala_nyeri?.length ? data.pilihan_skala_nyeri : ['0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '10'])
      resetForm()
      if (billingLocked.value && records.value.length) {
        fillForm(records.value[0], false)
        formVisible.value = true
      }
    } catch (err) {
      error.value = err.message
      notifikasi.gagal(err.message || 'EWS Ranap tidak dapat dibaca.')
    } finally {
      loading.value = false
    }
  }
  
  function editRecord(item) {
    if (billingLocked.value) {
      notifikasi.peringatan('Kunjungan sudah masuk billing. EWS Ranap hanya dapat dilihat.')
      return
    }
    if (!item.bisa_diubah) {
      notifikasi.peringatan('EWS ini hanya dapat diedit oleh petugas yang membuatnya.')
      return
    }
    editingKey.value = recordKey(item)
    formVisible.value = true
    fillForm(item, true)
    nextTick(() => document.querySelector('.ews-ranap-page .clinical-form-card')?.scrollIntoView({ behavior: 'smooth', block: 'start' }))
  }
  
  function viewRecord(item) {
    editingKey.value = null
    formVisible.value = true
    fillForm(item, false)
    nextTick(() => document.querySelector('.ews-ranap-page .clinical-form-card')?.scrollIntoView({ behavior: 'smooth', block: 'start' }))
  }
  
  function fillForm(item, sebagaiEdit = false) {
    if (sebagaiEdit) editingKey.value = recordKey(item)
    hapusSemuaErrorValidasi()
    selectedOfficer.value = {
      nip: item.nip || '',
      nama: item.nama_petugas || item.nip || '',
      jabatan: item.jabatan || '',
    }
    Object.assign(form, {
      ...emptyForm(),
      ...item,
      no_rawat: props.patient.no_rawat,
      nip: item.nip,
    })
    hitungOtomatis()
  }
  
  function payload() {
    hitungOtomatis()
    const body = {
      no_rawat: props.patient.no_rawat,
      nip: selectedOfficer.value?.nip || selectedOfficer.value?.kode || form.nip || petugas.value.nip,
      tanggal: form.tanggal,
      jam: form.jam,
      pernafasan: form.pernafasan,
      score_pernafasan: form.score_pernafasan,
      saturasi: form.saturasi,
      score_saturasi: form.score_saturasi,
      alat: form.alat,
      score_alat: form.score_alat,
      suhu: form.suhu,
      score_suhu: form.score_suhu,
      denyut: form.denyut,
      score_denyut: form.score_denyut,
      tekanan: form.tekanan,
      diastol: form.diastol,
      score_tekanan: form.score_tekanan,
      kesadaran: form.kesadaran,
      score_kesadaran: form.score_kesadaran,
      total_score: form.total_score,
      klasifikasi: form.klasifikasi,
      respon: form.respon,
      tindakan: form.tindakan,
      frekuensi: form.frekuensi,
      skala_nyeri: form.skala_nyeri,
      bb: form.bb,
      tb: form.tb,
      lk: form.lk,
      lp: form.lp,
      masuk1: form.masuk1,
      masuk2: form.masuk2,
      jumlahmasuk: form.jumlahmasuk,
      keluar1: form.keluar1,
      keluar2: form.keluar2,
      keluar3: form.keluar3,
      keluar4: form.keluar4,
      keluar5: form.keluar5,
      jumlahkeluar: form.jumlahkeluar,
      bc: form.bc,
    }
    return Object.fromEntries(Object.entries(body).map(([key, value]) => [key, nilaiPayload(value)]))
  }
  
  function nilaiPayload(value) {
    if (value === undefined || value === null) return ''
    return String(value).trim()
  }
  
  async function submitForm() {
    if (billingLocked.value) {
      notifikasi.peringatan('Kunjungan sudah masuk billing. EWS Ranap tidak dapat disimpan.')
      return
    }
    if (!validasiFormEws()) return
    saving.value = true
    try {
      const body = { ...payload() }
      if (editingKey.value) await ubahEwsRanap(props.token, { kunci_lama: editingKey.value, ...body })
      else await simpanEwsRanap(props.token, body)
      notifikasi.sukses(editingKey.value ? 'EWS Ranap berhasil diperbarui.' : 'EWS Ranap berhasil disimpan.')
      await loadRecords()
      formVisible.value = true
    } catch (err) {
      fokusDariPesanBackend(err.message)
      notifikasi.gagal(err.message || 'EWS Ranap tidak dapat disimpan.')
    } finally {
      saving.value = false
    }
  }
  
  function daftarFieldWajib() {
    return [
      { nama: 'tanggal', label: 'Tanggal', nilai: () => form.tanggal },
      { nama: 'jam', label: 'Jam', nilai: () => form.jam },
      { nama: 'nip', label: 'Petugas', nilai: () => selectedOfficer.value?.nip || selectedOfficer.value?.kode || form.nip },
      { nama: 'pernafasan', label: 'Pernafasan', nilai: () => form.pernafasan },
      { nama: 'saturasi', label: 'Saturasi O2', nilai: () => form.saturasi },
      { nama: 'alat', label: 'Alat Bantu O2', nilai: () => form.alat },
      { nama: 'suhu', label: 'Suhu', nilai: () => form.suhu },
      { nama: 'denyut', label: 'Denyut Jantung', nilai: () => form.denyut },
      { nama: 'tekanan', label: 'Sistolik', nilai: () => form.tekanan },
      { nama: 'diastol', label: 'Diastolik', nilai: () => form.diastol },
      { nama: 'kesadaran', label: 'Kesadaran', nilai: () => form.kesadaran },
      { nama: 'respon', label: 'Respon Klinis', nilai: () => form.respon },
      { nama: 'tindakan', label: 'Tindakan', nilai: () => form.tindakan },
      { nama: 'frekuensi', label: 'Frekuensi Monitoring', nilai: () => form.frekuensi },
    ]
  }
  
  function fieldKosong(value) {
    return value === undefined || value === null || String(value).trim() === ''
  }
  
  function hapusSemuaErrorValidasi() {
    Object.keys(validationErrors).forEach((key) => delete validationErrors[key])
  }
  
  function validasiFormEws() {
    hapusSemuaErrorValidasi()
    hitungOtomatis()
    const fieldKosongPertama = daftarFieldWajib().find((field) => fieldKosong(field.nilai()))
    if (!fieldKosongPertama) return true
  
    validationErrors[fieldKosongPertama.nama] = `${fieldKosongPertama.label} wajib diisi`
    notifikasi.peringatan(`${fieldKosongPertama.label} wajib diisi.`)
    fokusKeField(fieldKosongPertama.nama)
    return false
  }
  
  function fokusKeField(namaField) {
    const selector = namaField === 'nip'
      ? '.ews-ranap-page .ews-field-petugas input, .ews-ranap-page .ews-field-petugas button'
      : `#ews-${namaField}`
  
    formVisible.value = true
    nextTick(() => {
      const element = document.querySelector(selector)
      const target = element?.matches?.('input, textarea, select, button, [tabindex]')
        ? element
        : element?.querySelector?.('input, textarea, select, button, [tabindex]')
      target?.scrollIntoView({ behavior: 'smooth', block: 'center' })
      window.setTimeout(() => {
        target?.focus?.({ preventScroll: true })
        target?.select?.()
      }, 250)
    })
  }
  
  function fokusDariPesanBackend(pesan = '') {
    const teks = String(pesan).toLowerCase()
    const peta = [
      ['petugas', 'nip'],
      ['tanggal', 'tanggal'],
      ['jam', 'jam'],
      ['pernafasan', 'pernafasan'],
      ['saturasi', 'saturasi'],
      ['alat bantu', 'alat'],
      ['suhu', 'suhu'],
      ['denyut', 'denyut'],
      ['sistolik', 'tekanan'],
      ['diastolik', 'diastol'],
      ['tekanan', 'tekanan'],
      ['kesadaran', 'kesadaran'],
      ['respon', 'respon'],
      ['tindakan', 'tindakan'],
      ['frekuensi', 'frekuensi'],
    ]
    const cocok = peta.find(([kata]) => teks.includes(kata))
    if (!cocok) return
    const field = daftarFieldWajib().find((item) => item.nama === cocok[1])
    if (!field) return
    validationErrors[field.nama] = pesan
    fokusKeField(field.nama)
  }
  
  async function deleteRecord(item) {
    if (billingLocked.value) {
      notifikasi.peringatan('Kunjungan sudah masuk billing. EWS Ranap tidak dapat dihapus.')
      return
    }
    if (!item.bisa_diubah) {
      notifikasi.peringatan('EWS ini hanya dapat dihapus oleh petugas yang membuatnya.')
      return
    }
    deleting.value = true
    try {
      await hapusEwsRanap(props.token, recordKey(item))
      notifikasi.sukses('EWS Ranap berhasil dihapus.')
      deleteTarget.value = null
      await loadRecords()
    } catch (err) {
      notifikasi.gagal(err.message || 'EWS Ranap tidak dapat dihapus.')
    } finally {
      deleting.value = false
    }
  }
  
  function hitungOtomatis() {
    form.score_pernafasan = skorPernafasan(form.pernafasan)
    form.score_saturasi = skorSaturasi(form.saturasi)
    form.score_alat = form.alat === 'Ya' ? '2' : '0'
    form.score_suhu = skorSuhu(form.suhu)
    form.score_denyut = skorDenyut(form.denyut)
    form.score_tekanan = skorTekanan(form.tekanan)
    form.score_kesadaran = form.kesadaran === 'P-V-U' ? '3' : '0'
    const total = [
      form.score_pernafasan, form.score_saturasi, form.score_alat, form.score_suhu,
      form.score_denyut, form.score_tekanan, form.score_kesadaran,
    ].reduce((sum, nilai) => sum + toNumber(nilai), 0)
    form.total_score = String(total)
    form.klasifikasi = klasifikasi(total)
    form.respon = responKlinis(total)
    form.tindakan = tindakanKlinis(total)
    form.frekuensi = frekuensiMonitoring(total)
    form.jumlahmasuk = String(totalMasuk.value)
    form.jumlahkeluar = String(totalKeluar.value)
    form.bc = String(balanceCairan.value)
  }
  
  function toNumber(value) {
    const parsed = Number.parseFloat(String(value ?? '').replace(',', '.'))
    return Number.isFinite(parsed) ? parsed : 0
  }
  
  function skorPernafasan(value) {
    const n = toNumber(value)
    if (!n) return ''
    if (n <= 8) return '3'
    if (n <= 11) return '1'
    if (n <= 20) return '0'
    if (n <= 24) return '2'
    return '3'
  }
  
  function skorSaturasi(value) {
    const n = toNumber(value)
    if (!n) return ''
    if (n <= 91) return '3'
    if (n <= 93) return '2'
    if (n <= 95) return '1'
    return '0'
  }
  
  function skorSuhu(value) {
    const n = toNumber(value)
    if (!n) return ''
    if (n <= 35) return '3'
    if (n <= 36) return '1'
    if (n <= 38) return '0'
    if (n <= 39) return '1'
    return '2'
  }
  
  function skorDenyut(value) {
    const n = toNumber(value)
    if (!n) return ''
    if (n <= 40) return '3'
    if (n <= 50) return '1'
    if (n <= 90) return '0'
    if (n <= 110) return '1'
    if (n <= 130) return '2'
    return '3'
  }
  
  function skorTekanan(value) {
    const n = toNumber(value)
    if (!n) return ''
    if (n <= 90) return '3'
    if (n <= 100) return '2'
    if (n <= 110) return '1'
    if (n <= 219) return '0'
    return '3'
  }
  
  function klasifikasi(total) {
    if (total === 0) return 'Sangat Rendah'
    if (total <= 4) return 'Rendah'
    if (total <= 6) return 'Sedang'
    return 'Tinggi'
  }
  
  function responKlinis(total) {
    if (total === 0) return 'Dilakukan monitoring'
    if (total <= 4) return 'Harus segera dievaluasi oleh perawat terdaftar yang kompeten, harus memutuskan apakah perubahan frekuensi pemantauan klinis atau wajib eskalasi perawatan klinis'
    if (total <= 6) return 'Harus segera melakukan tinjauan mendesak oleh klinis yang terampil dengan kompetensi dalam penilaian penyakit akut di bangsal, biasanya oleh dokter atau perawat dengan mempertimbangkan apakah eskalasi perawatan ke tim perawatan kritis diperlukan'
    return 'Harus segera memberikan penilaian darurat secara klinis oleh tim critical care outreach atau code blue dengan kompetensi penanganan pasien kritis dan biasanya terjadi transfer pasien ke area perawatan dengan alat bantu'
  }
  
  function tindakanKlinis(total) {
    if (total === 0) return 'Melanjutkan monitoring'
    if (total <= 4) return 'Perawat melakukan assessment atau meningkatkan frekuensi monitor'
    if (total <= 6) return 'Perawat berkolaborasi dengan tim/pemberian assessment kegawatan/meningkatkan perawatan dengan fasilitas monitor yang lengkap'
    return 'Berkolaborasi dengan tim medis/pemberian assessment kegawatan/pindah ruang HCU/ICU'
  }
  
  function frekuensiMonitoring(total) {
    if (total === 0) return 'Minimal 12 Jam'
    if (total <= 4) return 'Minimal 4-6 Jam'
    if (total <= 6) return 'Minimal 1 Jam'
    return 'Bed side monitor/every time'
  }
  
  function scoreClass(total) {
    const n = toNumber(total)
    if (n >= 7) return 'danger'
    if (n >= 5) return 'warning'
    if (n >= 1) return 'info'
    return 'safe'
  }
  return {
    loading,
    saving,
    deleting,
    error,
    records,
    petugas,
    selectedOfficer,
    canChooseOfficer,
    billingLocked,
    editingKey,
    deleteTarget,
    formVisible,
    kataKunciRiwayat,
    pilihanAlat,
    pilihanKesadaran,
    pilihanSkalaNyeri,
    form,
    validationErrors,
    tableRecordsTampil,
    resetForm,
    loadRecords,
    editRecord,
    viewRecord,
    submitForm,
    deleteRecord,
    klasifikasi,
    scoreClass,
  }
}
