<script setup lang="ts">
import { bidangSbar } from '../../../types/sbar'
import type { CatatanSbar } from '../../../types/sbar'
defineProps<{ catatan: CatatanSbar }>()
</script>

<template>
  <div class="sbar-isi">
    <p>{{ catatan.data.tgl_perawatan }} {{ catatan.data.jam_rawat }} WITA · {{ catatan.nama_petugas }}</p>
    <section v-for="bidang in bidangSbar" :key="bidang.key">
      <h4>{{ bidang.huruf }} · {{ bidang.label }}</h4>
      <p>{{ catatan.data[bidang.key] || 'Belum diisi' }}</p>
    </section>
    <section>
      <h4>Instruksi</h4>
      <p>{{ catatan.data.instruksi || 'Belum diisi' }}</p>
    </section>
    <section>
      <h4>Verifikasi</h4>
      <p>{{ catatan.status || (catatan.terkunci ? 'Terkunci — ada waktu validasi instruksi' : 'Belum diverifikasi') }}</p>
      <p v-if="catatan.validator">
        {{ catatan.nama_validator || catatan.validator }} · {{ catatan.tanggal_validasi }} {{ catatan.jam_validasi }} WITA
      </p>
    </section>
  </div>
</template>

<style scoped>
.sbar-isi {
  color: var(--text);
  font-size: 13px;
}
.sbar-isi section {
  padding: 12px 0;
  border-top: 1px solid var(--line);
}
.sbar-isi h4 {
  margin: 0 0 6px;
  font-size: 14px;
}
.sbar-isi p {
  margin: 6px 0;
  white-space: pre-wrap;
  overflow-wrap: anywhere;
  line-height: 1.6;
}
</style>
