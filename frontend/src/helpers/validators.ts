import type { IValidator } from "./constants";
import { FCapitalize } from "./functions";

//validate user Login
export function ValidUserLogin(value: string): IValidator<string>{
  // if(value.match(/^[a-zA-Z0-9-]+$/) === null){
  //   return {value: value, error: 'Некорректные символы в логине!'};
  // }
  // if(value.match(/[a-zA-Z]+/) === null){
  //   return {value: value, error: 'Логин должен соджержать латинские буквы'};
  // }
  if(value.length < 4){
    return {value: value, error: 'Логин слишком короткий!'};
  }
  if(value.length > 60){
    return {value: value, error: 'Логин слишком длинный!'};
  }
  return {value: value, error: ''};
}

//validate user Password
export function ValidUserPassword(value: string): IValidator<string>{
  // if(value.match(/^[a-zA-Z0-9-]+$/) === null){
  //   return {value: value, error: 'Некорректные символы в пароле!'};
  // }
  // if(value.match(/[a-zA-Z]+/) === null){
  //   return {value: value, error: 'Пароль должен соджержать латинские буквы'};
  // }
  // if(value.match(/[a-z]+/) === null){
  //   return {value: value, error: 'Пароль должен сожержать строчную букву!'};
  // }
  // if(value.match(/[A-Z]+/) === null){
  //   return {value: value, error: 'Пароль должен сожержать прописную букву!'};
  // }
  // if(value.match(/[0-9]+/) === null){
  //   return {value: value, error: 'Пароль должен содержать число!'};
  // }
  // if(value.length < 8){
  if(value.length < 4){
    return {value: value, error: 'Пароль слишком короткий!'};
  }
  if(value.length > 50){
    return {value: value, error: 'Пароль слишком длинный!'};
  }
  return {value: value, error: ''};
}


//validate user FirstName
export function ValidUserFirstName(value: string): IValidator<string>{
  if(value.match(/^[а-яА-Яa-zA-ZёЁ]+([-'`]{1}[а-яА-Яa-zA-ZёЁ]+)?$/) === null){
    return {value: FCapitalize(value.toLowerCase()), error: 'Некорректное имя!'};
  }
  if(value.length < 2){
    return {value: FCapitalize(value.toLowerCase()), error: 'Очень короткое имя!'};
  }
  if(value.length > 30){
    return {value: FCapitalize(value.toLowerCase()), error: 'Очень длинное имя!'};
  }
  return {value: FCapitalize(value.toLowerCase()), error: ''}
}

//validate user Lastname
export function ValidUserLastName(value: string): IValidator<string>{
  if(value.match(/^[а-яА-Яa-zA-ZёЁ]+([-'`]{1}[а-яА-Яa-zA-ZёЁ]+)?$/) === null){
    return {value: FCapitalize(value.toLowerCase()), error: 'Некорректная фамилия!'};
  }
  if(value.length < 2){
    return {value: FCapitalize(value.toLowerCase()), error: 'Очень короткая фамилия!'};
  }
  if(value.length > 30){
    return {value: FCapitalize(value.toLowerCase()), error: 'Очень длинная фамилия!'};
  }
  return {value: FCapitalize(value.toLowerCase()), error: ''}
}

//validate user Thirdname
export function ValidUserMiddleName(value: string): IValidator<string>{
  if(value.length === 0){
    return {value: FCapitalize(value.toLowerCase()), error: ''}
  }
  if(value.match(/^[а-яА-Яa-zA-ZёЁ]+([-'`]{1}[а-яА-Яa-zA-ZёЁ]+)?$/) === null){
    return {value: FCapitalize(value.toLowerCase()), error: 'Некорректное отчество!'};
  }
  if(value.length < 2){
    return {value: FCapitalize(value.toLowerCase()), error: 'Очень короткое отчество!'};
  }
  if(value.length > 30){
    return {value: FCapitalize(value.toLowerCase()), error: 'Очень длинное отчество!'};
    
  }
  return {value: FCapitalize(value.toLowerCase()), error: ''}
}