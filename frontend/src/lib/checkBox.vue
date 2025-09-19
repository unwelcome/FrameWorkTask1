<template>
  <div class="flex flex-row gap-2 items-center">
    <div class="flex flex-col p-0.5 border-2 border-bg-input rounded-lg cursor-pointer" @click="toggle">
      <div class="w-5 h-5 select-none">
        <img v-if="isChecked" src="../assets/icons/icon_check.svg"/>
      </div>
    </div>
    <div class="cursor-pointer" @click="toggle">
      <p class="text-text-main select-none">{{ placeholder }}</p>
    </div>
  </div>
</template>
<script lang="ts">
export default {
  emits: ['change:checked'],
  props: {
    placeholder: {
      type: String,
      required: false,
      default: '',
    },
    defaultChecked: {
      type: Boolean,
      required: false,
      default: false,
    }
  },
  data() {
    return {
      isChecked: this.defaultChecked,
    }
  },
  methods: {
    toggle() {
      this.isChecked = !this.isChecked;
      this.$emit('change:checked', this.isChecked);
    }
  }
}
</script>