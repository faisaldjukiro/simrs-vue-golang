<script setup lang="ts">
import { onMounted, ref } from 'vue'
import LoginPage from './Components/LoginPage.vue'
import Toast from './Components/Ui/Toast.vue'
import HomePage from './Pages/Home/Home.vue'
import { currentUser, login, logout } from './lib/faisal/api'
import { useNotifikasi } from './lib/shared/useNotifikasi'

const user = ref(null)
const loading = ref(false)
const checkingSession = ref(true)
const errorMessage = ref('')
const loginModalOpen = ref(false)
const logoutConfirmOpen = ref(false)
const notifikasi = useNotifikasi()

function savedToken() {
  return localStorage.getItem('simrs_access_token') || sessionStorage.getItem('simrs_access_token')
}

function clearSession() {
  localStorage.removeItem('simrs_access_token')
  sessionStorage.removeItem('simrs_access_token')
  user.value = null
}

function clearNavigation() {
  Object.keys(localStorage)
    .filter((key) => key.startsWith('sirapi.navigation.') || key.startsWith('sirapi.patient_sidebar.'))
    .forEach((key) => localStorage.removeItem(key))
}

function openLoginModal() {
  errorMessage.value = ''
  loginModalOpen.value = true
}

function closeLoginModal() {
  if (loading.value) return
  errorMessage.value = ''
  loginModalOpen.value = false
}

onMounted(async () => {
  const token = savedToken()
  if (!token) {
    checkingSession.value = false
    return
  }

  try {
    user.value = await currentUser(token)
  } catch {
    clearSession()
    notifikasi.peringatan('Sesi login sudah habis. Silakan login ulang.')
  } finally {
    checkingSession.value = false
  }
})

async function handleLogin(form) {
  if (loading.value) return
  loading.value = true
  errorMessage.value = ''

  try {
    const result = await login({ username: form.username, password: form.password })
    const storage = form.remember ? localStorage : sessionStorage
    clearSession()
    storage.setItem('simrs_access_token', result.access_token)
    user.value = result.user
    loginModalOpen.value = false
    notifikasi.sukses(`Selamat datang, ${result.user.name}`)
  } catch (error) {
    errorMessage.value = error.message
    notifikasi.gagal(error.message || 'Login gagal')
  } finally {
    loading.value = false
  }
}

async function handleLogout() {
  const token = savedToken()
  loading.value = true
  try {
    if (token) await logout(token)
  } catch {
    notifikasi.peringatan('Server tidak merespons, sesi lokal tetap dihapus.')
  } finally {
    clearNavigation()
    clearSession()
    logoutConfirmOpen.value = false
    loading.value = false
    notifikasi.info('Anda sudah logout dari SIRAPI.')
  }
}

function requestLogout() {
  if (loading.value) return
  logoutConfirmOpen.value = true
}
</script>

<template>
  <Toast />

  <div v-if="checkingSession" class="session-loader" aria-live="polite">
    <img src="/img/icon_rsas.png" alt="RSAS" />
    <span class="loader-ring"></span>
    <p>Memeriksa sesi petugas...</p>
  </div>

  <template v-else>
    <HomePage
      :user="user"
      :token="savedToken() || ''"
      :loading="loading"
      @login="openLoginModal"
      @logout="requestLogout"
    />

    <Transition name="modal-fade">
      <div v-if="loginModalOpen" class="auth-modal-overlay" @click.self="closeLoginModal">
        <div class="auth-modal-card">
          <button class="auth-modal-close" type="button" aria-label="Tutup login" @click="closeLoginModal">x</button>
          <LoginPage
            :loading="loading"
            :error-message="errorMessage"
            mode="modal"
            @submit="handleLogin"
          />
        </div>
      </div>
    </Transition>

    <Transition name="modal-fade">
      <div v-if="logoutConfirmOpen" class="logout-confirm-backdrop" @click.self="logoutConfirmOpen = false">
        <section class="logout-confirm-card" role="dialog" aria-modal="true" aria-labelledby="logout-confirm-title">
          <div class="logout-confirm-icon" aria-hidden="true">↪</div>
          <h2 id="logout-confirm-title">Keluar dari SIRAPI?</h2>
          <p>Sesi Anda akan diakhiri dan halaman kerja yang sedang dibuka akan ditutup.</p>
          <div class="logout-confirm-actions">
            <button type="button" class="secondary" :disabled="loading" @click="logoutConfirmOpen = false">Tidak, kembali</button>
            <button type="button" class="danger" :disabled="loading" @click="handleLogout">{{ loading ? 'Mengeluarkan...' : 'Ya, logout' }}</button>
          </div>
        </section>
      </div>
    </Transition>
  </template>
</template>
