package entities

import (
	"github.com/unwelcome/FrameWorkTask1/v1/gateway/pkg/utils"
)

type RegisterRequest struct {
	Email      string `json:"email"`
	Password   string `json:"password"`
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	Patronymic string `json:"patronymic"`
}
type RegisterResponse struct {
	UserUUID string `json:"user_uuid"`
}

func (e *RegisterRequest) Validate() error {
	// Валидация email
	if err := utils.ValidateEmail(e.Email); err != nil {
		return err
	}

	// Валидация password
	if err := utils.ValidatePassword(e.Password, 8, 30); err != nil {
		return err
	}

	// Валидация first_name
	if err := utils.ValidateFirstName(e.FirstName, 2, 30); err != nil {
		return err
	}

	// Валидация last_name
	if err := utils.ValidateLastName(e.LastName, 2, 30); err != nil {
		return err
	}

	// Валидация patronymic
	if err := utils.ValidatePatronymic(e.Patronymic, 2, 30); err != nil {
		return err
	}

	return nil
}
