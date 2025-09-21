<template>
  <toggle-group :title="title" :default-show-body="defaultShowBody">
    <div class="flex flex-col items-stretch max-h-[600px] ml-5">
      <!--Header-->
      <apllications-group-item class="bg-bg-2 mr-[26px]" :data="{id: 'SID', title: 'Название', created_by: 'Создал', created_at: 'Дата создания', status: 'Статус', responsible_manager: 'Ответств. менеджер', responsible_engineer: 'Ответств. инженер', closed_at: 'Дата закрытия'}"/>
      <!--List-->
      <div class="flex flex-col scrollable pr-5">
        <apllications-group-item 
          v-for="(application, index) in applicationsList" 
          :key="application.id" 
          :data="application"
          :class="{'bg-bg-4': index % 2 == 0, 'bg-bg-5': index % 2 != 0}"
          />
      </div>
    </div>
  </toggle-group>
</template>
<script lang="ts">
import type { PropType } from 'vue';
import toggleGroup from '@/features/toggleGroup.vue';
import apllicationsGroupItem from '../features/apllicationsGroupItem.vue';

export default {
  components: {
    toggleGroup,
    apllicationsGroupItem,
  },
  props: {
    title: {
      type: String,
      required: true,
    },
    applicationsList: {
      type: Array as PropType<{id: number, title: string}[]>,
      required: true,
    },
    defaultShowBody: {
      type: Boolean, 
      required: false,
      default: false,
    }
  },
  data(){
    return {
      showBody: this.defaultShowBody,

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