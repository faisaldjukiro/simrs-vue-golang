# Catatan Integrasi E-Klaim IDRG

Sumber referensi: collection Postman `E-KLAIM IDRG.postman_collection.json` dan manual Web Service E-Klaim 5.10.x yang dikirim user.

## Urutan utama

```text
1. new_claim
2. set_claim_data
3. idrg_diagnosa_set
4. idrg_procedure_set
5. grouper stage 1, grouper idrg
6. grouper stage 2, grouper idrg, jika ada topup_codes
7. idrg_grouper_final
8. idrg_to_inacbg_import
9. inacbg_diagnosa_set
10. inacbg_procedure_set
11. grouper stage 1, grouper inacbg
12. grouper stage 2, grouper inacbg, jika ada special CMG
13. inacbg_grouper_final
14. claim_final
15. send_claim_individual
16. claim_print
```

## Payload yang sudah dipakai backend Go

### new_claim

```json
{
  "metadata": {
    "method": "new_claim"
  },
  "data": {
    "nomor_kartu": "...",
    "nomor_sep": "...",
    "nomor_rm": "...",
    "nama_pasien": "...",
    "tgl_lahir": "YYYY-MM-DD 00:00:00",
    "gender": "1 atau 2"
  }
}
```

### set_claim_data

`coder_nik` diambil dari user login:

```text
username login -> SIMRS.inacbg_coder_nik.nik -> no_ik
```

Jika tidak ditemukan, backend memakai fallback `INACBG_CODER_NIK` dari `.env`.

Payload utama:

```json
{
  "metadata": {
    "method": "set_claim_data",
    "nomor_sep": "..."
  },
  "data": {
    "nomor_sep": "...",
    "nomor_kartu": "...",
    "tgl_masuk": "YYYY-MM-DD 08:00:00",
    "tgl_pulang": "YYYY-MM-DD 09:00:00",
    "cara_masuk": "gp",
    "jenis_rawat": "1 atau 2",
    "kelas_rawat": "1/2/3/vip/vvip",
    "adl_sub_acute": "0",
    "adl_chronic": "0",
    "icu_indikator": "0",
    "icu_los": "0",
    "ventilator_hour": "0",
    "ventilator": {
      "use_ind": "0",
      "start_dttm": "",
      "stop_dttm": ""
    },
    "upgrade_class_ind": "0",
    "add_payment_pct": "0",
    "birth_weight": "0",
    "sistole": 110,
    "diastole": 60,
    "discharge_status": "1",
    "tarif_rs": {},
    "nomor_kartu_t": "kartu_jkn",
    "dializer_single_use": "0",
    "kantong_darah": 0,
    "alteplase_ind": 0,
    "tarif_poli_eks": "0",
    "nama_dokter": "...",
    "kode_tarif": "BP",
    "payor_id": "3",
    "payor_cd": "JKN",
    "cob_cd": 0,
    "coder_nik": "..."
  }
}
```

### IDRG coding dan grouping

```json
{
  "metadata": {
    "method": "idrg_diagnosa_set",
    "nomor_sep": "..."
  },
  "data": {
    "nomor_sep": "...",
    "diagnosa": "S73.02#S87.9#E11.9"
  }
}
```

```json
{
  "metadata": {
    "method": "idrg_procedure_set",
    "nomor_sep": "..."
  },
  "data": {
    "nomor_sep": "...",
    "procedure": "81.53#86.28+2"
  }
}
```

```json
{
  "metadata": {
    "method": "grouper",
    "stage": "1",
    "grouper": "idrg"
  },
  "data": {
    "nomor_sep": "..."
  }
}
```

## Catatan keamanan database

Untuk saat ini backend Go masih menjaga koneksi SIMRS lama sebagai read-only untuk proses query. Log hasil WS E-Klaim belum ditulis ke tabel SIMRS lama seperti `inacbg_data` / `inacbg_klaim_baru2`, kecuali nanti user menyetujui penulisan data ke SIMRS lama.
