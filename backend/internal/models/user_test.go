package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUserRole(t *testing.T) {
	t.Run("RoleConstants", func(t *testing.T) {
		assert.Equal(t, UserRole("doctor"), RoleDoctor)
		assert.Equal(t, UserRole("nurse"), RoleNurse)
		assert.Equal(t, UserRole("admin"), RoleAdmin)
	})
}

func TestUserMethods(t *testing.T) {
	t.Run("IsDoctor", func(t *testing.T) {
		doctorUser := &User{Role: RoleDoctor}
		nurseUser := &User{Role: RoleNurse}
		adminUser := &User{Role: RoleAdmin}

		assert.True(t, doctorUser.IsDoctor())
		assert.False(t, nurseUser.IsDoctor())
		assert.False(t, adminUser.IsDoctor())
	})

	t.Run("IsNurse", func(t *testing.T) {
		doctorUser := &User{Role: RoleDoctor}
		nurseUser := &User{Role: RoleNurse}
		adminUser := &User{Role: RoleAdmin}

		assert.False(t, doctorUser.IsNurse())
		assert.True(t, nurseUser.IsNurse())
		assert.False(t, adminUser.IsNurse())
	})

	t.Run("IsAdmin", func(t *testing.T) {
		doctorUser := &User{Role: RoleDoctor}
		nurseUser := &User{Role: RoleNurse}
		adminUser := &User{Role: RoleAdmin}

		assert.False(t, doctorUser.IsAdmin())
		assert.False(t, nurseUser.IsAdmin())
		assert.True(t, adminUser.IsAdmin())
	})

	t.Run("CanAccessPatients", func(t *testing.T) {
		doctorUser := &User{Role: RoleDoctor}
		nurseUser := &User{Role: RoleNurse}
		adminUser := &User{Role: RoleAdmin}

		assert.True(t, doctorUser.CanAccessPatients())
		assert.True(t, nurseUser.CanAccessPatients())
		assert.False(t, adminUser.CanAccessPatients())
	})

	t.Run("CanViewSensitiveData", func(t *testing.T) {
		doctorUser := &User{Role: RoleDoctor}
		nurseUser := &User{Role: RoleNurse}
		adminUser := &User{Role: RoleAdmin}

		assert.True(t, doctorUser.CanViewSensitiveData())
		assert.False(t, nurseUser.CanViewSensitiveData())
		assert.False(t, adminUser.CanViewSensitiveData())
	})

	t.Run("CanModifyPatients", func(t *testing.T) {
		doctorUser := &User{Role: RoleDoctor}
		nurseUser := &User{Role: RoleNurse}
		adminUser := &User{Role: RoleAdmin}

		assert.True(t, doctorUser.CanModifyPatients())
		assert.True(t, nurseUser.CanModifyPatients())
		assert.False(t, adminUser.CanModifyPatients())
	})

	t.Run("CanCreatePatients", func(t *testing.T) {
		doctorUser := &User{Role: RoleDoctor}
		nurseUser := &User{Role: RoleNurse}
		adminUser := &User{Role: RoleAdmin}

		assert.True(t, doctorUser.CanCreatePatients())
		assert.False(t, nurseUser.CanCreatePatients())
		assert.False(t, adminUser.CanCreatePatients())
	})

	t.Run("CanViewAuditLogs", func(t *testing.T) {
		doctorUser := &User{Role: RoleDoctor}
		nurseUser := &User{Role: RoleNurse}
		adminUser := &User{Role: RoleAdmin}

		assert.False(t, doctorUser.CanViewAuditLogs())
		assert.False(t, nurseUser.CanViewAuditLogs())
		assert.True(t, adminUser.CanViewAuditLogs())
	})

	t.Run("CanManageUsers", func(t *testing.T) {
		doctorUser := &User{Role: RoleDoctor}
		nurseUser := &User{Role: RoleNurse}
		adminUser := &User{Role: RoleAdmin}

		assert.False(t, doctorUser.CanManageUsers())
		assert.False(t, nurseUser.CanManageUsers())
		assert.True(t, adminUser.CanManageUsers())
	})
}

func TestUserTableName(t *testing.T) {
	user := &User{}
	assert.Equal(t, "users", user.TableName())
}