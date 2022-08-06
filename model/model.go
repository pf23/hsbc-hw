package model

var (
	nilUser  = User{}
	nilRole  = Role{}
	nilToken = Token{}
)

type User struct {
	Name         string
	PwdEncrypted string
	roles        []*Role
	token        *Token
}

type Role struct {
	Name    string
	deleted bool
}

type Token struct {
	ID              string
	ExpiredAtInUsec int64
	invalid         bool
	user            *User
}

// AuthenticateAuthorizationEngine defines db level interfaces
type AuthenticateAuthorizationEngine interface {
	CreateUser(u User) StatusCode
	DeleteUser(u User) StatusCode
	CreateRole(r Role) StatusCode
	DeleteRole(r Role) StatusCode
	AddUserRole(u User, r Role) StatusCode
	Authenticate(u User) (Token, StatusCode)
	Invalidate(t string) StatusCode
	CheckRole(t, r string) StatusCode
	AllRoles(t string) ([]Role, StatusCode)
	Shutdown()
}
