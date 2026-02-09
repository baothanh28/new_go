package handler

import (
	"errors"
	"net/http"
	"strconv"
	"github.com/labstack/echo/v4"
	"myapp/internal/service/product/model"
	"myapp/internal/service/product/service"
)

// ProductTestOnlyHandler handles product test only HTTP requests
type ProductTestOnlyHandler struct {
	service *service.ProductTestOnlyService
}

// NewProductTestOnlyHandler creates a new product test only handler
func NewProductTestOnlyHandler(service *service.ProductTestOnlyService) *ProductTestOnlyHandler {
	return &ProductTestOnlyHandler{service: service}
}

// CreateProductTestOnly handles product test only creation
// POST /api/product-test-only
func (h *ProductTestOnlyHandler) CreateProductTestOnly(c echo.Context) error {
	var req model.CreateProductTestOnlyRequest
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

	entity, err := h.service.CreateProductTestOnly(c.Request().Context(), &req)
	if err != nil {
		if errors.Is(err, service.ErrCodeExists) {
			return c.JSON(http.StatusConflict, map[string]string{
				"error": "Code already exists",
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to create product test only",
		})
	}

	return c.JSON(http.StatusCreated, entity.ToResponse())
}

// GetProductTestOnly handles retrieving a single product test only by ID
// GET /api/product-test-only/:id
func (h *ProductTestOnlyHandler) GetProductTestOnly(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid ID format",
		})
	}

	entity, err := h.service.GetProductTestOnlyByID(c.Request().Context(), uint(id))
	if err != nil {
		if errors.Is(err, service.ErrProductTestOnlyNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "Product test only not found",
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to get product test only",
		})
	}

	return c.JSON(http.StatusOK, entity.ToResponse())
}

// GetProductTestOnlyByCode handles retrieving a product test only by code
// GET /api/product-test-only/code/:code
func (h *ProductTestOnlyHandler) GetProductTestOnlyByCode(c echo.Context) error {
	code := c.Param("code")
	if code == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Code is required",
		})
	}

	entity, err := h.service.GetProductTestOnlyByCode(c.Request().Context(), code)
	if err != nil {
		if errors.Is(err, service.ErrProductTestOnlyNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "Product test only not found",
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to get product test only",
		})
	}

	return c.JSON(http.StatusOK, entity.ToResponse())
}

// GetAllProductTestOnly handles retrieving all product test only records
// GET /api/product-test-only
func (h *ProductTestOnlyHandler) GetAllProductTestOnly(c echo.Context) error {
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	offset, _ := strconv.Atoi(c.QueryParam("offset"))

	entities, err := h.service.GetAllProductTestOnly(c.Request().Context(), limit, offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to get product test only records",
		})
	}

	responses := make([]*model.ProductTestOnlyResponse, len(entities))
	for i, entity := range entities {
		responses[i] = entity.ToResponse()
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"items":  responses,
		"limit":  limit,
		"offset": offset,
	})
}

// GetProductTestOnlyByType handles retrieving product test only records by type
// GET /api/product-test-only/type/:type
func (h *ProductTestOnlyHandler) GetProductTestOnlyByType(c echo.Context) error {
	entityType := c.Param("type")
	if entityType == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Type is required",
		})
	}

	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	offset, _ := strconv.Atoi(c.QueryParam("offset"))

	entities, err := h.service.GetProductTestOnlyByType(c.Request().Context(), entityType, limit, offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to get product test only records by type",
		})
	}

	responses := make([]*model.ProductTestOnlyResponse, len(entities))
	for i, entity := range entities {
		responses[i] = entity.ToResponse()
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"items":  responses,
		"limit":  limit,
		"offset": offset,
	})
}

// SearchProductTestOnly handles searching product test only records by name
// GET /api/product-test-only/search
func (h *ProductTestOnlyHandler) SearchProductTestOnly(c echo.Context) error {
	name := c.QueryParam("name")
	if name == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Name query parameter is required",
		})
	}

	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	offset, _ := strconv.Atoi(c.QueryParam("offset"))

	entities, err := h.service.SearchProductTestOnly(c.Request().Context(), name, limit, offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to search product test only records",
		})
	}

	responses := make([]*model.ProductTestOnlyResponse, len(entities))
	for i, entity := range entities {
		responses[i] = entity.ToResponse()
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"items":  responses,
		"limit":  limit,
		"offset": offset,
	})
}

// UpdateProductTestOnly handles updating a product test only
// PUT /api/product-test-only/:id
func (h *ProductTestOnlyHandler) UpdateProductTestOnly(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid ID format",
		})
	}

	var req model.UpdateProductTestOnlyRequest
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

	entity, err := h.service.UpdateProductTestOnly(c.Request().Context(), uint(id), &req)
	if err != nil {
		if errors.Is(err, service.ErrProductTestOnlyNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "Product test only not found",
			})
		}
		if errors.Is(err, service.ErrCodeExists) {
			return c.JSON(http.StatusConflict, map[string]string{
				"error": "Code already exists",
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to update product test only",
		})
	}

	return c.JSON(http.StatusOK, entity.ToResponse())
}

// DeleteProductTestOnly handles deleting a product test only
// DELETE /api/product-test-only/:id
func (h *ProductTestOnlyHandler) DeleteProductTestOnly(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid ID format",
		})
	}

	err = h.service.DeleteProductTestOnly(c.Request().Context(), uint(id))
	if err != nil {
		if errors.Is(err, service.ErrProductTestOnlyNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "Product test only not found",
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to delete product test only",
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Product test only deleted successfully",
	})
}
