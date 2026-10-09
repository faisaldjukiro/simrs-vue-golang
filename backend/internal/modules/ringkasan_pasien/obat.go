package ringkasan_pasien

// Sumber obat dipisahkan agar resep, racikan, dan transaksi tidak saling
// menggandakan jumlah atau dianggap sebagai bukti administrasi obat ke pasien.
var sumberObat = []struct{ kode, nama, query string }{
	{
		"resep", "Nomor resep pada kunjungan ini, lintas layanan; bukan bukti pemberian obat",
		`SELECT r.no_resep,
			CASE WHEN r.tgl_peresepan IS NULL OR r.tgl_peresepan='0000-00-00'
				THEN NULL ELSE DATE_FORMAT(TIMESTAMP(r.tgl_peresepan,r.jam_peresepan),'%Y-%m-%d %H:%i:%s') END waktu,
			COALESCE(d.nm_dokter,r.kd_dokter) dokter, r.status layanan
		FROM resep_obat r
		LEFT JOIN dokter d ON d.kd_dokter=r.kd_dokter
		WHERE r.no_rawat=?
		ORDER BY r.tgl_peresepan DESC,r.jam_peresepan DESC,r.no_resep DESC`,
	},
	{
		"resep_item", "Item resep nonracikan; jumlah sesuai resep, bukan dosis yang telah diberikan",
		`SELECT r.no_resep, i.kode_brng kode_obat,
			COALESCE(b.nama_brng,i.kode_brng) obat,
			CAST(i.jml AS CHAR) jumlah, b.kode_sat satuan, i.aturan_pakai
		FROM resep_obat r
		JOIN resep_dokter i ON i.no_resep=r.no_resep
		LEFT JOIN databarang b ON b.kode_brng=i.kode_brng
		WHERE r.no_rawat=?
		ORDER BY r.tgl_peresepan DESC,r.jam_peresepan DESC,r.no_resep DESC,i.kode_brng`,
	},
	{
		"resep_racikan", "Racikan pada resep; jumlah sediaan racikan, bukan jumlah bahan penyusunnya",
		`SELECT r.no_resep, CAST(i.no_racik AS CHAR) no_racik,
			i.nama_racik, CAST(i.jml_dr AS CHAR) jumlah,
			i.aturan_pakai, i.keterangan
		FROM resep_obat r
		JOIN resep_dokter_racikan i ON i.no_resep=r.no_resep
		WHERE r.no_rawat=?
		ORDER BY r.tgl_peresepan DESC,r.jam_peresepan DESC,r.no_resep DESC,i.no_racik`,
	},
	{
		"obat", "Baris transaksi detail_pemberian_obat pada kunjungan ini, lintas layanan; bukan konfirmasi pemberian oleh perawat dan belum dikurangi retur",
		`SELECT DATE_FORMAT(TIMESTAMP(o.tgl_perawatan,o.jam),'%Y-%m-%d %H:%i:%s') waktu,
			o.kode_brng kode_obat, COALESCE(b.nama_brng,o.kode_brng) obat,
			CAST(o.jml AS CHAR) jumlah, b.kode_sat satuan,
			o.status layanan, o.kd_bangsal, o.no_batch, o.no_faktur
		FROM detail_pemberian_obat o
		LEFT JOIN databarang b ON b.kode_brng=o.kode_brng
		WHERE o.no_rawat=?
		ORDER BY o.tgl_perawatan DESC,o.jam DESC,o.kode_brng,o.kd_bangsal,o.no_batch,o.no_faktur`,
	},
}
