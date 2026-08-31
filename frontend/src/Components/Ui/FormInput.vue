<script setup lang="ts">
import { computed, useAttrs, useId } from 'vue'
import Select from './Select.vue'

defineOptions({ inheritAttrs: false })

const model = defineModel({ default: '' })
const attrs = useAttrs()
const generatedId = useId()

const props = defineProps({
  label: { type: String, required: true },
  jenis: { type: String, default: 'input' },
  type: { type: String, default: 'text' },
  id: { type: String, default: '' },
  placeholder: { type: String, default: '' },
  options: { type: Array, default: () => [] },
  optionLabel: { type: String, default: 'label' },
  optionValue: { type: String, default: 'value' },
  rows: { type: Number, default: 3 },
  maxlength: { type: [Number, String], default: undefined },
  inputmode: { type: String, default: undefined },
  step: { type: [Number, String], default: undefined },
  required: Boolean,
  disabled: Boolean,
  hint: { type: String, default: '' },
  error: { type: String, default: '' },
})

const inputId = computed(() => props.id || generatedId)
type ModeInput = 'none' | 'text' | 'search' | 'email' | 'tel' | 'url' | 'numeric' | 'decimal'
const inputMode = computed<ModeInput | undefined>(() => props.inputmode as ModeInput | undefined)
const rootClass = computed(() => ['form-input-field', attrs.class, { invalid: props.error }])
const inheritedAttrs = computed(() => {
  const { class: _class, ...rest } = attrs
  return rest
})
</script>

<template>
  <label :class="rootClass" :for="inputId">
    <span class="form-input-label">
      {{ label }}
      <i v-if="required" aria-hidden="true">*</i>
    </span>

    <textarea
      v-if="jenis === 'textarea'"
      :id="inputId"
      v-model="model"
      class="form-input-element form-input-textarea"
      :rows="rows"
      :placeholder="placeholder"
      :maxlength="maxlength"
      :required="required"
      :disabled="disabled"
      v-bind="inheritedAttrs"
    />

    <Select
      v-else-if="jenis === 'select'"
      :id="inputId"
      v-model="model"
      class="form-input-element form-input-select"
      :options="options"
      :option-label="optionLabel"
      :option-value="optionValue"
      :placeholder="placeholder || `Pilih ${label}`"
      :disabled="disabled"
      v-bind="inheritedAttrs"
    />

    <input
      v-else
      :id="inputId"
      v-model="model"
      class="form-input-element"
      :type="type"
      :placeholder="placeholder"
      :maxlength="maxlength"
      :inputmode="inputMode"
      :step="step"
      :required="required"
      :disabled="disabled"
      v-bind="inheritedAttrs"
    />

    <small v-if="error" class="form-input-message error">{{ error }}</small>
    <small v-else-if="hint" class="form-input-message">{{ hint }}</small>
  </label>
</template>
