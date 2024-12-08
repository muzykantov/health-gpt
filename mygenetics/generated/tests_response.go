package generated

import "time"

type TestsResponse []struct {
	ID              int       `json:"id,omitempty"`
	CreatedAt       time.Time `json:"createdAt,omitempty"`
	CodeLab         string    `json:"codeLab,omitempty"`
	Activated       bool      `json:"activated,omitempty"`
	Status          any       `json:"status,omitempty"`
	SystemStatus    string    `json:"systemStatus,omitempty"`
	ReadinessPeriod any       `json:"readinessPeriod,omitempty"`
	Type            struct {
		ID   int    `json:"id,omitempty"`
		Name string `json:"name,omitempty"`
		Code string `json:"code,omitempty"`
	} `json:"type,omitempty"`
	Profile struct {
		ID   int    `json:"id,omitempty"`
		Name string `json:"name,omitempty"`
	} `json:"profile,omitempty"`
}
