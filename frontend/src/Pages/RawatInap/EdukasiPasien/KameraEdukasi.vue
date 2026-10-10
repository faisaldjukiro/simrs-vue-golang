<script setup lang="ts">
import { ref, watch } from 'vue'
import { Camera, RefreshCw, X } from '@lucide/vue'
import { useKameraEdukasi } from './useKameraEdukasi'
const props = defineProps<{ label: string; nama: string; pratinjau: string; baru: boolean; aktif: boolean; disabled?: boolean }>()
const emit = defineEmits<{ buka: []; tutup: []; tangkap: [file: File]; batal: [] }>()
const { video, memuat, mengambil, siap, pesan, selesai, ambil, gantiKamera } = useKameraEdukasi(props,
  file => emit('tangkap', file), () => emit('tutup'))
const gagal = ref(false)
watch(() => props.pratinjau, () => { gagal.value = false })
</script>

<template>
  <section class="edukasi-kamera" :aria-label="label">
    <h4>{{ label }}</h4>
    <p>{{ nama || 'Nama belum diisi' }}</p>
    <template v-if="aktif">
      <video ref="video" autoplay muted playsinline :aria-label="'Kamera ' + label" />
      <p v-if="memuat" role="status">Membuka kamera…</p>
      <div class="edukasi-kamera-aksi">
        <button type="button" class="clinical-button primary" :disabled="disabled || !siap || mengambil" @click="ambil">
          <Camera :size="15" /> {{ mengambil ? 'Mengambil…' : 'Ambil Foto' }}
        </button>
        <button type="button" class="clinical-button secondary" :disabled="disabled || memuat || mengambil" @click="gantiKamera">
          <RefreshCw :size="15" /> Ganti Kamera
        </button>
        <button type="button" class="clinical-button secondary" @click="selesai"><X :size="15" /> Tutup Kamera</button>
      </div>
    </template>
    <template v-else>
      <img v-if="pratinjau && !gagal" :src="pratinjau" :alt="label" @error="gagal = true" />
      <p v-if="gagal" role="alert">Foto tersimpan tidak dapat dimuat. Periksa koneksi server berkas.</p>
      <div class="edukasi-kamera-aksi">
        <button type="button" class="clinical-button secondary" :disabled="disabled" @click="emit('buka')">
          <Camera :size="15" /> {{ pratinjau ? 'Ambil Ulang Foto' : 'Buka Kamera' }}
        </button>
        <button v-if="baru" type="button" class="clinical-button secondary" :disabled="disabled" @click="emit('batal')">
          <X :size="15" /> Batalkan Foto Baru
        </button>
      </div>
    </template>
    <p v-if="pesan" class="patient-error" role="alert">{{ pesan }}</p>
    <p v-if="baru">Foto baru akan dikirim saat catatan disimpan.</p>
  </section>
</template>

<style src="@/Pages/RawatInap/EdukasiPasien/kamera-edukasi.css" scoped></style>
