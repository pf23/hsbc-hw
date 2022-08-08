package serving

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"testing"

	mdl "hsbc-hw/model"

	"github.com/stretchr/testify/assert"
)

const (
	serverPort = ":8083"
	serverAddr = "http://localhost:8083"
)

var (
	cli *http.Client
	srv *http.Server
)

func initialize() {
	cli = &http.Client{}
	srv = &http.Server{Addr: serverPort}
	go func() {
		srv.ListenAndServe()
	}()
}

type req2resp struct {
	url      string
	method   string
	payload  string
	respCode mdl.StatusCode
	httpCode int
}

func expected(url, method, payload string, respCode mdl.StatusCode, httpCode int) req2resp {
	return req2resp{
		url:      url,
		method:   method,
		payload:  payload,
		respCode: respCode,
		httpCode: httpCode,
	}
}

func makeRequestsAndAssert(t *testing.T, seq ...req2resp) {
	for _, v := range seq {
		req, _ := http.NewRequest(
			v.method,
			serverAddr+v.url,
			strings.NewReader(v.payload),
		)
		resp, err := cli.Do(req)
		assert.Nil(t, err)

		b, _ := ioutil.ReadAll(resp.Body)
		resp.Body.Close()

		data := new(ResponseCommon)
		assert.Nil(t, json.Unmarshal(b, data))
		assert.Equal(t, v.respCode, data.Status)
		assert.Equal(t, v.httpCode, resp.StatusCode)
	}
}

func TestBasic(t *testing.T) {
	newEngineForTesting()
	makeRequestsAndAssert(t,
		expected("/user", "POST", `{"user_name": "qwer", "password": "qsc123"}`,
			mdl.UserCreated, 200),
		expected("/user", "POST", `{"user_name": "qwer", "password": "qsc123"}`,
			mdl.UserAlreadyExisting, 400),
		expected("/user", "DELETE", `{"user_name": "qwer", "password": "qsc1234"}`,
			mdl.UserPasswordNotMatch, 400),
		expected("/user", "DELETE", `{"user_name": "qwer", "password": "qsc123"}`,
			mdl.UserDeleted, 200),
		expected("/user", "DELETE", `{"user_name": "qwer", "password": "qsc123"}`,
			mdl.UserNotFound, 400),
	)
}

func TestMain(m *testing.M) {
	initialize()
	exitCode := m.Run()
	srv.Shutdown(context.Background())
	os.Exit(exitCode)
}
