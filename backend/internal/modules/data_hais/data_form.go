package data_hais

import (
	"encoding/json"
	"fmt"
)

// DataForm menerima angka JSON hanya untuk indikator numerik HAIs.
// Semua nilai tetap divalidasi sebelum diteruskan ke database.
type DataForm map[string]string

func (d *DataForm) UnmarshalJSON(data []byte) error {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return fmt.Errorf("%w: data harus berupa objek form", ErrValidasi)
	}
	if fields == nil {
		*d = nil
		return nil
	}
	hasil := make(DataForm, len(fields))
	for key, raw := range fields {
		var teks string
		if string(raw) != "null" && json.Unmarshal(raw, &teks) == nil {
			hasil[key] = teks
			continue
		}
		angka := false
		for _, kolom := range Kolom()[1:13] {
			if key == kolom {
				angka = true
				break
			}
		}
		var nilai json.Number
		if angka && string(raw) != "null" && json.Unmarshal(raw, &nilai) == nil {
			hasil[key] = nilai.String()
			continue
		}
		// Jangan menampilkan isi payload pasien di pesan kesalahan.
		return fmt.Errorf("%w: tipe isian form tidak sesuai; indikator harus angka atau teks angka, isian lainnya harus teks", ErrValidasi)
	}
	*d = hasil
	return nil
}
