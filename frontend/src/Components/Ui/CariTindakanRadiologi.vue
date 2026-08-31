<script setup lang="ts">
import { X } from '@lucide/vue'
import { ref, watch } from 'vue'
import InputPencarian from './InputPencarian.vue'
import { cariTindakanRadiologi } from '../../lib/faisal/api'

const model = defineModel<Record<string, any>[]>({ default: () => [] })
const props = defineProps({
  token: { type: String, required: true },
  noRawat: { type: String, required: true },
  disabled: Boolean,
  required: Boolean,
})

const pilihan = ref<Record<string, any>>({})
const rupiah = (value) => new Intl.NumberFormat('id-ID', {
  style: 'currency', currency: 'IDR', maximumFractionDigits: 0,
}).format(Number(value || 0))
const cariTindakan = async (kataKunci) => {
  const daftar = await cariTindakanRadiologi(props.token, props.noRawat, kataKunci)
  return (Array.isArray(daftar) ? daftar : []).map((item) => ({
    ...item,
    keterangan: [item.kode_cara_bayar, item.kelas].filter(Boolean).join(' · '),
  }))
}

watch(pilihan, (item) => {
  if (!item?.kode) return
  const daftar = Array.isArray(model.value) ? model.value : []
  if (!daftar.some((tindakan) => tindakan.kode === item.kode)) {
    model.value = [...daftar, { ...item }]
  }
  pilihan.value = {}
})

function hapus(kode) {
  if (!props.disabled) model.value = model.value.filter((item) => item.kode !== kode)
}
</script>

<template>
  <div class="handling-treatment handling-treatment-multiple radiology-treatment-search">
    <InputPencarian
      v-model="pilihan"
      label="Tindakan Radiologi"
      placeholder="Cari kode atau nama tindakan radiologi..."
      :search="cariTindakan"
      description-field="keterangan"
      right-field="total"
      :right-formatter="rupiah"
      :disabled="disabled"
      :required="required && model.length === 0"
    />
    <div v-if="model.length" class="handling-treatment-list">
      <article v-for="(item, index) in model" :key="item.kode">
        <span class="handling-treatment-number">{{ index + 1 }}</span>
        <span>
          <strong>{{ item.nama || item.kode }}</strong>
          <small>{{ item.kode }}<template v-if="item.kelas"> · {{ item.kelas }}</template><template v-if="item.kode_cara_bayar"> · {{ item.kode_cara_bayar }}</template></small>
        </span>
        <strong class="handling-treatment-price">{{ rupiah(item.total) }}</strong>
        <button v-if="!disabled" type="button" title="Hapus tindakan" @click="hapus(item.kode)"><X :size="14" /></button>
      </article>
    </div>
  </div>
</template>
