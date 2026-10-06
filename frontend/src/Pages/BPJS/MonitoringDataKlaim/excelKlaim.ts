import type { FilterKlaim, HasilMonitoringKlaim, KlaimMonitoring } from '../../../types/monitoringDataKlaim'

type Sel = string | number | null | undefined


function biaya(value: Sel): Sel {
  if (value == null || value === '') return null
  const teks = String(value).trim()
  return /^-?\d+(\.\d+)?$/.test(teks) && Number.isFinite(Number(teks)) ? Number(teks) : teks
}

// Workbook XLSX asli. Library dimuat hanya saat ekspor dijalankan.
export async function dokumenExcelKlaim(rows: KlaimMonitoring[], hasil: HasilMonitoringKlaim, filter: FilterKlaim, pencarian: string): Promise<Uint8Array<ArrayBuffer>> {
  if (rows.length > 1048575) throw new Error('Jumlah klaim melebihi batas baris Excel. Persempit periode atau pencarian.')
  const { default: ExcelJS } = await import('exceljs')
  const workbook = new ExcelJS.Workbook()
  workbook.creator = 'SIRAPI'
  const headers = [
    'No.', 'No. SEP', 'No. FPK', 'No. RM', 'Nama Peserta', 'No. Kartu',
    'Jenis Rawat', 'Poli', 'Kelas Rawat', 'No. Rawat SIMRS', 'Dokter SIMRS',
    'Sumber Dokter', 'Keterangan Pencocokan', 'Kode INA-CBG', 'Nama INA-CBG',
    'Tanggal SEP', 'Tanggal Pulang', 'Status', 'Tarif RS', 'Tarif Grouper',
    'Pengajuan', 'Disetujui', 'Top Up',
  ]
  const data: Sel[][] = rows.map((r, i) => [
    i + 1, r.noSEP, r.noFPK, r.peserta?.noMR, r.peserta?.nama, r.peserta?.noKartu,
    r.jenisPelayanan, r.poli, r.kelasRawat, r.dokter_simrs?.no_rawat,
    r.dokter_simrs?.nama_dokter, r.dokter_simrs?.sumber, r.dokter_simrs?.keterangan,
    r.Inacbg?.kode, r.Inacbg?.nama, r.tglSep, r.tglPulang, r.status,
    ...['byTarifRS', 'byTarifGruper', 'byPengajuan', 'bySetujui', 'byTopup'].map(k => biaya(r.biaya?.[k])),
  ])
  const info: Sel[][] = [
    ['Periode hasil', filter.tanggal_mulai + ' s.d. ' + filter.tanggal_selesai],
    ['Jenis pelayanan', filter.jenis_pelayanan],
    ['Status klaim (1 Proses, 2 Pending, 3 Terbayar)', filter.status_klaim],
    ['Pencarian tabel', pencarian],
    ['Jumlah baris diekspor', rows.length],
    ['Cakupan', 'Semua baris hasil pencarian tabel, bukan hanya halaman aktif.'],
    ['Sumber dokter', 'SIMRS saat penarikan data; bukan penetapan dokter dari BPJS.'],
    ['Peringatan SIMRS', hasil.peringatan_simrs || ''],
    ['Kelengkapan VClaim', hasil.tanggal_gagal?.length ? 'Sebagian tanggal/pelayanan gagal. Hasil tidak lengkap.' : 'Tidak ada tanggal gagal dilaporkan.'],
    ...(hasil.tanggal_gagal || []).map(r => ['Tanggal/pelayanan gagal', r.tanggal, r.jenis_pelayanan, r.pesan]),
    ...(hasil.tanggal_tanpa_data || []).map(t => ['Tanggal tanpa data', t]),
  ]
  function sheet(nama: string, isi: Sel[][], lebar: number[]) {
    const worksheet = workbook.addWorksheet(nama, {
      views: [{ state: 'frozen', ySplit: 1 }],
    })
    worksheet.columns = lebar.map(width => ({ width }))
    worksheet.addRows(isi.map(row => row.map(value => value ?? null)))
    worksheet.eachRow(row => {
      row.eachCell(cell => {
        cell.alignment = { vertical: 'top', wrapText: true }
        if (typeof cell.value === 'string') cell.numFmt = '@'
      })
    })
    const header = worksheet.getRow(1)
    header.font = { bold: true, color: { argb: 'FFFFFFFF' } }
    header.fill = { type: 'pattern', pattern: 'solid', fgColor: { argb: 'FF0D9488' } }
    header.height = 30
    return worksheet
  }
  const klaim = sheet('Data Klaim', [headers, ...data], [
    7, 24, 24, 16, 32, 24, 18, 25, 16, 24, 45, 28, 45, 18, 45, 18, 18, 18,
    20, 20, 20, 20, 20,
  ])
  klaim.autoFilter = { from: { row: 1, column: 1 }, to: { row: rows.length + 1, column: headers.length } }
  for (let col = 19; col <= 23; col++) klaim.getColumn(col).numFmt = '#,##0.00'
  sheet('Keterangan', [['Informasi', 'Nilai'], ...info], [44, 70, 25, 70])
  return new Uint8Array(await workbook.xlsx.writeBuffer())
}

export async function unduhExcelKlaim(rows: KlaimMonitoring[], hasil: HasilMonitoringKlaim, filter: FilterKlaim, pencarian: string, bolehUnduh: () => boolean = () => true) {
  const doc = await dokumenExcelKlaim(rows, hasil, filter, pencarian)
  if (!bolehUnduh()) return
  const url = URL.createObjectURL(new Blob([doc], { type: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet' }))
  const link = document.createElement('a')
  link.href = url
  link.download = `monitoring-klaim-${filter.tanggal_mulai}-${filter.tanggal_selesai}.xlsx`
  document.body.append(link)
  link.click()
  link.remove()
  window.setTimeout(() => URL.revokeObjectURL(url), 1000)
}
