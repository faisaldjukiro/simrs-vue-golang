import { onMounted, ref, watch } from 'vue'

export function useParafEdukasi(
  props: { modelValue: string; disabled?: boolean },
  ubah: (nilai: string) => void,
) {
  const kanvas = ref<HTMLCanvasElement | null>(null)
  const pesan = ref('')
  let pointer: number | null = null
  let terakhir = { x: 0, y: 0 }
  let jarak = 0
  let dikirim = ''
  let generasi = 0
  function tampilkan() {
    const id = ++generasi
    pointer = null
    const ctx = kanvas.value?.getContext('2d')
    if (!ctx) return
    ctx.fillStyle = '#fff'
    ctx.fillRect(0, 0, 720, 240)
    if (props.modelValue.startsWith('data:image/png;base64,')) {
      const gambar = new Image()
      gambar.onload = () => { if (id === generasi) ctx.drawImage(gambar, 0, 0, 720, 240) }
      gambar.onerror = () => { if (id === generasi) pesan.value = 'Paraf tidak dapat dimuat. Gambar ulang paraf.' }
      gambar.src = props.modelValue
    }
  }
  function titik(event: PointerEvent) {
    const area = kanvas.value!.getBoundingClientRect()
    return {
      x: Math.max(0, Math.min(720, (event.clientX - area.left) * 720 / area.width)),
      y: Math.max(0, Math.min(240, (event.clientY - area.top) * 240 / area.height)),
    }
  }
  function mulai(event: PointerEvent) {
    if (props.disabled || pointer !== null || event.button !== 0) return
    const ctx = kanvas.value?.getContext('2d')
    if (!ctx) { pesan.value = 'Browser tidak mendukung kotak paraf.'; return }
    generasi++
    pointer = event.pointerId
    kanvas.value!.setPointerCapture(pointer)
    terakhir = titik(event)
    jarak = 0
    pesan.value = ''
  }
  function gerak(event: PointerEvent) {
    if (props.disabled || pointer !== event.pointerId) return
    const ctx = kanvas.value!.getContext('2d')!
    const baru = titik(event)
    ctx.strokeStyle = '#111'
    ctx.lineWidth = 3
    ctx.lineCap = 'round'
    ctx.lineJoin = 'round'
    ctx.beginPath()
    ctx.moveTo(terakhir.x, terakhir.y)
    ctx.lineTo(baru.x, baru.y)
    ctx.stroke()
    jarak += Math.hypot(baru.x - terakhir.x, baru.y - terakhir.y)
    terakhir = baru
  }
  function selesai(event: PointerEvent) {
    if (pointer !== event.pointerId) return
    if (props.disabled) { tampilkan(); return }
    gerak(event)
    pointer = null
    if (jarak < 8) { tampilkan(); return }
    dikirim = kanvas.value!.toDataURL('image/png')
    if (dikirim.length > 256 * 1024) { pesan.value = 'Paraf terlalu besar. Gambar ulang paraf.'; tampilkan(); return }
    ubah(dikirim)
  }
  function hapus() {
    if (props.disabled) return
    dikirim = ''
    ubah('')
    pesan.value = ''
    tampilkan()
  }
  onMounted(tampilkan)
  watch(() => props.modelValue, nilai => {
    if (!nilai) dikirim = ''
    if (!nilai || nilai !== dikirim) tampilkan()
  }, { flush: 'post' })
  watch(() => props.disabled, () => { if (pointer !== null) tampilkan() })
  function batal() { if (pointer !== null) tampilkan() }
  return { kanvas, pesan, mulai, gerak, selesai, batal, hapus }
}
