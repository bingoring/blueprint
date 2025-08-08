package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// StandardResponse 표준 응답 구조체
type StandardResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// ResponseWrapper 미들웨어 - 모든 응답을 표준 구조로 래핑
func ResponseWrapper() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// 원본 writer를 래핑
		writer := &responseWriter{
			ResponseWriter: c.Writer,
			context:        c,
		}
		c.Writer = writer

		// 다음 핸들러 실행
		c.Next()
	})
}

// responseWriter는 gin.ResponseWriter를 래핑하여 응답을 가로챕니다
type responseWriter struct {
	gin.ResponseWriter
	context *gin.Context
}

// 성공 응답 헬퍼 함수들
func Success(c *gin.Context, data interface{}, message string) {
	c.JSON(http.StatusOK, StandardResponse{
		Success: true,
		Data:    data,
		Message: message,
	})
}

func SuccessWithStatus(c *gin.Context, status int, data interface{}, message string) {
	c.JSON(status, StandardResponse{
		Success: true,
		Data:    data,
		Message: message,
	})
}

// 에러 응답 헬퍼 함수들
func Error(c *gin.Context, status int, error string, message string) {
	c.JSON(status, StandardResponse{
		Success: false,
		Error:   error,
		Message: message,
	})
}

func BadRequest(c *gin.Context, error string) {
	Error(c, http.StatusBadRequest, error, "Bad Request")
}

func Unauthorized(c *gin.Context, error string) {
	Error(c, http.StatusUnauthorized, error, "Unauthorized")
}

func InternalServerError(c *gin.Context, error string) {
	Error(c, http.StatusInternalServerError, error, "Internal Server Error")
}

func NotFound(c *gin.Context, error string) {
	Error(c, http.StatusNotFound, error, "Not Found")
}

func Conflict(c *gin.Context, error string) {
	Error(c, http.StatusConflict, error, "Conflict")
}
