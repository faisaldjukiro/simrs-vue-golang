<script setup>
import { Check, LoaderCircle, Search, UserRound, X } from '@lucide/vue'
import { onBeforeUnmount, ref, watch } from 'vue'
import { cariPetugasCppt } from '../../lib/faisal/api'

const model = defineModel({
  default: () => ({ nip: '', nama: '', jabatan: '' }),
})

const props = defineProps({
  token: { type: String, required: true },
  label: { type: String, default: 'Petugas' },
  placeholder: { type: String, default: 'Cari NIK, nama, atau jabatan petugas...' },
  disabled: Boolean,
  required: Boolean,
})

const query = ref('')
const results = ref([])
const loading = ref(false)
const open = ref(false)
let debounceTimer
let requestOrder = 0

function selectOfficer(officer) {
  model.value = { ...officer }
  query.value = ''
  results.value = []
  open.value = false
}

function clearSelection() {
  if (props.disabled) return
  model.value = { nip: '', nama: '', jabatan: '' }
  query.value = ''
  results.value = []
}

async function searchOfficer(keyword) {
  const currentRequest = ++requestOrder
  loading.value = true
  try {
    const data = await cariPetugasCppt(props.token, keyword)
    if (currentRequest !== requestOrder) return
    results.value = Array.isArray(data) ? data : []
    open.value = true
  } catch {
    if (currentRequest !== requestOrder) return
    results.value = []
    open.value = true
  } finally {
    if (currentRequest === requestOrder) loading.value = false
  }
}

watch(query, (value) => {
  window.clearTimeout(debounceTimer)
  const keyword = value.trim()
  if (props.disabled || keyword.length < 2) {
    results.value = []
    open.value = false
    loading.value = false
    return
  }
  debounceTimer = window.setTimeout(() => searchOfficer(keyword), 300)
})

onBeforeUnmount(() => window.clearTimeout(debounceTimer))
</script>

<template>
  <div class="staff-search" :class="{ disabled }">
    <span class="form-input-label">
      {{ label }}
      <i v-if="required" aria-hidden="true">*</i>
    </span>

    <div v-if="model?.nip" class="staff-search-selected">
      <UserRound :size="15" />
      <span><strong>{{ model.nama || model.nip }}</strong><small>{{ model.nip }}<template v-if="model.jabatan"> · {{ model.jabatan }}</template></small></span>
      <Check v-if="disabled" :size="15" />
      <button v-else type="button" title="Ganti petugas" aria-label="Ganti petugas" @click="clearSelection"><X :size="14" /></button>
    </div>

    <div v-else class="staff-search-box">
      <Search :size="15" />
      <input v-model="query" type="search" :placeholder="placeholder" :disabled="disabled" @focus="query.trim().length >= 2 && (open = true)" />
      <LoaderCircle v-if="loading" class="spin" :size="15" />
    </div>

    <div v-if="open && !disabled" class="staff-search-results">
      <button v-for="officer in results" :key="officer.nip" type="button" @click="selectOfficer(officer)">
        <UserRound :size="15" />
        <span><strong>{{ officer.nama || officer.nip }}</strong><small>{{ officer.nip }}<template v-if="officer.jabatan"> · {{ officer.jabatan }}</template></small></span>
      </button>
      <p v-if="!loading && results.length === 0">Petugas tidak ditemukan.</p>
    </div>

    <small v-if="disabled" class="staff-search-hint">Petugas mengikuti akun yang sedang login.</small>
  </div>
</template>
