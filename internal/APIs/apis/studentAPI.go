package apis

import (
	"db_project2/internal/services/subservices"
	"net/http"

	"github.com/gin-gonic/gin"
)

type StudentHandler struct {
	studentService *subservices.StudentService
}

func NewStudentHandler(service *subservices.StudentService) *StudentHandler {
	return &StudentHandler{studentService: service}
}

func InitStudentAPI(router *gin.Engine, studentService *subservices.StudentService) {
	handler := NewStudentHandler(studentService)
	studentRoutes := router.Group("/student")
	{
		studentRoutes.GET("/resources", handler.ListAvailableResources)
		studentRoutes.GET("/loans", handler.ListLoans)
		studentRoutes.PATCH("/update-password", handler.UpdatePassword)
	}
}

// ListAvailableResources handles listing resources available for students.
func (h *StudentHandler) ListAvailableResources(c *gin.Context) {
	resources, err := h.studentService.GetAvailableResources()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch available resources"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"resources": resources})
}

// ListLoans handles listing the current loans for the student.
func (h *StudentHandler) ListLoans(c *gin.Context) {
	// Assume student ID is extracted from authentication middleware
	studentID := c.GetInt("student_id")

	loans, err := h.studentService.GetLoansByStudentID(studentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch loans"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"loans": loans})
}

// UpdatePassword handles updating the password for the student.
func (h *StudentHandler) UpdatePassword(c *gin.Context) {
	var reqData struct {
		OldPassword string `json:"old_password" binding:"required"`
		NewPassword string `json:"new_password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&reqData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	// Assume student ID is extracted from authentication middleware
	studentID := c.GetInt("student_id")

	err := h.studentService.ChangePassword(studentID, reqData.OldPassword, reqData.NewPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update password", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password updated successfully"})
}
