package ventilator

type Master struct {
	KodeVentilator  string  `json:"kode_ventilator"`
	Nama            string  `json:"nama"`
	Merk            string  `json:"merk"`
	Model           string  `json:"model"`
	NomorSeri       string  `json:"nomor_seri"`
	Ruangan         string  `json:"ruangan"`
	Status          string  `json:"status"`
	TerakhirRawat   *string `json:"tanggal_maintenance_terakhir"`
	BerikutnyaRawat *string `json:"tanggal_maintenance_berikutnya"`
	Keterangan      string  `json:"keterangan"`
}

type Pemakaian struct {
	ID                    uint64  `json:"id"`
	NoRawat               string  `json:"no_rawat"`
	KodeVentilator        string  `json:"kode_ventilator"`
	NamaVentilator        string  `json:"nama_ventilator"`
	TanggalMulai          string  `json:"tanggal_mulai"`
	TanggalSelesai        *string `json:"tanggal_selesai"`
	Ruangan               string  `json:"ruangan"`
	Indikasi              string  `json:"indikasi"`
	JenisJalanNapas       string  `json:"jenis_jalan_napas"`
	UkuranJalanNapas      string  `json:"ukuran_jalan_napas"`
	KedalamanJalanNapas   string  `json:"kedalaman_jalan_napas"`
	Status                string  `json:"status"`
	DokterPenanggungJawab string  `json:"dokter_penanggung_jawab"`
	PetugasPemasangan     string  `json:"petugas_pemasangan"`
	Catatan               string  `json:"catatan"`
}

type Setting struct {
	ID              uint64  `json:"id"`
	IDPemakaian     uint64  `json:"id_pemakaian"`
	Waktu           string  `json:"waktu_setting"`
	Mode            string  `json:"mode"`
	FiO2            float64 `json:"fio2"`
	PEEP            float64 `json:"peep"`
	TidalVolume     float64 `json:"tidal_volume"`
	Frekuensi       float64 `json:"frekuensi_set"`
	PressureControl float64 `json:"pressure_control"`
	PressureSupport float64 `json:"pressure_support"`
	Petugas         string  `json:"petugas"`
	Catatan         string  `json:"catatan"`
}

type Monitoring struct {
	ID           uint64  `json:"id"`
	IDPemakaian  uint64  `json:"id_pemakaian"`
	Waktu        string  `json:"waktu_monitoring"`
	Kesadaran    string  `json:"kesadaran"`
	TekananDarah string  `json:"tekanan_darah"`
	Nadi         int     `json:"nadi"`
	Respirasi    int     `json:"respirasi"`
	Suhu         float64 `json:"suhu"`
	SpO2         float64 `json:"spo2"`
	Tahap        string  `json:"tahap"`
	Petugas      string  `json:"petugas"`
	Catatan      string  `json:"catatan"`
}

type ChecklistVAP struct {
	ID              uint64  `json:"id"`
	IDPemakaian     uint64  `json:"id_pemakaian"`
	Waktu           string  `json:"waktu_checklist"`
	ElevasiKepala   bool    `json:"elevasi_kepala"`
	PerawatanMulut  bool    `json:"perawatan_mulut"`
	Suction         bool    `json:"suction"`
	EvaluasiSedasi  bool    `json:"evaluasi_sedasi"`
	SAT             bool    `json:"sat"`
	SBT             bool    `json:"sbt"`
	PencegahanDVT   bool    `json:"pencegahan_dvt"`
	PencegahanUlkus bool    `json:"pencegahan_ulkus"`
	TekananCuff     float64 `json:"tekanan_cuff"`
	Petugas         string  `json:"petugas"`
	Catatan         string  `json:"catatan"`
}

type DataPasien struct {
	Master          []Master       `json:"master"`
	Pemakaian       []Pemakaian    `json:"pemakaian"`
	Setting         []Setting      `json:"setting"`
	Monitoring      []Monitoring   `json:"monitoring"`
	ChecklistVAP    []ChecklistVAP `json:"checklist_vap"`
	BillingTerkunci bool           `json:"billing_terkunci"`
}
