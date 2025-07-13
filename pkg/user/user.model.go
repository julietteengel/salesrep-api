package user

import (
	"database/sql/driver"
	common2 "github.com/julietteengel/salesrep-api/pkg/common"
	"strings"
	"time"
)

// UserRole - System roles for sales coaching
type UserRole string

const (
	SalesRepRole    UserRole = "sales_rep"
	TeamManagerRole UserRole = "team_manager"
	AdminRole       UserRole = "admin"
)

func (ur *UserRole) Scan(value interface{}) error {
	if str, ok := value.(string); ok {
		*ur = UserRole(str)
	}
	return nil
}

func (ur UserRole) Value() (driver.Value, error) {
	if string(ur) == "" {
		return SalesRepRole, nil
	}
	return string(ur), nil
}

type User struct {
	common2.BaseModelV2

	// Personal information
	FirstName *string `json:"first_name"`
	LastName  *string `json:"last_name"`
	Email     string  `gorm:"uniqueIndex;not null" validate:"required" json:"email"`
	Phone     *string `validate:"omitempty,e164" json:"phone,omitempty"`
	Avatar    *string `json:"avatar,omitempty"`
	Language  string  `gorm:"default:'en'" json:"language"`

	// Hierarchy and roles
	Role        UserRole `gorm:"type:varchar(50);not null;default:'sales_rep'" json:"role"`
	ManagerID   *uint    `json:"manager_id,omitempty"`
	Manager     *User    `gorm:"foreignKey:ManagerID" json:"manager,omitempty"`
	TeamMembers []User   `gorm:"foreignKey:ManagerID" json:"team_members,omitempty"`

	// Status and onboarding
	EmailVerified bool `gorm:"default:false" json:"email_verified"`

	// Sales-specific fields
	SalesTarget       *float64 `json:"sales_target,omitempty"`
	CallsTarget       *int     `json:"calls_target,omitempty"`
	PreferredCoaching *string  `json:"preferred_coaching,omitempty"`

	// Integrations
	CRMIntegration      *string `json:"crm_integration,omitempty"` // salesforce, hubspot, etc.
	CRMExternalID       *string `json:"crm_external_id,omitempty"`
	CalendarIntegration *string `json:"calendar_integration,omitempty"` // google, outlook, etc.

	// Notification preferences
	NotificationsEnabled     bool `gorm:"default:true" json:"notifications_enabled"`
	CoachingNotifications    bool `gorm:"default:true" json:"coaching_notifications"`
	PerformanceNotifications bool `gorm:"default:true" json:"performance_notifications"`

	// Activity tracking
	LastLogin    *time.Time `json:"last_login,omitempty"`
	LastCallDate *time.Time `json:"last_call_date,omitempty"`
	TotalCalls   int        `gorm:"default:0" json:"total_calls"`
	TotalRevenue float64    `gorm:"default:0" json:"total_revenue"`

	// Archiving
	ArchivedAt *time.Time `json:"archived_at,omitempty"`
}

// User helper methods
func (u *User) IsManager() bool {
	return u.Role == TeamManagerRole || u.Role == AdminRole
}

func (u *User) IsAdmin() bool {
	return u.Role == AdminRole
}

func (u *User) IsSalesRep() bool {
	return u.Role == SalesRepRole
}

func (u *User) GetFullName() string {
	firstName := ""
	lastName := ""
	if u.FirstName != nil {
		firstName = *u.FirstName
	}
	if u.LastName != nil {
		lastName = *u.LastName
	}
	if firstName == "" && lastName == "" {
		return u.Email
	}
	return strings.TrimSpace(firstName + " " + lastName)
}

func (u *User) CanViewUser(targetUserID uint) bool {
	// Admin can see everything
	if u.IsAdmin() {
		return true
	}
	// Can see own data
	if u.ID == targetUserID {
		return true
	}
	// Manager can see team members
	if u.IsManager() {
		for _, member := range u.TeamMembers {
			if member.ID == targetUserID {
				return true
			}
		}
	}
	return false
}

func (u *User) GetAccessibleUserIDs() []uint {
	if u.IsAdmin() {
		return nil // nil means "all users"
	}

	accessible := []uint{u.ID} // Always own data

	if u.IsManager() {
		for _, member := range u.TeamMembers {
			accessible = append(accessible, member.ID)
		}
	}

	return accessible
}

func (u *User) GetTeamMemberIDs() []uint {
	var ids []uint
	for _, member := range u.TeamMembers {
		ids = append(ids, member.ID)
	}
	return ids
}
