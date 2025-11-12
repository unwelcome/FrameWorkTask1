package entities

import "github.com/unwelcome/FrameWorkTask1/v1/gateway/pkg/utils"

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
	e.FirstName = utils.FCapitalize(e.FirstName)

	// Валидация last_name
	if err := utils.ValidateLastName(e.LastName, 2, 30); err != nil {
		return err
	}
	e.LastName = utils.FCapitalize(e.LastName)

	// Валидация patronymic
	if err := utils.ValidatePatronymic(e.Patronymic, 2, 30); err != nil {
		return err
	}
	e.Patronymic = utils.FCapitalize(e.Patronymic)

	return nil
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
type LoginResponse struct {
	UserUUID     string `json:"user_uuid"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func (e *LoginRequest) Validate() error {
	// Валидация email
	if err := utils.ValidateEmail(e.Email); err != nil {
		return err
	}

	// Валидация password
	if err := utils.ValidatePassword(e.Password, 8, 30); err != nil {
		return err
	}

	return nil
}

type GetUserRequest struct {
	UserUUID string `json:"user_uuid"`
}
type GetUserResponse struct {
	UserUUID   string `json:"user_uuid"`
	Email      string `json:"email"`
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	Patronymic string `json:"patronymic"`
	CreatedAt  string `json:"created_at"`
}

func (e *GetUserRequest) Validate() error {
	// Валидация user_uuid
	if err := utils.ValidateUUID(e.UserUUID); err != nil {
		return err
	}

	return nil
}

type ChangePasswordRequest struct {
	Password string `json:"password"`
}
type ChangePasswordRequestFull struct {
	UserUUID string `json:"user_uuid"`
	Password string `json:"password"`
}
type ChangePasswordResponse struct {
}

func (e *ChangePasswordRequestFull) Validate() error {
	// Валидация user_uuid
	if err := utils.ValidateUUID(e.UserUUID); err != nil {
		return err
	}

	// Валидация password
	if err := utils.ValidatePassword(e.Password, 8, 30); err != nil {
		return err
	}

	return nil
}

type UpdateUserBioRequest struct {
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	Patronymic string `json:"patronymic"`
}
type UpdateUserBioRequestFull struct {
	UserUUID   string `json:"user_uuid"`
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	Patronymic string `json:"patronymic"`
}
type UpdateUserBioResponse struct {
}

func (e *UpdateUserBioRequestFull) Validate() error {
	// Валидация user_uuid
	if err := utils.ValidateUUID(e.UserUUID); err != nil {
		return err
	}

	// Валидация first_name
	if err := utils.ValidateFirstName(e.FirstName, 2, 30); err != nil {
		return err
	}
	e.FirstName = utils.FCapitalize(e.FirstName)

	// Валидация last_name
	if err := utils.ValidateLastName(e.LastName, 2, 30); err != nil {
		return err
	}
	e.LastName = utils.FCapitalize(e.LastName)

	// Валидация patronymic
	if err := utils.ValidatePatronymic(e.Patronymic, 2, 30); err != nil {
		return err
	}
	e.Patronymic = utils.FCapitalize(e.Patronymic)

	return nil
}

type DeleteUserRequest struct {
	TargetUUID string `json:"tagret_uuid"`
}
type DeleteUserRequestFull struct {
	InitiatorUUID string `json:"initiator_uuid"`
	TargetUUID    string `json:"tagret_uuid"`
}
type DeleteUserResponse struct {
}

func (e *DeleteUserRequestFull) Validate() error {
	// Валидация initiator_uuid
	if err := utils.ValidateUUID(e.InitiatorUUID); err != nil {
		return err
	}

	// Валидация target_uuid
	if err := utils.ValidateUUID(e.TargetUUID); err != nil {
		return err
	}

	return nil
}

type GetAllActiveTokensRequest struct {
	UserUUID string `json:"user_uuid"`
}
type GetAllActiveTokensResponse struct {
	Tokens []*TokenInfo `json:"tokens"`
}
type TokenInfo struct {
	Token string `json:"token"`
}

func (e *GetAllActiveTokensRequest) Validate() error {
	// Валидация user_uuid
	if err := utils.ValidateUUID(e.UserUUID); err != nil {
		return err
	}

	return nil
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}
type RefreshTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func (e *RefreshTokenRequest) Validate() error {
	// Валидация jwt токена
	if err := utils.ValidateJWT(e.RefreshToken); err != nil {
		return err
	}

	return nil
}

type RevokeTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}
type RevokeTokenResponse struct {
}

func (e *RevokeTokenRequest) Validate() error {
	// Валидация jwt токена
	if err := utils.ValidateJWT(e.RefreshToken); err != nil {
		return err
	}

	return nil
}

type RevokeAllTokensRequest struct {
	UserUUID string `json:"user_uuid"`
}
type RevokeAllTokensResponse struct {
}

func (e *RevokeAllTokensRequest) Validate() error {
	// Валидация user_uuid
	if err := utils.ValidateUUID(e.UserUUID); err != nil {
		return err
	}

	return nil
}
