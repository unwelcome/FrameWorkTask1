<template>
  <div class="flex flex-col items-stretch border-2 border-border-main rounded-lg h-full">
    <div class="flex flex-col items-stretch gap-2 p-2">

      <h1 class="text-lg cursor-default">{{ title }}</h1>
      
      <div class="flex flex-col gap-2 items-stretch max-h-[200px] scrollable">
        <div v-for="i of 30" class="flex flex-row gap-2 items-center">
          <departmentEmployeeItem :user-data="{id: 4 + i, first_name: 'Имя', second_name: 'Фамилия', third_name: 'Отчество'}" :redirect-to-profile="true"/>
        </div>
      </div>
    </div>

    <div class="flex flex-col items-stretch">
      <div class="border-b-2 border-border-main"></div> <!-- Horizontal line -->
      <div class="flex flex-col items-stretch p-2 px-4">
        <searchBar :placeholder="'Введите ФИО сотрудника'"/>
      </div>
    </div>
  </div>
</template>
<script lang="ts">
import departmentEmployeeItem from '../features/departmentEmployeeItem.vue';
export default {
  components: {
    departmentEmployeeItem,
  },
  props: {
    title: {
      type: String,
      required: true,
    }
  }
}
</script>