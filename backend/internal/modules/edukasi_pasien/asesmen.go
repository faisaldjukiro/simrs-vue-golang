package edukasi_pasien

import (
	"fmt"
	"strings"
	"time"
)

func validasiAsesmen(in *Input, fotoBaru map[string]bool) error {
	fotoPetugas := fotoBaru["foto"] || (in.Data["foto"] != "" && in.Data["foto"] != "-")
	fotoPenerima := fotoBaru["foto_penerima"] || (in.Data["foto_penerima"] != "" && in.Data["foto_penerima"] != "-")
	adaFoto := fotoPetugas || fotoPenerima
	adaParaf := in.Data["paraf_petugas"] != "" || in.Data["paraf_penerima"] != ""
	if adaFoto && adaParaf {
		return fmt.Errorf("%w: pilih dua paraf atau dua foto, tidak boleh bersamaan", ErrValidasi)
	}
	if in.Data["status_verifikasi"] == "Terverifikasi" {
		if adaFoto && (!fotoPetugas || !fotoPenerima) {
			return fmt.Errorf("%w: lengkapi foto petugas dan penerima edukasi", ErrValidasi)
		}
		if strings.TrimSpace(in.Data["nama_penerima"]) == "" {
			return fmt.Errorf("%w: lengkapi nama penerima edukasi", ErrValidasi)
		}
	}
	if in.Data["status_verifikasi"] == "Terverifikasi" && !adaFoto {
		if in.Data["paraf_petugas"] == "" || in.Data["paraf_penerima"] == "" {
			return fmt.Errorf("%w: lengkapi paraf petugas dan penerima, atau gunakan foto edukasi", ErrValidasi)
		}
	}
	for _, pihak := range []struct{ key, nama string }{{"paraf_petugas", "petugas"}, {"paraf_penerima", "penerima"}} {
		if err := validasiParaf(in.Data[pihak.key]); err != nil {
			return fmt.Errorf("%w (%s edukasi)", err, pihak.nama)
		}
	}

	for _, k := range []string{"materi", "catatan_verifikasi"} {
		in.Data[k] = strings.TrimSpace(in.Data[k])
		if len(in.Data[k]) > 65535 {
			return fmt.Errorf("%w: %s maksimal 65535 byte", ErrValidasi, k)
		}
	}
	switch in.Data["tingkat_pemahaman"] {
	case "", "Belum memahami", "Sebagian memahami", "Memahami":
	default:
		return fmt.Errorf("%w: tingkat pemahaman tidak valid", ErrValidasi)
	}
	switch in.Data["status_verifikasi"] {
	case "", "Belum diverifikasi":
		if in.Data["tanggal_verifikasi"] != "" || in.Data["nip_verifikator"] != "" {
			return fmt.Errorf("%w: waktu dan petugas verifikasi hanya diisi saat status Terverifikasi", ErrValidasi)
		}
	case "Terverifikasi":
		for _, bidang := range []struct{ key, label string }{
			{"kemampuan_membaca", "kemampuan membaca"}, {"tingkat_pendidikan", "tingkat pendidikan"},
			{"bahasa", "bahasa"}, {"hambatan_emosional", "hambatan emosional"}, {"motivasi", "motivasi"},
			{"keterbatasan_fisik", "keterbatasan fisik"}, {"keterbatasan_kognitif", "keterbatasan kognitif"},
			{"kesediaan_menerima", "kesediaan menerima informasi"}, {"nilai_budaya", "nilai budaya / keyakinan"},
			{"materi", "materi edukasi"}, {"penerima", "penerima edukasi"}, {"catatan_verifikasi", "catatan hasil verifikasi"},
		} {
			if strings.TrimSpace(in.Data[bidang.key]) == "" {
				return fmt.Errorf("%w: lengkapi %s sebelum verifikasi", ErrValidasi, bidang.label)
			}
		}
		if in.Data["tingkat_pemahaman"] == "" || in.Data["tanggal_verifikasi"] == "" || in.Data["nip_verifikator"] == "" {
			return fmt.Errorf("%w: lengkapi tingkat pemahaman, waktu, dan petugas verifikasi", ErrValidasi)
		}
		waktu := strings.ReplaceAll(in.Data["tanggal_verifikasi"], "T", " ")
		if len(waktu) == 16 {
			waktu += ":00"
		}
		v, err := time.Parse("2006-01-02 15:04:05", waktu)
		if err != nil || v.Year() < 1000 {
			return fmt.Errorf("%w: waktu verifikasi tidak valid", ErrValidasi)
		}
		edukasi, _ := time.Parse("2006-01-02 15:04:05", in.Data["tgl_perawatan"]+" "+in.Data["jam_rawat"])
		if v.Before(edukasi) {
			return fmt.Errorf("%w: waktu verifikasi tidak boleh sebelum waktu edukasi", ErrValidasi)
		}
		in.Data["tanggal_verifikasi"] = waktu
	default:
		return fmt.Errorf("%w: status verifikasi tidak valid", ErrValidasi)
	}
	return nil
}

// DATETIME kosong disimpan NULL, bukan tanggal nol. Kolom asesmen opsional
// yang kosong juga tetap NULL; snapshot API menyajikannya sebagai string kosong.
func nilaiSimpan(kolom, nilai string) any {
	if nilai == "" {
		for _, k := range Kolom()[10:] {
			if kolom == k {
				return nil
			}
		}
	}
	return nilai
}
