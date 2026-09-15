package testing

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/omar-shahieen/moneyflow/internal/config"
	"github.com/omar-shahieen/moneyflow/internal/handler"
	"github.com/omar-shahieen/moneyflow/internal/middleware"
	"github.com/omar-shahieen/moneyflow/internal/repository"
	"github.com/omar-shahieen/moneyflow/internal/server"
	"github.com/omar-shahieen/moneyflow/internal/service"
	"github.com/stretchr/testify/require"
)

func init() {
	gin.SetMode(gin.TestMode)
}

type TestGinServer struct {
	Engine *gin.Engine
	Server *httptest.Server
	Config *config.Config
	t      testing.TB
}

func SetupTestGinServer(t *testing.T) (*TestGinServer, func()) {
	t.Helper()

	testDB, dbCleanup := SetupTestDB(t)

	srv, err := server.New(testDB.Config, nil, nil)
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	repos := repository.NewRepositories(srv)
	services, _ := service.NewServices(srv, repos, srv.Logger)
	handlers := handler.NewHandlers(srv, services)

	router := gin.New()
	middlewares := middleware.NewMiddlewares(srv)

	router.Use(
		middlewares.Global.CORS(),
		middlewares.Global.Secure(),
		middleware.RequestID(),
		middlewares.Global.Recover(),
	)

	_ = handlers

	ts := httptest.NewServer(router)

	return &TestGinServer{
		Engine: router,
		Server: ts,
		Config: testDB.Config,
		t:      t,
	}, func() {
		ts.Close()
		if testDB.Pool != nil {
			testDB.Pool.Close()
		}
		dbCleanup()
	}
}

func (s *TestGinServer) MakeRequest(method, path string, body interface{}) *httptest.ResponseRecorder {
	var reqBody *bytes.Buffer
	if body != nil {
		jsonBytes, err := json.Marshal(body)
		if err != nil {
			s.T().Fatalf("failed to marshal request body: %v", err)
		}
		reqBody = bytes.NewBuffer(jsonBytes)
	} else {
		reqBody = bytes.NewBuffer(nil)
	}

	req, err := http.NewRequest(method, path, reqBody)
	if err != nil {
		s.T().Fatalf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	s.Engine.ServeHTTP(w, req)

	return w
}

func (s *TestGinServer) T() testing.TB {
	return s.t
}

func AssertJSONStatus(t *testing.T, w *httptest.ResponseRecorder, expectedStatus int) {
	t.Helper()
	require.Equal(t, expectedStatus, w.Code, "unexpected status code")
}

func ParseJSONResponse(t *testing.T, w *httptest.ResponseRecorder) map[string]interface{} {
	t.Helper()
	var result map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &result)
	require.NoError(t, err, "failed to parse JSON response")
	return result
}
