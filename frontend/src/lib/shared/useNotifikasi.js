import { useToast } from 'primevue/usetoast'

export function useNotifikasi() {
  const toast = useToast()

  function tampilkan(severity, detail, summary, life = 2800) {
    toast.add({
      group: 'simrs',
      severity,
      summary,
      detail,
      life,
    })
  }

  return {
    sukses: (detail, summary = 'Berhasil') => tampilkan('success', detail, summary),
    info: (detail, summary = 'Informasi') => tampilkan('info', detail, summary),
    peringatan: (detail, summary = 'Perhatian') => tampilkan('warn', detail, summary),
    gagal: (detail, summary = 'Gagal') => tampilkan('error', detail, summary),
  }
}
