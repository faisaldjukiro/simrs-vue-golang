<script setup>
const props = defineProps({
  currentTab: { type: String, required: true },
  menus: { type: Array, required: true },
  isMenuDisabled: { type: Function, default: () => false },
})

const emit = defineEmits(['select'])

function selectMenu(label) {
  if (props.isMenuDisabled(label)) return
  emit('select', label)
}
</script>

<template>
  <header class="ribbon-bar">
    <button
      v-for="item in menus"
      :key="item.label"
      type="button"
      class="ribbon-item"
      :class="[
        { active: currentTab === item.label, logout: item.label === 'Logout', disabled: isMenuDisabled(item.label) },
        `tone-${item.tone}`,
      ]"
      :disabled="isMenuDisabled(item.label)"
      @click="selectMenu(item.label)"
    >
      <component :is="item.icon" :size="20" />
      <span>{{ item.label }}</span>
    </button>
  </header>
</template>
