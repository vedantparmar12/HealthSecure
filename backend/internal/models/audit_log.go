package models

import (
	"time"

	"gorm.io/gorm"
)

type AuditAction string

const (
	ActionLogin            AuditAction = "LOGIN"
	ActionLogout           AuditAction = "LOGOUT"
	ActionView             AuditAction = "VIEW"
	ActionCreate           AuditAction = "CREATE"
	ActionUpdate           AuditAction = "UPDATE"
	ActionDelete           AuditAction = "DELETE"
	ActionEmergencyRequest AuditAction = "EMERGENCY_REQUEST"
	ActionEmergencyAccess  AuditAction = "EMERGENCY_ACCESS"
	ActionUnauthorized     AuditAction = "UNAUTHORIZED_ACCESS"
)

type AuditLog struct {
	ID           uint        `json:"id" gorm:"primaryKey;type:bigint unsigned;autoIncrement"`
	UserID       *uint       `json:"user_id,omitempty" gorm:"index;type:bigint unsigned"`
	PatientID    *uint       `json:"patient_id,omitempty" gorm:"index;type:bigint unsigned"`
	RecordID     *uint       `json:"record_id,omitempty" gorm:"index;type:bigint unsigned"`
	Action       AuditAction `json:"action" gorm:"not null;index"`
	Resource     string      `json:"resource" gorm:"not null"`
	IPAddress    string      `json:"ip_address" gorm:"not null"`
	UserAgent    string      `json:"user_agent" gorm:"type:text"`
	EmergencyUse bool        `json:"emergency_use" gorm:"default:false;index"`
	Reason       string      `json:"reason,omitempty" gorm:"type:text"`
	Success      bool        `json:"success" gorm:"default:true;index"`
	ErrorMessage string      `json:"error_message,omitempty"`
	Timestamp    time.Time   `json:"timestamp" gorm:"autoCreateTime;index"`

	User    User     `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Patient *Patient `json:"patient,omitempty" gorm:"foreignKey:PatientID"`
}

func (al *AuditLog) BeforeCreate(tx *gorm.DB) (err error) {
	if al.Timestamp.IsZero() {
		al.Timestamp = time.Now()
	}
	return
}

func (al *AuditLog) IsSecurityEvent() bool {
	return al.Action == ActionUnauthorized || 
		   al.Action == ActionEmergencyAccess || 
		   !al.Success
}

func (al *AuditLog) IsEmergencyAccess() bool {
	return al.EmergencyUse || al.Action == ActionEmergencyAccess || al.Action == ActionEmergencyRequest
}

func (al *AuditLog) IsPatientDataAccess() bool {
	return al.PatientID != nil && (al.Action == ActionView || al.Action == ActionUpdate)
}

func (al *AuditLog) RequiresNotification() bool {
	return al.IsSecurityEvent() || (al.IsEmergencyAccess() && al.Success)
}

func (al *AuditLog) GetSeverity() string {
	if !al.Success {
		return "HIGH"
	}
	if al.IsEmergencyAccess() {
		return "MEDIUM"
	}
	if al.Action == ActionUnauthorized {
		return "HIGH"
	}
	return "LOW"
}

func (al *AuditLog) TableName() string {
	return "audit_logs"
}

type AuditLogFilter struct {
	UserID      *uint
	PatientID   *uint
	Action      *AuditAction
	Success     *bool
	Emergency   *bool
	StartTime   *time.Time
	EndTime     *time.Time
	IPAddress   string
	Limit       int
	Offset      int
}

func (f *AuditLogFilter) Apply(db *gorm.DB) *gorm.DB {
	query := db

	if f.UserID != nil {
		query = query.Where("user_id = ?", *f.UserID)
	}
	if f.PatientID != nil {
		query = query.Where("patient_id = ?", *f.PatientID)
	}
	if f.Action != nil {
		query = query.Where("action = ?", *f.Action)
	}
	if f.Success != nil {
		query = query.Where("success = ?", *f.Success)
	}
	if f.Emergency != nil {
		query = query.Where("emergency_use = ?", *f.Emergency)
	}
	if f.StartTime != nil {
		query = query.Where("timestamp >= ?", *f.StartTime)
	}
	if f.EndTime != nil {
		query = query.Where("timestamp <= ?", *f.EndTime)
	}
	if f.IPAddress != "" {
		query = query.Where("ip_address = ?", f.IPAddress)
	}

	query = query.Order("timestamp DESC")

	if f.Limit > 0 {
		query = query.Limit(f.Limit)
	}
	if f.Offset > 0 {
		query = query.Offset(f.Offset)
	}

	return query
}