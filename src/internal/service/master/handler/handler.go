package handler

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"myapp/internal/service/master/model"
	"myapp/internal/service/master/service"
)

// Handler handles master HTTP requests
type Handler struct {
	service *service.Service
}

// NewHandler creates a new master handler
func NewHandler(service *service.Service) *Handler {
	return &Handler{
		service: service,
	}
}

// CreateMaster handles master record creation
// POST /api/masters
func (h *Handler) CreateMaster(c echo.Context) error {
	var req model.CreateMasterRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
	}

	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
	}

	master, err := h.service.CreateMaster(c.Request().Context(), &req)
	if err != nil {
		if errors.Is(err, service.ErrCodeExists) {
			return c.JSON(http.StatusConflict, map[string]string{
				"error": err.Error(),
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to create master record",
		})
	}

	return c.JSON(http.StatusCreated, master.ToResponse())
}

// GetMaster handles retrieving a master record by ID
// GET /api/masters/:id
func (h *Handler) GetMaster(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid master ID",
		})
	}

	master, err := h.service.GetMasterByID(c.Request().Context(), uint(id))
	if err != nil {
		if errors.Is(err, service.ErrMasterNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "Master record not found",
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to get master record",
		})
	}

	return c.JSON(http.StatusOK, master.ToResponse())
}

// GetMasters handles retrieving all master records
// GET /api/masters
func (h *Handler) GetMasters(c echo.Context) error {
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	offset, _ := strconv.Atoi(c.QueryParam("offset"))
	masterType := c.QueryParam("type")
	search := c.QueryParam("search")
	activeOnly := c.QueryParam("active") == "true"

	if limit <= 0 {
		limit = 20
	}

	var masters []*model.Master
	var err error

	if search != "" {
		masters, err = h.service.SearchMasters(c.Request().Context(), search, limit, offset)
	} else if masterType != "" {
		masters, err = h.service.GetMastersByType(c.Request().Context(), masterType, limit, offset)
	} else if activeOnly {
		masters, err = h.service.GetActiveMasters(c.Request().Context(), limit, offset)
	} else {
		masters, err = h.service.GetAllMasters(c.Request().Context(), limit, offset)
	}

	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to get master records",
		})
	}

	responses := make([]*model.MasterResponse, len(masters))
	for i, master := range masters {
		responses[i] = master.ToResponse()
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"masters": responses,
		"limit":   limit,
		"offset":  offset,
	})
}

// UpdateMaster handles master record update
// PUT /api/masters/:id
func (h *Handler) UpdateMaster(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid master ID",
		})
	}

	var req model.UpdateMasterRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
	}

	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
	}

	master, err := h.service.UpdateMaster(c.Request().Context(), uint(id), &req)
	if err != nil {
		if errors.Is(err, service.ErrMasterNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "Master record not found",
			})
		}
		if errors.Is(err, service.ErrCodeExists) {
			return c.JSON(http.StatusConflict, map[string]string{
				"error": err.Error(),
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to update master record",
		})
	}

	return c.JSON(http.StatusOK, master.ToResponse())
}

// DeleteMaster handles master record deletion
// DELETE /api/masters/:id
func (h *Handler) DeleteMaster(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid master ID",
		})
	}

	if err := h.service.DeleteMaster(c.Request().Context(), uint(id)); err != nil {
		if errors.Is(err, service.ErrMasterNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "Master record not found",
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to delete master record",
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Master record deleted successfully",
	})
}

// Health returns the basic health status of the service
// GET /health
func (h *Handler) Health(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]interface{}{
		"status":  "healthy",
		"service": "master-service",
		"time":    time.Now().UTC(),
	})
}
