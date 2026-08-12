<script setup>
import {
  CheckCircle2,
  Eye,
  EyeOff,
  Fingerprint,
  LoaderCircle,
  LockKeyhole,
  ShieldCheck,
  Sparkles,
  UserRound,
} from '@lucide/vue'
import { computed, onMounted, reactive, ref } from 'vue'

defineProps({
  loading: Boolean,
  errorMessage: { type: String, default: '' },
})

const emit = defineEmits(['submit'])
const form = reactive({ username: '', password: '', remember: false })
const showPassword = ref(false)
const loaded = ref(false)
const year = computed(() => new Date().getFullYear())

onMounted(() => window.setTimeout(() => { loaded.value = true }, 80))

function submit() {
  emit('submit', { ...form })
}
</script>

<template>
  <main class="login-page">
    <div class="background-gradient"></div>
    <div class="ambient" aria-hidden="true">
      <span class="orb orb-one"></span>
      <span class="orb orb-two"></span>
      <span class="orb orb-three"></span>
      <i v-for="index in 8" :key="index" :class="`star star-${index}`"></i>
    </div>

    <section class="login-shell" :class="{ loaded }">
      <aside class="hospital-panel">
        <img class="hospital-image" src="/img/benner.png" alt="RS Prof. Dr. H. Aloei Saboe" />
        <div class="hospital-overlay"></div>
        <div class="hospital-bottom-fade"></div>
        <div class="accent-line"></div>

        <div class="hospital-content">
          <header class="brand">
            <img src="/img/icon_rsas.png" alt="RSAS" />
            <div>
              <p>SIRAVA</p>
              <h1>RS Prof. Dr. H. Aloei Saboe</h1>
            </div>
          </header>

          <div class="hospital-message">
            <span class="status-pill">
              <CheckCircle2 :size="16" />
              Sistem pelayanan aktif
            </span>
            <h2>Ruang kerja digital untuk pelayanan yang lebih tenang.</h2>
            <p>SIRAVA, Sistem Informasi Rumah Sakit Terintegrasi, menyatukan data registrasi, rawat jalan, IGD, penunjang, dan farmasi dalam satu portal yang rapi dan mudah dipantau.</p>
          </div>

          <div class="stat-grid">
            <article><strong>24/7</strong><span>Monitoring</span></article>
            <article><strong>SIRAVA</strong><span>Terintegrasi</span></article>
            <article><strong>RSAS</strong><span>Gorontalo</span></article>
          </div>
        </div>
      </aside>

      <div class="form-panel">
        <div class="form-container">
          <header class="mobile-brand brand">
            <img src="/img/icon_rsas.png" alt="RSAS" />
            <div>
              <p>SIRAVA</p>
              <h1>RS Prof. Dr. H. Aloei Saboe</h1>
            </div>
          </header>

          <div class="form-heading">
            <span class="login-pill"><Fingerprint :size="16" /> Login petugas</span>
            <h2>Selamat bertugas</h2>
            <p>Silakan masuk untuk melanjutkan pelayanan pasien hari ini.</p>
          </div>

          <form @submit.prevent="submit">
            <label class="field">
              <span>Username</span>
              <div class="input-wrap">
                <UserRound :size="20" :class="{ active: form.username }" />
                <input v-model.trim="form.username" type="text" autocomplete="username" placeholder="Masukkan username" autofocus required />
              </div>
            </label>

            <label class="field">
              <span>Password</span>
              <div class="input-wrap">
                <LockKeyhole :size="20" :class="{ active: form.password }" />
                <input v-model="form.password" :type="showPassword ? 'text' : 'password'" autocomplete="current-password" placeholder="Masukkan password" required />
                <button class="icon-button" type="button" :aria-label="showPassword ? 'Sembunyikan password' : 'Tampilkan password'" @click="showPassword = !showPassword">
                  <EyeOff v-if="showPassword" :size="20" />
                  <Eye v-else :size="20" />
                </button>
              </div>
            </label>

            <div v-if="errorMessage" class="form-error" role="alert">
              <ShieldCheck :size="17" />
              {{ errorMessage }}
            </div>

            <label class="remember">
              <input v-model="form.remember" type="checkbox" />
              <span>Ingat sesi</span>
            </label>

            <button class="submit-button" type="submit" :disabled="loading">
              <i></i>
              <LoaderCircle v-if="loading" class="spin" :size="20" />
              <Sparkles v-else :size="20" />
              <span>{{ loading ? 'Memeriksa akun...' : 'Masuk SIRAVA' }}</span>
            </button>
          </form>

          <div class="security-note">
            <ShieldCheck :size="17" />
            <p><strong>Akses aman.</strong> Aktivitas masuk mengikuti akun dan hak akses yang diberikan administrator.</p>
          </div>

          <footer>SIRAVA &copy; {{ year }} RS Prof. Dr. H. Aloei Saboe</footer>
        </div>
      </div>
    </section>
  </main>
</template>
