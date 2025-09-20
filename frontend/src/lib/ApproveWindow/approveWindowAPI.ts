import { useApproveWindowStore } from "./store/approveWindowStore"

export enum ApproveWindowAnswerType{
  'good', 'normal', 'bad'
}

export interface IApproveWindowAnswers{
  id: number,
  type: ApproveWindowAnswerType,
  text: string,
  icon?: string,
}

export interface IApproveWindow{
  id: number,
  title: string,
  text: string,
  answers: IApproveWindowAnswers[],
}

interface IPromiseQueueItem {
  id: number;
  resolve: (value: number | PromiseLike<number>) => void;
  reject: (reason?: any) => void;
}

const promiseQueue: IPromiseQueueItem[] = [];

export function useApproveWindowAPI(){
  const approveWindowStore = useApproveWindowStore();

  const createApproveWindow = (title: string, text: string, answers: IApproveWindowAnswers[]):Promise<number> => {
    const appwroveWindowId = approveWindowStore.createApproveWindow(title, text, answers);

    let resolveFn: (value: number | PromiseLike<number>) => void;
    let rejectFn: (reason?: any) => void;

    const promiseToReturn = new Promise<number>((resolve, reject) => {
      resolveFn = resolve;
      rejectFn = reject;
    });

    promiseQueue.push({
      id: appwroveWindowId,
      resolve: resolveFn!,
      reject: rejectFn!,
    });

    return promiseToReturn;
  }

  const updateApproveWindowTitle = (id: number, title: string):boolean => {
    return approveWindowStore.updateApproveWindowTitle(id, title);
  }

  const deleteApproveWindow = (id: number):boolean => {
    for(let i = 0; i < promiseQueue.length; i++){
      if(promiseQueue[i].id === id){
        promiseQueue[i].reject();
        promiseQueue.splice(i, 1);
        break;
      }
    }

    return approveWindowStore.deleteApproveWindow(id);
  }

  const deleteAllApproveWindow = ():boolean => {
    for(let item of promiseQueue){
      item.reject();
    }
    promiseQueue.filter(item => false); // empty array

    return approveWindowStore.deleteAllApproveWindow();
  }

  const clearQueue = ():boolean => {
    for(let i = 1; i < promiseQueue.length; i++){
      promiseQueue[i].reject();
    }
    promiseQueue.splice(1, promiseQueue.length - 2);

    return approveWindowStore.clearQueue();
  }

  const resolveFirst = (returnID: number):boolean => {
    promiseQueue[0].resolve(returnID);
    promiseQueue.splice(0, 1);

    return approveWindowStore.popFirst();
  }

  const approveWindowAnswerTypesToString = (type: ApproveWindowAnswerType): string => {
    switch(type){
      case ApproveWindowAnswerType.good: return 'good';
      case ApproveWindowAnswerType.normal: return 'normal';
      case ApproveWindowAnswerType.bad: return 'bad';
    }
  }

  return {
    createApproveWindow,
    updateApproveWindowTitle,
    clearQueue,
    resolveFirst,
    deleteApproveWindow,
    deleteAllApproveWindow,
    approveWindowAnswerTypesToString,
    getApproveWindowAnswerType: ApproveWindowAnswerType,
    getApproveWindowQueue: approveWindowStore.getApproveWindowQueue,
  }
}