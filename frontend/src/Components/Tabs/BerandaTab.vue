<script setup lang="ts">
// @ts-nocheck -- migrasi TypeScript bertahap; kontrak data modul lama belum sepenuhnya bertipe.
import { CheckCircle2, Heart, MapPin, Server } from '@lucide/vue'

defineProps({
  dashboardLoading: Boolean,
  dashboardError: { type: String, default: '' },
  quickStats: { type: Array, required: true },
  serviceStatuses: { type: Array, required: true },
  isAuthenticated: Boolean,
})

const emit = defineEmits(['select-tab'])
</script>

<template>
  <section class="dashboard-grid">
    <div class="dashboard-left">
      <article class="hero-card">
        <img src="/img/benner.png" alt="RS Prof. Dr. H. Aloei Saboe" />
        <div class="hero-overlay"></div>
        <div class="hero-content">
          <span class="simrs-label"><Heart :size="14" /> Sistem Informasi Rumah Sakit Pelayanan Terintegrasi</span>
          <div class="hero-title">
            <img src="/img/icon_rsas.png" alt="Logo RSAS" />
            <h1>SIRAPI</h1>
          </div>
          <p class="location"><MapPin :size="16" /> PROF. DR. H. ALOEI SABOE GORONTALO</p>
          <p class="hero-description">SIRAPI adalah Sistem Informasi Rumah Sakit Pelayanan Terintegrasi untuk administrasi medis dan pelayanan digital. Membantu koordinasi antardepartemen secara cepat, responsif, dan aman.</p>
          <div class="connection-info" :class="{ connection_error: dashboardError }">
            <CheckCircle2 :size="20" />
            {{ dashboardLoading ? 'Memeriksa koneksi database SIMRS lama...' : dashboardError || 'Database SIMRS lama terhubung dan siap digunakan.' }}
          </div>
        </div>
      </article>

      <div class="quick-stats">
        <button v-for="stat in quickStats" :key="stat.label" type="button" :disabled="!isAuthenticated" @click="emit('select-tab', stat.tab)">
          <div><strong>{{ stat.value }}</strong><span :class="`tone-bg-${stat.tone}`"><component :is="stat.icon" :size="20" /></span></div>
          <p>{{ stat.label }}</p>
        </button>
      </div>
    </div>

    <aside class="monitor-card">
      <div>
        <header><div><h2>Koneksi</h2></div><i><Server :size="22" /></i></header>
        <div class="service-list">
          <div v-for="status in serviceStatuses" :key="status.label">
            <span>{{ status.label }}</span>
            <strong :class="status.tone"><i></i>{{ status.value }}</strong>
          </div>
        </div>
      </div>
    </aside>
  </section>
</template>
