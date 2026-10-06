import type { Pasien } from '../../../types/domain'
import type { HasilKardeks } from '../../../types/kardeks'
import { parameterKardeks } from '../../../types/kardeks'
import type { KopPemulangan } from '../../../types/perencanaanPemulangan'
import css from './cetak-kardeks.css?raw'
import { rekapCairan } from './useCairan'

export async function cetakKardeks(
  win: Window,
  pasien: Pasien,
  hasil: HasilKardeks,
  slots: { label: string; tanggal: string; baris: Record<string, string>[] }[],
  sumber: string,
  kop: KopPemulangan,
  grafik: { label: string; min: number; max: number; titik: { x: number; y: number; nilai: number; waktu: string }[] },
) {
  const doc = win.document
  doc.title = 'Kardeks ICU'
  doc.documentElement.lang = 'id'
  doc.body.replaceChildren()
  const style = doc.createElement('style')
  style.textContent = css
  doc.head.append(style)
  function el(tag: string, teks = '', kelas = '') {
    const node = doc.createElement(tag)
    node.textContent = teks
    node.className = kelas
    return node
  }
  const header = el('header')
  const identitas = el('div')
  identitas.append(
    el('h1', kop.nama),
    el('p', [kop.alamat, kop.kabupaten, kop.propinsi].filter(Boolean).join(', ')),
    el('p', [kop.kontak, kop.email].filter(Boolean).join(' | ')),
  )
  const images = ['logo_kota.png', 'icon_rsas.png'].map((file, i) => {
    const img = doc.createElement('img')
    img.alt = i === 0 ? 'Logo kota' : 'Logo rumah sakit'
    img.src = new URL(import.meta.env.BASE_URL + 'img/' + file, window.location.href).href
    return img
  })
  header.append(images[0], identitas, images[1])
  doc.body.append(header, el('h2', 'KARDEKS ICU — RINGKASAN DATA TERCATAT'))
  doc.body.append(
    el('p', [pasien.nm_pasien, 'RM: ' + pasien.no_rkm_medis, 'No. Rawat: ' + pasien.no_rawat].join(' | ')),
    el('p', 'Periode: ' + hasil.mulai + ' sampai sebelum ' + hasil.selesai + ' WITA'),
    el('p', 'Hanya baca. Resep bukan bukti pemberian; transaksi obat bukan konfirmasi waktu pemberian. — berarti belum tercatat, bukan nol.'),
    el('h3', 'Observasi — ' + sumber),
  )
  if (grafik.titik.length) {
    doc.body.append(el('p', 'Grafik ' + grafik.label + ' · Skala data ' + grafik.min + '–' + grafik.max + ' · Titik tanpa interpolasi.'))
    const ns = 'http://www.w3.org/2000/svg'
    const svg = doc.createElementNS(ns, 'svg')
    svg.setAttribute('viewBox', '0 0 800 220')
    svg.setAttribute('class', 'grafik')
    const axis = doc.createElementNS(ns, 'path')
    axis.setAttribute('d', 'M55 25 V180 H745')
    axis.setAttribute('fill', 'none')
    axis.setAttribute('stroke', '#111')
    svg.append(axis)
    for (const [x, y, teks] of [
      [5, 38, String(grafik.max)], [5, 175, String(grafik.min)],
      [55, 207, hasil.mulai.slice(11, 16)], [650, 207, '+24 jam'],
    ] as [number, number, string][]) {
      const label = doc.createElementNS(ns, 'text')
      label.setAttribute('x', String(x))
      label.setAttribute('y', String(y))
      label.textContent = teks
      svg.append(label)
    }
    for (const p of grafik.titik) {
      const point = doc.createElementNS(ns, 'circle')
      point.setAttribute('cx', String(p.x))
      point.setAttribute('cy', String(p.y))
      point.setAttribute('r', '3')
      svg.append(point)
    }
    doc.body.append(svg)
  }
  function table(headers: string[], rows: string[][], classe = '') {
    const t = doc.createElement('table')
    t.className = classe
    const head = t.createTHead().insertRow()
    for (const label of headers) head.append(el('th', label))
    const body = t.createTBody()
    for (const values of rows) {
      const row = body.insertRow()
      for (const value of values) row.insertCell().textContent = value || '—'
    }
    return t
  }
  doc.body.append(table(
    ['Parameter', ...slots.map(s => s.label + '\n' + s.tanggal.slice(5))],
    parameterKardeks.map(p => [
      p.label + ' ' + p.unit,
      ...slots.map(s => s.baris.map(r => r.waktu.slice(11) + ': ' + (r[p.key] || '—')).join('\n')),
    ]),
    'matrix',
  ))
  doc.body.append(el('p', 'CVP dan ICP belum tersedia. Pengaturan ventilator ditampilkan pada waktu pencatatan, tanpa meneruskan nilai otomatis.'))
  const cairan = rekapCairan(hasil)
  if (cairan) {
    doc.body.append(el('h3', 'Rekap cairan tercatat (mL), tanpa IWL'))
    doc.body.append(el('p', 'Masuk: ' + (cairan.masuk ?? '—') + ' | Keluar: ' + (cairan.keluar ?? '—') + ' | Selisih: ' + (cairan.balance ?? '—')))
    doc.body.append(el('p', 'Hanya catatan sepenuhnya dalam periode. Belum dicatat bukan nol. Rekap bukan konfirmasi kelengkapan. Catatan lintas batas tidak dijumlahkan: ' + cairan.lintas))
  }
  for (const b of hasil.bagian) {
    const section = el('section', '', 'bagian')
    section.append(el('h3', b.nama))
    if (!b.baris.length) section.append(el('p', 'Belum ada catatan pada periode ini.'))
    else {
      const keys = Object.keys(b.baris[0])
      section.append(table(keys.map(k => k.replaceAll('_', ' ')), b.baris.map(r => keys.map(k => r[k]))))
    }
    doc.body.append(section)
  }
  await Promise.all(images.map(img => new Promise<void>((resolve, reject) => {
    const timer = window.setTimeout(() => finish(false), 15000)
    function finish(ok: boolean) {
      window.clearTimeout(timer)
      img.onload = null
      img.onerror = null
      if (ok) resolve()
      else reject(new Error('Logo cetak tidak dapat dimuat. Periksa file public/img.'))
    }
    img.onload = () => finish(true)
    img.onerror = () => finish(false)
    if (img.complete) finish(img.naturalWidth > 0)
  })))
  await doc.fonts.ready
  if (!win.closed) {
    win.focus()
    win.requestAnimationFrame(() => { if (!win.closed) win.print() })
  }
}
