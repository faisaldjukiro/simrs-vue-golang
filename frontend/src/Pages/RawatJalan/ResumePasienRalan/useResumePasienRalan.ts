// @ts-nocheck -- migrasi TypeScript bertahap; kontrak data modul lama belum sepenuhnya bertipe.
import { computed, reactive, ref, watch } from "vue"
import { cariCodingResumePasien, hapusResumePasien, referensiResumePasien, resumePasienData, simpanResumePasien, ubahResumePasien, validasiCodingResumePasien } from "../../../lib/faisal/api"
import { useNotifikasi } from "../../../lib/shared/useNotifikasi"

export function useResumePasienRalan(props) {
  const notifikasi = useNotifikasi()
  const loading = ref(false)
  const saving = ref(false)
  const deleting = ref(false)
  const tersedia = ref(false)
  const billingLocked = ref(false)
  const hapusTerbuka = ref(false)
  const doctor = ref({})
  const pilihan = reactive({ kondisi_pulang: ['Hidup', 'Meninggal'] })
  const form = reactive(formKosong())
  const statusCoding = reactive({})
  const referensiTerbuka = ref(null)
  const referensiLoading = ref(false)
  const referensiSearch = ref('')
  const referensiItems = ref([])
  const referensiDipilih = ref([])
  
  const pilihanSelect = computed(() => ({
    kondisi_pulang: pilihan.kondisi_pulang.map((value) => ({ label: value, value })),
  }))
  
  const catatanResume = [
    {
      key: 'keluhan_utama',
      label: 'Keluhan Utama / Riwayat Penyakit',
      required: true,
      references: [
        { jenis: 'keluhan', label: 'Keluhan' },
        { jenis: 'pemeriksaan', label: 'Pemeriksaan' },
      ],
    },
    { key: 'jalannya_penyakit', label: 'Jalannya Penyakit', required: true },
    { key: 'pemeriksaan_penunjang', label: 'Pemeriksaan Penunjang', references: [{ jenis: 'radiologi', label: 'Radiologi' }] },
    { key: 'hasil_laborat', label: 'Hasil Laboratorium', references: [{ jenis: 'laboratorium', label: 'Laboratorium' }] },
    { key: 'obat_pulang', label: 'Obat Pulang', references: [{ jenis: 'obat', label: 'Obat Pasien' }] },
  ]
  
  const referensiMeta = {
    keluhan: {
      title: 'Ambil Keluhan Pasien',
      description: 'Data diambil dari keluhan pemeriksaan rawat jalan/rawat inap pada nomor rawat ini.',
      target: 'keluhan_utama',
    },
    pemeriksaan: {
      title: 'Ambil Pemeriksaan / SOAP',
      description: 'Data diambil dari kolom pemeriksaan rawat jalan/rawat inap pada nomor rawat ini.',
      target: 'keluhan_utama',
    },
    radiologi: {
      title: 'Ambil Hasil Radiologi',
      description: 'Data diambil dari hasil radiologi pasien.',
      target: 'pemeriksaan_penunjang',
    },
    laboratorium: {
      title: 'Ambil Hasil Laboratorium',
      description: 'Data diambil dari detail hasil laboratorium pasien.',
      target: 'hasil_laborat',
    },
    obat: {
      title: 'Ambil Obat Pasien',
      description: 'Data diambil dari obat yang sudah pernah diberikan pada pasien.',
      target: 'obat_pulang',
    },
  }
  
  const referensiSemuaTerpilih = computed(() => (
    referensiItems.value.length > 0
    && referensiItems.value.every((item) => referensiDipilih.value.some((dipilih) => referensiKey(dipilih) === referensiKey(item)))
  ))
  
  const diagnosa = [
    { kode: 'kd_diagnosa_utama', nama: 'diagnosa_utama', label: 'Diagnosa Utama', required: true },
    { kode: 'kd_diagnosa_sekunder', nama: 'diagnosa_sekunder', label: 'Diagnosa Sekunder 1' },
    { kode: 'kd_diagnosa_sekunder2', nama: 'diagnosa_sekunder2', label: 'Diagnosa Sekunder 2' },
    { kode: 'kd_diagnosa_sekunder3', nama: 'diagnosa_sekunder3', label: 'Diagnosa Sekunder 3' },
    { kode: 'kd_diagnosa_sekunder4', nama: 'diagnosa_sekunder4', label: 'Diagnosa Sekunder 4' },
  ]
  
  const prosedur = [
    { kode: 'kd_prosedur_utama', nama: 'prosedur_utama', label: 'Prosedur Utama' },
    { kode: 'kd_prosedur_sekunder', nama: 'prosedur_sekunder', label: 'Prosedur Sekunder 1' },
    { kode: 'kd_prosedur_sekunder2', nama: 'prosedur_sekunder2', label: 'Prosedur Sekunder 2' },
    { kode: 'kd_prosedur_sekunder3', nama: 'prosedur_sekunder3', label: 'Prosedur Sekunder 3' },
  ]
  
  function formKosong() {
    return {
      no_rawat: props.patient?.no_rawat || '',
      kode_dokter: '',
      kondisi_pulang: 'Hidup',
      keluhan_utama: '',
      jalannya_penyakit: '',
      pemeriksaan_penunjang: '',
      hasil_laborat: '',
      diagnosa_utama: '',
      kd_diagnosa_utama: '',
      diagnosa_sekunder: '',
      kd_diagnosa_sekunder: '',
      diagnosa_sekunder2: '',
      kd_diagnosa_sekunder2: '',
      diagnosa_sekunder3: '',
      kd_diagnosa_sekunder3: '',
      diagnosa_sekunder4: '',
      kd_diagnosa_sekunder4: '',
      prosedur_utama: '',
      kd_prosedur_utama: '',
      prosedur_sekunder: '',
      kd_prosedur_sekunder: '',
      prosedur_sekunder2: '',
      kd_prosedur_sekunder2: '',
      prosedur_sekunder3: '',
      kd_prosedur_sekunder3: '',
      obat_pulang: '',
    }
  }
  
  function resetStatusCoding() {
    Object.keys(statusCoding).forEach((key) => delete statusCoding[key])
  }
  
  function dariServer(value = {}) {
    resetStatusCoding()
    Object.assign(form, formKosong(), value, { no_rawat: props.patient.no_rawat })
  }
  
  function daftarCoding(jenis) {
    return jenis === 'diagnosa' ? diagnosa : prosedur
  }
  
  function kodeUtama(jenis, item) {
    const daftar = daftarCoding(jenis)
    const posisi = daftar.findIndex((entry) => entry.kode === item.kode)
    return posisi <= 0 || !daftar.slice(0, posisi).some((entry) => String(form[entry.kode] || '').trim())
  }
  
  function kodeDuplikat(jenis, item, kode) {
    return daftarCoding(jenis).some((entry) => entry.kode !== item.kode && String(form[entry.kode] || '').trim().toUpperCase() === kode)
  }
  
  async function periksaCoding(item, jenis) {
    const kode = String(form[item.kode] || '').trim().toUpperCase()
    if (!kode) {
      form[item.nama] = ''
      delete statusCoding[item.kode]
      return true
    }
    if (kodeDuplikat(jenis, item, kode)) {
      form[item.nama] = ''
      statusCoding[item.kode] = { valid: false, error: `Kode ${kode} sudah digunakan.` }
      return false
    }
    statusCoding[item.kode] = { loading: true, hint: 'Memeriksa kode ke master SIMRS...' }
    try {
      const data = await validasiCodingResumePasien(props.token, {
        jenis,
        kode,
        utama: kodeUtama(jenis, item) ? '1' : '0',
      })
      if (String(form[item.kode] || '').trim().toUpperCase() !== kode) return false
      if (!data?.valid) {
        form[item.nama] = ''
        statusCoding[item.kode] = { valid: false, error: data?.pesan || `Kode ${kode} tidak valid.` }
        return false
      }
      form[item.kode] = data.kode
      form[item.nama] = data.nama
      statusCoding[item.kode] = { valid: true, hint: data.pesan || 'Kode valid', im: data.im === '1' }
      return true
    } catch (error) {
      if (String(form[item.kode] || '').trim().toUpperCase() !== kode) return false
      form[item.nama] = ''
      statusCoding[item.kode] = { valid: false, error: error.message || 'Kode tidak dapat diperiksa.' }
      return false
    }
  }
  
  function nilaiCoding(item) {
    if (!form[item.kode]) return {}
    return {
      kode: form[item.kode],
      nama: form[item.nama],
      penanda: statusCoding[item.kode]?.im ? 'IM' : '',
    }
  }
  
  async function cariPilihanCoding(item, jenis, kataKunci) {
    return cariCodingResumePasien(props.token, {
      jenis,
      q: kataKunci,
      utama: kodeUtama(jenis, item) ? '1' : '0',
    })
  }
  
  function pilihCoding(item, jenis, pilihanCoding) {
    if (!pilihanCoding?.kode) {
      form[item.kode] = ''
      form[item.nama] = ''
      delete statusCoding[item.kode]
      return
    }
    const kode = String(pilihanCoding.kode).trim().toUpperCase()
    if (!pilihanCoding.valid) {
      form[item.kode] = kode
      form[item.nama] = pilihanCoding.nama || ''
      statusCoding[item.kode] = { valid: false, error: pilihanCoding.pesan || `Kode ${kode} tidak valid.` }
      notifikasi.peringatan(pilihanCoding.pesan || `Kode ${kode} tidak valid.`)
      return
    }
    if (kodeDuplikat(jenis, item, kode)) {
      form[item.kode] = kode
      form[item.nama] = pilihanCoding.nama || ''
      statusCoding[item.kode] = { valid: false, error: `Kode ${kode} sudah digunakan.` }
      notifikasi.peringatan(`Kode ${kode} sudah digunakan.`)
      return
    }
    form[item.kode] = kode
    form[item.nama] = pilihanCoding.nama || ''
    statusCoding[item.kode] = { valid: true, hint: pilihanCoding.pesan || 'Kode valid', im: pilihanCoding.im === '1' }
  }
  
  async function periksaSemuaCoding() {
    const daftar = [
      ...diagnosa.filter((item) => String(form[item.kode] || '').trim()).map((item) => [item, 'diagnosa']),
      ...prosedur.filter((item) => String(form[item.kode] || '').trim()).map((item) => [item, 'prosedur']),
    ]
    const hasil = await Promise.all(daftar.map(([item, jenis]) => periksaCoding(item, jenis)))
    return hasil.every(Boolean)
  }
  
  async function bukaReferensi(jenis) {
    if (billingLocked.value) return
    referensiTerbuka.value = { jenis, ...referensiMeta[jenis] }
    referensiSearch.value = ''
    referensiDipilih.value = []
    await muatReferensi()
  }
  
  async function muatReferensi() {
    if (!referensiTerbuka.value?.jenis) return
    referensiLoading.value = true
    try {
      referensiItems.value = await referensiResumePasien(props.token, {
        no_rawat: props.patient.no_rawat,
        jenis: referensiTerbuka.value.jenis,
        q: referensiSearch.value,
      })
      referensiDipilih.value = []
    } catch (error) {
      referensiItems.value = []
      referensiDipilih.value = []
      notifikasi.gagal(error.message || 'Referensi resume pasien tidak dapat dibaca.')
    } finally {
      referensiLoading.value = false
    }
  }
  
  function referensiKey(item) {
    return `${item?.sumber || ''}|${item?.tanggal || ''}|${item?.jam || ''}|${item?.isi || ''}`
  }
  
  function referensiAktif(item) {
    const key = referensiKey(item)
    return referensiDipilih.value.some((dipilih) => referensiKey(dipilih) === key)
  }
  
  function toggleReferensi(item) {
    const key = referensiKey(item)
    if (referensiAktif(item)) {
      referensiDipilih.value = referensiDipilih.value.filter((dipilih) => referensiKey(dipilih) !== key)
      return
    }
    referensiDipilih.value = [...referensiDipilih.value, item]
  }
  
  function pilihSemuaReferensi() {
    if (!referensiItems.value.length) return
    referensiDipilih.value = referensiSemuaTerpilih.value ? [] : [...referensiItems.value]
  }
  
  function tambahReferensiTerpilih() {
    const target = referensiTerbuka.value?.target
    const isi = referensiDipilih.value.map((item) => String(item?.isi || '').trim()).filter(Boolean).join(', ')
    if (!target || !isi) return
    const sebelumnya = String(form[target] || '').trim()
    form[target] = sebelumnya ? `${sebelumnya}, ${isi}` : isi
    referensiTerbuka.value = null
    referensiDipilih.value = []
  }
  
  async function muat() {
    if (!props.patient.no_rawat) return
    loading.value = true
    try {
      const data = await resumePasienData(props.token, props.patient.no_rawat)
      tersedia.value = Boolean(data?.tersedia)
      billingLocked.value = Boolean(data?.billing_terkunci)
      Object.assign(pilihan, data?.pilihan || {})
      dariServer(data?.resume)
      doctor.value = data?.dokter || {}
    } catch (error) {
      notifikasi.gagal(error.message || 'Resume pasien rawat jalan tidak dapat dibaca.')
    } finally {
      loading.value = false
    }
  }
  
  function payload() {
    return {
      ...form,
      no_rawat: props.patient.no_rawat,
      kode_dokter: doctor.value?.kode || '',
    }
  }
  
  async function simpan() {
    if (billingLocked.value) {
      notifikasi.peringatan('Billing sudah terverifikasi. Resume hanya dapat dilihat.')
      return
    }
    if (!doctor.value?.kode || !form.keluhan_utama.trim() || !form.jalannya_penyakit.trim() || !form.kd_diagnosa_utama.trim()) {
      notifikasi.peringatan('Dokter, keluhan utama, jalannya penyakit, dan diagnosa utama wajib diisi.')
      return
    }
    if (!await periksaSemuaCoding()) {
      notifikasi.peringatan('Periksa kembali kode diagnosa atau prosedur yang ditandai.')
      return
    }
    saving.value = true
    try {
      const response = tersedia.value
        ? await ubahResumePasien(props.token, payload())
        : await simpanResumePasien(props.token, payload())
      notifikasi.sukses(response?.pesan || 'Resume pasien rawat jalan berhasil disimpan.')
      await muat()
    } catch (error) {
      notifikasi.gagal(error.message || 'Resume pasien rawat jalan gagal disimpan.')
    } finally {
      saving.value = false
    }
  }
  
  async function hapus() {
    deleting.value = true
    try {
      const response = await hapusResumePasien(props.token, props.patient.no_rawat)
      notifikasi.sukses(response?.pesan || 'Resume pasien rawat jalan berhasil dihapus.')
      hapusTerbuka.value = false
      await muat()
    } catch (error) {
      notifikasi.gagal(error.message || 'Resume pasien rawat jalan gagal dihapus.')
    } finally {
      deleting.value = false
    }
  }
  
  watch(() => props.patient.no_rawat, muat, { immediate: true })
  return {
    loading,
    saving,
    deleting,
    tersedia,
    billingLocked,
    hapusTerbuka,
    doctor,
    form,
    statusCoding,
    referensiTerbuka,
    referensiLoading,
    referensiSearch,
    referensiItems,
    referensiDipilih,
    pilihanSelect,
    catatanResume,
    referensiSemuaTerpilih,
    diagnosa,
    prosedur,
    nilaiCoding,
    cariPilihanCoding,
    pilihCoding,
    bukaReferensi,
    muatReferensi,
    referensiAktif,
    toggleReferensi,
    pilihSemuaReferensi,
    tambahReferensiTerpilih,
    simpan,
    hapus,
  }
}
