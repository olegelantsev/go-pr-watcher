package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zalando/go-keyring"
)

func TestStoreSecret(t *testing.T) {
	keyring.MockInit()

	storage := NewSecretStorage()
	storage.store("super-secret")

	token, err := storage.read()
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, token, "super-secret", "Should be equal strings")
}
