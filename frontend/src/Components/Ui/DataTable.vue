<script setup lang="ts">
import PrimeDataTable from 'primevue/datatable'

withDefaults(defineProps<{
  rows?: Record<string, any>[]
  dataKey?: string
  loading?: boolean
  emptyMessage?: string
  loadingMessage?: string
  paginator?: boolean
  lazy?: boolean
  first?: number
  rowsPerPage?: number
  rowsPerPageOptions?: number[]
  totalRecords?: number
  clickable?: boolean
}>(), {
  rows: () => [],
  dataKey: 'id',
  loading: false,
  emptyMessage: 'Data tidak ditemukan.',
  loadingMessage: 'Memuat data...',
  paginator: false,
  lazy: false,
  first: 0,
  rowsPerPage: 100,
  rowsPerPageOptions: () => [50, 100, 200, 500],
  totalRecords: 0,
  clickable: false,
})

const emit = defineEmits(['page', 'row-click'])
</script>

<template>
  <div class="patient-table-wrap data-table-wrap">
    <PrimeDataTable
      class="data-table"
      :class="{ 'data-table-clickable': clickable }"
      :value="rows"
      :data-key="dataKey"
      :loading="loading"
      :paginator="paginator"
      :lazy="lazy"
      :first="first"
      :rows="rowsPerPage"
      :total-records="totalRecords || rows.length"
      :rows-per-page-options="rowsPerPageOptions"
      paginator-template="RowsPerPageDropdown FirstPageLink PrevPageLink PageLinks NextPageLink LastPageLink CurrentPageReport"
      current-page-report-template="{first} - {last} dari {totalRecords} data"
      @page="emit('page', $event)"
      @row-click="emit('row-click', $event)"
    >
      <slot />

      <template #empty>
        <div class="patient-empty">{{ emptyMessage }}</div>
      </template>
    </PrimeDataTable>
  </div>
</template>
