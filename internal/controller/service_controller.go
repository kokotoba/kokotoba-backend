package controller

import (
	"net/http"

	"kokotoba-backend/internal/model"
	"kokotoba-backend/internal/view"
)

type ServiceController struct {
	name    string
	version string
}

func NewServiceController(name, version string) *ServiceController {
	return &ServiceController{name: name, version: version}
}

func (c *ServiceController) Show(w http.ResponseWriter, _ *http.Request) {
	_ = view.JSON(w, http.StatusOK, model.ServiceInfoResponse{
		Name:    c.name,
		Version: c.version,
		Message: "API はここから実装します",
	})
}
