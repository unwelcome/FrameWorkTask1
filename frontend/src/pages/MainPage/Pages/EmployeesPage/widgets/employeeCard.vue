<template>
  <div class="bg-bg-3 rounded-xl p-2 flex flex-col items-center gap-2 h-full justify-start">
    <userAvatar class="text-4xl p-4 cursor-pointer" :id="2" :first-name="'Имя'" :last-name="'Фамилия'" @click="$emit('click:profile')"/>
    
    <div class="flex flex-col gap-2 w-full px-2 cursor-default">
      <div class="flex flex-col">
        <p>Фамилия</p>
        <p>Имя</p>
        <p>Отчетсво</p>
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
export default {
  emits: ['click:profile'],
  props: {
    showBtns: {
      type: Boolean,
      required: false,
      default: false,
    }
  }
}
</script>