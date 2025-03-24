package utils

import (
	"log"
	"strings"

	"github.com/gofrs/uuid/v5"
)

func GenerateUuid() (string, error) {
	// Create a Version 4 UUID.
	u2, err := uuid.NewV4()
	if err != nil {
		log.Fatalf("failed to generate UUID: %v", err)
		return "", err
	}

	return u2.String(), nil
}

func ExtractUUIDFromUrl(path string, desiredUrl string) (string, string) {
	if strings.HasPrefix(path, "/"+desiredUrl+"/") {
		id := strings.TrimPrefix(path, "/"+desiredUrl+"/")
		return id, ""
	} else {
		return "", "not found"
	}
}
