package validate

import (
	"fmt"
	"regexp"
)

const (
	PasswordMinLen         = 8
	PasswordMaxLen         = 72
	NameMinLen             = 2
	NameMaxLen             = 30
	CompanyTitleMaxLen     = 255
	DepartmentTitleMaxLen  = 255
	ApplicationTitleMaxLen = 255
)

var (
	reUUID                 = regexp.MustCompile(`^[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$`)
	reEmail                = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	reWhitespace           = regexp.MustCompile(`\s`)
	rePasswordLower        = regexp.MustCompile(`[a-zа-яё]`)
	rePasswordUpper        = regexp.MustCompile(`[A-ZА-ЯЁ]`)
	rePasswordDigit        = regexp.MustCompile(`[0-9]`)
	reJoinCode             = regexp.MustCompile(`^[0-9]{6}$`)
	reUserVerificationCode = regexp.MustCompile(`^[0-9]{6}$`)
	reTitle                = regexp.MustCompile(`^[a-zA-Zа-яА-ЯёЁ0-9\s\-_&.,№#()'"°+]+$`)
	reName                 = regexp.MustCompile(`^[a-zA-Zа-яА-ЯёЁ]+(['\-][a-zA-Zа-яА-ЯёЁ]+)*$`)
)

func UUID(uuid string) error {
	if uuid == "" {
		return fmt.Errorf("uuid missed")
	}
	if !reUUID.MatchString(uuid) {
		return fmt.Errorf("incorrect uuid")
	}
	return nil
}

func Email(email string) error {
	if email == "" {
		return fmt.Errorf("email missed")
	}
	if !reEmail.MatchString(email) {
		return fmt.Errorf("incorrect email")
	}
	return nil
}

func Password(password string) error {
	if password == "" {
		return fmt.Errorf("password missed")
	}
	if reWhitespace.MatchString(password) {
		return fmt.Errorf("password must not contain spaces")
	}
	if len(password) < PasswordMinLen {
		return fmt.Errorf("password should be more than %d characters", PasswordMinLen)
	}
	if len([]byte(password)) > PasswordMaxLen {
		return fmt.Errorf("password should be less than %d characters", PasswordMaxLen)
	}
	if !rePasswordLower.MatchString(password) {
		return fmt.Errorf("password must contain lowercase character")
	}
	if !rePasswordUpper.MatchString(password) {
		return fmt.Errorf("password must contain uppercase character")
	}
	if !rePasswordDigit.MatchString(password) {
		return fmt.Errorf("password must contain digit")
	}
	return nil
}

func FirstName(s string) error {
	if s == "" {
		return fmt.Errorf("first name missed")
	}
	if !reName.MatchString(s) {
		return fmt.Errorf("first name contains incorrect characters")
	}
	if len(s) < NameMinLen {
		return fmt.Errorf("first name must be more than %d characters", NameMinLen)
	}
	if len(s) > NameMaxLen {
		return fmt.Errorf("first name must be less than %d characters", NameMaxLen)
	}
	return nil
}

func LastName(s string) error {
	if s == "" {
		return fmt.Errorf("last name missed")
	}
	if !reName.MatchString(s) {
		return fmt.Errorf("last name contains incorrect characters")
	}
	if len(s) < NameMinLen {
		return fmt.Errorf("last name must be more than %d characters", NameMinLen)
	}
	if len(s) > NameMaxLen {
		return fmt.Errorf("last name must be less than %d characters", NameMaxLen)
	}
	return nil
}

func Patronymic(s string) error {
	if s == "" {
		return nil
	}
	if !reName.MatchString(s) {
		return fmt.Errorf("patronymic contains incorrect characters")
	}
	return nil
}

func CompanyJoinCode(code string) error {
	if code == "" {
		return fmt.Errorf("join code missed")
	}
	if !reJoinCode.MatchString(code) {
		return fmt.Errorf("invalid join code")
	}
	return nil
}

func UserVerificationCode(code string) error {
	if code == "" {
		return fmt.Errorf("verification code missed")
	}
	if !reUserVerificationCode.MatchString(code) {
		return fmt.Errorf("invalid verification code")
	}
	return nil
}

func CompanyTitle(title string) error {
	return checkTitle(title, "company", CompanyTitleMaxLen)
}

func DepartmentTitle(title string) error {
	return checkTitle(title, "department", DepartmentTitleMaxLen)
}

func ApplicationTitle(title string) error {
	return checkTitle(title, "application", ApplicationTitleMaxLen)
}

func ApplicationDescription(description string) error {
	if description == "" {
		return fmt.Errorf("description missed")
	}
	return nil
}

func Number(num int, min, max *int, title string) error {
	if title == "" {
		title = "digit"
	}
	if min != nil && num < *min {
		return fmt.Errorf("%s should be more than %d", title, *min)
	}
	if max != nil && num > *max {
		return fmt.Errorf("%s should be less than %d", title, *max)
	}
	return nil
}

func IntPtr(n int) *int { return &n }

func checkTitle(title, msg string, maxLen int) error {
	if title == "" {
		return fmt.Errorf("%s title missed", msg)
	}
	if len([]rune(title)) > maxLen {
		return fmt.Errorf("%s title is too long", msg)
	}
	if !reTitle.MatchString(title) {
		return fmt.Errorf("%s title contains incorrect characters", msg)
	}
	return nil
}
