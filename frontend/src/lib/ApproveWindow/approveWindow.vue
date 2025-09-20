<template>
  <div v-if="ApproveWindowAPI.getApproveWindowQueue.length !== 0" class="absolute left-1/2 top-10 z-40 rounded-lg overflow-hidden" style="transform: translate(-50%, 0);">
    <div class="bg-sky-600 px-2">
      <h2 class="text-text-main font-medium">{{ getFirstWindow.title }}</h2>
    </div>
    <div class="bg-bg-3 p-4 flex flex-col min-w-60 max-w-96 gap-4 border-2 border-t-0 border-bg-input rounded-b-lg">
      <h3 class="text-justify text-text-main">{{ getFirstWindow.text }}</h3>
      
      <div class="flex flex-row justify-center gap-2 items-center">
        <textButton
          v-for="btn of getFirstWindow.answers" 
          :key="btn.id"
          class="grow"
          :class="{
            'btn-main': btn.type === ApproveWindowAPI.getApproveWindowAnswerType.good,
            'btn-delete':  btn.type === ApproveWindowAPI.getApproveWindowAnswerType.bad,
          }"
          @click="resolveAppworveWindow(btn.id)"
          :text="btn.text" 
        />
      </div>
    </div>
  </div>
</template>
<script lang="ts">
import { useApproveWindowAPI } from './approveWindowAPI';

export default {
  data(){
    return{
      ApproveWindowAPI: useApproveWindowAPI(), 
    }
  },
  computed: {
    getFirstWindow(){
      return this.ApproveWindowAPI.getApproveWindowQueue[0];
    }
  },
  methods: {
    resolveAppworveWindow(returnID: number){
      this.ApproveWindowAPI.resolveFirst(returnID);
    }
  }
}
</script>
