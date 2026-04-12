package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type Role string

const (
	RoleAdmin Role = "admin"
	RoleUser  Role = "user"
)

var (
	ErrInvalidRole = errors.New("invalid user role")
)

type User struct {
	ID        uuid.UUID
	Email     string
	Password  string
	Role      Role
	CreatedAt time.Time
}

func NewUser(email string, password string, role Role) (*User, error) {
	if !role.IsValid() {
		return nil, ErrInvalidRole
	}

	return &User{
		ID:        uuid.New(),
		Email:     email,
		Password:  password,
		Role:      role,
		CreatedAt: time.Now().UTC(),
	}, nil
}

func (r Role) IsValid() bool {
	switch r {
	case RoleAdmin, RoleUser:
		return true
	default:
		return false
	}
}

func (u *User) IsAdmin() bool {
	return u.Role == RoleAdmin
}

func (u *User) IsUser() bool {
	return u.Role == RoleUser
}
