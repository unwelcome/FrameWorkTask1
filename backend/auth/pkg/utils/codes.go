package utils

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

// GenerateTwoFACode Генерирует случайный 6-значный код для 2FA авторизации пользователя
func GenerateTwoFACode() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1_000_000))
	if err != nil {
		return "", fmt.Errorf("generate 6-digit code: %w", err)
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}
