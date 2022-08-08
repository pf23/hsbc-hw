package model

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/binary"
	"sync"
	"time"
)

var (
	userShardSize  uint32 = 1024
	tokenShardSize uint32 = 1024
)

type userPartition struct {
	sync.RWMutex // rw lock for concurrent control
	users        map[string]*User
}

type tokenPartition struct {
	sync.RWMutex // rw lock for concurrent control
	tokens       map[string]*Token
}

type inmemEngine struct {
	// Inmem Lookup tables
	users    []*userPartition  // UserName - User
	tokens   []*tokenPartition // TokenID - User
	roles    map[string]*Role  // RoleName - Role
	rolelock sync.RWMutex

	// For token expiration
	tokenTTL                   time.Duration
	tokenExpirationCheckPeriod time.Duration

	// Signal to exit back ground routines
	exitChan chan struct{}
}

// NewInmemEngine inits a new instance of inmemEngine and start background job
// to delete expired tokens.
func NewInmemEngine() AuthenticateAuthorizationEngine {
	e := &inmemEngine{
		users:                      make([]*userPartition, userShardSize),
		tokens:                     make([]*tokenPartition, tokenShardSize),
		roles:                      make(map[string]*Role),
		tokenTTL:                   2 * time.Hour,
		tokenExpirationCheckPeriod: time.Millisecond * 200,
		exitChan:                   make(chan struct{}),
	}
	for i := uint32(0); i < userShardSize; i++ {
		e.users[i] = &userPartition{users: make(map[string]*User)}
	}
	for i := uint32(0); i < tokenShardSize; i++ {
		e.tokens[i] = &tokenPartition{tokens: make(map[string]*Token)}
	}
	go e.deleteExpiredTokens()
	return e
}

func (e *inmemEngine) SetTokenTTL(du time.Duration) {
	e.tokenTTL = du
}

func (e *inmemEngine) CreateUser(u User) StatusCode {
	p := e.getUserPartition(u.Name)
	p.Lock()
	defer p.Unlock()

	if _, ok := p.users[u.Name]; ok {
		return UserAlreadyExisting
	}
	p.users[u.Name] = &User{
		Name:         u.Name,
		PwdEncrypted: u.PwdEncrypted,
	}
	return UserCreated
}

func (e *inmemEngine) DeleteUser(u User) StatusCode {
	p := e.getUserPartition(u.Name)
	p.Lock()
	defer p.Unlock()

	if status := e.checkUserPassword(p, u); status != OK {
		return status
	}
	if t := p.users[u.Name].token; t != nil {
		t.user = nil
		t.invalid = true
		p.users[u.Name].token = nil
	}
	delete(p.users, u.Name)
	return UserDeleted
}

func (e *inmemEngine) CreateRole(r Role) StatusCode {
	e.rolelock.Lock()
	defer e.rolelock.Unlock()

	if _, ok := e.roles[r.Name]; ok {
		return RoleAlreadyExisting
	}
	e.roles[r.Name] = &Role{
		Name: r.Name,
	}
	return RoleCreated
}

func (e *inmemEngine) DeleteRole(r Role) StatusCode {
	e.rolelock.Lock()
	defer e.rolelock.Unlock()

	cur, ok := e.roles[r.Name]
	if !ok {
		return RoleNotFound
	}
	cur.deleted = true
	delete(e.roles, r.Name)
	return RoleDeleted
}

func (e *inmemEngine) AddUserRole(u User, r Role) StatusCode {
	p := e.getUserPartition(u.Name)
	p.Lock()
	defer p.Unlock()

	cur, ok := p.users[u.Name]
	if !ok {
		return UserNotFound
	}
	if checkUserRole(r, p.users[u.Name]) {
		return UserRoleAlreadyExisting
	}
	e.rolelock.RLock()
	defer e.rolelock.RUnlock()
	rr, ok := e.roles[r.Name]
	if !ok {
		return RoleNotFound
	}
	cur.roles = append(cur.roles, rr)
	return UserRoleAdded
}

func (e *inmemEngine) Authenticate(u User) (Token, StatusCode) {
	p := e.getUserPartition(u.Name)
	p.Lock()
	defer p.Unlock()

	if status := e.checkUserPassword(p, u); status != OK {
		return nilToken, status
	}
	cur := p.users[u.Name]
	if cur.token != nil && !cur.token.invalid {
		// Extend the expiration of token
		cur.token.ExpiredAtInUsec = tokenExpirationInUsecFromTime(time.Now(), e.tokenTTL)
		return Token{ID: cur.token.ID}, TokenRenewed
	}
	cur.token = &Token{
		ID:              generateToken(u.Name, u.PwdEncrypted, time.Now()),
		ExpiredAtInUsec: tokenExpirationInUsecFromTime(time.Now(), e.tokenTTL),
		user:            cur,
	}

	pp := e.getTokePartition(cur.token.ID)
	pp.Lock()
	defer pp.Unlock()
	pp.tokens[cur.token.ID] = cur.token
	return Token{ID: cur.token.ID, ExpiredAtInUsec: cur.token.ExpiredAtInUsec}, TokenCreated
}

func (e *inmemEngine) Invalidate(t string) StatusCode {
	pp := e.getTokePartition(t)
	pp.Lock()
	defer pp.Unlock()

	token, status := getValidToken(pp, t)
	if token == nil {
		return status
	}
	token.invalid = true
	token.user = nil
	return TokenInvalidated
}

func (e *inmemEngine) CheckRole(t, r string) StatusCode {
	pp := e.getTokePartition(t)
	pp.RLock()
	defer pp.RUnlock()

	token, status := getValidToken(pp, t)
	if token == nil {
		return status
	}

	if checkUserRole(Role{Name: r}, token.user) {
		return TokenRoleOK
	}
	return TokenRoleNotFound
}

func (e *inmemEngine) AllRoles(t string) ([]Role, StatusCode) {
	pp := e.getTokePartition(t)
	pp.RLock()
	defer pp.RUnlock()

	token, status := getValidToken(pp, t)
	if token == nil {
		return nil, status
	}
	deleteInvalidRoles(token.user)
	res := make([]Role, 0, len(token.user.roles))
	for _, v := range token.user.roles {
		res = append(res, Role{Name: v.Name})
	}
	return res, OK
}

func (e *inmemEngine) Shutdown() {
	close(e.exitChan)
}

//
// lower level funcs
//
func (e *inmemEngine) deleteExpiredTokens() {
	t := time.NewTicker(e.tokenExpirationCheckPeriod)
	tokenShardIndex := 0
	for {
		select {
		case <-t.C:
			pp := e.tokens[tokenShardIndex]
			pp.Lock()
			nowInUsec := time.Now().UnixNano() / 1000
			for id, v := range pp.tokens {
				if v.ExpiredAtInUsec < nowInUsec {
					delete(pp.tokens, id)
				}
			}
			pp.Unlock()
			tokenShardIndex = (tokenShardIndex + 1) % int(tokenShardSize)
		case <-e.exitChan:
			t.Stop()
			return
		}
	}
}

func (e *inmemEngine) getUserPartition(name string) *userPartition {
	return e.users[hashStringToInt32(name)%tokenShardSize]
}

func (e *inmemEngine) getTokePartition(token string) *tokenPartition {
	return e.tokens[hashStringToInt32(token)%tokenShardSize]
}

func (e *inmemEngine) checkUserPassword(p *userPartition, u User) StatusCode {
	cur, ok := p.users[u.Name]
	if !ok {
		return UserNotFound
	}
	if cur.PwdEncrypted != u.PwdEncrypted {
		return UserPasswordNotMatch
	}
	return OK
}

func getValidToken(pp *tokenPartition, t string) (*Token, StatusCode) {
	token, ok := pp.tokens[t]
	if !ok {
		return nil, TokenNotFound
	}
	if expiredByTime(token.ExpiredAtInUsec, time.Now()) {
		return nil, TokenExpired
	}
	if token.invalid || token.user == nil {
		return nil, TokenIsInvalid
	}
	return token, OK
}

func deleteInvalidRoles(u *User) {
	if u == nil {
		return
	}
	res := make([]*Role, 0, len(u.roles))
	for i := 0; i < len(u.roles); i++ {
		if !u.roles[i].deleted {
			res = append(res, u.roles[i])
		}
	}
	u.roles = res
	return
}

func checkUserRole(r Role, u *User) bool {
	if u == nil {
		return false
	}
	deleteInvalidRoles(u)
	for _, v := range u.roles {
		if v.Name == r.Name && !v.deleted {
			return true
		}
	}
	return false
}

func tokenExpirationInUsecFromTime(t time.Time, du time.Duration) int64 {
	return t.Add(du).UnixNano() / 1000
}

func expiredByTime(usec int64, t time.Time) bool {
	return usec < (t.UnixNano() / 1000)
}

func hashStringToInt32(s string) uint32 {
	h := md5.New()
	h.Write([]byte(s))
	b := h.Sum(nil)
	return binary.BigEndian.Uint32(b)
}

func generateToken(name, pwd string, t time.Time) string {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(t.UnixNano()))
	h := md5.New()
	h.Write([]byte(name))
	h.Write([]byte(pwd))
	h.Write(b)
	return base64.URLEncoding.EncodeToString(h.Sum(nil))
}
