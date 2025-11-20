package entities

import "github.com/unwelcome/FrameWorkTask1/v1/gateway/pkg/utils"

type CreateCompanyRequest struct {
	Title string `json:"title"`
}
type CreateCompanyResponse struct {
	CompanyUUID string `json:"company_uuid"`
}

func (e *CreateCompanyRequest) Validate() error {
	// Валидация title
	if err := utils.ValidateCompanyTitle(e.Title); err != nil {
		return err
	}

	return nil
}
