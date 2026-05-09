package entities

type ApplicationStatistic struct {
	Created          int `db:"created"`
	Assigned         int `db:"assigned"`
	Redirected       int `db:"redirected"`
	Rejected         int `db:"rejected"`
	InProgress       int `db:"in_progress"`
	OnHold           int `db:"on_hold"`
	AwaitingApproval int `db:"awaiting_approval"`
	Completed        int `db:"completed"`
	Cancelled        int `db:"cancelled"`
	Failed           int `db:"failed"`
	Archived         int `db:"archived"`
}

type GetCompanyApplicationStatisticDTO struct {
	CompanyUUID string
	WithDeleted bool
}

type GetEmployeeApplicationStatisticDTO struct {
	CompanyUUID string
	TargetUUID  string
	TargetRole  string
}
