import { computed, onBeforeUnmount, onMounted, reactive, ref } from "vue"
import { buatPerangkatWhatsApp, daftarPerangkatWhatsApp, daftarPesanWhatsApp, detailPerangkatWhatsApp, hubungkanPerangkatWhatsApp, kirimPesanWhatsApp, kodePasanganWhatsApp, qrPerangkatWhatsApp, statusWhatsAppGateway } from "../../../lib/faisal/api"
import { useNotifikasi } from "../../../lib/shared/useNotifikasi"

export function useWhatsAppGateway(props) {
  const notifikasi = useNotifikasi()
  
  const loading = ref(true)
  const proses = ref('')
  const error = ref('')
  const konfigurasi = ref({ dikonfigurasi: false })
  const responsPerangkat = ref([])
  const responsPesan = ref([])
  const modalQR = ref(false)
  const modalPair = ref(false)
  const qrURL = ref('')
  const perangkatAktif = ref(null)
  const kodePairing = ref('')
  const formPerangkat = reactive({ nama: '' })
  const formPair = reactive({ nomor: '' })
  const formPesan = reactive({ device_id: '', tujuan: '', nama_penerima: '', external_id: '', idempotency_key: '', pesan: '' })
  
  function ambilArray(data, candidates = []) {
    if (Array.isArray(data)) return data
    for (const key of candidates) if (Array.isArray(data?.[key])) return data[key]
    if (Array.isArray(data?.items)) return data.items
    return []
  }
  
  const perangkat = computed(() => ambilArray(responsPerangkat.value, ['devices', 'perangkat']))
  const pesan = computed(() => ambilArray(responsPesan.value, ['messages', 'pesan']))
  const jumlahTerhubung = computed(() => perangkat.value.filter((item) => ['connected', 'online'].includes(String(item?.status || item?.connection_status).toLowerCase())).length)
  const pilihanPerangkat = computed(() => perangkat.value.map((item) => ({
    label: `${item.name || item.nama || 'Perangkat'}${item.phone ? ` - ${item.phone}` : ''}`,
    value: item.id || item.device_id,
  })).filter((item) => item.value))
  
  function idPerangkat(item) { return item?.id || item?.device_id || '' }
  function warnaStatus(status) { return ['connected', 'sent', 'delivered', 'read'].includes(String(status).toLowerCase()) ? 'success' : ['failed', 'disconnected'].includes(String(status).toLowerCase()) ? 'danger' : 'pending' }
  function warnaStatusPerangkat(item) { return warnaStatus(item?.status || item?.connection_status) }
  function labelStatus(status) {
    return ({ connected: 'Terhubung', online: 'Terhubung', disconnected: 'Terputus', queued: 'Antrean', processing: 'Diproses', sent: 'Terkirim', delivered: 'Diterima', read: 'Dibaca', failed: 'Gagal' })[String(status || '').toLowerCase()] || String(status || '-').replaceAll('_', ' ')
  }
  function formatWaktu(value) {
    if (!value) return '-'
    const tanggal = new Date(value)
    return Number.isNaN(tanggal.getTime()) ? value : tanggal.toLocaleString('id-ID', { dateStyle: 'medium', timeStyle: 'short' })
  }
  function idempotencyBaru() { return `sirapi-${Date.now()}-${Math.random().toString(36).slice(2, 8)}` }
  function tunggu(ms) { return new Promise((resolve) => window.setTimeout(resolve, ms)) }
  function qrBelumSiap(err) { return /qr.+not ready|connect the device|poll this endpoint/i.test(String(err?.message || '')) }
  
  async function muatData() {
    loading.value = true
    error.value = ''
    try {
      konfigurasi.value = await statusWhatsAppGateway(props.token)
      if (!konfigurasi.value.dikonfigurasi) {
        error.value = 'URL_WHATSAPP dan KEY_WHATSAPP belum diisi pada env backend.'
        return
      }
      const [dataPerangkat, dataPesan] = await Promise.all([
        daftarPerangkatWhatsApp(props.token),
        daftarPesanWhatsApp(props.token).catch(() => []),
      ])
      responsPerangkat.value = dataPerangkat
      responsPesan.value = dataPesan
      if (!formPesan.device_id && pilihanPerangkat.value.length) formPesan.device_id = pilihanPerangkat.value[0].value
    } catch (err) {
      error.value = err.message || 'WhatsApp Gateway tidak dapat dibaca.'
    } finally {
      loading.value = false
    }
  }
  
  async function buatPerangkat() {
    if (proses.value) return
    if (!formPerangkat.nama.trim()) {
      notifikasi.peringatan('Nama perangkat WhatsApp wajib diisi.')
      return
    }
    proses.value = 'buat'
    try {
      await buatPerangkatWhatsApp(props.token, formPerangkat.nama.trim())
      formPerangkat.nama = ''
      notifikasi.sukses('Perangkat WhatsApp berhasil dibuat.')
      await muatData()
    } catch (err) { notifikasi.gagal(err.message) } finally { proses.value = '' }
  }
  
  async function refreshPerangkat(item) {
    proses.value = `status-${idPerangkat(item)}`
    try {
      await detailPerangkatWhatsApp(props.token, idPerangkat(item))
      await muatData()
    } catch (err) { notifikasi.gagal(err.message) } finally { proses.value = '' }
  }
  
  async function hubungkan(item) {
    proses.value = `connect-${idPerangkat(item)}`
    try {
      const hasil = await hubungkanPerangkatWhatsApp(props.token, idPerangkat(item))
      notifikasi.info(hasil?.already_connected ? 'Perangkat sudah terhubung.' : 'Perangkat sedang dihubungkan. SIRAPI menunggu QR dari gateway...')
      await muatData()
      if (!hasil?.already_connected) await bukaQR(item)
    } catch (err) { notifikasi.gagal(err.message) } finally { proses.value = '' }
  }
  
  async function bukaQR(item) {
    perangkatAktif.value = item
    proses.value = `qr-${idPerangkat(item)}`
    try {
      let errorTerakhir = null
      for (let percobaan = 1; percobaan <= 20; percobaan += 1) {
        try {
          const gambar = await qrPerangkatWhatsApp(props.token, idPerangkat(item))
          if (qrURL.value) URL.revokeObjectURL(qrURL.value)
          qrURL.value = URL.createObjectURL(gambar)
          modalQR.value = true
          return
        } catch (err) {
          errorTerakhir = err
          if (!qrBelumSiap(err)) throw err
          if (percobaan < 20) await tunggu(1500)
        }
      }
      throw new Error(qrBelumSiap(errorTerakhir)
        ? 'QR belum tersedia setelah 30 detik. Tekan Hubungkan lalu coba buka QR kembali.'
        : errorTerakhir?.message || 'QR perangkat tidak dapat dibaca.')
    } catch (err) { notifikasi.peringatan(err.message || 'QR belum siap. Hubungkan perangkat lalu coba lagi.') } finally { proses.value = '' }
  }
  
  function bukaPair(item) {
    perangkatAktif.value = item
    formPair.nomor = ''
    kodePairing.value = ''
    modalPair.value = true
  }
  
  async function mintaKodePairing() {
    if (!formPair.nomor.trim() || !perangkatAktif.value || proses.value) return
    proses.value = 'pair'
    try {
      const hasil = await kodePasanganWhatsApp(props.token, idPerangkat(perangkatAktif.value), formPair.nomor)
      kodePairing.value = hasil?.code || hasil?.pairing_code || hasil?.pair_code || ''
      notifikasi.sukses('Kode pasangan berhasil dibuat.')
    } catch (err) { notifikasi.gagal(err.message) } finally { proses.value = '' }
  }
  
  async function kirimPesan() {
    if (proses.value) return
    if (!formPesan.device_id || !formPesan.tujuan.trim() || !formPesan.pesan.trim()) {
      notifikasi.peringatan('Perangkat, nomor tujuan, dan isi pesan wajib diisi.')
      return
    }
    proses.value = 'kirim'
    try {
      if (!formPesan.idempotency_key) formPesan.idempotency_key = idempotencyBaru()
      await kirimPesanWhatsApp(props.token, { ...formPesan })
      notifikasi.sukses('Pesan diterima antrean WhatsApp Gateway. Status 202 belum berarti pesan sudah terkirim.')
      Object.assign(formPesan, { tujuan: '', nama_penerima: '', external_id: '', idempotency_key: idempotencyBaru(), pesan: '' })
      responsPesan.value = await daftarPesanWhatsApp(props.token).catch(() => responsPesan.value)
    } catch (err) { notifikasi.gagal(err.message) } finally { proses.value = '' }
  }
  
  onMounted(() => { formPesan.idempotency_key = idempotencyBaru(); muatData() })
  onBeforeUnmount(() => { if (qrURL.value) URL.revokeObjectURL(qrURL.value) })
  return {
    loading,
    proses,
    error,
    konfigurasi,
    modalQR,
    modalPair,
    qrURL,
    perangkatAktif,
    kodePairing,
    formPerangkat,
    formPair,
    formPesan,
    perangkat,
    pesan,
    jumlahTerhubung,
    pilihanPerangkat,
    idPerangkat,
    warnaStatus,
    warnaStatusPerangkat,
    labelStatus,
    formatWaktu,
    muatData,
    buatPerangkat,
    refreshPerangkat,
    hubungkan,
    bukaQR,
    bukaPair,
    mintaKodePairing,
    kirimPesan,
  }
}
