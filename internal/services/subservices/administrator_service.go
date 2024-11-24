package subservices

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

type AdministratorService struct {
	db *gorm.DB
}

func NewAdministratorServiceInstance(db *gorm.DB) *AdministratorService {
	return &AdministratorService{db: db}
}

func (a *AdministratorService) AddResource(bookCode, rack, barcode string, price float64, purchaseDate string) error {
	if price < 0 {
		return fmt.Errorf("price cannot be negative")
	}

	// Insert into the "Book_copy" table
	err := a.db.Table("book_copy").Create(map[string]interface{}{
		"book_code":     bookCode,     
		"rack_number":   rack,        
		"barcode":       barcode,      
		"price":         price,        
		"purchase_date": purchaseDate, 
		"is_available":  true,         
	}).Error
	if err != nil {
		return fmt.Errorf("failed to add resource: %w", err)
	}

	return nil
}

func (a *AdministratorService) ActivateCard(studentID int) error {
	var currentStatus bool

	err := a.db.Table("librarycard").
		Select("status").
		Where("student_id = ?", studentID).
		Scan(&currentStatus).Error
	if err != nil {
		return fmt.Errorf("failed to fetch card status for student_id %d: %w", studentID, err)
	}

	newStatus := !currentStatus

	result := a.db.Table("librarycard").
		Where("student_id = ?", studentID).
		Update("status", newStatus)
	if result.Error != nil {
		return fmt.Errorf("failed to update card status for student_id %d: %w", studentID, result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("no library card found for student_id %d", studentID)
	}

	return nil
}

func (a *AdministratorService) CreateStudentWithCard(firstname, lastname, email, phone, postalAddress string) error {
	tx := a.db.Begin()

	// Insert student and get the student ID
	var studentID int
	err := tx.Table("student").Create(map[string]interface{}{
		"first_name":     firstname,
		"last_name":      lastname,
		"email":          email,
		"phone":          phone,
		"postal_address": postalAddress,
	}).Error
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to create student: %w", err)
	}

	// Fetch the student_id using email (assuming it's unique)
	err = tx.Table("student").Select("student_id").Where("email = ?", email).Scan(&studentID).Error
	if err != nil || studentID == 0 {
		tx.Rollback()
		return fmt.Errorf("failed to fetch student_id: %w", err)
	}

	// Insert library card for the student
	err = tx.Table("librarycard").Create(map[string]interface{}{
		"student_id":      studentID,
		"activation_date": time.Now(),
		"status":          true, // Default to active
		"resource":        "General", // Default resource type
	}).Error
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to create library card: %w", err)
	}

	// Create a user for the student
	username := fmt.Sprintf("%s.%s", firstname, lastname)
	password := "default_password"
	err = tx.Table("User").Create(map[string]interface{}{
		"username":   username,
		"password":   password,
		"user_role":  "Student",
		"student_id": studentID,
	}).Error
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to create user: %w", err)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
