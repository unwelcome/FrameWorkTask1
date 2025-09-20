import { defineStore } from "pinia";
import { type IApproveWindow, type IApproveWindowAnswers } from "../approveWindowAPI";

export const useApproveWindowStore = defineStore('approveWindow', {
  state() {
    return{
      approveWindowQueue: [] as IApproveWindow[],
      approveWindowID: 0,
    }
  },
  actions: {
    createApproveWindow(title: string, text: string, answers: IApproveWindowAnswers[]): number{
      if(answers.length === 0) return -1;

      const currentID = this.approveWindowID++;
      this.approveWindowQueue.push({
        id: currentID,
        title: title,
        text: text,
        answers: answers,
      });

      return currentID;
    },
    updateApproveWindowTitle(id: number, title: string): boolean{
      for(let item of this.approveWindowQueue){
        if(item.id === id){
          item.title = title;
          return true;
        }
      }
      return false;
    },
    deleteApproveWindow(id: number): boolean{
      const index = this.approveWindowQueue.findIndex((item:IApproveWindow) => item.id === id);
      if (index !== -1) {
        this.approveWindowQueue.splice(index, 1);
        return true;
      }
      return false;
    },
    popFirst(): boolean{
      try{
        this.approveWindowQueue.splice(0, 1);
        return true;
      }catch(e) {
        return false;
      }
    },
    clearQueue(): boolean{
      if(this.approveWindowQueue.length > 1) this.approveWindowQueue = [ this.approveWindowQueue[0] ];
      return true;
    },
    deleteAllApproveWindow(): boolean{
      this.approveWindowQueue = [];
      return true;
    }
  },
  getters: {
    getApproveWindowQueue(): IApproveWindow[]{
      return this.approveWindowQueue;
    }
  }
});