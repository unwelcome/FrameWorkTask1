<template>
  <div class="flex flex-col grow items-stretch">
    <header class="flex flex-row h-20 items-center gap-5 shrink-0">
      <searchBar class="w-[380px]" :placeholder="'Поиск по ФИО'" @change:input="(text: string) => employeeQuery = text"/>
      <textButton class="btn-main text-lg h-10 px-6" :text="'Добавить сотрудника'"/>
    </header>
    <main class="flex flex-col items-stretch bg-bg-2 grow rounded-tl-xl overflow-hidden pt-5 pl-5">
      <div class="flex flex-col justify-start items-stretch">
        <div class="employees-container scrollable pr-5 grow">
          <div v-for="employee in employeesFilteredList" :key="employee.id" class="employee-item h-[300px] max-w-[250px]">
            <employeeCard :employee-data="employee" :show-btns="employee.id % 2 == 0" :redirect-to-profile="true"/>
          </div>
        </div>
      </div>
    </main>
  </div>
</template>
<script lang="ts">
import employeeCard from './widgets/employeeCard.vue';
export default {
  components: {
    employeeCard,
  },
  data(){
    return {
      employeeQuery: '',

      employeesList: [
        {id: 1, first_name: 'Имя', second_name: 'Фамилия1', third_name: 'Отчество'},
        {id: 2, first_name: 'Имя', second_name: 'Фамилия2', third_name: 'Отчество'},
        {id: 3, first_name: 'Имя', second_name: 'Фамилия2', third_name: 'Отчество'},
        {id: 4, first_name: 'Имя', second_name: 'Фамилия2', third_name: 'Отчество'},
        {id: 5, first_name: 'Имя', second_name: 'Фамилия2', third_name: 'Отчество'},
        {id: 6, first_name: 'Имя', second_name: 'Фамилия3', third_name: 'Отчество'},
        {id: 7, first_name: 'Имя', second_name: 'Фамилия4', third_name: 'Отчество'},
        {id: 8, first_name: 'Имя', second_name: 'Фамилия5', third_name: 'Отчество'},
        {id: 9, first_name: 'Имя', second_name: 'Фамилия6', third_name: 'Отчество'},
        {id: 10, first_name: 'Имя', second_name: 'Фамилия7', third_name: 'Отчество'},
        {id: 11, first_name: 'Имя', second_name: 'Фамилия8', third_name: 'Отчество'},
        {id: 12, first_name: 'Имя', second_name: 'Фамилия9', third_name: 'Отчество'},
        {id: 13, first_name: 'Имя', second_name: 'Фамилия10', third_name: 'Отчество'},
        {id: 14, first_name: 'Имя', second_name: 'Фамилия11', third_name: 'Отчество'},
        {id: 15, first_name: 'Имя', second_name: 'Фамилия12', third_name: 'Отчество'},
        {id: 16, first_name: 'Имя', second_name: 'Фамилия13', third_name: 'Отчество'},
        {id: 17, first_name: 'Имя', second_name: 'Фамилия14', third_name: 'Отчество'},
        {id: 18, first_name: 'Имя', second_name: 'Фамилия15', third_name: 'Отчество'},
        {id: 19, first_name: 'Имя', second_name: 'Фамилия16', third_name: 'Отчество'},
      ],

      employeesFilteredList: [] as {id: number, first_name: string, second_name: string, third_name: string}[],
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
<style scoped>
  .employees-container {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
    gap: 20px;
    /* justify-items: start; */
  }
</style>