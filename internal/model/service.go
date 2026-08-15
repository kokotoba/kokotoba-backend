package model

type HealthResponse struct {
	Status   string `json:"status"`
	Database string `json:"database"`
}

type ServiceInfoResponse struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Message string `json:"message"`
}
