import { computed, onBeforeUnmount, reactive, ref } from 'vue'

export function useBuktiEdukasi(
  form: Record<string, string>,
  urlTersimpan: (key: 'foto' | 'foto_penerima') => string,
  terkunci: () => boolean, gagal: (pesan: string) => void,
) {
  const jenis = ref('')
  const kameraAktif = ref('')
  function bidangFoto(key: 'foto' | 'foto_penerima') {
    const file = ref<File | null>(null)
    const lokal = ref('')
    const ada = computed(() => !!file.value || (!!form[key] && form[key] !== '-'))
    const pratinjau = computed(() => lokal.value || (form[key] && form[key] !== '-' ? urlTersimpan(key) : ''))
    function batal() {
      if (lokal.value) URL.revokeObjectURL(lokal.value)
      lokal.value = ''
      file.value = null
    }
    function tangkap(hasil: File) {
      if (terkunci() || jenis.value !== 'foto') return
      if (!['image/jpeg', 'image/png'].includes(hasil.type) || !hasil.size || hasil.size > 10 * 1024 * 1024) {
        gagal('Foto kamera harus JPG/PNG, maksimal 10 MB.'); return
      }
      batal()
      file.value = hasil
      lokal.value = URL.createObjectURL(hasil)
      gagal('')
    }
    onBeforeUnmount(batal)
    return reactive({ file, ada, pratinjau, batal, tangkap })
  }
  const fotoPetugas = bidangFoto('foto')
  const fotoPenerima = bidangFoto('foto_penerima')
  const adaFoto = computed(() => fotoPetugas.ada || fotoPenerima.ada)
  const buktiGanda = computed(() => adaFoto.value && !!(form.paraf_petugas || form.paraf_penerima))
  function batalFoto() { fotoPetugas.batal(); fotoPenerima.batal(); kameraAktif.value = '' }
  function pilihJenis(jenisBaru: string) {
    if (terkunci() || !['', 'paraf', 'foto'].includes(jenisBaru)) return
    kameraAktif.value = ''
    jenis.value = jenisBaru
    if (jenisBaru !== 'foto') { batalFoto(); form.foto = ''; form.foto_penerima = '' }
    if (jenisBaru !== 'paraf') { form.paraf_petugas = ''; form.paraf_penerima = '' }
  }
  function reset() {
    batalFoto()
    jenis.value = ''
    form.foto = ''
    form.foto_penerima = ''
    form.paraf_petugas = ''
    form.paraf_penerima = ''
  }
  function muatJenis() { jenis.value = buktiGanda.value ? '' : form.paraf_petugas || form.paraf_penerima ? 'paraf' : adaFoto.value ? 'foto' : '' }
  return { jenis, kameraAktif, fotoPetugas, fotoPenerima, adaFoto, buktiGanda, batalFoto, pilihJenis, reset, muatJenis }
}
