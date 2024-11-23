package database

import (
	"fmt"
	"os"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func DatabaseInit() *gorm.DB {
	// Set up environment variables or replace with actual values
	host := "localhost"
	port := "5432"
	user := "postgres"
	password := "postgres"
	dbName := "my_database3"
	sslMode := "disable"

	// Connection string for the target database
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s", host, port, user, password, dbName, sslMode)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to the PostgreSQL database: %v", err)
	}

	// Assign to global DB variable
	DB = db

	// Execute schema SQL file
	/* schemaPath := "/home/t4nk/db_project2/pkg/database/migrations/sql.sql"
	if err := ExecuteSchemaFile(db, schemaPath); err != nil {
		log.Fatalf("Failed to execute schema file: %v", err)
	} */

	return DB
}

func ExecuteSchemaFile(db *gorm.DB, filePath string) error {
	// Read the schema file
	schema, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("error reading schema file: %w", err)
	}

	// Execute the schema SQL
	if err := db.Exec(string(schema)).Error; err != nil {
		return fmt.Errorf("error executing schema SQL: %w", err)
	}

	log.Println("Schema file executed successfully.")
	return nil
}
