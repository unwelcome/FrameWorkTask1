import { useStatusWindowStore } from "./store/statusWindowStore";

//special types
export enum StatusCodes {
    'error', 'info', 'loading', 'success'
};
export enum StatusTypes {
    'classic', 'border', 'detail'
};
export interface IStatusWindow{
    id: number,
    type: StatusTypes,
    status: StatusCodes,
    text: string,
    time: number,
    closable: boolean,
    showAnimation: boolean,
    showTimeBar: boolean,
    headerText: string,
};
export interface IStatusWindowCreate{
    type?: StatusTypes,
    status: StatusCodes,
    text: string,
    time?: number,
    closable?: boolean,
    showAnimation?: boolean,
    showTimeBar?: boolean,
    headerText?: string,
}
//default status window config
const StatusWindowType = StatusTypes.border;
const StatusWindowTime = 3000;
const StatusWindowClosable = true;
const ShowAnimation = true;
const ShowTimeBar = true;
//API
export function useStatusWindowAPI() {
    const statusWindowStore = useStatusWindowStore();

    const createStatusWindow = (obj: IStatusWindowCreate): number => { // type?, status, text, time?, closable?, showAnimation?, showTimeBar?, headerText?
        const type = (obj.type === undefined) ? StatusWindowType : obj.type;
        const time = (obj.time === undefined) ? StatusWindowTime : obj.time;
        const closable = (obj.closable === undefined) ? StatusWindowClosable : obj.closable;
        const showAnimation = (obj.showAnimation === undefined) ? ShowAnimation : obj.showAnimation;
        const showTimeBar = (obj.showTimeBar === undefined) ? ShowTimeBar : obj.showTimeBar;
        const headerText = (obj.headerText === undefined) ? statusCodesToString(obj.status) : obj.headerText;

        return statusWindowStore.showStatusWindow(type, obj.status, obj.text, time, closable, showAnimation, showTimeBar, headerText);
    };
    
    const updateStatusWindowText = (id: number, text: string): boolean => {
        return statusWindowStore.updateStatusWindowText(id, text);
    };

    const deleteStatusWindow = (id: number): boolean => {
        return statusWindowStore.deleteStatusWindow(id);
    };
    
    const deleteAllStatusWindows = (): boolean => {
        return statusWindowStore.deleteAllStatusWindows();
    };

    const statusCodesToString = (status: StatusCodes): string => {
        switch(status){
          case StatusCodes.error: return 'error';
          case StatusCodes.loading: return 'loading';
          case StatusCodes.success: return 'success';
          case StatusCodes.info: return 'info';
        }
    }
    
    return {
        createStatusWindow,
        updateStatusWindowText,
        deleteStatusWindow,
        deleteAllStatusWindows,
        statusCodesToString,
        getTypes: StatusTypes,
        getCodes: StatusCodes,
        getTime: StatusWindowTime,
        getAllStatusWindows: statusWindowStore.getAllStatusWindows,
    };
}



