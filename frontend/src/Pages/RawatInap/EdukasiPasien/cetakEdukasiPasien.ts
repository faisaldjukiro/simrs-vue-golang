import type { Pasien } from '../../../types/domain'
import type { KopPemulangan } from '../../../types/perencanaanPemulangan'
import { bidangAsesmen, bidangEdukasi, bidangVerifikasi, type CatatanEdukasi } from '../../../types/edukasiPasien'
import styleCetak from './cetak-edukasi.css?raw'

export async function cetakEdukasiPasien(
  win: Window,
  pasien: Pasien,
  catatan: CatatanEdukasi[],
  kop: KopPemulangan,
  masihAktif: () => boolean,
) {
  const doc = win.document
  const namaPasien = pasien.nama_pasien || pasien.nm_pasien || '-'
  const noRekamMedis = pasien.no_rekam_medis || pasien.no_rkm_medis || '-'
  doc.title = `Edukasi Pasien - ${noRekamMedis}`
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
  const gambar: Promise<void>[] = []
  function muatGambar(src: string, alt: string, wajib: boolean) {
    const img = doc.createElement('img')
    img.alt = alt
    gambar.push(new Promise<void>((resolve, reject) => {
      const timer = window.setTimeout(() => selesai(false), 15000)
      let tuntas = false
      function selesai(ok: boolean) {
        if (tuntas) return
        tuntas = true
        window.clearTimeout(timer)
        img.onload = null
        img.onerror = null
        if (ok) resolve()
        else if (wajib) reject(new Error('Logo kop belum dapat dimuat. Periksa file di public/img lalu coba cetak kembali.'))
        else {
          img.replaceWith(elemen('p', 'Foto dokumentasi tersimpan, tetapi tidak dapat dimuat.'))
          resolve()
        }
      }
      img.onload = () => selesai(true)
      img.onerror = () => selesai(false)
      img.src = src
      // Tunggu elemen terpasang sebelum mengganti gambar yang gagal dimuat.
      queueMicrotask(() => { if (img.complete) selesai(img.naturalWidth > 0) })
    }))
    return img
  }
  for (const r of catatan) {
    const lembar = elemen('article', '', 'lembar-edukasi')
    const header = elemen('header', '', 'kop')
    const identitasRS = elemen('div', '', 'identitas-rs')
    identitasRS.append(elemen('h1', kop.nama))
    identitasRS.append(elemen('p', [kop.alamat, kop.kabupaten, kop.propinsi].filter(Boolean).join(', ')))
    identitasRS.append(elemen('p', [
      kop.kontak ? 'Telp. ' + kop.kontak : '', kop.email ? 'Email: ' + kop.email : '',
    ].filter(Boolean).join(' | ')))
    const logo = (file: string, alt: string) => muatGambar(
      new URL(import.meta.env.BASE_URL + 'img/' + file, window.location.href).href, alt, true,
    )
    header.append(logo('logo_kota.png', 'Logo kota'), identitasRS, logo('icon_rsas.png', 'Logo rumah sakit'))
    lembar.append(header, elemen('h2', 'CATATAN EDUKASI PASIEN'))
    const identitas = elemen('div', '', 'identitas-pasien')
    for (const [label, value] of [
      ['Nama pasien', namaPasien], ['No. RM', noRekamMedis],
      ['No. Rawat', pasien.no_rawat], ['Ruangan', r.nama_ruangan || r.data.kd_ruangan],
      ['Tanggal edukasi', r.data.tgl_perawatan], ['Jam edukasi (WITA)', r.data.jam_rawat],
      ['Petugas', r.nama_petugas],
    ]) {
      const baris = elemen('div')
      baris.append(elemen('strong', label + ': '), elemen('span', value || '-'))
      identitas.append(baris)
    }
    lembar.append(identitas)
    function bagian(judul: string, entries: { key: string; label: string }[]) {
      lembar.append(elemen('h3', judul))
      const table = doc.createElement('table')
      const body = table.createTBody()
      for (const b of entries) {
        const tr = body.insertRow()
        const th = doc.createElement('th')
        th.scope = 'row'
        th.textContent = b.label
        tr.append(th)
        tr.insertCell().textContent = (b.key === 'nama_verifikator' ? r.nama_verifikator : r.data[b.key]) || 'Belum dicatat'
      }
      lembar.append(table)
    }
    bagian('Pelaksanaan Edukasi', [...bidangEdukasi])
    bagian('Asesmen Kebutuhan Belajar', bidangAsesmen)
    bagian('Pemahaman dan Verifikasi', [
      ...bidangVerifikasi.filter(b => b.key !== 'nip_verifikator'),
      { key: 'nama_verifikator', label: 'Nama Verifikator' },
    ])
    if (r.foto_url || r.data.foto) {
      const dokumentasi = elemen('section', '', 'dokumentasi')
      dokumentasi.append(elemen('h3', 'Foto Dokumentasi'))
      dokumentasi.append(r.foto_url
        ? muatGambar(r.foto_url, 'Foto dokumentasi edukasi pasien', false)
        : elemen('p', 'Foto dokumentasi tersimpan, tetapi tidak tersedia.'))
      lembar.append(dokumentasi)
    }
    doc.body.append(lembar)
  }
  await Promise.all(gambar)
  await doc.fonts.ready
  if (win.closed) return
  if (!masihAktif()) { win.close(); return }
  win.focus()
  win.requestAnimationFrame(() => {
    if (win.closed) return
    if (masihAktif()) win.print()
    else win.close()
  })
}
