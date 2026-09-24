import { computed, onBeforeUnmount, reactive, ref, watch } from 'vue'
import { request } from '../../../lib/shared/http'
import { useNotifikasi } from '../../../lib/shared/useNotifikasi'
import { bidangChecklist, type CatatanChecklist, type PropsChecklist } from '../../../types/checklistPreOperasi'

export function useChecklistPreOperasi(props: PropsChecklist) {
  const notifikasi = useNotifikasi()
  const loading = ref(false)
  const saving = ref(false)
  const error = ref('')
  const errorSimpan = ref('')
  const peringatan = ref('')
  const formVisible = ref(true)
  const keyword = ref('')
  const records = ref<CatatanChecklist[]>([])
  const editing = ref<CatatanChecklist | null>(null)
  const detail = ref<CatatanChecklist | null>(null)
  const hapusTarget = ref<CatatanChecklist | null>(null)
  const tanggal = ref('')
  const form = reactive<Record<string, string>>({})
  const pilihan = reactive<Record<string, Record<string, string>>>({})
  let generasi = 0
  let urutanMuat = 0

  const nomorRawat = computed(() => String(props.patient.no_rawat || ''))
  const filteredRows = computed(() => {
    const q = keyword.value.toLocaleLowerCase().trim()
    return records.value
      .filter(r => [r.tanggal, r.sumber, ...Object.values(r.data)].join(' ').toLocaleLowerCase().includes(q))
      .map(r => ({ ...r, kunci: r.sumber + ':' + (r.id || r.tanggal) }))
  })
  function api<T>(path = '', options: RequestInit = {}) {
    return request<T>('/api/checklist-pre-operasi' + path, {
      ...options,
      headers: { Authorization: 'Bearer ' + props.token },
    })
  }
  function waktuSekarang() {
    // Jam klinis WITA, tidak bergantung pada zona waktu komputer pengguna.
    tanggal.value = new Date(Date.now() + 8 * 60 * 60 * 1000).toISOString().slice(0, 19)
  }
  function reset() {
    editing.value = null
    errorSimpan.value = ''
    for (const b of bidangChecklist) {
      form[b.kode] = ''
      if (b.jenis) pilihan[b.kode] = {}
    }
    waktuSekarang()
    const jadwal = props.jadwalOperasi
    if (jadwal && jadwal.no_rawat === nomorRawat.value) {
      form.tindakan = jadwal.nama_paket
      pilihan.kd_dokter_bedah = { kode: jadwal.kd_dokter, nama: jadwal.nama_dokter }
      // Tanggal pemeriksaan tetap waktu pencatatan, bukan waktu operasi
      // yang mungkin masih di masa depan. Jawaban klinis tidak diisi otomatis.
    }
  }
  async function muat() {
    const konteks = generasi
    const urutan = ++urutanMuat
    if (!nomorRawat.value) {
      error.value = 'Pilih kunjungan pasien terlebih dahulu.'
      return
    }
    loading.value = true
    error.value = ''
    try {
      const hasil = await api<{ catatan: CatatanChecklist[]; peringatan: string }>(
        '?no_rawat=' + encodeURIComponent(nomorRawat.value),
      )
      if (konteks !== generasi || urutan !== urutanMuat) return
      records.value = hasil.catatan
      peringatan.value = hasil.peringatan
    } catch (e) {
      if (konteks === generasi && urutan === urutanMuat) error.value = pesan(e)
    } finally {
      if (konteks === generasi && urutan === urutanMuat) loading.value = false
    }
  }
  async function cariReferensi(jenis: string, q: string) {
    return api<Record<string, string>[]>('/referensi?' + new URLSearchParams({ jenis, q }))
  }
  function edit(r: CatatanChecklist) {
    if (!r.bisa_ubah || saving.value) return
    reset()
    editing.value = r
    tanggal.value = r.tanggal.replace(' ', 'T')
    for (const b of bidangChecklist) {
      form[b.kode] = r.data[b.kode] || ''
      if (b.jenis) pilihan[b.kode] = { kode: r.data[b.kode], nama: r.data[b.kode + '_nama'] }
    }
    formVisible.value = true
  }
  async function simpan() {
    if (saving.value) return
    errorSimpan.value = ''
    const data: Record<string, string> = {}
    for (const b of bidangChecklist) {
      data[b.kode] = (b.jenis ? pilihan[b.kode]?.kode : form[b.kode]) || ''
      if (!b.kode.startsWith('keterangan_') && !data[b.kode].trim()) {
        errorSimpan.value = b.label + ' wajib diisi.'
        return
      }
    }
    if (!tanggal.value || !nomorRawat.value) {
      errorSimpan.value = 'Tanggal dan kunjungan pasien wajib diisi.'
      return
    }
    let waktu = tanggal.value.replace('T', ' ')
    if (waktu.length === 16) waktu += ':00'
    const konteks = generasi
    saving.value = true
    try {
      const hasil = await api<{ pesan: string }>(editing.value ? '/' + (editing.value.sumber === 'Khanza' ? 'khanza' : editing.value.id) : '', {
        method: editing.value ? 'PUT' : 'POST',
        body: JSON.stringify({
          asli: editing.value?.sumber === 'Khanza' ? editing.value : undefined,
          no_rawat: nomorRawat.value,
          tanggal: waktu,
          versi: editing.value?.versi || 0,
          data,
        }),
      })
      if (konteks !== generasi) return
      notifikasi.sukses(hasil.pesan)
      reset()
      await muat()
    } catch (e) {
      if (konteks === generasi) errorSimpan.value = pesan(e)
    } finally {
      if (konteks === generasi) saving.value = false
    }
  }
  async function hapus() {
    const r = hapusTarget.value
    if (!r || !r.bisa_ubah || saving.value) return
    const konteks = generasi
    saving.value = true
    try {
      const hasil = await api<{ pesan: string }>('/' + (r.sumber === 'Khanza' ? 'khanza' : r.id), {
        method: 'DELETE',
        body: JSON.stringify({ no_rawat: r.no_rawat, versi: r.versi, asli: r.sumber === 'Khanza' ? r : undefined }),
      })
      if (konteks !== generasi) return
      hapusTarget.value = null
      if (editing.value?.id === r.id) reset()
      notifikasi.sukses(hasil.pesan)
      await muat()
    } catch (e) {
      if (konteks === generasi) notifikasi.gagal(pesan(e))
    } finally {
      if (konteks === generasi) saving.value = false
    }
  }
  function pesan(e: unknown) {
    return e instanceof Error ? e.message : 'Data belum dapat diproses.'
  }
  function cetak() {
    const r = detail.value
    if (!r) return
    const jendela = window.open('', '_blank', 'width=900,height=750')
    if (!jendela) {
      notifikasi.peringatan('Izinkan popup untuk mencetak checklist.')
      return
    }
    jendela.opener = null
    const doc = jendela.document
    doc.title = 'Checklist Pre Operasi'
    const style = doc.createElement('style')
    style.textContent = 'body{font:13px Arial,sans-serif;color:#111;margin:24px}table{width:100%;border-collapse:collapse}td{padding:8px;border:1px solid #bbb;overflow-wrap:anywhere}td:first-child{width:38%}tr{break-inside:avoid}h1{font-size:20px}'
    doc.head.append(style)
    const judul = doc.createElement('h1')
    judul.textContent = 'Checklist Pre Operasi'
    const identitas = doc.createElement('p')
    identitas.textContent = [
      props.patient.nama_pasien || props.patient.nm_pasien || '-',
      'RM: ' + (props.patient.no_rekam_medis || props.patient.no_rkm_medis || '-'),
      r.no_rawat, r.tanggal + ' WITA', 'Sumber: ' + r.sumber,
    ].join(' · ')
    const tabel = doc.createElement('table')
    for (const b of bidangChecklist) {
      const row = tabel.insertRow()
      row.insertCell().textContent = b.label
      row.insertCell().textContent = r.data[b.kode + '_nama'] || r.data[b.kode] || '-'
    }
    doc.body.append(judul, identitas, tabel)
    jendela.focus()
    jendela.print()
  }
  watch(() => [props.token, nomorRawat.value], () => {
    generasi++
    loading.value = false
    saving.value = false
    records.value = []
    peringatan.value = ''
    keyword.value = ''
    detail.value = null
    hapusTarget.value = null
    reset()
    void muat()
  }, { immediate: true })
  onBeforeUnmount(() => { generasi++ })

  return {
    loading, saving, error, errorSimpan, peringatan, formVisible, keyword, filteredRows,
    editing, detail, hapusTarget, tanggal, form, pilihan,
    waktuSekarang, reset, muat, cariReferensi, edit, simpan, hapus, cetak,
  }
}
