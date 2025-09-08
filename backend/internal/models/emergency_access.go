package models

import (
	"time"

	"gorm.io/gorm"
)

type EmergencyAccessStatus string

const (
	EmergencyStatusPending  EmergencyAccessStatus = "pending"
	EmergencyStatusActive   EmergencyAccessStatus = "active"
	EmergencyStatusUsed     EmergencyAccessStatus = "used"
	EmergencyStatusExpired  EmergencyAccessStatus = "expired"
	EmergencyStatusRevoked  EmergencyAccessStatus = "revoked"
)

type EmergencyAccess struct {
	ID          uint                  `json:"id" gorm:"primaryKey;type:bigint unsigned;autoIncrement"`
	UserID      uint                  `json:"user_id" gorm:"not null;index;type:bigint unsigned"`
	PatientID   uint                  `json:"patient_id" gorm:"not null;index;type:bigint unsigned"`
	Reason      string                `json:"reason" gorm:"not null;type:text"`
	AccessToken string                `json:"access_token" gorm:"unique;not null"`
	Status      EmergencyAccessStatus `json:"status" gorm:"default:'pending';index"`
	ExpiresAt   time.Time             `json:"expires_at" gorm:"not null;index"`
	UsedAt      *time.Time            `json:"used_at,omitempty"`
	RevokedAt   *time.Time            `json:"revoked_at,omitempty"`
	RevokedBy   *uint                 `json:"revoked_by,omitempty" gorm:"type:bigint unsigned"`
	CreatedAt   time.Time             `json:"created_at"`

	User      User  `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Patient   Patient `json:"patient,omitempty" gorm:"foreignKey:PatientID"`
	RevokedByUser *User `json:"revoked_by_user,omitempty" gorm:"foreignKey:RevokedBy"`
}

func (ea *EmergencyAccess) BeforeCreate(tx *gorm.DB) (err error) {
	if ea.CreatedAt.IsZero() {
		ea.CreatedAt = time.Now()
	}
	if ea.Status == "" {
		ea.Status = EmergencyStatusPending
	}
	return
}

func (ea *EmergencyAccess) IsActive() bool {
	now := time.Now()
	return ea.Status == EmergencyStatusActive && 
		   ea.ExpiresAt.After(now) && 
		   ea.RevokedAt == nil
}

func (ea *EmergencyAccess) IsExpired() bool {
	now := time.Now()
	return ea.ExpiresAt.Before(now)
}

func (ea *EmergencyAccess) IsRevoked() bool {
	return ea.RevokedAt != nil
}

func (ea *EmergencyAccess) IsUsed() bool {
	return ea.UsedAt != nil
}

func (ea *EmergencyAccess) CanBeActivated() bool {
	now := time.Now()
	return ea.Status == EmergencyStatusPending && 
		   ea.ExpiresAt.After(now) && 
		   ea.RevokedAt == nil
}

func (ea *EmergencyAccess) Activate() error {
	if !ea.CanBeActivated() {
		return gorm.ErrInvalidValue
	}
	
	now := time.Now()
	ea.Status = EmergencyStatusActive
	ea.UsedAt = &now
	return nil
}

func (ea *EmergencyAccess) Revoke(revokedByUserID uint) error {
	if ea.IsRevoked() {
		return gorm.ErrInvalidValue
	}
	
	now := time.Now()
	ea.Status = EmergencyStatusRevoked
	ea.RevokedAt = &now
	ea.RevokedBy = &revokedByUserID
	return nil
}

func (ea *EmergencyAccess) GetRemainingTime() time.Duration {
	now := time.Now()
	if ea.ExpiresAt.Before(now) {
		return 0
	}
	return ea.ExpiresAt.Sub(now)
}

func (ea *EmergencyAccess) UpdateStatus() {
	if ea.IsRevoked() {
		ea.Status = EmergencyStatusRevoked
		return
	}
	
	if ea.IsExpired() {
		ea.Status = EmergencyStatusExpired
		return
	}
	
	if ea.IsUsed() {
		ea.Status = EmergencyStatusUsed
		return
	}
	
	if ea.CanBeActivated() {
		ea.Status = EmergencyStatusPending
		return
	}
	
	if ea.IsActive() {
		ea.Status = EmergencyStatusActive
		return
	}
}

func (ea *EmergencyAccess) TableName() string {
	return "emergency_access"
}

type EmergencyAccessFilter struct {
	UserID    *uint
	PatientID *uint
	Status    *EmergencyAccessStatus
	Active    *bool
	StartTime *time.Time
	EndTime   *time.Time
	Limit     int
	Offset    int
}

func (f *EmergencyAccessFilter) Apply(db *gorm.DB) *gorm.DB {
	query := db

	if f.UserID != nil {
		query = query.Where("user_id = ?", *f.UserID)
	}
	if f.PatientID != nil {
		query = query.Where("patient_id = ?", *f.PatientID)
	}
	if f.Status != nil {
		query = query.Where("status = ?", *f.Status)
	}
	if f.Active != nil && *f.Active {
		now := time.Now()
		query = query.Where("status = ? AND expires_at > ? AND revoked_at IS NULL", 
			EmergencyStatusActive, now)
	}
	if f.StartTime != nil {
		query = query.Where("created_at >= ?", *f.StartTime)
	}
	if f.EndTime != nil {
		query = query.Where("created_at <= ?", *f.EndTime)
	}

	query = query.Order("created_at DESC")

	if f.Limit > 0 {
		query = query.Limit(f.Limit)
	}
	if f.Offset > 0 {
		query = query.Offset(f.Offset)
	}

	return query
}