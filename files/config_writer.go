package files

import (
	"gopkg.in/ini.v1"
	"strings"
)

// ReadConfig reads from config file
func ReadConfig(configPath string) (*ini.File, error) {
	return ini.Load(configPath)
}

// ReadListFromConfig read csv seperated string and covert to array
func ReadListFromConfig(stringArray string) []string {
	return strings.Split(stringArray, ",")
}

// ConvertArrayToConfigList convert array to csv seperated string
func ConvertArrayToConfigList(array []string) string {
	return strings.Join(array, ",")
}
