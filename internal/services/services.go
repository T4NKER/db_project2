package services

import (
	subservices "db_project2/internal/services/subservices"

	"gorm.io/gorm"
)

var (
    StudentServiceInstance *subservices.StudentService
    LibraryAgentServiceInstance *subservices.LibraryAgentService
    AdministratorServiceInstance *subservices.AdministratorService
	AuthServiceInstance *subservices.AuthService
)

func InitServices(db *gorm.DB) {
	StudentServiceInstance = subservices.NewStudentServiceInstance(db)
	LibraryAgentServiceInstance = subservices.NewLibraryAgentServiceInstance(db)
	AdministratorServiceInstance = subservices.NewAdministratorServiceInstance(db)
	AuthServiceInstance = subservices.NewAuthServiceInstance(db)
} 