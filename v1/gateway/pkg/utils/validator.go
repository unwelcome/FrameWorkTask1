package utils

import (
	"fmt"
	"regexp"
	"slices"
	"strings"
	"unicode"
)

// ValidateUUID Проверяет uuid по регулярному выражению
func ValidateUUID(uuid string) error {
	if uuid == "" {
		return fmt.Errorf("uuid missed")
	}

	pattern := `^[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$`
	if !regexp.MustCompile(pattern).MatchString(uuid) {
		return fmt.Errorf("incorrect uuid")
	}
	return nil
}

// ValidateJWT Проверка строки на схожесть на jwt токен
func ValidateJWT(token string) error {
	if token == "" {
		return fmt.Errorf("token missed")
	}

	pattern := `^[A-Za-z0-9_-]{20,}\.[A-Za-z0-9_-]{20,}\.[A-Za-z0-9_-]{20,}$`
	if !regexp.MustCompile(pattern).MatchString(token) {
		return fmt.Errorf("incorrect token")
	}
	return nil
}

// ValidateEmail Проверяет email по регулярному выражению
func ValidateEmail(email string) error {
	if email == "" {
		return fmt.Errorf("email missed")
	}

	pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`

	if !regexp.MustCompile(pattern).MatchString(email) {
		return fmt.Errorf("incorrect email")
	}
	return nil
}

// ValidatePassword Проверка пароля пользователя
func ValidatePassword(password string, minLen, maxLen int) error {
	if password == "" {
		return fmt.Errorf("password missed")
	}

	if strings.TrimSpace(password) != password {
		return fmt.Errorf("password contains extra spaces")
	}

	if len(password) < minLen {
		return fmt.Errorf("password should be more than %d characters", minLen)
	}
	if len(password) > maxLen {
		return fmt.Errorf("password should be less than %d characters", maxLen)
	}

	if !regexp.MustCompile(`^[a-zа-яA-ZА-ЯёЁ0-9]+$`).MatchString(password) {
		return fmt.Errorf("password cointains incorrect characters")
	}

	if !regexp.MustCompile(`[a-zа-яё]`).MatchString(password) {
		return fmt.Errorf("password must cointain lowercase character")
	}

	if !regexp.MustCompile(`[A-ZА-ЯЁ]`).MatchString(password) {
		return fmt.Errorf("password must cointain uppercase character")
	}

	if !regexp.MustCompile(`[0-9]`).MatchString(password) {
		return fmt.Errorf("password must cointain digit")
	}

	return nil
}

// ValidateFirstName Проверка имени пользователя
func ValidateFirstName(s string, minLen, maxLen int) error {
	s = strings.TrimSpace(s)

	if s == "" {
		return fmt.Errorf("first name missed")
	}

	if !checkStringCharacters(s) {
		return fmt.Errorf("first name contains incorrect characters")
	}

	length := len(s)
	if length < minLen {
		return fmt.Errorf("first name must be more than %d characters", minLen)
	}

	if length > maxLen {
		return fmt.Errorf("first name must be less than %d characters", maxLen)
	}

	return nil
}

// ValidateLastName Проверка фамилии пользователя
func ValidateLastName(s string, minLen, maxLen int) error {
	s = strings.TrimSpace(s)

	if s == "" {
		return fmt.Errorf("last name missed")
	}

	if !checkStringCharacters(s) {
		return fmt.Errorf("last name contains incorrect characters")
	}

	length := len(s)
	if length < minLen {
		return fmt.Errorf("last name must be more than %d characters", minLen)
	}

	if length > maxLen {
		return fmt.Errorf("last name must be less than %d characters", maxLen)
	}

	return nil
}

// ValidatePatronymic Проверка отчества пользователя (при наличии)
func ValidatePatronymic(s string, minLen, maxLen int) error {
	s = strings.TrimSpace(s)

	if s == "" {
		return nil
	}

	if !checkStringCharacters(s) {
		return fmt.Errorf("patronymic contains incorrect characters")
	}

	return nil
}

// ValidateCompanyTitle Проверяет название компании
func ValidateCompanyTitle(title string) error {
	if title == "" {
		return fmt.Errorf("title missed")
	}

	if strings.TrimSpace(title) != title {
		return fmt.Errorf("title contains useless spaces")
	}

	if len([]rune(title)) > 250 {
		return fmt.Errorf("title is too long")
	}

	pattern := `^[a-zA-Zа-яА-ЯёЁ0-9\s\-_&.,№#()'"°+]+$`
	if !regexp.MustCompile(pattern).MatchString(title) {
		return fmt.Errorf("title contains incorrect characters")
	}

	return nil
}

// ValidateCompanyJoinCode Проверяет код добавления в компанию
func ValidateCompanyJoinCode(code string) error {
	if code == "" {
		return fmt.Errorf("join code missed")
	}

	pattern := `^[0-9]{8}$`
	if !regexp.MustCompile(pattern).MatchString(code) {
		return fmt.Errorf("invalid join code")
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

// ValidateNumber проверка числа
func ValidateNumber(num, min, max int, title string) error {
	if title == "" {
		title = "digit"
	}

	if num < min {
		return fmt.Errorf("%s should be more than %d", title, min)
	}

	if max != 0 && num > max {
		return fmt.Errorf("%s should be less than %d", title, max)
	}

	return nil
}

// ValidateIsArrayContain Проверяет, содержится ли объект в массиве объектов
func ValidateIsArrayContain[T int | string | float64](str T, arr []T) bool {
	return slices.Contains(arr, str)
}

func checkStringCharacters(str string) bool {
	return regexp.MustCompile(`^[a-zA-Zа-яА-ЯёЁ]+([\'\-][a-zA-Zа-яА-ЯёЁ]+)*$`).MatchString(str)
}
