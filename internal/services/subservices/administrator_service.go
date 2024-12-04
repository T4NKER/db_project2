package subservices

import (
	"fmt"
	"log"
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
	result := a.db.Exec("CALL toggle_card_status(?)", studentID)
	if result.Error != nil {
		return fmt.Errorf("failed to toggle card status for student_id %d: %w", studentID, result.Error)
	}

	log.Printf("Successfully toggled library card status for student_id %d", studentID)
	return nil
}

func (a *AdministratorService) CreateStudentWithCard(firstname, lastname, email, phone, postalAddress string) error {
	tx := a.db.Begin()

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

	err = tx.Table("student").Select("student_id").Where("email = ?", email).Scan(&studentID).Error
	if err != nil || studentID == 0 {
		tx.Rollback()
		return fmt.Errorf("failed to fetch student_id: %w", err)
	}


	var resourceID int
	err = tx.Table("resource").Select("resource_id").Where("resource_type = ?", "Book").Scan(&resourceID).Error
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to fetch resource ID for 'Book': %w", err)
	}

	err = tx.Table("librarycard").Create(map[string]interface{}{
		"student_id":      studentID,
		"activation_date": time.Now(),
		"status":          true,
		"resource_id":     resourceID, 
	}).Error
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to create library card: %w", err)
	}

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

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
