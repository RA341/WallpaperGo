package files

import (
	"gopkg.in/ini.v1"
	"strings"
)

// ReadConfig reads from config file
func ReadConfig(configPath string) (*ini.File, error) {
	return ini.Load(configPath)
}

// ReadSubredditList read csv seperated string and covert to array
func ReadSubredditList(stringArray string) []string {
	tmp := strings.Split(stringArray, ",")
	if tmp[0] == "" { // first element is empty string if no subreddits are found
		return nil
	}
	return tmp
}

// ConvertArrayToConfigList convert array to csv seperated string
func ConvertArrayToConfigList(array []string) string {
	return strings.Join(array, ",")
}
