package user

import "time"

type User struct {
	ID        string     `json:"id" gorm:"primaryKey;type:varchar(36)"`
	Username  string     `json:"username" gorm:"type:varchar(50);uniqueIndex;not null"`
	Password  string     `json:"-" gorm:"type:varchar(255);not null"`
	Email     string     `json:"email" gorm:"type:varchar(100);uniqueIndex"`
	RealName  string     `json:"real_name" gorm:"type:varchar(50)"`
	Role      string     `json:"role" gorm:"type:varchar(20);not null;default:'user'"`
	Status    string     `json:"status" gorm:"type:varchar(20);not null;default:'active'"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	LastLogin *time.Time `json:"last_login"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token string    `json:"token"`
	User  *UserInfo `json:"user"`
}

type UserInfo struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	RealName string `json:"real_name"`
	Role     string `json:"role"`
}

const (
	RoleAdmin      = "admin"
	RoleUser       = "user"
	RoleFinance    = "finance"
	StatusActive   = "active"
	StatusInactive = "inactive"
	StatusLocked   = "locked"
)

func (u *User) ToUserInfo() *UserInfo {
	return &UserInfo{
		ID:       u.ID,
		Username: u.Username,
		Email:    u.Email,
		RealName: u.RealName,
		Role:     u.Role,
	}
}

func (u *User) IsActive() bool {
	return u.Status == StatusActive
}

func (u *User) IsAdmin() bool {
	return u.Role == RoleAdmin
}

func (u *User) IsFinance() bool {
	return u.Role == RoleFinance
}

func (u *User) CanManualAudit() bool {
	return u.Role == RoleAdmin || u.Role == RoleFinance
}
