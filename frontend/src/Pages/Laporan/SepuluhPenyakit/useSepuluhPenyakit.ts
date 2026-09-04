import { computed, onMounted, reactive, ref } from "vue";
import { laporan10PenyakitData } from "../../../lib/faisal/api";
import { useNotifikasi } from "../../../lib/shared/useNotifikasi";

interface Penyakit {
  kode: string;
  nama: string;
  diagnosa_lain: number;
  laki_hidup: number;
  perempuan_hidup: number;
  laki_meninggal: number;
  perempuan_meninggal: number;
  jumlah: number;
}

interface Ringkasan {
  jumlah_penyakit: number;
  jumlah_diagnosa: number;
  laki_laki: number;
  perempuan: number;
  meninggal: number;
}

function tanggalHariIni() {
  const sekarang = new Date();
  const tahun = sekarang.getFullYear();
  const bulan = String(sekarang.getMonth() + 1).padStart(2, "0");
  const tanggal = String(sekarang.getDate()).padStart(2, "0");
  return `${tahun}-${bulan}-${tanggal}`;
}

export function useSepuluhPenyakit(props: { token: string }) {
  const notifikasi = useNotifikasi();
  const loading = ref(false);
  const error = ref("");
  const data = ref<Penyakit[]>([]);
  const grafikVisible = ref(true);
  const hariIni = tanggalHariIni();
  const filter = reactive({
    tanggal_mulai: hariIni,
    tanggal_selesai: hariIni,
    status: "Ralan",
    q: "",
  });
  const ringkasan = ref<Ringkasan>({
    jumlah_penyakit: 0,
    jumlah_diagnosa: 0,
    laki_laki: 0,
    perempuan: 0,
    meninggal: 0,
  });

  const grafik = computed(() => {
    const maksimum = Math.max(...data.value.map((item) => item.jumlah), 1);
    return data.value.map((item) => ({
      ...item,
      persen: (item.jumlah / maksimum) * 100,
    }));
  });

  async function muat() {
    loading.value = true;
    error.value = "";
    try {
      const hasil = await laporan10PenyakitData(props.token, filter);
      data.value = hasil?.data ?? [];
      ringkasan.value = hasil?.ringkasan ?? ringkasan.value;
      notifikasi.sukses(
        `${data.value.length} penyakit berhasil ditarik.`,
        "Laporan berhasil dimuat",
      );
    } catch (err: any) {
      error.value = err.message || "Laporan 10 penyakit tidak dapat dibaca.";
      notifikasi.gagal(error.value);
    } finally {
      loading.value = false;
    }
  }

  function reset() {
    filter.status = "Ralan";
    filter.q = "";
    muat();
  }

  function excel() {
    if (!data.value.length) {
      notifikasi.peringatan("Tidak ada data untuk diekspor.");
      return;
    }

    const baris = data.value
      .map(
        (item, index) => `
          <tr>
            <td>${index + 1}</td>
            <td>${item.kode}</td>
            <td>${item.nama}</td>
            <td>${item.diagnosa_lain}</td>
            <td>${item.laki_hidup}</td>
            <td>${item.perempuan_hidup}</td>
            <td>${item.laki_meninggal}</td>
            <td>${item.perempuan_meninggal}</td>
            <td>${item.jumlah}</td>
          </tr>`,
      )
      .join("");
    const html = `
      <meta charset="UTF-8">
      <h2>Laporan 10 Penyakit</h2>
      <p>Periode ${filter.tanggal_mulai} sampai ${filter.tanggal_selesai}</p>
      <table border="1">
        <tr>
          <th>No.</th><th>Kode</th><th>Nama Penyakit</th><th>Diagnosa Lain</th>
          <th>Laki-laki Hidup</th><th>Perempuan Hidup</th>
          <th>Laki-laki Meninggal</th><th>Perempuan Meninggal</th><th>Jumlah</th>
        </tr>
        ${baris}
      </table>`;
    const blob = new Blob(["\ufeff" + html], {
      type: "application/vnd.ms-excel",
    });
    const tautan = document.createElement("a");
    tautan.href = URL.createObjectURL(blob);
    tautan.download = `laporan-10-penyakit-${filter.tanggal_mulai}-${filter.tanggal_selesai}.xls`;
    tautan.click();
    URL.revokeObjectURL(tautan.href);
  }

  onMounted(muat);

  return {
    loading,
    error,
    data,
    filter,
    ringkasan,
    grafik,
    grafikVisible,
    muat,
    reset,
    excel,
    cetak: () => window.print(),
  };
}
