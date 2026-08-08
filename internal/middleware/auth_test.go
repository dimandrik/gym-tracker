package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"gym_tracker/internal/auth"

	"github.com/golang-jwt/jwt/v5"
)

const testSecret = "test-secret"

func TestAuthMiddleware_NoToken(t *testing.T) {
	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
	})

	handler := AuthMiddleware(testSecret)(next)

	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()

	handler(rec, req)

	if nextCalled {
		t.Error("next should not be called when there is no token")
	}
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
}

func TestAuthMiddleware_ValidToken(t *testing.T) {
	token, err := auth.GenerateToken("user-123", testSecret)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	var nextCalled bool
	var gotUserID any
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		gotUserID = r.Context().Value(UserIDKey)
	})

	handler := AuthMiddleware(testSecret)(next)

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	handler(rec, req)

	if !nextCalled {
		t.Fatal("next should be called when the token is valid")
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if gotUserID != "user-123" {
		t.Errorf("expected context userID %q, got %v", "user-123", gotUserID)
	}
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
	})

	handler := AuthMiddleware(testSecret)(next)

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer not-a-real-token")
	rec := httptest.NewRecorder()

	handler(rec, req)

	if nextCalled {
		t.Error("next should not be called when the token is invalid")
	}
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
}

func TestAuthMiddleware_WrongSecret(t *testing.T) {
	token, err := auth.GenerateToken("user-123", "some-other-secret")
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
	})

	handler := AuthMiddleware(testSecret)(next)

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	handler(rec, req)

	if nextCalled {
		t.Error("next should not be called when the token was signed with a different secret")
	}
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
}

func TestAuthMiddleware_ExpiredToken(t *testing.T) {
	claims := auth.Claims{
		UserID: "user-123",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
		},
	}
	expired, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(testSecret))
	if err != nil {
		t.Fatalf("failed to sign test token: %v", err)
	}

	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
	})

	handler := AuthMiddleware(testSecret)(next)

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer "+expired)
	rec := httptest.NewRecorder()

	handler(rec, req)

	if nextCalled {
		t.Error("next should not be called when the token is expired")
	}
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
}

func TestAuthMiddleware_WrongAuthScheme(t *testing.T) {
	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
	})

	handler := AuthMiddleware(testSecret)(next)

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Basic not-a-bearer-token")
	rec := httptest.NewRecorder()

	handler(rec, req)

	if nextCalled {
		t.Error("next should not be called when the Authorization header does not use the Bearer scheme")
	}
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
}

// TrimPrefix is a no-op when the prefix is absent, so a bare token
// (no "Bearer " prefix at all) is still accepted as-is by the current
// implementation. This test documents that behavior rather than an
// intended security boundary.
func TestAuthMiddleware_TokenWithoutBearerPrefixStillValidates(t *testing.T) {
	token, err := auth.GenerateToken("user-123", testSecret)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
	})

	handler := AuthMiddleware(testSecret)(next)

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", token)
	rec := httptest.NewRecorder()

	handler(rec, req)

	if !nextCalled {
		t.Error("expected next to be called since TrimPrefix leaves a prefix-less token untouched")
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}
