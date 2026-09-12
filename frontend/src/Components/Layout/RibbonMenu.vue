<script setup lang="ts">
// @ts-nocheck -- migrasi TypeScript bertahap; kontrak data modul lama belum sepenuhnya bertipe.
import { LockKeyhole } from '@lucide/vue'

const props = defineProps({
  currentTab: { type: String, required: true },
  menus: { type: Array, required: true },
  isMenuDisabled: { type: Function, default: () => false },
  menuOpen: Boolean,
})

const emit = defineEmits(['select'])

function selectMenu(label) {
  if (props.isMenuDisabled(label)) return
  emit('select', label)
}
</script>

<template>
  <nav class="sirapi-ribbon" aria-label="Navigasi utama">
    <button
      v-for="item in menus"
      :key="item.label"
      type="button"
      class="sirapi-ribbon-item"
      :class="[
        { active: currentTab === item.label, logout: item.label === 'Logout', disabled: isMenuDisabled(item.label) },
        `tone-${item.tone}`,
      ]"
      :disabled="isMenuDisabled(item.label)"
      :title="isMenuDisabled(item.label) ? 'Anda tidak memiliki akses ke modul ini' : item.label"
      :aria-label="isMenuDisabled(item.label) ? `${item.label} - tidak memiliki akses` : item.label"
      :aria-current="currentTab === item.label && item.label !== 'Menu' ? 'page' : undefined"
      :aria-expanded="item.label === 'Menu' ? menuOpen : undefined"
      :aria-haspopup="item.label === 'Menu' ? 'dialog' : undefined"
      @click="selectMenu(item.label)"
    >
      <component :is="item.icon" :size="20" />
      <span>{{ item.label }}</span>
      <LockKeyhole v-if="isMenuDisabled(item.label)" class="ribbon-lock" :size="11" aria-hidden="true" />
    </button>
  </nav>
</template>

<style src="./ribbon-menu.css" scoped></style>
