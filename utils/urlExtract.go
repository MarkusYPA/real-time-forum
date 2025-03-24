package utils

import (
	"strings"
)

func ExtractFromUrl(path string, desiredUrl string) (string, string) {
	if strings.HasPrefix(path, "/"+desiredUrl+"/") {
		id := strings.TrimPrefix(path, "/"+desiredUrl+"/")
		return id, ""
	} else {
		return "", "not found"
	}
}
