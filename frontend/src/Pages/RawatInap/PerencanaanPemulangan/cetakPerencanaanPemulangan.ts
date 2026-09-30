import type { Pasien } from '../../../types/domain'
import { penilaianPemulangan } from '../../../types/perencanaanPemulangan'
import type { CatatanPemulangan, KopPemulangan } from '../../../types/perencanaanPemulangan'
import styleCetak from './cetak-pemulangan.css?raw'

export async function cetakFormulirPemulangan(
  win: Window,
  pasien: Pasien,
  catatan: CatatanPemulangan,
  kop: KopPemulangan,
) {
  const doc = win.document
  doc.title = 'Perencanaan Pemulangan Pasien'
  doc.documentElement.lang = 'id'
  doc.body.replaceChildren()
  const style = doc.createElement('style')
  style.textContent = styleCetak
  doc.head.append(style)
  function elemen(tag: string, teks = '', kelas = '') {
    const node = doc.createElement(tag)
    node.textContent = teks
    node.className = kelas
    return node
  }
  const header = elemen('header', '', 'kop')
  const images: HTMLImageElement[] = []
  function logo(file: string, alt: string) {
    const img = doc.createElement('img')
    img.alt = alt
    img.src = new URL(import.meta.env.BASE_URL + 'img/' + file, window.location.href).href
    images.push(img)
    return img
  }
  const identitasRS = elemen('div', '', 'identitas-rs')
  identitasRS.append(elemen('h1', kop.nama))
  identitasRS.append(elemen('p', [kop.alamat, kop.kabupaten, kop.propinsi].filter(Boolean).join(', ')))
  identitasRS.append(elemen('p', [
    kop.kontak ? 'Telp. ' + kop.kontak : '',
    kop.email ? 'Email: ' + kop.email : '',
  ].filter(Boolean).join(' | ')))
  header.append(logo('logo_kota.png', 'Logo kota'), identitasRS, logo('icon_rsas.png', 'Logo rumah sakit'))
  doc.body.append(header, elemen('h2', 'PERENCANAAN PEMULANGAN PASIEN'))
  const identitas = elemen('div', '', 'identitas-pasien')
  for (const [label, value] of [
    ['Nama pasien', pasien.nm_pasien],
    ['No. RM', pasien.no_rkm_medis],
    ['No. Rawat', pasien.no_rawat],
    ['Rencana pulang', catatan.data.rencana_pulang],
  ]) {
    const baris = elemen('div')
    baris.append(elemen('strong', label + ': '), elemen('span', String(value || '-')))
    identitas.append(baris)
  }
  doc.body.append(identitas)
  for (const [label, value] of [
    ['Diagnosis medis', catatan.data.diagnosa_medis],
    ['Alasan masuk / dirawat', catatan.data.alasan_masuk],
  ]) {
    const p = elemen('p', '', 'ringkasan')
    p.append(elemen('strong', label + ': '), elemen('span', value || '-'))
    doc.body.append(p)
  }
  const table = doc.createElement('table')
  const colgroup = doc.createElement('colgroup')
  for (const width of ['7%', '43%', '18%', '32%']) {
    const col = doc.createElement('col')
    col.style.width = width
    colgroup.append(col)
  }
  table.append(colgroup)
  const head = table.createTHead().insertRow()
  for (const label of ['No.', 'Penilaian kebutuhan', 'Jawaban', 'Keterangan']) head.append(elemen('th', label))
  const body = table.createTBody()
  penilaianPemulangan.forEach((p, index) => {
    const tr = body.insertRow()
    for (const value of [String(index + 1), p.label, catatan.data[p.key], catatan.data['keterangan_' + p.key]]) {
      tr.insertCell().textContent = value || '-'
    }
  })
  doc.body.append(table)
  const tandaTangan = elemen('section', '', 'tanda-tangan')
  tandaTangan.append(elemen('p', 'Tanggal tanda tangan: ................................', 'tanggal-ttd'))
  const pihak = elemen('div', '', 'pihak')
  for (const [label, nama, nip] of [
    ['Pasien / Keluarga', catatan.data.nama_pasien_keluarga, ''],
    ['Petugas', catatan.nama_petugas, catatan.data.nip],
  ]) {
    const kolom = elemen('div', '', 'penanda-tangan')
    kolom.append(elemen('p', label), elemen('div', '', 'ruang-ttd'), elemen('strong', nama || '-'))
    if (nip) kolom.append(elemen('p', 'NIP: ' + nip))
    pihak.append(kolom)
  }
  tandaTangan.append(pihak)
  doc.body.append(tandaTangan)
  await Promise.all(images.map(img => new Promise<void>((resolve, reject) => {
    const timer = window.setTimeout(() => selesai(false), 15000)
    function selesai(ok: boolean) {
      window.clearTimeout(timer)
      img.onload = null
      img.onerror = null
      if (ok) resolve()
      else reject(new Error('Logo kop belum dapat dimuat. Periksa file di public/img lalu coba cetak kembali.'))
    }
    img.onload = () => selesai(true)
    img.onerror = () => selesai(false)
    if (img.complete) selesai(img.naturalWidth > 0)
  })))
  await doc.fonts.ready
  if (!win.closed) {
    win.focus()
    win.requestAnimationFrame(() => { if (!win.closed) win.print() })
  }
}
