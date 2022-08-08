# Simple authentication and authorization service

## Feature notations

This is a Golang implementation of `a simple authentication and authorization service`. According to requirements, the project

1. hosts a HTTP server to provide services,
2. uses in-memory storage (can be adapted to other storage by conforming to interfaces),
3. is implemented mostly by Go built-in packages (except a unit-test library),

For detailed HTTP API document, please check [serving/API.md)](serving/API.md)

## Directory

```markdown
.
├── cmd                     # executable
│   ├── go.mod              # go module files
│   ├── go.sum              # go module files
│   └── server.go           # entrypoint of server
│
├── model                   # data relation model and storage engine
│   ├── go.mod
│   ├── go.sum
│   ├── inmem_test.go       # unit tests for inmem.go
│   ├── inmem.go            # in-memory implementation of interface in model.go
│   ├── model.go            # data model and storage interface definition
│   └── status.go           # status code and description
│
├── serving                 # implementation of services
│   ├── go.mod
│   ├── go.sum
│   ├── handler_test.go     # function tests for HTTP implementation
│   ├── handler.go          # handlers for HTTP APIs
│   └── README.md           # HTTP API documentations
│
├── stresstest              # (todo) stresstest for serving implementation
|   └── .....
│
├── build_docker.sh
├── build.sh
├── format.sh               # script to format the code
├── hsbc-hw.md              # task description and requirements
├── README.md
├── run_docker.sh
├── run.sh
└── server.Dockerfile
```

## How to build & run

* Build & run under Linux or MacOS

  Requirements: go version >= 1.15

  ```sh
  chmod +x build.sh
  ./build.sh # you'll see both unit tests output and build information.

  chmod +x run.sh
  ./run.sh
  ```

  You'll see server logs from terminal like

  ```
  2022/08/08 07:16:17 authenticate_server: start listen on :8080
  ```

  By `curl http://127.0.0.1:8080/`, if you receive a `Hello` message, it means the http serving is already serving.

* Build from docker

  ```sh
  chmod +x build_docker.sh
  ./build_docker.sh

  chmod +x run_docker.sh
  ./run_docker.sh
  ```
