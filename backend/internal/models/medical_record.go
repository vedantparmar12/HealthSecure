package models

import (
	"time"

	"gorm.io/gorm"
)

type SeverityLevel string

const (
	SeverityLow      SeverityLevel = "low"
	SeverityMedium   SeverityLevel = "medium"
	SeverityHigh     SeverityLevel = "high"
	SeverityCritical SeverityLevel = "critical"
)

type MedicalRecord struct {
	ID          uint          `json:"id" gorm:"primaryKey;type:bigint unsigned;autoIncrement"`
	PatientID   uint          `json:"patient_id" gorm:"not null;index;type:bigint unsigned"`
	Patient     Patient       `json:"patient,omitempty" gorm:"foreignKey:PatientID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	DoctorID    uint          `json:"doctor_id" gorm:"not null;index;type:bigint unsigned"`
	Doctor      User          `json:"doctor,omitempty" gorm:"foreignKey:DoctorID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Diagnosis   string        `json:"diagnosis" gorm:"type:text"`
	Treatment   string        `json:"treatment" gorm:"type:text"`
	Notes       string        `json:"notes" gorm:"type:text"`
	Medications string        `json:"medications" gorm:"type:text"`
	Severity    SeverityLevel `json:"severity" gorm:"type:varchar(20)"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
}

func (mr *MedicalRecord) BeforeCreate(tx *gorm.DB) (err error) {
	if mr.CreatedAt.IsZero() {
		mr.CreatedAt = time.Now()
	}
	if mr.UpdatedAt.IsZero() {
		mr.UpdatedAt = time.Now()
	}
	return
}

func (mr *MedicalRecord) BeforeUpdate(tx *gorm.DB) (err error) {
	mr.UpdatedAt = time.Now()
	return
}

func (mr *MedicalRecord) IsCritical() bool {
	return mr.Severity == SeverityCritical
}

func (mr *MedicalRecord) IsHighSeverity() bool {
	return mr.Severity == SeverityHigh || mr.Severity == SeverityCritical
}

func (mr *MedicalRecord) CanBeAccessedByRole(role UserRole, userID uint) bool {
	switch role {
	case RoleDoctor:
		return true
	case RoleNurse:
		return mr.Severity != SeverityCritical
	case RoleAdmin:
		return false
	default:
		return false
	}
}

func (mr *MedicalRecord) SanitizeForRole(role UserRole) *MedicalRecord {
	sanitized := *mr

	switch role {
	case RoleNurse:
		if mr.Severity == SeverityCritical {
			sanitized.Diagnosis = "[RESTRICTED - Doctor Only]"
			sanitized.Treatment = "[RESTRICTED - Doctor Only]"
			sanitized.Medications = "[RESTRICTED - Doctor Only]"
		}
	case RoleAdmin:
		return nil
	}

	return &sanitized
}

func (mr *MedicalRecord) TableName() string {
	return "medical_records"
}