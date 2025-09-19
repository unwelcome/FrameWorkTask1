<template>
  <div class="btn">
    <p>{{ text }}</p>
  </div>
</template>
<script lang="ts">
export default {
  props: {
    text: {
      type: String,
      required: true,
    }
  }
}
</script>