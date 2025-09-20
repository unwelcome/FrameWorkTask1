<template>
  <div class="flex flex-col grow items-stretch">
    <header class="flex flex-row h-20 items-center">
    </header>
    <main class="flex flex-col items-stretch bg-bg-2 grow rounded-tl-xl p-5">
      <div class="flex flex-col gap-6">
        <div class="flex flex-row gap-5 items-stretch">
          <div class="flex flex-col justify-center items-center">
            <userAvatar class="text-6xl p-6" :id="3" :first-name="'Имя'" :last-name="'Фамилия'"/>
          </div>
          <div class="flex flex-col justify-center">
            <h1 class="text-2xl mb-2 cursor-default">Фамилия Имя Отчетсво</h1>
            <h3 class="text-lg cursor-default">Должность: <span>инженер</span></h3>
            <h3 class="text-lg cursor-default">Отдел: <span>Отдел Энергетики</span></h3>
            <h3 class="text-lg cursor-default">Дата регистрации: <span>12.03.2012</span></h3>
          </div>
        </div>
        <div class="flex flex-col">
          <h2 class="text-lg mb-1 cursor-default">Статистика:</h2>
          <h3 class="text-base cursor-default">Создано заявлений: <span>1234</span></h3>
          <h3 class="text-base cursor-default">Выполнено заявлений: <span>1229</span></h3>
          <h3 class="text-base cursor-default">Провалено заявлений: <span>2</span></h3>
          <h3 class="text-base cursor-default">В работе заявлений: <span>8</span></h3>
        </div>
        <div class="flex flex-col">
          <h2 class="text-lg mb-1 cursor-default">Контактные данные:</h2>
          <h3 class="text-base cursor-default">Email: <span class="cursor-pointer" @click="copyToClipboard('user_email@mail.ru')">user_email@mail.ru</span></h3>
          <h3 class="text-base cursor-default">Тел.: <span class="cursor-pointer" @click="copyToClipboard('+7 (987) 654-32-10')">+7 (987) 654-32-10</span></h3>
        </div>
      </div>

    </main>
  </div>
</template>
<script lang="ts">
import { useStatusWindowAPI } from '@/lib/StatusWindow/statusWindowAPI';

export default {
  data(){
    return {
      SW: useStatusWindowAPI(),
    }
  },
  methods: {
    copyToClipboard(text: string) {
      navigator.clipboard.writeText(text);
      this.SW.createStatusWindow({status: this.SW.getCodes.success, text: 'Скопировано в буфер обмена', time: 1500});
    }
  }
}
</script>