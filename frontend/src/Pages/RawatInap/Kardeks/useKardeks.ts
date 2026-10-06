import { computed, onBeforeUnmount, ref, watch } from 'vue'
import { request } from '../../../lib/shared/http'
import { useNotifikasi } from '../../../lib/shared/useNotifikasi'
import type { PropsKardeks, HasilKardeks } from '../../../types/kardeks'
import { parameterKardeks } from '../../../types/kardeks'
import { cetakKardeks } from './cetakKardeks'
import type { KopPemulangan } from '../../../types/perencanaanPemulangan'

export function useKardeks(props: PropsKardeks) {
  const loading = ref(false)
  const printing = ref(false)
  const inputVisible = ref(false)
  const error = ref('')
  const tanggal = ref('')
  const jam = ref('08:00')
  const hasil = ref<HasilKardeks | null>(null)
  const tab = ref('observasi')
  const sumber = ref('observasi')
  const grafikVisible = ref(false)
  const metrik = ref('nadi')
  const notifikasi = useNotifikasi()
  let generasi = 0
  let urutan = 0
  const header = () => ({ Authorization: 'Bearer ' + props.token })
  const observasi = computed(() => hasil.value?.bagian.find(b => b.kode === sumber.value))
  const perluMuat = computed(() => hasil.value?.mulai !== tanggal.value + ' ' + jam.value + ':00')
  const adaError = computed(() => !!hasil.value?.bagian.some(b => b.error))
  const waktu = (s: string) => Date.parse(s.replace(' ', 'T') + '+08:00')
  const slots = computed(() => {
    if (!hasil.value) return []
    const awal = waktu(hasil.value.mulai)
    return Array.from({ length: 24 }, (_, i) => {
      const mulai = awal + i * 3600000
      const label = new Date(mulai + 8 * 3600000).toISOString()
      return {
        label: label.slice(11, 16),
        tanggal: label.slice(0, 10),
        baris: (observasi.value?.baris || []).filter(r => waktu(r.waktu) >= mulai && waktu(r.waktu) < mulai + 3600000),
      }
    })
  })
  const bagianAktif = computed(() => {
    const grup: Record<string, string[]> = {
      terapi: ['resep', 'obat'],
      instruksi: ['pemeriksaan', 'instruksi', 'catatan'],
      penunjang: ['lab', 'radiologi'],
      ventilator: ['setting_ventilator', 'monitoring_ventilator'],
      cairan: ['cairan'],
    }
    return hasil.value?.bagian.filter(b => (grup[tab.value] || []).includes(b.kode)) || []
  })
  const judulGrafik = computed(() => parameterKardeks.find(p => p.key === metrik.value))
  const grafik = computed(() => {
    const raw = (observasi.value?.baris || []).flatMap(r => {
      const teks = (r[metrik.value] || '').trim().replace(',', '.')
      if (!/^\d+(\.\d+)?$/.test(teks)) return []
      const nilai = Number(teks)
      if (!Number.isFinite(nilai)) return []
      return [{ nilai, waktu: r.waktu }]
    })
    const min = raw.length ? Math.min(...raw.map(r => r.nilai)) : 0
    const max = raw.length ? Math.max(...raw.map(r => r.nilai)) : 0
    const rentang = max - min || 1
    const awal = hasil.value ? waktu(hasil.value.mulai) : 0
    return {
      min, max,
      titik: raw.map(r => ({
        ...r,
        x: 55 + (waktu(r.waktu) - awal) / 86400000 * 690,
        y: max === min ? 105 : 175 - (r.nilai - min) / rentang * 140,
      })),
    }
  })
  async function muat() {
    const id = ++urutan
    const konteks = generasi
    loading.value = true
    hasil.value = null
    error.value = ''
    try {
      const params = new URLSearchParams({ no_rawat: props.patient.no_rawat, tanggal: tanggal.value, jam: jam.value })
      const data = await request<HasilKardeks>('/api/kardeks?' + params, { headers: header() })
      if (id === urutan && konteks === generasi) hasil.value = data
    } catch (e) {
      if (id === urutan && konteks === generasi) error.value = e instanceof Error ? e.message : 'Kardeks gagal dimuat'
    } finally {
      if (id === urutan && konteks === generasi) loading.value = false
    }
  }
  async function cetak() {
    if (!hasil.value || loading.value || printing.value || perluMuat.value || adaError.value) return
    const win = window.open('', '_blank', 'width=1200,height=800')
    if (!win) { notifikasi.peringatan('Izinkan jendela cetak pada browser.'); return }
    const snapshot = hasil.value
    const pasien = { ...props.patient }
    const matrix = slots.value
    const namaSumber = observasi.value?.nama || ''
    const chart = { ...grafik.value, label: (judulGrafik.value?.label || '') + ' (' + (judulGrafik.value?.unit || '') + ')' }
    const konteks = generasi
    printing.value = true
    win.document.body.textContent = 'Menyiapkan kardeks...'
    try {
      const kop = await request<KopPemulangan>('/api/perencanaan-pemulangan/kop', { headers: header() })
      if (win.closed) return
      if (konteks !== generasi) { win.close(); return }
      await cetakKardeks(win, pasien, snapshot, matrix, namaSumber, kop, chart)
    } catch (e) {
      if (!win.closed) win.close()
      if (konteks === generasi) notifikasi.peringatan(e instanceof Error ? e.message : 'Cetak gagal')
    } finally {
      if (konteks === generasi) printing.value = false
    }
  }
  watch(() => [props.token, props.patient.no_rawat], () => {
    generasi++
    printing.value = false
    inputVisible.value = false
    tanggal.value = new Date(Date.now() + 8 * 3600000).toISOString().slice(0, 10)
    jam.value = '08:00'
    tab.value = 'observasi'
    sumber.value = 'observasi'
    grafikVisible.value = false
    void muat()
  }, { immediate: true })
  onBeforeUnmount(() => { generasi++ })
  function observasiBerubah() {
    sumber.value = 'observasi'
    tab.value = 'observasi'
    void muat()
  }
  return { loading, printing, inputVisible, observasiBerubah, error, tanggal, jam, hasil, tab, sumber, grafikVisible, metrik, perluMuat,
    adaError, observasi, slots, bagianAktif, judulGrafik, grafik, muat, cetak }
}
