<script setup lang="ts">
import { CalendarClock, Moon, Sun, UserRound, UsersRound } from '@lucide/vue'

defineProps({
  formattedDate: { type: String, required: true },
  formattedTime: { type: String, required: true },
  isDark: Boolean,
  user: { type: Object, default: null },
  onlineCount: { type: Number, default: 1 },
})

defineEmits(['toggle-theme'])
</script>

<template>
  <nav class="status-bar">
    <div class="date-time">
      <CalendarClock :size="16" />
      <span>{{ formattedDate }}</span>
      <strong>{{ formattedTime }} WITA</strong>
    </div>
    <span class="status-divider"></span>
    <div class="status-user">
      <UserRound :size="16" />
      <b>{{ user?.name || 'Belum login' }}</b>
      <code>{{ user?.username || 'guest' }}</code>
    </div>
    <span class="status-divider wide-only"></span>
    <div class="hospital-status wide-only"><i></i> RS Prof. Dr. H. Aloei Saboe</div>
    <div class="status-actions">
      <span class="online-badge"><UsersRound :size="15" /> {{ onlineCount }} Online</span>
      <button type="button" @click="$emit('toggle-theme')">
        <Sun v-if="isDark" :size="15" />
        <Moon v-else :size="15" />
        {{ isDark ? 'Light' : 'Dark' }}
      </button>
    </div>
  </nav>
</template>
