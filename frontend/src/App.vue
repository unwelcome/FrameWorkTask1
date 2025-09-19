<template>
  <header>
    <div class=" bg-green-200 rounded-md p-4">
      <p>Hello world!</p>
    </div>

    <div class="m-10 flex flex-row gap-8 justify-center">
      <div class="flex flex-col gap-4 items-center">
        <TextButton :text="'bebra'" class="btn-main text-2xl px-6 py-1"/>
        <TextButton :text="'bebra'" class="btn-disabled"/>
        <TextButton :text="'bebra'" class="btn-change"/>
        <TextButton :text="'bebra'" class="btn-delete"/>
      </div>
      
      <div class="flex flex-col gap-4 items-center">
        <IconButton class="btn-main">
          <img class="h-6 w-6" src="./assets/icons/icon_settings.svg"/>
        </IconButton>
        <IconButton class="btn-disabled">
          <img class="h-6 w-6" src="./assets/icons/icon_settings.svg"/>
        </IconButton>
        <IconButton class="btn-change">
          <img class="h-6 w-6" src="./assets/icons/icon_settings.svg"/>
        </IconButton>
        <IconButton class="btn-delete">
          <img class="h-6 w-6" src="./assets/icons/icon_settings.svg"/>
        </IconButton>
      </div>

      <div class="flex flex-col gap-4 items-center">
        <IconTextButton class="btn-main" :text="'bebra'">
          <img class="h-6 w-6" src="./assets/icons/icon_settings.svg"/>
        </IconTextButton>
        <IconTextButton class="btn-disabled" :text="'bebra'">
          <img class="h-6 w-6" src="./assets/icons/icon_settings.svg"/>
        </IconTextButton>
        <IconTextButton class="btn-change text-2xl px-6 py-1" :text="'bebra'">
          <img class="h-6 w-6" src="./assets/icons/icon_settings.svg"/>
        </IconTextButton>
        <IconTextButton class="btn-delete" :text="'bebra'">
          <img class="h-6 w-6" src="./assets/icons/icon_settings.svg"/>
        </IconTextButton>
      </div>

      <div class="flex flex-col gap-4 items-center">
        <IconTextButton class="btn-main" :text="'bebra'" :reverse="true">
          <img class="h-6 w-6" src="./assets/icons/icon_settings.svg"/>
        </IconTextButton>
        <IconTextButton class="btn-disabled" :text="'bebra'" :reverse="true">
          <img class="h-6 w-6" src="./assets/icons/icon_settings.svg"/>
        </IconTextButton>
        <IconTextButton class="btn-change" :text="'bebra'" :reverse="true">
          <img class="h-6 w-6" src="./assets/icons/icon_settings.svg"/>
        </IconTextButton>
        <IconTextButton class="btn-delete" :text="'bebra'" :reverse="true">
          <img class="h-6 w-6" src="./assets/icons/icon_settings.svg"/>
        </IconTextButton>
      </div>
    </div>  
  </header>
  
  <div class="flex flex-col items-center">
    <div class="flex flex-col gap-4 w-[400px] h-[400px] bg-bg-1 rounded-xl p-4">
      <loginInput :placeholder="'Login'" :validator="(text: string) => { return text.length > 10 }" @change:input="(text: string) => loginText = text">
        <img class="w-7 h-7" src="./assets/icons/icon_user.svg"/>
      </loginInput>
      <p class="text-text-main text-base">Login: {{ loginText }}</p>

      <loginInput :placeholder="'Password'" :is-password="true" :validator="(text: string) => { return text.length > 10 }" @change:input="(text: string) => passwordText = text">
        <img class="w-7 h-7" src="./assets/icons/icon_lock.svg"/>
      </loginInput>
      <p class="text-text-main text-base">Password: {{ passwordText }}</p>
      
      <checkBox :placeholder="'Не выходить из аккаунта'" :default-checked="true"/>
      
      <searchBar 
        :placeholder="'Поиск по ID'" 
        :hint-enabled="true"
        :hint-array="testUsers"
        :hint-search-function="searchFunction"
        :hint-component="testUserItem"
        :hint-on-select-function="onSelectFunction"
        @change:input="(text: string) => searchText = text"
      />
      <p class="text-text-main text-base">Search: {{ searchText }}</p>
    </div>
  </div>

  <RouterView />
</template>
<script lang="ts">

import loginInput from './lib/loginInput.vue';
import checkBox from './lib/checkBox.vue';
import searchBar from './lib/searchBar.vue';
import testUserItem from './shared/testUserItem.vue';
import { markRaw } from 'vue';

interface User {
  id: number,
  fio: string,
}

export default {
  components: {
    loginInput,
    checkBox,
    searchBar,
  },
  data() {
    return {
      loginText: '',
      passwordText: '',
      searchText: '',

      testUsers: [
        {id: 1, fio: 'a'},
        {id: 2, fio: 'aa'},
        {id: 3, fio: 'ab'},
        {id: 4, fio: 'abc'},
        {id: 5, fio: 'abc'},
        {id: 6, fio: 'abcd'},
        {id: 7, fio: 'abce'},
        {id: 8, fio: 'ac'},
        {id: 9, fio: 'ad'},
        {id: 10, fio: 'b'},
        {id: 11, fio: 'c'},
        {id: 12, fio: 'd'},
      ],

      testUserItem: markRaw(testUserItem), // превращаем компонент в переменную
    }
  },
  methods: {
    searchFunction(query: string, array: User[]):User[] {
      const queryLower = query.toLowerCase();
      return array.filter(user => user.fio.toLowerCase().startsWith(queryLower));
    },
    onSelectFunction(data: User):string {
      console.log(`selected user with id: ${data.id} and fio: ${data.fio}`);
      return data.fio;
    }
  }
}
</script>
