package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const testSecret = "test-secret"

func TestGenerateToken_ReturnsNonEmptyToken(t *testing.T) {
	token, err := GenerateToken("user-123", testSecret)
	if err != nil {
		t.Fatalf("GenerateToken returned error: %v", err)
	}
	if token == "" {
		t.Fatal("GenerateToken returned an empty token")
	}
}

func TestGenerateAndValidateToken_Success(t *testing.T) {
	token, err := GenerateToken("user-123", testSecret)
	if err != nil {
		t.Fatalf("GenerateToken returned error: %v", err)
	}

	claims, err := ValidateToken(token, testSecret)
	if err != nil {
		t.Fatalf("ValidateToken returned error: %v", err)
	}
	if claims.UserID != "user-123" {
		t.Fatalf("expected UserID %q, got %q", "user-123", claims.UserID)
	}
}

func TestValidateToken_WrongSecret(t *testing.T) {
	token, err := GenerateToken("user-123", testSecret)
	if err != nil {
		t.Fatalf("GenerateToken returned error: %v", err)
	}

	if _, err := ValidateToken(token, "some-other-secret"); err == nil {
		t.Fatal("expected ValidateToken to return an error for a token signed with a different secret, got nil")
	}
}

func TestValidateToken_MalformedToken(t *testing.T) {
	if _, err := ValidateToken("not-a-valid-token", testSecret); err == nil {
		t.Fatal("expected ValidateToken to return an error for a malformed token, got nil")
	}
}

func TestValidateToken_EmptyToken(t *testing.T) {
	if _, err := ValidateToken("", testSecret); err == nil {
		t.Fatal("expected ValidateToken to return an error for an empty token, got nil")
	}
}

func TestValidateToken_ExpiredToken(t *testing.T) {
	claims := Claims{
		UserID: "user-123",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(testSecret))
	if err != nil {
		t.Fatalf("failed to sign test token: %v", err)
	}

	if _, err := ValidateToken(signed, testSecret); err == nil {
		t.Fatal("expected ValidateToken to return an error for an expired token, got nil")
	}
}
