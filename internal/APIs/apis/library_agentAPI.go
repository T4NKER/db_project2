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

func NewLibraryAgentHandler(service *subservices.LibraryAgentService) *LibraryAgentHandler {
	return &LibraryAgentHandler{libraryAgentService: service}
}

func InitLibraryAgentAPI(router *gin.Engine, agentService *subservices.LibraryAgentService) {
	handler := NewLibraryAgentHandler(agentService)
	agentRoutes := router.Group("/library-agent")
	{
		agentRoutes.GET("/overdue-loans", handler.ListOverdueLoans)
		agentRoutes.POST("/return-resource", handler.MarkResourceAsReturned)
		agentRoutes.GET("/student-profile/:student_id", handler.ViewStudentProfile)
		agentRoutes.POST("/assign-resource", handler.AssignResource)
	}
}

func (h *LibraryAgentHandler) ListOverdueLoans(c *gin.Context) {
	overdueLoans, err := h.libraryAgentService.GetOverdueLoans()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch overdue loans"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"overdue_loans": overdueLoans})
}

func (h *LibraryAgentHandler) MarkResourceAsReturned(c *gin.Context) {
	var reqData struct {
		LoanID int `form:"loan_id" binding:"required"`
	}

	if err := c.ShouldBind(&reqData); err != nil {
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

func (h *LibraryAgentHandler) ViewStudentProfile(c *gin.Context) {
	studentIDStr := c.Param("student_id")

	studentID, err := strconv.Atoi(studentIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid student ID"})
		return
	}

	studentProfile, err := h.libraryAgentService.GetStudentProfile(studentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch student profile", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"student_profile": studentProfile})
}

func (h *LibraryAgentHandler) AssignResource(c *gin.Context) {
	var reqData struct {
		StudentID           int    `form:"student_id" binding:"required"`
		BookCode         string `form:"book_code" binding:"required"`
	}

	if err := c.ShouldBind(&reqData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid input",
			"details": err.Error(),
		})
		return
	}

	err := h.libraryAgentService.AssignResource(reqData.StudentID, reqData.BookCode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to assign resource",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Resource assigned successfully",
	})
}
