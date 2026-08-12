<script setup>
import { DoorOpen } from '@lucide/vue'

defineEmits(['restart'])
</script>

<template>
  <main class="session-loader shutdown-screen" aria-live="polite">
    <DoorOpen :size="44" />
    <h1>Sesi SIRAVA Ditutup</h1>
    <p>Anda telah berhasil keluar dari sistem SIRAVA RS Prof. Dr. H. Aloei Saboe.</p>
    <button type="button" @click="$emit('restart')">Nyalakan Ulang Terminal</button>
  </main>
</template>
