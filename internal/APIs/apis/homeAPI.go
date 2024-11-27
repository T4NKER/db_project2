package apis

import (
	"log"
	"net/http"
	"strings"

	"db_project2/internal/services/subservices"

	"github.com/gin-gonic/gin"
)

type HomeHandler struct {
	authService *subservices.AuthService
}

func NewHomeHandler(authService *subservices.AuthService) *HomeHandler {
	return &HomeHandler{authService: authService}
}


func InitHomeAPI(router *gin.Engine, authService *subservices.AuthService) {
	handler := NewHomeHandler(authService)
	router.LoadHTMLGlob("templates/*.html")



	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})
	router.POST("/login", handler.Login)
	router.POST("/logout", handler.Logout)



	adminRoutes := router.Group("/admin")
	adminRoutes.Use(handler.AuthMiddleware("admin"))
	router.GET("/admin/dashboard", func(c *gin.Context) {
		c.HTML(http.StatusOK, "admin_dashboard.html", nil)
	})

	agentRoutes := router.Group("/library_agent")
	agentRoutes.Use(handler.AuthMiddleware("library_agent"))
	router.GET("/library_agent/dashboard", func(c *gin.Context) {
		c.HTML(http.StatusOK, "library_agent_dashboard.html", nil)
	})

	studentRoutes := router.Group("/student")
	studentRoutes.Use(handler.AuthMiddleware("student"))
	router.GET("/student/dashboard", func(c *gin.Context) {
		c.HTML(http.StatusOK, "student_dashboard.html", nil)
	})
}

func (h *HomeHandler) Login(c *gin.Context) {
	var loginRequest struct {
		Username string `form:"username" binding:"required"`
		Password string `form:"password" binding:"required"`
	}

	if err := c.ShouldBind(&loginRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	role, studentID, err := h.authService.ValidateCredentials(loginRequest.Username, loginRequest.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	role = strings.ToLower(role)

	token := h.authService.CreateSession(loginRequest.Username, role)

	redirectURL := ""
	switch role {
	case "admin":
		redirectURL = "/admin/dashboard"
	case "libraryagent":
		redirectURL = "/library_agent/dashboard"
	case "student":
		redirectURL = "/student/dashboard"
	}
	c.JSON(http.StatusOK, gin.H{"message": "Login successful", "redirect": redirectURL, "token": token, "studentID": studentID})
}


func (h *HomeHandler) Logout(c *gin.Context) {
	token := c.GetHeader("Authorization")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing token"})
		return
	}

	h.authService.InvalidateSession(token)
	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

func (h *HomeHandler) AuthMiddleware(requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing token"})
			c.Abort()
			return
		}

		log.Println("PRINTING TOKEN: ", token)

		role, ok := h.authService.GetSessionRole(token)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired session"})
			c.Abort()
			return
		}

		log.Println("Requiredrole is ", requiredRole)
		log.Println("ROle is ", role)

		if role != requiredRole && requiredRole != "" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied for your role"})
			c.Abort()
			return
		}

		c.Next()
	}
}
