import * as LucideIcons from '@lucide/vue'
const keys = Object.keys(LucideIcons)
  .filter(key => key !== 'default' && key !== 'createLucideIcon' && /^[A-Z]/.test(key) && !key.endsWith('Icon'))
console.log(keys.length)
console.log(keys.slice(0, 10))
