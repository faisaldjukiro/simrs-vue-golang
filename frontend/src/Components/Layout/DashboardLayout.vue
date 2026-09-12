<script setup lang="ts">
// @ts-nocheck -- migrasi TypeScript bertahap; kontrak data modul lama belum sepenuhnya bertipe.
import {
  Search,
  X,
  Folder,
  ArrowLeft,
  ChevronRight,
  LayoutGrid,
  LockKeyhole,
} from '@lucide/vue'
import { computed, nextTick, ref, watch } from 'vue'
import FormInput from '../Ui/FormInput.vue'
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
const menuDialog = ref<HTMLDialogElement | null>(null)

watch(menuSearchModel, (val) => {
  if (val) activeCategory.value = null
})
watch(menuOpenModel, (val) => {
  if (val) activeCategory.value = null
})

// Dialog native mengelola fokus keyboard dan mengembalikannya saat ditutup.
watch(
  () => props.menuOpen,
  async (open) => {
    await nextTick()
    if (open && !menuDialog.value?.open) menuDialog.value?.showModal()
    if (!open && menuDialog.value?.open) menuDialog.value.close()
  },
  { immediate: true },
)

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

const visibleGroups = computed(() => {
  if (!menuSearchModel.value.trim() && activeCategory.value) {
    return { [activeCategory.value]: groupedMenus.value[activeCategory.value] || [] }
  }
  return groupedMenus.value
})

function selectMenu(label: string) {
  if (props.isMenuDisabled(label)) return
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
      :menu-open="menuOpen"
      :is-menu-disabled="isMenuDisabled"
      @select="selectMenu"
    />

    <div class="dashboard-body">
      <div class="dashboard-ambient" aria-hidden="true"></div>
      <nav v-if="currentTab !== 'Menu'" class="shell-location" aria-label="Lokasi halaman">
        <button type="button" @click="selectMenu('Beranda')">Beranda</button>
        <ChevronRight :size="14" aria-hidden="true" />
        <span aria-current="page">{{ currentTab }}</span>
        <button
          type="button"
          class="shell-browse"
          aria-haspopup="dialog"
          :aria-expanded="menuOpen"
          @click="selectMenu('Menu')"
        >
          <LayoutGrid :size="16" /> Semua menu
        </button>
      </nav>
      <slot></slot>

      <AppFooter :is-dark="isDark" />
    </div>

    <dialog
      ref="menuDialog"
      class="shell-menu-modal"
      aria-labelledby="shell-menu-title"
      @cancel.prevent="menuOpenModel = false"
      @close="menuOpenModel = false"
      @click.self="menuOpenModel = false"
    >
      <section class="shell-menu-panel">
        <header>
          <div>
            <i><Search :size="20" /></i>
            <div>
              <h2 id="shell-menu-title">Menu SIRAPI</h2>
              <p>Temukan layanan berdasarkan nama atau kategori.</p>
            </div>
          </div>
          <button type="button" aria-label="Tutup menu" @click="menuOpenModel = false">
            <X :size="20" />
          </button>
        </header>
        <div class="shell-menu-search">
          <FormInput
            v-model="menuSearchModel"
            label="Cari menu"
            type="search"
            placeholder="Nama layanan, laporan, atau kategori..."
            autofocus
          />
        </div>
        <div class="shell-menu-list">
          <template v-if="!menuSearchModel.trim() && !activeCategory">
            <button
              v-for="(menus, catName) in groupedMenus"
              :key="catName"
              type="button"
              @click="activeCategory = String(catName)"
            >
              <i class="shell-menu-icon"><Folder :size="21" /></i>
              <span>
                <strong>{{ catName }}</strong>
                <small>{{ menus.length }} modul</small>
              </span>
              <ChevronRight class="shell-menu-arrow" :size="16" aria-hidden="true" />
            </button>
          </template>
          <template v-else>
            <button
              v-if="!menuSearchModel.trim() && activeCategory"
              class="shell-menu-back"
              type="button"
              @click="activeCategory = null"
            >
              <ArrowLeft :size="16" /> Kembali ke kategori
            </button>
            <template v-for="(menus, catName) in visibleGroups" :key="catName">
              <h3 class="shell-menu-category">
                {{ catName }}
                <span>{{ menus.length }} modul</span>
              </h3>
              <button
                v-for="menu in menus"
                :key="menu.label"
                type="button"
                :disabled="isMenuDisabled(menu.label)"
                :class="{ 'is-active': currentTab === menu.label }"
                :aria-current="currentTab === menu.label ? 'page' : undefined"
                :title="isMenuDisabled(menu.label) ? 'Anda tidak memiliki akses ke modul ini' : menu.label"
                @click="selectMenu(menu.label)"
              >
                <i class="shell-menu-icon">
                  <component :is="menu.icon" :size="21" />
                </i>
                <span>
                  <strong>{{ menu.label }}</strong>
                  <small>{{ menu.description }}</small>
                  <small v-if="isMenuDisabled(menu.label)" class="shell-access-note">
                    <LockKeyhole :size="12" /> Akses terbatas
                  </small>
                  <small v-else-if="currentTab === menu.label" class="shell-access-note">
                    Sedang dibuka
                  </small>
                </span>
              </button>
            </template>
          </template>
          <p v-if="filteredMenus.length === 0" class="shell-menu-empty" role="status">
            Menu tidak ditemukan. Coba nama layanan atau kategori lain.
          </p>
        </div>
        <footer>
          <span>{{ filteredMenus.length }} menu{{ menuSearchModel.trim() ? ' ditemukan' : ' terdaftar' }}</span>
          <span><LockKeyhole :size="12" /> Menu terkunci memerlukan hak akses.</span>
        </footer>
      </section>
    </dialog>
  </main>
</template>

<style src="./dashboard-navigation.css" scoped></style>
