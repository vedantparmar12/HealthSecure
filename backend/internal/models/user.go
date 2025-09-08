package models

import (
	"time"

	"gorm.io/gorm"
)

type UserRole string

const (
	RoleDoctor UserRole = "doctor"
	RoleNurse  UserRole = "nurse"
	RoleAdmin  UserRole = "admin"
)

type User struct {
	ID        uint      `json:"id" gorm:"primaryKey;type:bigint unsigned;autoIncrement"`
	Email     string    `json:"email" gorm:"type:varchar(191);unique;not null;index"`
	Password  string    `json:"-" gorm:"type:varchar(255);not null"`
	Role      UserRole  `json:"role" gorm:"not null;type:enum('doctor','nurse','admin')"`
	Name      string    `json:"name" gorm:"type:varchar(100);not null"`
	Active    bool      `json:"active" gorm:"default:true"`
	LastLogin *time.Time `json:"last_login"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	if u.CreatedAt.IsZero() {
		u.CreatedAt = time.Now()
	}
	if u.UpdatedAt.IsZero() {
		u.UpdatedAt = time.Now()
	}
	return
}

func (u *User) BeforeUpdate(tx *gorm.DB) (err error) {
	u.UpdatedAt = time.Now()
	return
}

func (u *User) IsDoctor() bool {
	return u.Role == RoleDoctor
}

func (u *User) IsNurse() bool {
	return u.Role == RoleNurse
}

func (u *User) IsAdmin() bool {
	return u.Role == RoleAdmin
}

func (u *User) CanAccessPatientData() bool {
	return u.Role == RoleDoctor || u.Role == RoleNurse
}

func (u *User) CanAccessSensitiveData() bool {
	return u.Role == RoleDoctor
}

func (u *User) CanManageUsers() bool {
	return u.Role == RoleAdmin
}