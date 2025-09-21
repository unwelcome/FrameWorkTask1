<template>
  <div class="flex flex-col grow items-stretch">
    <header class="flex flex-row h-20 items-center shrink-0">
      <searchBar class="w-[380px]" :placeholder="'Поиск по SID'"/>
    </header>
    <main class="flex flex-col items-stretch bg-bg-2 grow rounded-tl-xl overflow-hidden pl-5 pt-5">
      <div class="flex flex-col grow gap-5 items-stretch scrollable pr-5">
        <applicationsGroup :title="'Активные заявления'" :applications-list="applicationsList" :default-show-body="true"/>
        <applicationsGroup :title="'Завершенные заявления'" :applications-list="applicationsList"/>
        <applicationsGroup :title="'Приостановленные заявления'" :applications-list="applicationsList"/>
        <applicationsGroup :title="'Проваленные заявления'" :applications-list="applicationsList"/>
      </div>
    </main>
  </div>
</template>
<script lang="ts">
import applicationsGroup from './widgets/applicationsGroup.vue';
export default {
  components: {
    applicationsGroup,
  },
  data(){
    return {
      applicationsList: [
        {id: 1, title: 'Неполажка в розетке'},
        {id: 2, title: 'Неполажка в розетке'},
        {id: 3, title: 'Неполажка в розетке'},
        {id: 4, title: 'Неполажка в розетке'},
        {id: 5, title: 'Неполажка в розетке'},
        {id: 6, title: 'Неполажка в розетке'},
        {id: 7, title: 'Неполажка в розетке'},
        {id: 8, title: 'Неполажка в розетке'},
        {id: 9, title: 'Неполажка в розетке'},
      ],
    }
  }
}
</script>