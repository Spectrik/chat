package user

import (
	"strings"
)

type Role uint

const (
    RoleUser Role = iota
    RoleMod
    RoleAdmin
)

var roleFromString = map[string]Role{
	"user":  RoleUser,
	"mod":   RoleMod,
	"admin": RoleAdmin,
}

func (r Role) String() string {
	switch r {
	case RoleUser:
		return "user"
	case RoleMod:
		return "mod"
	case RoleAdmin:
		return "admin"
	default:
		return "unknown"
	}
}

func RoleFromString(s string) (Role, bool) {
	r, ok := roleFromString[strings.ToLower(s)]
	return r, ok
}

type User struct {
	Username     string
	Role         Role
	Disabled     bool
}

func (u User) Equal(other User) bool {
    return u.Username == other.Username &&
           u.Role == other.Role &&
           u.Disabled == other.Disabled
}
