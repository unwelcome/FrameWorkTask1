<template>
  <div class="flex flex-col w-full bg-bg-3 rounded-lg gap-3">
    <header class="flex flex-col justify-center w-full pr-3">
      <div class="flex flex-row items-center">

        <div class="flex flex-row gap-2 items-center p-2 pr-5 cursor-pointer" @click="showBody=!showBody">
          <div class="flex flex-col justify-center items-center">
            <img src="../../../../../assets/icons/icon_arrow_up.svg" class="select-none" :class="{'rotate-180': showBody}"/>
          </div>
          <h1 class="text-lg select-none">{{ departmentData.title }}</h1>
        </div>

        <div class="border-b-2 border-bg-input grow"></div> <!-- Horizontal line -->
      </div>
    </header>

    <main v-if="showBody" class="flex flex-row gap-5 items-stretch px-5 pb-5">
      <div class="grow">
        <department-employee-list :title="'Менеджеры:'"/>
      </div>
      <div class="grow">
        <department-employee-list :title="'Инженеры:'"/>
      </div>
      <div class="flex flex-col gap-2 border-2 border-border-main p-2 rounded-lg shrink-0 max-w-1/3">
        <h2>Информация:</h2>
        <div class="flex flex-col">
          <h3>Название: <span>Отдел Строительства</span></h3>
          <h3>Создан: <span>14.04.2012</span></h3>
          <h3>Кол-во менеджеров: <span>4</span></h3>
          <h3>Кол-во инженеров: <span>15</span></h3>
          <h3>Создано заявок: <span>14928</span></h3>
          <h3>Завершено заявок: <span>54921</span></h3>
          <h3>Активно заявок: <span>134</span></h3>
        </div>
        <div class="flex flex-row gap-6 items-center justify-center mt-auto">
          <textButton class="btn-change" :text="'Изменить'"/>
          <textButton class="btn-delete" :text="'Удалить отдел'"/>
        </div>
      </div>
    </main>
  </div>
</template>
<script lang="ts">
import type { PropType } from 'vue';
import departmentEmployeeList from './departmentEmployeeList.vue';
export default {
  components: {
    departmentEmployeeList,
  },
  props: {
    departmentData: {
      type: Object as PropType<{id: number, title: string}>,
      required: true,
    },
    defaultShowBody: {
      type: Boolean,
      required: false,
      default: false,
    },
  },
  data(){
    return {
      showBody: this.defaultShowBody,
    }
  }
}
</script>