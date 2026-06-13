package entities

import (
	"fmt"
	"strings"

	"github.com/unwelcome/FrameWorkTask1/backend/gateway/pkg/utils"
	"github.com/unwelcome/FrameWorkTask1/backend/shared/validate"
)

// ─── Shared response types ────────────────────────────────────────────────────

// TokenInfo описывает активную сессию пользователя.
type TokenInfo struct {
	TokenHash      string `json:"token_hash"`
	IP             string `json:"ip"`
	LastIP         string `json:"last_ip"`
	ISP            string `json:"isp,omitempty"`
	CountryCode    string `json:"country_code,omitempty"`
	CountryName    string `json:"country_name,omitempty"`
	City           string `json:"city,omitempty"`
	Timezone       string `json:"timezone,omitempty"`
	DeviceType     string `json:"device_type,omitempty"`
	OS             string `json:"os,omitempty"`
	OSVersion      string `json:"os_version,omitempty"`
	Browser        string `json:"browser,omitempty"`
	BrowserVersion string `json:"browser_version,omitempty"`
	CreatedAt      string `json:"created_at"`
	LastActiveAt   string `json:"last_active_at"`
}

// ─── Register ─────────────────────────────────────────────────────────────────

type RegisterRequest struct {
	Email      string `json:"email"`
	Password   string `json:"password"`
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	Patronymic string `json:"patronymic"`
}

func (e *RegisterRequest) Validate() error {
	e.Email = strings.TrimSpace(e.Email)
	if err := validate.Email(e.Email); err != nil {
		return err
	}
	e.Password = strings.TrimSpace(e.Password)
	if err := validate.Password(e.Password); err != nil {
		return err
	}
	e.FirstName = utils.FCapitalize(strings.TrimSpace(e.FirstName))
	if err := validate.FirstName(e.FirstName); err != nil {
		return err
	}
	e.LastName = utils.FCapitalize(strings.TrimSpace(e.LastName))
	if err := validate.LastName(e.LastName); err != nil {
		return err
	}
	e.Patronymic = utils.FCapitalize(strings.TrimSpace(e.Patronymic))
	if err := validate.Patronymic(e.Patronymic); err != nil {
		return err
	}
	return nil
}

// ─── Login ────────────────────────────────────────────────────────────────────

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
type LoginResponse struct {
	UserUUID     string `json:"user_uuid"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	SessionUUID  string `json:"session_uuid"`
}

func (e *LoginRequest) Validate() error {
	e.Email = strings.TrimSpace(e.Email)
	if err := validate.Email(e.Email); err != nil {
		return err
	}
	e.Password = strings.TrimSpace(e.Password)
	if err := validate.Password(e.Password); err != nil {
		return err
	}
	return nil
}

// ─── GetUser ──────────────────────────────────────────────────────────────────

type GetUserRequest struct {
	UserUUID string `json:"-"`
}
type GetUserResponse struct {
	UserUUID    string `json:"user_uuid"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	Patronymic  string `json:"patronymic"`
	Description string `json:"description,omitempty"`
	CreatedAt   string `json:"created_at"`
	DeletedAt   string `json:"deleted_at,omitempty"`
}

func (e *GetUserRequest) Validate() error {
	e.UserUUID = strings.TrimSpace(e.UserUUID)
	return validate.UUID(e.UserUUID)
}

// ─── ChangePassword ───────────────────────────────────────────────────────────

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password"`
	Password    string `json:"password"`
}
type ChangePasswordResponse struct{}

func (e *ChangePasswordRequest) Validate() error {
	e.OldPassword = strings.TrimSpace(e.OldPassword)
	if err := validate.Password(e.OldPassword); err != nil {
		return fmt.Errorf("old_password: %w", err)
	}
	e.Password = strings.TrimSpace(e.Password)
	if err := validate.Password(e.Password); err != nil {
		return fmt.Errorf("password: %w", err)
	}
	return nil
}

// ─── UpdateUserBio ────────────────────────────────────────────────────────────

type UpdateUserBioRequest struct {
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	Patronymic  string `json:"patronymic"`
	Description string `json:"description"`
}
type UpdateUserBioResponse struct{}

func (e *UpdateUserBioRequest) Validate() error {
	e.FirstName = utils.FCapitalize(strings.TrimSpace(e.FirstName))
	if err := validate.FirstName(e.FirstName); err != nil {
		return err
	}
	e.LastName = utils.FCapitalize(strings.TrimSpace(e.LastName))
	if err := validate.LastName(e.LastName); err != nil {
		return err
	}
	e.Patronymic = utils.FCapitalize(strings.TrimSpace(e.Patronymic))
	if err := validate.Patronymic(e.Patronymic); err != nil {
		return err
	}
	e.Description = strings.TrimSpace(e.Description)
	if err := validate.UserDescription(e.Description); err != nil {
		return fmt.Errorf("description: %w", err)
	}
	return nil
}

// ─── DeleteUser ───────────────────────────────────────────────────────────────

type DeleteUserResponse struct{}

// ─── GetAllActiveTokens ───────────────────────────────────────────────────────

type GetAllActiveTokensRequest struct{}
type GetAllActiveTokensResponse struct {
	Tokens []*TokenInfo `json:"tokens"`
}

// ─── RefreshToken ─────────────────────────────────────────────────────────────

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}
type RefreshTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func (e *RefreshTokenRequest) Validate() error {
	e.RefreshToken = strings.TrimSpace(e.RefreshToken)
	if err := utils.ValidateJWT(e.RefreshToken); err != nil {
		return err
	}
	return nil
}

// ─── RevokeToken ──────────────────────────────────────────────────────────────

type RevokeTokenRequest struct {
	TokenHash string `json:"token_hash"`
}
type RevokeTokenResponse struct{}

func (e *RevokeTokenRequest) Validate() error {
	e.TokenHash = strings.TrimSpace(e.TokenHash)
	if e.TokenHash == "" {
		return fmt.Errorf("token_hash is required")
	}
	return nil
}

// ─── RevokeAllTokens ──────────────────────────────────────────────────────────

type RevokeAllTokensRequest struct{}
type RevokeAllTokensResponse struct{}

// ─── VerifyAccount ────────────────────────────────────────────────────────────

type VerifyAccountRequest struct {
	Email string `json:"email"`
	Code  string `json:"code"`
}
type VerifyAccountResponse struct{}

func (e *VerifyAccountRequest) Validate() error {
	e.Email = strings.TrimSpace(e.Email)
	if err := validate.Email(e.Email); err != nil {
		return err
	}
	e.Code = strings.TrimSpace(e.Code)
	if err := validate.UserVerificationCode(e.Code); err != nil {
		return err
	}
	return nil
}

// ─── ResendVerificationCode ───────────────────────────────────────────────────

type ResendVerificationCodeRequest struct {
	Email string `json:"email"`
}
type ResendVerificationCodeResponse struct{}

func (e *ResendVerificationCodeRequest) Validate() error {
	e.Email = strings.TrimSpace(e.Email)
	return validate.Email(e.Email)
}

// ─── ForgotPassword ───────────────────────────────────────────────────────────

type ForgotPasswordRequest struct {
	Email string `json:"email"`
}
type ForgotPasswordResponse struct{}

func (e *ForgotPasswordRequest) Validate() error {
	e.Email = strings.TrimSpace(e.Email)
	return validate.Email(e.Email)
}

// ─── ResetPassword ────────────────────────────────────────────────────────────

type ResetPasswordRequest struct {
	ResetToken  string `json:"reset_token"`
	NewPassword string `json:"new_password"`
}
type ResetPasswordResponse struct{}

func (e *ResetPasswordRequest) Validate() error {
	e.ResetToken = strings.TrimSpace(e.ResetToken)
	if err := utils.ValidateJWT(e.ResetToken); err != nil {
		return fmt.Errorf("reset_token: %w", err)
	}
	e.NewPassword = strings.TrimSpace(e.NewPassword)
	return validate.Password(e.NewPassword)
}

// ─── Verify2FA ────────────────────────────────────────────────────────────────

type Verify2FARequest struct {
	SessionUUID string `json:"session_uuid"`
	Code        string `json:"code"`
}
type Verify2FAResponse struct {
	UserUUID     string `json:"user_uuid"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func (e *Verify2FARequest) Validate() error {
	e.SessionUUID = strings.TrimSpace(e.SessionUUID)
	if err := validate.UUID(e.SessionUUID); err != nil {
		return err
	}
	e.Code = strings.TrimSpace(e.Code)
	return validate.User2FACCode(e.Code)
}

// ─── RestoreAccount ───────────────────────────────────────────────────────────

type RestoreAccountRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
type RestoreAccountResponse struct{}

func (e *RestoreAccountRequest) Validate() error {
	e.Email = strings.TrimSpace(e.Email)
	if err := validate.Email(e.Email); err != nil {
		return err
	}
	e.Password = strings.TrimSpace(e.Password)
	return validate.Password(e.Password)
}

// ─── UpdateUser2FA ────────────────────────────────────────────────────────────

type UpdateUser2FARequest struct {
	UserUUID  string `json:"-"`
	Enable2FA bool   `json:"enable_2fa"`
}
type UpdateUser2FAResponse struct{}

func (e *UpdateUser2FARequest) Validate() error {
	e.UserUUID = strings.TrimSpace(e.UserUUID)
	return validate.UUID(e.UserUUID)
}
