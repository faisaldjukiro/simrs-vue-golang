<script setup lang="ts">
import { useId } from 'vue'
import { useParafEdukasi } from './useParafEdukasi'
const props = defineProps<{ modelValue: string; disabled?: boolean; nama: string; label?: string }>()
const emit = defineEmits<{ 'update:modelValue': [nilai: string] }>()
const id = useId()
const { kanvas, pesan, mulai, gerak, selesai, batal, hapus } = useParafEdukasi(props, nilai => emit('update:modelValue', nilai))
</script>

<template>
  <div class="edukasi-paraf">
    <p :id="id">Gambar paraf menggunakan mouse, pena, atau layar sentuh. Paraf ikut disimpan bersama catatan.</p>
    <canvas ref="kanvas" width="720" height="240" :aria-describedby="id" :aria-label="label || 'Kotak gambar paraf petugas'"
      :aria-disabled="disabled" @pointerdown.prevent="mulai" @pointermove.prevent="gerak"
      @pointerup.prevent="selesai" @pointercancel="batal" @lostpointercapture="batal">
      Browser tidak mendukung kotak paraf.
    </canvas>
    <strong>{{ nama || 'Lengkapi nama terlebih dahulu' }}</strong>
    <p v-if="pesan" role="alert">{{ pesan }}</p>
    <button type="button" class="clinical-button secondary" :disabled="disabled" @click="hapus">Hapus / Gambar Ulang Paraf</button>
  </div>
</template>

<style src="@/Pages/RawatInap/EdukasiPasien/paraf-edukasi.css" scoped></style>
