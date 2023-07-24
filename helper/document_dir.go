package helper

import (
	"os/user"
	"path/filepath"
	"runtime"
)

func GetDocumentsDir() (string, error) {
	// Get the current user's information
	currentUser, err := user.Current()
	if err != nil {
		return "", err
	}

	// Retrieve the home directory
	homeDir := currentUser.HomeDir

	// Check the operating system and build the documents directory path accordingly
	var documentsDir string
	switch runtimeOS := runtime.GOOS; runtimeOS {
	case "windows":
		documentsDir = filepath.Join(homeDir, "Documents")
	case "darwin": // macOS
		documentsDir = filepath.Join(homeDir, "Documents")
	default: // Linux and other platforms
		documentsDir = filepath.Join(homeDir, "Documents")
	}

	return documentsDir, nil
}
