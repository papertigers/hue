package config

type CreateUser struct {
	DeviceType string `json:"devicetype"`
}

// CreateResult struct
type CreateUserResult struct {
	Success struct {
		Username string `json:"username"`
	} `json:"success"`
}
