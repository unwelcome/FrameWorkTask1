<template>
  <div class="grid-container h-11 shrink-0 text-base">
      <!--SID-->
      <p class="text-center">{{ data.id }}</p>

      <!--Title-->
      <p class="">{{ data.title }}</p>

      <!--Created by-->
      <p v-if="typeof data.created_by === 'string'" class="text-center">Создал</p>
      <div v-else class="flex flex-col items-center">
        <userAvatar 
          class="text-base !p-1.5 cursor-pointer" 
          :id="data.created_by.id" 
          :first-name="data.created_by.first_name" 
          :second-name="data.created_by.second_name" 
          @click="redirectToUserProfile(data.created_by.id)"
        />
      </div>

      <!--Created at-->
      <p class="leading-4 text-center">{{ data.created_at }}</p>

      <!--Status-->
      <p class="text-center">{{ data.status }}</p>

      <!--Reponsible manager-->
      <p v-if="typeof data.responsible_manager === 'string'" class="leading-4 text-center">Ответств. менеджер</p>
      <div v-else class="flex flex-col items-center">
        <userAvatar 
          class="text-base !p-1.5 cursor-pointer" 
          :id="data.responsible_manager.id" 
          :first-name="data.responsible_manager.first_name" 
          :second-name="data.responsible_manager.second_name"
          @click="redirectToUserProfile(data.responsible_manager.id)"
          />
      </div>

      <!--Reponsible engineer-->
      <p v-if="typeof data.responsible_engineer === 'string'" class="leading-4 text-center">Ответств. инженер</p>
      <div v-else class="flex flex-col items-center">
        <userAvatar 
          class="text-base !p-1.5 cursor-pointer" 
          :id="data.responsible_engineer.id" 
          :first-name="data.responsible_engineer.first_name" 
          :second-name="data.responsible_engineer.second_name"
          @click="redirectToUserProfile(data.responsible_engineer.id)"
          />
      </div>

      <!--Closed at-->
      <p class="leading-4 text-center">{{ data.closed_at }}</p>
  </div>
</template>
<script lang="ts">
import type { PropType } from 'vue';

export default {
  props: {
    data: {
      type: Object as PropType<{
        id: number | string, 
        title: string, 
        created_by: string | {
          id: number, 
          first_name: string, 
          second_name: string
        }, 
        created_at: string, 
        status: string, 
        responsible_manager: string | {
          id: number, 
          first_name: string, 
          second_name: string
        }, 
        responsible_engineer: string | {
          id: number, 
          first_name: string, 
          second_name: string
        }, 
        closed_at: string
      }>,
      required: true,
    },
    redirectToProfile: {
      type: Boolean, 
      require: false,
      default: true,
    }
  },
  methods: {
    redirectToUserProfile(userID: number){
      if (this.redirectToProfile) this.$router.push({name: 'ProfilePage', params: {id: userID}});
    }
  }
}
</script>
<style lang="css" scoped>
  .grid-container {
    display: grid;
    grid-template-columns: 50px 1fr 80px 100px 100px 100px 100px 100px;
    align-items: center;
  }
</style>