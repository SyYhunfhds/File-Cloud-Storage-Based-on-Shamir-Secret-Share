package v1

import (
	"os"
	"strings"
)

func getBasename(info os.FileInfo) string {
	return strings.Split(info.Name(), ".")[0]
}
func getExtname(info os.FileInfo) string {
	parts := strings.Split(info.Name(), ".")
	if len(parts) <= 1 {
		return ""
	}

	return parts[len(parts)-1]
}

func validateExtnameFunc(matches map[string]struct{}) func(ext string) bool {
	return func(ext string) bool {
		if _, exists := matches[ext]; exists {
			return true
		}
		return false
	}
}
