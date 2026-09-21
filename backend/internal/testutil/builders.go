package testutil

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"

	"security-solution/config"
	"security-solution/models"
	"security-solution/services"
)

// MakeUser creates a user with the given role and returns it.
func MakeUser(t *testing.T, role string) *models.User {
	t.Helper()
	u := &models.User{
		Email:     fmt.Sprintf("u-%s@test.local", uuid.NewString()[:8]),
		Phone:     fmt.Sprintf("080%d", time.Now().UnixNano()%100000000),
		FirstName: "Test",
		LastName:  "User",
		Password:  "test-password-123",
		Role:      role,
		Status:    "active",
	}
	// Hash the password via AuthService.Register
	authSvc := services.NewAuthService()
	created, err := authSvc.Register(u)
	if err != nil {
		t.Fatalf("MakeUser: %v", err)
	}
	return created
}

// MakeUnit creates a unit owned by headAdmin (also adds them as head admin).
func MakeUnit(t *testing.T, headAdmin *models.User) *models.SecurityUnit {
	t.Helper()
	unit := &models.SecurityUnit{
		Name:               fmt.Sprintf("Unit-%s", uuid.NewString()[:8]),
		Type:               "community",
		Status:             "active",
		IsVerified:         true,
		RegistrationNumber: fmt.Sprintf("REG-%s", uuid.NewString()[:8]),
	}
	if err := config.DB.Create(unit).Error; err != nil {
		t.Fatalf("MakeUnit: %v", err)
	}
	if headAdmin != nil {
		MakeMembership(t, headAdmin, unit, "unit_admin", true)
	}
	return unit
}

// MakeMembership attaches user to unit with role. isHeadAdmin flags head admin.
func MakeMembership(t *testing.T, user *models.User, unit *models.SecurityUnit, role string, isHeadAdmin bool) *models.UnitMembership {
	t.Helper()
	m := &models.UnitMembership{
		UnitID:      unit.ID,
		UserID:      user.ID,
		Role:        role,
		Status:      models.MembershipActive,
		IsHeadAdmin: isHeadAdmin,
	}
	now := time.Now().UTC()
	m.AcceptedAt = &now
	m.VerifiedAt = &now
	if err := config.DB.Create(m).Error; err != nil {
		t.Fatalf("MakeMembership: %v", err)
	}
	return m
}

// MakeCase creates a case reported by reporter within unit.
func MakeCase(t *testing.T, reporter *models.User, unit *models.SecurityUnit) *models.Case {
	t.Helper()
	c := &models.Case{
		TrackingID:  fmt.Sprintf("CS-TEST-%s", uuid.NewString()[:8]),
		Title:       "Test case",
		Description: "Created by testutil",
		Status:      "pending",
		Priority:    "medium",
		IsPublic:    true,
		UnitID:      unit.ID,
		ReportedBy:  reporter.ID,
	}
	if err := config.DB.Create(c).Error; err != nil {
		t.Fatalf("MakeCase: %v", err)
	}
	return c
}

// AuthHeader returns "Bearer <jwt>" for the user.
func AuthHeader(t *testing.T, user *models.User) string {
	t.Helper()
	authSvc := services.NewAuthService()
	token, err := authSvc.GenerateJWT(user)
	if err != nil {
		t.Fatalf("AuthHeader: %v", err)
	}
	return "Bearer " + token
}
