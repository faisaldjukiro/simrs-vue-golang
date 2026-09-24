import { computed, onBeforeUnmount, reactive, ref, watch } from 'vue'
import { request } from '../../../lib/shared/http'
import type { JadwalOperasi } from '../../../types/jadwalOperasi'
import type { HasilPendukung } from '../../../types/pendukungOperasi'
import { useNotifikasi } from '../../../lib/shared/useNotifikasi'

export interface PropsRiwayatPendukung {
  token: string
  patient: Record<string, unknown>
  jadwalOperasi: JadwalOperasi
  jenis: string
  judul: string
}

export function useRiwayatPendukung(props: PropsRiwayatPendukung) {
  const notifikasi = useNotifikasi()
  const loading = ref(false)
  const error = ref('')
  const q = ref('')
  const kosong = (): HasilPendukung => ({ kolom: [], catatan: [], form: [], pesan_kunci: '' })
  const hasil = ref<HasilPendukung>(kosong())
  const form = reactive<Record<string, string>>({})
  const pilihan = reactive<Record<string, Record<string, string>>>({})
  const editing = ref<Record<string, string> | null>(null)
  const hapusTarget = ref<Record<string, string> | null>(null)
  const saving = ref(false)
  const errorSimpan = ref('')
  const formVisible = ref(true)
  const bidang = computed(() => hasil.value.form || [])
  const terkunci = computed(() => saving.value || loading.value || !!error.value || !!hasil.value.pesan_kunci)
  const nilaiSkor = computed(() => {
    const nilai: Record<string, string> = {}
    let total = 0
    let lengkap = true
    for (const b of bidang.value) {
      if (!b.nilai_ke) continue
      const index = b.pilihan?.indexOf(form[b.kode]) ?? -1
      if (index < 0) lengkap = false
      nilai[b.nilai_ke] = index < 0 ? '' : String(index)
      if (index >= 0) total += index
    }
    nilai.penilaian_totalnilai = lengkap ? String(total) : ''
    return nilai
  })
  const detail = ref<Record<string, string> | null>(null)
  let urutan = 0
  const baris = computed(() => hasil.value.catatan
    .filter(r => Object.values(r).join(' ').toLowerCase().includes(q.value.trim().toLowerCase()))
    .map((r, i) => ({ ...r, _id: i + 1 })))
  const kolomUtama = computed(() => hasil.value.kolom.slice(0, 5))

  function reset() {
    editing.value = null
    errorSimpan.value = ''
    for (const k of Object.keys(form)) delete form[k]
    for (const k of Object.keys(pilihan)) delete pilihan[k]
    for (const b of bidang.value) form[b.kode] = ''
    form.tanggal = new Date(Date.now() + 8 * 3600000).toISOString().slice(0, 19)
    if (props.jenis === 'laporan_operasi') {
      form.tanggal = props.jadwalOperasi.tanggal + 'T' + props.jadwalOperasi.jam_mulai
      form.selesaioperasi = props.jadwalOperasi.tanggal + 'T' + props.jadwalOperasi.jam_selesai
    }
    if (bidang.value.some(b => b.kode === 'tindakan')) form.tindakan = props.jadwalOperasi.nama_paket
    if (bidang.value.some(b => b.kode === 'kd_dokter_bedah')) {
      pilihan.kd_dokter_bedah = { kode: props.jadwalOperasi.kd_dokter, nama: props.jadwalOperasi.nama_dokter }
    }
    if (props.jenis === 'penilaian_pre_operasi') {
      pilihan.kd_dokter = { kode: props.jadwalOperasi.kd_dokter, nama: props.jadwalOperasi.nama_dokter }
      form.rencana_tindakan_bedah = props.jadwalOperasi.nama_paket
    }
    if (props.jenis === 'penilaian_pre_anestesi') {
      form.tanggal_operasi = props.jadwalOperasi.tanggal + 'T' + props.jadwalOperasi.jam_mulai
    }
  }

  function edit(row: Record<string, string>) {
    if (terkunci.value || row._bisa_ubah !== 'true') return
    reset()
    editing.value = { ...row }
    for (const b of bidang.value) {
      const v = row[b.kode] || ''
      if (b.jenis === 'dokter' || b.jenis === 'petugas') {
        pilihan[b.kode] = { kode: v, nama: 'Kode: ' + v }
      } else {
        form[b.kode] = b.jenis === 'datetime-local'
          ? (v.startsWith('0000-') ? '' : v.replace(' ', 'T')) : v
      }
    }
    formVisible.value = true
  }

  function cari(jenis: string, q: string) {
    return request<Record<string, string>[]>('/api/checklist-pre-operasi/referensi?' + new URLSearchParams({ jenis, q }), {
      headers: { Authorization: 'Bearer ' + props.token },
    })
  }

  async function mutasi(hapus = false) {
    if (terkunci.value) return
    const asli = hapus ? hapusTarget.value : editing.value
    if (hapus && (!asli || asli._bisa_ubah !== 'true')) return
    errorSimpan.value = ''
    const data: Record<string, string> = {}
    for (const b of bidang.value) {
      data[b.kode] = b.jenis === 'dokter' || b.jenis === 'petugas'
        ? pilihan[b.kode]?.kode || '' : b.jenis === 'computed' ? nilaiSkor.value[b.kode] : form[b.kode] || ''
      if (!hapus && b.wajib && !data[b.kode]?.trim()) {
        errorSimpan.value = b.label + ' wajib diisi.'
        return
      }
    }
    const id = urutan
    saving.value = true
    try {
      const respons = await request<{ pesan: string }>('/api/jadwal-operasi/pendukung', {
        method: hapus ? 'DELETE' : asli ? 'PUT' : 'POST',
        headers: { Authorization: 'Bearer ' + props.token },
        body: JSON.stringify({
          jenis: props.jenis,
          no_rawat: String(props.patient.no_rawat || ''),
          jadwal: props.jadwalOperasi,
          data,
          asli,
        }),
      })
      if (id !== urutan) return
      notifikasi.sukses(respons.pesan)
      hapusTarget.value = null
      reset()
      await muat()
    } catch (e) {
      if (id === urutan) {
        errorSimpan.value = e instanceof Error ? e.message : 'Catatan gagal diproses.'
        notifikasi.gagal(errorSimpan.value)
      }
    } finally {
      saving.value = false
    }
  }

  function cetak(row: Record<string, string>) {
    const jendela = window.open('', '_blank')
    if (!jendela) { notifikasi.peringatan('Izinkan pop-up untuk mencetak.'); return }
    jendela.opener = null
    const doc = jendela.document
    doc.title = props.judul
    const judul = doc.createElement('h1')
    judul.textContent = props.judul
    const pasien = doc.createElement('p')
    pasien.textContent = String(props.patient.nama_pasien || props.patient.nm_pasien || '') + ' · ' + props.patient.no_rawat
    const tabel = doc.createElement('table')
    for (const kode of hasil.value.kolom) {
      const tr = tabel.insertRow()
      tr.insertCell().textContent = bidang.value.find(b => b.kode === kode)?.label || label(kode)
      tr.insertCell().textContent = row[kode] || '—'
    }
    doc.body.append(judul, pasien, tabel)
    jendela.focus()
    jendela.print()
  }

  function label(kode: string) {
    const khusus: Record<string, string> = {
      kd_dokter: 'Kode Dokter', kd_dokter_bedah: 'Kode Dokter Bedah',
      kd_dokter_anestesi: 'Kode Dokter Anestesi', nip: 'NIP Petugas',
      kd_kamar: 'Kode Kamar', stts_pulang: 'Status Pulang', tgl_operasi: 'Tanggal Operasi',
      diagnosa_preop: 'Diagnosis Pre Operasi', diagnosa_postop: 'Diagnosis Post Operasi',
    }
    return khusus[kode] || kode.replaceAll('_', ' ').replace(/\b\w/g, c => c.toUpperCase())
  }

  async function muat() {
    const id = ++urutan
    error.value = ''
    detail.value = null
    loading.value = true
    try {
      const params = new URLSearchParams({
        jenis: props.jenis,
        no_rawat: String(props.patient.no_rawat || ''),
        tanggal: props.jadwalOperasi.tanggal,
      })
      const data = await request<HasilPendukung>('/api/jadwal-operasi/pendukung?' + params, {
        headers: { Authorization: 'Bearer ' + props.token },
      })
      if (id === urutan) {
        hasil.value = data
        if (!Object.keys(form).length) reset()
      }
    } catch (e) {
      if (id === urutan) error.value = e instanceof Error ? e.message : 'Riwayat gagal dimuat.'
    } finally {
      if (id === urutan) loading.value = false
    }
  }

  watch(() => [props.token, props.patient.no_rawat, props.jenis, props.jadwalOperasi.tanggal], () => {
    q.value = ''
    hasil.value = kosong()
    for (const k of Object.keys(form)) delete form[k]
    editing.value = null
    hapusTarget.value = null
    errorSimpan.value = ''
    void muat()
  }, { immediate: true })
  onBeforeUnmount(() => { urutan++ })
  return {
    loading, error, q, hasil, detail, baris, kolomUtama, label, muat,
    bidang, form, pilihan, editing, hapusTarget, saving, errorSimpan, formVisible,
    terkunci, nilaiSkor, reset, edit, cari, mutasi, cetak,
  }
}
