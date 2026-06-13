# Error Matrix — Auth Service

gRPC → HTTP mapping (из `gateway/internal/errors/grpcToFiber.go`):

| gRPC code | HTTP |
|---|---|
| InvalidArgument | 400 |
| NotFound | 404 |
| AlreadyExists | 409 |
| PermissionDenied | 403 |
| ResourceExhausted | 429 |
| Internal | 500 |
| Unauthenticated | 401 |

> **Все маршруты `/auth/*`** получают **401** от JWT middleware до вызова сервиса,
> если токен отсутствует или недействителен.

---

## Register · `POST /api/register`

| Сценарий | gRPC | HTTP | Сообщение | Примечание |
|---|---|---|---|---|
| Невалидный email | InvalidArgument | 400 | `invalid email` | |
| Невалидный пароль | InvalidArgument | 400 | `invalid password` | |
| Невалидное first_name | InvalidArgument | 400 | `invalid first name` | |
| Невалидное last_name | InvalidArgument | 400 | `invalid last name` | |
| Невалидное patronymic | InvalidArgument | 400 | `invalid patronymic` | |
| Email занят, аккаунт верифицирован | AlreadyExists | 409 | `email already registered` | → MQ: `registration-attempt.email` |
| Email занят, код верификации активен | AlreadyExists | 409 | `verification email already sent, please check your inbox` | |
| Ошибка GetUserByEmail при коллизии | Internal | 500 | `internal error` | |
| Ошибка БД | Internal | 500 | `internal error` | |
| **Успех** | — | **201** | `{user_uuid}` | → MQ: `verification.email` |

---

## Login · `POST /api/login`

| Сценарий | gRPC | HTTP | Сообщение                                            | Примечание |
|---|---|---|------------------------------------------------------|---|
| Невалидный email | InvalidArgument | 400 | `invalid email`                                      | |
| Невалидный пароль (формат) | InvalidArgument | 400 | `invalid password`                                   | |
| Email не найден | InvalidArgument | 400 | `wrong email or password`                            | timing dummy |
| Неверный пароль | InvalidArgument | 400 | `wrong email or password`                            | |
| Аккаунт удалён | PermissionDenied | 403 | `account is deleted, you have N hours to restore it` | N = hoursUntilAnonymization |
| Аккаунт не верифицирован | PermissionDenied | 403 | `account is not verified`                            | |
| Ошибка Save2FAData / SaveRefreshToken | … | 500 | `internal error`                                     | |
| **Успех (2FA включена)** | — | **200** | `{session_uuid}`                                     | → MQ: `2fa.email` |
| **Успех (2FA выключена)** | — | **200** | `{user_uuid, access_token, refresh_token}`           | |

---

## GetUser · `GET /auth/user/:user_uuid/info`

| Сценарий | gRPC | HTTP | Сообщение | Примечание |
|---|---|---|---|---|
| Нет JWT токена | Unauthenticated | 401 | middleware | |
| Невалидный UUID | InvalidArgument | 400 | `invalid user uuid` | |
| UUID не найден | NotFound | 404 | `user not found` | |
| **Успех (активный)** | — | **200** | `{user data}` | |
| **Успех (удалённый)** | — | **200** | `{user data, deleted_at≠""}` | возвращает 200, не 404 |
| **Успех (анонимизированный)** | — | **200** | `{user_uuid, остальные поля=""}` | пустые поля без признака |

---

## ChangePassword · `PATCH /auth/user/password`

| Сценарий                               | gRPC | HTTP | Сообщение | Примечание |
|----------------------------------------|---|---|---|---|
| Нет JWT токена                         | Unauthenticated | 401 | middleware | |
| Невалидный UUID                        | InvalidArgument | 400 | `invalid user uuid` | |
| Невалидный old_password (формат)       | InvalidArgument | 400 | `invalid old password` | |
| Невалидный password (формат)           | InvalidArgument | 400 | `invalid new password` | |
| Пользователь не найден                 | NotFound | 404 | `user not found` | |
| Неверный старый пароль                 | InvalidArgument | 400 | `wrong old password` | |
| Ошибка хеширования пароля              | Internal | 500 | `internal error` | |
| Ошибка UpdateUserPassword              | … | 404/500 | propagated | |
| Ошибка RevokeAllRefreshTokens (не 404) | … | 500 | `internal error` | |
| **Успех**                              | — | **200** | `{}` | → MQ: `password-changed.email`; токены отозваны |

---

## UpdateUserBio · `PATCH /auth/user/bio`

| Сценарий | gRPC | HTTP | Сообщение | Примечание |
|---|---|---|---|---|
| Нет JWT токена | Unauthenticated | 401 | middleware | |
| Невалидный UUID | InvalidArgument | 400 | `invalid user uuid` | |
| Невалидное first_name | InvalidArgument | 400 | `invalid first name` | |
| Невалидное last_name | InvalidArgument | 400 | `invalid last name` | |
| Невалидное patronymic | InvalidArgument | 400 | `invalid patronymic` | |
| Description > 500 символов | InvalidArgument | 400 | `invalid description` | |
| Пользователь не найден | NotFound | 404 | `user not found` | |
| **Успех** | — | **200** | `{}` | |

---

## DeleteUser · `DELETE /auth/user/account`

| Сценарий | gRPC | HTTP | Сообщение | Примечание |
|---|---|---|---|---|
| Нет JWT токена | Unauthenticated | 401 | middleware | |
| Невалидный initiator UUID | InvalidArgument | 400 | `invalid initiator uuid` | |
| Невалидный target UUID | InvalidArgument | 400 | `invalid target uuid` | |
| initiator ≠ target | PermissionDenied | 403 | `not enough rights` | |
| UUID не найден / уже удалён | NotFound | 404 | `user not found` | DB: `WHERE deleted_at IS NULL` |
| Ошибка RevokeAllRefreshTokens (не 404) | … | 500 | propagated | |
| **Успех** | — | **200** | `{}` | токены отозваны |

---

## RefreshToken · `POST /api/refresh`

| Сценарий | gRPC | HTTP | Сообщение | Примечание |
|---|---|---|---|---|
| Невалидный / просроченный JWT | InvalidArgument | 400 | текст из ParseToken | |
| Тип токена не refresh | InvalidArgument | 400 | `wrong token type` | |
| Токен не найден в Redis | NotFound | 404 | propagated | |
| UUID пользователя не найден | NotFound | 404 | `user not found` | |
| Ошибка создания токенов | Internal | 500 | `internal error` | |
| Ошибка ротации в Redis | … | 500 | propagated | |
| **Успех** | — | **200** | `{access_token, refresh_token}` | старый токен удалён |

---

## GetAllActiveTokens · `GET /auth/user/tokens`

| Сценарий | gRPC | HTTP | Сообщение | Примечание |
|---|---|---|---|---|
| Нет JWT токена | Unauthenticated | 401 | middleware | |
| Невалидный UUID | InvalidArgument | 400 | `invalid user uuid` | |
| Ошибка Redis | … | 500 | propagated | |
| **Успех** | — | **200** | `{tokens: [...]}` | пустой массив если токенов нет |

---

## RevokeToken · `DELETE /auth/user/revoke/token`

| Сценарий | gRPC | HTTP | Сообщение | Примечание |
|---|---|---|---|---|
| Нет JWT токена | Unauthenticated | 401 | middleware | |
| Невалидный UUID | InvalidArgument | 400 | `invalid user uuid` | |
| Пустой token_hash | InvalidArgument | 400 | `token hash missed` | |
| Токен не найден | NotFound | 404 | propagated | |
| **Успех** | — | **200** | `{}` | |

---

## RevokeAllTokens · `DELETE /auth/user/revoke/all`

| Сценарий | gRPC | HTTP | Сообщение | Примечание |
|---|---|---|---|---|
| Нет JWT токена | Unauthenticated | 401 | middleware | |
| Невалидный UUID | InvalidArgument | 400 | `invalid user uuid` | |
| Нет активных токенов | NotFound | **404** | propagated | |
| Ошибка Redis | … | 500 | propagated | |
| **Успех** | — | **200** | `{}` | |

---

## VerifyAccount · `POST /api/user/verify`

| Сценарий | gRPC | HTTP | Сообщение | Примечание |
|---|---|---|---|---|
| Невалидный email | InvalidArgument | 400 | `invalid email` | |
| Невалидный формат кода | InvalidArgument | 400 | `invalid verification code format` | |
| Email не найден | InvalidArgument | 400 | `invalid email or code` | timing dummy |
| Аккаунт уже верифицирован | AlreadyExists | 409 | `account already verified` | |
| Код не найден / истёк | InvalidArgument | 400 | `invalid or expired verification code` | |
| Ошибка счётчика попыток | Internal | 500 | `internal error` | |
| Превышен лимит попыток (>5) | ResourceExhausted | 429 | `too many attempts, please request a new verification code` | код удаляется |
| Неверный код | InvalidArgument | 400 | `invalid or expired verification code` | |
| Ошибка SetUserVerified | … | 500 | propagated | |
| **Успех** | — | **200** | `{}` | |

---

## ResendVerificationCode · `POST /api/user/verify/resend`

| Сценарий | gRPC | HTTP | Сообщение | Примечание |
|---|---|---|---|---|
| Невалидный email | InvalidArgument | 400 | `invalid email` | |
| Email не найден | — | **200** | `{}` (тихо) | не раскрывает, есть ли email |
| Аккаунт уже верифицирован | AlreadyExists | 409 | `account already verified` | |
| Ошибка SaveVerificationCode | … | 500 | propagated | |
| **Успех** | — | **200** | `{}` | → MQ: `verification.email`; сбрасывает счётчик попыток |

---

## ForgotPassword · `POST /api/forgot-password`

| Сценарий | gRPC | HTTP | Сообщение | Примечание |
|---|---|---|---|---|
| Невалидный email | InvalidArgument | 400 | `invalid email` | |
| Email не найден | — | **200** | `{}` (тихо) | |
| Аккаунт не верифицирован | — | **200** | `{}` (тихо) | не раскрывает состояние |
| Ошибка SaveRecoveryCode | … | 500 | propagated | |
| **Успех** | — | **200** | `{}` | → MQ: `recovery.email` |

---

## ResetPassword · `POST /api/reset-password`

| Сценарий | gRPC | HTTP | Сообщение | Примечание |
|---|---|---|---|---|
| Невалидный email | InvalidArgument | 400 | `invalid email` | |
| Невалидный формат кода | InvalidArgument | 400 | `invalid code format` | |
| Невалидный пароль | InvalidArgument | 400 | `invalid password` | |
| Email не найден | InvalidArgument | 400 | `invalid email or code` | timing dummy |
| Аккаунт не верифицирован | InvalidArgument | 400 | `invalid email or code` | timing dummy |
| Код не найден / истёк | InvalidArgument | 400 | `invalid or expired code` | |
| Ошибка счётчика попыток | Internal | 500 | `internal error` | |
| Превышен лимит попыток (>5) | ResourceExhausted | 429 | `too many attempts, please request a new recovery code` | код удаляется |
| Неверный код | InvalidArgument | 400 | `invalid or expired code` | |
| Ошибка хеширования | Internal | 500 | `internal error` | |
| Ошибка UpdateUserPassword | … | 404/500 | propagated | |
| Ошибка RevokeAllRefreshTokens (не 404) | … | 500 | propagated | |
| **Успех** | — | **200** | `{}` | токены отозваны |

---

## Verify2FA · `POST /api/verify-2fa`

| Сценарий | gRPC | HTTP | Сообщение | Примечание |
|---|---|---|---|---|
| Невалидный session_uuid | InvalidArgument | 400 | `invalid session uuid` | |
| Невалидный формат кода | InvalidArgument | 400 | `invalid code format` | |
| Сессия не найдена в Redis | NotFound | 404 | propagated | |
| Ошибка счётчика попыток | … | 500 | propagated | |
| Превышен лимит попыток (>5) | ResourceExhausted | 429 | `too many attempts, please try login again` | сессия удаляется |
| Неверный код | PermissionDenied | **403** | `invalid or expired code` | |
| Ошибка создания токенов | Internal | 500 | `internal error` | |
| Ошибка SaveRefreshToken | … | 500 | propagated | |
| **Успех** | — | **200** | `{user_uuid, access_token, refresh_token}` | |

---

## UpdateUser2FA · `PATCH /auth/user/2fa`

| Сценарий | gRPC | HTTP | Сообщение | Примечание |
|---|---|---|---|---|
| Нет JWT токена | Unauthenticated | 401 | middleware | |
| Невалидный UUID | InvalidArgument | 400 | `invalid user uuid` | |
| Пользователь не найден | NotFound | 404 | `user not found` | |
| **Успех** | — | **200** | `{}` | |

---

## RestoreAccount · `POST /api/restore-account`

| Сценарий | gRPC | HTTP | Сообщение | Примечание |
|---|---|---|---|---|
| Невалидный email | InvalidArgument | 400 | `invalid email` | |
| Невалидный пароль (формат) | InvalidArgument | 400 | `invalid password` | |
| Email не найден | InvalidArgument | 400 | `wrong email or password` | timing dummy |
| Неверный пароль | InvalidArgument | 400 | `wrong email or password` | |
| Аккаунт не удалён | InvalidArgument | 400 | `account is not deleted` | |
| Период восстановления истёк | PermissionDenied | 403 | `restoration period has expired` | |
| Ошибка RestoreUser | … | 500 | propagated | |
| **Успех** | — | **200** | `{}` | нужен отдельный Login |
