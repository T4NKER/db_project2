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
		Name           string `json:"name" binding:"required"`
		Email          string `json:"email" binding:"required,email"`
		Phone          string `json:"phone" binding:"required"`
		CardType       string `json:"card_type" binding:"required"`
		ActivationDate string `json:"activation_date" binding:"required"`
	}

	if err := c.ShouldBindJSON(&reqData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.administratorService.CreateStudentWithCard(
		reqData.Name,
		reqData.Email,
		reqData.Phone,
		reqData.CardType,
		reqData.ActivationDate,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create student", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Student created successfully"})
}

func (h *AdminHandler) ActivateCard(c *gin.Context) {
	var reqData struct {
		CardID int  `json:"card_id"`
		Status bool `json:"status"`
	}

	if err := c.ShouldBindJSON(&reqData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	err := h.administratorService.ActivateCard(reqData.CardID, reqData.Status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update card status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Card status updated successfully"})
}

func (h *AdminHandler) AddResource(c *gin.Context) {
	var reqData struct {
		Title        string  `json:"title"`
		Author       string  `json:"author"`
		ISBN         string  `json:"isbn"`
		Rack         string  `json:"rack"`
		Price        float64 `json:"price"`
		PurchaseDate string  `json:"purchase_date"`
	}

	if err := c.ShouldBindJSON(&reqData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	err := h.administratorService.AddResource(
		reqData.Title,
		reqData.Author,
		reqData.ISBN,
		reqData.Rack,
		reqData.Price,
		reqData.PurchaseDate,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add resource"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Resource added successfully"})
}

