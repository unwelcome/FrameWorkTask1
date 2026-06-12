package password

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

// Параметры Argon2id.
// Значения соответствуют RFC 9106 (второй рекомендованный вариант) и OWASP 2024.
const (
	memory      uint32 = 64 * 1024 // 64 MiB
	iterations  uint32 = 3
	parallelism uint8  = 4
	saltLen            = 16 // 128-bit соль
	keyLen      uint32 = 32 // 256-bit выход
)

// Hash хеширует пароль и возвращает строку в формате PHC:
//
//	$argon2id$v=19$m=65536,t=3,p=4$<base64-salt>$<base64-hash>
func Hash(plaintext string) (string, error) {
	salt := make([]byte, saltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("password: failed to generate salt: %w", err)
	}

	hash := argon2.IDKey([]byte(plaintext), salt, iterations, memory, parallelism, keyLen)

	return encode(salt, hash), nil
}

// Verify проверяет, совпадает ли plaintext с ранее сохранённым PHC-хешем.
// Использует constant-time сравнение для защиты от timing-атак.
func Verify(encodedHash, plaintext string) (bool, error) {
	salt, expectedHash, err := decode(encodedHash)
	if err != nil {
		return false, err
	}

	actualHash := argon2.IDKey([]byte(plaintext), salt, iterations, memory, parallelism, keyLen)

	return subtle.ConstantTimeCompare(expectedHash, actualHash) == 1, nil
}

// DummyHash возвращает фиктивный хеш для выравнивания времени ответа,
// когда пользователь не найден (защита от timing-атаки на перечисление аккаунтов).
// Вызывается один раз при старте сервиса.
func DummyHash() string {
	h, err := Hash("$timing-protection-dummy$")
	if err != nil {
		panic("password: failed to pre-compute dummy hash: " + err.Error())
	}
	return h
}

// ── PHC encoding ─────────────────────────────────────────────────────────────

func encode(salt, hash []byte) string {
	b64salt := base64.RawStdEncoding.EncodeToString(salt)
	b64hash := base64.RawStdEncoding.EncodeToString(hash)
	return fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, memory, iterations, parallelism,
		b64salt, b64hash,
	)
}

func decode(encodedHash string) (salt, hash []byte, err error) {
	parts := strings.Split(encodedHash, "$")
	// "$argon2id$v=19$m=...,t=...,p=...$<salt>$<hash>" → 6 parts после split по "$"
	if len(parts) != 6 {
		return nil, nil, fmt.Errorf("password: invalid hash format")
	}
	if parts[1] != "argon2id" {
		return nil, nil, fmt.Errorf("password: unsupported algorithm %q", parts[1])
	}

	salt, err = base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return nil, nil, fmt.Errorf("password: failed to decode salt: %w", err)
	}
	hash, err = base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return nil, nil, fmt.Errorf("password: failed to decode hash: %w", err)
	}

	return salt, hash, nil
}
