import { kategoriLaboratorium } from '../../../types/permintaanLaboratorium'
import type { PermintaanLaboratorium, PropsLaboratorium } from '../../../types/permintaanLaboratorium'

// Dokumen cetak terisolasi: tidak mencetak navigasi, pasien lain, atau kontrol edit.
export function cetakPermintaanLaboratorium(
  permintaan: PermintaanLaboratorium,
  pasien: PropsLaboratorium['patient'],
) {
  const frame = document.createElement('iframe')
  frame.title = `Cetak ${permintaan.nomor}`
  frame.style.cssText = 'position:fixed;width:0;height:0;border:0;bottom:0;left:0'
  document.body.append(frame)
  const doc = frame.contentDocument
  const win = frame.contentWindow
  if (!doc || !win) {
    frame.remove()
    return
  }
  doc.title = `Permintaan Laboratorium ${permintaan.nomor}`
  const style = doc.createElement('style')
  style.textContent = `
    @page { size: A4; margin: 16mm; }
    body { font: 12px/1.5 Arial, sans-serif; color: #111; background: #fff; }
    h1 { font-size: 20px; margin-bottom: 4px; }
    h2 { font-size: 14px; margin-top: 20px; }
    p { margin: 5px 0; white-space: pre-wrap; overflow-wrap: anywhere; }
    table { width: 100%; border-collapse: collapse; margin-top: 16px; }
    th, td { padding: 8px; border: 1px solid #999; text-align: left; vertical-align: top; }
    thead { display: table-header-group; }
    tr { break-inside: avoid; }
    footer { margin-top: 24px; font-size: 11px; }
  `
  doc.head.append(style)
  const tulis = (tag: string, teks: string, induk: HTMLElement = doc.body) => {
    const elemen = doc.createElement(tag)
    elemen.textContent = teks
    induk.append(elemen)
    return elemen
  }
  const kategori = kategoriLaboratorium.find((item) => item.value === permintaan.kategori)?.label || permintaan.kategori
  tulis('h1', `Permintaan Laboratorium — ${kategori}`)
  tulis('p', `Nomor: ${permintaan.nomor}`)
  tulis('p', `Tanggal/jam permintaan: ${permintaan.tanggal} ${permintaan.jam} WITA`)
  tulis('p', `Pasien: ${pasien.nama_pasien || pasien.nm_pasien || '-'}`)
  tulis('p', `No. RM: ${pasien.no_rm || pasien.no_rkm_medis || '-'} | No. rawat: ${permintaan.no_rawat}`)
  tulis('p', `Dokter perujuk: ${permintaan.nama_dokter} (${permintaan.kode_dokter})`)
  tulis('p', `Informasi tambahan: ${permintaan.informasi_tambahan}`)
  tulis('p', `Diagnosis klinis: ${permintaan.diagnosis_klinis}`)
  if (permintaan.kategori === 'PA') {
    tulis('h2', 'Informasi spesimen / riwayat PA')
    const p = permintaan.spesimen
    for (const [label, nilai] of [
      ['Tanggal pengambilan bahan', p.pengambilan_bahan],
      ['Diperoleh dengan', p.diperoleh_dengan],
      ['Lokasi jaringan', p.lokasi_jaringan],
      ['Diawetkan dengan', p.diawetkan_dengan],
      ['Pernah dilakukan di', p.pernah_dilakukan_di],
      ['Tanggal PA sebelumnya', p.tanggal_pa_sebelumnya],
      ['Nomor PA sebelumnya', p.nomor_pa_sebelumnya],
      ['Diagnosis PA sebelumnya', p.diagnosa_pa_sebelumnya],
    ]) tulis('p', `${label}: ${nilai || '-'}`)
  }
  const table = tulis('table', '')
  const head = tulis('tr', '', tulis('thead', '', table))
  for (const teks of ['No.', 'Kode', 'Pemeriksaan', 'Detail yang diminta']) tulis('th', teks, head)
  const body = tulis('tbody', '', table)
  permintaan.pemeriksaan.forEach((item, index) => {
    const row = tulis('tr', '', body)
    for (const teks of [String(index + 1), item.kode, item.nama, item.detail.map((detail) => detail.nama).join(', ') || '-']) {
      tulis('td', teks, row)
    }
  })
  tulis('footer', 'Dicetak dari SIRAPI. Dokumen permintaan pemeriksaan, bukan hasil laboratorium atau tanda tangan elektronik.')
  win.addEventListener('afterprint', () => frame.remove(), { once: true })
  win.requestAnimationFrame(() => {
    win.focus()
    win.print()
  })
}
