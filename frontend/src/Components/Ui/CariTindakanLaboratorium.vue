<script setup lang="ts">
// @ts-nocheck -- migrasi TypeScript bertahap; kontrak data modul lama belum sepenuhnya bertipe.
import { X } from '@lucide/vue'
import { ref, watch } from 'vue'
import { cariTindakanLaboratorium, detailTindakanLaboratorium } from '../../lib/faisal/api'
import InputPencarian from './InputPencarian.vue'

const model = defineModel<Record<string, any>[]>({ default: () => [] })
const props = defineProps({
  token: { type: String, required: true },
  noRawat: { type: String, required: true },
  disabled: Boolean,
  required: Boolean,
})

const pilihan = ref<Record<string, any>>({})
const loadingDetail = ref('')
const detailError = ref('')
const rupiah = (value) => new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR', maximumFractionDigits: 0 }).format(Number(value || 0))
const cariTindakan = async (kataKunci) => {
  const daftar = await cariTindakanLaboratorium(props.token, props.noRawat, kataKunci)
  return (Array.isArray(daftar) ? daftar : []).map((item) => ({
    ...item,
    keterangan: [item.kode_cara_bayar, item.kelas, `${item.jumlah_detail || 0} detail`].filter(Boolean).join(' · '),
  }))
}

watch(pilihan, async (item) => {
  if (!item?.kode) return
  const daftar = Array.isArray(model.value) ? model.value : []
  if (daftar.some((pemeriksaan) => pemeriksaan.kode === item.kode)) {
    pilihan.value = {}
    return
  }
  loadingDetail.value = item.kode
  detailError.value = ''
  try {
    const detail = await detailTindakanLaboratorium(props.token, props.noRawat, item.kode)
    model.value = [...daftar, {
      ...item,
      detail_opsi: (Array.isArray(detail) ? detail : []).map((pilihanDetail) => ({ ...pilihanDetail, dipilih: true })),
    }]
  } catch (error) {
    detailError.value = error.message || 'Pilihan detail laboratorium tidak dapat dibaca.'
  } finally {
    loadingDetail.value = ''
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
      :disabled="disabled"
      :required="required && model.length === 0"
    />
    <p v-if="loadingDetail" class="laboratory-detail-message">Membaca pilihan detail pemeriksaan...</p>
    <p v-if="detailError" class="laboratory-detail-message error">{{ detailError }}</p>
    <div v-if="model.length" class="laboratory-treatment-list">
      <article v-for="(item, index) in model" :key="item.kode" class="laboratory-treatment-item">
        <header>
          <span class="handling-treatment-number">{{ index + 1 }}</span>
          <span class="laboratory-treatment-name">
            <strong>{{ item.nama || item.kode }}</strong>
            <small>{{ item.kode }} · {{ item.kelas || '-' }} · {{ jumlahDipilih(item) }} dari {{ item.detail_opsi?.length || 0 }} detail dipilih</small>
          </span>
          <strong class="handling-treatment-price">{{ rupiah(item.total) }}</strong>
          <button v-if="!disabled" type="button" title="Hapus pemeriksaan" @click="hapus(item.kode)"><X :size="14" /></button>
        </header>

        <div v-if="item.detail_opsi?.length" class="laboratory-detail-options">
          <label class="laboratory-detail-select-all">
            <input type="checkbox" :checked="semuaDipilih(item)" :disabled="disabled" @change="pilihSemua(item, $event.target.checked)">
            <span>Pilih Semua Detail</span>
          </label>
          <label v-for="detail in item.detail_opsi" :key="detail.id" class="laboratory-detail-option">
            <input v-model="detail.dipilih" type="checkbox" :disabled="disabled">
            <span>
              <strong>{{ detail.nama }}</strong>
              <small>{{ [detail.satuan, detail.nilai_rujukan].filter(Boolean).join(' · ') || 'Tanpa satuan/nilai rujukan' }}</small>
            </span>
          </label>
        </div>
        <p v-else class="laboratory-no-detail">Pemeriksaan ini tidak mempunyai pilihan detail.</p>
      </article>
    </div>
  </div>
</template>
