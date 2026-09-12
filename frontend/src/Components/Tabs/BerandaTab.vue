<script setup lang="ts">
// @ts-nocheck -- migrasi TypeScript bertahap; kontrak data modul lama belum sepenuhnya bertipe.
import { ArrowRight, CheckCircle2, CircleAlert, LayoutGrid, LoaderCircle, LockKeyhole, MapPin, Server } from '@lucide/vue'

defineProps({
  dashboardLoading: Boolean,
  dashboardError: { type: String, default: '' },
  quickStats: { type: Array, required: true },
  serviceStatuses: { type: Array, required: true },
  isAuthenticated: Boolean,
  isMenuDisabled: { type: Function, default: () => false },
})

const emit = defineEmits(['select-tab'])
</script>

<template>
  <section class="beranda" aria-label="Beranda SIRAPI">
    <header class="beranda-heading">
      <div>
        <h1>Beranda SIRAPI</h1>
        <p>Ringkasan layanan dan koneksi dalam satu tampilan.</p>
      </div>
      <button type="button" class="beranda-menu" @click="emit('select-tab', 'Menu')">
        <LayoutGrid :size="18" />
        Jelajahi menu
        <ArrowRight :size="16" />
      </button>
    </header>

    <div class="beranda-overview">
      <article class="beranda-hero">
        <img class="beranda-photo" src="/img/benner.png" alt="Gedung RS Prof. Dr. H. Aloei Saboe" />
        <div class="beranda-hero-content">
          <div class="beranda-brand">
            <img src="/img/icon_rsas.png" alt="Logo RSAS" />
            <div>
              <span class="beranda-eyebrow">Sistem Informasi Rumah Sakit Pelayanan Terintegrasi</span>
              <h2>RS Prof. Dr. H. Aloei Saboe</h2>
            </div>
          </div>
          <p class="beranda-location">
            <MapPin :size="15" /> Kota Gorontalo
          </p>
          <p class="beranda-description">
            Mendukung administrasi medis dan pelayanan digital, serta koordinasi
            antardepartemen yang cepat, responsif, dan aman.
          </p>
          <div
            class="beranda-connection"
            :class="{ 'is-error': dashboardError, 'is-pending': dashboardLoading || !isAuthenticated }"
            role="status"
          >
            <LoaderCircle v-if="dashboardLoading" :size="18" class="beranda-spinner" />
            <LockKeyhole v-else-if="!isAuthenticated" :size="18" />
            <CircleAlert v-else-if="dashboardError" :size="18" />
            <CheckCircle2 v-else :size="18" />
            <span>{{
              !isAuthenticated
                ? 'Silakan login untuk melihat data pelayanan.'
                : dashboardLoading
                  ? 'Memeriksa koneksi SIMRS...'
                  : dashboardError || 'Database SIMRS terhubung dan siap digunakan.'
            }}</span>
          </div>
        </div>
      </article>

      <aside class="beranda-services" aria-labelledby="beranda-koneksi">
        <header>
          <div>
            <h2 id="beranda-koneksi">Status koneksi</h2>
            <p>Ketersediaan layanan terintegrasi</p>
          </div>
          <Server :size="21" />
        </header>
        <div class="beranda-service-list" aria-live="polite">
          <div v-for="status in serviceStatuses" :key="status.label" class="beranda-service">
            <span>{{ status.label }}</span>
            <strong :class="status.tone">
              <i aria-hidden="true"></i>
              {{ status.value }}
            </strong>
          </div>
        </div>
      </aside>
    </div>

    <section class="beranda-summary" aria-labelledby="beranda-ringkasan" :aria-busy="dashboardLoading">
      <header>
        <h2 id="beranda-ringkasan">Ringkasan pelayanan</h2>
      </header>
      <div class="beranda-stats">
        <button
          v-for="stat in quickStats"
          :key="stat.label"
          type="button"
          :disabled="isMenuDisabled(stat.tab)"
          :title="isMenuDisabled(stat.tab) ? 'Anda tidak memiliki akses ke modul ini' : `Buka ${stat.tab}`"
          @click="emit('select-tab', stat.tab)"
        >
          <span class="beranda-stat-heading">
            <span>{{ stat.label }}</span>
            <component :is="stat.icon" :size="20" />
          </span>
          <strong>{{
            dashboardLoading
              ? '…'
              : dashboardError || !isAuthenticated
                ? '—'
                : Number(stat.value).toLocaleString('id-ID')
          }}</strong>
          <span class="beranda-stat-link">
            <template v-if="isMenuDisabled(stat.tab)">
              <LockKeyhole :size="13" /> Akses terbatas
            </template>
            <template v-else>
              Buka layanan <ArrowRight :size="14" />
            </template>
          </span>
        </button>
      </div>
    </section>
  </section>
</template>

<style src="./beranda.css" scoped></style>
