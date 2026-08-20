<script setup>
import { ChevronDown, ChevronUp, ClipboardPlus, FlaskConical, Plus, Save, Trash2, X } from '@lucide/vue'
import { computed, onMounted, reactive, ref, watch } from 'vue'
import FormInput from '../../Components/Ui/FormInput.vue'
import InputPencarian from '../../Components/Ui/InputPencarian.vue'
import {
  hapusResep, resepCariDepo, resepCariDokter, resepCariObat, resepDaftar, resepDepoDefault, resepDetail,
  resepInfoPasien, resepMetodeRacik, resepNomorAuto, simpanResep,
} from '../../lib/faisal/api'
import { useNotifikasi } from '../../lib/shared/useNotifikasi'

const props = defineProps({
  token: { type: String, required: true },
  patient: { type: Object, required: true },
  moduleName: { type: String, default: 'Rawat Inap' },
  copiedResep: { type: Object, default: null },
})
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
function normalisasiItemResep(x){
  return {
    ...x,
    jml: angka(x.jml ?? x.jumlah ?? x.jumlah_obat, 1),
    harga: angka(x.harga, 0),
    aturan_pakai: x.aturan_pakai || '',
  }
}
function normalisasiRacikanCopy(x){
  return {
    ...x,
    no_racik: '',
    jml_dr: angka(x.jml_dr, 1),
    aturan_pakai: x.aturan_pakai || '',
    keterangan: x.keterangan || '',
    detail: (x.detail || []).map(d => ({
      ...d,
      jml: angka(d.jml ?? d.jumlah, 1),
      p1: angka(d.p1, 1),
      p2: angka(d.p2, 1),
      kandungan: d.kandungan || '',
    })),
  }
}

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
    methods.value=met.status==='fulfilled'&&Array.isArray(met.value)?met.value:[]
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
function addRacikMedicine(){if(!racikMedicine.value?.kode_brng)return;racikDraft.detail.push({...racikMedicine.value,jml:1,p1:1,p2:1});racikMedicine.value={}}
function addRacik(){
  if(!racikDraft.nama_racik||!racikDraft.kd_racik||!racikDraft.detail.length){notifikasi.gagal('Nama, metode, dan obat racikan wajib diisi');return}
  racikan.value.push({...racikDraft,detail:[...racikDraft.detail]})
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
  racikan.value=(d.racikan||[]).map(r=>({...r,jml_dr:angka(r.jml_dr,1),detail:(r.detail||[]).map(normalisasiItemResep)}))
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
onMounted(load)
</script>

<template>
  <section class="recipe-page">
    <div v-if="locked" class="recipe-lock">Billing sudah diproses. Resep tidak dapat ditambah, diubah, atau dihapus.</div>

    <article class="recipe-card recipe-form-card">
      <header class="recipe-panel-header">
        <div class="recipe-panel-title">
          <ClipboardPlus :size="16" />
          <span>Peresepan Dokter</span>
        </div>
        <div class="recipe-section-tools">
          <button type="button" class="recipe-template-button" :disabled="locked">
            <ClipboardPlus :size="15" />
            Template Resep
          </button>
          <button class="recipe-toggle" type="button" :title="formVisible ? 'Sembunyikan Form Input' : 'Tampilkan Form Input'" @click="formVisible=!formVisible">
            <ChevronUp v-if="formVisible" :size="18"/>
            <ChevronDown v-else :size="18"/>
          </button>
        </div>
      </header>

      <form v-show="formVisible" class="recipe-form cppt-form" @submit.prevent="save">
        <div class="recipe-grid top-grid">
          <InputPencarian v-model="doctor" label="NIP Dokter" code-field="kd_dokter" name-field="nm_dokter" description-field="spesialis" placeholder="NIP" :search="cariDokter" required :disabled="locked"/>
          <FormInput v-model="doctor.nm_dokter" label="Nama Dokter" placeholder="Nama dokter" disabled/>
          <FormInput v-model="doctor.spesialis" label="Jabatan" placeholder="-" disabled/>
          <FormInput v-model="form.tgl_peresepan" label="Tanggal Peresepan" type="date" required :disabled="locked"/>
          <FormInput v-model="form.jam" label="Jam Peresepan" type="time" step="1" required :disabled="locked"/>
          <FormInput v-model="form.no_resep" label="No. Resep" placeholder="Auto" disabled/>
          <FormInput v-model="form.judul" label="Judul Resep" placeholder="Judul" :disabled="locked"/>
          <InputPencarian
            v-model="depoGudang"
            label="Depo / Gudang"
            code-field="kd_bangsal"
            name-field="nm_bangsal"
            placeholder="Cari depo/gudang..."
            :search="cariDepo"
            required
            :disabled="locked"
          />
          <FormInput v-model="form.status" label="Tarif" jenis="select" :options="[{label:'Rawat Jalan',value:'ralan'},{label:'Rawat Inap',value:'ranap'}]" :disabled="locked"/>
        </div>

        <div class="recipe-total-strip">
          <div><small>Total</small><strong>{{ rupiah(total) }}</strong></div>
          <div><small>PPN 11%</small><strong class="warn">{{ rupiah(ppn) }}</strong></div>
          <div><small>Tagihan</small><strong>{{ rupiah(tagihan) }}</strong></div>
        </div>

        <div class="recipe-tabs">
          <button type="button" :class="{active:tab==='resep'}" @click="tab='resep'"><ClipboardPlus :size="15" />Resep<b>{{ items.length }}</b></button>
          <button type="button" :class="{active:tab==='racikan'}" @click="tab='racikan'"><FlaskConical :size="15" />Racikan<b>{{ racikan.length }}</b></button>
        </div>

        <section v-show="tab === 'resep'" class="recipe-tab-panel">
          <div class="recipe-search recipe-bulk-search">
            <FormInput v-model="keywordObat" label="Cari Obat" placeholder="Cari kode/nama obat..." :disabled="locked" @keydown.enter.prevent="cariBanyakObat"/>
          </div>
          <div v-if="hasilObat.length" class="recipe-pick-list">
            <button v-for="obat in hasilObat" :key="obat.kode_brng" type="button" :class="{selected:obatTerpilih(obat)}" @click="togglePilihanObat(obat)">
              <span class="recipe-check" aria-hidden="true">{{ obatTerpilih(obat) ? '✓' : '' }}</span>
              <span><b>{{ obat.nama_brng }}</b><small>{{ obat.kode_brng }} - {{ obat.kode_sat }} <template v-if="obat.jenis">- {{ obat.jenis }}</template></small></span>
              <strong class="stock" :class="{danger:Number(obat.stok||0)<=0}">Stok {{ obat.stok ?? 0 }}</strong>
              <strong>{{ rupiah(obat.harga) }}</strong>
            </button>
          </div>
          <div class="recipe-items">
            <template v-if="items.length">
              <div class="recipe-row recipe-row-head">
                <span>K</span><span>Jumlah</span><span>Kode Barang</span><span>Nama Barang</span><span>Satuan</span><span>Harga(Rp)</span><span>Jenis Obat</span><span>Aturan Pakai</span><span>H.Beli</span><span>Stok</span><span>Fornas</span><span>#</span>
              </div>
              <div v-for="(item, i) in items" :key="item.kode_brng" class="recipe-row">
                <input v-model="item.pilih" type="checkbox" :disabled="locked"/>
                <input v-model.number="item.jml" class="form-input-element recipe-inline-input" type="number" min="1" :disabled="locked"/>
                <code>{{ item.kode_brng }}</code>
                <strong>{{ item.nama_brng }}</strong>
                <span>{{ item.kode_sat || '-' }}</span>
                <span class="money">{{ rupiah(item.harga) }}</span>
                <span>{{ item.jenis || '-' }}</span>
                <input v-model="item.aturan_pakai" class="form-input-element recipe-inline-input" placeholder="mis: 3×1" :disabled="locked"/>
                <span class="money">{{ rupiah(item.h_beli) }}</span>
                <span :class="{danger:Number(item.stok||0)<=0}">{{ item.stok ?? '-' }}</span>
                <span><i v-if="item.kategori_fornas" class="recipe-badge">{{ item.kategori_fornas }}</i><template v-else>-</template></span>
                <button type="button" class="icon-button danger" :disabled="locked" @click="removeMedicine(i)"><Trash2 :size="15"/></button>
              </div>
            </template>
            <div v-else class="recipe-empty-box">
              <ClipboardPlus :size="26" />
              <span>Belum ada obat. Cari obat menggunakan kolom pencarian di atas.</span>
            </div>
          </div>
        </section>

        <section v-show="tab === 'racikan'" class="recipe-tab-panel">
          <div class="recipe-grid racik-form">
            <FormInput v-model="racikDraft.nama_racik" label="Nama Racikan" :disabled="locked"/>
            <FormInput v-model="racikDraft.jml_dr" label="Jumlah" type="number" :disabled="locked"/>
            <FormInput v-model="racikDraft.aturan_pakai" label="Aturan Pakai" :disabled="locked"/>
            <FormInput v-model="racikDraft.kd_racik" label="Metode Racik" jenis="select" :options="methods" option-label="nm_racik" option-value="kd_racik" :disabled="locked"/>
          </div>
          <div class="recipe-search compact">
            <InputPencarian v-model="racikMedicine" label="Obat Racikan" code-field="kode_brng" name-field="nama_brng" right-field="harga" :right-formatter="rupiah" :search="cariObat" placeholder="Cari obat untuk racikan..." :disabled="locked"/>
            <button type="button" class="recipe-action-button" :disabled="locked||!racikMedicine.kode_brng" @click="addRacikMedicine"><Plus :size="15"/> Tambah</button>
          </div>
          <div v-if="racikDraft.detail.length" class="selected-tags">
            <span v-for="(d,i) in racikDraft.detail" :key="i">{{ d.nama_brng }} <button type="button" @click="racikDraft.detail.splice(i,1)"><X :size="12"/></button></span>
          </div>
          <div class="recipe-racik-actions">
            <FormInput v-model="racikDraft.keterangan" label="Keterangan" :disabled="locked"/>
            <button type="button" class="recipe-action-button" :disabled="locked" @click="addRacik"><Plus :size="15"/> Tambah Racikan</button>
          </div>
          <div v-if="racikan.length" class="recipe-items">
            <div v-for="(r,i) in racikan" :key="i" class="racik-saved"><b>{{ r.nama_racik }}</b><small>{{ r.nm_racik || r.kd_racik }} - {{ r.detail?.length || 0 }} obat - {{ r.aturan_pakai || '-' }}</small></div>
          </div>
          <div v-else class="recipe-empty-box"><FlaskConical :size="26" /><span>Belum ada racikan.</span></div>
        </section>

        <footer class="recipe-actions">
          <button type="submit" class="recipe-action-button" :disabled="locked||saving"><Save :size="16"/> {{ saving ? 'Menyimpan...' : (editing ? 'Simpan Perubahan' : 'Simpan Resep') }}</button>
          <small>Aturan pakai wajib diisi untuk setiap obat.</small>
        </footer>
      </form>
    </article>

    <article class="recipe-card recipe-history-card">
      <header class="recipe-section-header">
        <div><span class="eyebrow">RIWAYAT RESEP</span><h3>Resep Pasien</h3><p>{{ recipes.length }} resep tersimpan</p></div>
        <button type="button" class="recipe-action-button secondary" :disabled="locked" @click="reset"><Plus :size="15"/> Resep Baru</button>
      </header>
      <div v-if="loading" class="empty-state">Memuat resep...</div>
      <div v-else-if="!recipes.length" class="empty-state">Belum ada resep untuk pasien ini.</div>
      <div v-else class="recipe-history">
        <div class="history-head">
          <span>No. Resep</span>
          <span>Tanggal / Dokter</span>
          <span>Isi Resep</span>
          <span>Status</span>
          <span>Total</span>
          <span>Aksi</span>
        </div>
        <div v-for="row in recipes" :key="row.no_resep" class="history-row">
          <div class="history-prescription">
            <b>{{ row.no_resep }}</b>
            <small>{{ row.judul || 'Tanpa judul resep' }}</small>
          </div>
          <div>
            <b>{{ tanggalPendek(row.tgl_peresepan) }} {{ jamPendek(row.jam) }}</b>
            <small>{{ row.nm_dokter || row.kd_dokter || '-' }}</small>
          </div>
          <div>
            <b>{{ row.jumlah_obat || 0 }} obat</b>
            <small>{{ row.jumlah_racik || 0 }} racikan</small>
          </div>
          <span class="recipe-status-pill">{{ labelStatusResep(row.status) }}</span>
          <strong class="history-total">{{ rupiah(row.total) }}</strong>
          <div class="history-actions">
            <button type="button" class="recipe-action-button secondary" :disabled="locked" @click="editRecipe(row)">Edit</button>
            <button type="button" class="icon-button danger" :disabled="locked" @click="remove(row)"><Trash2 :size="15"/></button>
          </div>
        </div>
      </div>
    </article>
  </section>
</template>

<style scoped>
.recipe-page{display:grid;gap:16px;color:var(--text)}
.recipe-card{border:1px solid var(--line);border-radius:12px;background:var(--surface);overflow:hidden}
.recipe-form-card{margin-top:12px}
.recipe-lock{padding:12px 16px;border:1px solid #e6a23c;color:#fbbf24;background:rgba(230,162,60,.12);border-radius:8px}
.recipe-panel-header{display:flex;align-items:center;justify-content:space-between;gap:16px;padding:14px 16px;background:color-mix(in srgb,#0d9488 10%,var(--surface));border-bottom:1px solid var(--line)}
.recipe-panel-title{display:flex;align-items:center;gap:9px;color:#0d9488;font-weight:700}
.recipe-section-tools{display:flex;align-items:center;gap:8px}
.recipe-template-button,.recipe-action-button,.recipe-mini-button,.recipe-toggle{display:inline-flex;align-items:center;justify-content:center;gap:8px;min-height:36px;border:0;border-radius:8px;background:#0d9488;color:#fff;font-weight:700;cursor:pointer;padding:0 14px}
.recipe-template-button{min-width:154px}.recipe-action-button.secondary,.recipe-mini-button{border:1px solid var(--line);background:var(--surface-soft);color:var(--text)}.recipe-toggle{width:36px;padding:0;background:transparent;color:var(--text)}.recipe-template-button:disabled,.recipe-action-button:disabled,.recipe-mini-button:disabled{opacity:.55;cursor:not-allowed}
.recipe-form{padding:34px 16px 16px;background:color-mix(in srgb,#0d9488 4%,var(--surface-soft))}.recipe-grid{display:grid;grid-template-columns:repeat(3,minmax(0,1fr));gap:12px}.top-grid :deep(.reference-search){grid-column:auto}
.recipe-total-strip{display:grid;grid-template-columns:1fr 1fr 1fr;margin-top:14px;border:1px solid var(--line);border-radius:8px;background:var(--surface);overflow:hidden}.recipe-total-strip div{display:flex;flex-direction:column;gap:4px;padding:12px 16px;border-right:1px solid var(--line)}.recipe-total-strip div:last-child{border-right:0;align-items:flex-end}.recipe-total-strip small{text-transform:uppercase;color:var(--muted);font-size:.72rem;font-weight:700}.recipe-total-strip strong{color:#0d9488;font-size:1.02rem}.recipe-total-strip .warn{color:#f59e0b}
.recipe-tabs{display:flex;gap:24px;margin-top:18px;border-bottom:1px solid var(--line)}.recipe-tabs button{display:inline-flex;align-items:center;gap:8px;border:0;border-bottom:2px solid transparent;background:transparent;color:var(--muted);padding:12px 0;font-weight:700;cursor:pointer}.recipe-tabs button.active{border-color:#0d9488;color:#0d9488}.recipe-tabs b{min-width:22px;border-radius:999px;background:rgba(13,148,136,.14);color:#0d9488;font-size:.72rem;padding:2px 7px;text-align:center}
.recipe-tab-panel{padding-top:14px}.recipe-search{display:grid;grid-template-columns:minmax(0,1fr);align-items:end;gap:8px}.recipe-search>.reference-search{min-width:0}.recipe-items{margin-top:14px;border:1px solid var(--line);border-radius:8px;overflow:auto;background:var(--surface)}.recipe-row{display:grid;grid-template-columns:32px 64px 92px minmax(230px,1.8fr) 68px 100px 100px 130px 100px 72px 78px 38px;gap:10px;align-items:center;min-width:1120px;padding:8px 10px;border-top:1px solid var(--line);color:var(--text)}.recipe-row:first-child{border-top:0}.recipe-row-head{position:sticky;top:0;z-index:1;font-size:.68rem;text-transform:uppercase;color:var(--muted);background:var(--surface-soft);font-weight:800}.recipe-row span{min-width:0;overflow:hidden;text-overflow:ellipsis;white-space:nowrap}.recipe-row small,.racik-saved small,.history-row small{color:var(--muted);font-size:.78rem}.recipe-row input{min-width:0}.recipe-row input[type="checkbox"]{width:16px;height:16px;justify-self:center;accent-color:#0d9488}.recipe-row code{overflow:hidden;color:#2563eb;text-overflow:ellipsis;white-space:nowrap}.recipe-row strong{overflow:hidden;text-overflow:ellipsis;white-space:nowrap}.recipe-row .money{font-family:ui-monospace,monospace;text-align:right}.recipe-row .danger{color:#e11d48}.recipe-badge{display:inline-flex;border-radius:999px;padding:3px 7px;color:#0f766e;background:rgba(20,184,166,.14);font-style:normal;font-size:.68rem;font-weight:700}
.recipe-action-button b{display:grid;min-width:20px;height:20px;place-items:center;border-radius:999px;background:rgba(255,255,255,.18);font-size:.72rem}.recipe-pick-list{display:grid;max-height:260px;overflow:auto;margin-top:10px;border:1px solid var(--line);border-radius:10px;background:var(--surface)}.recipe-pick-list button{display:grid;grid-template-columns:28px minmax(0,1fr) 90px 110px;align-items:center;gap:10px;border:0;border-bottom:1px solid var(--line);padding:10px 12px;color:var(--text);background:transparent;text-align:left;cursor:pointer}.recipe-pick-list button:last-child{border-bottom:0}.recipe-pick-list button:hover,.recipe-pick-list button.selected{background:rgba(13,148,136,.1)}.recipe-pick-list span:not(.recipe-check){display:grid;min-width:0;gap:3px}.recipe-pick-list b,.recipe-pick-list small{overflow:hidden;text-overflow:ellipsis;white-space:nowrap}.recipe-pick-list small{color:var(--muted);font-size:.75rem}.recipe-pick-list strong{color:#0d9488;white-space:nowrap}.recipe-pick-list .stock{color:var(--text);font-size:.78rem;text-align:right}.recipe-pick-list .stock.danger{color:#e11d48}.recipe-check{display:grid;width:22px;height:22px;place-items:center;border:1px solid var(--line);border-radius:7px;color:#fff;background:var(--surface-soft);font-weight:800}.recipe-pick-list button.selected .recipe-check{border-color:#0d9488;background:#0d9488}
.recipe-empty-box{display:grid;place-items:center;gap:10px;min-height:120px;color:var(--muted);border:1px dashed var(--line);border-radius:8px;margin-top:14px;background:var(--surface);text-align:center}.racik-form{grid-template-columns:repeat(4,minmax(0,1fr))}.compact{margin-top:12px}.selected-tags{display:flex;flex-wrap:wrap;gap:8px;margin-top:10px}.selected-tags span{display:inline-flex;align-items:center;gap:6px;padding:7px 10px;border:1px solid var(--line);border-radius:999px;background:var(--surface)}.selected-tags button{border:0;background:transparent;color:inherit;cursor:pointer}.recipe-racik-actions{display:grid;grid-template-columns:1fr auto;align-items:end;gap:10px;margin-top:12px}.racik-saved{display:flex;flex-direction:column;padding:10px;border-bottom:1px solid var(--line)}
.recipe-actions{display:flex;align-items:center;gap:10px;margin-top:18px;padding-top:14px;border-top:1px solid var(--line)}.recipe-actions small{color:var(--muted)}.recipe-section-header{display:flex;align-items:center;justify-content:space-between;gap:16px;padding:16px 18px;border-bottom:1px solid var(--line)}.recipe-section-header h3{margin:3px 0 0;font-size:1.05rem}.recipe-section-header p{margin:4px 0 0;color:var(--muted);font-size:.82rem}.recipe-history{display:grid;margin:16px 18px 18px;border:1px solid var(--line);border-radius:12px;overflow:auto;background:var(--surface)}.history-head,.history-row{display:grid;grid-template-columns:minmax(150px,1fr) minmax(250px,1.5fr) 120px 115px 120px 118px;align-items:center;gap:14px;min-width:960px}.history-head{padding:12px 14px;background:color-mix(in srgb,#0d9488 14%,var(--surface-soft));color:var(--muted);font-size:.7rem;font-weight:800;text-transform:uppercase;letter-spacing:.08em}.history-row{padding:14px;border-top:1px solid var(--line);background:var(--surface)}.history-row:nth-child(odd){background:color-mix(in srgb,#0d9488 4%,var(--surface))}.history-row>div{display:flex;min-width:0;flex-direction:column;gap:4px}.history-row b,.history-row strong{min-width:0;overflow:hidden;text-overflow:ellipsis;white-space:nowrap}.history-prescription b{color:#0d9488}.history-total{text-align:right}.history-actions{display:flex!important;align-items:center;justify-content:flex-end;flex-direction:row!important;gap:8px}.recipe-status-pill{display:inline-flex;width:max-content;max-width:100%;align-items:center;justify-content:center;border-radius:999px;padding:5px 10px;color:#0f766e;background:rgba(20,184,166,.14);font-size:.72rem;font-weight:800;white-space:nowrap}.empty-state{text-align:center;color:var(--muted);padding:30px}
.recipe-form :deep(.staff-search-results){z-index:120}
@media(max-width:1000px){.recipe-grid,.racik-form{grid-template-columns:1fr 1fr}.recipe-row{min-width:760px}.recipe-items{overflow-x:auto}.recipe-search{grid-template-columns:1fr auto}}
@media(max-width:650px){.recipe-grid,.racik-form,.recipe-search,.recipe-racik-actions{grid-template-columns:1fr}.recipe-total-strip{grid-template-columns:1fr}.recipe-total-strip div{border-right:0;border-bottom:1px solid var(--line)}.recipe-total-strip div:last-child{align-items:flex-start;border-bottom:0}.recipe-actions{align-items:flex-start;flex-direction:column}.history-row{grid-template-columns:1fr auto}}
</style>
