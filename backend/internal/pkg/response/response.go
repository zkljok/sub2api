// Package response provides standardized HTTP response helpers.
package response

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"log"
	"math"
	"net/http"
	"strings"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/util/logredact"
	"github.com/gin-gonic/gin"
)

const gzipMinResponseBytes = 1024

// Response 标准API响应格式
type Response struct {
	Code     int               `json:"code"`
	Message  string            `json:"message"`
	Reason   string            `json:"reason,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"`
	Data     any               `json:"data,omitempty"`
}

// PaginatedData 分页数据格式（匹配前端期望）
type PaginatedData struct {
	Items    any   `json:"items"`
	Total    int64 `json:"total"`
	Page     int   `json:"page"`
	PageSize int   `json:"page_size"`
	Pages    int   `json:"pages"`
}

// Success 返回成功响应
func Success(c *gin.Context, data any) {
	c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "success",
		Data:    data,
	})
}

// SuccessMaybeGzip returns the standard success envelope and gzips large JSON
// responses when the client supports it. It is intentionally opt-in so streaming
// and gateway responses keep their current behavior.
func SuccessMaybeGzip(c *gin.Context, data any) {
	payload, err := json.Marshal(Response{
		Code:    0,
		Message: "success",
		Data:    data,
	})
	if err != nil {
		Error(c, http.StatusInternalServerError, "failed to serialize response")
		return
	}
	DataMaybeGzip(c, http.StatusOK, "application/json; charset=utf-8", payload)
}

// DataMaybeGzip writes a response body, applying gzip only for large, ordinary
// responses. Range requests are left untouched because compressed byte ranges are
// not equivalent to original body ranges.
func DataMaybeGzip(c *gin.Context, status int, contentType string, payload []byte) {
	if !shouldGzip(c, payload) {
		c.Data(status, contentType, payload)
		return
	}

	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	if _, err := zw.Write(payload); err != nil {
		_ = zw.Close()
		c.Data(status, contentType, payload)
		return
	}
	if err := zw.Close(); err != nil {
		c.Data(status, contentType, payload)
		return
	}

	c.Header("Content-Encoding", "gzip")
	c.Header("Vary", appendVary(c.Writer.Header().Get("Vary"), "Accept-Encoding"))
	c.Data(status, contentType, buf.Bytes())
}

func shouldGzip(c *gin.Context, payload []byte) bool {
	if c == nil || c.Request == nil || len(payload) < gzipMinResponseBytes {
		return false
	}
	if strings.TrimSpace(c.GetHeader("Range")) != "" {
		return false
	}
	return acceptsGzip(c.GetHeader("Accept-Encoding"))
}

func acceptsGzip(header string) bool {
	for _, part := range strings.Split(header, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		segments := strings.Split(part, ";")
		if !strings.EqualFold(strings.TrimSpace(segments[0]), "gzip") {
			continue
		}
		for _, segment := range segments[1:] {
			kv := strings.SplitN(strings.TrimSpace(segment), "=", 2)
			if len(kv) == 2 && strings.EqualFold(kv[0], "q") && strings.TrimSpace(kv[1]) == "0" {
				return false
			}
		}
		return true
	}
	return false
}

func appendVary(existing, value string) string {
	for _, part := range strings.Split(existing, ",") {
		if strings.EqualFold(strings.TrimSpace(part), value) {
			return existing
		}
	}
	if strings.TrimSpace(existing) == "" {
		return value
	}
	return existing + ", " + value
}

// Created 返回创建成功响应
func Created(c *gin.Context, data any) {
	c.JSON(http.StatusCreated, Response{
		Code:    0,
		Message: "success",
		Data:    data,
	})
}

// Accepted 返回异步接受响应 (HTTP 202)
func Accepted(c *gin.Context, data any) {
	c.JSON(http.StatusAccepted, Response{
		Code:    0,
		Message: "accepted",
		Data:    data,
	})
}

// Error 返回错误响应
func Error(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, Response{
		Code:     statusCode,
		Message:  message,
		Reason:   "",
		Metadata: nil,
	})
}

// ErrorWithDetails returns an error response compatible with the existing envelope while
// optionally providing structured error fields (reason/metadata).
func ErrorWithDetails(c *gin.Context, statusCode int, message, reason string, metadata map[string]string) {
	c.JSON(statusCode, Response{
		Code:     statusCode,
		Message:  message,
		Reason:   reason,
		Metadata: metadata,
	})
}

// ErrorFrom converts an ApplicationError (or any error) into the envelope-compatible error response.
// It returns true if an error was written.
func ErrorFrom(c *gin.Context, err error) bool {
	if err == nil {
		return false
	}

	statusCode, status := infraerrors.ToHTTP(err)

	// Log internal errors with full details for debugging
	if statusCode >= 500 && c.Request != nil {
		log.Printf("[ERROR] %s %s\n  Error: %s", c.Request.Method, c.Request.URL.Path, logredact.RedactText(err.Error()))
	}

	ErrorWithDetails(c, statusCode, status.Message, status.Reason, status.Metadata)
	return true
}

// BadRequest 返回400错误
func BadRequest(c *gin.Context, message string) {
	Error(c, http.StatusBadRequest, message)
}

// Unauthorized 返回401错误
func Unauthorized(c *gin.Context, message string) {
	Error(c, http.StatusUnauthorized, message)
}

// Forbidden 返回403错误
func Forbidden(c *gin.Context, message string) {
	Error(c, http.StatusForbidden, message)
}

// NotFound 返回404错误
func NotFound(c *gin.Context, message string) {
	Error(c, http.StatusNotFound, message)
}

// InternalError 返回500错误
func InternalError(c *gin.Context, message string) {
	Error(c, http.StatusInternalServerError, message)
}

// Paginated 返回分页数据
func Paginated(c *gin.Context, items any, total int64, page, pageSize int) {
	pages := int(math.Ceil(float64(total) / float64(pageSize)))
	if pages < 1 {
		pages = 1
	}

	Success(c, PaginatedData{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
		Pages:    pages,
	})
}

// PaginationResult 分页结果（与pagination.PaginationResult兼容）
type PaginationResult struct {
	Total    int64
	Page     int
	PageSize int
	Pages    int
}

// PaginatedWithResult 使用PaginationResult返回分页数据
func PaginatedWithResult(c *gin.Context, items any, pagination *PaginationResult) {
	if pagination == nil {
		Success(c, PaginatedData{
			Items:    items,
			Total:    0,
			Page:     1,
			PageSize: 20,
			Pages:    1,
		})
		return
	}

	Success(c, PaginatedData{
		Items:    items,
		Total:    pagination.Total,
		Page:     pagination.Page,
		PageSize: pagination.PageSize,
		Pages:    pagination.Pages,
	})
}

// ParsePagination 解析分页参数
func ParsePagination(c *gin.Context) (page, pageSize int) {
	page = 1
	pageSize = 20

	if p := c.Query("page"); p != "" {
		if val, err := parseInt(p); err == nil && val > 0 {
			page = val
		}
	}

	// 支持 page_size 和 limit 两种参数名
	if ps := c.Query("page_size"); ps != "" {
		if val, err := parseInt(ps); err == nil && val > 0 && val <= 1000 {
			pageSize = val
		}
	} else if l := c.Query("limit"); l != "" {
		if val, err := parseInt(l); err == nil && val > 0 && val <= 1000 {
			pageSize = val
		}
	}

	return page, pageSize
}

func parseInt(s string) (int, error) {
	var result int
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, nil
		}
		result = result*10 + int(c-'0')
	}
	return result, nil
}
