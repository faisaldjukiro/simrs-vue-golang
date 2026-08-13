<script setup>
import { Check, LoaderCircle, Search, X } from '@lucide/vue'
import { onBeforeUnmount, ref, watch } from 'vue'

const model = defineModel({ default: () => ({}) })
const props = defineProps({
  label: { type: String, required: true },
  placeholder: { type: String, default: 'Ketik minimal 2 huruf...' },
  search: { type: Function, required: true },
  codeField: { type: String, default: 'kode' },
  nameField: { type: String, default: 'nama' },
  descriptionField: { type: String, default: '' },
  rightField: { type: String, default: '' },
  rightFormatter: { type: Function, default: null },
  required: Boolean,
  disabled: Boolean,
})

const query = ref('')
const results = ref([])
const loading = ref(false)
const open = ref(false)
let timer
let order = 0

function pilih(item) {
  model.value = { ...item }
  query.value = ''
  results.value = []
  open.value = false
}

function bersihkan() {
  if (props.disabled) return
  model.value = {}
  query.value = ''
  results.value = []
}

function punyaNilaiKanan(item) {
  if (!props.rightField) return false
  const value = item?.[props.rightField]
  return value !== undefined && value !== null && value !== ''
}

function nilaiKanan(item) {
  const value = item?.[props.rightField]
  return props.rightFormatter ? props.rightFormatter(value, item) : value
}

watch(query, (value) => {
  window.clearTimeout(timer)
  const keyword = value.trim()
  if (props.disabled || keyword.length < 2) {
    open.value = false
    results.value = []
    loading.value = false
    return
  }
  timer = window.setTimeout(async () => {
    const current = ++order
    loading.value = true
    try {
      const data = await props.search(keyword)
      if (current !== order) return
      results.value = Array.isArray(data) ? data : []
      open.value = true
    } catch {
      if (current === order) {
        results.value = []
        open.value = true
      }
    } finally {
      if (current === order) loading.value = false
    }
  }, 280)
})

onBeforeUnmount(() => window.clearTimeout(timer))
</script>

<template>
  <div class="staff-search reference-search" :class="{ disabled }">
    <span class="form-input-label">{{ label }} <i v-if="required">*</i></span>
    <div v-if="model?.[codeField]" class="staff-search-selected">
      <Check :size="15" />
      <span>
        <strong>{{ model[nameField] || model[codeField] }}</strong>
        <small>{{ model[codeField] }}<template v-if="descriptionField && model[descriptionField]"> - {{ model[descriptionField] }}</template></small>
      </span>
      <strong v-if="punyaNilaiKanan(model)" class="reference-search-value">{{ nilaiKanan(model) }}</strong>
      <button v-if="!disabled" type="button" title="Ganti pilihan" @click="bersihkan"><X :size="14" /></button>
    </div>
    <div v-else class="staff-search-box">
      <Search :size="15" />
      <input v-model="query" type="search" :placeholder="placeholder" :disabled="disabled" @focus="query.trim().length >= 2 && (open = true)" />
      <LoaderCircle v-if="loading" class="spin" :size="15" />
    </div>
    <div v-if="open && !disabled" class="staff-search-results">
      <button v-for="item in results" :key="item[codeField]" type="button" @click="pilih(item)">
        <Search :size="14" />
        <span><strong>{{ item[nameField] || item[codeField] }}</strong><small>{{ item[codeField] }}<template v-if="descriptionField && item[descriptionField]"> - {{ item[descriptionField] }}</template></small></span>
        <strong v-if="punyaNilaiKanan(item)" class="reference-search-value">{{ nilaiKanan(item) }}</strong>
      </button>
      <p v-if="!loading && results.length === 0">Data tidak ditemukan.</p>
    </div>
  </div>
</template>
