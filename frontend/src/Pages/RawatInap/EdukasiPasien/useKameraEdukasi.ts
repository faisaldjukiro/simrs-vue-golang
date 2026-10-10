import { nextTick, onBeforeUnmount, ref, watch } from 'vue'

export function useKameraEdukasi(
  props: { aktif: boolean; disabled?: boolean; label: string },
  tangkap: (file: File) => void, tutup: () => void,
) {
  const video = ref<HTMLVideoElement | null>(null)
  const memuat = ref(false)
  const mengambil = ref(false)
  const siap = ref(false)
  const pesan = ref('')
  const arah = ref<'user' | 'environment'>('user')
  let stream: MediaStream | null = null
  let generasi = 0
  function hentikan() {
    generasi++
    stream?.getTracks().forEach(track => track.stop())
    stream = null
    if (video.value) video.value.srcObject = null
    siap.value = false
    memuat.value = false
    mengambil.value = false
  }
  function selesai() { hentikan(); tutup() }
  async function buka() {
    hentikan()
    pesan.value = ''
    if (props.disabled || !props.aktif) return
    if (!window.isSecureContext || !navigator.mediaDevices?.getUserMedia) {
      pesan.value = 'Kamera memerlukan HTTPS atau localhost dan browser yang mendukung kamera.'
      return
    }
    const id = generasi
    memuat.value = true
    try {
      const hasil = await navigator.mediaDevices.getUserMedia({ audio: false, video: {
        facingMode: { ideal: arah.value }, width: { ideal: 1280 }, height: { ideal: 960 },
      } })
      if (id !== generasi || !props.aktif || props.disabled) { hasil.getTracks().forEach(t => t.stop()); return }
      stream = hasil
      hasil.getVideoTracks().forEach(track => track.addEventListener('ended', () => {
        if (id === generasi) { hentikan(); pesan.value = 'Kamera terputus. Tutup lalu buka kamera kembali.' }
      }))
      await nextTick()
      if (id !== generasi) return
      if (!video.value) { hentikan(); return }
      video.value.srcObject = hasil
      await video.value.play()
      if (id === generasi) siap.value = video.value.videoWidth > 0
    } catch (e) {
      if (id !== generasi) return
      hentikan()
      const nama = e instanceof Error ? e.name : ''
      pesan.value = nama === 'NotAllowedError' ? 'Izin kamera ditolak. Izinkan kamera pada pengaturan situs browser, lalu buka kembali.'
        : nama === 'NotFoundError' ? 'Kamera tidak ditemukan pada perangkat ini.'
          : 'Kamera tidak dapat dibuka. Pastikan kamera tidak sedang digunakan aplikasi lain.'
    } finally { if (id === generasi) memuat.value = false }
  }
  async function ambil() {
    const el = video.value
    if (!siap.value || props.disabled || mengambil.value || !el?.videoWidth || !el.videoHeight) return
    const id = generasi
    mengambil.value = true
    pesan.value = ''
    try {
      const canvas = document.createElement('canvas')
      const skala = Math.min(1, 1600 / Math.max(el.videoWidth, el.videoHeight))
      canvas.width = Math.round(el.videoWidth * skala)
      canvas.height = Math.round(el.videoHeight * skala)
      const ctx = canvas.getContext('2d')
      if (!ctx) throw new Error('canvas')
      ctx.drawImage(el, 0, 0, canvas.width, canvas.height)
      const blob = await new Promise<Blob | null>(resolve => canvas.toBlob(resolve, 'image/jpeg', 0.9))
      if (id !== generasi || props.disabled || !props.aktif) return
      if (!blob || !blob.size || blob.size > 10 * 1024 * 1024) throw new Error('foto')
      tangkap(new File([blob], 'edukasi-kamera.jpg', { type: 'image/jpeg' }))
      selesai()
    } catch { if (id === generasi) pesan.value = 'Foto gagal diambil. Silakan coba lagi.' }
    finally { if (id === generasi) mengambil.value = false }
  }
  function gantiKamera() { arah.value = arah.value === 'user' ? 'environment' : 'user'; void buka() }
  watch(() => [props.aktif, props.disabled], () => {
    if (props.aktif && !props.disabled) void buka()
    else { hentikan(); if (props.aktif) tutup() }
  }, { immediate: true })
  onBeforeUnmount(hentikan)
  return { video, memuat, mengambil, siap, pesan, selesai, ambil, gantiKamera }
}
