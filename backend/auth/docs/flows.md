# Flowcharts методов auth сервиса

Методы с нетривиальной логикой ветвлений. Линейные методы (UpdateUserBio, RevokeToken,
GetAllActiveTokens, UpdateUser2FA) не включены.

Все защищённые маршруты (`/auth/*`) неявно добавляют **401** от JWT middleware до вызова сервиса.

---

## Register

`POST /api/register`

```mermaid
flowchart TD
    A([Start]) --> V1{validate\nemail / password\nfirst/last name}
    V1 -->|fail| E1[/"400 invalid ..."/]

    V1 -->|ok| DB1[CreateUser в PostgreSQL]
    DB1 -->|success| RC[SaveVerificationCode\nв Redis]
    RC --> MQ1[/"→ MQ: verification.email"/]
    MQ1 --> OK[/"201 {user_uuid}"/]

    DB1 -->|AlreadyExists| DB2[GetUserByEmail]
    DB2 -->|error| E2[/"500 internal error"/]

    DB2 -->|ok| CHK{IsVerified?}
    CHK -->|true| MQ2[/"→ MQ: registration-attempt.email\nfire & forget"/]
    MQ2 --> E3[/"409 email already registered"/]

    CHK -->|false| RD1[GetVerificationCode\nиз Redis]
    RD1 -->|код активен| E4[/"409 verification email already sent"\ncheck your inbox/]

    RD1 -->|код истёк| DEL[DeleteUser старого\nневерифицированного аккаунта]
    DEL --> CRE2[CreateUser нового]
    CRE2 -->|error| E5[/"500 internal error"/]
    CRE2 -->|ok| RC

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
    DB1 -->|not found| DUMMY1[timing dummy:\nVerify dummyHash]
    DUMMY1 --> E2[/"400 wrong email or password"/]

    DB1 -->|ok| PWD{Verify\npassword}
    PWD -->|fail| E3[/"400 wrong email or password"/]

    PWD -->|ok| DEL{deleted_at\n!= nil?}
    DEL -->|true| E4[/"403 account is deleted\nyou have N hours to restore it"/]

    DEL -->|false| VER{IsVerified?}
    VER -->|false| E5[/"403 account not verified"/]

    VER -->|true| TFA{Enabled2FA?}

    TFA -->|true| S1[Save2FAData в Redis\nsessionUUID + code]
    S1 -->|error| E6[/"500 ... propagated"/]
    S1 -->|ok| MQ1[/"→ MQ: 2fa.email\nfire & forget"/]
    MQ1 --> OK1[/"200 {session_uuid}"/]

    TFA -->|false| TK[CreateTokens JWT]
    TK -->|error| E7[/"500 internal error"/]
    TK -->|ok| RD1[SaveRefreshToken в Redis]
    RD1 -->|error| E8[/"500 ... propagated"/]
    RD1 -->|ok| OK2[/"200 {user_uuid, access_token, refresh_token}"/]
```

---

## VerifyAccount

`POST /api/user/verify`

```mermaid
flowchart TD
    A([Start]) --> V1{validate email\n+ code format}
    V1 -->|fail| E1[/"400 invalid email / invalid verification code format"/]

    V1 -->|ok| DB1[GetUserByEmail]
    DB1 -->|not found| DUMMY[timing dummy:\nGetVerificationCode dummyUUID]
    DUMMY --> E2[/"400 invalid email or code"/]

    DB1 -->|ok| VER{IsVerified?}
    VER -->|true| E3[/"409 account already verified"/]

    VER -->|false| RD1[GetVerificationCode\nиз Redis]
    RD1 -->|not found / expired| E4[/"400 invalid or expired verification code"/]

    RD1 -->|ok| INC[IncrVerificationAttempts]
    INC -->|error| E5[/"500 internal error"/]

    INC -->|ok| ATT{attempts\n> 5?}
    ATT -->|true| DEL1[DeleteVerificationCode\nfire & forget]
    DEL1 --> E6[/"429 too many attempts\nplease request a new verification code"/]

    ATT -->|false| CMP{code\n== stored?}
    CMP -->|false| E7[/"400 invalid or expired verification code"/]

    CMP -->|true| DEL2[DeleteVerificationCode]
    DEL2 --> DB2[SetUserVerified в PostgreSQL]
    DB2 -->|error| E8[/"500 ... propagated"/]
    DB2 -->|ok| OK[/"200 {}"/]
```

---

## ResetPassword

`POST /api/reset-password`

```mermaid
flowchart TD
    A([Start]) --> V1{validate email\n+ code format\n+ password}
    V1 -->|fail| E1[/"400 invalid email / invalid code format / invalid password"/]

    V1 -->|ok| DB1[GetUserByEmail]
    DB1 -->|not found| DUMMY1[timing dummy:\nGetRecoveryCode dummyUUID]
    DUMMY1 --> E2[/"400 invalid email or code"/]

    DB1 -->|ok| VER{IsVerified?}
    VER -->|false| DUMMY2[timing dummy:\nGetRecoveryCode dummyUUID]
    DUMMY2 --> E3[/"400 invalid email or code"/]

    VER -->|true| RD1[GetRecoveryCode из Redis]
    RD1 -->|not found / expired| E4[/"400 invalid or expired code"/]

    RD1 -->|ok| INC[IncrRecoveryAttempts]
    INC -->|error| E5[/"500 internal error"/]

    INC -->|ok| ATT{attempts\n> 5?}
    ATT -->|true| DEL1[DeleteRecoveryCode]
    DEL1 --> E6[/"429 too many attempts\nplease request a new recovery code"/]

    ATT -->|false| CMP{code\n== stored?}
    CMP -->|false| E7[/"400 invalid or expired code"/]

    CMP -->|true| DEL2[DeleteRecoveryCode]
    DEL2 --> HASH[Hash new password]
    HASH -->|error| E8[/"500 internal error"/]
    HASH -->|ok| DB2[UpdateUserPassword в PostgreSQL]
    DB2 -->|error| E9[/"... propagated"/]
    DB2 -->|ok| RVK[RevokeAllRefreshTokens\nNotFound игнорируется]
    RVK -->|error| E10[/"... propagated"/]
    RVK -->|ok| OK[/"200 {}"/]
```

---

## Verify2FA

`POST /api/verify-2fa`

```mermaid
flowchart TD
    A([Start]) --> V1{validate\nsession_uuid\n+ code format}
    V1 -->|fail| E1[/"400 invalid session uuid / invalid code format"/]

    V1 -->|ok| RD1[Get2FAData из Redis\nпо session_uuid]
    RD1 -->|not found| E2[/"404 ... propagated"/]

    RD1 -->|ok| INC[Incr2FAAttempts]
    INC -->|error| E3[/"... propagated"/]

    INC -->|ok| ATT{attempts\n> 5?}
    ATT -->|true| DEL1[Delete2FAData]
    DEL1 --> E4[/"429 too many attempts\nplease try login again"/]

    ATT -->|false| CMP{code\n== stored?}
    CMP -->|false| E5[/"403 invalid or expired code"/]

    CMP -->|true| DEL2[Delete2FAData\nfire & forget]
    DEL2 --> TK[CreateTokens JWT]
    TK -->|error| E6[/"500 internal error"/]
    TK -->|ok| RD2[SaveRefreshToken в Redis]
    RD2 -->|error| E7[/"... propagated"/]
    RD2 -->|ok| OK[/"200 {user_uuid, access_token, refresh_token}"/]
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

    DEL -->|false| EXP{"time.Since(deleted_at)\n> 30 дней?"}
    EXP -->|true| E5[/"403 restoration period has expired"/]

    EXP -->|false| DB2[RestoreUser в PostgreSQL]
    DB2 -->|error| E6[/"... propagated"/]
    DB2 -->|ok| OK[/"200 {}"/]
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

    TTP -->|true| RD1[CheckRefreshTokenExists\nв Redis]
    RD1 -->|not found| E3[/"404 ... propagated"/]

    RD1 -->|ok| DB1[GetUser из PostgreSQL]
    DB1 -->|not found| E4[/"404 ... propagated"/]

    DB1 -->|ok| TK[CreateTokens JWT новая пара]
    TK -->|error| E5[/"500 internal error"/]
    TK -->|ok| RD2["RefreshToken в Redis\n(атомарно: удалить старый + сохранить новый)"]
    RD2 -->|error| E6[/"... propagated"/]
    RD2 -->|ok| OK[/"200 {access_token, refresh_token}"/]
```
