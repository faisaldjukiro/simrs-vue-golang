<script setup lang="ts">
import { X } from '@lucide/vue'
import { ref, watch } from 'vue'
import InputPencarian from './InputPencarian.vue'
import { cariTindakanPenanganan } from '../../lib/faisal/api'

const model = defineModel<Record<string, any>[]>({ default: () => [] })
const props = defineProps({
  token: { type: String, required: true },
  noRawat: { type: String, required: true },
  jenisRawat: { type: String, default: 'ranap' },
  multiple: { type: Boolean, default: true },
  disabled: Boolean,
  required: Boolean,
})

const pilihan = ref<Record<string, any>>({})
const labelKelas = (kelas) => {
  const value = String(kelas || '').trim()
  if (!value || value === '-') return ''
  return value.toLowerCase().startsWith('kelas') ? value : `Kelas ${value}`
}
const cariTindakan = async (kataKunci) => {
  const daftar = await cariTindakanPenanganan(props.token, props.noRawat, kataKunci, props.jenisRawat)
  return (Array.isArray(daftar) ? daftar : []).map((item) => ({
    ...item,
    keterangan: [item.kategori, labelKelas(item.kelas)].filter(Boolean).join(' · '),
  }))
}
const rupiah = (value) => new Intl.NumberFormat('id-ID', {
  style: 'currency',
  currency: 'IDR',
  maximumFractionDigits: 0,
}).format(Number(value || 0))

watch(pilihan, (item) => {
  if (!item?.kode) return
  const daftar = Array.isArray(model.value) ? model.value : []
  if (!props.multiple) {
    model.value = [{ ...item }]
  } else if (!daftar.some((tindakan) => tindakan.kode === item.kode)) {
    model.value = [...daftar, { ...item }]
  }
  pilihan.value = {}
})

function hapus(kode) {
  if (props.disabled) return
  model.value = model.value.filter((item) => item.kode !== kode)
}
</script>

<template>
  <div class="handling-treatment handling-treatment-multiple">
    <InputPencarian
      v-model="pilihan"
      label="Tindakan / Tagihan"
      :placeholder="jenisRawat === 'ralan' ? 'Cari kode atau nama tindakan IGD/rawat jalan...' : 'Cari kode atau nama tindakan rawat inap...'"
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
          <small>{{ item.kode }}<template v-if="item.kategori"> · {{ item.kategori }}</template><template v-if="labelKelas(item.kelas)"> · {{ labelKelas(item.kelas) }}</template></small>
        </span>
        <strong class="handling-treatment-price">{{ rupiah(item.total) }}</strong>
        <button v-if="!disabled" type="button" title="Hapus tindakan" @click="hapus(item.kode)"><X :size="14" /></button>
      </article>
    </div>
  </div>
</template>
