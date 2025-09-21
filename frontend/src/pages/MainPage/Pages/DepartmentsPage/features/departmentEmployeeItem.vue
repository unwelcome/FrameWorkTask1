<template>
  <div class="flex flex-row gap-2 items-center">
    <userAvatar :id="4" :first-name="'Имя'" :last-name="'Фамилия'" class="text-base cursor-pointer" @click="redirectToUserProfile"/>
    <p class="cursor-default text-lg">{{ userData.second_name + ' ' + userData.first_name + ' ' + userData.third_name }}</p>
  </div>
</template>
<script lang="ts">
import type { PropType } from 'vue';

export default {
  props: {
    userData: {
      type: Object as PropType<{id: number, first_name: string, second_name: string, third_name: string}>,
      required: true,
    },
    redirectToProfile: {
      type: Boolean,
      required: false,
      default: false,
    }
  },
  methods: {
    redirectToUserProfile(){
      if (this.redirectToProfile) this.$router.push({name: 'ProfilePage', params: {id: this.userData.id}});
    }
  }
}
</script>