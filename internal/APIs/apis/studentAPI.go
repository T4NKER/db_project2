package apis

import (
	"db_project2/internal/services/subservices"
	"log"
	"net/http"
	"strconv"

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
		studentRoutes.POST("/loans", handler.ListLoans)
		studentRoutes.PATCH("/update-password", handler.UpdatePassword)
	}
}

func (h *StudentHandler) ListAvailableResources(c *gin.Context) {
	resources, err := h.studentService.GetAvailableResources()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch available resources"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"resources": resources})
}

func (h *StudentHandler) ListLoans(c *gin.Context) {
	studentIDStr := c.PostForm("student_id")
	if studentIDStr == "" {
		c.JSON(400, gin.H{"error": "student_id is required"})
		return
	}

	studentID, err := strconv.Atoi(studentIDStr)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid student_id"})
		return
	}

	loans, err := h.studentService.GetLoansByStudentID(studentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch loans"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"loans": loans})
}

func (h *StudentHandler) UpdatePassword(c *gin.Context) {
	var reqData struct {
		OldPassword string `form:"old_password" binding:"required"`
		NewPassword string `form:"new_password" binding:"required"`
		StudentID   string `form:"student_id" binding:"required"`
	}

	if err := c.ShouldBind(&reqData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	studentIDInt, _ := strconv.Atoi(reqData.StudentID)

	log.Println(reqData)

	err := h.studentService.ChangePassword(studentIDInt, reqData.OldPassword, reqData.NewPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to update password",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password updated successfully"})
}
