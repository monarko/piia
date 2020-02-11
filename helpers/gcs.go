package helpers

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"time"

	"cloud.google.com/go/storage"
)

// GetSignedURL returns a signed url
func GetSignedURL(bucketName, filePath, serviceAccountFilePath string) (string, error) {
	err := os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", serviceAccountFilePath)
	if err != nil {
		return "", err
	}

	credentialFile, err := ioutil.ReadFile(serviceAccountFilePath)
	if err != nil {
		return "", err
	}

	credentialContent := make(map[string]string)
	if err := json.Unmarshal(credentialFile, &credentialContent); err != nil {
		return "", err
	}

	method := "GET"
	expires := time.Now().Add(time.Second * 60 * 10)
	googleStorageEmail := credentialContent["client_email"]
	googleStoragePrivateKey := credentialContent["private_key"]

	url, err := storage.SignedURL(bucketName, filePath, &storage.SignedURLOptions{
		GoogleAccessID: googleStorageEmail,
		PrivateKey:     []byte(googleStoragePrivateKey),
		Method:         method,
		Expires:        expires,
	})
	if err != nil {
		return "", err
	}
	return url, nil
}
