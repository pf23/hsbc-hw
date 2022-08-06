## HTTP API Document

Both resquest body and response body should be in Json format.

### Rule for Response

**HTTP Code**

For simplicity (and it satisfies for now), services only return an HTTP Code from below 3 alternatives,

* 200: operation succeeded or other normal cases.
* 400: invalid input or operation failed.
* 500: severe interval error.

**Body Format**

The response body format is defined as follow, with an internal status code and description message to explain what happened, carrying extra data if needed.

```json
{
  "status": 20002,
  "message": "user created",
  "data": {},  
}
```

**Status Code**

Below lists current set of status code and descriptions.

```
10000 unknown

20001 ok
20002 user created
20003 user deleted
20004 user role added
20005 role created
20006 role deleted
20007 token renewed
20008 token created
20009 token invalidated
20010 token role ok

40011 invalid argument
40012 user already existing
40013 user not found
40014 user password not match
40015 user role already existing
40016 role already existing
40017 role not found
40018 token not found
40019 token expired
40020 token is invalid
40021 token role not found

50000 internal
```

### Request & Response Document

| Function | URL | HTTP Method | Payload Demo | Succeeded Response Demo |
|---|---|---|---|---|
| CreateUser | /user | POST | {"user_name": "uname1", "password": "pwd1"} | {"status": 20002, "message": "user created"} |
| DeleteUser | /user | DELETE | {"user_name": "uname1", "password": "pwd1"} | {"status": 20003, "message": "user deleted"} |
| AddUserRole | /user/role | POST | {"user_name": "uname1", "role_name": "role1"} | {"status": 20004, "message": "user role added"} |
| AuthenticateUser | /user/auth | POST | {"user_name": "uname1", "password": "pwd1"} | {"status": 20008 or 20007, "message": "token created" or "token renewed", "data": {"toke": ""ZU6o9wcfvROW5YHh5ChMzw==", "expired_at_in_usec": 1659762467740160} | |
| CreateRole | /role | POST | {"role_name": "role1"} | {"status": 20005, "message": "role created"} |
| DeleteRole | /role | DELETE | {"role_name": "role1"} | {"status": 20006, "message": "role deleted"} |
| Invalidate | /token | DELETE | {"token": "ZU6o9wcfvROW5YHh5ChMzw=="} | {"status": 20009, "message": "token invalidated"} |
| CheckRole | /token/role | GET | {"token": "ZU6o9wcfvROW5YHh5ChMzw==", "role_name": "role1"} | {"status": 20010, "message": "token role ok"} |
| AllRoles | /token/roles | GET | {"token": "ZU6o9wcfvROW5YHh5ChMzw=="} | {"status": 20001, "message": "ok", data: {"token": ZU6o9wcfvROW5YHh5ChMzw==", "roles": ["role1", "role2", "role3"]} |
