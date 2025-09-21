<template>
  <div class="flex flex-col w-full bg-bg-3 rounded-lg gap-3">
    <header class="flex flex-col justify-center w-full pr-3">
      <div class="flex flex-row items-center">

        <div class="flex flex-row gap-2 items-center p-2 pr-5 cursor-pointer" @click="showBody=!showBody">
          <div class="flex flex-col justify-center items-center">
            <img src="../assets/icons/icon_arrow_up.svg" class="select-none" :class="{'rotate-180': showBody}"/>
          </div>
          <h1 class="text-lg select-none">{{ title }}</h1>
        </div>

        <div class="border-b-2 border-bg-input grow"></div> <!-- Horizontal line -->
      </div>
    </header>

    <main v-if="showBody" class="flex flex-col items-stretch grow">
      <slot></slot>
    </main>
  </div>
</template>
<script lang="ts">
export default {
  props: {
    title: {
      type: String,
      required: true,
    },
    defaultShowBody: {
      type: Boolean, 
      required: false,
      default: false,
    }
  },
  data(){
    return {
      showBody: this.defaultShowBody,
    }
  }
}
</script>