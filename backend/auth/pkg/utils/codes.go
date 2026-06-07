package utils

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

// GenerateVerificationCode Генерирует случайный 6-значный код для верификации аккаунта пользователя
func GenerateVerificationCode() string {
	n, err := rand.Int(rand.Reader, big.NewInt(1_000_000))
	if err != nil {
		return "000000"
	}
	return fmt.Sprintf("%06d", n.Int64())
}

// GenerateRecoveryCode Генерирует случайный 6-значный код для восстановления пароля пользователя
func GenerateRecoveryCode() string {
	n, err := rand.Int(rand.Reader, big.NewInt(1_000_000))
	if err != nil {
		return "000000"
	}
	return fmt.Sprintf("%06d", n.Int64())
}

// GenerateTwoFACode Генерирует случайный 6-значный код для 2FA авторизации пользователя
func GenerateTwoFACode() string {
	n, err := rand.Int(rand.Reader, big.NewInt(1_000_000))
	if err != nil {
		return "000000"
	}
	return fmt.Sprintf("%06d", n.Int64())
}
