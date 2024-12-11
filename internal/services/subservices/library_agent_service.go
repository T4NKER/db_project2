package subservices

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	"gorm.io/gorm"
)

type LibraryAgentService struct {
	db *gorm.DB
}

func NewLibraryAgentServiceInstance(db *gorm.DB) *LibraryAgentService {
	return &LibraryAgentService{db: db}
}

func (l *LibraryAgentService) GetOverdueLoans() ([]map[string]interface{}, error) {
	var overdueLoans []map[string]interface{}

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

	
	err := l.db.Raw(query).Scan(&overdueLoans).Error
	if err != nil {
		return nil, err
	}

	return overdueLoans, nil
}

func (l *LibraryAgentService) GetAllLoans() ([]map[string]interface{}, error) {
	var loans []map[string]interface{}

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
		WHERE l.return_date is NULL
    `

	err := l.db.Raw(query).Scan(&loans).Error
	if err != nil {
		return nil, err
	}

	return loans, nil
}

func (l *LibraryAgentService) AssignResource(studentID int, bookCode string) error {
	loanDate := time.Now()
	dueDate := loanDate.AddDate(0, 0, 15)

	var cardStatus struct {
		Status *bool 
	}

	tx := l.db.Begin()

	err := tx.Table("librarycard").
		Select("status").
		Where("student_id = ?", studentID).
		Scan(&cardStatus).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		tx.Rollback()
		log.Printf("Failed to check library card status for student_id %d: %v\n", studentID, err)
		return fmt.Errorf("failed to check library card status: %w", err)
	}

	if cardStatus.Status != nil && !*cardStatus.Status {
		tx.Rollback()
		log.Printf("Library card for student_id %d is not activated\n", studentID)
		return fmt.Errorf("library card for student_id %d is not activated", studentID)
	}

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

	err = tx.Table("book_copy").Where("copy_id = ?", copyID).Update("is_available", false).Error
	if err != nil {
		tx.Rollback()
		log.Printf("Failed to mark copy_id %d as unavailable: %v\n", copyID, err)
		return err
	}

	if err := tx.Commit().Error; err != nil {
		log.Printf("Failed to commit transaction for student_id %d and copy_id %d: %v\n", studentID, copyID, err)
		return err
	}

	log.Printf("Successfully assigned copy_id %d of book_code %s to student_id %d\n", copyID, bookCode, studentID)
	return nil
}

func (l *LibraryAgentService) MarkResourceReturned(loanID int) error {

	tx := l.db.Begin()

	var exists bool
	err := tx.Table("loan").Select("COUNT(*) > 0").Where("loan_id = ?", loanID).Find(&exists).Error
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to check loan existence: %w", err)
	}
	if !exists {
		tx.Rollback()
		return fmt.Errorf("loan_id %d does not exist", loanID)
	}

	var returnDate sql.NullTime
	err = tx.Table("loan").Select("return_date").Where("loan_id = ?", loanID).Scan(&returnDate).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	if returnDate.Valid {
		tx.Rollback()
		return fmt.Errorf("loan_id %d already has a return_date set", loanID)
	}

	err = tx.Table("loan").Where("loan_id = ?", loanID).Update("return_date", time.Now()).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	var copyID int
	err = tx.Table("loan").Select("copy_id").Where("loan_id = ?", loanID).Scan(&copyID).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Table("book_copy").Where("copy_id = ?", copyID).Update("is_available", true).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func (l *LibraryAgentService) GetStudentProfile(studentID int) (map[string]interface{}, error) {
	var profile map[string]interface{}
	var exists bool

	err := l.db.Table("student").Select("COUNT(*) > 0").Where("student_id = ?", studentID).Find(&exists).Error
	if err != nil {
		return nil, fmt.Errorf("failed to check student existence: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("student_id %d does not exist", studentID)
	}

	query := `
    SELECT 
    s.first_name, 
    s.last_name, 
    s.email, 
    s.phone, 
    s.postal_address, 
    COUNT(l.loan_id) AS total_loans, 
    COUNT(CASE WHEN l.loan_id IS NOT NULL AND l.return_date IS NULL THEN 1 END) AS active_loans
FROM Student s
LEFT JOIN Loan l ON s.student_id = l.student_id 
WHERE s.student_id = ?
GROUP BY s.first_name, s.last_name, s.email, s.phone, s.postal_address;
    `

	err = l.db.Raw(query, studentID).Scan(&profile).Error
	if err != nil {
		return nil, err
	}

	return profile, nil
}

func (l *LibraryAgentService) GetAllAvailableBooks() ([]map[string]interface{}, error) {
	var availableBooks []map[string]interface{}

	query := `
        SELECT
    b.book_code,
    b.title,
    COUNT(DISTINCT bc.copy_id) AS available_copies,
    STRING_AGG(DISTINCT CONCAT(a.first_name, ' ', a.last_name), ', ') AS authors,
    STRING_AGG(DISTINCT s.name, ', ') AS subjects,
    STRING_AGG(DISTINCT bl.language, ', ') AS languages
FROM book b
LEFT JOIN book_copy bc ON b.book_code = bc.book_code AND bc.is_available = TRUE
LEFT JOIN book_author ba ON b.book_code = ba.book_code
LEFT JOIN author a ON ba.author_id = a.author_id
LEFT JOIN book_subject bs ON b.book_code = bs.book_code
LEFT JOIN "Subject" s ON bs.subject_id = s.subject_id
LEFT JOIN book_language bl ON b.book_code = bl.book_code
GROUP BY b.book_code, b.title
HAVING COUNT(DISTINCT bc.copy_id) > 0
    `

	err := l.db.Raw(query).Scan(&availableBooks).Error
	if err != nil {
		return nil, err
	}

	return availableBooks, nil
}