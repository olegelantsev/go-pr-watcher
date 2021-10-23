package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"syscall"

	"github.com/rivo/tview"
	"golang.org/x/term"
	"gopkg.in/yaml.v2"
)

type RepoMap = map[string][]string

type Config struct {
	Repos RepoMap
	Token string
}

func loadToken() string {
	secretStorage := NewSecretStorage()
	token, err := secretStorage.read()
	if err != nil {
		if err.Error() == "secret not found in keyring" {
			token = readTokenFromInput()
			client := createGitHubClient(token)
			verifyToken(client)
			secretStorage.store(token)
		} else {
			log.Fatal(err)
		}
	}
	return token
}

func loadConfig(app *tview.Application) Config {
	var config Config

	yamlFile, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		log.Printf("yamlFile.Get err   #%v ", err)
	}
	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}
	config.Token = loadToken()
	return config
}

func readTokenFromInput() string {
	fmt.Fprintf(os.Stdout, "GitHub Personal Token:\n")
	bytebw, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		log.Fatal(err)
	}
	return string(bytebw)
}
