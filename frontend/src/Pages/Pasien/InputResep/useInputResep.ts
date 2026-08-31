// @ts-nocheck -- migrasi TypeScript bertahap; kontrak data modul lama belum sepenuhnya bertipe.
import { computed, onMounted, reactive, ref, watch } from "vue"
import { hapusResep, resepCariDepo, resepCariDokter, resepCariObat, resepDaftar, resepDepoDefault, resepDetail, resepInfoPasien, resepMetodeRacik, resepNomorAuto, simpanResep } from "../../../lib/faisal/api"
import { useNotifikasi } from "../../../lib/shared/useNotifikasi"

export function useInputResep(props) {
  const notifikasi = useNotifikasi()
  const loading = ref(false)
  const saving = ref(false)
  const locked = ref(false)
  const recipes = ref([])
  const methods = ref([])
  const doctor = ref({})
  const depoGudang = ref({})
  const medicine = ref({})
  const keywordObat = ref('')
  const hasilObat = ref([])
  const mencariObat = ref(false)
  const pilihanObat = reactive({})
  const editing = ref(false)
  const formVisible = ref(true)
  const tab = ref('resep')
  const copyTerakhir = ref(0)
  const form = reactive({ no_resep: '', tgl_peresepan: today(), jam: now(), no_rawat: props.patient?.no_rawat || '', kd_dokter: '', status: 'ralan', kd_bangsal: '', nm_bangsal: '', judul: '' })
  const items = ref([])
  const racikan = ref([])
  const racikDraft = reactive({ nama_racik: '', kd_racik: '', jml_dr: 1, aturan_pakai: '', keterangan: '', detail: [] })
  const racikMedicine = ref({})
  
  function today(){const d=new Date();return new Date(d.getTime()-d.getTimezoneOffset()*60000).toISOString().slice(0,10)}
  function now(){return new Date().toTimeString().slice(0,8)}
  function rupiah(v){return new Intl.NumberFormat('id-ID',{style:'currency',currency:'IDR',maximumFractionDigits:0}).format(Number(v||0))}
  function tanggalPendek(value){
    if(!value)return '-'
    const clean = String(value).slice(0,10)
    const [y,m,d] = clean.split('-')
    return y && m && d ? `${d}/${m}/${y}` : value
  }
  function jamPendek(value){return value ? String(value).slice(0,5) : '-'}
  function labelStatusResep(value){
    const status = String(value || '').toLowerCase()
    if(status === 'ranap') return 'Rawat Inap'
    if(status === 'ralan') return 'Rawat Jalan'
    return value || '-'
  }
  function statusRawat(){return props.moduleName === 'Rawat Inap' ? 'ranap' : 'ralan'}
  function angka(v, fallback = 0){const n=Number(v);return Number.isFinite(n)?n:fallback}
  function angkaRacik(v, fallback = 0){
    const text = String(v ?? '').replace(',', '.').replace(/[^\d.%.-]/g, '')
    if(text.includes('%')) return text
    const n = Number(text)
    return Number.isFinite(n) ? n : fallback
  }
  function angkaTampil(v, digit = 3){
    const n = Number(v)
    if(!Number.isFinite(n)) return ''
    return Number(n.toFixed(digit)).toString()
  }
  function normalisasiItemResep(x){
    return {
      ...x,
      jml: angka(x.jml ?? x.jumlah ?? x.jumlah_obat, 1),
      harga: angka(x.harga, 0),
      aturan_pakai: x.aturan_pakai || '',
    }
  }
  function normalisasiDetailRacik(x){
    return {
      ...x,
      jml: angka(x.jml ?? x.jumlah, 0),
      harga: angka(x.harga, 0),
      h_beli: angka(x.h_beli, 0),
      stok: angka(x.stok, 0),
      kapasitas: angka(x.kapasitas, 0),
      p1: angka(x.p1, 1),
      p2: angka(x.p2, 1),
      kandungan: x.kandungan || '',
      jenis: x.jenis || '',
      nama_industri: x.nama_industri || '',
      letak_barang: x.letak_barang || '',
    }
  }
  function normalisasiRacikanCopy(x){
    return {
      ...x,
      no_racik: '',
      jml_dr: angka(x.jml_dr, 1),
      aturan_pakai: x.aturan_pakai || '',
      keterangan: x.keterangan || '',
      detail: (x.detail || []).map(normalisasiDetailRacik),
    }
  }
  function normalisasiMetodeRacik(data){
    return (Array.isArray(data) ? data : [])
      .map(item => {
        const kd = item?.kd_racik ?? item?.kode ?? ''
        const nm = item?.nm_racik ?? item?.nama ?? kd
        return {
          ...item,
          kd_racik: String(kd || ''),
          nm_racik: String(nm || ''),
          kapasitas: angka(item?.kapasitas, 0),
        }
      })
      .filter(item => item.kd_racik && item.nm_racik)
  }
  function namaMetodeRacik(kdRacik){
    return methods.value.find(item => item.kd_racik === kdRacik)?.nm_racik || kdRacik || ''
  }
  const noRacikDraft = computed(() => String(racikan.value.length + 1))
  
  const totalObat = computed(() => items.value.reduce((s,i)=>s+Number(i.jml||0)*Number(i.harga||0),0))
  const totalRacikan = computed(() => racikan.value.reduce((sum,r)=>sum+(r.detail||[]).reduce((s,i)=>s+Number(i.jml||0)*Number(i.harga||0),0),0))
  const total = computed(() => totalObat.value + totalRacikan.value)
  const ppn = computed(() => Math.round(total.value * 0.11))
  const tagihan = computed(() => total.value + ppn.value)
  const cariDokter = q => resepCariDokter(props.token,q)
  const cariDepo = q => resepCariDepo(props.token,q)
  const cariObat = q => resepCariObat(props.token,{q,kd_bangsal:form.kd_bangsal,kelas:props.patient?.kelas||''})
  
  async function reset(){
    formVisible.value = true
    Object.assign(form,{no_resep:'',tgl_peresepan:today(),jam:now(),no_rawat:props.patient?.no_rawat||'',kd_dokter:doctor.value?.kd_dokter||'',status:statusRawat(),judul:''})
    items.value = []
    racikan.value = []
    hasilObat.value = []
    keywordObat.value = ''
    Object.keys(pilihanObat).forEach(k=>delete pilihanObat[k])
    editing.value = false
    tab.value = 'resep'
    try{const nomor=await resepNomorAuto(props.token,form.tgl_peresepan);form.no_resep=nomor?.no_resep||''}catch{form.no_resep=''}
  }
  
  async function load(){
    if(!props.patient?.no_rawat)return
    loading.value=true
    try{
      const [info,list]=await Promise.all([resepInfoPasien(props.token,props.patient.no_rawat),resepDaftar(props.token,props.patient.no_rawat)])
      locked.value=Boolean(info?.billing_terkunci)
      recipes.value=Array.isArray(list)?list:[]
      form.status=statusRawat()
      doctor.value={kd_dokter:props.patient?.kd_dokter||'',nm_dokter:props.patient?.nama_dokter||'',spesialis:props.patient?.spesialis||''}
      form.kd_dokter=doctor.value.kd_dokter
      const [met,depo,num]=await Promise.allSettled([resepMetodeRacik(props.token),resepDepoDefault(props.token,props.patient.no_rawat,statusRawat()),resepNomorAuto(props.token,form.tgl_peresepan)])
      if(met.status === 'fulfilled'){
        methods.value = normalisasiMetodeRacik(met.value)
        if(!methods.value.length) notifikasi.peringatan('Metode racik belum ditemukan di tabel metode_racik SIMRS Khanza.')
      }else{
        methods.value = []
        notifikasi.peringatan(`Metode racik tidak dapat dimuat: ${met.reason?.message || 'cek koneksi SIMRS Khanza'}`)
      }
      if(depo.status==='fulfilled'){
        form.kd_bangsal=depo.value?.kd_bangsal||''
        form.nm_bangsal=depo.value?.nm_bangsal||''
        depoGudang.value={kd_bangsal:form.kd_bangsal,nm_bangsal:form.nm_bangsal}
      }
      if(num.status==='fulfilled')form.no_resep=num.value?.no_resep||''
      await terapkanCopyResep(props.copiedResep)
    }catch(e){
      notifikasi.gagal(e.message)
    }finally{
      loading.value=false
    }
  }
  
  async function terapkanCopyResep(copy){
    if(!copy?.sumber || copyTerakhir.value === copy.copied_at)return
    copyTerakhir.value = copy.copied_at || Date.now()
    formVisible.value = true
    editing.value = false
    tab.value = 'resep'
    form.no_rawat = props.patient?.no_rawat || ''
    form.tgl_peresepan = today()
    form.jam = now()
    form.status = statusRawat()
    form.judul = copy.sumber?.judul ? `Copy ${copy.sumber.judul}` : `Copy resep ${copy.sumber?.no_resep || ''}`
    if(!form.no_resep){
      try{
        const nomor = await resepNomorAuto(props.token, form.tgl_peresepan)
        form.no_resep = nomor?.no_resep || ''
      }catch{
        form.no_resep = ''
      }
    }
    const dokterCopy = {
      kd_dokter: props.patient?.kd_dokter || copy.sumber?.kd_dokter || doctor.value?.kd_dokter || '',
      nm_dokter: props.patient?.nama_dokter || copy.sumber?.nm_dokter || doctor.value?.nm_dokter || '',
      spesialis: props.patient?.spesialis || doctor.value?.spesialis || '',
    }
    doctor.value = dokterCopy
    form.kd_dokter = dokterCopy.kd_dokter
    items.value = (copy.obat || []).map(normalisasiItemResep)
    Object.keys(pilihanObat).forEach(k=>delete pilihanObat[k])
    items.value.forEach(item=>{if(item.kode_brng) pilihanObat[item.kode_brng] = {...item}})
    racikan.value = (copy.racikan || []).map(normalisasiRacikanCopy)
    hasilObat.value = []
    keywordObat.value = ''
    notifikasi.sukses('Resep hasil copy sudah tampil di form. Silakan cek lalu simpan.')
    requestAnimationFrame(()=>document.querySelector('.recipe-form-card')?.scrollIntoView({behavior:'smooth',block:'start'}))
  }
  
  async function cariBanyakObat(){
    const q = keywordObat.value.trim()
    if(!form.kd_bangsal){
      hasilObat.value = []
      if(q.length >= 2) notifikasi.gagal('Depo / gudang wajib dipilih dulu')
      return
    }
    if(q.length < 2){
      hasilObat.value = []
      return
    }
    mencariObat.value = true
    try{
      hasilObat.value = await resepCariObat(props.token,{q,kd_bangsal:form.kd_bangsal,kelas:props.patient?.kelas||''})
    }catch(e){
      hasilObat.value = []
      notifikasi.gagal(e.message)
    }finally{
      mencariObat.value = false
    }
  }
  function obatTerpilih(item){return Boolean(pilihanObat[item.kode_brng])}
  function togglePilihanObat(item){
    if(!item?.kode_brng)return
    if(pilihanObat[item.kode_brng]){
      delete pilihanObat[item.kode_brng]
      const idx = items.value.findIndex(i=>i.kode_brng===item.kode_brng)
      if(idx >= 0) items.value.splice(idx,1)
      return
    }
    if(angka(item.harga, 0) <= 0){
      notifikasi.peringatan('Maaf, harga obat masih 0 sehingga tidak bisa dipilih.')
      return
    }
    const data = {...item,jml:1,aturan_pakai:'',ikhtisar_farmasi:''}
    pilihanObat[item.kode_brng] = data
    if(!items.value.some(i=>i.kode_brng===item.kode_brng)){
      items.value.push({...data})
    }
  }
  function removeMedicine(i){
    const item = items.value[i]
    if(item?.kode_brng) delete pilihanObat[item.kode_brng]
    items.value.splice(i,1)
  }
  function hitungJumlahRacik(detail){
    if(!detail) return
    const kapasitas = angka(detail.kapasitas, 0)
    const kandunganText = String(detail.kandungan ?? '').trim()
    if(!kapasitas || !kandunganText){
      detail.jml = 0
      return
    }
    if(kandunganText.includes('%')){
      const persen = angka(kandunganText.replace('%', ''), 0)
      const totalKandunganLain = racikDraft.detail.reduce((sum, item) => {
        if(item === detail || String(item.kandungan ?? '').includes('%')) return sum
        return sum + (angka(item.kapasitas, 0) * angka(item.jml, 0))
      }, 0)
      detail.jml = kapasitas ? Number(((totalKandunganLain * (persen / 100)) / kapasitas).toFixed(1)) : 0
      return
    }
    detail.jml = Number(((angka(racikDraft.jml_dr, 1) * angka(kandunganText, 0)) / kapasitas).toFixed(1))
  }
  function hitungKandunganDariP(detail){
    if(!detail) return
    const kapasitas = angka(detail.kapasitas, 0)
    const p1 = angka(detail.p1, 0)
    const p2 = angka(detail.p2, 0)
    if(!kapasitas || !p1 || !p2){
      detail.kandungan = ''
      detail.jml = 0
      return
    }
    detail.kandungan = angkaTampil(kapasitas * (p1 / p2), 3)
    hitungJumlahRacik(detail)
  }
  function hitungSemuaDetailRacik(){
    racikDraft.detail.forEach(hitungJumlahRacik)
  }
  function addRacikMedicine(){
    if(!racikMedicine.value?.kode_brng)return
    if(angka(racikMedicine.value.harga, 0) <= 0){
      notifikasi.peringatan('Maaf, harga obat masih 0 sehingga tidak bisa dipilih.')
      return
    }
    const detail = normalisasiDetailRacik({...racikMedicine.value,jml:0,p1:1,p2:1,kandungan:''})
    hitungKandunganDariP(detail)
    racikDraft.detail.push(detail)
    racikMedicine.value={}
  }
  function addRacik(){
    if(!racikDraft.nama_racik||!racikDraft.kd_racik||!racikDraft.detail.length){notifikasi.gagal('Nama, metode, dan obat racikan wajib diisi');return}
    racikan.value.push({...racikDraft,no_racik:noRacikDraft.value,nm_racik:namaMetodeRacik(racikDraft.kd_racik),detail:racikDraft.detail.map(item => ({...item}))})
    Object.assign(racikDraft,{nama_racik:'',kd_racik:'',jml_dr:1,aturan_pakai:'',keterangan:'',detail:[]})
  }
  async function editRecipe(row){
    if(locked.value)return
    const d=await resepDetail(props.token,row.no_resep)
    const kdBangsal = d.header.kd_bangsal || form.kd_bangsal
    const nmBangsal = d.header.nm_bangsal || form.nm_bangsal
    Object.assign(form,{no_resep:d.header.no_resep,tgl_peresepan:d.header.tgl_peresepan,jam:d.header.jam,no_rawat:d.header.no_rawat,kd_dokter:d.header.kd_dokter,status:d.header.status,kd_bangsal:kdBangsal,nm_bangsal:nmBangsal,judul:d.header.judul})
    doctor.value={kd_dokter:d.header.kd_dokter,nm_dokter:d.header.nm_dokter}
    depoGudang.value={kd_bangsal:kdBangsal,nm_bangsal:nmBangsal}
    items.value=(d.obat||[]).map(normalisasiItemResep)
    Object.keys(pilihanObat).forEach(k=>delete pilihanObat[k])
    items.value.forEach(item=>{if(item.kode_brng) pilihanObat[item.kode_brng] = {...item}})
    racikan.value=(d.racikan||[]).map(r=>({...r,jml_dr:angka(r.jml_dr,1),detail:(r.detail||[]).map(normalisasiDetailRacik)}))
    editing.value=true
    formVisible.value=true
    tab.value='resep'
    requestAnimationFrame(()=>document.querySelector('.recipe-form-card')?.scrollIntoView({behavior:'smooth',block:'start'}))
  }
  async function save(){
    if(locked.value)return
    if(!form.kd_dokter){notifikasi.gagal('Dokter peresep wajib dipilih');return}
    if(!items.value.length&&!racikan.value.length){notifikasi.gagal('Tambahkan obat atau racikan');return}
    saving.value=true
    try{
      await simpanResep(props.token,{...form,kd_dokter:form.kd_dokter||doctor.value.kd_dokter,obat:items.value,racikan:racikan.value,is_ubah:editing.value})
      notifikasi.sukses('Resep berhasil disimpan')
      await load()
      await reset()
    }catch(e){
      notifikasi.gagal(e.message)
    }finally{
      saving.value=false
    }
  }
  async function remove(row){
    if(locked.value)return
    if(!confirm(`Hapus resep ${row.no_resep}?`))return
    try{await hapusResep(props.token,row.no_resep);notifikasi.sukses('Resep dihapus');await load()}catch(e){notifikasi.gagal(e.message)}
  }
  
  watch(()=>doctor.value?.kd_dokter,v=>{form.kd_dokter=v||''})
  watch(()=>depoGudang.value, v=>{
    form.kd_bangsal = v?.kd_bangsal || ''
    form.nm_bangsal = v?.nm_bangsal || ''
    hasilObat.value = []
  })
  let timerCariObat
  watch(keywordObat,()=>{
    window.clearTimeout(timerCariObat)
    timerCariObat = window.setTimeout(cariBanyakObat, 300)
  })
  watch(()=>props.copiedResep, (copy)=>terapkanCopyResep(copy))
  watch(()=>racikDraft.jml_dr, hitungSemuaDetailRacik)
  onMounted(load)
  return {
    loading,
    saving,
    locked,
    recipes,
    methods,
    doctor,
    depoGudang,
    keywordObat,
    hasilObat,
    editing,
    formVisible,
    tab,
    form,
    items,
    racikan,
    racikDraft,
    racikMedicine,
    rupiah,
    tanggalPendek,
    jamPendek,
    labelStatusResep,
    noRacikDraft,
    total,
    ppn,
    tagihan,
    cariDokter,
    cariDepo,
    cariObat,
    reset,
    cariBanyakObat,
    obatTerpilih,
    togglePilihanObat,
    removeMedicine,
    hitungJumlahRacik,
    hitungKandunganDariP,
    hitungSemuaDetailRacik,
    addRacikMedicine,
    addRacik,
    editRecipe,
    save,
    remove,
  }
}
