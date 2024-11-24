package apis

import (
	"db_project2/internal/services/subservices"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AdminHandler struct {
	administratorService *subservices.AdministratorService
}

func NewAdminHandler(service *subservices.AdministratorService) *AdminHandler {
	return &AdminHandler{administratorService: service}
}

func InitAdministratorAPI(router *gin.Engine, adminService *subservices.AdministratorService) {
	handler := NewAdminHandler(adminService)
	adminRoutes := router.Group("/admin")
	{
		adminRoutes.POST("/create-student", handler.CreateStudent)
		adminRoutes.PATCH("/activate-card", handler.ActivateCard)
		adminRoutes.POST("/add-resource", handler.AddResource)
	}
}

func (h *AdminHandler) CreateStudent(c *gin.Context) {
	var reqData struct {
		FirstName     string `form:"first_name" binding:"required"`
		LastName      string `form:"last_name" binding:"required"`
		Email         string `form:"email" binding:"required,email"`
		Phone         string `form:"phone" binding:"required"`
		PostalAddress string `form:"postal_address" binding:"required"`
	}

	// Parse form data
	if err := c.ShouldBind(&reqData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Call the service
	err := h.administratorService.CreateStudentWithCard(
		reqData.FirstName,
		reqData.LastName,
		reqData.Email,
		reqData.Phone,
		reqData.PostalAddress,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create student", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Student created successfully"})
}

func (h *AdminHandler) ActivateCard(c *gin.Context) {
	var reqData struct {
		StudentID int `form:"student_id" binding:"required"`
	}

	// Parse form data
	if err := c.ShouldBind(&reqData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input", "details": err.Error()})
		return
	}

	// Call the service to toggle the card status
	err := h.administratorService.ActivateCard(reqData.StudentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update card status", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Card status updated successfully"})
}

func (h *AdminHandler) AddResource(c *gin.Context) {
	var reqData struct {
		BookCode     string  `form:"book_code" binding:"required"`   
		Rack         string  `form:"rack" binding:"required"`       
		Barcode      string  `form:"barcode" binding:"required"`    
		Price        float64 `form:"price" binding:"required"`      
		PurchaseDate string  `form:"purchase_date" binding:"required"` 
	}

	// Parse form data
	if err := c.ShouldBind(&reqData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input", "details": err.Error()})
		return
	}

	// Call the service
	err := h.administratorService.AddResource(
		reqData.BookCode,
		reqData.Rack,
		reqData.Barcode,
		reqData.Price,
		reqData.PurchaseDate,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add resource", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Resource added successfully"})
}
