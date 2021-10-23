package main

import (
	"log"

	"github.com/zalando/go-keyring"
)

type SecretStorage struct {
	service string
	user    string
}

func NewSecretStorage() SecretStorage {
	return SecretStorage{
		service: "go-pr-watcher",
		user:    "github-pat",
	}
}

func (storage *SecretStorage) store(secret string) {
	keyring.Set(storage.service, storage.user, secret)
}

func (storage *SecretStorage) read() (string, error) {
	return keyring.Get(storage.service, storage.user)
}

func (storage *SecretStorage) delete() {
	err := keyring.Delete(storage.service, storage.user)
	if err != nil {
		log.Fatal(err)
	}
}
