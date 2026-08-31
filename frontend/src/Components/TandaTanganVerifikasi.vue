<script setup lang="ts">
import { QrCode } from '@lucide/vue'
import { onMounted, ref, watch } from 'vue'

const props = defineProps({
  data: { type: Object, required: true },
})

const sumberQR = ref('')
const gagal = ref(false)

function tampilkanQRCode() {
  gagal.value = false
  sumberQR.value = props.data.url_qr_code
    ? `${props.data.url_qr_code}${props.data.url_qr_code.includes('?') ? '&' : '?'}v=${Date.now()}`
    : ''
}

function muatTandaTangan() {
  sumberQR.value = ''
  gagal.value = false
  if (!props.data.url_generator) {
    tampilkanQRCode()
    return
  }

  const pemicu = new Image()
  pemicu.onload = tampilkanQRCode
  pemicu.onerror = tampilkanQRCode
  pemicu.src = `${props.data.url_generator}${props.data.url_generator.includes('?') ? '&' : '?'}v=${Date.now()}`
}

onMounted(muatTandaTangan)
watch(() => props.data, muatTandaTangan, { deep: true })
</script>

<template>
  <article class="care-history-signature-card">
    <div class="care-history-signature-qr">
      <img v-if="sumberQR && !gagal" :src="sumberQR" :alt="`QR verifikasi ${data.dokter}`" @error="gagal = true" />
      <QrCode v-else :size="50" />
    </div>
    <span>{{ data.peran }}<template v-if="data.peran === 'Dokter DPJP' && data.urutan"> {{ data.urutan }}</template></span>
    <strong>{{ data.dokter || '-' }}</strong>
    <small>ID {{ data.kode_dokter || '-' }}</small>
    <em v-if="gagal">QR belum dapat dimuat dari server Khanza.</em>
  </article>
</template>
