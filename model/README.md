## Interfaces for data storage

```go
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
```

By implementing above interfaces, you can replace the default in memory storage.

### About in memory storage

The in memory data storage is built by sharded hashmaps with read-write locks to ensure safety of concurrent visiting. Check `inmem.go` for detailed implementation.

The expired tokens are deleted in a background routine, which will be invoked periodically and check if there are tokens to delete for within the hashmap shard (each period it'll check one shard).

```go
var (
  // shard sizes for users and tokes.
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

	// for token expiration
	tokenTTL                   time.Duration
	tokenExpirationCheckPeriod time.Duration

	// Signal to exit back ground routines
	exitChan chan struct{}
}
```
