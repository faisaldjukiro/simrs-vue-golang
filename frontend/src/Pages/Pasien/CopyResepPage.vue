<script setup>
import { ClipboardCopy, Copy, LoaderCircle, Pill, RefreshCcw, Save } from "@lucide/vue"
import { useCopyResep } from "./CopyResep/useCopyResep.js"

const props = defineProps({
  token: { type: String, required: true },
  patient: { type: Object, required: true },
  moduleName: { type: String, default: 'Rawat Inap' },
})
const emit = defineEmits(['copy-resep'])

const {
  loading,
  loadingDetail,
  locked,
  recipes,
  selected,
  detail,
  noRkmMedis,
  jumlahIsi,
  rupiah,
  tanggalPendek,
  jamPendek,
  labelStatus,
  load,
  pilihResep,
  salinResep,
} = useCopyResep(props, emit)
</script>

<template>
  <section class="copy-recipe-page">
    <div v-if="locked" class="copy-recipe-lock">
      Billing sudah diproses. Copy resep ke kunjungan ini tidak bisa dilakukan.
    </div>

    <article class="copy-recipe-card">
      <header class="copy-recipe-header">
        <div>
          <span class="eyebrow">COPY RESEP</span>
          <h3>Riwayat Resep Berdasarkan No. CM</h3>
          <p>Menampilkan seluruh resep pasien {{ noRkmMedis || '-' }} untuk disalin ke kunjungan saat ini.</p>
        </div>
        <button type="button" class="copy-recipe-button secondary" :disabled="loading" @click="load">
          <RefreshCcw :size="16" :class="{ spin: loading }" />
          Muat Ulang
        </button>
      </header>

      <div class="copy-recipe-layout">
        <section class="copy-recipe-list">
          <div v-if="loading" class="copy-recipe-empty">
            <LoaderCircle class="spin" :size="24" />
            <span>Memuat riwayat resep pasien...</span>
          </div>
          <div v-else-if="!recipes.length" class="copy-recipe-empty">
            <ClipboardCopy :size="28" />
            <span>Belum ada resep lama untuk No. CM ini.</span>
          </div>
          <button
            v-for="row in recipes"
            v-else
            :key="row.no_resep"
            type="button"
            class="copy-recipe-item"
            :class="{ active: selected?.no_resep === row.no_resep }"
            @click="pilihResep(row)"
          >
            <span>
              <strong>{{ row.no_resep }}</strong>
              <small>{{ tanggalPendek(row.tgl_peresepan) }} {{ jamPendek(row.jam) }} - {{ row.nm_dokter || '-' }}</small>
              <small>No. Rawat {{ row.no_rawat || '-' }}</small>
            </span>
            <span>
              <b>{{ rupiah(row.total) }}</b>
              <i>{{ row.jumlah_obat || 0 }} obat / {{ row.jumlah_racik || 0 }} racikan</i>
              <em>{{ labelStatus(row.status) }}</em>
            </span>
          </button>
        </section>

        <section class="copy-recipe-preview">
          <div v-if="loadingDetail" class="copy-recipe-empty">
            <LoaderCircle class="spin" :size="24" />
            <span>Membaca isi resep...</span>
          </div>

          <template v-else-if="detail?.header">
            <div class="copy-recipe-preview-head">
              <div>
                <span class="eyebrow">RESEP DIPILIH</span>
                <h3>{{ detail.header.no_resep }}</h3>
                <p>{{ tanggalPendek(detail.header.tgl_peresepan) }} {{ jamPendek(detail.header.jam) }} - {{ detail.header.nm_dokter || '-' }}</p>
              </div>
              <button type="button" class="copy-recipe-button" :disabled="locked || jumlahIsi === 0" @click="salinResep">
                <Copy :size="16" />
                Copy ke Input Resep
              </button>
            </div>

            <div class="copy-recipe-table">
              <div class="copy-recipe-table-head">
                <span>Jenis</span>
                <span>Kode</span>
                <span>Nama</span>
                <span>Jumlah</span>
                <span>Aturan Pakai</span>
              </div>
              <div v-for="item in detail.obat" :key="`obat-${item.kode_brng}`" class="copy-recipe-table-row">
                <span><Pill :size="14" /> Obat</span>
                <code>{{ item.kode_brng }}</code>
                <strong>{{ item.nama_brng }}</strong>
                <b>{{ item.jml }}</b>
                <small>{{ item.aturan_pakai || '-' }}</small>
              </div>
              <template v-for="racik in detail.racikan" :key="`racik-${racik.no_racik}`">
                <div class="copy-recipe-table-row racik">
                  <span>Racikan</span>
                  <code>{{ racik.kd_racik }}</code>
                  <strong>{{ racik.nama_racik }}</strong>
                  <b>{{ racik.jml_dr }}</b>
                  <small>{{ racik.aturan_pakai || '-' }}</small>
                </div>
                <div v-for="obat in racik.detail" :key="`${racik.no_racik}-${obat.kode_brng}`" class="copy-recipe-table-row child">
                  <span>Isi racik</span>
                  <code>{{ obat.kode_brng }}</code>
                  <strong>{{ obat.nama_brng }}</strong>
                  <b>{{ obat.jml }}</b>
                  <small>{{ obat.kandungan || '-' }}</small>
                </div>
              </template>
            </div>

            <footer class="copy-recipe-footer">
              <Save :size="15" />
              Setelah tombol copy diklik, data akan dibawa ke Input Resep. Resep lama tidak diubah.
            </footer>
          </template>

          <div v-else class="copy-recipe-empty">
            <ClipboardCopy :size="30" />
            <span>Pilih salah satu resep lama untuk melihat isinya.</span>
          </div>
        </section>
      </div>
    </article>
  </section>
</template>

<style src="./CopyResep/copy-resep.css" scoped></style>
