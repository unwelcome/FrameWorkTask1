<template>
  <div class="status-window status-window-classic" :class="`status-window-${status}`">

    <div class="status-window-body">
      <div>
        <img v-if="status === 'error'" class="status-window-image" src="../assets/icons/icon-error.svg"/>
        <img v-else-if="status === 'success'" class="status-window-image" src="../assets/icons/icon-success.svg" />
        <img v-else-if="status === 'loading'" class="status-window-image" src="../assets/icons/icon-loading.svg" />
        <img v-else-if="status === 'info'" class="status-window-image" src="../assets/icons/icon-info.svg" />
      </div>
      <p class="status-window-text">
        {{ animatedText }}
      </p>
      <img v-if="closable" class="status-window-close" src="../assets/icons/icon-close.svg" @click="$emit('closeWindow')"/>
    </div>

    <div v-if="showTimeBar && time > 0" 
      class="status-window-timebar" 
      :style="{'animation': time > 0 ? `timeLine ${time}ms linear forwards` : ''}">
    </div>
  </div>
</template>
<script lang="ts">
export default {
  emits: ['closeWindow'],
  props: {
    status: {
      type: String,
      required: true,
    },
    text: {
      type: String,
      required: true,
    },
    time: {
      type: Number,
      required: true,
    },
    closable:{
      type: Boolean,
      required: false,
      default: false,
    },
    showAnimation: {
      type: Boolean,
      required: false,
      default: false,
    },
    showTimeBar: {
      type: Boolean,
      required: false,
      default: false,
    }
  },
  data(){
    return{
      animatedText: this.text,
    }
  },
  mounted() {
   if(this.showAnimation && this.status === 'loading'){
    let dotCounter = 0;
    setInterval(() => {
      dotCounter = (++dotCounter) % 4;
      this.animatedText = this.text + '.'.repeat(dotCounter);
    }, 1000);
   } 
  }
}
</script>