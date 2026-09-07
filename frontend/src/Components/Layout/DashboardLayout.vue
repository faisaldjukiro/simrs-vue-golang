<script setup lang="ts">
// @ts-nocheck -- migrasi TypeScript bertahap; kontrak data modul lama belum sepenuhnya bertipe.
import {
  Search,
  X,
  Folder,
  ArrowLeft
} from '@lucide/vue'
import { computed, ref, watch } from 'vue'
import AppFooter from './AppFooter.vue'
import RibbonMenu from './RibbonMenu.vue'
import TopStatusBar from './TopStatusBar.vue'

const props = defineProps({
  isDark: Boolean,
  formattedDate: { type: String, required: true },
  formattedTime: { type: String, required: true },
  user: { type: Object, default: null },
  currentTab: { type: String, required: true },
  ribbonMenus: { type: Array, required: true },
  filteredMenus: { type: Array, required: true },
  menuOpen: Boolean,
  menuSearch: { type: String, default: '' },
  isMenuDisabled: { type: Function, default: () => false },
})

const emit = defineEmits([
  'select-menu',
  'toggle-theme',
  'update:menuOpen',
  'update:menuSearch',
])

const menuOpenModel = computed({
  get: () => props.menuOpen,
  set: (value) => emit('update:menuOpen', value),
})

const menuSearchModel = computed({
  get: () => props.menuSearch,
  set: (value) => emit('update:menuSearch', value),
})

const activeCategory = ref<string | null>(null)

watch(menuSearchModel, (val) => {
  if (val) activeCategory.value = null
})
watch(menuOpenModel, (val) => {
  if (val) activeCategory.value = null
})

const groupedMenus = computed(() => {
  const groups: Record<string, any[]> = {}
  if (!props.filteredMenus) return groups
  props.filteredMenus.forEach((menu: any) => {
    const cat = menu.category || 'Menu Lainnya'
    if (!groups[cat]) groups[cat] = []
    groups[cat].push(menu)
  })
  return groups
})

function selectMenu(label: string) {
  emit('select-menu', label)
}
</script>

<template>
  <main class="dashboard-app" :class="isDark ? 'theme-dark' : 'theme-light'">
    <TopStatusBar
      :formatted-date="formattedDate"
      :formatted-time="formattedTime"
      :is-dark="isDark"
      :user="user"
      @toggle-theme="emit('toggle-theme')"
    />

    <RibbonMenu
      :current-tab="currentTab"
      :menus="ribbonMenus"
      :is-menu-disabled="isMenuDisabled"
      @select="selectMenu"
    />

    <div class="dashboard-body">
      <div class="dashboard-ambient" aria-hidden="true"></div>
      <slot></slot>

      <AppFooter :is-dark="isDark" />
    </div>

    <Transition name="modal-fade">
      <div v-if="menuOpenModel" class="menu-overlay" @click.self="menuOpenModel = false">
        <section class="menu-dialog">
          <header>
            <div>
              <i><Search :size="20" /></i>
              <span><small>Pencarian Modul</small><strong>Menu SIRAPI</strong></span>
            </div>
            <button type="button" @click="menuOpenModel = false"><X :size="20" /></button>
          </header>
          <label>
            <Search :size="20" />
            <input v-model="menuSearchModel" type="search" placeholder="Ketik nama menu atau kategori..." autofocus />
          </label>
          <div class="menu-list">
            <!-- Mode Folder: Tampilkan Kategori -->
            <template v-if="!menuSearchModel && !activeCategory">
              <button
                v-for="(menus, catName) in groupedMenus"
                :key="catName"
                type="button"
                @click="activeCategory = String(catName)"
              >
                <i class="tone-bg-slate"><Folder :size="21" /></i>
                <span><strong>{{ catName }}</strong><small>{{ menus.length }} Modul</small></span>
              </button>
            </template>
            
            <!-- Mode Kategori: Tampilkan Isi Kategori -->
            <template v-else-if="!menuSearchModel && activeCategory">
              <button class="menu-back-btn" type="button" @click="activeCategory = null">
                <ArrowLeft :size="18" /> <span>Kembali ke Kategori</span>
              </button>
              <h4 class="menu-category-title">{{ activeCategory }}</h4>
              <button
                v-for="menu in groupedMenus[activeCategory]"
                :key="menu.label"
                type="button"
                :disabled="isMenuDisabled(menu.label)"
                @click="selectMenu(menu.label)"
              >
                <i :class="`tone-bg-${menu.tone}`"><component :is="menu.icon" :size="21" /></i>
                <span><strong>{{ menu.label }}</strong><small>{{ menu.description }}</small></span>
              </button>
            </template>

            <!-- Mode Pencarian: Tampilkan Semua yang Cocok -->
            <template v-else>
              <template v-for="(menus, catName) in groupedMenus" :key="catName">
                <h4 class="menu-category-title">{{ catName }}</h4>
                <button
                  v-for="menu in menus"
                  :key="menu.label"
                  type="button"
                  :disabled="isMenuDisabled(menu.label)"
                  @click="selectMenu(menu.label)"
                >
                  <i :class="`tone-bg-${menu.tone}`"><component :is="menu.icon" :size="21" /></i>
                  <span><strong>{{ menu.label }}</strong><small>{{ menu.description }}</small></span>
                </button>
              </template>
            </template>

            <p v-if="filteredMenus.length === 0">Menu tidak ditemukan.</p>
          </div>
          <footer>{{ filteredMenus.length }} menu tersedia</footer>
        </section>
      </div>
    </Transition>
  </main>
</template>
<style scoped>
.menu-category-title {
  grid-column: 1 / -1;
  margin: 10px 0 0;
  padding-bottom: 5px;
  border-bottom: 1px solid var(--line);
  color: var(--muted);
  font-size: 11px;
  font-weight: 850;
  letter-spacing: 0.1em;
  text-transform: uppercase;
}
.menu-category-title:first-child {
  margin-top: 0;
}
.menu-back-btn {
  grid-column: 1 / -1;
  display: flex !important;
  align-items: center;
  justify-content: center;
  gap: 10px;
  min-height: 40px;
  border: 1px dashed var(--line) !important;
  border-radius: 12px;
  color: var(--text) !important;
  background: transparent !important;
  font-size: 13px !important;
  font-weight: 850;
}
.menu-back-btn:hover {
  background: var(--surface-soft) !important;
  border-color: var(--muted) !important;
}
</style>
