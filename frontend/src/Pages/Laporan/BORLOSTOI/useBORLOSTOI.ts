import { computed, onMounted, reactive, ref } from "vue";
import { laporanBORLOSTOIData } from "../../../lib/faisal/api";
import { useNotifikasi } from "../../../lib/shared/useNotifikasi";

function bulanSekarang() {
  return new Date().toLocaleDateString("en-CA").slice(0, 7);
}

export function useBORLOSTOI(props: { token: string }) {
  const loading = ref(false);
  const error = ref("");
  const data = ref<any[]>([]);
  const ringkasan = ref<any>({});
  const pencarian = ref("");
  const filter = reactive({
    bulan_mulai: `${new Date().getFullYear()}-01`,
    bulan_selesai: bulanSekarang(),
  });
  const notifikasi = useNotifikasi();

  const dataTersaring = computed(() => {
    const kataKunci = pencarian.value.trim().toLowerCase();
    if (!kataKunci) return data.value;
    return data.value.filter((item) =>
      `${item.periode} ${item.tahun}`.toLowerCase().includes(kataKunci),
    );
  });

  async function muat() {
    loading.value = true;
    error.value = "";
    try {
      const hasil = await laporanBORLOSTOIData(props.token, filter);
      data.value = hasil?.data || [];
      ringkasan.value = hasil?.ringkasan || {};
      notifikasi.sukses(
        `${data.value.length} periode berhasil ditarik.`,
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
    if (!data.value.length) {
      notifikasi.peringatan("Tidak ada data untuk diekspor.");
      return;
    }
    const baris = data.value
      .map(
        (item, index) => `<tr>
          <td>${index + 1}</td>
          <td>${item.periode}</td>
          <td>${item.total_tempat_tidur}</td>
          <td>${item.jumlah_hari}</td>
          <td>${item.total_hari_perawatan}</td>
          <td>${item.pasien_keluar}</td>
          <td>${item.bor}%</td>
          <td>${item.los}</td>
          <td>${item.toi}</td>
        </tr>`,
      )
      .join("");
    const html = `<meta charset="UTF-8"><h2>Laporan BOR, LOS & TOI</h2><table border="1"><tr><th>No.</th><th>Periode</th><th>Tempat Tidur</th><th>Hari</th><th>Hari Perawatan</th><th>Pasien Keluar</th><th>BOR (%)</th><th>LOS (hari)</th><th>TOI (hari)</th></tr>${baris}</table>`;
    const tautan = document.createElement("a");
    tautan.href = URL.createObjectURL(
      new Blob(["\ufeff" + html], { type: "application/vnd.ms-excel" }),
    );
    tautan.download = `bor-los-toi-${filter.bulan_mulai}-${filter.bulan_selesai}.xls`;
    tautan.click();
    URL.revokeObjectURL(tautan.href);
  }

  onMounted(muat);
  return {
    loading,
    error,
    data,
    ringkasan,
    pencarian,
    filter,
    dataTersaring,
    muat,
    excel,
    cetak: () => window.print(),
  };
}
