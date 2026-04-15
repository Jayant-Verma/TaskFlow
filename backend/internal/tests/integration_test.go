//go:build integration
// +build integration

// Integration tests for the TaskFlow backend.
//
// Run with:
//   cd backend
//   go test -v -tags=integration ./internal/tests

package tests

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"taskflow-api/internal/database"
	"taskflow-api/internal/handlers"
	"taskflow-api/internal/middleware"
	"taskflow-api/internal/models"

	"github.com/golang-jwt/jwt/v5"
)

const defaultJWTSecret = "integration-test-secret"

func envOrDefault(key, def string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return def
}

func connectTestDB(t *testing.T) *sql.DB {
	t.Helper()

	os.Setenv("DB_USER", envOrDefault("DB_USER", "taskuser"))
	os.Setenv("DB_PASSWORD", envOrDefault("DB_PASSWORD", "taskpass"))
	os.Setenv("DB_NAME", envOrDefault("DB_NAME", "taskflow"))
	os.Setenv("DB_HOST", envOrDefault("DB_HOST", "127.0.0.1"))
	os.Setenv("DB_PORT", envOrDefault("DB_PORT", "5432"))

	db, err := database.ConnectAndMigrate()
	if err != nil {
		t.Fatalf("failed to connect and migrate test database: %v", err)
	}

	t.Cleanup(func() {
		db.Close()
	})

	return db
}

func truncateTables(t *testing.T, db *sql.DB) {
	t.Helper()

	_, err := db.ExecContext(context.Background(), "TRUNCATE TABLE tasks, projects, users RESTART IDENTITY CASCADE")
	if err != nil {
		t.Fatalf("failed to truncate integration tables: %v", err)
	}
}

func decodeJSONResponse[T any](t *testing.T, body *strings.Reader, out *T) {
	t.Helper()
	if err := json.NewDecoder(body).Decode(out); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}
}

func TestAuthRegisterLoginIntegration(t *testing.T) {
	db := connectTestDB(t)
	truncateTables(t, db)

	authHandler := &handlers.AuthHandler{DB: db, JWTSecret: []byte(defaultJWTSecret)}

	registerBody := strings.NewReader(`{"name":"Integration Tester","email":"integration@example.com","password":"password123"}`)
	registerReq := httptest.NewRequest(http.MethodPost, "/auth/register", registerBody)
	registerRec := httptest.NewRecorder()

	authHandler.Register(registerRec, registerReq)

	if registerRec.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, registerRec.Code, registerRec.Body.String())
	}

	var registerResp map[string]string
	decodeJSONResponse(t, strings.NewReader(registerRec.Body.String()), &registerResp)

	userID, ok := registerResp["id"]
	if !ok || userID == "" {
		t.Fatal("expected registered user id in response")
	}

	loginBody := strings.NewReader(`{"email":"integration@example.com","password":"password123"}`)
	loginReq := httptest.NewRequest(http.MethodPost, "/auth/login", loginBody)
	loginRec := httptest.NewRecorder()

	authHandler.Login(loginRec, loginReq)

	if loginRec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, loginRec.Code, loginRec.Body.String())
	}

	var loginResp map[string]string
	decodeJSONResponse(t, strings.NewReader(loginRec.Body.String()), &loginResp)

	accessToken, ok := loginResp["access_token"]
	if !ok || accessToken == "" {
		t.Fatal("expected access_token in login response")
	}

	parsedClaims := &models.JWTClaims{}
	token, err := jwt.ParseWithClaims(accessToken, parsedClaims, func(token *jwt.Token) (any, error) {
		return []byte(defaultJWTSecret), nil
	})
	if err != nil {
		t.Fatalf("failed to parse JWT token: %v", err)
	}
	if !token.Valid {
		t.Fatal("expected valid JWT token")
	}
	if parsedClaims.UserID != userID {
		t.Fatalf("expected JWT user_id %q, got %q", userID, parsedClaims.UserID)
	}
}

func TestProjectCreateAndListIntegration(t *testing.T) {
	db := connectTestDB(t)
	truncateTables(t, db)

	authHandler := &handlers.AuthHandler{DB: db, JWTSecret: []byte(defaultJWTSecret)}
	projectHandler := &handlers.ProjectHandler{DB: db}

	registerBody := strings.NewReader(`{"name":"Project Tester","email":"project@example.com","password":"password123"}`)
	registerReq := httptest.NewRequest(http.MethodPost, "/auth/register", registerBody)
	registerRec := httptest.NewRecorder()
	authHandler.Register(registerRec, registerReq)

	if registerRec.Code != http.StatusCreated {
		t.Fatalf("expected register status %d, got %d", http.StatusCreated, registerRec.Code)
	}

	var registerResp map[string]string
	decodeJSONResponse(t, strings.NewReader(registerRec.Body.String()), &registerResp)
	userID := registerResp["id"]

	claims := &models.JWTClaims{UserID: userID, Email: "project@example.com"}

	projectBody := strings.NewReader(`{"name":"Test Project","description":"Integration test project"}`)
	projectReq := httptest.NewRequest(http.MethodPost, "/projects", projectBody)
	projectReq = projectReq.WithContext(context.WithValue(projectReq.Context(), models.UserContextKey, claims))
	projectRec := httptest.NewRecorder()

	projectHandler.Create(projectRec, projectReq)
	if projectRec.Code != http.StatusCreated {
		t.Fatalf("expected project create status %d, got %d: %s", http.StatusCreated, projectRec.Code, projectRec.Body.String())
	}

	var createdProject models.Project
	decodeJSONResponse(t, strings.NewReader(projectRec.Body.String()), &createdProject)
	if createdProject.Name != "Test Project" {
		t.Fatalf("expected created project name %q, got %q", "Test Project", createdProject.Name)
	}

	listReq := httptest.NewRequest(http.MethodGet, "/projects", nil)
	listReq = listReq.WithContext(context.WithValue(listReq.Context(), models.UserContextKey, claims))
	listRec := httptest.NewRecorder()

	projectHandler.List(listRec, listReq)
	if listRec.Code != http.StatusOK {
		t.Fatalf("expected project list status %d, got %d: %s", http.StatusOK, listRec.Code, listRec.Body.String())
	}

	var listedProjects []models.Project
	decodeJSONResponse(t, strings.NewReader(listRec.Body.String()), &listedProjects)
	if len(listedProjects) != 1 {
		t.Fatalf("expected 1 project, got %d", len(listedProjects))
	}
	if listedProjects[0].ID != createdProject.ID {
		t.Fatalf("expected listed project id %q, got %q", createdProject.ID, listedProjects[0].ID)
	}
}

func TestAuthMiddlewareIntegration(t *testing.T) {
	expectedClaims := &models.JWTClaims{
		UserID: "middleware-user-id",
		Email:  "middleware@example.com",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, expectedClaims)
	tokenString, err := token.SignedString([]byte(defaultJWTSecret))
	if err != nil {
		t.Fatalf("failed to sign JWT token: %v", err)
	}

	wasCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wasCalled = true
		claims, ok := r.Context().Value(models.UserContextKey).(*models.JWTClaims)
		if !ok || claims == nil {
			t.Fatal("expected JWT claims in context")
		}
		if claims.Email != expectedClaims.Email {
			t.Fatalf("expected claims email %q, got %q", expectedClaims.Email, claims.Email)
		}
		w.WriteHeader(http.StatusOK)
	})

	handler := middleware.Auth([]byte(defaultJWTSecret), next)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tokenString))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if !wasCalled {
		t.Fatal("expected protected handler to be called")
	}
}
