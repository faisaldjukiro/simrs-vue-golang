<script setup>
import { ArrowDown, ArrowUp, FileCheck2, LoaderCircle, Save, ShieldAlert, Trash2 } from '@lucide/vue'
import { computed, nextTick, ref, watch } from 'vue'
import InputPencarian from '../../Components/Ui/InputPencarian.vue'
import { cariCodingDiagnosaPasien, diagnosaPasienData, hapusCodingDiagnosaPasien, simpanDiagnosaPasien } from '../../lib/faisal/api'
import { useNotifikasi } from '../../lib/shared/useNotifikasi'

const props = defineProps({
  token: { type: String, required: true },
  patient: { type: Object, required: true },
  moduleName: { type: String, required: true },
})

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
</script>

<template>
  <section class="patient-coding-page">
    <article class="patient-coding-card">
      <header class="patient-coding-header">
        <div>
          <span>CODING PASIEN</span>
          <h3>Diagnosa & Prosedur</h3>
          <p>Data disimpan langsung ke diagnosa_pasien dan prosedur_pasien SIMRS Khanza.</p>
        </div>
        <strong><FileCheck2 :size="16" /> {{ jumlahCoding }} coding</strong>
      </header>

      <div v-if="billingLocked" class="patient-coding-lock">
        <ShieldAlert :size="17" />
        <span><strong>Billing sudah terverifikasi.</strong> Diagnosa dan prosedur hanya dapat dilihat.</span>
      </div>

      <div v-if="loading" class="patient-coding-state">
        <LoaderCircle class="spin" :size="24" />
        <span>Memuat diagnosa dan prosedur pasien...</span>
      </div>

      <div v-else class="patient-coding-content">
        <section class="patient-coding-section">
          <header>
            <div><span>ICD-10</span><h4>Diagnosa</h4></div>
            <small>{{ diagnosa.length }} diagnosa</small>
          </header>

          <div class="patient-coding-list">
            <article v-for="(item, index) in diagnosa" :key="`diagnosa-${item.kode}`">
              <code>{{ item.kode }}</code>
              <div><strong>{{ item.nama }}</strong><span>{{ index === 0 ? 'Diagnosa utama' : `Diagnosa sekunder ${index}` }}</span></div>
              <i v-if="!item.tersimpan">Belum disimpan</i>
              <i v-else>{{ item.status_penyakit || 'Baru' }}<template v-if="item.im === '1'"> · IM</template></i>
              <div class="patient-coding-actions">
                <button type="button" title="Naikkan urutan pilihan baru" :disabled="billingLocked || item.tersimpan || index === 0 || diagnosa[index - 1]?.tersimpan" @click="pindahCoding('diagnosa', index, -1)"><ArrowUp :size="14" /></button>
                <button type="button" title="Turunkan urutan pilihan baru" :disabled="billingLocked || item.tersimpan || index === diagnosa.length - 1" @click="pindahCoding('diagnosa', index, 1)"><ArrowDown :size="14" /></button>
                <button type="button" class="danger" title="Hapus diagnosa" :disabled="billingLocked || deletingKey === `diagnosa-${item.kode}`" @click="hapusCoding('diagnosa', index)"><LoaderCircle v-if="deletingKey === `diagnosa-${item.kode}`" class="spin" :size="14" /><Trash2 v-else :size="14" /></button>
              </div>
            </article>
            <p v-if="diagnosa.length === 0">Belum ada diagnosa pasien.</p>
          </div>

          <InputPencarian
            v-model="diagnosaBaru"
            label="Tambah Diagnosa (ICD-10)"
            placeholder="Ketik minimal 3 karakter kode atau nama diagnosa..."
            :search="(kataKunci) => cariCoding('diagnosa', kataKunci)"
            right-field="penanda"
            :disabled="billingLocked || saving"
            @update:model-value="tambahCoding('diagnosa', $event)"
          />
        </section>

        <section class="patient-coding-section">
          <header>
            <div><span>ICD-9-CM</span><h4>Prosedur / Tindakan</h4></div>
            <small>{{ prosedur.length }} prosedur</small>
          </header>

          <div class="patient-coding-list">
            <article v-for="(item, index) in prosedur" :key="`prosedur-${item.kode}`" class="procedure">
              <code>{{ item.kode }}</code>
              <div><strong>{{ item.nama }}</strong><span>{{ index === 0 ? 'Prosedur utama' : `Prosedur sekunder ${index}` }}</span></div>
              <label class="patient-coding-quantity">
                <span>Jml</span>
                <input v-model.number="item.jumlah" type="number" min="1" max="999" :disabled="billingLocked || item.tersimpan" />
              </label>
              <i v-if="!item.tersimpan">Belum disimpan</i>
              <i v-else-if="item.im === '1'">IM</i>
              <div class="patient-coding-actions">
                <button type="button" title="Naikkan urutan pilihan baru" :disabled="billingLocked || item.tersimpan || index === 0 || prosedur[index - 1]?.tersimpan" @click="pindahCoding('prosedur', index, -1)"><ArrowUp :size="14" /></button>
                <button type="button" title="Turunkan urutan pilihan baru" :disabled="billingLocked || item.tersimpan || index === prosedur.length - 1" @click="pindahCoding('prosedur', index, 1)"><ArrowDown :size="14" /></button>
                <button type="button" class="danger" title="Hapus prosedur" :disabled="billingLocked || deletingKey === `prosedur-${item.kode}`" @click="hapusCoding('prosedur', index)"><LoaderCircle v-if="deletingKey === `prosedur-${item.kode}`" class="spin" :size="14" /><Trash2 v-else :size="14" /></button>
              </div>
            </article>
            <p v-if="prosedur.length === 0">Belum ada prosedur pasien.</p>
          </div>

          <InputPencarian
            v-model="prosedurBaru"
            label="Tambah Prosedur (ICD-9-CM)"
            placeholder="Ketik minimal 3 karakter kode atau nama prosedur..."
            :search="(kataKunci) => cariCoding('prosedur', kataKunci)"
            right-field="penanda"
            :disabled="billingLocked || saving"
            @update:model-value="tambahCoding('prosedur', $event)"
          />
        </section>

        <footer v-if="!billingLocked" class="patient-coding-footer">
          <button type="button" :disabled="saving" @click="simpan">
            <LoaderCircle v-if="saving" class="spin" :size="15" />
            <Save v-else :size="15" />
            {{ saving ? 'Menyimpan...' : `Simpan ${jumlahBaru} Pilihan Baru` }}
          </button>
        </footer>
      </div>
    </article>
  </section>
</template>

<style>
.patient-coding-page{margin-top:15px}.patient-coding-card{overflow:visible;border:1px solid var(--line);border-radius:16px;background:var(--surface-soft)}
.patient-coding-header{display:flex;align-items:center;justify-content:space-between;gap:16px;padding:17px 19px;border-bottom:1px solid var(--line);background:var(--surface)}.patient-coding-header>div>span,.patient-coding-section>header span{color:#0d9488;font-size:9px;font-weight:800;letter-spacing:.13em}.patient-coding-header h3{margin:5px 0 0;font-size:18px}.patient-coding-header p{margin:5px 0 0;color:var(--muted);font-size:11px}.patient-coding-header>strong{display:flex;align-items:center;gap:7px;padding:9px 11px;border-radius:10px;color:#0f766e;background:rgba(20,184,166,.12);font-size:11px}
.patient-coding-lock{display:flex;align-items:center;gap:9px;margin:14px 14px 0;padding:11px 13px;border:1px solid rgba(245,158,11,.3);border-radius:10px;color:#92400e;background:rgba(245,158,11,.1);font-size:11px}.patient-coding-state{display:grid;min-height:280px;place-items:center;align-content:center;gap:10px;color:var(--muted)}.patient-coding-content{display:grid;grid-template-columns:1fr 1fr;gap:14px;padding:14px}.patient-coding-section{position:relative;min-width:0;padding:14px;border:1px solid var(--line);border-radius:13px;background:var(--surface)}.patient-coding-section>header{display:flex;align-items:center;justify-content:space-between;padding-bottom:11px;border-bottom:1px solid var(--line)}.patient-coding-section>header h4{margin:4px 0 0;font-size:14px}.patient-coding-section>header small{color:var(--muted);font-size:10px}.patient-coding-list{min-height:90px;margin-bottom:15px}.patient-coding-list>article{display:grid;grid-template-columns:72px minmax(0,1fr) auto 94px;align-items:center;gap:10px;padding:11px 0;border-bottom:1px solid var(--line)}.patient-coding-list code{color:#2563eb;font-size:11px;font-weight:750}.patient-coding-list article>div:nth-child(2){min-width:0}.patient-coding-list article>div:nth-child(2) strong,.patient-coding-list article>div:nth-child(2) span{display:block;overflow:hidden;text-overflow:ellipsis;white-space:nowrap}.patient-coding-list article>div:nth-child(2) strong{color:var(--text);font-size:11px;font-weight:700}.patient-coding-list article>div:nth-child(2) span{margin-top:4px;color:var(--muted);font-size:9px}.patient-coding-list article>i{padding:4px 6px;border-radius:999px;color:#0f766e;background:rgba(20,184,166,.12);font-size:8px;font-style:normal;font-weight:800}.patient-coding-list>p{display:grid;min-height:80px;place-items:center;margin:0;color:var(--muted);font-size:11px}.patient-coding-actions{display:flex;justify-content:flex-end;gap:4px}.patient-coding-actions button{display:grid;width:27px;height:27px;place-items:center;border:1px solid var(--line);border-radius:7px;color:var(--muted);background:var(--surface-soft);cursor:pointer}.patient-coding-actions button:hover:not(:disabled){color:#0d9488;border-color:rgba(13,148,136,.4)}.patient-coding-actions button.danger:hover:not(:disabled){color:#dc2626;border-color:rgba(220,38,38,.35)}.patient-coding-actions button:disabled{cursor:not-allowed;opacity:.35}.patient-coding-section>.staff-search{position:relative}.patient-coding-section .staff-search-results{z-index:140}.patient-coding-footer{display:flex;grid-column:1/-1;justify-content:flex-end;padding-top:2px}.patient-coding-footer button{display:flex;height:40px;align-items:center;gap:7px;border-radius:10px;padding:0 14px;color:white;background:#0d9488;font-size:11px;font-weight:800;cursor:pointer}.patient-coding-footer button:disabled{cursor:not-allowed;opacity:.6}
.theme-dark .patient-coding-card,.sirapi-dark .patient-coding-card{border-color:#304258;background:#0c1625}.theme-dark .patient-coding-header,.theme-dark .patient-coding-section,.sirapi-dark .patient-coding-header,.sirapi-dark .patient-coding-section{border-color:#304258;background:#142132}.theme-dark .patient-coding-section>header,.theme-dark .patient-coding-list>article,.sirapi-dark .patient-coding-section>header,.sirapi-dark .patient-coding-list>article{border-color:#304258}.theme-dark .patient-coding-list code,.sirapi-dark .patient-coding-list code{color:#60a5fa}.theme-dark .patient-coding-list article>i,.theme-dark .patient-coding-header>strong,.sirapi-dark .patient-coding-list article>i,.sirapi-dark .patient-coding-header>strong{color:#5eead4;background:rgba(20,184,166,.14)}.theme-dark .patient-coding-actions button,.sirapi-dark .patient-coding-actions button{border-color:#3b4d64;color:#cbd5e1;background:#1c293b}.theme-dark .patient-coding-lock,.sirapi-dark .patient-coding-lock{color:#fcd34d;background:rgba(120,53,15,.24)}
.patient-coding-list>article.procedure{grid-template-columns:62px minmax(0,1fr) 54px auto 94px}.patient-coding-quantity{display:grid;gap:3px;color:var(--muted);font-size:8px;font-weight:800}.patient-coding-quantity input{width:52px;height:29px;padding:0 6px;border:1px solid var(--line);border-radius:7px;color:var(--text);background:var(--surface-soft);font:inherit;font-size:10px}.patient-coding-quantity input:disabled{opacity:.75}.theme-dark .patient-coding-quantity input,.sirapi-dark .patient-coding-quantity input{border-color:#3b4d64;color:#cbd5e1;background:#1c293b}
@media(max-width:1100px){.patient-coding-content{grid-template-columns:1fr}.patient-coding-footer{grid-column:auto}}@media(max-width:650px){.patient-coding-header{align-items:flex-start;flex-direction:column}.patient-coding-list>article{grid-template-columns:62px minmax(0,1fr) auto}.patient-coding-actions{grid-column:1/-1;justify-content:flex-end}}
</style>
