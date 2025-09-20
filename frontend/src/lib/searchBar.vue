<template>
  <div class="relative flex flex-row gap-2 items-center h-10 px-3 py-1 bg-bg-input rounded-[10px] cursor-text" @click="setFocus">
    
    <input 
      type="text" 
      class="grow text-base placeholder:text-text-placeholder placeholder:text-lg placeholder:select-none" 
      :placeholder="placeholder" 
      v-model="inputText"
      ref="inputRef"
      />
    
    <div class="cursor-pointer select-none" @click="setFocus">
      <img src="../assets/icons/icon_search.svg"/>
    </div>

    <div v-if="showHint" class="absolute bg-bg-input top-11 left-0 w-full rounded-[10px] cursor-default p-2">
      <div v-if="filteredHintArray.length > 0" class="flex flex-col items-stretch gap-2">
        
        <div v-for="item, index in filteredHintArray" :key="index">
          <component 
            :is="hintComponent"
            :data="item"
            @select="onItemSelect"
          />
        </div>

      </div>
      <div v-else class="text-base">
        <p>Ничего не найдено</p>
      </div>
    </div>
  </div>
</template>
<script lang="ts">
/**
 * Компонент поиска с списком подсказок
 * 
 * props:
 * placeholder - плейсхолдер в окошке поиска
 * 
 * hintEnabled - включает список с подсказками
 * 
 * hintArray - исходный список для поиска
 * 
 * hintSearchFunction - функция для поиска, входные параметры (query - ввод пользователя, array - массив hintArray),
 * возвращает массив, удовлетворяющий поиску
 * 
 * hintSelectEnabled - включает выбор элемента из списка с подсказками
 * 
 * hintOnSelectFunction - функция, срабатывающая при выборе пользователем элемента из списка подсказок,
 * входные параметры (data - объект из массива hintArray, на котором сработал click)
 * 
 * hintMaxArrayItems - макс. число объектов из массива hintArray, которое отображается в списке
 * 
 * hintComponent - компонент, отрисовывающий компоненты из массива hintArray на экране
 * чтобы передать компонент в hintComponent нужно в родительском компоненте подключить компонент-отрисовщик
 * import renderComponent from "..."
 * import { markRaw } from 'vue';
 * не прописывать его в components: {X}
 * а указать его в data:
 * data() {
 *  return {
 *    renderComponent: markRaw(renderComponent) // делает его нереактивным
 *  }
 * } 
 * 
 */


import type { PropType } from 'vue';

export default {
  emits: ['change:input'],
  props: {
    placeholder: {
      type: String,
      required: false,
      default: '',
    },
    hintEnabled: {
      type: Boolean,
      required: false,
      default: false,
    },
    hintArray: {
      type: Array as PropType<any[]>,
      required: false,
      default: [],
    },
    hintSearchFunction: {
      type: Function as PropType<(query: string, array: any[]) => any[]>,
      requred: false,
      default: (query: string, array: any[]) => { return []; }
    },
    hintSelectEnabled: {
      type: Boolean,
      required: false,
      default: true,
    },
    hintOnSelectFunction: {
      type: Function as PropType<(data: any) => string>,
      requred: false,
      default: (data: any) => { return ''; }
    },
    hintMaxArrayItems: {
      type: Number,
      required: false,
      default: 5,
    },
    hintComponent: {
      type: Object as PropType<any>,
      required: false,
      default: {}
    }
  },
  data() {
    return {
      inputText: '',
      showHint: false,
      isSelected: false,
      filteredHintArray: [] as any[],
    }
  },
  methods: {
    setFocus() {      
      (this.$refs.inputRef as HTMLInputElement).focus();
    },
    onItemSelect(itemBack: any[]){
      // если выключен выбор из списка подсказок - выходим
      if (!this.hintSelectEnabled) return;

      this.inputText = this.hintOnSelectFunction(itemBack);

      // убираем список с подсказками
      this.showHint = false;
      // устанавливаем флаг чтобы при срабатывании watch подсказки снова не появились в случае autocomplete
      this.isSelected = true;
    }
  },
  watch: {
    inputText(newText: string) {
      this.$emit('change:input', newText);
      
      if (this.hintEnabled && newText !== '') this.showHint = true;
      else this.showHint = false; 

      // если пользователь выбрал элемент, то убираем повторно список подсказок 
      // и убираем флажок isSelected, чтобы при следующем вводе он не мешал
      if (this.isSelected){
        this.showHint = false;
        this.isSelected = false;
      }

      if(this.hintEnabled) {
        this.filteredHintArray = this.hintSearchFunction(this.inputText, this.hintArray);

        // оставляем не более, чем hintMaxArrayItems элементов в массиве для отображения
        if (this.filteredHintArray.length > this.hintMaxArrayItems) this.filteredHintArray = this.filteredHintArray.slice(0, this.hintMaxArrayItems);
      }
    }
  }
}
</script>