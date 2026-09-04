import { computed, onMounted, reactive, ref } from "vue";
import {
  cariReferensiLaporanKunjunganRalan,
  laporanKunjunganRalanData,
} from "../../../lib/faisal/api";
import { useNotifikasi } from "../../../lib/shared/useNotifikasi";

export function useKunjunganRalan(props) {
  const notifikasi = useNotifikasi();
  const loading = ref(false),
    error = ref("");
  const data = ref<any[]>([]);
  const firstRow = ref(0);
  const grafikVisible = ref(false);
  const ringkasan = ref({
    total: 0,
    baru: 0,
    lama: 0,
    laki_laki: 0,
    perempuan: 0,
    berulang: 0,
    tidak_berulang: 0,
  });
  const today = new Date().toISOString().slice(0, 10);
  const filter = reactive({
    jenis: "detail",
    tanggal_mulai: today,
    tanggal_selesai: today,
    status: "",
    poli: "",
    dokter: "",
    penjamin: "",
    kabupaten: "",
    kecamatan: "",
    kelurahan: "",
    q: "",
  });
  const pilihan = reactive<Record<string, Record<string, any>>>({
    poli: {},
    dokter: {},
    penjamin: {},
    kabupaten: {},
    kecamatan: {},
    kelurahan: {},
  });
  const berulang = computed(() => filter.jenis === "berulang");
  const trenHarian = computed(() => {
    if (berulang.value) return [];
    const grup = new Map<string, number>();
    data.value.forEach((item) =>
      grup.set(item.tanggal, (grup.get(item.tanggal) || 0) + 1),
    );
    const daftar = [...grup].map(([label, nilai]) => ({ label, nilai }));
    const maksimum = Math.max(...daftar.map((x) => x.nilai), 1);
    return daftar.map((x) => ({ ...x, persen: (x.nilai / maksimum) * 100 }));
  });
  const kategoriTerbanyak = computed(() => {
    const field = berulang.value ? "status_kunjungan" : "poli";
    const grup = new Map<string, number>();
    data.value.forEach((item) => {
      const nama = item[field] || "Tidak diketahui";
      grup.set(nama, (grup.get(nama) || 0) + 1);
    });
    const daftar = [...grup]
      .map(([label, nilai]) => ({ label, nilai }))
      .sort((a, b) => b.nilai - a.nilai);
    const maksimum = Math.max(...daftar.map((x) => x.nilai), 1);
    return daftar.map((x) => ({ ...x, persen: (x.nilai / maksimum) * 100 }));
  });
  function peringkat(field: string, pisahkan = false) {
    const grup = new Map<string, number>();
    data.value.forEach((item) => {
      const nilai = String(item[field] || "").trim();
      const daftar = pisahkan
        ? nilai
            .split(",")
            .map((x) => x.trim())
            .filter(Boolean)
        : [nilai || "Tidak diketahui"];
      new Set(daftar).forEach((nama) =>
        grup.set(nama, (grup.get(nama) || 0) + 1),
      );
    });
    const daftar = [...grup]
      .map(([label, nilai]) => ({ label, nilai }))
      .sort((a, b) => b.nilai - a.nilai);
    const maksimum = Math.max(...daftar.map((x) => x.nilai), 1);
    return daftar.map((x) => ({ ...x, persen: (x.nilai / maksimum) * 100 }));
  }
  const penyakitTerbanyak = computed(() => peringkat("diagnosa", true));
  const dokterTerbanyak = computed(() => peringkat("dokter"));
  const persenGender = computed(() => {
    const total = ringkasan.value.laki_laki + ringkasan.value.perempuan;
    return total ? Math.round((ringkasan.value.laki_laki / total) * 100) : 0;
  });
  const persenBaru = computed(() => {
    const total = ringkasan.value.baru + ringkasan.value.lama;
    return total ? Math.round((ringkasan.value.baru / total) * 100) : 0;
  });
  async function muat() {
    loading.value = true;
    error.value = "";
    firstRow.value = 0;
    try {
      const hasil = await laporanKunjunganRalanData(props.token, {
        ...filter,
        poli: pilihan.poli.nama || "",
        dokter: pilihan.dokter.nama || "",
        penjamin: pilihan.penjamin.nama || "",
        kabupaten: pilihan.kabupaten.nama || "",
        kecamatan: pilihan.kecamatan.nama || "",
        kelurahan: pilihan.kelurahan.nama || "",
      });
      data.value = hasil?.data || [];
      ringkasan.value = hasil?.ringkasan || ringkasan.value;
      notifikasi.sukses(
        `${data.value.length} data kunjungan berhasil ditarik.`,
        "Laporan berhasil dimuat",
      );
    } catch (e: any) {
      error.value = e.message || "Laporan tidak dapat dibaca.";
      notifikasi.gagal(error.value);
    } finally {
      loading.value = false;
    }
  }
  function gantiHalaman(event: any) {
    firstRow.value = event.first || 0;
  }
  function cariReferensi(jenis: string, kataKunci: string) {
    return cariReferensiLaporanKunjunganRalan(props.token, jenis, kataKunci);
  }
  function gantiJenis(jenis: string) {
    filter.jenis = jenis;
    muat();
  }
  function excel() {
    if (!data.value.length)
      return notifikasi.peringatan("Tidak ada data untuk diekspor.");
    const label: Record<string, string> = {
      no_rawat: "No. Rawat",
      tanggal: "Tanggal",
      jam: "Jam",
      status_daftar: "Status Daftar",
      no_rm: "No. RM",
      nama_pasien: "Nama Pasien",
      jenis_kelamin: "Jenis Kelamin",
      umur: "Umur",
      alamat: "Alamat",
      kode_diagnosa: "Kode Diagnosis",
      diagnosa: "Diagnosis",
      dokter: "Dokter",
      poli: "Poliklinik",
      penjamin: "Cara Bayar",
      no_sep: "No. SEP",
      tanggal_lahir: "Tanggal Lahir",
      status_kunjungan: "Status Kunjungan",
      jumlah_kunjungan: "Jumlah Kunjungan",
    };
    const keys = Object.keys(data.value[0]);
    const aman = (value: any) =>
      String(value ?? "")
        .replaceAll("&", "&amp;")
        .replaceAll("<", "&lt;")
        .replaceAll(">", "&gt;")
        .replaceAll('"', "&quot;");
    const kepala =
      `<th>No.</th>` +
      keys.map((key) => `<th>${aman(label[key] || key)}</th>`).join("");
    const baris = data.value
      .map(
        (row, index) =>
          `<tr><td>${index + 1}</td>${keys.map((key) => `<td>${aman(row[key])}</td>`).join("")}</tr>`,
      )
      .join("");
    const isi = `
      <html xmlns:o="urn:schemas-microsoft-com:office:office"
        xmlns:x="urn:schemas-microsoft-com:office:excel">
        <head>
          <meta charset="UTF-8">
          <style>
            table { border-collapse: collapse; font-family: Arial; font-size: 10pt; }
            th { background: #0f6367; color: #fff; font-weight: bold; }
            th, td { border: 1px solid #94a3b8; padding: 6px; vertical-align: top; }
            tr:nth-child(even) { background: #f1f5f9; }
          </style>
        </head>
        <body>
          <h2>Laporan Kunjungan Rawat Jalan</h2>
          <p>Periode ${aman(filter.tanggal_mulai)} sampai ${aman(filter.tanggal_selesai)}</p>
          <table>
            <thead><tr>${kepala}</tr></thead>
            <tbody>${baris}</tbody>
          </table>
        </body>
      </html>`;
    const blob = new Blob(["\ufeff" + isi], {
      type: "application/vnd.ms-excel;charset=utf-8",
    });
    const a = document.createElement("a");
    a.href = URL.createObjectURL(blob);
    a.download = `laporan-kunjungan-ralan-${filter.tanggal_mulai}-${filter.tanggal_selesai}.xls`;
    a.click();
    URL.revokeObjectURL(a.href);
  }
  function cetak() {
    window.print();
  }
  onMounted(muat);
  return {
    loading,
    error,
    data,
    ringkasan,
    filter,
    pilihan,
    berulang,
    firstRow,
    grafikVisible,
    trenHarian,
    kategoriTerbanyak,
    penyakitTerbanyak,
    dokterTerbanyak,
    persenGender,
    persenBaru,
    muat,
    gantiJenis,
    gantiHalaman,
    cariReferensi,
    excel,
    cetak,
  };
}
