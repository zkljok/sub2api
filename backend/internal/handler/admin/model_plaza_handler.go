package admin

import (
	"database/sql"
	"strconv"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

type ModelPlazaHandler struct{ service *service.ModelPlazaService }

func NewModelPlazaHandler(service *service.ModelPlazaService) *ModelPlazaHandler {
	return &ModelPlazaHandler{service: service}
}

type modelPlazaVendorRequest struct {
	Name        string `json:"name" binding:"required,max=100"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
	Status      string `json:"status" binding:"omitempty,oneof=active disabled"`
	SortOrder   int    `json:"sort_order"`
}

type modelPlazaModelRequest struct {
	ModelName       string                            `json:"model_name" binding:"required,max=255"`
	DisplayName     string                            `json:"display_name" binding:"max=255"`
	Description     string                            `json:"description"`
	Icon            string                            `json:"icon"`
	Tags            []string                          `json:"tags"`
	VendorID        *int64                            `json:"vendor_id"`
	Endpoints       []string                          `json:"endpoints"`
	Status          string                            `json:"status" binding:"omitempty,oneof=active disabled"`
	NameRule        service.ModelPlazaNameRule        `json:"name_rule" binding:"omitempty,oneof=exact prefix suffix contains"`
	SortOrder       int                               `json:"sort_order"`
	PricingOverride service.ModelPlazaPricingOverride `json:"pricing_override"`
	AutoSynced      bool                              `json:"auto_synced"`
}

type modelPlazaBatchRequest struct {
	IDs       []int64                       `json:"ids" binding:"required"`
	Action    service.ModelPlazaBatchAction `json:"action" binding:"required"`
	VendorID  *int64                        `json:"vendor_id"`
	Tags      []string                      `json:"tags"`
	Endpoints []string                      `json:"endpoints"`
}

type modelPlazaBillingSettingsRequest struct {
	Enabled bool `json:"enabled"`
}

func (h *ModelPlazaHandler) ListVendors(c *gin.Context) {
	items, err := h.service.ListVendors(c.Request.Context(), true)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, items)
}

func (h *ModelPlazaHandler) CreateVendor(c *gin.Context) {
	var req modelPlazaVendorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	vendor := &service.ModelPlazaVendor{Name: req.Name, Description: req.Description, Icon: req.Icon, Status: req.Status, SortOrder: req.SortOrder}
	if err := h.service.CreateVendor(c.Request.Context(), vendor); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Created(c, vendor)
}

func (h *ModelPlazaHandler) UpdateVendor(c *gin.Context) {
	id, ok := modelPlazaID(c)
	if !ok {
		return
	}
	var req modelPlazaVendorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	vendor := &service.ModelPlazaVendor{ID: id, Name: req.Name, Description: req.Description, Icon: req.Icon, Status: req.Status, SortOrder: req.SortOrder}
	if err := h.service.UpdateVendor(c.Request.Context(), vendor); err != nil {
		if err == sql.ErrNoRows {
			response.NotFound(c, "Vendor not found")
			return
		}
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, vendor)
}

func (h *ModelPlazaHandler) DeleteVendor(c *gin.Context) {
	id, ok := modelPlazaID(c)
	if !ok {
		return
	}
	if err := h.service.DeleteVendor(c.Request.Context(), id); err != nil {
		if err == sql.ErrNoRows {
			response.NotFound(c, "Vendor not found")
			return
		}
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"message": "Vendor deleted"})
}

func (h *ModelPlazaHandler) ListModels(c *gin.Context) {
	items, err := h.service.ListModelsForAdmin(c.Request.Context(), true)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, items)
}

func (h *ModelPlazaHandler) GetBillingSettings(c *gin.Context) {
	settings, err := h.service.GetBillingSettings(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, settings)
}

func (h *ModelPlazaHandler) UpdateBillingSettings(c *gin.Context) {
	var req modelPlazaBillingSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	settings, err := h.service.UpdateBillingSettings(c.Request.Context(), service.ModelPlazaBillingSettings{Enabled: req.Enabled})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, settings)
}

func (h *ModelPlazaHandler) CreateModel(c *gin.Context) {
	var req modelPlazaModelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	model := modelPlazaRequestToService(req)
	if err := h.service.CreateModel(c.Request.Context(), &model); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Created(c, model)
}

func (h *ModelPlazaHandler) UpdateModel(c *gin.Context) {
	id, ok := modelPlazaID(c)
	if !ok {
		return
	}
	var req modelPlazaModelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	model := modelPlazaRequestToService(req)
	model.ID = id
	if err := h.service.UpdateModel(c.Request.Context(), &model); err != nil {
		if err == sql.ErrNoRows {
			response.NotFound(c, "Model not found")
			return
		}
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, model)
}

func (h *ModelPlazaHandler) DeleteModel(c *gin.Context) {
	id, ok := modelPlazaID(c)
	if !ok {
		return
	}
	if err := h.service.DeleteModel(c.Request.Context(), id); err != nil {
		if err == sql.ErrNoRows {
			response.NotFound(c, "Model not found")
			return
		}
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"message": "Model deleted"})
}

func (h *ModelPlazaHandler) BatchModels(c *gin.Context) {
	var req modelPlazaBatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	changed, err := h.service.BatchUpdateModels(c.Request.Context(), service.ModelPlazaBatchUpdate{
		IDs:       req.IDs,
		Action:    req.Action,
		VendorID:  req.VendorID,
		Tags:      req.Tags,
		Endpoints: req.Endpoints,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			response.NotFound(c, "Model not found")
			return
		}
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"updated": changed})
}

func (h *ModelPlazaHandler) Sync(c *gin.Context) {
	inserted, err := h.service.SyncFromChannels(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"inserted": inserted})
}

func modelPlazaID(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(strings.TrimSpace(c.Param("id")), 10, 64)
	if err != nil || id <= 0 {
		response.BadRequest(c, "Invalid id")
		return 0, false
	}
	return id, true
}

func modelPlazaRequestToService(req modelPlazaModelRequest) service.ModelPlazaModel {
	return service.ModelPlazaModel{
		ModelName: req.ModelName, DisplayName: req.DisplayName, Description: req.Description,
		Icon: req.Icon, Tags: req.Tags, VendorID: req.VendorID, Endpoints: req.Endpoints,
		Status: req.Status, NameRule: req.NameRule, SortOrder: req.SortOrder,
		PricingOverride: req.PricingOverride, AutoSynced: req.AutoSynced,
	}
}
