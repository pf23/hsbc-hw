package model

import "fmt"

// This file lists all response status code and description.
type StatusCode int

const (
	Unknown StatusCode = 10000 + iota
	OK      StatusCode = 20000 + iota
	UserCreated
	UserDeleted
	UserRoleAdded
	RoleCreated
	RoleDeleted
	TokenRenewed
	TokenCreated
	TokenInvalidated
	TokenRoleOK

	InvalidArgument StatusCode = 40000 + iota
	UserAlreadyExisting
	UserNotFound
	UserPasswordNotMatch
	UserRoleAlreadyExisting
	RoleAlreadyExisting
	RoleNotFound
	TokenNotFound
	TokenExpired
	TokenIsInvalid
	TokenRoleNotFound
	Internal StatusCode = 50000
)

var (
	codeDesc = map[StatusCode]string{
		Unknown:                 "unknown",
		OK:                      "ok",
		InvalidArgument:         "invalid argument",
		UserAlreadyExisting:     "user already existing",
		UserCreated:             "user created",
		UserDeleted:             "user deleted",
		UserNotFound:            "user not found",
		UserPasswordNotMatch:    "user password not match",
		UserRoleAlreadyExisting: "user role already existing",
		UserRoleAdded:           "user role added",
		RoleAlreadyExisting:     "role already existing",
		RoleCreated:             "role created",
		RoleDeleted:             "role deleted",
		RoleNotFound:            "role not found",
		TokenRenewed:            "token renewed",
		TokenCreated:            "token created",
		TokenNotFound:           "token not found",
		TokenExpired:            "token expired",
		TokenIsInvalid:          "token is invalid",
		TokenInvalidated:        "token invalidated",
		TokenRoleOK:             "token role ok",
		TokenRoleNotFound:       "token role not found",
	}
)

func (c StatusCode) String() string {
	v, ok := codeDesc[c]
	if !ok {
		return fmt.Sprintf("StatusCode %v", int(c))
	}
	return v
}

func (c StatusCode) HTTPCode() int {
	return int(c) / 100
}
