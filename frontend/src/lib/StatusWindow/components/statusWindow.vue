<template>
  <div class="status-window-wrapper">
    <TransitionGroup name="list">
      <div v-for="item in statusWindowApi.getAllStatusWindows" :key="item.id">

        <statusWindowClassic
          v-if="item.type === statusWindowApi.getTypes.classic" 
          :status="statusWindowApi.statusCodesToString(item.status)" 
          :text="item.text" 
          :time="item.time"
          :closable="item.closable"
          :show-animation="item.showAnimation"
          :show-time-bar="item.showTimeBar"
          @close-window="closeStatusWindow(item.id)"/>

        <statusWindowBorder
          v-else-if="item.type === statusWindowApi.getTypes.border" 
          :status="statusWindowApi.statusCodesToString(item.status)" 
          :text="item.text" 
          :time="item.time"
          :closable="item.closable"
          :show-animation="item.showAnimation"
          :show-time-bar="item.showTimeBar"
          :header-text="item.headerText"
          @close-window="closeStatusWindow(item.id)"/>

        <statusWindowDetailed
          v-else 
          :status="statusWindowApi.statusCodesToString(item.status)" 
          :text="item.text" 
          :time="item.time"
          :closable="item.closable"
          :show-animation="item.showAnimation"
          :show-time-bar="item.showTimeBar"
          :header-text="item.headerText"
          @close-window="closeStatusWindow(item.id)"/>

      </div>
    </TransitionGroup>
  </div>
</template>
<script lang="ts">

import '../assets/style/style.css';
import { useStatusWindowAPI } from "../statusWindowAPI";
import statusWindowClassic from './statusWindowClassic.vue';
import statusWindowBorder from './statusWindowBorder.vue';
import statusWindowDetailed from './statusWindowDetailed.vue';

export default {
  components: {
    statusWindowClassic,
    statusWindowBorder,
    statusWindowDetailed,
  },
  data() {
    return{
      statusWindowApi: useStatusWindowAPI(),
    }
  },
  methods: {
    closeStatusWindow(id: number){
      this.statusWindowApi.deleteStatusWindow(id);
    }
  }
};
</script>
<style scoped>
  .list-enter-active,
  .list-leave-active {
    transition: all 0.5s ease-out;
  }
  .list-enter-from,
  .list-leave-to {
    opacity: 0;
    transform: translateX(20px);
  }
</style>