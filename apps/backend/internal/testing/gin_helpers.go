package testing

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/omar-shahieen/moneyflow/internal/config"
	"github.com/omar-shahieen/moneyflow/internal/router"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
)

func init() {
	gin.SetMode(gin.TestMode)
}

type TestGinServer struct {
	Router *router.GinRouter
	Engine *gin.Engine
	DB     *pgxpool.Pool
	Config *config.Config
	Logger zerolog.Logger
	Server *httptest.Server
	t      testing.TB
}

func SetupTestGinServer(t *testing.T) (*TestGinServer, func()) {
	t.Helper()

	testDB, dbCleanup := SetupTestDB(t)

	logger := zerolog.New(zerolog.ConsoleWriter{Out: testWriter{t}}).
		Level(zerolog.InfoLevel).
		With().
		Timestamp().
		Logger()

	r := router.NewGinRouter(testDB.Config, testDB.Pool, logger)

	ts := httptest.NewServer(r.Engine())

	return &TestGinServer{
		Router: r,
		Engine: r.Engine(),
		DB:     testDB.Pool,
		Config: testDB.Config,
		Logger: logger,
		Server: ts,
		t:      t,
	}, func() {
		ts.Close()
		if testDB.Pool != nil {
			testDB.Pool.Close()
		}
		dbCleanup()
	}
}

type testWriter struct {
	t *testing.T
}

func (tw testWriter) Write(p []byte) (n int, err error) {
	tw.t.Log(string(p))
	return len(p), nil
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
