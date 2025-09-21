<template>
  <div class="flex flex-col grow items-stretch">
    <header class="flex flex-row h-20 items-center gap-5 shrink-0">
      <searchBar class="w-[380px]" :placeholder="'Поиск по ФИО'"/>
      <textButton class="btn-main text-lg h-10 px-6" :text="'Добавить сотрудника'"/>
    </header>
    <main class="flex flex-col items-stretch bg-bg-2 grow rounded-tl-xl overflow-hidden pt-5 pl-5">
      <div class="employees-container scrollable pr-5">
        <div v-for="i of 100" class="employee-item">
          <employeeCard :show-btns="i % 2 == 0" :redirect-to-profile="true"/>
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
  }

}
</script>
<style scoped>
  .employees-container {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
    gap: 20px;
  }

  .employee-item {
    min-width: 0; /* Важно для предотвращения переполнения */
  }
</style>