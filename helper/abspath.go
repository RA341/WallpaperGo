package helper

import (
	"log"
	"path/filepath"
)

func ConvertToAbsPath(path string) string {

	tmp, err := filepath.Abs(path)

	if err != nil {
		log.Fatalln("Failed to get absolute directory: ", err)
	}

	return tmp
}
