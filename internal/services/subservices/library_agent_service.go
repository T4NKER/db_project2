package subservices

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"gorm.io/gorm"
)

type LibraryAgentService struct {
	db *gorm.DB
}

// NewLibraryAgentServiceInstance initializes a new LibraryAgentService instance
func NewLibraryAgentServiceInstance(db *gorm.DB) *LibraryAgentService {
	return &LibraryAgentService{db: db}
}

// GetOverdueLoans retrieves a list of overdue loans along with student and book details
func (l *LibraryAgentService) GetOverdueLoans() ([]map[string]interface{}, error) {
	var overdueLoans []map[string]interface{}

	// Corrected SQL query
	query := `
        SELECT
            l.loan_id,
            l.due_date,
            CONCAT(s.first_name, ' ', s.last_name) AS student_name,
            b.title AS book_title,
            bc.barcode
        FROM loan l
        JOIN student s ON l.student_id = s.student_id
        JOIN book_copy bc ON l.copy_id = bc.copy_id
        JOIN book b ON bc.book_code = b.book_code
        WHERE l.due_date < CURRENT_DATE AND return_date IS NULL
    `

	// Execute the query
	err := l.db.Raw(query).Scan(&overdueLoans).Error
	if err != nil {
		return nil, err
	}

	return overdueLoans, nil
}

// AssignResource assigns a book copy to a user by creating a loan record
func (l *LibraryAgentService) AssignResource(studentID int, bookCode string) error {
	loanDate := time.Now()
	dueDate := loanDate.AddDate(0, 0, 15)

	// Start a database transaction
	tx := l.db.Begin()

	var isCardActivated bool
	err := tx.Table("librarycard").
		Select("status").
		Where("student_id = ?", studentID).
		Scan(&isCardActivated).Error
	if err != nil {
		tx.Rollback()
		log.Printf("Failed to check library card status for student_id %d: %v\n", studentID, err)
		return fmt.Errorf("failed to check library card status: %w", err)
	}

	if !isCardActivated {
		tx.Rollback()
		log.Printf("Library card for student_id %d is not activated\n", studentID)
		return fmt.Errorf("library card for student_id %d is not activated", studentID)
	}

	// Log all available copies for the given book_code
	var availableCopies []int
	err = tx.Table("book_copy").
		Select("copy_id").
		Where("book_code = ? AND is_available = ?", bookCode, true).
		Scan(&availableCopies).Error
	if err != nil {
		tx.Rollback()
		log.Printf("Failed to fetch available copies for book_code %s: %v\n", bookCode, err)
		return err
	}

	var copyID int
	err = tx.Table("book_copy").
		Select("copy_id").
		Where("book_code = ? AND is_available = ?", bookCode, true).
		Limit(1).Scan(&copyID).Error
	if err != nil || copyID == 0 {
		tx.Rollback()
		log.Printf("No available copy found for book_code %s\n", bookCode)
		return fmt.Errorf("no available copy for book_code: %s", bookCode)
	}

	err = tx.Table("loan").Create(map[string]interface{}{
		"student_id": studentID,
		"copy_id":    copyID,
		"loan_date":  loanDate,
		"due_date":   dueDate,
	}).Error
	if err != nil {
		tx.Rollback()
		log.Printf("Failed to create loan for student_id %d and copy_id %d: %v\n", studentID, copyID, err)
		return err
	}

	// Mark the book copy as unavailable
	err = tx.Table("book_copy").Where("copy_id = ?", copyID).Update("is_available", false).Error
	if err != nil {
		tx.Rollback()
		log.Printf("Failed to mark copy_id %d as unavailable: %v\n", copyID, err)
		return err
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		log.Printf("Failed to commit transaction for student_id %d and copy_id %d: %v\n", studentID, copyID, err)
		return err
	}

	log.Printf("Successfully assigned copy_id %d of book_code %s to student_id %d\n", copyID, bookCode, studentID)
	return nil
}

// MarkResourceReturned updates a loan record and the associated book copy status
func (l *LibraryAgentService) MarkResourceReturned(loanID int) error {
	// Start a database transaction
	tx := l.db.Begin()

	var returnDate sql.NullTime
	err := tx.Table("loan").Select("return_date").Where("loan_id = ?", loanID).Scan(&returnDate).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	if returnDate.Valid {
		tx.Rollback()
		return fmt.Errorf("loan_id %d already has a return_date set", loanID)
	}

	// Update the return_date for the loan record
	err = tx.Table("loan").Where("loan_id = ?", loanID).Update("return_date", time.Now()).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	// Fetch the copy_id associated with the loan
	var copyID int
	err = tx.Table("loan").Select("copy_id").Where("loan_id = ?", loanID).Scan(&copyID).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	// Update the book copy to mark it as available
	err = tx.Table("book_copy").Where("copy_id = ?", copyID).Update("is_available", true).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	// Commit the transaction
	return tx.Commit().Error
}

func (l *LibraryAgentService) GetStudentProfile(studentID int) (map[string]interface{}, error) {
	var profile map[string]interface{}

	query := `
        SELECT 
    u.username,
    s.phone,
    s.postal_address,
    COUNT(l.loan_id) AS total_loans
FROM "User" u
LEFT JOIN Student s ON u.student_id = s.student_id -- Joining User with Student to access phone and postal_address
LEFT JOIN Loan l ON s.student_id = l.student_id -- Joining Loan table with Student based on student_id
WHERE u.user_id = ? -- Replace ? with the actual user_id value when executing the query
GROUP BY u.username, s.phone, s.postal_address; 

    `

	err := l.db.Raw(query, studentID).Scan(&profile).Error
	if err != nil {
		return nil, err
	}

	return profile, nil
}
