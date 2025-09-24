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
        {id: 1, title: 'Проблема в розетке', created_by: {id: 1, first_name: 'Имя', second_name: 'Фамилия'}, created_at: '21.05.2013', status: 'В работе', responsible_manager: {id: 2, first_name: 'Имя', second_name: 'Фамилия'}, responsible_engineer: {id: 3, first_name: 'Имя', second_name: 'Фамилия'}, closed_at: '25.05.2013'},
        {id: 2, title: 'Проблема в розетке', created_by: {id: 1, first_name: 'Имя', second_name: 'Фамилия'}, created_at: '21.05.2013', status: 'В работе', responsible_manager: {id: 2, first_name: 'Имя', second_name: 'Фамилия'}, responsible_engineer: {id: 3, first_name: 'Имя', second_name: 'Фамилия'}, closed_at: '25.05.2013'},
        {id: 3, title: 'Проблема в розетке', created_by: {id: 1, first_name: 'Имя', second_name: 'Фамилия'}, created_at: '21.05.2013', status: 'В работе', responsible_manager: {id: 2, first_name: 'Имя', second_name: 'Фамилия'}, responsible_engineer: {id: 3, first_name: 'Имя', second_name: 'Фамилия'}, closed_at: '25.05.2013'},
        {id: 4, title: 'Проблема в розетке', created_by: {id: 1, first_name: 'Имя', second_name: 'Фамилия'}, created_at: '21.05.2013', status: 'В работе', responsible_manager: {id: 2, first_name: 'Имя', second_name: 'Фамилия'}, responsible_engineer: {id: 3, first_name: 'Имя', second_name: 'Фамилия'}, closed_at: '25.05.2013'},
        {id: 5, title: 'Проблема в розетке', created_by: {id: 1, first_name: 'Имя', second_name: 'Фамилия'}, created_at: '21.05.2013', status: 'В работе', responsible_manager: {id: 2, first_name: 'Имя', second_name: 'Фамилия'}, responsible_engineer: {id: 3, first_name: 'Имя', second_name: 'Фамилия'}, closed_at: '25.05.2013'},
        {id: 6, title: 'Проблема в розетке', created_by: {id: 1, first_name: 'Имя', second_name: 'Фамилия'}, created_at: '21.05.2013', status: 'В работе', responsible_manager: {id: 2, first_name: 'Имя', second_name: 'Фамилия'}, responsible_engineer: {id: 3, first_name: 'Имя', second_name: 'Фамилия'}, closed_at: '25.05.2013'},
        {id: 7, title: 'Проблема в розетке', created_by: {id: 1, first_name: 'Имя', second_name: 'Фамилия'}, created_at: '21.05.2013', status: 'В работе', responsible_manager: {id: 2, first_name: 'Имя', second_name: 'Фамилия'}, responsible_engineer: {id: 3, first_name: 'Имя', second_name: 'Фамилия'}, closed_at: '25.05.2013'},
        {id: 8, title: 'Проблема в розетке', created_by: {id: 1, first_name: 'Имя', second_name: 'Фамилия'}, created_at: '21.05.2013', status: 'В работе', responsible_manager: {id: 2, first_name: 'Имя', second_name: 'Фамилия'}, responsible_engineer: {id: 3, first_name: 'Имя', second_name: 'Фамилия'}, closed_at: '25.05.2013'},
        {id: 9, title: 'Проблема в розетке', created_by: {id: 1, first_name: 'Имя', second_name: 'Фамилия'}, created_at: '21.05.2013', status: 'В работе', responsible_manager: {id: 2, first_name: 'Имя', second_name: 'Фамилия'}, responsible_engineer: {id: 3, first_name: 'Имя', second_name: 'Фамилия'}, closed_at: '25.05.2013'},
        {id: 10, title: 'Проблема в розетке', created_by: {id: 1, first_name: 'Имя', second_name: 'Фамилия'}, created_at: '21.05.2013', status: 'В работе', responsible_manager: {id: 2, first_name: 'Имя', second_name: 'Фамилия'}, responsible_engineer: {id: 3, first_name: 'Имя', second_name: 'Фамилия'}, closed_at: '25.05.2013'},
        {id: 11, title: 'Проблема в розетке', created_by: {id: 1, first_name: 'Имя', second_name: 'Фамилия'}, created_at: '21.05.2013', status: 'В работе', responsible_manager: {id: 2, first_name: 'Имя', second_name: 'Фамилия'}, responsible_engineer: {id: 3, first_name: 'Имя', second_name: 'Фамилия'}, closed_at: '25.05.2013'},
        {id: 12, title: 'Проблема в розетке', created_by: {id: 1, first_name: 'Имя', second_name: 'Фамилия'}, created_at: '21.05.2013', status: 'В работе', responsible_manager: {id: 2, first_name: 'Имя', second_name: 'Фамилия'}, responsible_engineer: {id: 3, first_name: 'Имя', second_name: 'Фамилия'}, closed_at: '25.05.2013'},
        {id: 13, title: 'Проблема в розетке', created_by: {id: 1, first_name: 'Имя', second_name: 'Фамилия'}, created_at: '21.05.2013', status: 'В работе', responsible_manager: {id: 2, first_name: 'Имя', second_name: 'Фамилия'}, responsible_engineer: {id: 3, first_name: 'Имя', second_name: 'Фамилия'}, closed_at: '25.05.2013'},
        {id: 14, title: 'Проблема в розетке', created_by: {id: 1, first_name: 'Имя', second_name: 'Фамилия'}, created_at: '21.05.2013', status: 'В работе', responsible_manager: {id: 2, first_name: 'Имя', second_name: 'Фамилия'}, responsible_engineer: {id: 3, first_name: 'Имя', second_name: 'Фамилия'}, closed_at: '25.05.2013'},
        {id: 15, title: 'Проблема в розетке', created_by: {id: 1, first_name: 'Имя', second_name: 'Фамилия'}, created_at: '21.05.2013', status: 'В работе', responsible_manager: {id: 2, first_name: 'Имя', second_name: 'Фамилия'}, responsible_engineer: {id: 3, first_name: 'Имя', second_name: 'Фамилия'}, closed_at: '25.05.2013'},
      ],
    }
  }
}
</script>