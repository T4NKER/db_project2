package subservices

import(
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
	err := s.db.Table("books").Select("title, author, isbn").Where("available = ?", true).Find(&resources).Error
	return resources, err
}

func (s *StudentService) ChangePassword(studentID int, oldPassword, newPassword string) error {
	var storedPassword string
	err := s.db.Table("students").Select("password").Where("id = ?", studentID).Scan(&storedPassword).Error
	if err != nil || storedPassword != oldPassword {
		return err 
	}

	return s.db.Table("students").Where("id = ?", studentID).Update("password", newPassword).Error
}

func (s *StudentService) GetLoansByStudentID(studentID int) ([]map[string]interface{}, error) {
	var loans []map[string]interface{}
    err := s.db.Table("loans").Select("book_title, loan_date, due_date").Where("student_id =?", studentID).Find(&loans).Error
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
