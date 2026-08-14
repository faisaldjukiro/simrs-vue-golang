<script setup>
import {
  Bed,
  CalendarClock,
  CreditCard,
  FileText,
  IdCard,
  MapPin,
  Phone,
  Stethoscope,
  UserRound,
  VenusAndMars,
} from '@lucide/vue'
import { computed } from 'vue'

const props = defineProps({
  patient: { type: Object, required: true },
  moduleName: { type: String, required: true },
  activeSection: { type: String, required: true },
})

function nilai(...daftar) {
  return daftar.find((item) => item !== undefined && item !== null && String(item).trim() !== '') || '-'
}

function jenisKelamin(kode) {
  const nilaiKode = String(kode || '').trim().toUpperCase()
  if (nilaiKode === 'L') return 'Laki-laki'
  if (nilaiKode === 'P') return 'Perempuan'
  return kode || '-'
}

const poliAtauKamar = computed(() => nilai(props.patient.kamar, props.patient.poliklinik))
const warnaJenisKelamin = computed(() => {
  const kode = String(props.patient.jenis_kelamin || '').trim().toUpperCase()
  if (kode === 'P') return 'gender-female'
  if (kode === 'L') return 'gender-male'
  return 'gender-neutral'
})
const ringkasan = computed(() => [
  { label: jenisKelamin(props.patient.jenis_kelamin), icon: VenusAndMars, className: warnaJenisKelamin.value },
  { label: `Umur ${nilai(props.patient.umur)}`, icon: CalendarClock },
  { label: `Lahir ${nilai(props.patient.tanggal_lahir)}`, icon: CalendarClock },
])

const detailUtama = computed(() => [
  { label: 'No. Rawat', value: nilai(props.patient.no_rawat), icon: FileText, mono: true },
  { label: 'No. RM', value: nilai(props.patient.no_rekam_medis), icon: IdCard, mono: true },
  { label: 'Ruang/Poli', value: poliAtauKamar.value, icon: Bed },
  { label: 'No. Telepon', value: nilai(props.patient.no_telepon), icon: Phone },
  { label: 'Cara Bayar', value: nilai(props.patient.penjamin), icon: CreditCard },
])
</script>

<template>
  <header class="patient-identity-header" :class="warnaJenisKelamin">
    <div class="patient-identity-heading">
      <i><UserRound :size="28" /></i>
      <div>
        <div class="patient-identity-badges">
          <span>{{ activeSection }}</span>
          <small>{{ moduleName }}</small>
        </div>
        <h2>{{ patient.nama_pasien || `RM ${patient.no_rekam_medis}` }}</h2>
        <p>{{ poliAtauKamar }}</p>
        <div class="patient-identity-summary">
          <span v-for="item in ringkasan" :key="item.label" :class="item.className">
            <component :is="item.icon" :size="14" />
            {{ item.label }}
          </span>
        </div>
      </div>
      <b class="patient-identity-status">{{ patient.status || 'Aktif' }}</b>
    </div>

    <div class="patient-identity-grid">
      <article v-for="item in detailUtama" :key="item.label">
        <span><component :is="item.icon" :size="14" />{{ item.label }}</span>
        <strong :class="{ mono: item.mono }" :title="item.value">{{ item.value }}</strong>
      </article>
    </div>

    <div class="patient-identity-contact">
      <article>
        <span><MapPin :size="14" />Alamat</span>
        <strong :title="patient.alamat || 'Alamat belum tersedia'">{{ patient.alamat || 'Alamat belum tersedia' }}</strong>
      </article>
      <article>
        <span><Stethoscope :size="14" />DPJP / Dokter</span>
        <strong :title="patient.dokter || 'DPJP belum tersedia'">{{ patient.dokter || 'DPJP belum tersedia' }}</strong>
      </article>
    </div>
  </header>
</template>
