import { useToast } from 'primevue/usetoast'

type TingkatNotifikasi = 'success' | 'info' | 'warn' | 'error'

export function useNotifikasi() {
  const toast = useToast()

  function tampilkan(severity: TingkatNotifikasi, detail: string, summary: string, life = 2800) {
    toast.add({
      group: 'simrs',
      severity,
      summary,
      detail,
      life,
    })
  }

  return {
    sukses: (detail: string, summary = 'Berhasil') => tampilkan('success', detail, summary),
    info: (detail: string, summary = 'Informasi') => tampilkan('info', detail, summary),
    peringatan: (detail: string, summary = 'Perhatian') => tampilkan('warn', detail, summary),
    gagal: (detail: string, summary = 'Gagal') => tampilkan('error', detail, summary),
  }
}
