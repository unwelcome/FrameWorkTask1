# Жизненный цикл аккаунта

## Диаграмма состояний

```mermaid
stateDiagram-v2
    direction LR

    [*] --> Unverified  : Register
    Unverified --> Active    : VerifyAccount
    Unverified --> Deleted   : DeleteUser
    Active     --> Deleted   : DeleteUser (отзывает токены)
    Deleted    --> Active    : RestoreAccount (≤30 дней)
    Deleted    --> Anonymized: cleanup 00:00 UTC (>30 дней)
    Anonymized --> [*]
```

---

## Таблица переходов

| Из | В | Метод | Ключевое условие | Side effects |
|---|---|---|---|---|
| — | Unverified | `Register` | уникальный email | → письмо верификации |
| Unverified | Active | `VerifyAccount` | правильный код, ≤5 попыток | код удаляется из Redis |
| Unverified | Deleted | `DeleteUser` | initiator == target | отзыв токенов |
| Active | Deleted | `DeleteUser` | initiator == target | отзыв всех токенов |
| Deleted | Active | `RestoreAccount` | верный пароль, ≤30 дней | — |
| Deleted | Anonymized | cleanup worker | `deleted_at < now - 30d`, 00:00 UTC | email/name/password_hash → NULL |

---

## Доступные операции по состоянию

| Операция | Unverified | Active | Deleted | Anonymized |
|---|:---:|:---:|:---:|:---:|
| Login | ✗ 403 | ✓ | ✗ 403 | ✗ 400¹ |
| VerifyAccount | ✓ | ✗ 409 | ✓² | ✗² |
| ChangePassword | ✓ | ✓ | ✓ | ✗ 400³ |
| UpdateUserBio | ✓ | ✓ | ✓ | ✓ |
| UpdateUser2FA | ✓ | ✓ | ✓ | ✓ |
| DeleteUser | ✓ | ✓ | ✗ 404 | ✗ 404 |
| RestoreAccount | ✓⁴ | ✗ 400 | ✓ | ✗ 400³ |
| RefreshToken | ✓ | ✓ | ✗⁵ | ✗⁵ |
| GetUser | ✓ | ✓ | ✓ (deleted_at≠"") | ✓ (поля="") |

¹ email = NULL → не находится по email → "wrong email or password"  
² после мягкого удаления код верификации ещё может быть в Redis  
³ `password_hash` = NULL → `Verify("", pwd)` → false → "wrong email or password"  
⁴ восстанавливается в **Unverified** — нужна повторная верификация  
⁵ защищено косвенно: `DeleteUser` отзывает все токены; `CheckRefreshTokenExists` вернёт 404
