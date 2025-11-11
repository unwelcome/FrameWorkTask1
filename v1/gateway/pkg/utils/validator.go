package utils

import (
	"fmt"
	"regexp"
	"strings"
)

// ValidateEmail проверяет email по регулярному выражению
func ValidateEmail(email string) error {
	pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`

	matched, _ := regexp.MatchString(pattern, email)
	if !matched {
		return fmt.Errorf("incorrect email")
	}
	return nil
}

// ValidatePassword Проверка пароля пользователя
func ValidatePassword(password string, minLen, maxLen int) error {
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

// ValidateNumber проверка числа
func ValidateNumber(num, min, max int, title string) error {
	if title == "" {
		title = "digit"
	}

	if num < min {
		return fmt.Errorf("%s should be more than %d", title, min)
	}

	if num > max {
		return fmt.Errorf("%s chould be less than %d", title, max)
	}

	return nil
}

func checkStringCharacters(str string) bool {
	return regexp.MustCompile(`^[a-zA-Zа-яА-ЯёЁ]+([\'\-][a-zA-Zа-яА-ЯёЁ]+)*$`).MatchString(str)
}
