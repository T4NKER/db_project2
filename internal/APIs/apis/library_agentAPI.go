package apis

import (
	"db_project2/internal/services/subservices"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type LibraryAgentHandler struct {
	libraryAgentService *subservices.LibraryAgentService
}

// NewLibraryAgentHandler initializes a new handler for library agents
func NewLibraryAgentHandler(service *subservices.LibraryAgentService) *LibraryAgentHandler {
	return &LibraryAgentHandler{libraryAgentService: service}
}

// InitLibraryAgentAPI sets up the routes for library agent operations
func InitLibraryAgentAPI(router *gin.Engine, agentService *subservices.LibraryAgentService) {
	handler := NewLibraryAgentHandler(agentService)
	agentRoutes := router.Group("/library-agent")
	{
		agentRoutes.GET("/overdue-loans", handler.ListOverdueLoans)
		agentRoutes.PATCH("/return-resource", handler.MarkResourceAsReturned)
		agentRoutes.GET("/student-profile/:student_id", handler.ViewStudentProfile)
		agentRoutes.POST("/assign-resource", handler.AssignResource)
	}
}

// ListOverdueLoans handles fetching loans that are past their due date
func (h *LibraryAgentHandler) ListOverdueLoans(c *gin.Context) {
	overdueLoans, err := h.libraryAgentService.GetOverdueLoans()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch overdue loans"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"overdue_loans": overdueLoans})
}

// MarkResourceAsReturned handles marking a borrowed resource as returned
func (h *LibraryAgentHandler) MarkResourceAsReturned(c *gin.Context) {
	var reqData struct {
		LoanID int `json:"loan_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&reqData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input", "details": err.Error()})
		return
	}

	err := h.libraryAgentService.MarkResourceReturned(reqData.LoanID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to mark resource as returned", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Resource marked as returned"})
}

// ViewStudentProfile handles viewing the profile of a specific student by ID
func (h *LibraryAgentHandler) ViewStudentProfile(c *gin.Context) {
	// Get student_id as a string from the URL parameter
	studentIDStr := c.Param("student_id")

	// Convert student_id to an integer
	studentID, err := strconv.Atoi(studentIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid student ID"})
		return
	}

	// Call the service method to fetch the profile
	studentProfile, err := h.libraryAgentService.GetStudentProfile(studentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch student profile", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"student_profile": studentProfile})
}

// AssignResource assigns a resource (e.g., a book copy) to a user by creating a loan record
func (h *LibraryAgentHandler) AssignResource(c *gin.Context) {
	var reqData struct {
		UserID           int `json:"user_id" binding:"required"`
		CopyID           int `json:"copy_id" binding:"required"`
		LoanDurationDays int `json:"loan_duration_days" binding:"required"`
	}

	if err := c.ShouldBindJSON(&reqData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input", "details": err.Error()})
		return
	}

	err := h.libraryAgentService.AssignResource(reqData.UserID, reqData.CopyID, reqData.LoanDurationDays)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to assign resource", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Resource assigned successfully"})
}
