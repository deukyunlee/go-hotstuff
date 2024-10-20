package util

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"testing"
)

func TestHash(t *testing.T) {
	content := []byte("test content")
	expectedHash := sha256.Sum256(content)
	expectedHashStr := hex.EncodeToString(expectedHash[:])

	hashed := Hash(content)

	if hashed != expectedHashStr {
		t.Errorf("expected hash %s, but got %s", expectedHashStr, hashed)
	}
}

func TestDigest(t *testing.T) {
	testObject := map[string]interface{}{
		"key": "value",
	}

	msg, err := json.Marshal(testObject)
	if err != nil {
		t.Fatalf("unexpected error during manual marshalling: %v", err)
	}
	expectedHash := Hash(msg)

	digest, err := Digest(testObject)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if digest != expectedHash {
		t.Errorf("expected digest %s, but got %s", expectedHash, digest)
	}
}
