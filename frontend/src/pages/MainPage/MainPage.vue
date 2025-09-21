<template>
  <div class="flex flex-row w-full h-svh overflow-hidden bg-bg-1">
    <!--Sidebar-->
    <section class="flex flex-col w-[350px] shrink-0 gap-2 items-stretch">
      <!--sidebar links-->
      <div class="flex flex-col items-stretch gap-2 px-2 mt-20">
        <iconTextButton class="btn-sidebar justify-start text-xl py-1.5" :text="'Сотрудники'" @click="$router.push({name: 'EmployeesPage'})">
          <img class="w-8 h-8 select-none" src="../../assets/icons/icon_users.svg"/>
        </iconTextButton>

        <iconTextButton class="btn-sidebar justify-start text-xl py-1.5" :text="'Отделы'" @click="$router.push({name: 'DepartmentsPage'})">
          <img class="w-8 h-8 select-none" src="../../assets/icons/icon_departments.svg"/>
        </iconTextButton>

        <iconTextButton class="btn-sidebar justify-start text-xl py-1.5" :text="'Архив'" @click="$router.push({name: ''})">
          <img class="w-8 h-8 select-none" src="../../assets/icons/icon_archive.svg"/>
        </iconTextButton>

        <iconTextButton class="btn-sidebar justify-start text-xl py-1.5" :text="'Аналитика'" @click="$router.push({name: ''})">
          <img class="w-8 h-8 select-none" src="../../assets/icons/icon_analytics.svg"/>
        </iconTextButton>

        <iconTextButton class="btn-sidebar justify-start text-xl py-1.5" :text="'Исходящие заявления'" @click="$router.push({name: ''})">
          <img class="w-8 h-8 select-none" src="../../assets/icons/icon_outgoing_applications.svg"/>
        </iconTextButton>

        <iconTextButton class="btn-sidebar justify-start text-xl py-1.5" :text="'Входящие заявления'" @click="$router.push({name: ''})">
          <img class="w-8 h-8 select-none" src="../../assets/icons/icon_incoming_applications.svg"/>
        </iconTextButton>
      </div>
      <!--User profile-->
      <div class="flex flex-col items-stretch justify-center bg-bg-0 rounded-t-lg h-20 mt-auto px-4">
        <div class="flex flex-row gap-1 items-center">

          <userAvatar class="shrink-0 text-2xl cursor-pointer" :id="1" :first-name="'Имя'" :last-name="'Фамилия'" @click="$router.push({name: ''})"/>

          <p class="text-xl shrink-0 cursor-pointer select-none" @click="$router.push({name: ''})">Фамилия Имя</p>

          <iconButton class="ml-auto btn-main p-1" @click="$router.push({name: ''})">
            <img class="w-7 h-7" src="../../assets/icons/icon_settings.svg"/>
          </iconButton>
        </div>
      </div>
    </section>
    <!--Subpages-->
    <RouterView />
  </div>
</template>
<script lang="ts">

export default {

}
</script>