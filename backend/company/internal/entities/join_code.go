package entities

import "time"

type CreateCompanyJoinCodeDTO struct {
	CompanyUUID string
	Code        string
	TTL         time.Duration
}

type CheckJoinCodeExistsDTO struct {
	Code string
}

type CheckJoinCodeBelongToCompanyDTO struct {
	CompanyUUID string
	Code        string
}

type GetCompanyJoinCodesDTO struct {
	CompanyUUID string
}

type GetCompanyByJoinCodeDTO struct {
	Code string
}

type DeleteCompanyJoinCodeDTO struct {
	CompanyUUID string
	Code        string
}
