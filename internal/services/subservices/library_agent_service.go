package subservices

import (
	"errors"
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

	query := `
        SELECT
            l.loan_id,
            l.due_date,
            CONCAT(u.first_name, ' ', u.last_name) AS user_name,
            b.title AS book_title,
            bc.barcode
        FROM loan l
        JOIN user u ON l.user_id = u.user_id
        JOIN book_copy bc ON l.copy_id = bc.copy_id
        JOIN book b ON bc.book_code = b.book_code
        WHERE l.due_date < ? AND l.return_date IS NULL
    `

	err := l.db.Raw(query, time.Now()).Scan(&overdueLoans).Error
	if err != nil {
		return nil, err
	}

	return overdueLoans, nil
}

// AssignResource assigns a book copy to a user by creating a loan record
func (l *LibraryAgentService) AssignResource(userID, copyID int, loanDurationDays int) error {
	loanDate := time.Now()
	dueDate := loanDate.AddDate(0, 0, loanDurationDays)

	// Start a transaction to ensure consistency
	tx := l.db.Begin()

	// Ensure the book copy is available
	var status string
	err := tx.Table("book_copy").Select("status").Where("copy_id = ?", copyID).Scan(&status).Error
	if err != nil || status != "available" {
		tx.Rollback()
		return errors.New("book copy is not available")
	}

	// Create a new loan record
	err = tx.Table("loan").Create(map[string]interface{}{
		"user_id":   userID,
		"copy_id":   copyID,
		"loan_date": loanDate,
		"due_date":  dueDate,
	}).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	// Update the book copy status to borrowed
	err = tx.Table("book_copy").Where("copy_id = ?", copyID).Update("status", "borrowed").Error
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

// MarkResourceReturned updates a loan record and the associated book copy status
func (l *LibraryAgentService) MarkResourceReturned(loanID int) error {
	tx := l.db.Begin()

	// Update loan record to set return_date
	err := tx.Table("loans").Where("loan_id = ?", loanID).Update("return_date", time.Now()).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	// Get the copy_id associated with the loan
	var copyID int
	err = tx.Table("loans").Select("copy_id").Where("loan_id = ?", loanID).Scan(&copyID).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	// Update book copy status to available
	err = tx.Table("book_copies").Where("copy_id = ?", copyID).Update("status", "available").Error
	if err != nil {
		tx.Rollback()
		return err
	}

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
