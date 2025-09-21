<template>
  <div class="flex flex-col items-stretch border-2 border-border-main rounded-lg h-full">
    <div class="flex flex-col items-stretch gap-2 p-2 grow">

      <h1 class="text-lg cursor-default">{{ title }}</h1>
      
      <div class="flex flex-col gap-2 items-stretch max-h-[200px] scrollable">
        <div v-for="employee in employeesFilteredList" :key="employee.id" class="flex flex-row gap-2 items-center">
          <departmentEmployeeItem :user-data="employee" :redirect-to-profile="true"/>
        </div>
      </div>
    </div>

    <div class="flex flex-col items-stretch">
      <div class="border-b-2 border-border-main"></div> <!-- Horizontal line -->
      <div class="flex flex-col items-stretch p-2 px-4">
        <searchBar :placeholder="'Введите ФИО сотрудника'" @change:input="(text: string) => employeeQuery = text"/>
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
  },
  data() {
    return {
      employeeQuery: '',

      employeesList: [
        {id: 1, first_name: 'Имя', second_name: 'Фамилия1', third_name: 'Отчество'},
        {id: 2, first_name: 'Имя', second_name: 'Фамилия2', third_name: 'Отчество'},
        {id: 3, first_name: 'Имя', second_name: 'Фамилия3', third_name: 'Отчество'},
        {id: 4, first_name: 'Имя', second_name: 'Фамилия4', third_name: 'Отчество'},
        {id: 5, first_name: 'Имя', second_name: 'Фамилия5', third_name: 'Отчество'},
        {id: 6, first_name: 'Имя', second_name: 'Фамилия6', third_name: 'Отчество'},
        {id: 7, first_name: 'Имя', second_name: 'Фамилия7', third_name: 'Отчество'},
        {id: 8, first_name: 'Имя', second_name: 'Фамилия8', third_name: 'Отчество'},
        {id: 9, first_name: 'Имя', second_name: 'Фамилия9', third_name: 'Отчество'},
        {id: 10, first_name: 'Имя', second_name: 'Фамилия10', third_name: 'Отчество'},
        {id: 11, first_name: 'Имя', second_name: 'Фамилия11', third_name: 'Отчество'},
        {id: 12, first_name: 'Имя', second_name: 'Фамилия12', third_name: 'Отчество'},
        {id: 13, first_name: 'Имя', second_name: 'Фамилия13', third_name: 'Отчество'},
        {id: 14, first_name: 'Имя', second_name: 'Фамилия14', third_name: 'Отчество'},
        {id: 15, first_name: 'Имя', second_name: 'Фамилия15', third_name: 'Отчество'},
        {id: 16, first_name: 'Имя', second_name: 'Фамилия16', third_name: 'Отчество'},
      ],

      employeesFilteredList: [] as any[],
    }
  },
  watch: {
    employeeQuery: {
      handler(newValue: string) {
        if (newValue === '') this.employeesFilteredList = this.employeesList;
        else {
          const newValueLower = newValue.toLowerCase();
          this.employeesFilteredList = this.employeesList.filter(item => `${item.second_name} ${item.second_name} ${item.third_name}`.toLowerCase().startsWith(newValueLower));
        }
      },
      immediate: true
    }
  }
}
</script>