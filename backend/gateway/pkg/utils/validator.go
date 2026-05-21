package utils

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

var reJWT = regexp.MustCompile(`^[A-Za-z0-9_-]{20,}\.[A-Za-z0-9_-]{20,}\.[A-Za-z0-9_-]{20,}$`)

func ValidateJWT(token string) error {
	if token == "" {
		return fmt.Errorf("token missed")
	}
	if !reJWT.MatchString(token) {
		return fmt.Errorf("incorrect token")
	}
	return nil
}

func FCapitalize(str string) string {
	if str == "" {
		return ""
	}

	runes := []rune(strings.ToLower(str))
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}
