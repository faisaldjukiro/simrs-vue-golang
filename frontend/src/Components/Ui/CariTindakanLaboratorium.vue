<script setup lang="ts">
// @ts-nocheck -- migrasi TypeScript bertahap; kontrak data modul lama belum sepenuhnya bertipe.
import { ChevronDown, Trash2 } from '@lucide/vue'
import { computed, onBeforeUnmount, ref, watch } from 'vue'
import { cariTindakanLaboratorium, detailTindakanLaboratorium } from '../../lib/faisal/api'
import InputPencarian from './InputPencarian.vue'

const model = defineModel<Record<string, any>[]>({ default: () => [] })
const props = defineProps({
  token: { type: String, required: true },
  noRawat: { type: String, required: true },
  kategori: { type: String, default: 'PK' },
  disabled: Boolean,
  required: Boolean,
})

const pilihan = ref<Record<string, any>>({})
const emit = defineEmits<{ loading: [value: boolean] }>()
let aktif = true
onBeforeUnmount(() => {
  aktif = false
  emit('loading', false)
})
const loadingDetail = ref('')
const detailError = ref('')
const totalTarif = computed(() => model.value.reduce((total, item) => total + Number(item.total || 0), 0))
const rupiah = (value) => new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR', maximumFractionDigits: 0 }).format(Number(value || 0))
const cariTindakan = async (kataKunci) => {
  const daftar = await cariTindakanLaboratorium(props.token, props.noRawat, kataKunci, props.kategori)
  return (Array.isArray(daftar) ? daftar : []).map((item) => ({
    ...item,
    keterangan: [item.kode_cara_bayar, item.kelas, `${item.jumlah_detail || 0} detail`].filter(Boolean).join(' · '),
  }))
}

watch(pilihan, async (item) => {
  if (!item?.kode || props.disabled) return
  const konteks = `${props.noRawat}:${props.kategori}`
  const daftar = Array.isArray(model.value) ? model.value : []
  if (daftar.some((pemeriksaan) => pemeriksaan.kode === item.kode)) {
    pilihan.value = {}
    return
  }
  loadingDetail.value = item.kode
  emit('loading', true)
  detailError.value = ''
  try {
    const detail = await detailTindakanLaboratorium(props.token, props.noRawat, item.kode, props.kategori)
    if (!aktif || konteks !== `${props.noRawat}:${props.kategori}` || props.disabled) return
    model.value = [...model.value.filter((pemeriksaan) => pemeriksaan.kode !== item.kode), {
      ...item,
      detail_opsi: (Array.isArray(detail) ? detail : []).map((pilihanDetail) => ({ ...pilihanDetail, dipilih: true })),
    }]
  } catch (error) {
    detailError.value = error.message || 'Pilihan detail laboratorium tidak dapat dibaca.'
  } finally {
    loadingDetail.value = ''
    if (aktif) emit('loading', false)
  }
  pilihan.value = {}
})

function hapus(kode) {
  if (!props.disabled) model.value = model.value.filter((item) => item.kode !== kode)
}

function jumlahDipilih(item) {
  return (item.detail_opsi || []).filter((detail) => detail.dipilih).length
}

function semuaDipilih(item) {
  return item.detail_opsi?.length > 0 && jumlahDipilih(item) === item.detail_opsi.length
}

function pilihSemua(item, checked) {
  if (props.disabled) return
  item.detail_opsi.forEach((detail) => { detail.dipilih = checked })
}
</script>

<template>
  <div class="handling-treatment handling-treatment-multiple radiology-treatment-search">
    <InputPencarian
      v-model="pilihan"
      label="Pemeriksaan Laboratorium"
      placeholder="Cari kode atau nama pemeriksaan laboratorium..."
      :search="cariTindakan"
      description-field="keterangan"
      right-field="total"
      :right-formatter="rupiah"
      :disabled="disabled || Boolean(loadingDetail)"
      :required="required && model.length === 0"
    />
    <p v-if="loadingDetail" class="laboratory-detail-message">Membaca pilihan detail pemeriksaan...</p>
    <p v-if="detailError" class="laboratory-detail-message error">{{ detailError }}</p>
    <section v-if="model.length" class="lab-selected" aria-label="Pemeriksaan terpilih">
      <header class="lab-selected-heading">
        <strong>Pemeriksaan terpilih <span>{{ model.length }}</span></strong>
        <small>Buka detail untuk mengatur item pemeriksaan.</small>
      </header>
      <article v-for="(item, index) in model" :key="item.kode" class="lab-selected-row">
        <div class="lab-selected-main">
          <span class="lab-selected-number">{{ index + 1 }}</span>
          <div class="lab-selected-name">
            <strong>{{ item.nama || item.kode }}</strong>
            <small>{{ [item.kode, item.kelas !== '-' ? item.kelas : ''].filter(Boolean).join(' · ') }}</small>
          </div>
          <span class="lab-selected-price">{{ rupiah(item.total) }}</span>
          <div class="clinical-table-actions lab-selected-action">
            <button
              type="button"
              class="danger"
              :disabled="disabled"
              :title="`Hapus ${item.nama || item.kode}`"
              :aria-label="`Hapus ${item.nama || item.kode}`"
              @click="hapus(item.kode)"
            >
              <Trash2 :size="15" />
            </button>
          </div>
        </div>
        <details v-if="item.detail_opsi?.length" class="lab-selected-details">
          <summary>
            <ChevronDown :size="14" class="lab-selected-chevron" />
            <span>Detail pemeriksaan</span>
            <span class="lab-selected-count">{{ jumlahDipilih(item) }}/{{ item.detail_opsi.length }} dipilih</span>
          </summary>
          <div class="lab-selected-options">
            <label v-if="item.detail_opsi.length > 1" class="lab-selected-all">
              <input
                type="checkbox"
                :checked="semuaDipilih(item)"
                :indeterminate="jumlahDipilih(item) > 0 && !semuaDipilih(item)"
                :disabled="disabled"
                @change="pilihSemua(item, $event.target.checked)"
              >
              <span>Pilih semua detail</span>
            </label>
            <label v-for="detail in item.detail_opsi" :key="detail.id" class="lab-selected-option">
              <input v-model="detail.dipilih" type="checkbox" :disabled="disabled">
              <span>
                <strong>{{ detail.nama }}</strong>
                <small v-if="detail.satuan">Satuan: {{ detail.satuan }}</small>
                <small v-if="detail.nilai_rujukan">Nilai rujukan: {{ detail.nilai_rujukan }}</small>
              </span>
            </label>
          </div>
        </details>
      </article>
      <footer class="lab-selected-total">
        <span>Total tarif pemeriksaan</span>
        <strong>{{ rupiah(totalTarif) }}</strong>
      </footer>
    </section>
  </div>
</template>

<style src="/src/Components/Ui/cari-tindakan-laboratorium.css" scoped></style>
