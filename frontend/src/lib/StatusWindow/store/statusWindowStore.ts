import { defineStore } from "pinia";
import { StatusTypes, StatusCodes, type IStatusWindow } from "../statusWindowAPI";

export const useStatusWindowStore = defineStore('statusWindow', {
  state() {
    return{
      statusWindowList: [] as IStatusWindow[],
      statusWindowID: 0,
    }
  },
  actions: {
    showStatusWindow(type: StatusTypes, status: StatusCodes, text: string, time: number, closable: boolean, showAnimation: boolean, showTimeBar: boolean, headerText: string): number{
      const newStatusWindowID: number = this.statusWindowID++;

      const newStatusWindow: IStatusWindow = {
        id: newStatusWindowID, 
        type: type, 
        status: status, 
        text: text, 
        time: time, 
        closable: closable, 
        showAnimation: showAnimation, 
        showTimeBar: showTimeBar, 
        headerText: headerText
      };

      this.statusWindowList.push(newStatusWindow);

      if(time > 0){
        setTimeout(() => {
          this.deleteStatusWindow(newStatusWindowID);
        }, time);
      }
      return newStatusWindowID;
    },

    updateStatusWindowText(id: number, text: string): boolean{
      const index = this.statusWindowList.findIndex((item:IStatusWindow) => item.id === id);

      if(index !== -1){
        this.statusWindowList[index].text = text;
        return true;
      }
      return false;
    },

    deleteStatusWindow(id: number): boolean{
      const index = this.statusWindowList.findIndex((item:IStatusWindow) => item.id === id);
      if (index !== -1) {
        this.statusWindowList.splice(index, 1);
        return true;
      }
      return false;
    },

    deleteAllStatusWindows(): boolean{
      this.statusWindowList = [] as IStatusWindow[];
      return true;
    }
  },
  getters: {
    getAllStatusWindows(): IStatusWindow[]{
      return this.statusWindowList;
    }
  }
});