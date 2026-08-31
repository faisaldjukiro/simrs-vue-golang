import { computed, nextTick, ref, watch } from "vue"
import { cariCodingDiagnosaPasien, diagnosaPasienData, hapusCodingDiagnosaPasien, simpanDiagnosaPasien } from "../../../lib/faisal/api"
import { useNotifikasi } from "../../../lib/shared/useNotifikasi"

export function useDiagnosaPasien(props) {
  const notifikasi = useNotifikasi()
  const loading = ref(false)
  const saving = ref(false)
  const deletingKey = ref('')
  const billingLocked = ref(false)
  const diagnosa = ref([])
  const prosedur = ref([])
  const diagnosaBaru = ref({})
  const prosedurBaru = ref({})
  
  const statusPerawatan = computed(() => props.moduleName === 'Rawat Inap' ? 'Ranap' : 'Ralan')
  const jumlahCoding = computed(() => diagnosa.value.length + prosedur.value.length)
  const jumlahBaru = computed(() => [...diagnosa.value, ...prosedur.value].filter((item) => !item.tersimpan).length)
  
  function daftar(jenis) {
    return jenis === 'diagnosa' ? diagnosa.value : prosedur.value
  }
  
  function rapikanPrioritas(items) {
    items.forEach((item, index) => { item.prioritas = index + 1 })
  }
  
  async function cariCoding(jenis, kataKunci) {
    return cariCodingDiagnosaPasien(props.token, {
      jenis,
      q: kataKunci,
      utama: daftar(jenis).length === 0 ? '1' : '0',
    })
  }
  
  async function tambahCoding(jenis, item) {
    if (!item?.kode) return
    if (!item.valid) {
      notifikasi.peringatan(item.pesan || `Kode ${item.kode} tidak dapat digunakan.`)
      await nextTick()
      if (jenis === 'diagnosa') diagnosaBaru.value = {}
      else prosedurBaru.value = {}
      return
    }
    const items = daftar(jenis)
    if (items.some((entry) => entry.kode === item.kode)) {
      notifikasi.peringatan(`Kode ${item.kode} sudah ada dalam daftar.`, 'Duplikat Coding')
    } else {
      items.push({ ...item, prioritas: items.length + 1, jumlah: jenis === 'prosedur' ? 1 : undefined, tersimpan: false })
    }
    await nextTick()
    if (jenis === 'diagnosa') diagnosaBaru.value = {}
    else prosedurBaru.value = {}
  }
  
  async function hapusCoding(jenis, index) {
  	const items = daftar(jenis)
  	const item = items[index]
  	if (!item?.tersimpan) {
  		items.splice(index, 1)
  		rapikanPrioritas(items)
  		return
  	}
  	deletingKey.value = `${jenis}-${item.kode}`
  	try {
  		const response = await hapusCodingDiagnosaPasien(props.token, {
  			noRawat: props.patient.no_rawat,
  			status: statusPerawatan.value,
  			jenis,
  			kode: item.kode,
  		})
  		notifikasi.sukses(response?.pesan || `${jenis === 'diagnosa' ? 'Diagnosa' : 'Prosedur'} berhasil dihapus.`)
  		await muat()
  	} catch (error) {
  		notifikasi.gagal(error.message || 'Coding pasien gagal dihapus.')
  	} finally {
  		deletingKey.value = ''
  	}
  }
  
  function pindahCoding(jenis, index, arah) {
  	const items = daftar(jenis)
  	if (items[index]?.tersimpan) return
  	const tujuan = index + arah
  	if (tujuan < 0 || tujuan >= items.length) return
  	if (items[tujuan]?.tersimpan) return
    const item = items[index]
    if (tujuan === 0 && item.boleh_utama === false) {
      notifikasi.peringatan(`Kode ${item.kode} tidak dapat dijadikan coding utama (accpdx N).`)
      return
    }
    items.splice(index, 1)
    items.splice(tujuan, 0, item)
    rapikanPrioritas(items)
  }
  
  async function muat() {
    if (!props.patient?.no_rawat) return
    loading.value = true
    try {
      const data = await diagnosaPasienData(props.token, {
        no_rawat: props.patient.no_rawat,
        status: statusPerawatan.value,
      })
  		diagnosa.value = Array.isArray(data?.diagnosa) ? data.diagnosa.map((item) => ({ ...item, tersimpan: true })) : []
  		prosedur.value = Array.isArray(data?.prosedur) ? data.prosedur.map((item) => ({ ...item, jumlah: Number(item.jumlah) || 1, tersimpan: true })) : []
      billingLocked.value = Boolean(data?.billing_terkunci)
    } catch (error) {
      notifikasi.gagal(error.message || 'Diagnosa pasien tidak dapat dibaca.')
    } finally {
      loading.value = false
    }
  }
  
  async function simpan() {
    if (billingLocked.value) {
      notifikasi.peringatan('Billing sudah terverifikasi. Diagnosa dan prosedur hanya dapat dilihat.')
      return
    }
  	const diagnosaBaru = diagnosa.value.filter((item) => !item.tersimpan)
  	const prosedurBaru = prosedur.value.filter((item) => !item.tersimpan)
  	if (diagnosaBaru.length === 0 && prosedurBaru.length === 0) {
  		notifikasi.peringatan('Pilih minimal satu diagnosa atau prosedur baru.')
  		return
  	}
    saving.value = true
    try {
      const response = await simpanDiagnosaPasien(props.token, {
        no_rawat: props.patient.no_rawat,
        status: statusPerawatan.value,
  			diagnosa: diagnosaBaru.map(({ kode }) => ({ kode })),
  			prosedur: prosedurBaru.map(({ kode, jumlah }) => ({ kode, jumlah: Number(jumlah) || 1 })),
      })
  		notifikasi.sukses(response?.pesan || 'Diagnosa dan prosedur baru berhasil disimpan.')
      await muat()
    } catch (error) {
      notifikasi.gagal(error.message || 'Diagnosa dan prosedur pasien gagal disimpan.')
    } finally {
      saving.value = false
    }
  }
  
  watch([() => props.patient.no_rawat, statusPerawatan], muat, { immediate: true })
  return {
    loading,
    saving,
    deletingKey,
    billingLocked,
    diagnosa,
    prosedur,
    diagnosaBaru,
    prosedurBaru,
    jumlahCoding,
    jumlahBaru,
    cariCoding,
    tambahCoding,
    hapusCoding,
    pindahCoding,
    simpan,
  }
}
