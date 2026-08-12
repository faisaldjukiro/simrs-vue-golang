<script setup>
import PrimeDataTable from 'primevue/datatable'

defineProps({
  rows: { type: Array, default: () => [] },
  dataKey: { type: String, default: 'id' },
  loading: Boolean,
  emptyMessage: { type: String, default: 'Data tidak ditemukan.' },
  loadingMessage: { type: String, default: 'Memuat data...' },
  paginator: Boolean,
  lazy: Boolean,
  first: { type: Number, default: 0 },
  rowsPerPage: { type: Number, default: 100 },
  rowsPerPageOptions: { type: Array, default: () => [50, 100, 200, 500] },
  totalRecords: { type: Number, default: 0 },
})

const emit = defineEmits(['page'])
</script>

<template>
  <div class="patient-table-wrap data-table-wrap">
    <PrimeDataTable
      class="data-table"
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
    >
      <slot />

      <template #empty>
        <div class="patient-empty">{{ emptyMessage }}</div>
      </template>
    </PrimeDataTable>
  </div>
</template>
