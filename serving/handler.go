package serving

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	mdl "hsbc-hw/model"
)

var (
	mux    = make(map[string]map[string]func([]byte) ResponseCommon)
	engine mdl.AuthenticateAuthorizationEngine
)

func registerHandler(path, method string, h func([]byte) ResponseCommon) {
	m, ok := mux[path]
	if !ok {
		m = make(map[string]func([]byte) ResponseCommon)
		mux[path] = m
	}
	m[method] = h
}

func newMultiplexer(path string, m map[string]func([]byte) ResponseCommon) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		var resp ResponseCommon
		defer func() {
			code := resp.Status.HTTPCode()
			if code != 200 {
				resp.Data = nil
			}
			w.WriteHeader(code)
			log.Printf("%v %v, resp %+v", req.URL.Path, req.Method, resp)
			b, _ := json.Marshal(resp)
			w.Write(b)
		}()
		log.Printf("%v %v", req.URL.Path, req.Method)
		h, ok := m[req.Method]
		if !ok {
			resp = newResponse(mdl.InvalidArgument, fmt.Sprintf("method %v not implemented for url '%v'", req.Method, path))
			return
		}
		b, err := ioutil.ReadAll(req.Body)
		if err != nil {
			resp = newResponse(mdl.Internal, err.Error())
			return
		}
		req.Body.Close()
		resp = h(b)
	}
}

func init() {
	registerHandler("/user", "POST", CreateUser)
	registerHandler("/user", "DELETE", DeleteUser)
	registerHandler("/user/role", "POST", AddUserRole)
	registerHandler("/user/auth", "POST", AuthenticateUser)
	registerHandler("/role", "POST", CreateRole)
	registerHandler("/role", "DELETE", DeleteRole)
	registerHandler("/token", "DELETE", Invalidate)
	registerHandler("/token/role", "GET", CheckRole)
	registerHandler("/token/roles", "GET", AllRoles)
	for path, m := range mux {
		http.HandleFunc(path, newMultiplexer(path, m))
	}
	engine = mdl.NewInmemEngine()
}

func newEngineForTesting() {
	engine = mdl.NewInmemEngine()
}

func Cleanup() {
	engine.Shutdown()
}

func CreateUser(b []byte) ResponseCommon {
	in := new(CreateUserRequest)
	if err := json.Unmarshal(b, &in); err != nil {
		return newResponse(mdl.InvalidArgument, err.Error())
	}
	if in.UserName == "" || in.Password == "" {
		return newResponse(mdl.InvalidArgument, "empty user_name or password")
	}

	code := engine.CreateUser(mdl.User{
		Name:         in.UserName,
		PwdEncrypted: encryptPassword(in.Password),
	})
	return newResponse(code, code.String())
}

func DeleteUser(b []byte) ResponseCommon {
	in := new(DeleteUserRequest)
	if err := json.Unmarshal(b, &in); err != nil {
		return newResponse(mdl.InvalidArgument, err.Error())
	}
	code := engine.DeleteUser(mdl.User{
		Name:         in.UserName,
		PwdEncrypted: encryptPassword(in.Password),
	})
	return newResponse(code, code.String())
}

func AddUserRole(b []byte) ResponseCommon {
	in := new(AddUserRoleRequest)
	if err := json.Unmarshal(b, &in); err != nil {
		return newResponse(mdl.InvalidArgument, err.Error())
	}
	if in.UserName == "" || in.RoleName == "" {
		return newResponse(mdl.InvalidArgument, "empty user_name or role_name")
	}
	code := engine.AddUserRole(
		mdl.User{Name: in.UserName}, // no password required
		mdl.Role{Name: in.RoleName},
	)
	return newResponse(code, code.String())
}

func AuthenticateUser(b []byte) ResponseCommon {
	in := new(AuthenticateRequest)
	if err := json.Unmarshal(b, &in); err != nil {
		return newResponse(mdl.InvalidArgument, err.Error())
	}
	token, code := engine.Authenticate(mdl.User{
		Name:         in.UserName,
		PwdEncrypted: encryptPassword(in.Password),
	})
	return newResponseData(code, code.String(), AuthenticateResponse{
		Token:           token.ID,
		ExpiredAtInUsec: token.ExpiredAtInUsec,
	})
}

func CreateRole(b []byte) ResponseCommon {
	in := new(CreateRoleRequest)
	if err := json.Unmarshal(b, &in); err != nil {
		return newResponse(mdl.InvalidArgument, err.Error())
	}
	if in.RoleName == "" {
		return newResponse(mdl.InvalidArgument, "empty role_name")
	}
	code := engine.CreateRole(mdl.Role{Name: in.RoleName})
	return newResponse(code, code.String())
}

func DeleteRole(b []byte) ResponseCommon {
	in := new(DeleteRoleRequest)
	if err := json.Unmarshal(b, &in); err != nil {
		return newResponse(mdl.InvalidArgument, err.Error())
	}
	code := engine.DeleteRole(mdl.Role{Name: in.RoleName})
	return newResponse(code, code.String())
}

func Invalidate(b []byte) ResponseCommon {
	in := new(InvalidateRequest)
	if err := json.Unmarshal(b, &in); err != nil {
		return newResponse(mdl.InvalidArgument, err.Error())
	}
	code := engine.Invalidate(in.Token)
	return newResponse(code, code.String())
}

func CheckRole(b []byte) ResponseCommon {
	in := new(CheckRoleRequest)
	if err := json.Unmarshal(b, &in); err != nil {
		return newResponse(mdl.InvalidArgument, err.Error())
	}
	code := engine.CheckRole(in.Token, in.RoleName)
	return newResponse(code, code.String())
}

func AllRoles(b []byte) ResponseCommon {
	in := new(AllRolesRequest)
	if err := json.Unmarshal(b, &in); err != nil {
		return newResponse(mdl.InvalidArgument, err.Error())
	}
	roles, code := engine.AllRoles(in.Token)
	resp := AllRolesResponse{Token: in.Token}
	for _, r := range roles {
		resp.Roles = append(resp.Roles, r.Name)
	}
	return newResponseData(code, code.String(), resp)
}

// ResponseCommon
type ResponseCommon struct {
	Status  mdl.StatusCode `json:"status"`
	Message string         `json:"message,omitempty"`
	Data    interface{}    `json:"data,omitempty"`
}

func newResponse(s mdl.StatusCode, m string) ResponseCommon {
	return newResponseData(s, m, nil)
}

func newResponseData(s mdl.StatusCode, m string, data interface{}) ResponseCommon {
	return ResponseCommon{
		Status:  s,
		Message: m,
		Data:    data,
	}
}

type CreateUserRequest struct {
	UserName string `json:"user_name"`
	Password string `json:"password"`
}

type DeleteUserRequest struct {
	UserName string `json:"user_name"`
	Password string `json:"password"`
}

type CreateRoleRequest struct {
	RoleName string `json:"role_name"`
}

type DeleteRoleRequest struct {
	RoleName string `json:"role_name"`
}

type AddUserRoleRequest struct {
	UserName string `json:"user_name"`
	RoleName string `json:"role_name"`
}

type AuthenticateRequest struct {
	UserName string `json:"user_name"`
	Password string `json:"password"`
}

type AuthenticateResponse struct {
	Token           string `json:"token"`
	ExpiredAtInUsec int64  `json:"expired_at_in_usec"`
}

type InvalidateRequest struct {
	Token string `json:"token"`
}

type CheckRoleRequest struct {
	Token    string `json:"token"`
	RoleName string `json:"role_name"`
}

type AllRolesRequest struct {
	Token string `json:"token"`
}

type AllRolesResponse struct {
	Token string   `json:"token"`
	Roles []string `json:"roles"`
}

func encryptPassword(pwd string) string {
	return base64.URLEncoding.EncodeToString([]byte(pwd))
}
