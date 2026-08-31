// @ts-nocheck -- migrasi TypeScript bertahap; kontrak data modul lama belum sepenuhnya bertipe.
import { computed, reactive, ref, watch } from "vue"
import { cariCodingResumePasienRanap, hapusResumePasienRanap, referensiResumePasienRanap, resumePasienRanapData, simpanResumePasienRanap, ubahResumePasienRanap, validasiCodingResumePasienRanap } from "../../../lib/faisal/api"
import { useNotifikasi } from "../../../lib/shared/useNotifikasi"

export function useResumePasienRanap(props) {
  const notifikasi = useNotifikasi()
  const loading = ref(false)
  const saving = ref(false)
  const deleting = ref(false)
  const tersedia = ref(false)
  const billingLocked = ref(false)
  const hapusTerbuka = ref(false)
  const doctor = ref({})
  const pilihan = reactive({ cara_keluar: [], keadaan: [], dilanjutkan: [] })
  const form = reactive(formKosong())
  const statusCoding = reactive({})
  const referensiTerbuka = ref(null)
  const referensiLoading = ref(false)
  const referensiSearch = ref('')
  const referensiItems = ref([])
  const referensiDipilih = ref([])
  
  const referensiSemuaTerpilih = computed(() => (
    referensiItems.value.length > 0 && referensiItems.value.every((item) => referensiAktif(item))
  ))
  
  const pilihanSelect = computed(() => ({
    cara_keluar: pilihan.cara_keluar.map((value) => ({ label: value, value })),
    keadaan: pilihan.keadaan.map((value) => ({ label: value, value })),
    dilanjutkan: pilihan.dilanjutkan.map((value) => ({ label: value, value })),
  }))
  
  const catatanPerawatan = [
    { key: 'keluhan_utama', label: 'Keluhan Utama / Riwayat Penyakit', required: true, references: [{ jenis: 'keluhan', label: 'Keluhan' }] },
    { key: 'pemeriksaan_fisik', label: 'Pemeriksaan Fisik', references: [{ jenis: 'pemeriksaan', label: 'Pemeriksaan' }] },
    { key: 'jalannya_penyakit', label: 'Jalannya Penyakit Selama Perawatan', required: true },
    { key: 'pemeriksaan_penunjang', label: 'Pemeriksaan Penunjang Radiologi Terpenting', references: [{ jenis: 'radiologi', label: 'Radiologi' }] },
    { key: 'hasil_laborat', label: 'Pemeriksaan Penunjang Laboratorium Terpenting', references: [{ jenis: 'laboratorium', label: 'Laboratorium' }] },
    { key: 'tindakan_dan_operasi', label: 'Tindakan / Operasi Selama Perawatan', references: [{ jenis: 'tindakan', label: 'Tindakan' }] },
    { key: 'obat_di_rs', label: 'Obat-obatan Selama Perawatan', references: [{ jenis: 'obat', label: 'Obat RS' }] },
  ]
  
  const catatanPulang = [
    { key: 'diet', label: 'Diet', references: [{ jenis: 'diet', label: 'Diet' }] },
    { key: 'lab_belum', label: 'Hasil Laboratorium Belum Selesai', references: [{ jenis: 'lab_pending', label: 'Lab Pending' }] },
    { key: 'edukasi', label: 'Instruksi / Anjuran dan Edukasi' },
    { key: 'obat_pulang', label: 'Obat Pulang', references: [{ jenis: 'obat_pulang', label: 'Obat Pulang' }] },
  ]
  
  const referensiMeta = {
    keluhan: { judul: 'Ambil Keluhan Pasien', target: 'keluhan_utama', separator: ', ' },
    pemeriksaan: { judul: 'Ambil Pemeriksaan Fisik', target: 'pemeriksaan_fisik', separator: ', ' },
    radiologi: { judul: 'Ambil Hasil Radiologi', target: 'pemeriksaan_penunjang', separator: ', ' },
    laboratorium: { judul: 'Ambil Hasil Laboratorium', target: 'hasil_laborat', separator: ', ' },
    tindakan: { judul: 'Ambil Tindakan / Operasi', target: 'tindakan_dan_operasi', separator: ', ' },
    obat: { judul: 'Ambil Obat Selama RS', target: 'obat_di_rs', separator: ', ' },
    diet: { judul: 'Ambil Diet Pasien', target: 'diet', separator: ', ' },
    lab_pending: { judul: 'Ambil Lab Pending', target: 'lab_belum', separator: ', ' },
    obat_pulang: { judul: 'Ambil Obat Pulang', target: 'obat_pulang', separator: '\n' },
  }
  
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
  
  function sekarang() {
    const date = new Date(Date.now() - new Date().getTimezoneOffset() * 60000)
    return date.toISOString().slice(0, 16)
  }
  
  function formKosong() {
    return {
      no_rawat: props.patient?.no_rawat || '', kode_dokter: '', diagnosa_awal: '', alasan: '',
      keluhan_utama: '', pemeriksaan_fisik: '', jalannya_penyakit: '', pemeriksaan_penunjang: '',
      hasil_laborat: '', tindakan_dan_operasi: '', obat_di_rs: '', diagnosa_utama: '',
      kd_diagnosa_utama: '', diagnosa_sekunder: '', kd_diagnosa_sekunder: '', diagnosa_sekunder2: '',
      kd_diagnosa_sekunder2: '', diagnosa_sekunder3: '', kd_diagnosa_sekunder3: '', diagnosa_sekunder4: '',
      kd_diagnosa_sekunder4: '', prosedur_utama: '', kd_prosedur_utama: '', prosedur_sekunder: '',
      kd_prosedur_sekunder: '', prosedur_sekunder2: '', kd_prosedur_sekunder2: '', prosedur_sekunder3: '',
      kd_prosedur_sekunder3: '', alergi: '', diet: '', lab_belum: '', edukasi: '',
      cara_keluar: 'Atas Izin Dokter', ket_keluar: '', keadaan: 'Membaik', ket_keadaan: '',
      dilanjutkan: 'Kembali Ke RS', ket_dilanjutkan: '', kontrol: sekarang(), obat_pulang: '',
    }
  }
  
  function dariServer(value = {}) {
  	resetStatusCoding()
    Object.assign(form, formKosong(), value, {
      no_rawat: props.patient.no_rawat,
      kontrol: String(value.kontrol || sekarang()).replace(' ', 'T').slice(0, 16),
    })
  }
  
  function resetStatusCoding() {
    Object.keys(statusCoding).forEach((key) => delete statusCoding[key])
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
      const data = await validasiCodingResumePasienRanap(props.token, {
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
      statusCoding[item.kode] = {
        valid: true,
        hint: data.pesan || 'Kode valid',
        im: data.im === '1',
      }
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
    return cariCodingResumePasienRanap(props.token, {
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
    statusCoding[item.kode] = {
      valid: true,
      hint: pilihanCoding.pesan || 'Kode valid',
      im: pilihanCoding.im === '1',
    }
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
      referensiItems.value = await referensiResumePasienRanap(props.token, {
        no_rawat: props.patient.no_rawat,
        jenis: referensiTerbuka.value.jenis,
        q: referensiSearch.value,
      })
      referensiDipilih.value = []
    } catch (error) {
      referensiItems.value = []
      referensiDipilih.value = []
      notifikasi.gagal(error.message || 'Referensi resume pasien rawat inap tidak dapat dibaca.')
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
    const separator = referensiTerbuka.value?.separator || ', '
    const isi = referensiDipilih.value.map((item) => String(item?.isi || '').trim()).filter(Boolean).join(separator)
    if (!target || !isi) return
    const sebelumnya = String(form[target] || '').trim()
    form[target] = sebelumnya ? `${sebelumnya}${separator}${isi}` : isi
    referensiTerbuka.value = null
    referensiDipilih.value = []
  }
  
  async function muat() {
    if (!props.patient.no_rawat) return
    loading.value = true
    try {
      const data = await resumePasienRanapData(props.token, props.patient.no_rawat)
      tersedia.value = Boolean(data?.tersedia)
      billingLocked.value = Boolean(data?.billing_terkunci)
      Object.assign(pilihan, data?.pilihan || {})
      dariServer(data?.resume)
      doctor.value = data?.dokter || {}
    } catch (error) {
      notifikasi.gagal(error.message || 'Resume pasien rawat inap tidak dapat dibaca.')
    } finally {
      loading.value = false
    }
  }
  
  function payload() {
    return {
      ...form,
      no_rawat: props.patient.no_rawat,
      kode_dokter: doctor.value?.kode || '',
      kontrol: String(form.kontrol || '').replace('T', ' '),
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
        ? await ubahResumePasienRanap(props.token, payload())
        : await simpanResumePasienRanap(props.token, payload())
      notifikasi.sukses(response?.pesan || 'Resume pasien rawat inap berhasil disimpan.')
      await muat()
    } catch (error) {
      notifikasi.gagal(error.message || 'Resume pasien rawat inap gagal disimpan.')
    } finally {
      saving.value = false
    }
  }
  
  async function hapus() {
    deleting.value = true
    try {
      const response = await hapusResumePasienRanap(props.token, props.patient.no_rawat)
      notifikasi.sukses(response?.pesan || 'Resume pasien rawat inap berhasil dihapus.')
      hapusTerbuka.value = false
      await muat()
    } catch (error) {
      notifikasi.gagal(error.message || 'Resume pasien rawat inap gagal dihapus.')
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
    referensiSemuaTerpilih,
    pilihanSelect,
    catatanPerawatan,
    catatanPulang,
    diagnosa,
    prosedur,
    nilaiCoding,
    cariPilihanCoding,
    pilihCoding,
    bukaReferensi,
    muatReferensi,
    referensiKey,
    referensiAktif,
    toggleReferensi,
    pilihSemuaReferensi,
    tambahReferensiTerpilih,
    simpan,
    hapus,
  }
}
