import { computed, onMounted, reactive, ref } from "vue";
import {
  cariBangsalPenggunaanBed,
  laporanPenggunaanBedData,
} from "../../../lib/faisal/api";
import { useNotifikasi } from "../../../lib/shared/useNotifikasi";

function hariIni() {
  return new Date().toLocaleDateString("en-CA");
}
export function usePenggunaanBed(props: { token: string }) {
  const loading = ref(false),
    error = ref(""),
    data = ref<any[]>([]),
    ringkasan = ref<any>({});
  const filter = reactive({
    tanggal_mulai: hariIni(),
    tanggal_selesai: hariIni(),
  });
  const bangsal = ref<any>({});
  const pencarian = ref("");
  const notifikasi = useNotifikasi();
  const dataTersaring = computed(() => {
    const kataKunci = pencarian.value.trim().toLowerCase();
    if (!kataKunci) return data.value;

    return data.value.filter((item) =>
      `${item.kode_bangsal} ${item.nama_bangsal}`
        .toLowerCase()
        .includes(kataKunci),
    );
  });
  const grafik = computed(() => {
    const max = Math.max(...data.value.map((x) => x.frekuensi), 1);
    return data.value.map((x) => ({ ...x, persen: (x.frekuensi / max) * 100 }));
  });
  async function muat() {
    loading.value = true;
    error.value = "";
    try {
      const hasil = await laporanPenggunaanBedData(props.token, {
        ...filter,
        bangsal: bangsal.value?.nama || "",
      });
      data.value = hasil?.data || [];
      ringkasan.value = hasil?.ringkasan || {};
      notifikasi.sukses(
        `${data.value.length} bangsal berhasil ditarik.`,
        "Data berhasil dimuat",
      );
    } catch (e: any) {
      error.value = e.message;
      notifikasi.gagal(e.message);
    } finally {
      loading.value = false;
    }
  }
  function excel() {
    if (!data.value.length)
      return notifikasi.peringatan("Tidak ada data untuk diekspor.");
    const rows = data.value
      .map(
        (x, i) =>
          `<tr><td>${i + 1}</td><td>${x.kode_bangsal}</td><td>${x.nama_bangsal}</td><td>${x.total_bed}</td><td>${x.pasien_keluar}</td><td>${x.frekuensi}</td></tr>`,
      )
      .join("");
    const html = `<meta charset="UTF-8"><h2>Penggunaan Bed & Frekuensi</h2><table border="1"><tr><th>No</th><th>Kode Bangsal</th><th>Nama Bangsal</th><th>Total Bed Tersedia</th><th>Pasien Keluar</th><th>Frekuensi per Bed</th></tr>${rows}</table>`;
    const a = document.createElement("a");
    a.href = URL.createObjectURL(
      new Blob(["\ufeff" + html], { type: "application/vnd.ms-excel" }),
    );
    a.download = `penggunaan-bed-${filter.tanggal_mulai}-${filter.tanggal_selesai}.xls`;
    a.click();
    URL.revokeObjectURL(a.href);
  }
  onMounted(muat);
  return {
    loading,
    error,
    data,
    ringkasan,
    filter,
    bangsal,
    pencarian,
    dataTersaring,
    grafik,
    muat,
    excel,
    cariBangsal: (q: string) => cariBangsalPenggunaanBed(props.token, q),
    cetak: () => window.print(),
  };
}
