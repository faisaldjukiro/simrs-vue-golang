package idrg

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"simrs-backend/internal/config"
)

var (
	ErrEKlaimBelumAktif = errors.New("konfigurasi e-klaim belum lengkap")
	ErrInputTidakValid  = errors.New("input idrg tidak valid")
	ErrAksiBelumSiap    = errors.New("aksi idrg belum siap")
)

type Layanan struct {
	repositori *Repositori
	config     config.EKlaim
	client     *http.Client
}

type InputProses struct {
	Aksi     string         `json:"aksi"`
	NoRawat  string         `json:"no_rawat"`
	NoSEP    string         `json:"no_sep"`
	Username string         `json:"username"`
	Pasien   map[string]any `json:"pasien"`
	Klaim    map[string]any `json:"klaim"`
}

type HasilProses struct {
	Sukses  bool              `json:"sukses"`
	Aksi    string            `json:"aksi"`
	NoRawat string            `json:"no_rawat"`
	NoSEP   string            `json:"no_sep"`
	Pesan   string            `json:"pesan"`
	Tahapan []TahapanProses   `json:"tahapan,omitempty"`
	Hasil   *ResponseEClaim   `json:"hasil,omitempty"`
	Meta    map[string]string `json:"meta,omitempty"`
}

type TahapanProses struct {
	Nama  string         `json:"nama"`
	Hasil ResponseEClaim `json:"hasil"`
}

type ResponseEClaim struct {
	Sukses    bool           `json:"sukses"`
	Kode      int            `json:"kode"`
	Status    int            `json:"status"`
	Pesan     string         `json:"pesan"`
	ErrorNo   any            `json:"error_no"`
	Metadata  map[string]any `json:"metadata"`
	Data      any            `json:"data"`
	Response  any            `json:"response"`
	Duplicate any            `json:"duplicate"`
	Raw       map[string]any `json:"raw"`
}

func NewLayanan(repositori *Repositori, eklaimConfig config.EKlaim) *Layanan {
	return &Layanan{
		repositori: repositori,
		config:     eklaimConfig,
		client: &http.Client{
			Timeout: eklaimConfig.Timeout,
		},
	}
}

func (l *Layanan) Proses(ctx context.Context, input InputProses) (HasilProses, error) {
	input.Aksi = strings.TrimSpace(input.Aksi)
	input.NoRawat = strings.TrimSpace(input.NoRawat)
	input.NoSEP = strings.TrimSpace(input.NoSEP)

	if input.Aksi == "" || input.NoRawat == "" {
		return HasilProses{}, fmt.Errorf("%w: aksi dan nomor rawat wajib diisi", ErrInputTidakValid)
	}
	if !l.config.WSEnabled() {
		return HasilProses{}, ErrEKlaimBelumAktif
	}

	nomorSEP, err := l.nomorSEP(ctx, input)
	if err != nil {
		return HasilProses{}, err
	}
	if nomorSEP == "" {
		return HasilProses{}, fmt.Errorf("%w: nomor SEP tidak ditemukan untuk proses E-Klaim", ErrInputTidakValid)
	}
	input.NoSEP = nomorSEP

	switch input.Aksi {
	case "buat_klaim_baru":
		return l.buatKlaimBaru(ctx, input)
	case "atur_data_klaim":
		return l.aturDataKlaim(ctx, input)
	case "idrg_diagnosa_set":
		return l.setDiagnosaIDRG(ctx, input)
	case "idrg_diagnosa_get":
		hasil, err := l.panggilMetodeWS(ctx, "idrg_diagnosa_get", input.NoSEP, nil, nil)
		return hasilProsesTunggal(input, "Diagnosa IDRG selesai dibaca.", "idrg_diagnosa_get", hasil), err
	case "idrg_procedure_set":
		return l.setProsedurIDRG(ctx, input)
	case "idrg_procedure_get":
		hasil, err := l.panggilMetodeWS(ctx, "idrg_procedure_get", input.NoSEP, nil, nil)
		return hasilProsesTunggal(input, "Prosedur IDRG selesai dibaca.", "idrg_procedure_get", hasil), err
	case "get_claim_data", "ambil_data_klaim":
		hasil, err := l.panggilMetodeWS(ctx, "get_claim_data", input.NoSEP, nil, nil)
		return hasilProsesTunggal(input, "Data klaim selesai dibaca.", "get_claim_data", hasil), err
	case "search_diagnosis":
		var keyword string
		if val, ok := input.Klaim["keyword"].(string); ok {
			keyword = val
		}
		hasil, err := l.panggilMetodeWS(ctx, "search_diagnosis", "", map[string]any{"keyword": keyword}, nil)
		return hasilProsesTunggal(input, "Pencarian diagnosa selesai.", "search_diagnosis", hasil), err
	case "search_diagnosis_idrg":
		var keyword string
		if val, ok := input.Klaim["keyword"].(string); ok {
			keyword = val
		}
		hasil, err := l.panggilMetodeWS(ctx, "search_diagnosis_inagrouper", "", map[string]any{"keyword": keyword}, nil)
		return hasilProsesTunggal(input, "Pencarian diagnosa IDRG selesai.", "search_diagnosis_inagrouper", hasil), err
	case "search_procedures":
		var keyword string
		if val, ok := input.Klaim["keyword"].(string); ok {
			keyword = val
		}
		hasil, err := l.panggilMetodeWS(ctx, "search_procedures", "", map[string]any{"keyword": keyword}, nil)
		return hasilProsesTunggal(input, "Pencarian prosedur selesai.", "search_procedures", hasil), err
	case "search_procedures_idrg":
		var keyword string
		if val, ok := input.Klaim["keyword"].(string); ok {
			keyword = val
		}
		hasil, err := l.panggilMetodeWS(ctx, "search_procedures_inagrouper", "", map[string]any{"keyword": keyword}, nil)
		return hasilProsesTunggal(input, "Pencarian prosedur IDRG selesai.", "search_procedures_inagrouper", hasil), err
	case "grouping_idrg":
		return l.groupingIDRGLengkap(ctx, input)
	case "grouping_idrg_stage_1":
		return l.groupingIDRGStage1(ctx, input)
	case "grouping_idrg_stage_2":
		return l.groupingIDRGStage2(ctx, input)
	case "final_idrg":
		hasil, err := l.panggilMetodeWS(ctx, "idrg_grouper_final", input.NoSEP, nil, nil)
		return hasilProsesTunggal(input, "Final IDRG selesai diproses.", "final_idrg", hasil), err
	case "reedit_idrg":
		hasil, err := l.panggilMetodeWS(ctx, "idrg_grouper_reedit", input.NoSEP, nil, nil)
		return hasilProsesTunggal(input, "Re-edit IDRG selesai diproses.", "reedit_idrg", hasil), err
	case "idrg_ke_inacbg":
		hasil, err := l.panggilMetodeWS(ctx, "idrg_to_inacbg_import", input.NoSEP, nil, nil)
		return hasilProsesTunggal(input, "Import IDRG ke INA-CBG selesai diproses.", "idrg_ke_inacbg", hasil), err
	case "inacbg_diagnosa_get":
		hasil, err := l.panggilMetodeWS(ctx, "inacbg_diagnosa_get", input.NoSEP, nil, nil)
		return hasilProsesTunggal(input, "Diagnosa INA-CBG selesai dibaca.", "inacbg_diagnosa_get", hasil), err
	case "inacbg_procedure_get":
		hasil, err := l.panggilMetodeWS(ctx, "inacbg_procedure_get", input.NoSEP, nil, nil)
		return hasilProsesTunggal(input, "Prosedur INA-CBG selesai dibaca.", "inacbg_procedure_get", hasil), err
	case "grouping_inacbg":
		return l.groupingINACBGLengkap(ctx, input)
	case "grouping_inacbg_stage_2":
		return l.groupingINACBGStage2(ctx, input)
	case "final_inacbg":
		hasil, err := l.panggilMetodeWS(ctx, "inacbg_grouper_final", input.NoSEP, nil, nil)
		return hasilProsesTunggal(input, "Final INA-CBG selesai diproses.", "final_inacbg", hasil), err
	case "reedit_inacbg":
		hasil, err := l.panggilMetodeWS(ctx, "inacbg_grouper_reedit", input.NoSEP, nil, nil)
		return hasilProsesTunggal(input, "Re-edit INA-CBG selesai diproses.", "reedit_inacbg", hasil), err
	case "final_klaim":
		coderNIK, err := l.coderNIK(ctx, input)
		if err != nil {
			return HasilProses{}, err
		}
		hasil, err := l.panggilMetodeWS(ctx, "claim_final", input.NoSEP, map[string]any{"coder_nik": coderNIK}, nil)
		return hasilProsesTunggal(input, "Final Klaim selesai diproses.", "final_klaim", hasil), err
	case "reedit_klaim":
		hasil, err := l.panggilMetodeWS(ctx, "reedit_claim", input.NoSEP, nil, nil)
		return hasilProsesTunggal(input, "Edit ulang klaim selesai diproses.", "reedit_klaim", hasil), err
	case "send_claim_individual", "kirim_klaim":
		hasil, err := l.panggilMetodeWS(ctx, "send_claim_individual", input.NoSEP, nil, nil)
		return hasilProsesTunggal(input, "Kirim Klaim selesai diproses.", "send_claim_individual", hasil), err
	case "cetak_klaim":
		hasil, err := l.panggilMetodeWS(ctx, "claim_print", input.NoSEP, nil, nil)
		return hasilProsesTunggal(input, "Cetak Klaim selesai diproses.", "cetak_klaim", hasil), err
	case "validasi_sitb":
		hasil, err := l.panggilMetodeWS(ctx, "sitb_validate", input.NoSEP, map[string]any{
			"nomor_register_sitb": nilaiKlaim(input, "nomor_register_sitb"),
		}, nil)
		return hasilProsesTunggal(input, "Validasi SITB selesai diproses.", "validasi_sitb", hasil), err
	default:
		return HasilProses{}, fmt.Errorf("%w: aksi %q belum dibuat bertahap di backend Go", ErrAksiBelumSiap, input.Aksi)
	}
}

func (l *Layanan) buatKlaimBaru(ctx context.Context, input InputProses) (HasilProses, error) {
	hasil, err := l.panggilWS(ctx, map[string]any{
		"metadata": map[string]any{
			"method": "new_claim",
		},
		"data": map[string]any{
			"nomor_kartu": nilaiPasien(input, "no_kartu"),
			"nomor_sep":   input.NoSEP,
			"nomor_rm":    nilaiPasien(input, "no_rkm_medis"),
			"nama_pasien": nilaiPasien(input, "nm_pasien"),
			"tgl_lahir":   tanggalJam(nilaiPasien(input, "tgl_lahir"), "00:00:00"),
			"gender":      genderEklaim(nilaiPasien(input, "jk")),
		},
	})
	return hasilProsesTunggal(input, "New Claim selesai diproses.", "new_claim", hasil), err
}

func (l *Layanan) aturDataKlaim(ctx context.Context, input InputProses) (HasilProses, error) {
	coderNIK, err := l.coderNIK(ctx, input)
	if err != nil {
		return HasilProses{}, err
	}

	tanggalMasuk := fallbackString(nilaiKlaim(input, "tgl_masuk"), nilaiPasien(input, "tgl_masuk"))
	if tanggalMasuk == "" {
		tanggalMasuk = time.Now().Format("2006-01-02")
	}
	tanggalPulang := fallbackString(nilaiKlaim(input, "tgl_pulang"), nilaiPasien(input, "tgl_keluar"))
	if tanggalPulang == "" {
		tanggalPulang = tanggalMasuk
	}
	jamMasuk := fallbackString(nilaiKlaim(input, "jam_masuk"), "08:00:00")
	jamPulang := fallbackString(nilaiKlaim(input, "jam_pulang"), "09:00:00")
	icuIndikator := stringBinerDariBool(nilaiBoolKlaim(input, "icu_indikator"))
	ventilatorUseInd := stringBinerDariBool(nilaiBoolKlaim(input, "ventilator_use_ind"))
	upgradeClassInd := stringBinerDariBool(nilaiBoolKlaim(input, "upgrade_class_ind"))
	upgradeClass := ""
	upgradeLOS := "0"
	upgradePayor := ""
	addPaymentPct := "0"
	if upgradeClassInd == "1" {
		upgradeClass = fallbackString(nilaiKlaim(input, "upgrade_class_class"), "kelas_1")
		upgradeLOS = fallbackString(nilaiKlaim(input, "upgrade_class_los"), "0")
		upgradePayor = fallbackString(nilaiKlaim(input, "upgrade_class_payor"), "peserta")
		addPaymentPct = fallbackString(nilaiKlaim(input, "add_payment_pct"), "0")
	}

	payload := map[string]any{
		"nomor_kartu":     nilaiPasien(input, "no_kartu"),
		"tgl_masuk":       tanggalJam(tanggalMasuk, jamMasuk),
		"tgl_pulang":      tanggalJam(tanggalPulang, jamPulang),
		"cara_masuk":      fallbackString(nilaiKlaim(input, "cara_masuk"), "gp"),
		"jenis_rawat":     fallbackString(nilaiKlaim(input, "jenis_rawat"), jenisRawatEklaim(nilaiPasien(input, "jenis_rawat_data"))),
		"kelas_rawat":     fallbackString(nilaiKlaim(input, "kelas_rawat"), fallbackString(nilaiPasien(input, "kelas_rawat"), "3")),
		"adl_sub_acute":   fallbackString(nilaiKlaim(input, "adl_sub_acute"), "0"),
		"adl_chronic":     fallbackString(nilaiKlaim(input, "adl_chronic"), "0"),
		"icu_indikator":   icuIndikator,
		"icu_los":         fallbackString(nilaiKlaim(input, "icu_los"), "0"),
		"ventilator_hour": fallbackString(nilaiKlaim(input, "ventilator_hour"), "0"),
		"ventilator": map[string]string{
			"use_ind":    ventilatorUseInd,
			"start_dttm": tanggalWaktuKlaim(input, "ventilator_start_dttm"),
			"stop_dttm":  tanggalWaktuKlaim(input, "ventilator_stop_dttm"),
		},
		"upgrade_class_ind":   upgradeClassInd,
		"upgrade_class_class": upgradeClass,
		"upgrade_class_los":   upgradeLOS,
		"upgrade_class_payor": upgradePayor,
		"add_payment_pct":     addPaymentPct,
		"birth_weight":        fallbackString(nilaiKlaim(input, "birth_weight"), "0"),
		"sistole":             nilaiIntKlaim(input, "sistole", 110),
		"diastole":            nilaiIntKlaim(input, "diastole", 60),
		"discharge_status":    fallbackString(nilaiKlaim(input, "discharge_status"), "1"),
		"tarif_rs":            tarifRSPasien(input),
		"nomor_kartu_t":       fallbackString(nilaiKlaim(input, "nomor_kartu_t"), "kartu_jkn"),
		"kantong_darah":       nilaiIntKlaim(input, "kantong_darah", 0),
		"alteplase_ind":       nilaiIntKlaim(input, "alteplase_ind", 0),
		"tarif_poli_eks":      fallbackString(nilaiKlaim(input, "tarif_poli_eks"), "0"),
		"nama_dokter":         fallbackString(nilaiKlaim(input, "nama_dokter"), fallbackString(nilaiPasien(input, "nm_dokter"), "-")),
		"kode_tarif":          fallbackString(nilaiKlaim(input, "kode_tarif"), fallbackString(l.config.KodeTarif, "BP")),
		"payor_id":            "3",
		"payor_cd":            fallbackString(nilaiKlaim(input, "jaminan"), "JKN"),
		"cob_cd":              nilaiIntKlaim(input, "cob_cd", 0),
		"coder_nik":           coderNIK,
	}

	if apgar, ok := input.Klaim["apgar"]; ok && apgar != nil {
		payload["apgar"] = apgar
	}
	if persalinan, ok := input.Klaim["persalinan"]; ok && persalinan != nil {
		payload["persalinan"] = persalinan
	}

	hasil, err := l.panggilMetodeWS(ctx, "set_claim_data", input.NoSEP, payload, nil)
	return hasilProsesTunggal(input, "Set Data Klaim selesai diproses.", "set_claim_data", hasil), err
}

func (l *Layanan) groupingIDRGLengkap(ctx context.Context, input InputProses) (HasilProses, error) {
	hasilDiagnosa, err := l.setDiagnosaIDRG(ctx, input)
	if err != nil {
		return HasilProses{}, err
	}
	tahapan := append([]TahapanProses{}, hasilDiagnosa.Tahapan...)
	if !hasilDiagnosa.Sukses {
		return hasilProsesTahapan(input, "Set diagnosa IDRG gagal.", tahapan), nil
	}

	hasilProsedur, err := l.setProsedurIDRG(ctx, input)
	if err != nil {
		return HasilProses{}, err
	}
	tahapan = append(tahapan, hasilProsedur.Tahapan...)
	if !hasilProsedur.Sukses {
		return hasilProsesTahapan(input, "Set prosedur IDRG gagal.", tahapan), nil
	}

	hasilGrouping, err := l.groupingIDRGStage1(ctx, input)
	if err != nil {
		return HasilProses{}, err
	}
	tahapan = append(tahapan, hasilGrouping.Tahapan...)
	if !hasilGrouping.Sukses {
		return hasilProsesTahapan(input, "Grouping IDRG Tahap 1 gagal diproses.", tahapan), nil
	}

	topupCodes := strings.TrimSpace(nilaiKlaim(input, "topup_codes"))
	if topupCodes == "" {
		topupCodes = strings.Join(kodeTopupIDRG(hasilGrouping.Hasil), "#")
	}
	if topupCodes == "" {
		// Tahap 2 hanya berlaku ketika Tahap 1 mengembalikan topup_options.
		// Memaksakan nilai "#" membuat E-Klaim menjawab "Kasus tidak ada topup".
		return hasilProsesTahapan(input, "Grouping IDRG selesai diproses tanpa top-up.", tahapan), nil
	}
	hasilGrouping2, err := l.panggilGroupingWS(ctx, "idrg", "2", input.NoSEP, map[string]any{
		"topup_codes": topupCodes,
	})
	if err != nil {
		return HasilProses{}, err
	}
	tahapan = append(tahapan, TahapanProses{Nama: "idrg_grouper_stage_2", Hasil: hasilGrouping2})
	if !hasilGrouping2.Sukses {
		return hasilProsesTahapan(input, "Grouping IDRG Tahap 2 gagal diproses.", tahapan), nil
	}

	return hasilProsesTahapan(input, "Grouping IDRG Tahap 1 dan Tahap 2 selesai diproses.", tahapan), nil
}

func kodeTopupIDRG(hasil *ResponseEClaim) []string {
	if hasil == nil {
		return nil
	}

	kandidat := []any{hasil.Data, hasil.Response, hasil.Raw["response_idrg"]}
	kode := make([]string, 0)
	for _, nilai := range kandidat {
		objek, ok := nilai.(map[string]any)
		if !ok {
			continue
		}
		if responseIDRG, ada := objek["response_idrg"].(map[string]any); ada {
			objek = responseIDRG
		}
		opsi, ok := objek["topup_options"].([]any)
		if !ok {
			continue
		}
		for _, item := range opsi {
			pilihan, ok := item.(map[string]any)
			if !ok {
				continue
			}
			kodeTopup := strings.TrimSpace(fmt.Sprint(pilihan["code"]))
			if kodeTopup != "" && kodeTopup != "<nil>" {
				kode = append(kode, kodeTopup)
			}
		}
		if len(kode) > 0 {
			return kode
		}
	}

	return nil
}

func (l *Layanan) setDiagnosaIDRG(ctx context.Context, input InputProses) (HasilProses, error) {
	var diagPayload string
	if input.Klaim != nil && input.Klaim["diagnosa"] != nil {
		if ds, ok := input.Klaim["diagnosa"].(string); ok {
			diagPayload = strings.ToUpper(strings.TrimSpace(ds))
		} else if ds, ok := input.Klaim["diagnosa"].([]any); ok {
			var codes []string
			for _, d := range ds {
				if dm, ok := d.(map[string]any); ok {
					if code, ok := dm["kd_diag"].(string); ok && code != "" {
						codes = append(codes, strings.ToUpper(strings.TrimSpace(code)))
					}
				}
			}
			diagPayload = strings.Join(codes, "#")
		}
	}

	if diagPayload == "" {
		diagnosaHasil, err := l.repositori.AmbilDiagnosa(ctx, input.NoRawat)
		if err != nil {
			return HasilProses{}, err
		}

		diagnosa := ambilDaftarDiagnosa(diagnosaHasil)
		if len(diagnosa) == 0 {
			return HasilProses{}, fmt.Errorf("%w: diagnosa pasien belum ditemukan di SIMRS", ErrInputTidakValid)
		}
		diagPayload = stringDiagnosa(diagnosa)
	}

	hasilDiagnosa, err := l.panggilMetodeWS(ctx, "idrg_diagnosa_set", input.NoSEP, map[string]any{
		"diagnosa": diagPayload,
	}, nil)
	if err != nil {
		return HasilProses{}, err
	}
	return hasilProsesTunggal(input, "Set diagnosa IDRG selesai diproses.", "idrg_diagnosa_set", hasilDiagnosa), nil
}

func (l *Layanan) setProsedurIDRG(ctx context.Context, input InputProses) (HasilProses, error) {
	var procPayload string
	if input.Klaim != nil && input.Klaim["prosedur"] != nil {
		if ps, ok := input.Klaim["prosedur"].(string); ok {
			procPayload = strings.ToUpper(strings.TrimSpace(ps))
		} else if ps, ok := input.Klaim["prosedur"].([]any); ok {
			var codes []string
			for _, p := range ps {
				if pm, ok := p.(map[string]any); ok {
					if code, ok := pm["kd_prosedur"].(string); ok && code != "" {
						codes = append(codes, strings.ToUpper(strings.TrimSpace(code)))
					}
				}
			}
			procPayload = strings.Join(codes, "#")
		}
	}

	if procPayload == "" {
		prosedurHasil, err := l.repositori.AmbilProsedur(ctx, input.NoRawat)
		if err != nil {
			return HasilProses{}, err
		}
		procPayload = stringProsedur(ambilDaftarProsedur(prosedurHasil))
	}

	hasilProsedur, err := l.panggilMetodeWS(ctx, "idrg_procedure_set", input.NoSEP, map[string]any{
		"procedure": procPayload,
	}, nil)
	if err != nil {
		return HasilProses{}, err
	}
	return hasilProsesTunggal(input, "Set prosedur IDRG selesai diproses.", "idrg_procedure_set", hasilProsedur), nil
}

func (l *Layanan) groupingIDRGStage1(ctx context.Context, input InputProses) (HasilProses, error) {
	hasilGrouping, err := l.panggilGroupingWS(ctx, "idrg", "1", input.NoSEP, nil)
	if err != nil {
		return HasilProses{}, err
	}
	return hasilProsesTunggal(input, "Grouping IDRG Stage 1 selesai diproses.", "idrg_grouper_stage_1", hasilGrouping), nil
}

func (l *Layanan) groupingIDRGStage2(ctx context.Context, input InputProses) (HasilProses, error) {
	topupCodes := strings.TrimSpace(nilaiKlaim(input, "topup_codes"))
	if topupCodes == "" {
		return HasilProses{}, fmt.Errorf("%w: topup_codes wajib diisi untuk Grouping IDRG Stage 2", ErrInputTidakValid)
	}
	hasilGrouping, err := l.panggilGroupingWS(ctx, "idrg", "2", input.NoSEP, map[string]any{
		"topup_codes": topupCodes,
	})
	if err != nil {
		return HasilProses{}, err
	}
	return hasilProsesTunggal(input, "Grouping IDRG Stage 2 selesai diproses.", "idrg_grouper_stage_2", hasilGrouping), nil
}

func (l *Layanan) groupingINACBGLengkap(ctx context.Context, input InputProses) (HasilProses, error) {
	diagnosaPayload := strings.ToUpper(strings.TrimSpace(nilaiKlaim(input, "diagnosa")))
	if diagnosaPayload == "" {
		return HasilProses{}, fmt.Errorf("%w: coding INA-CBG belum diimpor atau diagnosa masih kosong", ErrInputTidakValid)
	}

	prosedurPayload := strings.ToUpper(strings.TrimSpace(nilaiKlaim(input, "prosedur")))
	if prosedurPayload == "" {
		prosedurPayload = "#"
	}

	tahapan := make([]TahapanProses, 0, 4)

	hasilDiagnosa, err := l.panggilMetodeWS(ctx, "inacbg_diagnosa_set", input.NoSEP, map[string]any{
		"diagnosa": diagnosaPayload,
	}, nil)
	if err != nil {
		return HasilProses{}, err
	}
	tahapan = append(tahapan, TahapanProses{Nama: "inacbg_diagnosa_set", Hasil: hasilDiagnosa})
	if !hasilDiagnosa.Sukses {
		return hasilProsesTahapan(input, "Set diagnosa INA-CBG gagal.", tahapan), nil
	}

	hasilProsedur, err := l.panggilMetodeWS(ctx, "inacbg_procedure_set", input.NoSEP, map[string]any{
		"procedure": prosedurPayload,
	}, nil)
	if err != nil {
		return HasilProses{}, err
	}
	tahapan = append(tahapan, TahapanProses{Nama: "inacbg_procedure_set", Hasil: hasilProsedur})
	if !hasilProsedur.Sukses {
		return hasilProsesTahapan(input, "Set prosedur INA-CBG gagal.", tahapan), nil
	}

	hasilGrouping1, err := l.panggilGroupingWS(ctx, "inacbg", "1", input.NoSEP, nil)
	if err != nil {
		return HasilProses{}, err
	}
	tahapan = append(tahapan, TahapanProses{Nama: "inacbg_grouper_stage_1", Hasil: hasilGrouping1})

	if !hasilGrouping1.Sukses {
		return hasilProsesTahapan(input, "Grouping INA-CBG Stage 1 gagal.", tahapan), nil
	}

	specialCmgCode := strings.TrimSpace(nilaiKlaim(input, "special_cmg"))
	if specialCmgCode == "" {
		// Format resmi E-Klaim untuk melanjutkan Tahap 2 tanpa Special CMG.
		specialCmgCode = "#"
	}

	hasilGrouping2, err := l.panggilGroupingWS(ctx, "inacbg", "2", input.NoSEP, map[string]any{
		"special_cmg": specialCmgCode,
	})
	if err != nil {
		return HasilProses{}, err
	}
	tahapan = append(tahapan, TahapanProses{Nama: "inacbg_grouper_stage_2", Hasil: hasilGrouping2})
	if !hasilGrouping2.Sukses {
		return hasilProsesTahapan(input, "Grouping INA-CBG Tahap 2 gagal.", tahapan), nil
	}

	return hasilProsesTahapan(input, "Grouping INA-CBG Tahap 1 dan Tahap 2 selesai diproses.", tahapan), nil
}

func (l *Layanan) groupingINACBGStage2(ctx context.Context, input InputProses) (HasilProses, error) {
	specialCmgCode := strings.TrimSpace(nilaiKlaim(input, "special_cmg"))
	if specialCmgCode == "" {
		return HasilProses{}, fmt.Errorf("%w: special_cmg wajib diisi untuk Grouping INA-CBG Tahap 2", ErrInputTidakValid)
	}

	hasilGrouping, err := l.panggilGroupingWS(ctx, "inacbg", "2", input.NoSEP, map[string]any{
		"special_cmg": specialCmgCode,
	})
	if err != nil {
		return HasilProses{}, err
	}

	return hasilProsesTunggal(input, "Grouping INA-CBG Tahap 2 selesai diproses.", "inacbg_grouper_stage_2", hasilGrouping), nil
}

func (l *Layanan) nomorSEP(ctx context.Context, input InputProses) (string, error) {
	if input.NoSEP != "" {
		return input.NoSEP, nil
	}
	return l.repositori.NomorSEP(ctx, input.NoRawat)
}

func (l *Layanan) coderNIK(ctx context.Context, input InputProses) (string, error) {
	coderNIK, err := l.repositori.AmbilCoderNIK(ctx, input.Username)
	if err != nil {
		return "", err
	}
	if coderNIK == "" {
		coderNIK = strings.TrimSpace(l.config.CoderNIK)
	}
	if coderNIK == "" {
		return "", fmt.Errorf("%w: NIK Coder masih kosong untuk user login %q", ErrInputTidakValid, input.Username)
	}
	return coderNIK, nil
}

func (l *Layanan) panggilMetodeWS(ctx context.Context, method string, nomorSEP string, data map[string]any, metadata map[string]any) (ResponseEClaim, error) {
	payloadMetadata := map[string]any{
		"method":    method,
		"nomor_sep": nomorSEP,
	}
	for key, value := range metadata {
		payloadMetadata[key] = value
	}

	payloadData := map[string]any{"nomor_sep": nomorSEP}
	for key, value := range data {
		payloadData[key] = value
	}

	return l.panggilWS(ctx, map[string]any{
		"metadata": payloadMetadata,
		"data":     payloadData,
	})
}

func (l *Layanan) panggilGroupingWS(ctx context.Context, grouper string, stage string, nomorSEP string, data map[string]any) (ResponseEClaim, error) {
	payloadData := map[string]any{"nomor_sep": nomorSEP}
	for key, value := range data {
		payloadData[key] = value
	}

	return l.panggilWS(ctx, map[string]any{
		"metadata": map[string]any{
			"method":  "grouper",
			"stage":   stage,
			"grouper": grouper,
		},
		"data": payloadData,
	})
}

func (l *Layanan) panggilWS(ctx context.Context, payload map[string]any) (ResponseEClaim, error) {
	bodyJSON, err := json.Marshal(payload)
	if err != nil {
		return ResponseEClaim{}, fmt.Errorf("payload E-Klaim tidak valid: %w", err)
	}

	terenkripsi, err := l.enkripsiINA(bodyJSON)
	if err != nil {
		return ResponseEClaim{}, err
	}

	var response *http.Response
	var lastErr error
	for percobaan := 0; percobaan <= l.config.Retry; percobaan++ {
		request, err := http.NewRequestWithContext(ctx, http.MethodPost, l.config.WSURL, strings.NewReader(terenkripsi))
		if err != nil {
			return ResponseEClaim{}, fmt.Errorf("membuat request E-Klaim gagal: %w", err)
		}
		request.Header.Set("Content-Type", "application/json")

		response, lastErr = l.client.Do(request)
		if lastErr == nil {
			break
		}
	}
	if lastErr != nil {
		return ResponseEClaim{}, fmt.Errorf("menghubungi WS E-Klaim gagal: %w", lastErr)
	}
	defer response.Body.Close()

	rawResponse, err := io.ReadAll(response.Body)
	if err != nil {
		return ResponseEClaim{}, fmt.Errorf("membaca response E-Klaim gagal: %w", err)
	}

	body, err := l.uraiResponseWS(rawResponse)
	if err != nil {
		return ResponseEClaim{}, err
	}
	return rapikanResponse(body, response.StatusCode), nil
}

func (l *Layanan) enkripsiINA(data []byte) (string, error) {
	kunci, err := kunciHex(l.config.Key)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(kunci)
	if err != nil {
		return "", fmt.Errorf("membuat cipher E-Klaim gagal: %w", err)
	}

	iv := make([]byte, aes.BlockSize)
	if _, err := rand.Read(iv); err != nil {
		return "", fmt.Errorf("membuat IV E-Klaim gagal: %w", err)
	}

	plain := pkcs7Pad(data, aes.BlockSize)
	encrypted := make([]byte, len(plain))
	cipher.NewCBCEncrypter(block, iv).CryptBlocks(encrypted, plain)

	signature := signatureINA(encrypted, kunci)
	return base64.StdEncoding.EncodeToString(append(append(signature, iv...), encrypted...)), nil
}

func (l *Layanan) uraiResponseWS(raw []byte) (map[string]any, error) {
	raw = bytes.TrimSpace(raw)
	var langsung map[string]any
	if err := json.Unmarshal(raw, &langsung); err == nil && langsung != nil {
		return langsung, nil
	}

	kandidat := []string{string(raw)}
	awal := bytes.IndexByte(raw, '\n')
	akhir := bytes.LastIndexByte(raw, '\n')
	if awal >= 0 && akhir > awal {
		kandidat = append(kandidat, strings.TrimSpace(string(raw[awal+1:akhir])))
	}

	var lastErr error
	for _, item := range kandidat {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}

		decrypted, err := l.dekripsiINA(item)
		if err != nil {
			lastErr = err
			continue
		}

		var decoded map[string]any
		if err := json.Unmarshal(decrypted, &decoded); err != nil {
			lastErr = err
			continue
		}
		return decoded, nil
	}
	if lastErr != nil {
		return nil, lastErr
	}
	return nil, fmt.Errorf("response E-Klaim bukan JSON valid")
}

func (l *Layanan) dekripsiINA(data string) ([]byte, error) {
	kunci, err := kunciHex(l.config.Key)
	if err != nil {
		return nil, err
	}

	decoded, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return nil, fmt.Errorf("response E-Klaim bukan base64 valid: %w", err)
	}
	if len(decoded) < 10+aes.BlockSize {
		return nil, fmt.Errorf("response E-Klaim terlalu pendek")
	}

	signature := decoded[:10]
	iv := decoded[10 : 10+aes.BlockSize]
	encrypted := decoded[10+aes.BlockSize:]
	if len(encrypted)%aes.BlockSize != 0 {
		return nil, fmt.Errorf("response E-Klaim bukan blok AES valid")
	}
	if !hmac.Equal(signature, signatureINA(encrypted, kunci)) {
		return nil, fmt.Errorf("signature response E-Klaim tidak cocok")
	}

	block, err := aes.NewCipher(kunci)
	if err != nil {
		return nil, fmt.Errorf("membuat cipher E-Klaim gagal: %w", err)
	}
	plain := make([]byte, len(encrypted))
	cipher.NewCBCDecrypter(block, iv).CryptBlocks(plain, encrypted)

	return pkcs7Unpad(plain, aes.BlockSize)
}

func kunciHex(raw string) ([]byte, error) {
	kunci, err := hex.DecodeString(strings.TrimSpace(raw))
	if err != nil || len(kunci) != 32 {
		return nil, fmt.Errorf("INACBG_KEY harus berupa 64 karakter hex / 256-bit key")
	}
	return kunci, nil
}

func signatureINA(encrypted []byte, kunci []byte) []byte {
	mac := hmac.New(sha256.New, kunci)
	mac.Write(encrypted)
	return mac.Sum(nil)[:10]
}

func pkcs7Pad(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padText...)
}

func pkcs7Unpad(data []byte, blockSize int) ([]byte, error) {
	if len(data) == 0 || len(data)%blockSize != 0 {
		return nil, fmt.Errorf("padding response E-Klaim tidak valid")
	}
	padding := int(data[len(data)-1])
	if padding == 0 || padding > blockSize || padding > len(data) {
		return nil, fmt.Errorf("padding response E-Klaim tidak valid")
	}
	for _, value := range data[len(data)-padding:] {
		if int(value) != padding {
			return nil, fmt.Errorf("padding response E-Klaim tidak valid")
		}
	}
	return data[:len(data)-padding], nil
}

func ambilDaftarDiagnosa(hasil HasilCoding) []Diagnosa {
	if daftar, ok := hasil.Data["diagnosa"].([]Diagnosa); ok {
		return daftar
	}
	return []Diagnosa{}
}

func ambilDaftarProsedur(hasil HasilCoding) []Prosedur {
	if daftar, ok := hasil.Data["prosedur"].([]Prosedur); ok {
		return daftar
	}
	return []Prosedur{}
}

func stringDiagnosa(daftar []Diagnosa) string {
	kode := make([]string, 0, len(daftar))
	for _, item := range daftar {
		if nilai := strings.TrimSpace(item.Kode); nilai != "" {
			kode = append(kode, nilai)
		}
	}
	return strings.Join(kode, "#")
}

func stringProsedur(daftar []Prosedur) string {
	kode := make([]string, 0, len(daftar))
	for _, item := range daftar {
		nilai := strings.TrimSpace(item.Kode)
		if nilai == "" {
			continue
		}
		if item.Multiplicity > 1 {
			nilai = nilai + "+" + strconv.Itoa(item.Multiplicity)
		}
		kode = append(kode, nilai)
	}
	return strings.Join(kode, "#")
}

func nilaiPasien(input InputProses, key string) string {
	if input.Pasien == nil {
		return ""
	}
	return nilaiMap(input.Pasien, key)
}

func nilaiKlaim(input InputProses, key string) string {
	if input.Klaim == nil {
		return ""
	}
	return nilaiMap(input.Klaim, key)
}

func nilaiMap(data map[string]any, key string) string {
	value, exists := data[key]
	if !exists || value == nil {
		return ""
	}
	return strings.TrimSpace(fmt.Sprint(value))
}

func nilaiBoolKlaim(input InputProses, key string) bool {
	if input.Klaim == nil {
		return false
	}
	value, exists := input.Klaim[key]
	if !exists || value == nil {
		return false
	}
	switch nilai := value.(type) {
	case bool:
		return nilai
	case string:
		nilai = strings.ToLower(strings.TrimSpace(nilai))
		return nilai == "1" || nilai == "true" || nilai == "ya" || nilai == "yes"
	case float64:
		return nilai > 0
	case int:
		return nilai > 0
	}
	return false
}

func nilaiIntKlaim(input InputProses, key string, fallback int) int {
	if input.Klaim == nil {
		return fallback
	}
	value, exists := input.Klaim[key]
	if !exists || value == nil {
		return fallback
	}
	return intDariAny(value, fallback)
}

func stringBinerDariBool(value bool) string {
	if value {
		return "1"
	}
	return "0"
}

func fallbackString(value string, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" || value == "<nil>" {
		return fallback
	}
	return value
}

func tanggalJam(tanggal string, jam string) string {
	tanggal = strings.TrimSpace(tanggal)
	if len(tanggal) >= 10 {
		tanggal = tanggal[:10]
	}
	if tanggal == "" || tanggal == "<nil>" {
		tanggal = time.Now().Format("2006-01-02")
	}
	return tanggal + " " + jamLengkap(jam)
}

func jamLengkap(jam string) string {
	jam = strings.TrimSpace(jam)
	if jam == "" || jam == "<nil>" {
		return "00:00:00"
	}
	if len(jam) == 5 {
		return jam + ":00"
	}
	if len(jam) > 8 {
		return jam[:8]
	}
	return jam
}

func tanggalWaktuKlaim(input InputProses, key string) string {
	nilai := nilaiKlaim(input, key)
	if nilai == "" {
		return ""
	}
	nilai = strings.ReplaceAll(nilai, "T", " ")
	if len(nilai) == 16 {
		return nilai + ":00"
	}
	return nilai
}

func genderEklaim(jk string) string {
	if strings.ToUpper(strings.TrimSpace(jk)) == "L" {
		return "1"
	}
	return "2"
}

func jenisRawatEklaim(jenisRawat string) string {
	if strings.ToLower(strings.TrimSpace(jenisRawat)) == "ralan" {
		return "2"
	}
	return "1"
}

func tarifRSPasien(input InputProses) map[string]any {
	if input.Klaim != nil {
		if tarif, ok := input.Klaim["tarif_rs"].(map[string]any); ok && tarif != nil {
			return tarif
		}
	}
	if input.Pasien == nil {
		return tarifKosong()
	}
	if tarif, ok := input.Pasien["tarif_rs"].(map[string]any); ok && tarif != nil {
		return tarif
	}
	return tarifKosong()
}

func rapikanResponse(body map[string]any, statusHTTP int) ResponseEClaim {
	metadata, _ := body["metadata"].(map[string]any)
	kode := statusHTTP
	if rawKode, exists := metadata["code"]; exists {
		kode = intDariAny(rawKode, statusHTTP)
	}

	pesan := "Ok"
	if kode < 200 || kode >= 300 {
		pesan = "Gagal memproses E-Klaim"
	}
	if rawPesan, exists := metadata["message"]; exists {
		pesan = fmt.Sprint(rawPesan)
	}

	errorNo := metadata["error_no"]
	duplikasiSEP := fmt.Sprint(errorNo) == "E2007"
	sukses := (statusHTTP >= 200 && statusHTTP < 300 && kode >= 200 && kode < 300) || duplikasiSEP
	if duplikasiSEP {
		pesan = "Nomor SEP sudah ada di E-Klaim. Silakan lanjutkan ke Set Data Klaim."
	}

	data := body["data"]
	if response, exists := body["response"]; exists {
		data = response
	}

	duplicate := any([]any{})
	if value, exists := body["duplicate"]; exists {
		duplicate = value
	}

	return ResponseEClaim{
		Sukses:    sukses,
		Kode:      kode,
		Status:    statusHTTP,
		Pesan:     pesan,
		ErrorNo:   errorNo,
		Metadata:  metadata,
		Data:      data,
		Response:  body["response"],
		Duplicate: duplicate,
		Raw:       body,
	}
}

func intDariAny(value any, fallback int) int {
	switch nilai := value.(type) {
	case float64:
		return int(nilai)
	case int:
		return nilai
	case string:
		angka, err := strconv.Atoi(nilai)
		if err == nil {
			return angka
		}
	}
	return fallback
}

func hasilProsesTunggal(input InputProses, pesan string, nama string, hasil ResponseEClaim) HasilProses {
	return hasilProsesTahapan(input, pesan, []TahapanProses{{Nama: nama, Hasil: hasil}})
}

func hasilProsesTahapan(input InputProses, pesan string, tahapan []TahapanProses) HasilProses {
	sukses := true
	var hasilTerakhir *ResponseEClaim
	for i := range tahapan {
		hasilTerakhir = &tahapan[i].Hasil
		if !tahapan[i].Hasil.Sukses {
			sukses = false
		}
	}

	return HasilProses{
		Sukses:  sukses,
		Aksi:    input.Aksi,
		NoRawat: input.NoRawat,
		NoSEP:   input.NoSEP,
		Pesan:   pesan,
		Tahapan: tahapan,
		Hasil:   hasilTerakhir,
	}
}
