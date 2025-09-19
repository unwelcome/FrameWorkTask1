<template>
  <div class="btn">
    <p v-if="reverse">{{ text }}</p>
    <slot></slot>
    <p v-if="!reverse">{{ text }}</p>
  </div>
</template>
<script lang="ts">
export default {
  props: {
    text: {
      type: String,
      required: true,
    },
    reverse: {
      type: Boolean,
      required: false,
      default: false,
    }
  }
}
</script>