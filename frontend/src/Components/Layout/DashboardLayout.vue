<script setup>
import {
  Search,
  X,
} from '@lucide/vue'
import { computed } from 'vue'
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

function selectMenu(label) {
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
            <input v-model="menuSearchModel" type="search" placeholder="Ketik nama menu..." autofocus />
          </label>
          <div class="menu-list">
            <button
              v-for="menu in filteredMenus"
              :key="menu.label"
              type="button"
              :disabled="isMenuDisabled(menu.label)"
              @click="selectMenu(menu.label)"
            >
              <i :class="`tone-bg-${menu.tone}`"><component :is="menu.icon" :size="21" /></i>
              <span><strong>{{ menu.label }}</strong><small>{{ menu.description }}</small></span>
            </button>
            <p v-if="filteredMenus.length === 0">Menu tidak ditemukan.</p>
          </div>
          <footer>{{ filteredMenus.length }} menu tersedia</footer>
        </section>
      </div>
    </Transition>
  </main>
</template>
