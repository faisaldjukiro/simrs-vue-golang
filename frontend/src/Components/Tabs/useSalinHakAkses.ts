import { computed, ref, watch } from 'vue'

interface PenggunaSumber {
  id: number
  username: string
  nama?: string
  permission?: string[]
}

interface PropsSalinAkses {
  pengguna: PenggunaSumber[]
  targetId?: number
  permissionSaatIni: string[]
  disabled?: boolean
}

export function useSalinHakAkses(props: PropsSalinAkses, terapkan: (kode: string[]) => void) {
  const pilihan = ref<Record<string, any>>({})
  const pratinjau = ref<{ nama: string; permission: string[] } | null>(null)
  const sumber = computed(() => props.pengguna.find(
    (user) => String(user.id) === pilihan.value.kode && user.id !== props.targetId,
  ))
  const tambahan = computed(() => pratinjau.value?.permission.filter(
    (kode) => !props.permissionSaatIni.includes(kode),
  ) || [])
  const dicabut = computed(() => props.permissionSaatIni.filter(
    (kode) => !pratinjau.value?.permission.includes(kode),
  ))

  async function cari(kata: string) {
    const q = kata.trim().toLocaleLowerCase('id-ID')
    return props.pengguna
      .filter((user) => user.id !== props.targetId
        && `${user.nama || ''} ${user.username}`.toLocaleLowerCase('id-ID').includes(q))
      .slice(0, 30)
      .map((user) => ({
        kode: String(user.id),
        nama: user.nama || user.username,
        keterangan: `${user.username} · ${user.permission?.includes('*') ? 'Admin / semua akses' : `${user.permission?.length || 0} permission`}`,
      }))
  }

  function siapkan() {
    if (props.disabled || !sumber.value) return
    const kode = [...new Set(sumber.value.permission || [])]
    pratinjau.value = {
      nama: `${sumber.value.nama || sumber.value.username} (${sumber.value.username})`,
      permission: kode.includes('*') ? ['*'] : kode,
    }
  }

  function konfirmasi() {
    if (props.disabled || !pratinjau.value || !pratinjau.value.permission.length) return
    terapkan([...pratinjau.value.permission])
    pratinjau.value = null
    pilihan.value = {}
  }

  watch(() => pilihan.value.kode, () => { pratinjau.value = null })
  watch(() => props.targetId, () => {
    pilihan.value = {}
    pratinjau.value = null
  })

  return { pilihan, pratinjau, sumber, tambahan, dicabut, cari, siapkan, konfirmasi }
}
