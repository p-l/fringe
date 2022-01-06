package system

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/sethvargo/go-password/password"
)

type Secrets struct {
	Radius string `json:"radius"`
	JWT    string `json:"jwt"`
}

const (
	radiusSecretLen       = 64
	jwtSecretLen          = 128
	secretNumDigit        = 2
	secretNumSymbols      = 2
	secretsFilePermission = 0o600
)

func generateSecret(length int) string {
	secret, err := password.Generate(length, secretNumDigit, secretNumSymbols, false, true)
	if err != nil {
		log.Panicf("could not generate secrets: %v", err)
	}

	return secret
}

func LoadSecretsFromFile(file string) Secrets {
	var secrets Secrets

	// If file not exist create it
	if _, err := os.Stat(file); errors.Is(err, os.ErrNotExist) {
		secrets = Secrets{
			Radius: generateSecret(radiusSecretLen),
			JWT:    generateSecret(jwtSecretLen),
		}
		SaveSecretsToFile(secrets, file)

		return secrets
	}

	bytes, err := ioutil.ReadFile(filepath.Clean(file))
	if err != nil {
		log.Panicf("could not read secrets repository file %s: %v", file, err)
	}

	err = json.Unmarshal(bytes, &secrets)
	if err != nil {
		log.Panicf("could not parse secrets from file %s: %v", file, err)
	}

	// Ensure all secrets exists
	if len(secrets.Radius) == 0 {
		secrets.Radius = generateSecret(radiusSecretLen)
		SaveSecretsToFile(secrets, file)
	}

	if len(secrets.JWT) == 0 {
		secrets.JWT = generateSecret(jwtSecretLen)
		SaveSecretsToFile(secrets, file)
	}

	return secrets
}

func SaveSecretsToFile(secrets Secrets, file string) {
	jsonData, _ := json.MarshalIndent(secrets, "", " ")
	_ = ioutil.WriteFile(filepath.Clean(file), jsonData, secretsFilePermission)
}
