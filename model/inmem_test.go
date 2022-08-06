package model

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var (
	u1               = User{Name: "u1", PwdEncrypted: "xxxx"}
	u12              = User{Name: "u1", PwdEncrypted: "yyyy"}
	u2               = User{Name: "u2", PwdEncrypted: "zzzz"}
	r1               = Role{Name: "r1"}
	r2               = Role{Name: "r2"}
	r3               = Role{Name: "r3"}
	tokenNotExisting = Token{ID: "__not_existing__"}
)

func statusCodeEqual(t *testing.T, expected, actual StatusCode) {
	assert.Equal(t, expected, actual)
	assert.Equal(t, expected.String(), actual.String())
}

func TestTokenGenerator(t *testing.T) {
	tt, _ := time.Parse(time.RFC3339, "2022-08-06T15:04:05Z07:00")
	token := generateToken("__test_user__", "__pass_word__", tt)
	assert.Equal(t, "ITnCejjs1pXtVDpI-6eDfw==", token)
}

func TestUserBasic(t *testing.T) {
	e := NewInmemEngine()
	statusCodeEqual(t, UserCreated, e.CreateUser(u1))
	statusCodeEqual(t, UserAlreadyExisting, e.CreateUser(u1))
	statusCodeEqual(t, UserPasswordNotMatch, e.DeleteUser(u12))
	statusCodeEqual(t, UserDeleted, e.DeleteUser(u1))
	statusCodeEqual(t, UserNotFound, e.DeleteUser(u1))
	e.Shutdown()
}

func TestRoleBasic(t *testing.T) {
	e := NewInmemEngine()
	statusCodeEqual(t, RoleCreated, e.CreateRole(r1))
	statusCodeEqual(t, RoleAlreadyExisting, e.CreateRole(r1))
	statusCodeEqual(t, RoleDeleted, e.DeleteRole(r1))
	statusCodeEqual(t, RoleNotFound, e.DeleteRole(r1))
}

func TestUserRole(t *testing.T) {
	e := NewInmemEngine()
	statusCodeEqual(t, UserCreated, e.CreateUser(u1))
	statusCodeEqual(t, RoleNotFound, e.AddUserRole(u1, r1))
	statusCodeEqual(t, RoleCreated, e.CreateRole(r1))
	statusCodeEqual(t, UserRoleAdded, e.AddUserRole(u1, r1))
	statusCodeEqual(t, UserRoleAlreadyExisting, e.AddUserRole(u1, r1))
}

func TestAuthenticate(t *testing.T) {
	e := NewInmemEngine()
	var code StatusCode
	statusCodeEqual(t, UserCreated, e.CreateUser(u1))
	_, code = e.Authenticate(u12)
	statusCodeEqual(t, UserPasswordNotMatch, code)
	_, code = e.Authenticate(u2)
	statusCodeEqual(t, UserNotFound, code)
	_, code = e.Authenticate(u1)
	statusCodeEqual(t, TokenCreated, code)
	_, code = e.Authenticate(u1)
	statusCodeEqual(t, TokenRenewed, code)
}

func TestInvalidate(t *testing.T) {
	e := NewInmemEngine()
	var code StatusCode
	var token Token
	statusCodeEqual(t, UserCreated, e.CreateUser(u1))
	token, code = e.Authenticate(u1)
	statusCodeEqual(t, TokenCreated, code)
	statusCodeEqual(t, TokenNotFound, e.Invalidate(tokenNotExisting.ID))
	statusCodeEqual(t, TokenInvalidated, e.Invalidate(token.ID))
	statusCodeEqual(t, TokenIsInvalid, e.CheckRole(token.ID, r1.Name))
}

func TestTokenExpired(t *testing.T) {
	tokenShardSize = 1
	e := NewInmemEngine()
	e.(*inmemEngine).SetTokenTTL(time.Millisecond * 100)
	var code StatusCode
	var token Token
	statusCodeEqual(t, UserCreated, e.CreateUser(u1))
	token, code = e.Authenticate(u1)
	statusCodeEqual(t, TokenCreated, code)
	statusCodeEqual(t, TokenNotFound, e.Invalidate(tokenNotExisting.ID))
	time.Sleep(time.Millisecond * 100)
	statusCodeEqual(t, TokenExpired, e.Invalidate(token.ID))
	time.Sleep(time.Millisecond * 150)
	statusCodeEqual(t, TokenNotFound, e.Invalidate(token.ID))
	tokenShardSize = 1024
}

func TestDeleteUserWithToken(t *testing.T) {
	e := NewInmemEngine()
	var code StatusCode
	statusCodeEqual(t, UserCreated, e.CreateUser(u1))
	_, code = e.Authenticate(u1)
	statusCodeEqual(t, TokenCreated, code)
	statusCodeEqual(t, UserDeleted, e.DeleteUser(u1))
}

func TestCheckRole(t *testing.T) {
	e := NewInmemEngine()
	var code StatusCode
	var token Token
	statusCodeEqual(t, TokenNotFound, e.CheckRole(tokenNotExisting.ID, r1.Name))
	statusCodeEqual(t, UserCreated, e.CreateUser(u1))
	token, code = e.Authenticate(u1)
	statusCodeEqual(t, TokenCreated, code)
	statusCodeEqual(t, TokenRoleNotFound, e.CheckRole(token.ID, u1.Name))
	statusCodeEqual(t, RoleCreated, e.CreateRole(r1))
	statusCodeEqual(t, UserRoleAdded, e.AddUserRole(u1, r1))
	statusCodeEqual(t, TokenRoleOK, e.CheckRole(token.ID, r1.Name))
}

func TestAllRoles(t *testing.T) {
	e := NewInmemEngine()
	var code StatusCode
	var token Token
	var rs []Role
	_, code = e.AllRoles(tokenNotExisting.ID)
	statusCodeEqual(t, TokenNotFound, code)
	statusCodeEqual(t, UserCreated, e.CreateUser(u1))
	token, code = e.Authenticate(u1)
	statusCodeEqual(t, TokenCreated, code)
	statusCodeEqual(t, RoleCreated, e.CreateRole(r1))
	statusCodeEqual(t, RoleCreated, e.CreateRole(r2))
	statusCodeEqual(t, RoleCreated, e.CreateRole(r3))
	statusCodeEqual(t, UserRoleAdded, e.AddUserRole(u1, r1))
	statusCodeEqual(t, UserRoleAdded, e.AddUserRole(u1, r2))
	rs, code = e.AllRoles(token.ID)
	statusCodeEqual(t, OK, code)
	assert.Equal(t, 2, len(rs))
	assert.Equal(t, r1.Name, rs[0].Name)
	assert.Equal(t, r2.Name, rs[1].Name)
	statusCodeEqual(t, RoleDeleted, e.DeleteRole(r1))
	rs, code = e.AllRoles(token.ID)
	statusCodeEqual(t, OK, code)
	assert.Equal(t, 1, len(rs))
	assert.Equal(t, r2.Name, rs[0].Name)
}

func TestConcurrent(t *testing.T) {

}
