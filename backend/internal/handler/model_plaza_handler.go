package handler

import (
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// ModelPlazaHandler exposes the public, anonymous model plaza snapshot.
type ModelPlazaHandler struct{ service *service.ModelPlazaService }

func NewModelPlazaHandler(service *service.ModelPlazaService) *ModelPlazaHandler {
	return &ModelPlazaHandler{service: service}
}

func (h *ModelPlazaHandler) Get(c *gin.Context) {
	snapshot, err := h.service.Snapshot(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, snapshot)
}
