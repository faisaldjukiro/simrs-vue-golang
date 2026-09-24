<script setup lang="ts">
import { Copy } from '@lucide/vue'
import InputPencarian from '../Ui/InputPencarian.vue'
import { useSalinHakAkses } from './useSalinHakAkses'

const props = defineProps<{
  pengguna: { id: number; username: string; nama?: string; permission?: string[] }[]
  targetId?: number
  permissionSaatIni: string[]
  disabled?: boolean
}>()
const emit = defineEmits<{ terapkan: [permission: string[]] }>()
const { pilihan, pratinjau, sumber, tambahan, dicabut, cari, siapkan, konfirmasi } = useSalinHakAkses(
  props, (kode) => emit('terapkan', kode),
)
</script>

<template>
  <section class="copy-access">
    <InputPencarian
      v-model="pilihan"
      label="Salin Hak Akses dari User"
      placeholder="Cari nama atau username sumber..."
      description-field="keterangan"
      :search="cari"
      :disabled="disabled"
    />
    <p>Hanya permission yang disalin. Identitas, password, dan status aktif tidak berubah.</p>
    <button type="button" class="clinical-button secondary" :disabled="disabled || !sumber" @click="siapkan">
      <Copy :size="15" /> Tinjau Hak Akses
    </button>
    <div v-if="pratinjau" class="copy-access-preview" role="region" aria-label="Konfirmasi salin hak akses">
      <strong>Sumber: {{ pratinjau.nama }}</strong>
      <p>Daftar permission pada form akan diganti dengan salinan ini. Belum disimpan ke server.</p>
      <p v-if="pratinjau.permission.includes('*')" role="alert">
        <strong>Perhatian: salinan ini memberikan akses admin penuh, termasuk mengatur akses user lain.</strong>
      </p>
      <p v-if="dicabut.includes('*')"><strong>Akses admin penuh pada form akan dicabut.</strong></p>
      <p v-if="!pratinjau.permission.length">User sumber belum memiliki permission. Pilih user lain.</p>
      <template v-else>
        <p>Ditambahkan: {{ tambahan.join(', ') || 'Tidak ada' }}</p>
        <p>Dihapus dari form: {{ dicabut.join(', ') || 'Tidak ada' }}</p>
        <details>
          <summary>Lihat {{ pratinjau.permission.length }} permission hasil salinan</summary>
          <p>{{ pratinjau.permission.join(', ') }}</p>
        </details>
      </template>
      <div class="copy-access-actions">
        <button type="button" class="clinical-button secondary" :disabled="disabled" @click="pratinjau = null">Batal</button>
        <button
          type="button"
          class="clinical-button primary"
          :disabled="disabled || !pratinjau.permission.length"
          @click="konfirmasi"
        >
          Ya, Terapkan ke Form
        </button>
      </div>
    </div>
    <p>Setelah diterapkan, sesuaikan checklist lalu klik Simpan. Akses kedua user tetap independen.</p>
  </section>
</template>

<style scoped>
.copy-access {
  display: grid;
  gap: 12px;
  padding: 16px;
  border: 1px solid var(--line);
  border-radius: 10px;
  color: var(--text);
  background: var(--surface);
  min-width: 0;
}

.copy-access p {
  margin: 0;
  color: var(--muted);
  font-size: 12px;
  line-height: 1.6;
  overflow-wrap: anywhere;
}

.copy-access strong {
  color: var(--text);
  font-size: 13px;
}

.copy-access-preview {
  display: grid;
  gap: 10px;
  padding-top: 12px;
  border-top: 1px solid var(--line);
}

.copy-access summary {
  cursor: pointer;
  font-size: 13px;
}

.copy-access-actions {
  display: flex;
  justify-content: flex-end;
  flex-wrap: wrap;
  gap: 8px;
}
</style>
