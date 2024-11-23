package API

import (
	apis "db_project2/internal/APIs/apis"
	services "db_project2/internal/services"
	"github.com/gin-gonic/gin"
)

func InitAPI(router *gin.Engine) {
	apis.InitAdministratorAPI(router, services.AdministratorServiceInstance)
	apis.InitLibraryAgentAPI(router, services.LibraryAgentServiceInstance)
	apis.InitStudentAPI(router, services.StudentServiceInstance)
	apis.InitHomeAPI(router)
}