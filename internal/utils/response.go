// Package utils provides shared helpers used across handlers and services.
package utils

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response is the canonical JSON envelope returned by every endpoint
//
//	{ "success": true,  "data":  <payload> }
//	{ "success": false, "error": "<message>" }
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Meta    *Meta       `json:"meta,omitempty"`
}

// Meta carries pagination information for list endpoints.
type Meta struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"pageSize"`
	TotalItems int64 `json:"totalItems"`
	TotalPages int   `json:"totalPages"`
}

// OK sends a 200 success response.
func OK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{Success: true, Data: data})
}

// OKList sends a 200 success response with pagination metadata.
func OKList(c *gin.Context, data interface{}, meta *Meta) {
	c.JSON(http.StatusOK, Response{Success: true, Data: data, Meta: meta})
}

// Created sends a 201 success response.
func Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, Response{Success: true, Data: data})
}

// BadRequest sends a 400 error response.
func BadRequest(c *gin.Context, msg string) {
	c.JSON(http.StatusBadRequest, Response{Success: false, Error: msg})
}

// Unauthorized sends a 401 error response.
func Unauthorized(c *gin.Context, msg string) {
	c.JSON(http.StatusUnauthorized, Response{Success: false, Error: msg})
}

// Forbidden sends a 403 error response.
func Forbidden(c *gin.Context) {
	c.JSON(http.StatusForbidden, Response{Success: false, Error: "forbidden"})
}

// NotFound sends a 404 error response.
func NotFound(c *gin.Context, resource string) {
	c.JSON(http.StatusNotFound, Response{Success: false, Error: resource + " not found"})
}

// Conflict sends a 409 error response.
func Conflict(c *gin.Context, msg string) {
	c.JSON(http.StatusConflict, Response{Success: false, Error: msg})
}

// InternalError sends a 500 error response.
// The raw err is NOT exposed to the client to avoid leaking internals.
func InternalError(c *gin.Context) {
	c.JSON(http.StatusInternalServerError, Response{Success: false, Error: "internal server error"})
}

// Pagination parses page / pageSize query params with sane defaults.
func Pagination(c *gin.Context) (page, pageSize, offset int) {
	page = intQuery(c, "page", 1)
	if page < 1 {
		page = 1
	}
	pageSize = intQuery(c, "pageSize", 20)
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset = (page - 1) * pageSize
	return
}

// intQuery parses a query param safely
func intQuery(c *gin.Context, key string, defaultVal int) int {
	v := 0
	if _, err := c.GetQuery(key); err {
		// ignore the parse error; fall back to defalt
		fmt.Sscanf(c.Query(key), "%d", &v)
	}

	if v == 0 {
		return defaultVal
	}

	return v
}

// BuildMeta constructs a Meta struct for list responses.
func BuildMeta(page, pageSize int, total int64) *Meta {
	totalPages := int(total) / pageSize
	if int(total)%pageSize != 0 {
		totalPages++
	}
	return &Meta{
		Page:       page,
		PageSize:   pageSize,
		TotalItems: total,
		TotalPages: totalPages,
	}
}
