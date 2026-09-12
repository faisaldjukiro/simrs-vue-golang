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
  <header class="status-bar shell-status">
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
      <button
        type="button"
        :aria-label="isDark ? 'Aktifkan mode terang' : 'Aktifkan mode gelap'"
        :title="isDark ? 'Aktifkan mode terang' : 'Aktifkan mode gelap'"
        @click="$emit('toggle-theme')"
      >
        <Sun v-if="isDark" :size="15" />
        <Moon v-else :size="15" />
        <span>{{ isDark ? 'Mode terang' : 'Mode gelap' }}</span>
      </button>
    </div>
  </header>
</template>

<style scoped>
.shell-status {
  height: auto;
  min-height: 46px;
  padding: 7px 24px;
}

.status-user b {
  max-width: 200px;
  overflow: hidden;
  text-overflow: ellipsis;
  font-weight: 600;
}

.status-actions button {
  min-height: 32px;
}

.status-actions button:focus-visible {
  outline: 2px solid currentColor;
  outline-offset: 3px;
}

@media (max-width: 1100px) {
  .wide-only {
    display: none;
  }

  .status-user code {
    display: none;
  }
}

@media (max-width: 600px) {
  .shell-status {
    gap: 8px;
    padding: 8px 12px;
  }

  .date-time {
    flex-wrap: wrap;
    gap: 3px 8px;
    font-size: 11px;
  }

  .date-time > svg {
    display: none;
  }

  .date-time strong {
    padding: 3px 6px;
    letter-spacing: normal;
  }

  .status-actions button {
    width: 34px;
    height: 34px;
    justify-content: center;
    padding: 0;
  }

  .status-actions button span {
    display: none;
  }
}
</style>
