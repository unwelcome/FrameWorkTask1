<template>
  <div class="flex flex-col justify-center items-center bg-bg-1 h-svh">
    <div class="flex flex-col gap-8 items-stretch py-4 px-8 w-[400px] h-[450px] rounded-xl bg-bg-2">
      <h1 class="text-2xl text-center cursor-default">Добро пожаловать!</h1>

      <div class="flex flex-col gap-4">
        <loginInput 
          :placeholder="'Логин'" 
          :validator="validUserLogin" 
          @change:input="(text: string) => loginInput = text"
        >
          <img class="w-6 h-6" src="../../assets/icons/icon_user.svg"/>
        </loginInput>

        <loginInput 
          :placeholder="'Пароль'" 
          :validator="validUserPassword" 
          :is-password="true" 
          @change:input="(text: string) => passwordInput = text"
        >
          <img class="w-6 h-6" src="../../assets/icons/icon_lock.svg"/>
        </loginInput>

        <checkBox 
          :placeholder="'Не выходить из аккаунта'" 
          :default-checked="false"
          @change:checked="(value: boolean) => saveToCookie = value"
        />
      </div>

      <div class="flex flex-col items-center gap-2 mt-auto">
        <textButton class="btn-main self-center py-2 px-24 text-xl" :text="'Вход'" @click="loginUser"/>

        <a class="text-text-link text-center cursor-pointer hover:underline" @click="">Забыли пароль?</a>
      </div>
    </div>
  </div>
</template>
<script lang="ts">
import loginInput from '@/lib/loginInput.vue';
import checkBox from '@/lib/checkBox.vue';
import { ValidUserLogin, ValidUserPassword } from '@/helpers/validators';
import { authApi } from '@/api/auth.api';

export default {
  components: {
    loginInput,
    checkBox,
  },
  data() {
    return {
      loginInput: '',
      passwordInput: '',
      saveToCookie: false,
    }
  },
  computed: {
    validUserLogin(){
      return (input: string) => ValidUserLogin(input).error === '';
    },
    validUserPassword(){
      return (input: string) => ValidUserPassword(input).error === '';
    },
  },
  methods: {
    async loginUser(){
      const userLogin = ValidUserLogin(this.loginInput);
      const userPassword = ValidUserPassword(this.passwordInput);
      
      if (userLogin.error === '' && userLogin.value !== '' && userPassword.error === '' && userPassword.value !== ''){
        console.log("send to backend: ", {login: userLogin.value, password: userPassword.value})
        const response = await authApi.login({login: userLogin.value, password: userPassword.value});

        this.$router.push({name: 'MainPage'});
      }else console.log('wrong inputs')
    }
  }
}
</script>