<template>
  <statusWindow />
  <approveWindow />

  <RouterView />
</template>
<script lang="ts">

import statusWindow from './lib/StatusWindow/components/statusWindow.vue';
import { useStatusWindowAPI } from './lib/StatusWindow/statusWindowAPI';

import approveWindow from './lib/ApproveWindow/approveWindow.vue';
import { useApproveWindowAPI } from './lib/ApproveWindow/approveWindowAPI';

export default {
  components: {
    statusWindow,
    approveWindow
  },
  data() {
    return { 
      // SW: useStatusWindowAPI(),
      // AW: useApproveWindowAPI(),
    }
  },
  mounted() {
    // this.SW.createStatusWindow({status: this.SW.getCodes.success, text: "Success text", time: -1, headerText: "Hello, world!"});
    // this.SW.createStatusWindow({status: this.SW.getCodes.error, text: "Error text", time: -1});
    // this.SW.createStatusWindow({status: this.SW.getCodes.loading, text: "Loading text", time: -1});
    // this.SW.createStatusWindow({status: this.SW.getCodes.info, text: "Info text", time: -1});

    // this.AW.createApproveWindow("Red or Blue", "Select red btn or blue btn", [
    //   {
    //     id: 1,
    //     type: this.AW.getApproveWindowAnswerType.bad,
    //     text: "Red",
    //   },
    //   {
    //     id: 2,
    //     type: this.AW.getApproveWindowAnswerType.good,
    //     text: "Blue",
    //   },
    // ])
  },
}
</script>
