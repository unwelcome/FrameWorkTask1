package password

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/argon2"
)

// Параметры Argon2id для НОВЫХ хешей.
// Значения соответствуют RFC 9106 (второй рекомендованный вариант) и OWASP 2024.
//
// parallelism=1: на сервере под нагрузкой высокий parallelism лишь множит
// контеншн потоков (p потоков на каждое вычисление), не усиливая защиту —
// memory-hardness обеспечивается параметром memory. Старые хеши с другим p
// остаются валидными: Verify читает параметры из самого хранимого хеша.
const (
	memory      uint32 = 64 * 1024 // 64 MiB
	iterations  uint32 = 3
	parallelism uint8  = 1
	saltLen            = 16 // 128-bit соль
	keyLen      uint32 = 32 // 256-bit выход
)

// ErrOverloaded возвращается, когда превышен лимит одновременных вычислений
var ErrOverloaded = errors.New("password: hashing capacity exceeded")

// hashSlots — семафор, ограничивающий число одновременных вычислений Argon2.
// nil (Configure не вызывали) → ограничения нет (для unit-тестов и CLI-утилит).
var (
	hashSlots      chan struct{}
	hashAcquireTTL = 3 * time.Second
)

// Setup инициализирует ограничитель одновременных вычислений Argon2.
// Вызывается один раз при старте сервиса, до приёма трафика.
//   - maxConcurrent — максимум одновременных хеширований
//   - acquireTimeout — сколько ждать освобождения слота, прежде чем вернуть ErrOverloaded
func Setup(maxConcurrent int, acquireTimeout time.Duration) {
	if maxConcurrent < 1 {
		maxConcurrent = 1
	}
	hashSlots = make(chan struct{}, maxConcurrent)
	if acquireTimeout > 0 {
		hashAcquireTTL = acquireTimeout
	}
}

// acquire занимает слот семафора с учётом таймаута и отмены ctx.
func acquire(ctx context.Context) error {
	if hashSlots == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(ctx, hashAcquireTTL)
	defer cancel()

	select {
	case hashSlots <- struct{}{}:
		return nil
	case <-ctx.Done():
		return ErrOverloaded
	}
}

func release() {
	if hashSlots != nil {
		<-hashSlots
	}
}

// Hash хеширует пароль и возвращает строку в формате PHC:
//
//	$argon2id$v=19$m=65536,t=3,p=1$<base64-salt>$<base64-hash>
//
// Перед вычислением занимает слот семафора; при перегрузке возвращает ErrOverloaded.
func Hash(ctx context.Context, plaintext string) (string, error) {
	if err := acquire(ctx); err != nil {
		return "", err
	}
	defer release()

	salt := make([]byte, saltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("password: failed to generate salt: %w", err)
	}

	hash := argon2.IDKey([]byte(plaintext), salt, iterations, memory, parallelism, keyLen)

	return encode(salt, hash), nil
}

// Verify проверяет, совпадает ли plaintext с ранее сохранённым PHC-хешем.
// Параметры Argon2 (m, t, p) читаются из самого хеша, поэтому хеши, созданные
// с другими параметрами, остаются валидными. Использует constant-time сравнение.
// Перед вычислением занимает слот семафора; при перегрузке возвращает ErrOverloaded.
func Verify(ctx context.Context, encodedHash, plaintext string) (bool, error) {
	p, err := decode(encodedHash)
	if err != nil {
		return false, err
	}

	if err := acquire(ctx); err != nil {
		return false, err
	}
	defer release()

	actualHash := argon2.IDKey([]byte(plaintext), p.salt, p.iterations, p.memory, p.parallelism, uint32(len(p.hash)))

	return subtle.ConstantTimeCompare(p.hash, actualHash) == 1, nil
}

// DummyHash возвращает фиктивный хеш для выравнивания времени ответа,
// когда пользователь не найден (защита от timing-атаки на перечисление аккаунтов).
// Вызывается один раз при старте сервиса.
func DummyHash() string {
	h, err := Hash(context.Background(), "$timing-protection-dummy$")
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

// phcParams — разобранное содержимое PHC-строки Argon2id.
type phcParams struct {
	memory      uint32
	iterations  uint32
	parallelism uint8
	salt        []byte
	hash        []byte
}

func decode(encodedHash string) (phcParams, error) {
	var p phcParams

	parts := strings.Split(encodedHash, "$")
	// "$argon2id$v=19$m=...,t=...,p=...$<salt>$<hash>" → 6 частей после split по "$"
	if len(parts) != 6 {
		return p, fmt.Errorf("password: invalid hash format")
	}
	if parts[1] != "argon2id" {
		return p, fmt.Errorf("password: unsupported algorithm %q", parts[1])
	}
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &p.memory, &p.iterations, &p.parallelism); err != nil {
		return p, fmt.Errorf("password: failed to parse parameters: %w", err)
	}

	var err error
	p.salt, err = base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return p, fmt.Errorf("password: failed to decode salt: %w", err)
	}
	p.hash, err = base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return p, fmt.Errorf("password: failed to decode hash: %w", err)
	}

	return p, nil
}
