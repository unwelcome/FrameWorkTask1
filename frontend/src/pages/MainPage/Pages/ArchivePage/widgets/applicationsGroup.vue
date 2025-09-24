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
      type: Array as PropType<{id: number,title: string,created_by: {id: number,first_name: string,second_name: string,},created_at: string,status: string,responsible_manager: {id: number,first_name: string,second_name: string,},responsible_engineer: {id: number, first_name: string, second_name: string}, closed_at: string}[]>,
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
    }
  }
}
</script>