package auth

import (
	"testing"
	"time"
)

func TestHashPasswordAndComparePassword(t *testing.T) {
	hash, err := HashPassword("s3cr3t-pass")
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}

	if err := ComparePassword(hash, "s3cr3t-pass"); err != nil {
		t.Fatalf("ComparePassword() error = %v", err)
	}

	if err := ComparePassword(hash, "wrong-pass"); err == nil {
		t.Fatal("ComparePassword() expected mismatch error")
	}
}

func TestIssueTokenPairAndParseToken(t *testing.T) {
	pair, err := IssueTokenPair("secret", time.Hour, 24*time.Hour, 42, "alice")
	if err != nil {
		t.Fatalf("IssueTokenPair() error = %v", err)
	}

	accessClaims, err := ParseTokenWithType("secret", pair.AccessToken, TokenTypeAccess)
	if err != nil {
		t.Fatalf("ParseTokenWithType(access) error = %v", err)
	}
	if accessClaims.UserID != 42 || accessClaims.Username != "alice" {
		t.Fatalf("unexpected access claims: %+v", accessClaims)
	}

	refreshClaims, err := ParseTokenWithType("secret", pair.RefreshToken, TokenTypeRefresh)
	if err != nil {
		t.Fatalf("ParseTokenWithType(refresh) error = %v", err)
	}
	if refreshClaims.ID != pair.RefreshTokenJTI {
		t.Fatalf("unexpected refresh jti: got %s want %s", refreshClaims.ID, pair.RefreshTokenJTI)
	}
}

func TestParseTokenWithTypeRejectsWrongType(t *testing.T) {
	pair, err := IssueTokenPair("secret", time.Hour, 24*time.Hour, 7, "bob")
	if err != nil {
		t.Fatalf("IssueTokenPair() error = %v", err)
	}

	if _, err := ParseTokenWithType("secret", pair.AccessToken, TokenTypeRefresh); err == nil {
		t.Fatal("ParseTokenWithType() expected token type mismatch")
	}
}
