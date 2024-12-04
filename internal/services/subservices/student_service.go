package subservices

import (
	"fmt"
	"log"

	"gorm.io/gorm"
)

type StudentService struct {
	db *gorm.DB
}

func NewStudentServiceInstance(db *gorm.DB) *StudentService {
	return &StudentService{db: db}
}

func (s *StudentService) GetAvailableResources() ([]map[string]interface{}, error) {
	var resources []map[string]interface{}

	err := s.db.Table("available_copies").
		Select("title, authors, languages, publisher, available_copies").
		Scan(&resources).Error

	if err != nil {
		log.Printf("Error fetching available resources: %v", err)
		return nil, err
	}

	return resources, nil
}

func (s *StudentService) ChangePassword(studentID int, oldPassword, newPassword string) error {
    var storedPassword string

    err := s.db.Table("User").Select("password").Where("student_id = ?", studentID).Scan(&storedPassword).Error
    if err != nil {
        return fmt.Errorf("failed to fetch stored password: %w", err)
    }

    if storedPassword != oldPassword {
        return fmt.Errorf("old password does not match")
    }

    result := s.db.Table("User").Where("student_id = ?", studentID).Update("password", newPassword)
    if result.Error != nil {
        return fmt.Errorf("failed to update password: %w", result.Error)
    }

    if result.RowsAffected == 0 {
        return fmt.Errorf("no rows were updated, check if student_id is valid")
    }

    return nil
}

func (s *StudentService) GetLoansByStudentID(studentID int) ([]map[string]interface{}, error) {
	var loans []map[string]interface{}

	err := s.db.Table("loan").
    Select("book.title AS book_title, loan.loan_date AS loan_date, loan.due_date AS due_date").
    Joins("JOIN book_copy ON loan.copy_id = book_copy.copy_id").
    Joins("JOIN book ON book_copy.book_code = book.book_code").
    Where("loan.student_id = ?", studentID).
    Find(&loans).Error

	return loans, err
}

func (s *StudentService) ViewStudentProfile(studentID int) (map[string]interface{}, error) {
	var profile map[string]interface{}
	err := s.db.Table("users").
		Select("first_name, last_name, email, phone, postal_address").
		Where("user_id = ?", studentID).
		First(&profile).Error
	return profile, err
}
