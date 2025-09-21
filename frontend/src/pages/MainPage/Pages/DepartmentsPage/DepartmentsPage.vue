<template>
  <div class="flex flex-col grow items-stretch">
    <header class="flex flex-row h-20 items-center gap-5 shrink-0">
      <searchBar class="w-[380px]" :placeholder="'Поиск по названию'" @change:input="(text: string) => departmentQuery = text"/>
      <textButton class="btn-main text-lg h-10 px-6" :text="'Добавить отдел'"/>
    </header>
    <main class="flex flex-col items-stretch bg-bg-2 grow rounded-tl-xl pt-5 pl-5 overflow-hidden">
      <div class="flex flex-col items-stretch gap-5 scrollable pr-5">
        <div v-for="department in departmentsFilteredList" :key="department.id">
          <departmentCard :department-data="department"/>
        </div>
      </div>
    </main>
  </div>
</template>
<script lang="ts">
import departmentCard from './widgets/departmentCard.vue';
export default {
  components: {
    departmentCard,
  },
  data(){
    return {
      departmentQuery: '',

      departmentsList: [
        {id: 1, title: 'Отдел Строительства 1'},
        {id: 2, title: 'Отдел Строительства 11'},
        {id: 3, title: 'Отдел Строительства 12'},
        {id: 4, title: 'Отдел Строительства 12'},
        {id: 5, title: 'Отдел Строительства 123'},
        {id: 6, title: 'Отдел Строительства 124'},
        {id: 7, title: 'Отдел Строительства 2'},
      ],

      departmentsFilteredList: [] as any[],
    }
  },
  watch: {
    departmentQuery:{
      handler(newValue: string) {
        if (newValue === '') this.departmentsFilteredList = this.departmentsList;
        else {
          const newValueLower = newValue.toLowerCase();
          this.departmentsFilteredList = this.departmentsList.filter(item => item.title.toLowerCase().startsWith(newValueLower));
        }
      },
      immediate: true,
    },
  }
}
</script>