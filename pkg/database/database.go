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
	host := "localhost"
	port := "5433"
	user := "postgres"
	password := "postgres"
	dbName := "my_database3"
	sslMode := "disable"

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s", host, port, user, password, dbName, sslMode)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to the PostgreSQL database: %v", err)
	}

	DB = db

	log.Println("PRinting database schema...")

	PrintDatabaseSchema()

	// Execute schema SQL file
	/*  schemaPath := "/home/t4nk/db_project2/pkg/database/migrations/sql.sql"
	if err := ExecuteSchemaFile(db, schemaPath); err != nil {
		log.Fatalf("Failed to execute schema file: %v", err)
	}  */

	PrintDatabaseSchema()

	return DB
}

func PrintDatabaseSchema() {
	if DB == nil {
		log.Fatalf("Database is not initialized. Please initialize the database first.")
	}

	// Query for table information
	var tables []struct {
		TableName string
	}
	if err := DB.Raw(`
		SELECT table_name 
		FROM information_schema.tables 
		WHERE table_schema = 'public' AND table_type = 'BASE TABLE';
	`).Scan(&tables).Error; err != nil {
		log.Fatalf("Error fetching tables: %v", err)
	}

	fmt.Println("Tables:")
	for _, table := range tables {
		fmt.Printf("- %s\n", table.TableName)

		var columns []struct {
			ColumnName    string
			DataType      string
			IsNullable    string
			ColumnDefault string
		}
		if err := DB.Raw(`
			SELECT column_name, data_type, is_nullable, column_default
			FROM information_schema.columns
			WHERE table_name = ?;
		`, table.TableName).Scan(&columns).Error; err != nil {
			log.Fatalf("Error fetching columns for table %s: %v", table.TableName, err)
		}

		fmt.Println("  Columns:")
		for _, column := range columns {
			fmt.Printf("    - %s (%s, Nullable: %s, Default: %s)\n",
				column.ColumnName, column.DataType, column.IsNullable, column.ColumnDefault)
		}
	}
}

func ExecuteSchemaFile(db *gorm.DB, filePath string) error {
	schema, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("error reading schema file: %w", err)
	}

	if err := db.Exec(string(schema)).Error; err != nil {
		return fmt.Errorf("error executing schema SQL: %w", err)
	}

	log.Println("Schema file executed successfully.")
	return nil
}
