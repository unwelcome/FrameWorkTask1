<template>
  <div class="flex flex-row items-center gap-2 text-lg cursor-pointer" @click="$emit('select', data)">
    <div class="w-10 h-10 flex flex-col items-center justify-center shrink-0 rounded-full bg-green-500">
      <p>{{ data.id }}</p>
    </div>
    <p>{{ data.fio }}</p>
  </div>
</template> 
<script lang="ts">
import type { PropType } from 'vue';


interface User {
  id: number,
  fio: string
}

export default {
  emits: ['select'],
  props: {
    data:{
      type: Object as PropType<User>,
      required: true,
    } 
  }
}
</script>