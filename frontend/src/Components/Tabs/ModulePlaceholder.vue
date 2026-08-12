<script setup>
import { LayoutDashboard } from '@lucide/vue'

defineProps({
  currentTab: { type: String, required: true },
  ribbonMenus: { type: Array, required: true },
})

const emit = defineEmits(['back'])
</script>

<template>
  <section class="module-placeholder">
    <span :class="`tone-bg-${ribbonMenus.find((item) => item.label === currentTab)?.tone || 'slate'}`">
      <component :is="ribbonMenus.find((item) => item.label === currentTab)?.icon || LayoutDashboard" :size="30" />
    </span>
    <p>Modul SIRAVA</p>
    <h1>{{ currentTab }}</h1>
    <div>Struktur halaman {{ currentTab }} akan dimigrasikan pada tahap berikutnya.</div>
    <button type="button" @click="emit('back')">Kembali ke Beranda</button>
  </section>
</template>
