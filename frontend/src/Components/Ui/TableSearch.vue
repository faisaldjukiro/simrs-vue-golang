<script setup lang="ts">
import { Search, X } from '@lucide/vue'

const model = defineModel({ default: '' })

defineProps({
  placeholder: { type: String, default: 'Cari data...' },
  total: { type: Number, default: 0 },
  filtered: { type: Number, default: 0 },
})
</script>

<template>
  <label class="table-search">
    <Search :size="16" />
    <input v-model="model" type="search" :placeholder="placeholder" />
    <span v-if="model" class="table-search-count">{{ filtered }} / {{ total }}</span>
    <button v-if="model" type="button" title="Bersihkan pencarian" @click="model = ''">
      <X :size="14" />
    </button>
  </label>
</template>

<style scoped>
.table-search {
  display: flex;
  width: min(380px, 42vw);
  min-width: 0;
  max-width: 100%;
  height: 42px;
  align-items: center;
  gap: 9px;
  border: 1px solid var(--line);
  border-radius: 12px;
  padding: 0 11px;
  color: var(--muted);
  background: var(--surface-soft);
}

.table-search input {
  flex: 1;
  min-width: 0;
  height: 100%;
  border: 0;
  outline: 0;
  color: var(--text);
  background: transparent;
  font-size: 13px;
  font-weight: 650;
}

.table-search > svg {
  flex-shrink: 0;
}

.table-search input::placeholder {
  color: var(--muted);
  opacity: .85;
}

.table-search-count {
  flex: 0 0 auto;
  white-space: nowrap;
  border-radius: 999px;
  padding: 4px 7px;
  color: #0d9488;
  background: rgba(20, 184, 166, .1);
  font-size: 10px;
  font-weight: 850;
}

.table-search button {
  display: inline-grid;
  width: 26px;
  height: 26px;
  flex: 0 0 auto;
  place-items: center;
  border: 0;
  border-radius: 8px;
  color: var(--muted);
  background: transparent;
  cursor: pointer;
}

.table-search button:hover {
  color: var(--text);
  background: rgba(20, 184, 166, .1);
}

:global(.theme-dark) .table-search,
:global(.sirapi-dark) .table-search {
  border-color: rgba(148, 163, 184, .18);
  background: #1e293b;
}

@media (max-width: 760px) {
  .table-search {
    width: 100%;
  }
}
</style>
