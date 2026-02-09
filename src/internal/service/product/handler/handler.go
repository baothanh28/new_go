package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/base/src/internal/service/product/model"
	"github.com/base/src/internal/service/product/service"
	"github.com/labstack/echo/v4"
)

// Handler handles product HTTP requests
type Handler struct {
	service *service.Service
}

// NewHandler creates a new product handler
func NewHandler(service *service.Service) *Handler {
	return &Handler{
		service: service,
	}
}

// CreateProduct handles product creation
// POST /api/products
func (h *Handler) CreateProduct(c echo.Context) error {
	var req model.CreateProductRequest
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

	product, err := h.service.CreateProduct(c.Request().Context(), &req)
	if err != nil {
		if errors.Is(err, service.ErrSKUExists) {
			return c.JSON(http.StatusConflict, map[string]string{
				"error": err.Error(),
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to create product",
		})
	}

	return c.JSON(http.StatusCreated, product.ToResponse())
}

// GetProduct handles retrieving a product by ID
// GET /api/products/:id
func (h *Handler) GetProduct(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid product ID",
		})
	}

	product, err := h.service.GetProductByID(c.Request().Context(), uint(id))
	if err != nil {
		if errors.Is(err, service.ErrProductNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "Product not found",
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to get product",
		})
	}

	return c.JSON(http.StatusOK, product.ToResponse())
}

// GetProducts handles retrieving all products
// GET /api/products
func (h *Handler) GetProducts(c echo.Context) error {
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	offset, _ := strconv.Atoi(c.QueryParam("offset"))
	category := c.QueryParam("category")
	search := c.QueryParam("search")
	activeOnly := c.QueryParam("active") == "true"

	if limit <= 0 {
		limit = 20
	}

	var products []*model.Product
	var err error

	if search != "" {
		products, err = h.service.SearchProducts(c.Request().Context(), search, limit, offset)
	} else if category != "" {
		products, err = h.service.GetProductsByCategory(c.Request().Context(), category, limit, offset)
	} else if activeOnly {
		products, err = h.service.GetActiveProducts(c.Request().Context(), limit, offset)
	} else {
		products, err = h.service.GetAllProducts(c.Request().Context(), limit, offset)
	}

	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to get products",
		})
	}

	responses := make([]*model.ProductResponse, len(products))
	for i, product := range products {
		responses[i] = product.ToResponse()
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"products": responses,
		"limit":    limit,
		"offset":   offset,
	})
}

// UpdateProduct handles product update
// PUT /api/products/:id
func (h *Handler) UpdateProduct(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid product ID",
		})
	}

	var req model.UpdateProductRequest
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

	product, err := h.service.UpdateProduct(c.Request().Context(), uint(id), &req)
	if err != nil {
		if errors.Is(err, service.ErrProductNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "Product not found",
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to update product",
		})
	}

	return c.JSON(http.StatusOK, product.ToResponse())
}

// DeleteProduct handles product deletion
// DELETE /api/products/:id
func (h *Handler) DeleteProduct(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid product ID",
		})
	}

	if err := h.service.DeleteProduct(c.Request().Context(), uint(id)); err != nil {
		if errors.Is(err, service.ErrProductNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "Product not found",
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to delete product",
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Product deleted successfully",
	})
}

// UpdateStock handles stock update
// PATCH /api/products/:id/stock
func (h *Handler) UpdateStock(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid product ID",
		})
	}

	var req struct {
		Quantity int `json:"quantity" validate:"required"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
	}

	if err := h.service.UpdateStock(c.Request().Context(), uint(id), req.Quantity); err != nil {
		if errors.Is(err, service.ErrProductNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "Product not found",
			})
		}
		if errors.Is(err, service.ErrInsufficientStock) {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": err.Error(),
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to update stock",
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Stock updated successfully",
	})
}
