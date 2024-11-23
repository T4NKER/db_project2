package server

import (
	"db_project2/internal/services"
	"db_project2/pkg/database"
	API "db_project2/internal/APIs"
	"log"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func Start() {
	db := database.DatabaseInit()

	router := gin.Default()
	router.Use(cors.New(LoadCorsConfig()))
	services.InitServices(db) 
	API.InitAPI(router)

	err := router.Run(":3000")
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
	log.Println("Server successfully started at http://localhost:3000")
}
