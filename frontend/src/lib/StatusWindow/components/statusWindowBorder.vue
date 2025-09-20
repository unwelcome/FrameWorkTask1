<template>
  <div class="status-window status-window-border" :class="`status-window-${status}`">
    <div class="status-window-body">

      <img v-if="status === 'error'" class="status-window-image" src="../assets/icons/icon-error.svg"/>
      <img v-else-if="status === 'success'" class="status-window-image" src="../assets/icons/icon-success.svg" />
      <img v-else-if="status === 'loading'" class="status-window-image" src="../assets/icons/icon-loading.svg" />
      <img v-else-if="status === 'info'" class="status-window-image" src="../assets/icons/icon-info.svg" />

      <div class="status-window-border-text-wrapper">
        <p class="status-window-border-header">{{ headerText }}</p>
        <p class="status-window-text">
          {{ animatedText }}
        </p>
      </div>

      <img v-if="closable" class="status-window-close" src="../assets/icons/icon-close.svg" @click="closeWindow"/>
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
    headerText: {
      type: String,
      required: false,
      default: 'Alert'
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
      intervalID: null as null | number,
    }
  },
  mounted() {
    if(this.showAnimation && this.status === 'loading'){
      let dotCounter = 0;
      this.intervalID = setInterval(() => {
        dotCounter = (++dotCounter) % 4;
        this.animatedText = this.text + '.'.repeat(dotCounter);
      }, 1000);
    } 
  },
  methods: {
    closeWindow(){
      if(this.intervalID !== null) clearInterval(this.intervalID);
      this.$emit('closeWindow');
    }
  }
}
</script>