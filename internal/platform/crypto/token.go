package crypto

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

type tokenClaims struct {
	Subject   string `json:"sub"`
	IssuedAt  int64  `json:"iat"`
	ExpiresAt int64  `json:"exp"`
}

// IssueToken creates an encrypted token for the given subject.
// The secret must be at least 32 bytes; it is hashed to produce a 32-byte AES key.
func IssueToken(subject string, expiry time.Duration, secret []byte) (string, error) {
	if subject == "" {
		return "", errors.New("crypto: subject must not be empty")
	}
	if len(secret) == 0 {
		return "", errors.New("crypto: secret must not be empty")
	}

	now := time.Now()
	claims := tokenClaims{
		Subject:   subject,
		IssuedAt:  now.Unix(),
		ExpiresAt: now.Add(expiry).Unix(),
	}

	payload, err := json.Marshal(claims)
	if err != nil {
		return "", fmt.Errorf("crypto: marshaling claims: %w", err)
	}

	key := deriveKey(secret)
	ct, err := Seal(payload, key)
	if err != nil {
		return "", fmt.Errorf("crypto: sealing token: %w", err)
	}

	return base64.RawURLEncoding.EncodeToString(ct), nil
}

// VerifyToken decrypts and validates a token, returning the subject claim.
func VerifyToken(token string, secret []byte) (string, error) {
	if token == "" {
		return "", errors.New("crypto: token must not be empty")
	}
	if len(secret) == 0 {
		return "", errors.New("crypto: secret must not be empty")
	}

	ct, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil {
		return "", errors.New("crypto: invalid token encoding")
	}

	key := deriveKey(secret)
	payload, err := Open(ct, key)
	if err != nil {
		return "", errors.New("crypto: invalid or tampered token")
	}

	var claims tokenClaims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return "", errors.New("crypto: invalid token payload")
	}

	if time.Now().Unix() > claims.ExpiresAt {
		return "", errors.New("crypto: token expired")
	}

	return claims.Subject, nil
}

// deriveKey produces a 32-byte AES key from an arbitrary-length secret.
func deriveKey(secret []byte) []byte {
	h := sha256.Sum256(secret)
	return h[:]
}
