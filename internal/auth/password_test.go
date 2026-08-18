package auth

import "testing"

func TestHashPassword(t *testing.T) {
	hash, err := HashPassword("correct-horse-battery-staple")
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}
	if hash == "" {
		t.Fatal("HashPassword returned an empty hash")
	}
	if hash == "correct-horse-battery-staple" {
		t.Fatal("HashPassword returned the plaintext password instead of a hash")
	}
}

func TestHashPassword_DifferentHashesForSamePassword(t *testing.T) {
	password := "correct-horse-battery-staple"

	hash1, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}
	hash2, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}

	if hash1 == hash2 {
		t.Fatal("HashPassword produced identical hashes for the same password on two calls (missing salt)")
	}
}

func TestHashPassword_TooLong(t *testing.T) {
	// bcrypt отклоняет пароли длиннее 72 байт
	longPassword := make([]byte, 73)
	for i := range longPassword {
		longPassword[i] = 'a'
	}

	if _, err := HashPassword(string(longPassword)); err == nil {
		t.Fatal("expected HashPassword to return an error for a password longer than 72 bytes, got nil")
	}
}

func TestCheckPassword_CorrectPassword(t *testing.T) {
	password := "correct-horse-battery-staple"
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}

	if err := CheckPassword(password, hash); err != nil {
		t.Fatalf("CheckPassword failed for the correct password: %v", err)
	}
}

func TestCheckPassword_WrongPassword(t *testing.T) {
	hash, err := HashPassword("correct-horse-battery-staple")
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}

	if err := CheckPassword("wrong-password", hash); err == nil {
		t.Fatal("expected CheckPassword to return an error for the wrong password, got nil")
	}
}

func TestCheckPassword_EmptyPassword(t *testing.T) {
	hash, err := HashPassword("correct-horse-battery-staple")
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}

	if err := CheckPassword("", hash); err == nil {
		t.Fatal("expected CheckPassword to return an error for an empty password, got nil")
	}
}

func TestCheckPassword_InvalidHash(t *testing.T) {
	if err := CheckPassword("correct-horse-battery-staple", "not-a-real-bcrypt-hash"); err == nil {
		t.Fatal("expected CheckPassword to return an error for a malformed hash, got nil")
	}
}
