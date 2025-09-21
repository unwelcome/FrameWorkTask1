<template>
  <div class="flex flex-col grow items-stretch">
    <header class="flex flex-row h-20 items-center gap-5 shrink-0">
      <searchBar class="w-[380px]" :placeholder="'Поиск по названию'"/>
      <textButton class="btn-main text-lg h-10 px-6" :text="'Добавить отдел'"/>
    </header>
    <main class="flex flex-col items-stretch bg-bg-2 grow rounded-tl-xl pt-5 pl-5 overflow-hidden">
      <div class="flex flex-col items-stretch gap-5 scrollable pr-5">
        <div v-for="i of 10">
          <departmentCard />
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
  }
}
</script>