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

Email-enumeration protection: все ветки (успех, email занят, email занят неверифицированным)
возвращают **201 {}** без раскрытия информации об аккаунте.

| Сценарий | gRPC | HTTP | Сообщение | Примечание |
|---|---|---|---|---|
| Невалидный email | InvalidArgument | 400 | `invalid email` | |
| Невалидный пароль | InvalidArgument | 400 | `invalid password` | |
| Невалидное first_name | InvalidArgument | 400 | `invalid first name` | |
| Невалидное last_name | InvalidArgument | 400 | `invalid last name` | |
| Невалидное patronymic | InvalidArgument | 400 | `invalid patronymic` | |
| Ошибка хеширования пароля | Internal | 500 | `internal error` | |
| Ошибка GetUserByEmail при коллизии | Internal | 500 | `internal error` | |
| Ошибка БД (не AlreadyExists) | Internal | 500 | `internal error` | |
| **Успех** | — | **201** | `{}` | → MQ: `verification.email` |
| **Email занят, верифицирован** | — | **201** | `{}` (тихо) | → MQ: `registration-attempt.email` |
| **Email занят, не верифицирован** | — | **201** | `{}` (тихо) | → MQ: `verification.email` (новый токен) |

---

## Login · `POST /api/login`

| Сценарий | gRPC | HTTP | Сообщение | Примечание |
|---|---|---|---|---|
| Невалидный email | InvalidArgument | 400 | `invalid email` | |
| Невалидный пароль (формат) | InvalidArgument | 400 | `invalid password` | |
| Email не найден | InvalidArgument | 400 | `wrong email or password` | timing dummy |
| Неверный пароль | InvalidArgument | 400 | `wrong email or password` | |
| Аккаунт удалён | PermissionDenied | 403 | `account is deleted[, you have N hours/minutes to restore it]` | N округляется вниз; < 1 мин — без счётчика |
| Аккаунт не верифицирован | PermissionDenied | 403 | `account is not verified` | |
| 2FA: cooldown между отправками | ResourceExhausted | 429 | `please wait before requesting a new 2FA code` | per-account, только после верного пароля |
| 2FA: суточный лимит (>5 писем) | ResourceExhausted | 429 | `daily 2FA email limit reached` | per-account |
| Ошибка Save2FAData / SaveSession | … | 500 | `internal error` | |
| **Успех (2FA включена)** | — | **200** | `{session_uuid}` | → MQ: `2fa.email` |
| **Успех (2FA выключена)** | — | **200** | `{user_uuid, access_token, refresh_token}` | → MQ: `login-notification.email` |

---

## GetUser · `GET /auth/user/{user_uuid}/info`

| Сценарий | gRPC | HTTP | Сообщение | Примечание |
|---|---|---|---|---|
| Нет JWT токена | Unauthenticated | 401 | middleware | |
| Невалидный UUID | InvalidArgument | 400 | `invalid user uuid` | |
| Requester ≠ target и не коллеги | — | 403 | `access denied` | gateway: CheckColleagues → company service |
| Ошибка CheckColleagues | … | 500 | propagated | от company service |
| UUID не найден | NotFound | 404 | `user not found` | |
| **Успех (активный)** | — | **200** | `{user data}` | |
| **Успех (удалённый)** | — | **200** | `{user data, deleted_at≠""}` | возвращает 200, не 404 |
| **Успех (анонимизированный)** | — | **200** | `{user_uuid, остальные поля=""}` | пустые поля без признака |

---

## ChangePassword · `PATCH /auth/user/password`

| Сценарий | gRPC | HTTP | Сообщение | Примечание |
|---|---|---|---|---|
| Нет JWT токена | Unauthenticated | 401 | middleware | |
| Невалидный UUID | InvalidArgument | 400 | `invalid user uuid` | |
| Невалидный old_password (формат) | InvalidArgument | 400 | `invalid old password` | |
| Невалидный password (формат) | InvalidArgument | 400 | `invalid new password` | |
| Пользователь не найден | NotFound | 404 | `user not found` | |
| Аккаунт удалён | PermissionDenied | 403 | `account is deleted...` | |
| Неверный старый пароль | InvalidArgument | 400 | `wrong old password` | |
| Ошибка хеширования пароля | Internal | 500 | `internal error` | |
| Ошибка UpdateUserPassword | … | 404/500 | propagated | |
| Ошибка RevokeAllSessions (не 404) | … | 500 | propagated | |
| **Успех** | — | **200** | `{}` | → MQ: `password-changed.email`; все сессии отозваны |

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
| Аккаунт удалён | PermissionDenied | 403 | `account is deleted...` | |
| **Успех** | — | **200** | `{}` | |

---

## DeleteUser · `DELETE /auth/user/account`

| Сценарий | gRPC | HTTP | Сообщение | Примечание |
|---|---|---|---|---|
| Нет JWT токена | Unauthenticated | 401 | middleware | |
| Невалидный UUID | InvalidArgument | 400 | `invalid user uuid` | gateway берёт UUID из JWT |
| UUID не найден / уже удалён | NotFound | 404 | `user not found` | DB: `WHERE deleted_at IS NULL` |
| Ошибка RevokeAllSessions (не 404) | … | 500 | propagated | |
| **Успех** | — | **200** | `{}` | все сессии отозваны |

---

## GetAllActiveSessions · `GET /auth/user/sessions`

| Сценарий | gRPC | HTTP | Сообщение | Примечание |
|---|---|---|---|---|
| Нет JWT токена | Unauthenticated | 401 | middleware | |
| Невалидный UUID | InvalidArgument | 400 | `invalid user uuid` | |
| Ошибка Redis | … | 500 | propagated | |
| **Успех** | — | **200** | `{tokens: [...]}` | пустой массив если нет сессий; истёкшие очищаются |

---

## RefreshToken · `POST /api/refresh`

| Сценарий | gRPC | HTTP | Сообщение | Примечание |
|---|---|---|---|---|
| Невалидный / просроченный JWT | InvalidArgument | 400 | текст из ParseToken | |
| Тип токена не refresh | InvalidArgument | 400 | `wrong token type` | |
| Сессия не найдена в Redis | NotFound | 404 | `session not found` | |
| UUID пользователя не найден | NotFound | 404 | propagated | |
| Аккаунт удалён | PermissionDenied | 403 | `account deleted` | |
| Аккаунт не верифицирован | PermissionDenied | 403 | `account not verified` | |
| Ошибка создания токенов | Internal | 500 | `internal error` | |
| Ошибка ротации в Redis | … | 500 | propagated | Watch/TxPipelined |
| **Успех** | — | **200** | `{access_token, refresh_token}` | старая сессия удалена атомарно |

---

## RevokeSession · `DELETE /auth/user/session`

| Сценарий | gRPC | HTTP | Сообщение | Примечание |
|---|---|---|---|---|
| Нет JWT токена | Unauthenticated | 401 | middleware | |
| Невалидный UUID | InvalidArgument | 400 | `invalid user uuid` | |
| Пустой token_hash | InvalidArgument | 400 | `token hash missed` | |
| Токен не найден / не принадлежит пользователю | NotFound | 404 | `session not found` | |
| **Успех** | — | **200** | `{}` | |

---

## RevokeAllSessions · `DELETE /auth/user/sessions`

| Сценарий | gRPC | HTTP | Сообщение | Примечание |
|---|---|---|---|---|
| Нет JWT токена | Unauthenticated | 401 | middleware | |
| Невалидный UUID | InvalidArgument | 400 | `invalid user uuid` | |
| Нет активных сессий | NotFound | **404** | propagated | сервис не игнорирует NotFound в этом случае |
| Ошибка Redis | … | 500 | propagated | |
| **Успех** | — | **200** | `{}` | |

---

## VerifyAccount · `POST /api/user/verify`

Принимает JWT magic-link токен из письма (одноразовый).

| Сценарий | gRPC | HTTP | Сообщение | Примечание |
|---|---|---|---|---|
| Невалидный / просроченный JWT | InvalidArgument | 400 | `invalid or expired verification token` | |
| Тип токена не verification | InvalidArgument | 400 | `invalid or expired verification token` | |
| Токен уже использован (в blacklist) | InvalidArgument | 400 | `invalid or expired verification token` | |
| Email из claims не найден | InvalidArgument | 400 | `invalid or expired verification token` | |
| Аккаунт уже верифицирован | AlreadyExists | 409 | `account already verified` | |
| Ошибка SetUserVerified | … | 500 | propagated | |
| **Успех** | — | **200** | `{}` | токен добавлен в blacklist (TTL = remaining lifetime) |

---

## ResendVerificationCode · `POST /api/user/verify/resend`

Не раскрывает, есть ли email или верифицирован ли аккаунт.

| Сценарий | gRPC | HTTP | Сообщение | Примечание |
|---|---|---|---|---|
| Невалидный email | InvalidArgument | 400 | `invalid email` | |
| Email не найден | — | **200** | `{}` (тихо) | не раскрывает, есть ли email |
| Аккаунт уже верифицирован | — | **200** | `{}` (тихо) | не раскрывает состояние |
| Ошибка создания JWT | Internal | 500 | `internal error` | |
| **Успех** | — | **200** | `{}` | → MQ: `verification.email` |

---

## ForgotPassword · `POST /api/forgot-password`

| Сценарий | gRPC | HTTP | Сообщение | Примечание |
|---|---|---|---|---|
| Невалидный email | InvalidArgument | 400 | `invalid email` | |
| Email не найден | — | **200** | `{}` (тихо) | |
| Аккаунт не верифицирован | — | **200** | `{}` (тихо) | не раскрывает состояние |
| Аккаунт удалён | — | **200** | `{}` (тихо) | не раскрывает состояние |
| Ошибка создания JWT | Internal | 500 | `internal error` | |
| **Успех** | — | **200** | `{}` | → MQ: `recovery.email` |

---

## ResetPassword · `POST /api/reset-password`

Принимает JWT reset-password токен из письма (одноразовый, TTL 15 мин).

| Сценарий | gRPC | HTTP | Сообщение | Примечание |
|---|---|---|---|---|
| Невалидный новый пароль | InvalidArgument | 400 | `invalid password` | |
| Невалидный / просроченный JWT | InvalidArgument | 400 | `invalid or expired reset token` | |
| Тип токена не reset_password | InvalidArgument | 400 | `invalid or expired reset token` | |
| Токен уже использован (в blacklist) | InvalidArgument | 400 | `invalid or expired reset token` | |
| Email из claims не найден | InvalidArgument | 400 | `invalid or expired reset token` | |
| Аккаунт не верифицирован или удалён | InvalidArgument | 400 | `invalid or expired reset token` | |
| Ошибка хеширования | Internal | 500 | `internal error` | |
| Ошибка UpdateUserPassword | … | 404/500 | propagated | |
| Ошибка RevokeAllSessions (не 404) | … | 500 | propagated | |
| **Успех** | — | **200** | `{}` | токен в blacklist; → MQ: `password-reset.email`; все сессии отозваны |

---

## Verify2FA · `POST /api/verify-2fa`

| Сценарий | gRPC | HTTP | Сообщение | Примечание |
|---|---|---|---|---|
| Невалидный session_uuid | InvalidArgument | 400 | `invalid session uuid` | |
| Невалидный формат кода | InvalidArgument | 400 | `invalid code format` | |
| Сессия не найдена в Redis | NotFound | 404 | propagated | |
| Ошибка счётчика попыток | … | 500 | propagated | |
| Превышен лимит попыток (>5) | ResourceExhausted | 429 | `too many attempts, please try login again` | сессия удаляется |
| Неверный код | InvalidArgument | **400** | `invalid or expired code` | |
| Ошибка создания токенов | Internal | 500 | `internal error` | |
| Ошибка SaveSession | … | 500 | propagated | |
| **Успех** | — | **200** | `{user_uuid, access_token, refresh_token}` | → MQ: `login-notification.email` |

---

## UpdateUser2FA · `PATCH /auth/user/2fa`

| Сценарий | gRPC | HTTP | Сообщение | Примечание |
|---|---|---|---|---|
| Нет JWT токена | Unauthenticated | 401 | middleware | |
| Невалидный UUID | InvalidArgument | 400 | `invalid user uuid` | |
| Пользователь не найден | NotFound | 404 | `user not found` | |
| Аккаунт удалён | PermissionDenied | 403 | `account is deleted...` | |
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
| Период восстановления истёк | PermissionDenied | 403 | `restoration period has expired` | cleanup ещё не случился, но nextCleanupTime ≤ now |
| Ошибка RestoreUser | … | 500 | propagated | |
| **Успех** | — | **200** | `{}` | нужен отдельный Login |
