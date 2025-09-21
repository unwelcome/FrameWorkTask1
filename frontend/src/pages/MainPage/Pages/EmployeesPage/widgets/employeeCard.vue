<template>
  <div class="bg-bg-3 rounded-xl p-2 flex flex-col items-center gap-2 h-full justify-start">
    <userAvatar class="text-4xl p-4 cursor-pointer" :id="employeeData.id" :first-name="employeeData.first_name" :second-name="employeeData.second_name" @click="redirectToUserProfile"/>
    
    <div class="flex flex-col gap-2 w-full px-2 cursor-default">
      <div class="flex flex-col">
        <p>{{ employeeData.second_name }}</p>
        <p>{{ employeeData.first_name }}</p>
        <p>{{ employeeData.third_name }}</p>
      </div>
      <div class="flex flex-col">
        <p>Роль: <span>инженер</span></p>
        <p>Отдел: <span>Электрика</span></p>
        <p>Стаж: <span>20 лет</span></p>
      </div>
    </div>

    <div v-if="showBtns" class="flex flex-row gap-6 items-center justify-center">
      <iconButton class="btn-change p-1.5">
        <img class="w-5 h-5" src="../../../../../assets/icons/icon_edit.svg"/>
      </iconButton>
      <iconButton class="btn-delete p-1.5">
        <img class="w-5 h-5" src="../../../../../assets/icons/icon_user_remove.svg"/>
      </iconButton>
    </div>
  </div>
</template>
<script lang="ts">
import type { PropType } from 'vue';

export default {
  props: {
    employeeData: {
      type: Object as PropType<{id: number, first_name: string, second_name: string, third_name: string}>,
      required: true,
    },
    showBtns: {
      type: Boolean,
      required: false,
      default: false,
    },
    redirectToProfile: {
      type: Boolean,
      required: false,
      default: false,
    }
  },
  methods: {
    redirectToUserProfile(){
      if (this.redirectToProfile) this.$router.push({name: 'ProfilePage', params: {id: 12}});
    }
  }
}
</script>