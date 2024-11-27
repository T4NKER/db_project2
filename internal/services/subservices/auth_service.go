package subservices

import (
	"errors"
	"sync"

	"gorm.io/gorm"
)

type AuthService struct {
	sessions sync.Map
	db       *gorm.DB
}

func NewAuthServiceInstance(database *gorm.DB) *AuthService {
	return &AuthService{db: database, sessions: sync.Map{}}
}

// User model represents the "User" table in the database
type User struct {
	UserID    int    `gorm:"primaryKey"`
	Username  string `gorm:"unique"`
	Password  string
	UserRole  string `gorm:"check:user_role IN ('Admin', 'LibraryAgent', 'Student')"`
	StudentID int   `gorm:"foreignKey:StudentID"`
}

// ValidateCredentials checks the username and password against the database
func (a *AuthService) ValidateCredentials(username, password string) (string, int, error) {
	var user User
	err := a.db.Table("User").Where("username = ? AND password = ?", username, password).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", 0, errors.New("invalid credentials")
		}
		return "",0, errors.New("database error: " + err.Error())
	}
	return user.UserRole, user.StudentID, nil
}

// CreateSession creates a new session token
func (a *AuthService) CreateSession(username, role string) string {
	token := username + "_token"
	a.sessions.Store(token, role)
	return token
}

// InvalidateSession deletes a session token
func (a *AuthService) InvalidateSession(token string) {
	a.sessions.Delete(token)
}

// GetSessionRole retrieves the role associated with a session token
func (a *AuthService) GetSessionRole(token string) (string, bool) {
	role, ok := a.sessions.Load(token)
	if !ok {
		return "", false
	}
	return role.(string), true
}
