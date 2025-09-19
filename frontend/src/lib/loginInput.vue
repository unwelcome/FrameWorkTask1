<template>
  <div class="flex flex-row items-stretch bg-bg-1 rounded-lg border-2 border-bg-input">
    <div class="flex flex-col shrink-0 justify-center px-2 py-1 border-r-2 border-bg-input cursor-pointer select-none" @click="iconClick">
      <slot></slot>
    </div>
    <div class="flex flex-col grow justify-center px-2">
      <input 
        :type="inputType" 
        class="text-text-main text-lg placeholder:text-text-placeholder-input placeholder:select-none"
        :class="{'text-text-wrong': !isInputCorrect}"
        v-model="inputText" 
        :placeholder="placeholder"
        ref="inputRef"
        />
    </div>
  </div>
</template>
<script lang="ts">
export default {
  emits: ['change:input'],
  props: {
    placeholder: {
      type: String,
      required: false,
      default: '' 
    },
    validator: {
      type: Function,
      required: false,
      default: (text: string) => { return true; }
    },
    isPassword: {
      type: Boolean,
      required: false,
      default: false,
    }
  },
  data() {
    return {
      inputText: '',

      inputType: this.isPassword ? 'password' : 'text',
    }
  },
  computed: {
    isInputCorrect() {
      if (this.inputText === '') return true;
      else return this.validator(this.inputText);
    }
  },
  methods: {
    iconClick() {
      if (this.isPassword) {
        if (this.inputType === 'password') this.inputType = 'text';
        else this.inputType = 'password';
      }
      
      (this.$refs.inputRef as HTMLInputElement).focus();
    }
  },
  watch: {
    inputText(newText: string) {
      this.$emit('change:input', newText);
    }
  }
}
</script>