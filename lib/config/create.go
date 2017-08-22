package config

type CreateUser struct {
	DeviceType string `json:"devicetype"`
}

// CreateResult struct
type CreateResult struct {
	Success struct {
		Username string `json:"username"`
	} `json:"success"`
}
