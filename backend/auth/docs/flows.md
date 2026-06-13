# Flowcharts методов auth сервиса

Методы с нетривиальной логикой ветвлений. Линейные методы (UpdateUserBio,
GetAllActiveSessions, RevokeSession, UpdateUser2FA) не включены.

Все защищённые маршруты (`/auth/*`) неявно добавляют **401** от JWT middleware до вызова сервиса.

---

## Register

`POST /api/register`

Email-enumeration protection: все ветви (новый пользователь, email занят верифицированным,
email занят неверифицированным) возвращают одинаковый **201 {}**.

```mermaid
flowchart TD
    A([Start]) --> V1{validate\nemail / password\nfirst/last/patronymic}
    V1 -->|fail| E1[/"400 invalid ..."/]

    V1 -->|ok| HASH[Hash password Argon2id]
    HASH -->|error| E0[/"500 internal error"/]
    HASH -->|ok| DB1[CreateUser в PostgreSQL]

    DB1 -->|success| TK1[CreateVerificationToken JWT\n48h TTL]
    TK1 --> MQ1[/"→ MQ: verification.email\nfire & forget"/]
    MQ1 --> OK[/"201 {}"/]

    DB1 -->|AlreadyExists| DB2[GetUserByEmail]
    DB2 -->|error| E2[/"500 internal error"/]

    DB2 -->|ok| CHK{IsVerified?}
    CHK -->|true| MQ2[/"→ MQ: registration-attempt.email\nfire & forget"/]
    MQ2 --> OK2[/"201 {} (тихо)"/]

    CHK -->|false| TK2[CreateVerificationToken JWT]
    TK2 --> MQ3[/"→ MQ: verification.email\nfire & forget"/]
    MQ3 --> OK3[/"201 {} (тихо)"/]

    DB1 -->|other error| E6[/"500 internal error"/]
```

---

## Login

`POST /api/login`

```mermaid
flowchart TD
    A([Start]) --> V1{validate email\n+ password}
    V1 -->|fail| E1[/"400 invalid email / invalid password"/]

    V1 -->|ok| DB1[GetUserByEmail]
    DB1 -->|not found| DUMMY[timing dummy:\nVerify dummyHash]
    DUMMY --> E2[/"400 wrong email or password"/]

    DB1 -->|ok| PWD{Verify\npassword}
    PWD -->|fail| E3[/"400 wrong email or password"/]

    PWD -->|ok| DEL{deleted_at\n!= nil?}
    DEL -->|true| E4[/"403 account is deleted\nyou have N hours to restore it"/]

    DEL -->|false| VER{IsVerified?}
    VER -->|false| E5[/"403 account not verified"/]

    VER -->|true| TFA{Enabled2FA?}

    TFA -->|true| RL1[Acquire2FAEmailCooldown\nв Redis]
    RL1 -->|error| E_rl1[/"... propagated"/]
    RL1 -->|cooldown active| E6[/"429 please wait before\nrequesting a new 2FA code"/]
    RL1 -->|ok| RL2[Incr2FAEmailDailyCount\nв Redis]
    RL2 -->|error| E_rl2[/"... propagated"/]
    RL2 -->|count > 5| E7[/"429 daily 2FA email limit reached"/]
    RL2 -->|ok| S1[Save2FAData в Redis\nsessionUUID + code]
    S1 -->|error| E8[/"... propagated"/]
    S1 -->|ok| MQ1[/"→ MQ: 2fa.email\nfire & forget"/]
    MQ1 --> OK1[/"200 {session_uuid}"/]

    TFA -->|false| TK[CreateTokens JWT]
    TK -->|error| E9[/"500 internal error"/]
    TK -->|ok| RD1[SaveSession в Redis]
    RD1 -->|error| E10[/"... propagated"/]
    RD1 -->|ok| MQ2[/"→ MQ: login-notification.email\nfire & forget"/]
    MQ2 --> OK2[/"200 {user_uuid, access_token, refresh_token}"/]
```

---

## VerifyAccount

`POST /api/user/verify`

Использует JWT magic-link (одноразовый, TTL 48h). После подтверждения токен добавляется
в blacklist в Redis с TTL = оставшееся время жизни токена.

```mermaid
flowchart TD
    A([Start]) --> PT[ParseVerificationToken JWT]
    PT -->|invalid / expired| E1[/"400 invalid or expired verification token"/]

    PT -->|ok| TTP{TokenType\n== verification?}
    TTP -->|false| E2[/"400 invalid or expired verification token"/]

    TTP -->|true| BL1[IsVerificationTokenBlacklisted\nв Redis по claims.ID]
    BL1 -->|error| E3[/"... propagated"/]
    BL1 -->|blacklisted| E4[/"400 invalid or expired verification token"/]

    BL1 -->|ok| DB1[GetUserByEmail из PostgreSQL\nпо claims.Email]
    DB1 -->|not found| E5[/"400 invalid or expired verification token"/]

    DB1 -->|ok| VER{IsVerified?}
    VER -->|true| E6[/"409 account already verified"/]

    VER -->|false| DB2[SetUserVerified в PostgreSQL]
    DB2 -->|error| E7[/"... propagated"/]
    DB2 -->|ok| BL2["AddToVerificationTokenBlacklist\nTTL = remaining lifetime\nfire & forget"]
    BL2 --> OK[/"200 {}"/]
```

---

## ResendVerificationCode

`POST /api/user/verify/resend`

Не раскрывает, есть ли email в системе или верифицирован ли аккаунт.

```mermaid
flowchart TD
    A([Start]) --> V1{validate email}
    V1 -->|fail| E1[/"400 invalid email"/]

    V1 -->|ok| DB1[GetUserByEmail]
    DB1 -->|not found| LOG1[log.Warn]
    LOG1 --> OK1[/"200 {} (тихо)"/]

    DB1 -->|ok| VER{IsVerified?}
    VER -->|true| LOG2[log.Warn]
    LOG2 --> OK2[/"200 {} (тихо)"/]

    VER -->|false| TK[CreateVerificationToken JWT]
    TK -->|error| E2[/"500 internal error"/]
    TK -->|ok| MQ[/"→ MQ: verification.email\nfire & forget"/]
    MQ --> OK3[/"200 {}"/]
```

---

## ForgotPassword

`POST /api/forgot-password`

Не раскрывает состояние аккаунта.

```mermaid
flowchart TD
    A([Start]) --> V1{validate email}
    V1 -->|fail| E1[/"400 invalid email"/]

    V1 -->|ok| DB1[GetUserByEmail]
    DB1 -->|not found| LOG1[log.Warn]
    LOG1 --> OK1[/"200 {} (тихо)"/]

    DB1 -->|ok| CHK{IsVerified\n&& !deleted?}
    CHK -->|false| LOG2[log.Warn]
    LOG2 --> OK2[/"200 {} (тихо)"/]

    CHK -->|true| TK[CreateResetPasswordToken JWT\n15min TTL]
    TK -->|error| E2[/"500 internal error"/]
    TK -->|ok| MQ[/"→ MQ: recovery.email\nfire & forget"/]
    MQ --> OK3[/"200 {}"/]
```

---

## ResetPassword

`POST /api/reset-password`

Использует JWT из письма (одноразовый, TTL 15min). После сброса токен добавляется
в blacklist, затем отзываются все активные сессии.

```mermaid
flowchart TD
    A([Start]) --> V1{validate\nnew_password}
    V1 -->|fail| E1[/"400 invalid password"/]

    V1 -->|ok| PT[ParseResetPasswordToken JWT]
    PT -->|invalid / expired| E2[/"400 invalid or expired reset token"/]

    PT -->|ok| TTP{TokenType\n== reset_password?}
    TTP -->|false| E3[/"400 invalid or expired reset token"/]

    TTP -->|true| BL1[IsResetTokenBlacklisted\nв Redis по claims.ID]
    BL1 -->|error| E4[/"... propagated"/]
    BL1 -->|blacklisted| E5[/"400 invalid or expired reset token"/]

    BL1 -->|ok| DB1[GetUserByEmail из PostgreSQL\nпо claims.Email]
    DB1 -->|not found| E6[/"400 invalid or expired reset token"/]

    DB1 -->|ok| CHK{"IsVerified\n&& deleted_at == nil?"}
    CHK -->|false| E7[/"400 invalid or expired reset token"/]

    CHK -->|true| HASH[Hash new password Argon2id]
    HASH -->|error| E8[/"500 internal error"/]
    HASH -->|ok| DB2[UpdateUserPassword в PostgreSQL]
    DB2 -->|error| E9[/"... propagated"/]
    DB2 -->|ok| BL2["AddToResetTokenBlacklist\nTTL = remaining lifetime\nfire & forget"]
    BL2 --> MQ[/"→ MQ: password-reset.email\nfire & forget"/]
    MQ --> RVK[RevokeAllSessions в Redis\nNotFound игнорируется]
    RVK -->|error| E10[/"... propagated"/]
    RVK -->|ok| OK[/"200 {}"/]
```

---

## ChangePassword

`PATCH /auth/user/password`

```mermaid
flowchart TD
    A([Start]) --> V1{validate UUID\n+ old_password\n+ new_password}
    V1 -->|fail| E1[/"400 invalid ..."/]

    V1 -->|ok| DB1[GetUser из PostgreSQL]
    DB1 -->|not found| E2[/"404 propagated"/]

    DB1 -->|ok| DEL{deleted_at\n!= nil?}
    DEL -->|true| E3[/"403 account is deleted..."/]

    DEL -->|false| PWD{Verify\nold password}
    PWD -->|fail| E4[/"400 wrong old password"/]

    PWD -->|ok| HASH[Hash new password Argon2id]
    HASH -->|error| E5[/"500 internal error"/]
    HASH -->|ok| DB2[UpdateUserPassword в PostgreSQL]
    DB2 -->|error| E6[/"... propagated"/]
    DB2 -->|ok| MQ[/"→ MQ: password-changed.email\nfire & forget"/]
    MQ --> RVK[RevokeAllSessions в Redis\nNotFound игнорируется]
    RVK -->|error| E7[/"... propagated"/]
    RVK -->|ok| OK[/"200 {}"/]
```

---

## RefreshToken

`POST /api/refresh`

```mermaid
flowchart TD
    A([Start]) --> V1[ParseToken JWT]
    V1 -->|invalid / expired| E1[/"400 ... parse error message"/]

    V1 -->|ok| TTP{TokenType\n== refresh?}
    TTP -->|false| E2[/"400 wrong token type"/]

    TTP -->|true| RD1[CheckSessionExists\nв Redis]
    RD1 -->|not found| E3[/"404 session not found"/]

    RD1 -->|ok| DB1[GetUser из PostgreSQL]
    DB1 -->|not found| E4[/"404 propagated"/]

    DB1 -->|ok| DEL{deleted_at\n!= nil?}
    DEL -->|true| E5[/"403 account deleted"/]

    DEL -->|false| VER{IsVerified?}
    VER -->|false| E6[/"403 account not verified"/]

    VER -->|true| TK[CreateTokens JWT новая пара]
    TK -->|error| E7[/"500 internal error"/]
    TK -->|ok| RD2["RefreshToken в Redis\n(Watch + TxPipelined:\nудалить старый + сохранить новый)"]
    RD2 -->|error| E8[/"... propagated"/]
    RD2 -->|ok| OK[/"200 {access_token, refresh_token}"/]
```

---

## Verify2FA

`POST /api/verify-2fa`

```mermaid
flowchart TD
    A([Start]) --> V1{validate\nsession_uuid\n+ code format}
    V1 -->|fail| E1[/"400 invalid session uuid / invalid code format"/]

    V1 -->|ok| RD1[Get2FAData из Redis\nпо session_uuid]
    RD1 -->|not found| E2[/"404 propagated"/]

    RD1 -->|ok| INC[Incr2FAAttempts]
    INC -->|error| E3[/"... propagated"/]

    INC -->|ok| ATT{attempts\n> 5?}
    ATT -->|true| DEL1[Delete2FAData\nfire & forget]
    DEL1 --> E4[/"429 too many attempts\nplease try login again"/]

    ATT -->|false| CMP{code\n== stored?}
    CMP -->|false| E5[/"400 invalid or expired code"/]

    CMP -->|true| DEL2[Delete2FAData\nfire & forget]
    DEL2 --> TK[CreateTokens JWT]
    TK -->|error| E6[/"500 internal error"/]
    TK -->|ok| RD2[SaveSession в Redis]
    RD2 -->|error| E7[/"... propagated"/]
    RD2 -->|ok| MQ[/"→ MQ: login-notification.email\nfire & forget"/]
    MQ --> OK[/"200 {user_uuid, access_token, refresh_token}"/]
```

---

## RestoreAccount

`POST /api/restore-account`

```mermaid
flowchart TD
    A([Start]) --> V1{validate email\n+ password}
    V1 -->|fail| E1[/"400 invalid email / invalid password"/]

    V1 -->|ok| DB1[GetUserByEmail]
    DB1 -->|not found| DUMMY[timing dummy:\nVerify dummyHash]
    DUMMY --> E2[/"400 wrong email or password"/]

    DB1 -->|ok| PWD{Verify\npassword}
    PWD -->|fail| E3[/"400 wrong email or password"/]

    PWD -->|ok| DEL{deleted_at\n== nil?}
    DEL -->|true| E4[/"400 account is not deleted"/]

    DEL -->|false| EXP{"nextCleanupTime\nafter(deleted_at + 30d)\n<= now?"}
    EXP -->|true| E5[/"403 restoration period has expired"/]

    EXP -->|false| DB2[RestoreUser в PostgreSQL]
    DB2 -->|error| E6[/"... propagated"/]
    DB2 -->|ok| OK[/"200 {}"/]
```
