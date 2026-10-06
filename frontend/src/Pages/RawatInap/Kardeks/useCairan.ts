import { computed, onBeforeUnmount, reactive, ref, watch } from 'vue'
import type { PropsKardeks, HasilKardeks } from '../../../types/kardeks'
import { request } from '../../../lib/shared/http'
import { useNotifikasi } from '../../../lib/shared/useNotifikasi'

export function rekapCairan(hasil: HasilKardeks) {
  const bagian = hasil.bagian.find(b => b.kode === 'cairan')
  if (!bagian || bagian.error) return null
  const baris = bagian.baris.filter(r => r.waktu_mulai >= hasil.mulai && r.waktu_selesai <= hasil.selesai)
  const jumlah = (jenis: string) => {
    const data = baris.filter(r => r.jenis === jenis)
    return data.length ? data.reduce((n, r) => n + Math.round(Number(r.volume_ml) * 100), 0) / 100 : null
  }
  const masuk = jumlah('Masuk')
  const keluar = jumlah('Keluar')
  return {
    masuk,
    keluar,
    balance: masuk !== null && keluar !== null ? Math.round((masuk - keluar) * 100) / 100 : null,
    lintas: bagian.baris.length - baris.length,
  }
}

export function useCairan(props: PropsKardeks, berubah: () => void) {
  const form = reactive({ waktu_mulai: '', waktu_selesai: '', jenis: 'Masuk', kategori: 'Infus', rincian: '', volume_ml: '', catatan: '' })
  const saving = ref(false)
  const error = ref('')
  const notifikasi = useNotifikasi()
  let konteks = 0
  const kategori = computed(() => (form.jenis === 'Masuk'
    ? ['Infus', 'Transfusi', 'Oral', 'NGT', 'Lainnya']
    : ['Urine', 'Drain', 'Cairan Lambung', 'Lainnya']).map(value => ({ label: value, value })))
  watch(() => form.jenis, () => { form.kategori = kategori.value[0].value })
  watch(() => [props.token, props.patient.no_rawat], () => {
    konteks++
    saving.value = false
    error.value = ''
    Object.assign(form, { waktu_mulai: '', waktu_selesai: '', jenis: 'Masuk', kategori: 'Infus', rincian: '', volume_ml: '', catatan: '' })
  })
  onBeforeUnmount(() => { konteks++ })
  async function simpan() {
    if (saving.value) return
    error.value = ''
    const volume = String(form.volume_ml).trim().replace(',', '.')
    if (!/^\d+(\.\d{1,2})?$/.test(volume) || !form.waktu_mulai || form.waktu_selesai <= form.waktu_mulai) {
      error.value = 'Isi volume nonnegatif (maksimal dua desimal), serta waktu selesai setelah mulai.'
      return
    }
    const id = konteks
    const waktu = (s: string) => s.replace('T', ' ') + (s.length === 16 ? ':00' : '')
    saving.value = true
    try {
      await request('/api/kardeks/cairan', {
        method: 'POST',
        headers: { Authorization: 'Bearer ' + props.token },
        body: JSON.stringify({ ...form, no_rawat: props.patient.no_rawat, volume_ml: Number(volume), waktu_mulai: waktu(form.waktu_mulai), waktu_selesai: waktu(form.waktu_selesai) }),
      })
      if (id !== konteks) return
      form.volume_ml = ''
      form.catatan = ''
      notifikasi.sukses('Catatan cairan berhasil disimpan ke SIMRS.')
      berubah()
    } catch (e) {
      if (id === konteks) error.value = e instanceof Error ? e.message : 'Penyimpanan gagal'
    } finally {
      if (id === konteks) saving.value = false
    }
  }
  return { form, saving, error, kategori, simpan }
}
