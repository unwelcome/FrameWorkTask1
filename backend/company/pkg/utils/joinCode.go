package utils

import (
	"crypto/rand"
	"math/big"
)

// GenerateJoinCode Генерирует криптографически случайную строку цифр длиной JoinCodeLength
func GenerateJoinCode(joinCodeLength int) (string, error) {
	digits := make([]byte, joinCodeLength)

	for i := 0; i < joinCodeLength; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", err
		}
		digits[i] = byte('0' + n.Int64())
	}

	return string(digits), nil
}
