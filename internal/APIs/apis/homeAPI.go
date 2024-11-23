package apis

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type HomeHandler struct{}

func NewHomeHandler() *HomeHandler {
	return &HomeHandler{}
}

func InitHomeAPI(router *gin.Engine) {
	handler := NewHomeHandler()
	router.GET("/", handler.Home)
}

func (h *HomeHandler) Home(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Welcome to the Library Management System API",
		"status":  "Running",
		"version": "1.0.0",
	})
}
