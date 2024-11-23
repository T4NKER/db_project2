package subservices

import (
	"gorm.io/gorm"
)

type AdministratorService struct {
	db *gorm.DB
}

func NewAdministratorServiceInstance(db *gorm.DB) *AdministratorService {
	return &AdministratorService{db: db}
}

func (a *AdministratorService) AddResource(title, author, isbn, rack string, price float64, purchaseDate string) error {
	return a.db.Table("Book").Create(map[string]interface{}{
		"title":         title,
		"author":        author,
		"isbn":          isbn,
		"rack":          rack,
		"price":         price,
		"purchase_date": purchaseDate,
	}).Error
}

func (a *AdministratorService) ActivateCard(cardID int, status bool) error {
	return a.db.Table("LibraryCard").Where("id = ?", cardID).Update("status", status).Error
}

func (a *AdministratorService) CreateStudentWithCard(name, email, phone, cardType, activationDate string) error {
	tx := a.db.Begin()

	var studentID int
	err := tx.Table("Student").Create(map[string]interface{}{
		"first_name": name,
		"email":      email,
		"phone":      phone,
	}).Scan(&studentID).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	// Insert card
	err = tx.Table("LibraryCard").Create(map[string]interface{}{
		"student_id":      studentID,
		"card_type":       cardType,
		"activation_date": activationDate,
	}).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}
