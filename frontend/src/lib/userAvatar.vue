<template>
  <div class="flex flex-col justify-center items-center rounded-full p-2 aspect-square cursor-default" :class="getBgColor">
    <p class="text-text-main select-none">{{ lastName.charAt(0) + firstName.charAt(0) }}</p>
  </div>
</template>
<script lang="ts">
export default {
  props: {
    id: {
      type: Number,
      required: true,
    },
    lastName: {
      type: String,
      required: true,
    },
    firstName: {
      type: String,
      required: true,
    }
  },
  computed: {
    getBgColor(){
      const colorIndex = (this.id % 16) + 1;
      switch (colorIndex) {
        case 1: return 'bg-avatar-1';
        case 2: return 'bg-avatar-2';
        case 3: return 'bg-avatar-3';
        case 4: return 'bg-avatar-4';
        case 5: return 'bg-avatar-5';
        case 6: return 'bg-avatar-6';
        case 7: return 'bg-avatar-7';
        case 8: return 'bg-avatar-8';
        case 9: return 'bg-avatar-9';
        case 10: return 'bg-avatar-10';
        case 11: return 'bg-avatar-11';
        case 12: return 'bg-avatar-12';
        case 13: return 'bg-avatar-13';
        case 14: return 'bg-avatar-14';
        case 15: return 'bg-avatar-15';
        case 16: return 'bg-avatar-16';
      }
    }
  }
}
</script>